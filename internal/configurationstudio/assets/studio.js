"use strict";
const { message, text, show, currentMessage } = ControlLanguage;
class LocalError extends Error {
  constructor(copy) { super(text(copy)); this.copy = copy; }
}
const $ = (id) => document.getElementById(id);
const form = $("configuration-form");
const mode = $("creation-mode");
const workspaceChoice = $("existing-workspace");
const configurationChoice = $("existing-configuration");
const draftChoice = $("existing-draft");
const sectionChoice = $("draft-section");
const timeoutFields = [
  ["handshakeSeconds", "timeout-handshake", 1, 300, "Handshake"],
  ["readSeconds", "timeout-read", 0, 86400, "Read"],
  ["writeSeconds", "timeout-write", 1, 300, "Write"],
  ["idleSeconds", "timeout-idle", 0, 86400, "Idle"],
].map(([key, id, min, max, label]) => ({ key, input: $(id), min, max, label, output: $(`${id}-error`) }));
const submitButton = $("submit-button");
const publishButton = $("publish-draft"), reloadDraftButton = $("reload-draft");
const statusOutput = $("flow-status");
const result = $("result");
const openCreatedDraftButton = $("open-created-draft"), startAnotherButton = $("start-another");
const resultState = $("result-state");
const resultVersion = $("result-version");
const resultURL = $("result-url");
const controls = [mode, workspaceChoice, configurationChoice, $("reload-lists"),
  $("workspace-name"), $("configuration-name"), $("listener-host"), $("listener-port"), sectionChoice,
  ...timeoutFields.map(({ input }) => input), submitButton];
let busy = false, locked = false, loading = false, listGeneration = 0, listController;
let workspaces = new Map(), configurations = new Map(), drafts = new Map();
let createdDraft = null;
let openedDraft = null, saveUnknown = false;
let publishNeedsReload = false, publicationRecovery = null;
let editorSection = "listener";
let choices = { mode: mode.value, workspace: "", configuration: "", draft: "" };
const editing = () => mode.value === "existing-draft";
const hostInput = $("listener-host"), portInput = $("listener-port");
const listenerErrors = { host: $("listener-host-error"), port: $("listener-port-error") };
const listenerFields = { host: hostInput, port: portInput };
const rejectedListenerValues = new Map();
const listenerErrorSummary = "Check the highlighted Listener fields.";
const timeoutErrorSummary = "Check the highlighted Timeout fields.";

function validIPv4(host) {
  const octets = host.split(".");
  return octets.length === 4 && octets.every((part) =>
    part.length > 0 && part.length <= 3 && !/[^0-9]/.test(part) &&
    (part.length === 1 || part[0] !== "0") && Number(part) <= 255);
}

function validIPv6(host) {
  if (host.includes(".")) {
    const suffix = host.lastIndexOf(":") + 1;
    if (!validIPv4(host.slice(suffix))) return false;
    // An embedded IPv4 address occupies two 16-bit groups.
    host = host.slice(0, suffix) + "0:0";
  }
  const sides = host.split("::");
  if (sides.length > 2) return false;
  const groups = sides.flatMap((side) => side ? side.split(":") : []);
  if (groups.some((group) => !group || group.length > 4 || /[^0-9a-f]/i.test(group))) return false;
  return sides.length === 1 ? groups.length === 8 : groups.length < 8;
}

function listenerHost(value) {
  // Go strings.TrimSpace follows Unicode White_Space, including NEL but excluding BOM.
  const host = value.replace(/^\p{White_Space}+|\p{White_Space}+$/gu, "");
  if (!host) return { error: "Enter a Listener host." };
  if ([...host].length > 255) return { error: "Listener host must not exceed 255 characters." };

  if (host.includes(".") && !/[^0-9.]/.test(host)) {
    if (validIPv4(host)) return { value: host };
    return { error: "Enter a valid IPv4 address for a dotted numeric host." };
  }
  if (host.includes(":")) return validIPv6(host) ? { value: host } : { error: "Enter a valid IPv6 address." };
  const labels = (host.endsWith(".") ? host.slice(0, -1) : host).split(".");
  if (labels.every((label) => label.length > 0 && label.length <= 63 &&
    !/[^a-z0-9-]/i.test(label) && !label.startsWith("-") && !label.endsWith("-"))) return { value: host };
  return { error: "Enter a valid IP address or hostname." };
}

function listenerPort(raw) {
  const error = "Enter a whole port number between 1 and 65535.";
  if (!raw || /[^0-9]/.test(raw)) return { error };
  const digits = raw.replace(/^0+/, "");
  if (!digits || digits.length > 5 || (digits.length === 5 && digits > "65535")) return { error };
  return { value: Number(digits) };
}

function setFieldError(input, output, copy = "") {
  show(output, copy);
  if (copy) input.setAttribute("aria-invalid", "true");
  else input.removeAttribute("aria-invalid");
}

function setListenerError(field, copy) {
  setFieldError(listenerFields[field], listenerErrors[field], copy);
}

function clearListenerSummary() {
  if (currentMessage(statusOutput) === listenerErrorSummary) setStatus("Ready.");
}

function clearListenerErrors() {
  rejectedListenerValues.clear();
  setListenerError("host");
  setListenerError("port");
  clearListenerSummary();
}

