"use strict";
const form = document.getElementById("configuration-form");
const submitButton = document.getElementById("submit-button");
const statusOutput = document.getElementById("flow-status");
const result = document.getElementById("result");
const resultState = document.getElementById("result-state");
const resultVersion = document.getElementById("result-version");
const resultURL = document.getElementById("result-url");
class DisplayedFlowError extends Error {}
function setStatus(message, kind) {
  statusOutput.textContent = message;
  statusOutput.className = `status ${kind || ""}`.trim();
}
function isPositiveSafeInteger(value) {
  return Number.isSafeInteger(value) && value > 0;
}
function requiredID(value) {
  if (!isPositiveSafeInteger(value)) {
    throw new Error("Unexpected response from the Control Service.");
  }
  return value;
}
function requiredVersionNumber(value) {
  if (!Number.isSafeInteger(value) || value < 1) {
    throw new Error("Unexpected response from the Control Service.");
  }
  return value;
}
function backendError(payload) {
  if (
    payload !== null &&
    typeof payload === "object" &&
    payload.error !== null &&
    typeof payload.error === "object" &&
    typeof payload.error.message === "string" &&
    payload.error.message.length > 0 &&
    typeof payload.error.code === "string" &&
    payload.error.code.length > 0
  ) {
    return `${payload.error.message} (${payload.error.code})`;
  }
  return null;
}
async function requestJSON(path, method, expectedStatus, body) {
  const options = {
    method,
    headers: { "Content-Type": "application/json" },
  };
  if (body !== undefined) {
    options.body = JSON.stringify(body);
  }
  const response = await fetch(path, options);
  let payload;
  try {
    payload = await response.json();
  } catch (_) {
    throw new Error("Unexpected response from the Control Service.");
  }
  if (response.status !== expectedStatus) {
    const safeError = backendError(payload);
    if (safeError !== null) {
      setStatus(`${safeError}. The flow stopped; resources created by earlier steps may remain in memory.`, "error");
      throw new DisplayedFlowError();
    }
    throw new Error("Unexpected response from the Control Service.");
  }
  if (payload === null || typeof payload !== "object" || Array.isArray(payload)) {
    throw new Error("Unexpected response from the Control Service.");
  }
  return payload;
}
function inputValues() {
  const workspaceName = document.getElementById("workspace-name").value.trim();
  const configurationName = document.getElementById("configuration-name").value.trim();
  const listenerHost = document.getElementById("listener-host").value.trim();
  const portValue = document.getElementById("listener-port").value;
  const listenerPort = Number(portValue);
  if (workspaceName === "" || configurationName === "" || listenerHost === "" || portValue === "") {
    throw new Error("Complete every field before creating the configuration.");
  }
  if (!Number.isInteger(listenerPort) || listenerPort < 1 || listenerPort > 65535) {
    throw new Error("Listener port must be a whole number between 1 and 65535.");
  }
  return { workspaceName, configurationName, listenerHost, listenerPort };
}
function preliminaryURL(host, port) {
  const renderedHost = host.includes(":") ? `[${host}]` : host;
  return `ws://${renderedHost}:${port}/ws`;
}
function showPublished(version) {
  requiredID(version.id);
  requiredID(version.configurationId);
  const number = requiredVersionNumber(version.number);
  if (version.state !== "Published") {
    throw new Error("Unexpected response from the Control Service.");
  }
  if (version.listener === null || typeof version.listener !== "object") {
    throw new Error("Unexpected response from the Control Service.");
  }
  const host = version.listener.host;
  const port = version.listener.port;
  if (typeof host !== "string" || host.length === 0 || !Number.isInteger(port) || port < 1 || port > 65535) {
    throw new Error("Unexpected response from the Control Service.");
  }
  if (
    version.listener.tls === null ||
    typeof version.listener.tls !== "object" ||
    version.listener.tls.enabled !== false
  ) {
    throw new Error("Unexpected response from the Control Service.");
  }
  resultState.textContent = version.state;
  resultVersion.textContent = String(number);
  resultURL.textContent = preliminaryURL(host, port);
  result.hidden = false;
  setStatus("Configuration version published. Published does not mean Running.", "success");
}
async function runFlow(values) {
  setStatus("Creating workspace…", "");
  const workspace = await requestJSON("/api/v1/workspaces", "POST", 201, {
    name: values.workspaceName,
    description: "",
  });
  const workspaceID = requiredID(workspace.id);
  setStatus("Creating configuration…", "");
  const configuration = await requestJSON(
    `/api/v1/workspaces/${workspaceID}/configurations`,
    "POST",
    201,
    { name: values.configurationName, description: "" },
  );
  const configurationID = requiredID(configuration.id);
  if (requiredID(configuration.workspaceId) !== workspaceID) {
    throw new Error("Unexpected response from the Control Service.");
  }
  setStatus("Creating draft version…", "");
  const draft = await requestJSON(
    `/api/v1/workspaces/${workspaceID}/configurations/${configurationID}/versions`,
    "POST",
    201,
    {},
  );
  const versionID = requiredID(draft.id);
  requiredVersionNumber(draft.number);
  if (
    requiredID(draft.configurationId) !== configurationID ||
    draft.state !== "Draft"
  ) {
    throw new Error("Unexpected response from the Control Service.");
  }
  setStatus("Updating listener…", "");
  const updated = await requestJSON(
    `/api/v1/workspaces/${workspaceID}/configurations/${configurationID}/versions/${versionID}/listener`,
    "PUT",
    200,
    { host: values.listenerHost, port: values.listenerPort },
  );
  if (
    requiredID(updated.id) !== versionID ||
    requiredID(updated.configurationId) !== configurationID ||
    updated.state !== "Draft" ||
    updated.listener === null ||
    typeof updated.listener !== "object" ||
    updated.listener.host !== values.listenerHost ||
    updated.listener.port !== values.listenerPort ||
    updated.listener.tls === null ||
    typeof updated.listener.tls !== "object" ||
    updated.listener.tls.enabled !== false
  ) {
    throw new Error("Unexpected response from the Control Service.");
  }
  setStatus("Publishing draft…", "");
  const published = await requestJSON(
    `/api/v1/workspaces/${workspaceID}/configurations/${configurationID}/versions/${versionID}/publish`,
    "POST",
    200,
    {},
  );
  if (requiredID(published.id) !== versionID || requiredID(published.configurationId) !== configurationID) {
    throw new Error("Unexpected response from the Control Service.");
  }
  showPublished(published);
}
form.addEventListener("submit", async (event) => {
  event.preventDefault();
  result.hidden = true;
  let values;
  try {
    values = inputValues();
  } catch (error) {
    setStatus(error.message, "error");
    return;
  }
  submitButton.disabled = true;
  try {
    await runFlow(values);
  } catch (error) {
    if (!(error instanceof DisplayedFlowError)) {
      if (error instanceof TypeError) {
        setStatus("The request outcome is unknown. Inspect Control Service state before starting another attempt; resources from earlier successful steps may remain in memory.", "error");
      } else {
        const message = error.message || "Unexpected response from the Control Service.";
        setStatus(`${message} The flow stopped; resources from earlier successful steps may remain in memory.`, "error");
      }
    }
  }
});
