"use strict";
// Run: node --test internal/configurationstudio/studio_test.js (no npm).
const test = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const script = fs.readFileSync(path.join(__dirname, "assets/studio.js"), "utf8");
const base = "/api/v1/workspaces";
class Element {
  constructor() { this.value = ""; this.textContent = ""; this.hidden = false; this.disabled = false; this.listeners = {}; this.children = []; }
  addEventListener(event, callback) { this.listeners[event] = callback; }
  append(node) { this.children.push(node); }
  replaceChildren(...nodes) { this.children = nodes; this.value = ""; }
  fire(event) { return this.listeners[event]({ preventDefault() {} }); }
}
const reply = (payload, status = 200) => ({ status, json: async () => payload });
const deferred = () => { let resolve; const promise = new Promise((yes) => { resolve = yes; }); return { resolve, promise }; };
async function settle() { for (let i = 0; i < 8; i++) await new Promise(setImmediate); }
function start(route) {
  const ids = ["configuration-form", "creation-mode", "existing-workspace", "existing-configuration", "submit-button", "flow-status", "result", "result-state", "result-version", "result-url", "reload-lists", "workspace-name", "configuration-name", "listener-host", "listener-port", "list-status", "existing-resources", "configuration-choice", "workspace-fields", "configuration-fields", "listener-fields", "result-url-row", "start-another"];
  const nodes = Object.fromEntries(ids.map((id) => [id, new Element()]));
  Object.assign(nodes["creation-mode"], { value: "new-workspace" });
  nodes["workspace-name"].value = "Workspace"; nodes["configuration-name"].value = "Configuration";
  nodes["listener-host"].value = "127.0.0.1"; nodes["listener-port"].value = "8080";
  const requests = [], timers = new Map(); let timerID = 0;
  vm.runInNewContext(script, {
    document: { getElementById: (id) => nodes[id], createElement: () => new Element() }, AbortController,
    setTimeout: (callback) => { const id = ++timerID; timers.set(id, callback); return id; },
    clearTimeout: (id) => timers.delete(id),
    fetch: async (url, options) => { requests.push({ url, options }); return route(url, options); },
  });
  return {
    nodes, requests, timers,
    select: async (id, value) => { nodes[id].value = value; await nodes[id].fire("change"); },
    submit: () => nodes["configuration-form"].fire("submit"),
    again: () => nodes["start-another"].fire("click"),
    reload: () => nodes["reload-lists"].fire("click"),
  };
}
function version(state = "Draft", extra = {}) {
  return { id: 3, configurationId: 2, number: 1, state,
    listener: { host: "127.0.0.1", port: 8080, tls: { enabled: false } }, ...extra };
}
function routes(url, options) {
  if (options.method === "GET") {
    if (url === base) return reply([{ id: 1, name: "Workspace" }, { id: 4, name: "Other" }]);
    return reply([{ id: 2, workspaceId: Number(url.split("/")[4]), name: "Configuration" }]);
  }
  if (url === base) return reply({ id: 1 }, 201);
  if (url.endsWith("/configurations")) return reply({ id: 2, workspaceId: 1 }, 201);
  if (url.endsWith("/versions")) return reply(version(), 201);
  return reply(version(url.endsWith("/publish") ? "Published" : "Draft"));
}
async function existing(app) {
  await app.select("creation-mode", "existing-configuration");
  await app.select("existing-workspace", "1");
  await app.select("existing-configuration", "2");
}
const mutations = (app) => app.requests.filter(({ options }) => options.method !== "GET");
test("original new parent flow remains complete and same-origin", async () => {
  const app = start(routes); await app.submit();
  assert.deepEqual(mutations(app).map(({ options }) => options.method), ["POST", "POST", "POST", "PUT", "POST"]);
  assert.equal(app.nodes["result-state"].textContent, "Published");
  for (const { options } of app.requests) {
    assert.equal(options.mode, "same-origin"); assert.equal(options.redirect, "error"); assert.equal(options.cache, "no-store");
  }
});
test("existing Configuration creates only Draft without parent mutations or publish", async () => {
  const app = start(routes); await existing(app); await app.submit();
  assert.deepEqual(mutations(app).map(({ url }) => url), [`${base}/1/configurations/2/versions`]);
  assert.equal(app.nodes["result-state"].textContent, "Draft");
  assert.equal(app.nodes["result-url-row"].hidden, true);
});
test("new Configuration uses selected Workspace and original publish steps", async () => {
  const app = start(routes); await app.select("creation-mode", "new-configuration");
  await app.select("existing-workspace", "1"); await app.submit();
  assert.equal(mutations(app).length, 4);
  assert.equal(mutations(app)[0].url, `${base}/1/configurations`);
  assert.equal(app.nodes["result-state"].textContent, "Published");
});
test("Workspace changes clear Configuration and stale result", async () => {
  const app = start(routes); await existing(app);
  app.nodes.result.hidden = false;
  await app.select("existing-workspace", "4");
  assert.equal(app.nodes["existing-configuration"].value, "");
  assert.equal(app.nodes.result.hidden, true);
  await app.submit(); assert.equal(mutations(app).length, 0);
});
test("Configuration changes clear prior result", async () => {
  const app = start(routes); await existing(app); app.nodes.result.hidden = false;
  await app.select("existing-configuration", "");
  assert.equal(app.nodes.result.hidden, true); await app.submit(); assert.equal(mutations(app).length, 0);
});
test("late list response cannot overwrite another mode", async () => {
  const old = deferred(); const app = start((url, options) => url === base && options.method === "GET" ? old.promise : routes(url, options));
  const pending = app.select("creation-mode", "existing-configuration");
  await app.select("creation-mode", "new-workspace");
  old.resolve(reply([{ id: 99, name: "Old" }])); await pending;
  assert.equal(app.nodes["existing-workspace"].children.length, 1);
  assert.equal(app.nodes["existing-resources"].hidden, true);
});
test("out-of-order Configuration list cannot replace newer Workspace scope", async () => {
  const old = deferred(); const app = start((url, options) => url === `${base}/1/configurations` ? old.promise : routes(url, options));
  await app.select("creation-mode", "existing-configuration");
  const pending = app.select("existing-workspace", "1");
  await app.select("existing-workspace", "4");
  old.resolve(reply([{ id: 2, workspaceId: 1, name: "Old" }])); await pending;
  assert.equal(app.nodes["existing-configuration"].children[1].textContent, "Configuration (#2)");
  await app.select("existing-configuration", "2"); await app.submit();
  assert.equal(mutations(app)[0].url, `${base}/4/configurations/2/versions`);
});
for (const [name, payload] of [
  ["unsafe ID", [{ id: Number.MAX_SAFE_INTEGER + 1, name: "bad" }]],
  ["zero ID", [{ id: 0, name: "bad" }]],
  ["duplicate ID", [{ id: 1, name: "a" }, { id: 1, name: "b" }]],
  ["wrong payload", {}],
]) test(`rejects Workspace list ${name}`, async () => {
  const app = start(() => reply(payload)); await app.select("creation-mode", "existing-configuration");
  assert.match(app.nodes["list-status"].textContent, /unavailable or invalid/);
  assert.equal(app.nodes["existing-workspace"].disabled, true);
  await app.submit(); assert.equal(mutations(app).length, 0);
});
test("wrong Configuration workspaceId is rejected", async () => {
  const app = start((url, options) => url.endsWith("/configurations") ? reply([{ id: 2, workspaceId: 9, name: "wrong" }]) : routes(url, options));
  await app.select("creation-mode", "existing-configuration"); await app.select("existing-workspace", "1");
  assert.equal(app.nodes["existing-configuration"].disabled, true);
  await app.submit(); assert.equal(mutations(app).length, 0);
});
test("empty Workspace and Configuration lists explain next action", async () => {
  const empty = start(() => reply([])); await empty.select("creation-mode", "existing-configuration");
  assert.match(empty.nodes["list-status"].textContent, /No Workspaces/);
  const app = start((url, options) => url.endsWith("/configurations") ? reply([]) : routes(url, options));
  await app.select("creation-mode", "existing-configuration"); await app.select("existing-workspace", "1");
  assert.match(app.nodes["list-status"].textContent, /No Configurations/);
});
test("failed list is reloadable without mutations", async () => {
  let fail = true; const app = start((url, options) => fail ? Promise.reject(new TypeError("offline")) : routes(url, options));
  await app.select("creation-mode", "existing-configuration");
  assert.match(app.nodes["list-status"].textContent, /unavailable/);
  fail = false; await app.reload();
  assert.equal(app.nodes["existing-workspace"].children.length, 3);
  assert.equal(mutations(app).length, 0);
});
test("tampered selected IDs never reach mutation routes", async () => {
  const app = start(routes); await existing(app);
  for (const bad of ["2x", "9007199254740992", "99", "0"]) {
    app.nodes["existing-configuration"].value = bad; await app.submit();
  }
  assert.equal(mutations(app).length, 0);
});
test("double submit and selection during mutation cannot duplicate or redirect action", async () => {
  const pending = deferred(); const app = start((url, options) => options.method === "POST" ? pending.promise : routes(url, options));
  await existing(app); const first = app.submit();
  await app.submit(); await app.select("existing-workspace", "4");
  assert.equal(mutations(app).length, 1);
  assert.equal(app.nodes["creation-mode"].disabled, true);
  pending.resolve(reply(version(), 201)); await first; await app.submit();
  assert.equal(mutations(app).length, 1);
});
for (const status of [404, 500]) test(`Draft HTTP ${status} never creates parents or claims success`, async () => {
  const app = start((url, options) => options.method === "POST" ? reply({}, status) : routes(url, options));
  await existing(app); await app.submit();
  assert.equal(app.nodes.result.hidden, true); assert.equal(mutations(app).length, 1);
  assert.equal(app.nodes["submit-button"].disabled, true);
  assert.match(app.nodes["flow-status"].textContent, /Inspect Dashboard/);
});
test("unknown mutation outcome requires explicit acknowledgement and refreshed selection", async () => {
  const app = start((url, options) => options.method === "POST" ? Promise.reject(new TypeError("offline")) : routes(url, options));
  await existing(app); await app.submit(); await app.submit();
  assert.equal(mutations(app).length, 1);
  assert.match(app.nodes["start-another"].textContent, /I checked Dashboard/);
  await app.again();
  assert.equal(app.nodes["existing-workspace"].value, "");
  await app.submit(); assert.equal(mutations(app).length, 1);
});
test("malformed Draft scope cannot become success", async () => {
  const app = start((url, options) => options.method === "POST" ? reply(version("Draft", { configurationId: 9 }), 201) : routes(url, options));
  await existing(app); await app.submit();
  assert.equal(app.nodes.result.hidden, true); assert.equal(mutations(app).length, 1);
});
test("resource labels use text, never HTML", async () => {
  const app = start(() => reply([{ id: 1, name: "<img src=x onerror=alert(1)>" }]));
  await app.select("creation-mode", "existing-configuration");
  assert.equal(app.nodes["existing-workspace"].children[1].textContent, "<img src=x onerror=alert(1)> (#1)");
});
test("bounded GET timeout yields recoverable list failure", async () => {
  const app = start((_, options) => new Promise((resolve, reject) => options.signal.addEventListener("abort", () => reject(new Error("aborted")))));
  const pending = app.select("creation-mode", "existing-configuration");
  for (const callback of app.timers.values()) callback(); await pending;
  assert.match(app.nodes["list-status"].textContent, /unavailable/);
  assert.equal(app.nodes["reload-lists"].disabled, false);
});

test("Reload lists refreshes Workspaces and clears deleted parent selection", async () => {
  let removed = false;
  const app = start((url, options) => removed && url === base ? reply([]) : routes(url, options));
  await existing(app); removed = true; await app.reload();
  assert.equal(app.nodes["existing-workspace"].value, "");
  assert.equal(app.nodes["existing-configuration"].value, "");
  assert.match(app.nodes["list-status"].textContent, /No Workspaces/);
  await app.submit(); assert.equal(mutations(app).length, 0);
});