function showListenerRejection(error) {
  if (!(error instanceof RequestError) || !error.listenerRejection) return null;
  const { field, message } = error.listenerRejection;
  const input = listenerFields[field];
  rejectedListenerValues.set(field, input.value);
  setListenerError(field, message);
  return input;
}

function validateListenerInput(field, validate) {
  const value = listenerFields[field].value;
  if (rejectedListenerValues.get(field) === value) return { error: currentMessage(listenerErrors[field]) };
  rejectedListenerValues.delete(field);
  const result = validate(value);
  setListenerError(field, result.error);
  return result;
}

function validateListenerFields() {
  const host = validateListenerInput("host", listenerHost);
  const port = validateListenerInput("port", listenerPort);
  if (host.error || port.error) {
    setStatus(listenerErrorSummary, "error");
    (host.error ? hostInput : portInput).focus();
    return null;
  }
  clearListenerSummary();
  return { host: host.value, port: port.value };
}

for (const [field, validate] of [["host", listenerHost], ["port", listenerPort]]) {
  const input = listenerFields[field];
  function updateListenerError() {
    validateListenerInput(field, validate);
    if (!currentMessage(listenerErrors.host) && !currentMessage(listenerErrors.port)) clearListenerSummary();
  }
  input.addEventListener("input", updateListenerError);
  input.addEventListener("blur", updateListenerError);
}

function validateTimeoutField({ input, min, max, label, output }) {
  const value = Number(input.value);
  if (!input.value || /[^0-9]/.test(input.value) || !Number.isInteger(value) || value < min || value > max) {
    setFieldError(input, output, message("{0} timeout must be a whole number between {1} and {2} seconds.", message(label), min, max));
    return undefined;
  }
  setFieldError(input, output);
  return value;
}

function validateTimeoutFields() {
  const body = {};
  let firstInvalid;
  for (const field of timeoutFields) {
    const value = validateTimeoutField(field);
    if (value === undefined) firstInvalid ??= field.input;
    else body[field.key] = value;
  }
  if (firstInvalid) {
    setStatus(timeoutErrorSummary, "error");
    firstInvalid.focus();
    return null;
  }
  return body;
}

function clearTimeoutErrors() {
  for (const { input, output } of timeoutFields) setFieldError(input, output);
  if (currentMessage(statusOutput) === timeoutErrorSummary) setStatus("Ready.");
}

for (const field of timeoutFields) {
  field.input.addEventListener("input", () => {
    if (!currentMessage(field.output)) return;
    validateTimeoutField(field);
    if (currentMessage(statusOutput) === timeoutErrorSummary &&
        timeoutFields.every(({ output }) => !currentMessage(output))) setStatus("Ready.");
  });
}

function dirty() {
  if (openedDraft && editorSection === "timeouts") return timeoutFields.some(({ key, input }) => input.value !== String(openedDraft.timeouts[key]));
  return !!openedDraft && (hostInput.value !== openedDraft.host || portInput.value !== String(openedDraft.port));
}
function rememberChoices() {
  choices = { mode: mode.value, workspace: workspaceChoice.value, configuration: configurationChoice.value, draft: draftChoice.value };
}
function restoreChoices() {
  mode.value = choices.mode; workspaceChoice.value = choices.workspace;
  configurationChoice.value = choices.configuration; draftChoice.value = choices.draft;
  sectionChoice.value = editorSection;
}
const confirmation = $("studio-confirmation");
let pendingConfirmation = null;

function restoreConfirmationFocus(trigger) {
  if (trigger.getClientRects().length && !trigger.disabled) trigger.focus();
  else if (!$("navigation-toggle").hidden) $("navigation-toggle").focus();
}

function finishConfirmation(confirmed) {
  const pending = pendingConfirmation;
  if (!pending) return;
  pendingConfirmation = null;
  confirmation.close();
  restoreConfirmationFocus(pending.trigger);
  if (!confirmed || busy || (locked && !pending.allowLocked)) {
    pending.resolve();
    return;
  }
  try { pending.resolve(pending.action()); }
  catch (error) { pending.reject(error); }
}

$("confirmation-cancel").addEventListener("click", () => finishConfirmation(false));
$("confirmation-accept").addEventListener("click", () => finishConfirmation(true));
confirmation.addEventListener("keydown", (event) => {
  if (event.key === "Escape") event.stopPropagation();
});
confirmation.addEventListener("cancel", (event) => {
  event.preventDefault();
  finishConfirmation(false);
});
window.addEventListener("pagehide", () => finishConfirmation(false));

function requestAction(trigger, action, copy = "Discard local Draft edits and leave this Draft?", allowLocked = false) {
  if (pendingConfirmation || busy || publicationRecovery || (locked && !allowLocked)) return;
  if (!dirty() && !saveUnknown) return action();
  return confirmAction(trigger, action, copy, "Discard Draft edits?", "Discard edits", saveUnknown, allowLocked);
}

function confirmAction(trigger, action, copy, title, accept, warning = false, allowLocked = false) {
  show($("confirmation-title"), title);
  show($("confirmation-accept"), accept);
  show($("confirmation-message"), copy);
  $("confirmation-warning").hidden = !warning;
  confirmation.setAttribute("aria-describedby", warning
    ? "confirmation-message confirmation-warning" : "confirmation-message");
  return new Promise((resolve, reject) => {
    pendingConfirmation = { trigger, action, allowLocked, resolve, reject };
    try { confirmation.showModal(); }
    catch (error) { pendingConfirmation = null; reject(error); }
  });
}

