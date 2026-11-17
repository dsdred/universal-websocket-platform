"use strict";
const $ = (id) => document.getElementById(id);
const form = $("configuration-form");
const mode = $("creation-mode");
const workspaceChoice = $("existing-workspace");
const configurationChoice = $("existing-configuration");
const submitButton = $("submit-button");
const statusOutput = $("flow-status");
const result = $("result");
const resultState = $("result-state");
const resultVersion = $("result-version");
const resultURL = $("result-url");
const controls = [mode, workspaceChoice, configurationChoice, $("reload-lists"),
  $("workspace-name"), $("configuration-name"), $("listener-host"), $("listener-port"), submitButton];
let busy = false, locked = false, loading = false, listGeneration = 0, listController;
let workspaces = new Map(), configurations = new Map();
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
  clearResult();
}
function syncControls() {
  for (const control of controls) control.disabled = busy || locked;
  workspaceChoice.disabled ||= loading || workspaces.size === 0;
  configurationChoice.disabled ||= loading || configurations.size === 0;
  submitButton.disabled ||= loading;
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
    if (response.status !== expectedStatus) throw new RequestError(`Control Service returned HTTP ${response.status}.`);
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
  if (busy || locked) return;
  invalidateLists();
  workspaces = new Map();
  placeholder(workspaceChoice, "Select a Workspace");
  clearConfigurations();
  const existing = mode.value === "existing-configuration";
  $("existing-resources").hidden = mode.value === "new-workspace";
  $("configuration-choice").hidden = !existing;
  $("workspace-fields").hidden = mode.value !== "new-workspace";
  $("configuration-fields").hidden = existing;
  $("listener-fields").hidden = existing;
  submitButton.textContent = existing ? "Create Draft" : "Create and publish";
  setStatus("Ready.");
  syncControls();
  if (mode.value !== "new-workspace") return loadLists();
}
mode.addEventListener("change", updateMode);
workspaceChoice.addEventListener("change", () => {
  if (busy || locked) return;
  invalidateLists(); clearConfigurations();
  setStatus("Ready."); syncControls();
  if (mode.value === "existing-configuration" && workspaceChoice.value) {
    try { return loadLists(selectedID(workspaceChoice, workspaces)); }
    catch (error) { setStatus(error.message, "error"); }
  }
});
configurationChoice.addEventListener("change", () => { if (!busy && !locked) { clearResult(); setStatus("Ready."); } });
$("reload-lists").addEventListener("click", () => {
  if (busy || locked) return;
  clearConfigurations();
  workspaces = new Map(); placeholder(workspaceChoice, "Select a Workspace");
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
form.addEventListener("submit", async (event) => {
  event.preventDefault();
  if (busy || locked || loading) return;
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
  locked = false; $("start-another").hidden = true;
  return updateMode();
});
updateMode();
