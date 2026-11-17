"use strict";
const $ = (id) => document.getElementById(id);
const form = $("configuration-form");
const mode = $("creation-mode");
const workspaceChoice = $("existing-workspace");
const configurationChoice = $("existing-configuration");
const draftChoice = $("existing-draft");
const sectionChoice = $("draft-section");
const timeoutFields = [
  ["handshakeSeconds", $("timeout-handshake"), 1, 300, "Handshake"],
  ["readSeconds", $("timeout-read"), 0, 86400, "Read"],
  ["writeSeconds", $("timeout-write"), 1, 300, "Write"],
  ["idleSeconds", $("timeout-idle"), 0, 86400, "Idle"],
];
const submitButton = $("submit-button");
const statusOutput = $("flow-status");
const result = $("result");
const resultState = $("result-state");
const resultVersion = $("result-version");
const resultURL = $("result-url");
const controls = [mode, workspaceChoice, configurationChoice, $("reload-lists"),
  $("workspace-name"), $("configuration-name"), $("listener-host"), $("listener-port"), sectionChoice,
  ...timeoutFields.map(([, input]) => input), submitButton];
let busy = false, locked = false, loading = false, listGeneration = 0, listController;
let workspaces = new Map(), configurations = new Map(), drafts = new Map();
let openedDraft = null, saveUnknown = false;
let editorSection = "listener";
let choices = { mode: mode.value, workspace: "", configuration: "", draft: "" };
const editing = () => mode.value === "existing-draft";
const hostInput = $("listener-host"), portInput = $("listener-port");
function dirty() {
  if (openedDraft && editorSection === "timeouts") return timeoutFields.some(([key, input]) => input.value !== String(openedDraft.timeouts[key]));
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
function allowNavigation(message = "Discard local Draft edits and leave this Draft?") {
  if (busy || locked) { restoreChoices(); return false; }
  if ((dirty() || saveUnknown) && !window.confirm(message + (saveUnknown ? " The previous save outcome is unknown; stored values must be reloaded before another save." : ""))) {
    restoreChoices(); return false;
  }
  return true;
}
function draftStatus(message) { $("draft-status").textContent = message; }
function clearDrafts() {
  drafts = new Map(); openedDraft = null; saveUnknown = false;
  editorSection = "listener"; sectionChoice.value = editorSection;
  placeholder(draftChoice, "Select a Draft");
  draftStatus("Select a Configuration to discover Drafts.");
}
function updateDirty() {
  if (editing() && openedDraft && !saveUnknown) draftStatus(dirty() ? "Unsaved Draft changes." : "Draft section is unchanged.");
  syncControls();
}
for (const input of [hostInput, portInput, ...timeoutFields.map(([, input]) => input)]) input.addEventListener("input", updateDirty);
window.addEventListener("beforeunload", (event) => {
  if (busy || dirty() || saveUnknown) { event.preventDefault(); event.returnValue = ""; }
});
class RequestError extends Error {}
function setStatus(message, kind = "") {
  statusOutput.textContent = message;
  statusOutput.className = `status ${kind}`.trim();
}
function requiredID(value) {
  if (!Number.isSafeInteger(value) || value < 1) throw new Error("Invalid resource ID in Control Service response.");
  return value;
}
function selectedID(select, items) {
  if (!/^[1-9][0-9]*$/.test(select.value)) throw new Error("Select an available resource first.");
  const id = requiredID(Number(select.value));
  if (!items.has(id)) throw new Error("Select an available resource first.");
  return id;
}
function clearResult() {
  result.hidden = true;
  resultState.textContent = resultVersion.textContent = resultURL.textContent = "";
}
function placeholder(select, label) {
  const option = document.createElement("option");
  option.value = ""; option.textContent = label;
  select.replaceChildren(option);
}
function clearConfigurations() {
  configurations = new Map();
  placeholder(configurationChoice, "Select a Configuration");
  clearDrafts();
  clearResult();
}
function syncControls() {
  for (const control of controls) control.disabled = busy || locked;
  workspaceChoice.disabled ||= loading || workspaces.size === 0;
  configurationChoice.disabled ||= loading || configurations.size === 0;
  submitButton.disabled ||= loading;
  draftChoice.disabled = busy || locked || loading || drafts.size === 0;
  $("reload-draft").disabled = busy || locked || loading || !openedDraft;
  sectionChoice.disabled ||= loading || !openedDraft;
  $("timeout-fields").hidden = !editing() || editorSection !== "timeouts";
  $("listener-fields").hidden = editing() ? editorSection !== "listener" : mode.value === "existing-configuration";
  for (const [, input] of timeoutFields) input.disabled ||= !editing() || !openedDraft || loading;
  if (editing()) {
    submitButton.textContent = editorSection === "timeouts" ? "Save timeouts" : "Save listener";
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
    if (response.status !== expectedStatus) throw Object.assign(new RequestError(`Control Service returned HTTP ${response.status}.`), { status: response.status });
    const payload = await response.json();
    if (payload === null || typeof payload !== "object") throw new Error("Unexpected Control Service response.");
    return payload;
  } finally {
    if (timer !== null) clearTimeout(timer);
  }
}
function resourceList(payload, workspaceID) {
  if (!Array.isArray(payload)) throw new Error("Invalid resource list.");
  const items = new Map();
  for (const item of payload) {
    if (!item || typeof item !== "object" || typeof item.name !== "string" || !item.name.trim()) throw new Error("Invalid resource list.");
    const id = requiredID(item.id);
    if (items.has(id) || (workspaceID !== undefined && requiredID(item.workspaceId) !== workspaceID)) throw new Error("Invalid resource scope.");
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
  $("list-status").textContent = "Loading resources…";
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
    $("list-status").textContent = items.size ? "Select a resource to continue." : workspaceID === undefined
      ? "No Workspaces. Choose New Workspace and Configuration to create one."
      : "No Configurations. Choose New Configuration in existing Workspace to create one.";
  } catch (_) {
    if (generation !== listGeneration) return;
    $("list-status").textContent = "Resources unavailable or invalid. Reload lists to try again.";
  } finally {
    clearTimeout(timer);
    if (generation === listGeneration) { loading = false; syncControls(); }
  }
}
function updateMode() {
  if (!allowNavigation()) return;
  invalidateLists();
  workspaces = new Map();
  placeholder(workspaceChoice, "Select a Workspace");
  clearConfigurations();
  const existing = mode.value === "existing-configuration" || editing();
  $("existing-resources").hidden = mode.value === "new-workspace";
  $("configuration-choice").hidden = !existing;
  $("workspace-fields").hidden = mode.value !== "new-workspace";
  $("configuration-fields").hidden = existing;
  $("listener-fields").hidden = existing && !editing();
  $("draft-fields").hidden = !editing();
  submitButton.textContent = editing() ? "Save listener" : existing ? "Create Draft" : "Create and publish";
  rememberChoices();
  setStatus("Ready.");
  syncControls();
  if (mode.value !== "new-workspace") return loadLists();
}
mode.addEventListener("change", updateMode);
workspaceChoice.addEventListener("change", () => {
  if (!allowNavigation()) return;
  invalidateLists(); clearConfigurations();
  rememberChoices();
  setStatus("Ready."); syncControls();
  if ((mode.value === "existing-configuration" || editing()) && workspaceChoice.value) {
    try { return loadLists(selectedID(workspaceChoice, workspaces)); }
    catch (error) { setStatus(error.message, "error"); }
  }
});
configurationChoice.addEventListener("change", () => {
  if (!allowNavigation()) return;
  invalidateLists(); clearDrafts(); clearResult(); rememberChoices(); setStatus("Ready."); syncControls();
  if (editing() && configurationChoice.value) return loadDrafts();
});
$("reload-lists").addEventListener("click", () => {
  if (!allowNavigation()) return;
  clearConfigurations();
  workspaces = new Map(); placeholder(workspaceChoice, "Select a Workspace");
  rememberChoices();
  return loadLists();
});
function inputValues() {
  if (!["new-workspace", "new-configuration", "existing-configuration"].includes(mode.value)) throw new Error("Select an action.");
  const values = { mode: mode.value };
  if (mode.value !== "new-workspace") values.workspaceID = selectedID(workspaceChoice, workspaces);
  if (mode.value === "existing-configuration") {
    values.configurationID = selectedID(configurationChoice, configurations);
    return values;
  }
  values.workspaceName = $("workspace-name").value.trim();
  values.configurationName = $("configuration-name").value.trim();
  values.listenerHost = $("listener-host").value.trim();
  values.listenerPort = Number($("listener-port").value);
  if ((mode.value === "new-workspace" && !values.workspaceName) || !values.configurationName || !values.listenerHost) throw new Error("Complete every visible field.");
  if (!Number.isInteger(values.listenerPort) || values.listenerPort < 1 || values.listenerPort > 65535) throw new Error("Listener port must be a whole number between 1 and 65535.");
  return values;
}
function versionResponse(version, configurationID, state, versionID) {
  if (Array.isArray(version) || requiredID(version.configurationId) !== configurationID || version.state !== state ||
      !Number.isSafeInteger(version.number) || version.number < 1 || (versionID !== undefined && requiredID(version.id) !== versionID)) throw new Error("Unexpected version response or scope.");
  requiredID(version.id);
}
function preliminaryURL(host, port) {
  const renderedHost = host.includes(":") ? `[${host}]` : host;
  return `ws://${renderedHost}:${port}/ws`;
}
function showVersion(version, published) {
  resultState.textContent = version.state;
  resultVersion.textContent = String(version.number);
  $("result-url-row").hidden = !published;
  if (published) {
    const listener = version.listener;
    if (!listener || typeof listener.host !== "string" || !listener.host || !Number.isInteger(listener.port) || listener.port < 1 || listener.port > 65535 || !listener.tls || listener.tls.enabled !== false) throw new Error("Unexpected listener response.");
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
    if (requiredID(configuration.workspaceId) !== workspaceID) throw new Error("Unexpected configuration scope.");
  }
  setStatus("Creating draft version…");
  const draft = await requestJSON(`/api/v1/workspaces/${workspaceID}/configurations/${configurationID}/versions`, "POST", 201, {});
  versionResponse(draft, configurationID, "Draft");
  if (values.mode === "existing-configuration") { showVersion(draft, false); return; }
  const versionID = draft.id;
  setStatus("Updating listener…");
  const updated = await requestJSON(`/api/v1/workspaces/${workspaceID}/configurations/${configurationID}/versions/${versionID}/listener`, "PUT", 200,
    { host: values.listenerHost, port: values.listenerPort });
  versionResponse(updated, configurationID, "Draft", versionID);
  if (!updated.listener || updated.listener.host !== values.listenerHost || updated.listener.port !== values.listenerPort || !updated.listener.tls || updated.listener.tls.enabled !== false) throw new Error("Unexpected listener response.");
  setStatus("Publishing draft…");
  const published = await requestJSON(`/api/v1/workspaces/${workspaceID}/configurations/${configurationID}/versions/${versionID}/publish`, "POST", 200, {});
  versionResponse(published, configurationID, "Published", versionID);
  showVersion(published, true);
}
function draftProjection(version, configurationID) {
  versionResponse(version, configurationID, version.state);
  if (!["Draft", "Validated", "Published", "Archived"].includes(version.state)) throw new Error("Invalid version state.");
  const listener = version.listener;
  if (!listener || typeof listener.host !== "string" || !listener.host.trim() || listener.host.length > 255 ||
      !Number.isInteger(listener.port) || listener.port < 1 || listener.port > 65535) throw new Error("Invalid listener response.");
  const timeouts = {};
  for (const [key, , min, max] of timeoutFields) {
    const value = listener.timeouts && listener.timeouts[key];
    if (!Number.isInteger(value) || value < min || value > max) throw new Error("Invalid timeout response.");
    timeouts[key] = value;
  }
  return { id: version.id, configurationID, number: version.number, state: version.state, host: listener.host, port: listener.port, timeouts };
}
function restoreEditorInputs() {
  if (!openedDraft) return;
  hostInput.value = openedDraft.host; portInput.value = String(openedDraft.port);
  for (const [key, input] of timeoutFields) input.value = String(openedDraft.timeouts[key]);
}
sectionChoice.addEventListener("change", () => {
  if (loading || !allowNavigation("Discard local Draft edits and change the editing section?")) { restoreChoices(); return; }
  if (!["listener", "timeouts"].includes(sectionChoice.value)) { restoreChoices(); return; }
  editorSection = sectionChoice.value;
  restoreEditorInputs();
  if (!saveUnknown) draftStatus("Editing section changed. No unsaved changes.");
  syncControls();
});
function editorScope() {
  return { workspaceID: selectedID(workspaceChoice, workspaces), configurationID: selectedID(configurationChoice, configurations) };
}
function versionsPath(scope) { return `/api/v1/workspaces/${scope.workspaceID}/configurations/${scope.configurationID}/versions`; }
async function loadDrafts(versionID) {
  invalidateLists();
  const generation = listGeneration;
  listController = new AbortController();
  const timer = setTimeout(() => generation === listGeneration && listController.abort(), 10000);
  loading = true; syncControls(); draftStatus("Loading Drafts…");
  try {
    const scope = editorScope();
    const payload = await requestJSON(versionsPath(scope), "GET", 200, undefined, listController.signal);
    if (generation !== listGeneration) return;
    if (!Array.isArray(payload)) throw new Error("Invalid version list.");
    const items = new Map(), seen = new Set();
    for (const version of payload) {
      const item = draftProjection(version, scope.configurationID);
      if (seen.has(item.id)) throw new Error("Duplicate version ID.");
      seen.add(item.id);
      if (item.state === "Draft") items.set(item.id, item);
    }
    const selected = versionID === undefined ? null : items.get(versionID);
    if (versionID !== undefined && !selected) throw new Error("Selected Draft is no longer available. Local edits are retained.");
    drafts = items;
    placeholder(draftChoice, "Select a Draft");
    for (const item of items.values()) {
      const option = document.createElement("option");
      option.value = String(item.id); option.textContent = `Draft v${item.number} (#${item.id})`; draftChoice.append(option);
    }
    if (selected) {
      openedDraft = { ...selected, workspaceID: scope.workspaceID };
      draftChoice.value = String(selected.id);
      restoreEditorInputs();
      saveUnknown = false; clearResult();
      draftStatus("Draft opened. Choose listener address or timeouts to edit.");
      setStatus("Draft opened without creating or publishing a version.");
    } else draftStatus(items.size ? "Select a Draft to open it." : "No Drafts in this Configuration.");
    rememberChoices();
  } catch (error) {
    if (generation !== listGeneration) return;
    restoreChoices();
    draftStatus("Drafts unavailable or invalid. Local edits are retained. Reload to try again.");
    setStatus(error instanceof RequestError ? error.message : "Could not open Drafts. Check the selection and reload.", "error");
  } finally {
    clearTimeout(timer);
    if (generation === listGeneration) { loading = false; syncControls(); }
  }
}
draftChoice.addEventListener("change", () => {
  if (!allowNavigation()) return;
  try { return loadDrafts(selectedID(draftChoice, drafts)); }
  catch (error) { restoreChoices(); setStatus(error.message, "error"); }
});
$("reload-draft").addEventListener("click", () => {
  if (!openedDraft || !allowNavigation("Discard local Draft edits and load the current stored values from Control Service?")) return;
  return loadDrafts(openedDraft.id);
});
async function saveDraft() {
  if (!openedDraft || saveUnknown || !dirty()) return;
  let scope, body;
  try {
    scope = editorScope();
    if (selectedID(draftChoice, drafts) !== openedDraft.id || scope.workspaceID !== openedDraft.workspaceID || scope.configurationID !== openedDraft.configurationID) throw new Error("Draft selection changed. Reload before saving.");
    if (editorSection === "timeouts") {
      body = {};
      for (const [key, input, min, max, label] of timeoutFields) {
        const value = Number(input.value);
        if (!/^[0-9]+$/.test(input.value) || !Number.isInteger(value) || value < min || value > max) throw new Error(`${label} timeout must be a whole number between ${min} and ${max} seconds.`);
        body[key] = value;
      }
    } else {
      const host = hostInput.value.trim(), port = Number(portInput.value);
      if (!host || host.length > 255 || !Number.isInteger(port) || port < 1 || port > 65535) throw new Error("Enter a listener host and whole port between 1 and 65535.");
      body = { host, port };
    }
  } catch (error) { setStatus(error.message, "error"); return; }
  const section = editorSection;
  invalidateLists(); busy = true; syncControls(); setStatus(`Saving Draft ${section}…`);
  try {
    const suffix = section === "timeouts" ? "/listener/timeouts" : "/listener";
    const response = await requestJSON(`${versionsPath(scope)}/${openedDraft.id}${suffix}`, "PUT", 200, body);
    const saved = draftProjection(response, scope.configurationID);
    const savedSection = section === "timeouts" ? saved.timeouts : saved;
    if (saved.id !== openedDraft.id || saved.number !== openedDraft.number || saved.state !== "Draft" ||
        Object.keys(body).some((key) => savedSection[key] !== body[key])) throw new Error("Unexpected saved section response.");
    // A section Save acknowledges only that section; unrelated baselines require an explicit fresh read.
    openedDraft = { ...openedDraft, ...(section === "timeouts" ? { timeouts: saved.timeouts } : { host: saved.host, port: saved.port }) };
    drafts.set(saved.id, openedDraft); restoreEditorInputs();
    draftStatus("Draft section saved. No unsaved changes."); setStatus(`${section === "timeouts" ? "Timeouts" : "Listener"} saved to the existing Draft. It has not been published.`, "success");
  } catch (error) {
    saveUnknown = !(error instanceof RequestError) || error.status >= 500;
    draftStatus(saveUnknown ? "Save outcome unknown. Edits retained; use Reload Draft to reconcile before saving again." : "Save failed. Unsaved Draft changes are retained.");
    setStatus(error instanceof RequestError ? error.message : "The save outcome is unknown.", "error");
  } finally { busy = false; restoreChoices(); syncControls(); }
}
form.addEventListener("reset", (event) => {
  event.preventDefault();
  if (!allowNavigation()) return;
  clearDrafts();
  return updateMode();
});
form.addEventListener("submit", async (event) => {
  event.preventDefault();
  if (busy || locked || loading) return;
  if (editing()) return saveDraft();
  clearResult();
  let values;
  try { values = inputValues(); }
  catch (error) { setStatus(error.message, "error"); return; }
  invalidateLists(); busy = true; syncControls();
  let success = false;
  try { await runFlow(values); success = true; }
  catch (error) {
    clearResult();
    const message = error instanceof RequestError ? error.message : "The request outcome is unknown.";
    setStatus(`${message} Inspect Dashboard before another attempt; resources from earlier steps may remain in memory.`, "error");
  } finally {
    busy = false; locked = true; syncControls();
    $("start-another").textContent = success ? "Start another action" : "I checked Dashboard — start another action";
    $("start-another").hidden = false;
  }
});
$("start-another").addEventListener("click", () => {
  if (busy || !locked) return;
  if ((dirty() || saveUnknown) && !window.confirm("Discard local Draft edits?")) return;
  clearDrafts();
  locked = false; $("start-another").hidden = true;
  return updateMode();
});
updateMode();