function guardChoice(select, action, copy) {
  select.addEventListener("change", () => {
    const proposed = select.value;
    restoreChoices();
    return requestAction(select, () => {
      select.value = proposed;
      return action();
    }, copy);
  });
}

function discardAndLeave(url) {
  // Confirmation authorizes discarding local edits before attempting navigation.
  // Keep beforeunload armed for any subsequent edit, even if this navigation fails.
  restoreEditorInputs();
  if (saveUnknown) {
    // Reopening from the existing list performs a fresh read before enabling Save.
    openedDraft = null;
    draftChoice.value = "";
    saveUnknown = false;
    rememberChoices();
    draftStatus("Select a Draft to open it.");
  } else updateDirty();
  syncControls();
  setStatus("Ready.");
  try { window.location.assign(url); }
  catch (_) { setStatus("Could not open the requested page.", "error"); }
}

document.addEventListener("click", (event) => {
  if (event.defaultPrevented || event.button !== 0 ||
      event.altKey || event.ctrlKey || event.metaKey || event.shiftKey) return;
  const link = event.target.closest("a[href]");
  if (!link || link.hasAttribute("download") || (link.target && link.target !== "_self")) return;
  const destination = new URL(link.href);
  const current = window.location;
  if (destination.origin !== current.origin || !["/", "/dashboard"].includes(destination.pathname)) return;
  const sameDocument = destination.pathname === current.pathname && destination.search === current.search;
  if (sameDocument && (destination.hash || (!$("navigation-toggle").hidden &&
      destination.href === current.href && $("control-navigation").contains(link)))) return;
  if (!busy && !dirty() && !saveUnknown) return;
  event.preventDefault();
  return requestAction(link, () => discardAndLeave(destination.href),
    "Discard local Draft edits and leave this page?", true);
}, true);

