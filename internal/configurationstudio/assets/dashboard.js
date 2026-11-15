"use strict";
(() => {
  const $ = (id) => document.getElementById(id);
  const workspace = $("workspace"), configuration = $("configuration");
  const base = "/api/v1/workspaces";
  let generation = 0, controller, healthController;
  class ViewError extends Error {}
  function requireShape(condition) {
    if (!condition) throw new ViewError("Unsupported response. No data from this response is shown.");
  }
  const object = (v) => v !== null && typeof v === "object" && !Array.isArray(v);
  const id = (v) => Number.isSafeInteger(v) && v > 0;
  function entities(data, parentKey, parentID) {
    requireShape(Array.isArray(data));
    const seen = new Set();
    for (const item of data) {
      requireShape(object(item) && id(item.id) && typeof item.name === "string" && item.name.length > 0);
      requireShape(!seen.has(item.id) && (!parentKey || (id(item[parentKey]) && String(item[parentKey]) === parentID)));
      seen.add(item.id);
    }
    return data;
  }
  function timestamp(value) {
    return typeof value === "string" && Number.isFinite(Date.parse(value)) ? new Date(value).toLocaleString() : "Unavailable";
  }
  function option(select, text, value = "") {
    const node = document.createElement("option");
    node.textContent = text; node.value = value; select.append(node);
  }
  function clearSelect(select, message) {
    select.replaceChildren(); option(select, message); select.disabled = true;
  }
  function fill(select, items, previous, label) {
    select.replaceChildren(); option(select, items.length ? `Select ${label}` : `No ${label}s`);
    for (const item of items) option(select, `${item.name} · ID ${item.id}`, String(item.id));
    select.disabled = items.length === 0;
    select.value = items.some((item) => String(item.id) === previous) ? previous : "";
  }
  function versionsMessage(message) {
    $("versions").replaceChildren();
    const row = document.createElement("tr"), cell = document.createElement("td");
    cell.colSpan = 4; cell.textContent = message; row.append(cell); $("versions").append(row);
    $("version-count").textContent = "Not loaded";
  }
  function clearVersions() { versionsMessage("Select a configuration to view its versions."); }
  function clearConfigurations() {
    clearSelect(configuration, "Select a workspace");
    $("configuration-meta").textContent = "Not loaded"; clearVersions();
  }
  async function get(path, signal) {
    let response;
    try {
      response = await fetch(path, { method: "GET", mode: "same-origin", credentials: "same-origin", cache: "no-store", redirect: "error", signal });
    } catch (error) {
      if (signal.aborted) throw error;
      throw new ViewError("Offline / unreachable. Check the local Control Service and use Refresh.");
    }
    if (!response.ok) throw new ViewError(`Request error (HTTP ${response.status}). Use Refresh to try again.`);
    requireShape((response.headers.get("Content-Type") || "").split(";")[0].trim() === "application/json");
    try { return await response.json(); }
    catch (_) { throw new ViewError("Unsupported response. Expected valid JSON."); }
  }
  async function health() {
    if (healthController) healthController.abort();
    const cycle = new AbortController(), signal = cycle.signal;
    healthController = cycle;
    const deadline = setTimeout(() => cycle.abort(), 15000);
    $("health-status").textContent = "Loading…";
    $("health-time").textContent = "Checking now…";
    try {
      const data = await get("/health", signal);
      requireShape(object(data) && data.status === "ok");
      if (!signal.aborted) $("health-status").textContent = "Available";
    } catch (error) {
      if (!signal.aborted) $("health-status").textContent = error.message;
    }
    clearTimeout(deadline);
    if (healthController === cycle) {
      if (signal.aborted) $("health-status").textContent = "Not confirmed — request timed out";
      $("health-time").textContent = `Checked in this browser: ${new Date().toLocaleString()}`;
    }
  }
  async function loadConfigurations(signal, previous) {
    const parent = workspace.value;
    if (!parent) return;
    clearSelect(configuration, "Loading…");
    const items = entities(await get(`${base}/${parent}/configurations`, signal), "workspaceId", parent);
    if (signal.aborted) return;
    fill(configuration, items, previous, "configuration");
    $("configuration-meta").textContent = `${items.length} configurations in selected workspace`;
  }
  async function loadVersions(signal) {
    const parent = configuration.value;
    if (!parent) return;
    versionsMessage("Loading versions…");
    const data = await get(`${base}/${workspace.value}/configurations/${parent}/versions`, signal);
    requireShape(Array.isArray(data));
    const seen = new Set();
    for (const v of data) {
      requireShape(object(v) && id(v.id) && id(v.configurationId) && String(v.configurationId) === parent);
      requireShape(!seen.has(v.id) && Number.isInteger(v.number) && v.number > 0 && v.number <= 4294967295);
      requireShape(["Draft", "Published", "Archived"].includes(v.state));
      requireShape(v.listener === undefined || v.listener === null || (object(v.listener) && typeof v.listener.host === "string" && Number.isInteger(v.listener.port) && v.listener.port >= 0 && v.listener.port <= 65535));
      seen.add(v.id);
    }
    if (signal.aborted) return;
    $("versions").replaceChildren();
    if (!data.length) versionsMessage("No versions in this configuration.");
    for (const v of data) {
      const row = document.createElement("tr");
      for (const text of [`v${v.number} · ID ${v.id}`, v.state, v.listener && v.listener.host && v.listener.port ? `${v.listener.host} / ${v.listener.port}` : "Unavailable", timestamp(v.updatedAt)]) {
        const cell = document.createElement("td"); cell.textContent = text; row.append(cell);
      }
      row.children[1].className = "badge";
      $("versions").append(row);
    }
    $("version-count").textContent = `${data.length} versions in selected configuration`;
  }
  async function run(level) {
    if (controller) controller.abort();
    controller = new AbortController();
    const signal = controller.signal, current = ++generation;
    const oldWorkspace = workspace.value, oldConfiguration = configuration.value;
    // A single request-cycle deadline, never a polling/retry timer.
    const controllerForCycle = controller;
    const deadline = setTimeout(() => controllerForCycle.abort(), 15000);
    $("browser").setAttribute("aria-busy", "true");
    $("refresh").disabled = true;
    $("data-status").className = "status";
    $("data-status").textContent = "Loading configuration data…";
    let healthRequest = Promise.resolve();
    try {
      if (level === "all") {
        clearSelect(workspace, "Loading…"); $("workspace-meta").textContent = "Not loaded";
        clearConfigurations(); healthRequest = health();
        const items = entities(await get(base, signal));
        if (signal.aborted) return;
        fill(workspace, items, oldWorkspace, "workspace");
        $("workspace-meta").textContent = `${items.length} workspaces · fetched in this browser: ${new Date().toLocaleString()}`;
      }
      if (level !== "versions") {
        clearConfigurations(); await loadConfigurations(signal, level === "all" ? oldConfiguration : "");
      } else clearVersions();
      await loadVersions(signal);
      if (!signal.aborted) $("data-status").textContent = workspace.options.length === 1 ? "No workspaces. Create one in Configuration Studio." : `Data loaded in this browser: ${new Date().toLocaleString()}. Select a workspace and configuration to browse.`;
    } catch (error) {
      if (current === generation && !signal.aborted) {
        $("data-status").textContent = error.message;
        $("data-status").className = "status error";
        if (configuration.disabled && workspace.value) clearSelect(configuration, "Not loaded — use Refresh");
        if (workspace.disabled) clearSelect(workspace, "Not loaded — use Refresh");
        versionsMessage("Data unavailable. Use Refresh to try again.");
      }
    } finally {
      clearTimeout(deadline);
      await healthRequest;
      if (current === generation) {
        if (signal.aborted) {
          $("data-status").textContent = "Request timed out. Use Refresh to try again.";
          if (workspace.disabled) clearSelect(workspace, "Not loaded — use Refresh");
          if (configuration.disabled) clearSelect(configuration, "Not loaded — use Refresh");
          versionsMessage("Data unavailable. Use Refresh to try again.");
        }
        $("browser").setAttribute("aria-busy", "false"); $("refresh").disabled = false;
      }
    }
  }
  workspace.addEventListener("change", () => run("configurations"));
  configuration.addEventListener("change", () => run("versions"));
  $("refresh").addEventListener("click", () => run("all"));
  run("all");
})();