function draftStatus(copy) { show($("draft-status"), copy); }
function clearDrafts() {
  drafts = new Map(); openedDraft = null; saveUnknown = false;
  publishNeedsReload = false;
  editorSection = "listener"; sectionChoice.value = editorSection;
  clearListenerErrors();
  clearTimeoutErrors();
  placeholder(draftChoice, "Select a Draft");
  draftStatus("Select a Configuration to discover Drafts.");
}
function updateDirty() {
  if (editing() && openedDraft && !saveUnknown && !publishNeedsReload && !publicationRecovery) draftStatus(dirty() ? "Unsaved Draft changes." : "Draft section is unchanged.");
  syncControls();
}
for (const input of [hostInput, portInput, ...timeoutFields.map(({ input }) => input)]) input.addEventListener("input", updateDirty);
window.addEventListener("beforeunload", (event) => {
  if (busy || dirty() || saveUnknown) { event.preventDefault(); event.returnValue = ""; }
});
class RequestError extends LocalError {}
function setStatus(message, kind = "") {
  show(statusOutput, message);
  statusOutput.className = `status ${kind}`.trim();
}
function requiredID(value) {
  if (!Number.isSafeInteger(value) || value < 1) throw new LocalError("Invalid resource ID in Control Service response.");
  return value;
}
function selectedID(select, items) {
  if (!/^[1-9][0-9]*$/.test(select.value)) throw new LocalError("Select an available resource first.");
  const id = requiredID(Number(select.value));
  if (!items.has(id)) throw new LocalError("Select an available resource first.");
  return id;
}
function clearResult() {
  createdDraft = null;
  openCreatedDraftButton.hidden = true;
  result.hidden = true;
  show(resultState, "");
  resultVersion.textContent = resultURL.textContent = "";
}
function placeholder(select, label) {
  const option = document.createElement("option");
  option.value = ""; show(option, label);
  select.replaceChildren(option);
}
function clearConfigurations() {
  configurations = new Map();
  placeholder(configurationChoice, "Select a Configuration");
  clearDrafts();
  clearResult();
}
function syncControls() {
  for (const control of controls) control.disabled = busy || locked || !!publicationRecovery;
  openCreatedDraftButton.disabled = busy;
  startAnotherButton.disabled = busy;
  workspaceChoice.disabled ||= loading || workspaces.size === 0;
  configurationChoice.disabled ||= loading || configurations.size === 0;
  submitButton.disabled ||= loading;
  draftChoice.disabled = busy || locked || loading || !!publicationRecovery || drafts.size === 0;
  reloadDraftButton.disabled = busy || locked || loading || (!openedDraft && !publicationRecovery);
  show(reloadDraftButton, publicationRecovery ? "Refresh versions" : "Reload Draft");
  publishButton.disabled = !!pendingConfirmation || !canPublish(openedDraft);
  sectionChoice.disabled ||= loading || !openedDraft;
  $("timeout-fields").hidden = !editing() || editorSection !== "timeouts";
  $("listener-fields").hidden = editing() ? editorSection !== "listener" : mode.value === "existing-configuration";
  for (const { input } of timeoutFields) input.disabled ||= !editing() || !openedDraft || loading;
  if (editing()) {
    show(submitButton, editorSection === "timeouts" ? "Save timeouts" : "Save listener");
    hostInput.disabled ||= !openedDraft || loading;
    portInput.disabled ||= !openedDraft || loading;
    submitButton.disabled ||= !openedDraft || saveUnknown || !dirty();
  }
}
function invalidateLists() {
  listGeneration++;
  if (listController) listController.abort();
  loading = false;
}
async function requestJSON(path, method, expectedStatus, body, signal) {
  const controller = signal ? null : new AbortController();
  const timer = controller ? setTimeout(() => controller.abort(), 10000) : null;
  const options = { method, mode: "same-origin", credentials: "same-origin", cache: "no-store",
    redirect: "error", signal: signal || controller.signal };
  if (body !== undefined) {
    options.headers = { "Content-Type": "application/json" };
    options.body = JSON.stringify(body);
  }
  try {
    const response = await fetch(path, options);
    if (response.status !== expectedStatus) {
      const failure = Object.assign(new RequestError(message("Control Service returned HTTP {0}.", response.status)), { status: response.status });
      if (response.status === 400 && method === "PUT" && path.endsWith("/listener")) {
        const payload = await response.json().catch(() => null);
        const details = payload?.error;
        const match = details?.code === "validation_failed" && typeof details.message === "string" &&
          /^(host|port): (.+)$/.exec(details.message);
        if (match) {
          const [, field, reason] = match;
          const known = field === "host"
            ? ["must not be empty", "must not exceed 255 characters", "must be a valid IP address or hostname"]
            : ["must be between 1 and 65535"];
          failure.listenerRejection = { field, message: known.includes(reason) ? reason
            : field === "host" ? "Control Service rejected the listener host." : "Control Service rejected the listener port." };
        }
      }
      throw failure;
    }
    const payload = await response.json();
    if (payload === null || typeof payload !== "object") throw new LocalError("Unexpected Control Service response.");
    return payload;
  } finally {
    if (timer !== null) clearTimeout(timer);
  }
}
function resourceList(payload, workspaceID) {
  if (!Array.isArray(payload)) throw new LocalError("Invalid resource list.");
  const items = new Map();
  for (const item of payload) {
    if (!item || typeof item !== "object" || typeof item.name !== "string" || !item.name.trim()) throw new LocalError("Invalid resource list.");
    const id = requiredID(item.id);
    if (items.has(id) || (workspaceID !== undefined && requiredID(item.workspaceId) !== workspaceID)) throw new LocalError("Invalid resource scope.");
    items.set(id, item);
  }
  return items;
}
function fill(select, items, label) {
  placeholder(select, label);
  for (const [id, item] of items) {
    const option = document.createElement("option");
    option.value = String(id); option.textContent = `${item.name} (#${id})`;
    select.append(option);
  }
}
async function loadLists(workspaceID) {
  invalidateLists();
  const generation = listGeneration;
  listController = new AbortController();
  const timer = setTimeout(() => listController && generation === listGeneration && listController.abort(), 10000);
  loading = true;
  show($("list-status"), "Loading resources…");
  syncControls();
  try {
    const path = workspaceID === undefined ? "/api/v1/workspaces" : `/api/v1/workspaces/${workspaceID}/configurations`;
    const payload = await requestJSON(path, "GET", 200, undefined, listController.signal);
    if (generation !== listGeneration) return;
    const items = resourceList(payload, workspaceID);
    if (workspaceID === undefined) {
      workspaces = items;
      fill(workspaceChoice, items, "Select a Workspace");
    } else {
      configurations = items;
      fill(configurationChoice, items, "Select a Configuration");
    }
    show($("list-status"), items.size ? "Select a resource to continue." : workspaceID === undefined
      ? "No Workspaces. Choose New Workspace and Configuration to create one."
      : "No Configurations. Choose New Configuration in existing Workspace to create one.");
  } catch (_) {
    if (generation !== listGeneration) return;
    show($("list-status"), "Resources unavailable or invalid. Reload lists to try again.");
  } finally {
    clearTimeout(timer);
    if (generation === listGeneration) { loading = false; syncControls(); }
  }
}
function showModeFields() {
  const existing = mode.value === "existing-configuration" || editing();
  $("existing-resources").hidden = mode.value === "new-workspace";
  $("configuration-choice").hidden = !existing;
  $("workspace-fields").hidden = mode.value !== "new-workspace";
  $("configuration-fields").hidden = existing;
  $("listener-fields").hidden = existing && !editing();
  $("draft-fields").hidden = !editing();
  show(submitButton, editing() ? "Save listener" : existing ? "Create Draft" : "Create and publish");
}
function updateMode() {
  invalidateLists();
  workspaces = new Map();
  placeholder(workspaceChoice, "Select a Workspace");
  clearConfigurations();
  showModeFields();
  rememberChoices();
  setStatus("Ready.");
  syncControls();
  if (mode.value !== "new-workspace") return loadLists();
}
guardChoice(mode, updateMode);
guardChoice(workspaceChoice, () => {
  invalidateLists(); clearConfigurations();
  rememberChoices();
  setStatus("Ready."); syncControls();
  if ((mode.value === "existing-configuration" || editing()) && workspaceChoice.value) {
    try { return loadLists(selectedID(workspaceChoice, workspaces)); }
    catch (error) { setStatus(error.copy, "error"); }
  }
});
guardChoice(configurationChoice, () => {
  invalidateLists(); clearDrafts(); clearResult(); rememberChoices(); setStatus("Ready."); syncControls();
  if (editing() && configurationChoice.value) return loadDrafts();
});
$("reload-lists").addEventListener("click", (event) => requestAction(event.currentTarget, () => {
  clearConfigurations();
  workspaces = new Map(); placeholder(workspaceChoice, "Select a Workspace");
  rememberChoices();
  return loadLists();
}));
function inputValues(listener) {
  if (!["new-workspace", "new-configuration", "existing-configuration"].includes(mode.value)) throw new LocalError("Select an action.");
  const values = { mode: mode.value };
  if (mode.value !== "new-workspace") values.workspaceID = selectedID(workspaceChoice, workspaces);
  if (mode.value === "existing-configuration") {
    values.configurationID = selectedID(configurationChoice, configurations);
    return values;
  }
  values.workspaceName = $("workspace-name").value.trim();
  values.configurationName = $("configuration-name").value.trim();
  if ((mode.value === "new-workspace" && !values.workspaceName) || !values.configurationName) throw new LocalError("Complete every visible field.");
  values.listenerHost = listener.host;
  values.listenerPort = listener.port;
  return values;
}
function versionResponse(version, configurationID, state, versionID) {
  if (Array.isArray(version) || requiredID(version.configurationId) !== configurationID || version.state !== state ||
      !Number.isSafeInteger(version.number) || version.number < 1 || (versionID !== undefined && requiredID(version.id) !== versionID)) throw new LocalError("Unexpected version response or scope.");
  requiredID(version.id);
}
function preliminaryURL(host, port) {
  const renderedHost = host.includes(":") ? `[${host}]` : host;
  return `ws://${renderedHost}:${port}/ws`;
}
function showVersion(version, published) {
  show($("result-heading"), "Created version");
  show(resultState, version.state);
  resultVersion.textContent = String(version.number);
  $("result-url-row").hidden = !published;
  $("result-url-warning").hidden = false;
  if (published) {
    const listener = version.listener;
    if (!listener || typeof listener.host !== "string" || !listener.host || !Number.isInteger(listener.port) || listener.port < 1 || listener.port > 65535 || !listener.tls || listener.tls.enabled !== false) throw new LocalError("Unexpected listener response.");
    resultURL.textContent = preliminaryURL(listener.host, listener.port);
  }
  result.hidden = false;
  setStatus(published ? "Configuration version published. Published does not mean Running." : "Draft created. It has not been configured or published by Studio.", "success");
}
async function runFlow(values) {
  let workspaceID = values.workspaceID;
  if (values.mode === "new-workspace") {
    setStatus("Creating workspace…");
    const workspace = await requestJSON("/api/v1/workspaces", "POST", 201, { name: values.workspaceName, description: "" });
    workspaceID = requiredID(workspace.id);
  }
  let configurationID = values.configurationID;
  if (values.mode !== "existing-configuration") {
    setStatus("Creating configuration…");
    const configuration = await requestJSON(`/api/v1/workspaces/${workspaceID}/configurations`, "POST", 201,
      { name: values.configurationName, description: "" });
    configurationID = requiredID(configuration.id);
    if (requiredID(configuration.workspaceId) !== workspaceID) throw new LocalError("Unexpected configuration scope.");
  }
  setStatus("Creating draft version…");
  const draft = await requestJSON(`/api/v1/workspaces/${workspaceID}/configurations/${configurationID}/versions`, "POST", 201, {});
  versionResponse(draft, configurationID, "Draft");
  if (values.mode === "existing-configuration") {
    if (typeof draft.createdAt !== "string" || !Number.isFinite(Date.parse(draft.createdAt))) {
      throw new LocalError("Unexpected version response or scope.");
    }
    createdDraft = { workspaceID, configurationID: draft.configurationId, versionID: draft.id, createdAt: draft.createdAt };
    showVersion(draft, false);
    openCreatedDraftButton.hidden = false;
    return;
  }
  const versionID = draft.id;
  setStatus("Updating listener…");
  let updated;
  try {
    updated = await requestJSON(`/api/v1/workspaces/${workspaceID}/configurations/${configurationID}/versions/${versionID}/listener`, "PUT", 200,
      { host: values.listenerHost, port: values.listenerPort });
  } catch (error) {
    if (error instanceof RequestError && error.listenerRejection) {
      error.partialCreation = values.mode === "new-workspace"
        ? message("Created Workspace #{0}, Configuration #{1} and Draft #{2}.", workspaceID, configurationID, versionID)
        : message("Created Configuration #{0} in Workspace #{1} and Draft #{2}.", configurationID, workspaceID, versionID);
    }
    throw error;
  }
  versionResponse(updated, configurationID, "Draft", versionID);
  if (!updated.listener || updated.listener.host !== values.listenerHost || updated.listener.port !== values.listenerPort || !updated.listener.tls || updated.listener.tls.enabled !== false) throw new LocalError("Unexpected listener response.");
  setStatus("Publishing draft…");
  const published = await requestJSON(`/api/v1/workspaces/${workspaceID}/configurations/${configurationID}/versions/${versionID}/publish`, "POST", 200, {});
  versionResponse(published, configurationID, "Published", versionID);
  showVersion(published, true);
}
function draftProjection(version, configurationID) {
  versionResponse(version, configurationID, version.state);
  if (!["Draft", "Validated", "Published", "Archived"].includes(version.state)) throw new LocalError("Invalid version state.");
  const listener = version.listener;
  if (!listener || typeof listener.host !== "string" || !listener.host.trim() || listener.host.length > 255 ||
      !Number.isInteger(listener.port) || listener.port < 1 || listener.port > 65535) throw new LocalError("Invalid listener response.");
  const timeouts = {};
  for (const { key, min, max } of timeoutFields) {
    const value = listener.timeouts && listener.timeouts[key];
    if (!Number.isInteger(value) || value < min || value > max) throw new LocalError("Invalid timeout response.");
    timeouts[key] = value;
  }
  const updatedAt = typeof version.updatedAt === "string" && Number.isFinite(Date.parse(version.updatedAt)) ? version.updatedAt : null;
  return { id: version.id, configurationID, number: version.number, state: version.state, host: listener.host, port: listener.port,
    timeouts, updatedAt, createdAt: typeof version.createdAt === "string" ? version.createdAt : null };
}
function restoreEditorInputs() {
  if (!openedDraft) return;
  clearListenerErrors();
  clearTimeoutErrors();
  hostInput.value = openedDraft.host; portInput.value = String(openedDraft.port);
  for (const { key, input } of timeoutFields) input.value = String(openedDraft.timeouts[key]);
}
guardChoice(sectionChoice, () => {
  if (loading || !["listener", "timeouts"].includes(sectionChoice.value)) { restoreChoices(); return; }
  editorSection = sectionChoice.value;
  restoreEditorInputs();
  if (!saveUnknown) draftStatus("Editing section changed. No unsaved changes.");
  syncControls();
}, "Discard local Draft edits and change the editing section?");
function editorScope() {
  return { workspaceID: selectedID(workspaceChoice, workspaces), configurationID: selectedID(configurationChoice, configurations) };
}
function versionsPath(scope) { return `/api/v1/workspaces/${scope.workspaceID}/configurations/${scope.configurationID}/versions`; }
async function readVersions(scope, signal) {
  const payload = await requestJSON(versionsPath(scope), "GET", 200, undefined, signal);
  if (!Array.isArray(payload)) throw new LocalError("Invalid version list.");
  const versions = new Map();
  for (const version of payload) {
    const item = draftProjection(version, scope.configurationID);
    if (versions.has(item.id)) throw new LocalError("Duplicate version ID.");
    versions.set(item.id, item);
  }
  return versions;
}

function applyDraftList(versions, scope, selected = null) {
  drafts = new Map([...versions].filter(([, item]) => item.state === "Draft"));
  placeholder(draftChoice, "Select a Draft");
  for (const item of drafts.values()) {
    const option = document.createElement("option");
    option.value = String(item.id); show(option, message("Draft v{0} (#{1})", item.number, item.id)); draftChoice.append(option);
  }
  openedDraft = selected ? { ...selected, workspaceID: scope.workspaceID } : null;
  if (openedDraft) {
    draftChoice.value = String(openedDraft.id);
    restoreEditorInputs();
  }
  saveUnknown = false;
  publishNeedsReload = false;
  draftStatus(openedDraft ? "Draft opened. Choose listener address or timeouts to edit."
    : drafts.size ? "Select a Draft to open it." : "No Drafts in this Configuration.");
  rememberChoices();
}

openCreatedDraftButton.addEventListener("click", async () => {
  if (!createdDraft || !locked || busy || loading || pendingConfirmation) return;
  const target = createdDraft;
  busy = true;
  syncControls();
  setStatus("Opening created Draft…");
  try {
    const versions = await readVersions(target);
    const selected = versions.get(target.versionID);
    if (selected?.state !== "Draft" || selected.createdAt !== target.createdAt) {
      setStatus("The created version is no longer available as a Draft.", "error");
      return;
    }

    mode.value = "existing-draft";
    workspaceChoice.value = String(target.workspaceID);
    configurationChoice.value = String(target.configurationID);
    showModeFields();
    applyDraftList(versions, target, selected);
    clearResult();
    locked = false;
    startAnotherButton.hidden = true;
    setStatus("Draft opened without creating or publishing a version.");
  } catch (_) {
    setStatus("Could not open the created Draft. Try opening it again.", "error");
  } finally {
    busy = false;
    syncControls();
    (locked ? statusOutput : draftChoice).focus();
  }
});

async function loadDrafts(versionID) {
  invalidateLists();
  const generation = listGeneration;
  listController = new AbortController();
  const timer = setTimeout(() => generation === listGeneration && listController.abort(), 10000);
  loading = true; syncControls(); draftStatus("Loading Drafts…");
  try {
    const scope = editorScope();
    const versions = await readVersions(scope, listController.signal);
    if (generation !== listGeneration) return;
    const selected = versions.get(versionID);
    if (versionID !== undefined && selected?.state !== "Draft") {
      if (openedDraft?.id === versionID) publishNeedsReload = true;
      throw new LocalError("Selected Draft is no longer available. Local edits are retained.");
    }
    applyDraftList(versions, scope, selected);
    if (selected) {
      clearResult();
      setStatus("Draft opened without creating or publishing a version.");
    }
  } catch (error) {
    if (generation !== listGeneration) return;
    restoreChoices();
    draftStatus("Drafts unavailable or invalid. Local edits are retained. Reload to try again.");
    setStatus(error instanceof RequestError ? error.copy : "Could not open Drafts. Check the selection and reload.", "error");
  } finally {
    clearTimeout(timer);
    if (generation === listGeneration) { loading = false; syncControls(); }
  }
}
guardChoice(draftChoice, () => {
  try { return loadDrafts(selectedID(draftChoice, drafts)); }
  catch (error) { restoreChoices(); setStatus(error.copy, "error"); }
});
reloadDraftButton.addEventListener("click", async (event) => {
  if (busy || locked || loading || pendingConfirmation) return;
  if (publicationRecovery) {
    busy = true; syncControls();
    try { await refreshPublication(); }
    finally { busy = false; syncControls(); statusOutput.focus(); }
    return;
  }
  if (!openedDraft) return;
  return requestAction(event.currentTarget, () => loadDrafts(openedDraft.id),
    "Discard local Draft edits and load the current stored values from Control Service?");
});
async function saveDraft() {
  if (!openedDraft || saveUnknown || !dirty()) return;
  let scope, body;
  try {
    scope = editorScope();
    if (selectedID(draftChoice, drafts) !== openedDraft.id || scope.workspaceID !== openedDraft.workspaceID || scope.configurationID !== openedDraft.configurationID) throw new LocalError("Draft selection changed. Reload before saving.");
    body = editorSection === "timeouts" ? validateTimeoutFields() : validateListenerFields();
    if (!body) return;
  } catch (error) { setStatus(error.copy, "error"); return; }
  const section = editorSection;
  invalidateLists(); busy = true; syncControls(); setStatus(section === "timeouts" ? "Saving Draft timeouts…" : "Saving Draft listener…");
  let rejectedInput;
  try {
    const suffix = section === "timeouts" ? "/listener/timeouts" : "/listener";
    const response = await requestJSON(`${versionsPath(scope)}/${openedDraft.id}${suffix}`, "PUT", 200, body);
    const saved = draftProjection(response, scope.configurationID);
    const savedSection = section === "timeouts" ? saved.timeouts : saved;
    if (saved.id !== openedDraft.id || saved.number !== openedDraft.number || saved.state !== "Draft" ||
        Object.keys(body).some((key) => savedSection[key] !== body[key])) throw new LocalError("Unexpected saved section response.");
    // A section Save acknowledges only that section; unrelated baselines require an explicit fresh read.
    publishNeedsReload ||= section === "timeouts"
      ? saved.host !== openedDraft.host || saved.port !== openedDraft.port
      : timeoutFields.some(({ key }) => saved.timeouts[key] !== openedDraft.timeouts[key]);
    openedDraft = { ...openedDraft, updatedAt: saved.updatedAt,
      ...(section === "timeouts" ? { timeouts: saved.timeouts } : { host: saved.host, port: saved.port }) };
    drafts.set(saved.id, openedDraft); restoreEditorInputs();
    draftStatus(publishNeedsReload ? "The stored Draft changed. Reload Draft before publishing." : "Draft section saved. No unsaved changes.");
    setStatus(section === "timeouts" ? "Timeouts saved to the existing Draft. It has not been published." : "Listener saved to the existing Draft. It has not been published.", "success");
  } catch (error) {
    saveUnknown = !(error instanceof RequestError) || error.status >= 500;
    draftStatus(saveUnknown ? "Save outcome unknown. Edits retained; use Reload Draft to reconcile before saving again." : "Save failed. Unsaved Draft changes are retained.");
    rejectedInput = showListenerRejection(error);
    setStatus(rejectedInput ? listenerErrorSummary :
      (error instanceof RequestError ? error.copy : "The save outcome is unknown."), "error");
  } finally {
    busy = false; restoreChoices(); syncControls();
    rejectedInput?.focus();
  }
}
function sameDraftSelection(target) {
  return editing() && target === openedDraft && workspaceChoice.value === String(target.workspaceID) &&
    configurationChoice.value === String(target.configurationID) && draftChoice.value === String(target.id) &&
    workspaces.has(target.workspaceID) && configurations.has(target.configurationID) && drafts.has(target.id);
}

function canPublish(target) {
  return !!target?.updatedAt && target.state === "Draft" && !busy && !locked && !loading &&
    !publicationRecovery && !publishNeedsReload && !saveUnknown && !dirty() && sameDraftSelection(target);
}

publishButton.addEventListener("click", (event) => {
  const target = openedDraft;
  if (pendingConfirmation || !canPublish(target)) return;
  const name = configurations.get(target.configurationID).name;
  return confirmAction(event.currentTarget, () => publishDraft(target), message(
    "Publish Draft v{0} (#{1}) in Configuration {2} (#{3})? Any current Published version will be archived.",
    target.number, target.id, name, target.configurationID), "Publish Draft?", "Publish Draft");
});

async function refreshPublication() {
  const { target, outcome } = publicationRecovery;
  let versions;
  try { versions = await readVersions(target); }
  catch (error) {
    if (!(error instanceof RequestError) || error.status !== 404) {
      setStatus(outcome === "published" ? "Draft published. Version refresh failed. Use Refresh versions."
        : outcome === "unknown" ? "Publish outcome unknown. Use Refresh versions to check current server state."
          : "Draft unavailable. Version refresh failed. Use Refresh versions.", "error");
      return;
    }
    versions = new Map();
  }
  const current = versions.get(target.id);
  const observed = current?.number === target.number ? current : null;
  const selected = outcome !== "published" && observed?.state === "Draft" ? observed : null;
  applyDraftList(versions, target, selected);
  publicationRecovery = null;
  if (outcome === "unknown") {
    setStatus(message("Earlier publish outcome unknown. Current version state: {0}.", message(observed?.state || "Unavailable")));
  } else if (outcome === "published") {
    show(resultState, observed?.state || "Unavailable");
    setStatus(observed?.state === "Published" ? "Configuration version published. Published does not mean Running."
      : message("Draft published. Current version state: {0}. Published does not mean Running.", message(observed?.state || "Unavailable")), "success");
  } else {
    setStatus("Draft unavailable for publishing. Versions refreshed.", "error");
  }
}

async function publishDraft(target) {
  if (!canPublish(target)) return;
  busy = true; syncControls(); setStatus("Checking stored Draft…");
  try {
    let versions;
    try { versions = await readVersions(target); }
    catch (error) {
      if (error instanceof RequestError && error.status === 404 && sameDraftSelection(target) && !dirty()) {
        applyDraftList(new Map(), target);
        setStatus("Draft unavailable for publishing. Versions refreshed.", "error");
      } else setStatus("Could not check the stored Draft. Try Publish Draft again.", "error");
      return;
    }
    if (!sameDraftSelection(target) || dirty() || saveUnknown) {
      setStatus("Draft selection or edits changed. Check the Draft before publishing.", "error");
      return;
    }
    const current = versions.get(target.id);
    if (current?.state !== "Draft" || current.number !== target.number) {
      applyDraftList(versions, target);
      setStatus("Draft unavailable for publishing. Versions refreshed.", "error");
      return;
    }
    if (current.updatedAt !== target.updatedAt || current.host !== target.host || current.port !== target.port ||
        timeoutFields.some(({ key }) => current.timeouts[key] !== target.timeouts[key])) {
      publishNeedsReload = true;
      draftStatus("The stored Draft changed. Reload Draft before publishing.");
      setStatus("The stored Draft changed. Reload Draft before publishing.", "error");
      return;
    }
    clearResult();
    setStatus("Publishing draft…");
    try {
      const published = await requestJSON(`${versionsPath(target)}/${target.id}/publish`, "POST", 200, {});
      versionResponse(published, target.configurationID, "Published", target.id);
      if (published.number !== target.number) throw new LocalError("Unexpected version response or scope.");
      show($("result-heading"), "Version after publication");
      show(resultState, "Published");
      resultVersion.textContent = String(target.number);
      $("result-url-row").hidden = true;
      $("result-url-warning").hidden = true;
      result.hidden = false;
      openedDraft = null;
      publicationRecovery = { target, outcome: "published" };
    } catch (error) {
      if (error instanceof RequestError && error.status === 400) {
        setStatus("Control Service rejected publication. Check the stored Draft.", "error");
        return;
      }
      if (error instanceof RequestError && [404, 409].includes(error.status)) {
        publicationRecovery = { target, outcome: "unavailable" };
      } else {
        publicationRecovery = { target, outcome: "unknown" };
        setStatus("Publish outcome unknown. Use Refresh versions to check current server state.", "error");
        return;
      }
    }
    await refreshPublication();
  } finally {
    busy = false; restoreChoices(); syncControls(); statusOutput.focus();
  }
}

form.addEventListener("reset", (event) => {
  event.preventDefault();
  return requestAction(document.activeElement || mode, updateMode);
});
form.addEventListener("submit", async (event) => {
  event.preventDefault();
  if (pendingConfirmation || busy || locked || loading || publicationRecovery) return;
  if (editing()) return saveDraft();
  clearResult();
  const listener = mode.value === "existing-configuration" ? null : validateListenerFields();
  if (mode.value !== "existing-configuration" && !listener) return;
  let values;
  try { values = inputValues(listener); }
  catch (error) { setStatus(error.copy, "error"); return; }
  invalidateLists(); busy = true; syncControls();
  let success = false;
  try { await runFlow(values); success = true; }
  catch (error) {
    clearResult();
    showListenerRejection(error);
    const reason = error instanceof RequestError ? error.copy : "The request outcome is unknown.";
    setStatus(error instanceof RequestError && error.partialCreation
      ? message("{0} The Listener update was rejected. Studio did not publish the Draft. Inspect Dashboard before another attempt.", error.partialCreation)
      : message("{0} Inspect Dashboard before another attempt; resources from earlier steps may remain in memory.", message(reason)), "error");
  } finally {
    busy = false; locked = true; syncControls();
    show(startAnotherButton, success ? "Start another action" : "I checked Dashboard — start another action");
    startAnotherButton.hidden = false;
    if (!success) statusOutput.focus();
    else if (createdDraft) openCreatedDraftButton.focus();
  }
});
startAnotherButton.addEventListener("click", (event) => {
  if (!locked) return;
  return requestAction(event.currentTarget, () => {
    locked = false; startAnotherButton.hidden = true;
    return updateMode();
  }, "Discard local Draft edits?", true);
});
updateMode();
