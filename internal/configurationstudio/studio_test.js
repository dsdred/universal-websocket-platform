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
  constructor() {
    this.value = ""; this.textContent = ""; this.hidden = false; this.disabled = false;
    this.listeners = {}; this.children = []; this.attributes = {};
  }
  addEventListener(event, callback) { (this.listeners[event] ||= []).push(callback); }
  setAttribute(name, value) { this.attributes[name] = value; }
  removeAttribute(name) { delete this.attributes[name]; }
  focus() { this.focused = true; }
  append(node) { this.children.push(node); }
  replaceChildren(...nodes) { this.children = nodes; this.value = ""; }
  fire(event, details = {}) {
    let result;
    for (const callback of this.listeners[event]) result = callback({ preventDefault() {}, ...details });
    return result;
  }
}
const reply = (payload, status = 200) => ({ status, json: async () => payload });
const deferred = () => { let resolve; const promise = new Promise((yes) => { resolve = yes; }); return { resolve, promise }; };
async function settle() { for (let i = 0; i < 8; i++) await new Promise(setImmediate); }
function start(route) {
  const ids = ["configuration-form", "creation-mode", "existing-workspace", "existing-configuration", "submit-button", "flow-status", "result", "result-state", "result-version", "result-url", "reload-lists", "workspace-name", "configuration-name", "listener-host", "listener-host-error", "listener-port", "listener-port-error", "list-status", "existing-resources", "configuration-choice", "workspace-fields", "configuration-fields", "listener-fields", "result-url-row", "start-another", "existing-draft", "draft-fields", "draft-status", "reload-draft", "draft-section", "timeout-fields", "timeout-handshake", "timeout-read", "timeout-write", "timeout-idle"];
  const nodes = Object.fromEntries(ids.map((id) => [id, new Element()]));
  Object.assign(nodes["creation-mode"], { value: "new-workspace" });
  nodes["workspace-name"].value = "Workspace"; nodes["configuration-name"].value = "Configuration";
  nodes["listener-host"].value = "127.0.0.1"; nodes["listener-port"].value = "8080";
  const requests = [], timers = new Map(), confirms = [], windowEvents = {}; let timerID = 0;
  const decisions = [];
  vm.runInNewContext(script, {
    window: { confirm: (message) => { confirms.push(message); return decisions.length ? decisions.shift() : true; }, addEventListener: (event, callback) => { windowEvents[event] = callback; } },
    document: { getElementById: (id) => nodes[id], createElement: () => new Element() }, AbortController,
    setTimeout: (callback) => { const id = ++timerID; timers.set(id, callback); return id; },
    clearTimeout: (id) => timers.delete(id),
    fetch: async (url, options) => { requests.push({ url, options }); return route(url, options); },
  });
  return {
    nodes, requests, timers, confirms, decisions, windowEvents,
    select: async (id, value) => { nodes[id].value = value; await nodes[id].fire("change"); },
    submit: () => nodes["configuration-form"].fire("submit"),
    again: () => nodes["start-another"].fire("click"),
    reload: () => nodes["reload-lists"].fire("click"),
  };
}
function version(state = "Draft", extra = {}) {
  return { id: 3, configurationId: 2, number: 1, state,
    ...extra, listener: { host: "127.0.0.1", port: 8080, tls: { enabled: false },
      timeouts: { handshakeSeconds: 10, readSeconds: 0, writeSeconds: 10, idleSeconds: 0 }, ...extra.listener } };
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

const versionsURL = `${base}/1/configurations/2/versions`;
function editorRoutes(url, options) {
  if (options.method === "GET" && url.endsWith("/versions")) return reply([
    version(), version("Published", { id: 4, number: 2 }), version("Archived", { id: 5, number: 3 }),
    version("Draft", { id: 6, number: 4, listener: { host: "localhost", port: 9090 } }),
  ]);
  if (options.method === "PUT") return reply(version("Draft", { listener: JSON.parse(options.body) }));
  return routes(url, options);
}
async function editor(app, open = true) {
  await app.select("creation-mode", "existing-draft");
  await app.select("existing-workspace", "1");
  await app.select("existing-configuration", "2");
  if (open) await app.select("existing-draft", "3");
}
async function editHost(app, value = "localhost") {
  app.nodes["listener-host"].value = value;
  await app.nodes["listener-host"].fire("input");
}
test("editor lists only Drafts and fresh-opens selected ID using GET without mutations", async () => {
  const app = start(editorRoutes); await editor(app);
  assert.deepEqual(app.nodes["existing-draft"].children.map((n) => n.textContent), ["Select a Draft", "Draft v1 (#3)", "Draft v4 (#6)"]);
  assert.equal(app.requests.filter((r) => r.url === versionsURL).length, 2);
  assert.equal(mutations(app).length, 0);
  assert.equal(app.nodes["listener-host"].value, "127.0.0.1");
  assert.equal(app.nodes["submit-button"].disabled, true);
});
test("save sends one pinned listener PUT with only host/port and no publish or POST", async () => {
  const app = start(editorRoutes); await editor(app); await editHost(app, " localhost "); await app.submit();
  assert.equal(mutations(app).length, 1);
  assert.equal(mutations(app)[0].url, `${versionsURL}/3/listener`);
  assert.deepEqual(JSON.parse(mutations(app)[0].options.body), { host: "localhost", port: 8080 });
  assert.match(app.nodes["flow-status"].textContent, /existing Draft/);
  assert.equal(app.nodes["submit-button"].disabled, true);
  assert.equal(app.nodes["existing-draft"].value, "3");
});
for (const [name, action] of [
  ["mode", (app) => app.select("creation-mode", "new-workspace")],
  ["Workspace", (app) => app.select("existing-workspace", "4")],
  ["Configuration", (app) => app.select("existing-configuration", "")],
  ["Draft", (app) => app.select("existing-draft", "6")],
  ["reload lists", (app) => app.reload()],
  ["reload Draft", (app) => app.nodes["reload-draft"].fire("click")],
  ["reset", (app) => app.nodes["configuration-form"].fire("reset")],
]) test(`cancel dirty ${name} navigation restores all selections and local input`, async () => {
  const app = start(editorRoutes); await editor(app); await editHost(app);
  const count = app.requests.length; app.decisions.push(false); await action(app);
  assert.equal(app.nodes["creation-mode"].value, "existing-draft");
  assert.equal(app.nodes["existing-workspace"].value, "1");
  assert.equal(app.nodes["existing-configuration"].value, "2");
  assert.equal(app.nodes["existing-draft"].value, "3");
  assert.equal(app.nodes["listener-host"].value, "localhost");
  assert.equal(app.requests.length, count);
  assert.equal(app.confirms.length, 1);
});
test("confirmed dirty parent change clears dependent Draft and disables Save", async () => {
  const app = start(editorRoutes); await editor(app); await editHost(app); await app.select("existing-workspace", "4");
  assert.equal(app.nodes["existing-draft"].value, "");
  assert.equal(app.nodes["existing-configuration"].value, "");
  assert.equal(app.nodes["submit-button"].disabled, true);
});
test("beforeunload protects dirty edits and pending saves, clean editor is unblocked", async () => {
  const pending = deferred();
  const app = start((url, options) => options.method === "PUT" ? pending.promise : editorRoutes(url, options));
  await editor(app);
  const event = () => ({ prevented: false, preventDefault() { this.prevented = true; } });
  let e = event(); app.windowEvents.beforeunload(e); assert.equal(e.prevented, false);
  await editHost(app); e = event(); app.windowEvents.beforeunload(e); assert.equal(e.prevented, true);
  const saving = app.submit(); await editHost(app, "127.0.0.1");
  e = event(); app.windowEvents.beforeunload(e); assert.equal(e.prevented, true);
  pending.resolve(reply(version("Draft", { listener: { host: "localhost", port: 8080 } }))); await saving;
});
test("late open response cannot rebind a new mode or configuration", async () => {
  const pending = deferred(); let calls = 0;
  const app = start((url, options) => url === versionsURL && ++calls === 2 ? pending.promise : editorRoutes(url, options));
  await editor(app, false); const opening = app.select("existing-draft", "3");
  await app.select("creation-mode", "new-workspace");
  pending.resolve(reply([version()])); await opening;
  assert.equal(app.nodes["creation-mode"].value, "new-workspace");
  assert.equal(app.nodes["existing-draft"].value, "");
});
test("out-of-order opens retain most recent Draft even when aborted response arrives", async () => {
  const pending = deferred(); let calls = 0;
  const app = start((url, options) => url === versionsURL && ++calls === 2 ? pending.promise : editorRoutes(url, options));
  await editor(app, false); const opening = app.select("existing-draft", "3");
  await app.select("existing-draft", "6"); pending.resolve(reply([version()])); await opening;
  assert.equal(app.nodes["existing-draft"].value, "6");
  assert.equal(app.nodes["listener-host"].value, "localhost");
});
test("duplicate Save and navigation while pending cannot redirect pinned request", async () => {
  const pending = deferred(); const app = start((url, options) => options.method === "PUT" ? pending.promise : editorRoutes(url, options));
  await editor(app); await editHost(app); const saving = app.submit(); await app.submit();
  for (const id of ["creation-mode", "existing-workspace", "existing-configuration", "existing-draft", "listener-host", "listener-port", "reload-lists", "reload-draft"]) assert.equal(app.nodes[id].disabled, true, id);
  await app.select("existing-draft", "6"); await app.select("creation-mode", "new-workspace");
  assert.equal(app.nodes["existing-draft"].value, "3");
  assert.equal(app.nodes["creation-mode"].value, "existing-draft");
  pending.resolve(reply(version("Draft", { listener: { host: "localhost", port: 8080 } }))); await saving;
  assert.equal(mutations(app).length, 1);
});
for (const status of [400, 404, 409]) test(`save HTTP ${status} preserves selected Draft and input without fallback`, async () => {
  const app = start((url, options) => options.method === "PUT" ? reply({}, status) : editorRoutes(url, options));
  await editor(app); await editHost(app); await app.submit();
  assert.equal(app.nodes["existing-draft"].value, "3"); assert.equal(app.nodes["listener-host"].value, "localhost");
  assert.match(app.nodes["draft-status"].textContent, /retained/); assert.equal(mutations(app).length, 1);
});
for (const failure of ["network", "HTTP 500", "malformed response", "wrong saved scope", "wrong saved state", "wrong saved settings"]) test(`unknown ${failure} blocks retry until confirmed fresh GET reconciliation`, async () => {
  const app = start((url, options) => {
    if (options.method !== "PUT") return editorRoutes(url, options);
    if (failure === "network") throw new TypeError("private raw error");
    if (failure === "HTTP 500") return reply({}, 500);
    if (failure === "malformed response") return { status: 200, json: async () => { throw new SyntaxError("raw private error"); } };
    return reply(version(failure === "wrong saved state" ? "Published" : "Draft", { configurationId: failure === "wrong saved scope" ? 9 : 2 }));
  });
  await editor(app); await editHost(app); await app.submit(); await app.submit();
  assert.equal(mutations(app).length, 1); assert.equal(app.nodes["submit-button"].disabled, true);
  assert.match(app.nodes["draft-status"].textContent, /outcome unknown/);
  assert.equal(app.nodes["listener-host"].value, "localhost");
  const event = { preventDefault() { this.prevented = true; } }; app.windowEvents.beforeunload(event); assert.equal(event.prevented, true);
  app.decisions.push(false); await app.nodes["reload-draft"].fire("click");
  assert.equal(app.nodes["listener-host"].value, "localhost");
  await app.nodes["reload-draft"].fire("click");
  assert.equal(app.requests.at(-1).options.method, "GET");
  assert.equal(app.nodes["listener-host"].value, "127.0.0.1");
  assert.equal(app.nodes["submit-button"].disabled, true);
  await editHost(app); assert.equal(app.nodes["submit-button"].disabled, false);
});
for (const payload of [[], [version("Published")], [version("Archived")]]) test("missing/nonDraft reconciliation retains unknown lock and local input", async () => {
  let changed = false;
  const app = start((url, options) => {
    if (options.method === "PUT") { changed = true; throw new Error("offline"); }
    if (changed && url === versionsURL) return reply(payload);
    return editorRoutes(url, options);
  });
  await editor(app); await editHost(app); await app.submit(); await app.nodes["reload-draft"].fire("click");
  assert.equal(app.nodes["existing-draft"].value, "3"); assert.equal(app.nodes["listener-host"].value, "localhost");
  assert.equal(app.nodes["submit-button"].disabled, true); assert.equal(mutations(app).length, 1);
});
for (const [name, payload] of [
  ["wrong scope", [version("Draft", { configurationId: 9 })]],
  ["unsafe ID", [version("Draft", { id: Number.MAX_SAFE_INTEGER + 1 })]],
  ["duplicate IDs", [version(), version()]],
  ["invalid listener", [version("Draft", { listener: { host: "", port: 0 } })]],
  ["invalid state", [version("<img src=x>")]],
]) test(`invalid Draft discovery ${name} never opens or saves`, async () => {
  const app = start((url, options) => url === versionsURL ? reply(payload) : editorRoutes(url, options));
  await editor(app, false); await app.submit();
  assert.equal(app.nodes["existing-draft"].children.length, 1);
  assert.equal(app.nodes["submit-button"].disabled, true); assert.equal(mutations(app).length, 0);
});
test("reload network errors preserve old Draft input and selection", async () => {
  let fail = false; const app = start((url, options) => fail && url === versionsURL ? Promise.reject(new Error("secret")) : editorRoutes(url, options));
  await editor(app); await editHost(app); fail = true; await app.nodes["reload-draft"].fire("click");
  assert.equal(app.nodes["existing-draft"].value, "3"); assert.equal(app.nodes["listener-host"].value, "localhost");
  assert.match(app.nodes["draft-status"].textContent, /retained/);
});
test("tampered Draft ID or parents do not reach PUT", async () => {
  const app = start(editorRoutes); await editor(app); await editHost(app);
  for (const value of ["6", "9007199254740992", "3x", "99"]) { app.nodes["existing-draft"].value = value; await app.submit(); }
  assert.equal(mutations(app).length, 0);
});
test("invalid host and port stay local before PUT", async () => {
  const app = start(editorRoutes); await editor(app);
  for (const port of ["0", "65536", "1.5", ""]) { app.nodes["listener-port"].value = port; await app.submit(); }
  app.nodes["listener-port"].value = "8080"; await editHost(app, " "); await app.submit();
  assert.equal(mutations(app).length, 0);
});
test("version metadata remains allowlisted and no Secret references enter rendered labels", async () => {
  const app = start((url, options) => url === versionsURL ? reply([version("Draft", { authentication: { secretRef: "sensitive-reference" }, name: "<img src=x>" })]) : editorRoutes(url, options));
  await editor(app);
  for (const node of Object.values(app.nodes)) {
    assert.doesNotMatch(node.textContent, /sensitive-reference|<img/);
    for (const child of node.children) assert.doesNotMatch(child.textContent, /sensitive-reference|<img/);
  }
});

test("unknown timeout is not automatically retried and Save remains disabled", async () => {
  const app = start((url, options) => options.method === "PUT" ? new Promise((resolve, reject) => options.signal.addEventListener("abort", () => reject(new Error("timeout")))) : editorRoutes(url, options));
  await editor(app); await editHost(app); const saving = app.submit();
  for (const callback of app.timers.values()) callback(); await saving; await app.submit();
  assert.equal(mutations(app).length, 1); assert.equal(app.nodes["submit-button"].disabled, true);
  assert.equal(app.nodes["listener-host"].value, "localhost");
});

const timeoutIDs = ["timeout-handshake", "timeout-read", "timeout-write", "timeout-idle"];
const timeoutValues = { handshakeSeconds: 25, readSeconds: 40, writeSeconds: 30, idleSeconds: 90 };
function timeoutVersion(timeouts = timeoutValues, extra = {}) {
  return version("Draft", { ...extra, listener: { ...extra.listener, timeouts } });
}
function timeoutRoutes(url, options) {
  if (options.method === "GET" && url === versionsURL) return reply([timeoutVersion()]);
  if (options.method === "PUT") return reply(timeoutVersion(JSON.parse(options.body)));
  return editorRoutes(url, options);
}
async function timeoutEditor(app) { await editor(app); await app.select("draft-section", "timeouts"); }
async function editTimeout(app, value = "45", id = "timeout-handshake") {
  app.nodes[id].value = value; await app.nodes[id].fire("input");
}

test("timeout section loads current values, exposes no references and creates nothing", async () => {
  const app = start(timeoutRoutes); await timeoutEditor(app);
  assert.deepEqual(timeoutIDs.map((id) => app.nodes[id].value), ["25", "40", "30", "90"]);
  assert.equal(app.nodes["listener-fields"].hidden, true);
  assert.equal(app.nodes["timeout-fields"].hidden, false);
  assert.equal(app.nodes["submit-button"].textContent, "Save timeouts");
  assert.equal(app.nodes["submit-button"].disabled, true);
  assert.equal(mutations(app).length, 0);
});

test("timeout Save persists one pinned section, reopens same ID without duplicate versions or publication", async () => {
  let stored = timeoutVersion(timeoutValues, { listener: { host: "example.test", tls: { enabled: true, certificateRef: "not-for-ui" } }, authentication: { enabled: true }, routing: { defaultHandlerRef: "echo" } });
  const before = structuredClone(stored);
  const app = start((url, options) => {
    if (url === versionsURL && options.method === "GET") return reply([structuredClone(stored)]);
    if (url === `${versionsURL}/3/listener/timeouts` && options.method === "PUT") {
      stored.listener.timeouts = JSON.parse(options.body); return reply(structuredClone(stored));
    }
    return editorRoutes(url, options);
  });
  await timeoutEditor(app); await editTimeout(app); await app.submit();
  assert.deepEqual(mutations(app).map(({ url, options }) => [url, options.method]), [[`${versionsURL}/3/listener/timeouts`, "PUT"]]);
  assert.deepEqual(JSON.parse(mutations(app)[0].options.body), { ...timeoutValues, handshakeSeconds: 45 });
  assert.equal(app.nodes["submit-button"].disabled, true);
  await app.nodes["reload-draft"].fire("click");
  assert.equal(app.nodes["timeout-handshake"].value, "45");
  assert.equal(app.nodes["existing-draft"].value, "3");
  const after = structuredClone(stored); after.listener.timeouts = before.listener.timeouts;
  assert.deepEqual(after, before);
  for (const node of Object.values(app.nodes)) assert.doesNotMatch(node.textContent, /not-for-ui/);
});

for (const [id, invalid] of [
  ["timeout-handshake", "0"], ["timeout-handshake", "301"],
  ["timeout-write", "0"], ["timeout-write", "301"],
  ["timeout-read", "-1"], ["timeout-read", "86401"],
  ["timeout-idle", "-1"], ["timeout-idle", "86401"],
  ["timeout-read", ""], ["timeout-read", " "], ["timeout-idle", "1.5"], ["timeout-handshake", "1e2"],
]) test(`invalid ${id}=${JSON.stringify(invalid)} retains input without PUT`, async () => {
  const app = start(timeoutRoutes); await timeoutEditor(app); await editTimeout(app, invalid, id); await app.submit();
  assert.equal(mutations(app).length, 0); assert.equal(app.nodes[id].value, invalid);
  assert.match(app.nodes["flow-status"].textContent, /timeout must be a whole number/);
});

for (const values of [
  { handshakeSeconds: 1, readSeconds: 0, writeSeconds: 1, idleSeconds: 0 },
  { handshakeSeconds: 300, readSeconds: 86400, writeSeconds: 300, idleSeconds: 86400 },
]) test("timeout endpoints accept all four boundary values including explicit zero", async () => {
  const app = start(timeoutRoutes); await timeoutEditor(app);
  for (const [index, value] of Object.values(values).entries()) await editTimeout(app, String(value), timeoutIDs[index]);
  await app.submit(); assert.deepEqual(JSON.parse(mutations(app)[0].options.body), values);
  assert.equal(app.nodes["submit-button"].disabled, true);
});

for (const [name, action] of [
  ["section", (app) => app.select("draft-section", "listener")],
  ["mode", (app) => app.select("creation-mode", "new-workspace")],
  ["Workspace", (app) => app.select("existing-workspace", "4")],
  ["Configuration", (app) => app.select("existing-configuration", "")],
  ["Draft", (app) => app.select("existing-draft", "6")],
  ["reload", (app) => app.nodes["reload-draft"].fire("click")],
  ["reset", (app) => app.nodes["configuration-form"].fire("reset")],
]) test(`Cancel dirty timeout ${name} retains section, selections and values`, async () => {
  const app = start(timeoutRoutes); await timeoutEditor(app); await editTimeout(app);
  const count = app.requests.length; app.decisions.push(false); await action(app);
  assert.equal(app.nodes["draft-section"].value, "timeouts"); assert.equal(app.nodes["timeout-handshake"].value, "45");
  assert.equal(app.nodes["existing-draft"].value, "3"); assert.equal(app.requests.length, count);
  const event = { preventDefault() { this.prevented = true; } }; app.windowEvents.beforeunload(event); assert.equal(event.prevented, true);
});

test("confirmed section switch discards only local edits and keeps address baseline", async () => {
  const app = start(timeoutRoutes); await timeoutEditor(app); await editTimeout(app);
  await app.select("draft-section", "listener");
  assert.equal(app.nodes["listener-host"].value, "127.0.0.1");
  assert.equal(app.nodes["submit-button"].textContent, "Save listener");
  await app.select("draft-section", "timeouts");
  assert.equal(app.nodes["timeout-handshake"].value, "25"); assert.equal(mutations(app).length, 0);
});

test("address dirty section Cancel preserves TASK136 address input", async () => {
  const app = start(timeoutRoutes); await editor(app); await editHost(app);
  app.decisions.push(false); await app.select("draft-section", "timeouts");
  assert.equal(app.nodes["draft-section"].value, "listener"); assert.equal(app.nodes["listener-host"].value, "localhost");
  assert.equal(app.nodes["listener-fields"].hidden, false);
});

test("duplicate timeout Save and pending section changes cannot redirect request", async () => {
  const pending = deferred(); const app = start((url, options) => options.method === "PUT" ? pending.promise : timeoutRoutes(url, options));
  await timeoutEditor(app); await editTimeout(app); const saving = app.submit(); await app.submit();
  for (const id of ["draft-section", ...timeoutIDs]) assert.equal(app.nodes[id].disabled, true);
  await app.select("draft-section", "listener"); assert.equal(app.nodes["draft-section"].value, "timeouts");
  pending.resolve(reply(timeoutVersion({ ...timeoutValues, handshakeSeconds: 45 }))); await saving;
  assert.equal(mutations(app).length, 1); assert.equal(app.nodes["submit-button"].disabled, true);
});

for (const status of [400, 404, 409]) test(`timeout HTTP ${status} retains values and same Draft, no create fallback`, async () => {
  const app = start((url, options) => options.method === "PUT" ? reply({}, status) : timeoutRoutes(url, options));
  await timeoutEditor(app); await editTimeout(app); await app.submit();
  assert.equal(app.nodes["timeout-handshake"].value, "45"); assert.equal(app.nodes["existing-draft"].value, "3");
  assert.match(app.nodes["draft-status"].textContent, /retained/); assert.equal(mutations(app).length, 1);
});

for (const failure of ["network", "500", "invalid JSON", "ID", "scope", "number", "Published", "timeout values"]) test(`timeout unknown ${failure} prevents retry until confirmed fresh read`, async () => {
  const app = start((url, options) => {
    if (options.method !== "PUT") return timeoutRoutes(url, options);
    if (failure === "network") throw new Error("private internal message");
    if (failure === "500") return reply({}, 500);
    if (failure === "invalid JSON") return { status: 200, json: async () => { throw new Error("bad JSON"); } };
    return reply(timeoutVersion(failure === "timeout values" ? timeoutValues : { ...timeoutValues, handshakeSeconds: 45 },
      { id: failure === "ID" ? 6 : 3, configurationId: failure === "scope" ? 9 : 2,
        number: failure === "number" ? 2 : 1, state: failure === "Published" ? "Published" : "Draft" }));
  });
  await timeoutEditor(app); await editTimeout(app); await app.submit(); await app.submit();
  assert.equal(mutations(app).length, 1); assert.equal(app.nodes["timeout-handshake"].value, "45");
  assert.equal(app.nodes["submit-button"].disabled, true); assert.match(app.nodes["draft-status"].textContent, /outcome unknown/);
  app.decisions.push(false); await app.nodes["reload-draft"].fire("click"); assert.equal(app.nodes["timeout-handshake"].value, "45");
  await app.nodes["reload-draft"].fire("click"); assert.equal(app.nodes["timeout-handshake"].value, "25");
  await editTimeout(app); assert.equal(app.nodes["submit-button"].disabled, false);
});

test("unknown timeout result stays locked across section switch and failed reload", async () => {
  let failed = false;
  const app = start((url, options) => {
    if (options.method === "PUT") { failed = true; throw new Error("offline"); }
    if (failed && url === versionsURL) throw new Error("offline");
    return timeoutRoutes(url, options);
  });
  await timeoutEditor(app); await editTimeout(app); await app.submit();
  await app.nodes["reload-draft"].fire("click"); assert.equal(app.nodes["timeout-handshake"].value, "45");
  await app.select("draft-section", "listener"); await editHost(app); await app.submit();
  assert.equal(mutations(app).length, 1); assert.equal(app.nodes["submit-button"].disabled, true);
});

test("late timeout reload cannot overwrite another mode or its creation controls", async () => {
  const pending = deferred(); let calls = 0;
  const app = start((url, options) => url === versionsURL && ++calls === 3 ? pending.promise : timeoutRoutes(url, options));
  await timeoutEditor(app); const loading = app.nodes["reload-draft"].fire("click");
  await app.select("creation-mode", "new-workspace"); pending.resolve(reply([timeoutVersion({ ...timeoutValues, handshakeSeconds: 99 })])); await loading;
  assert.equal(app.nodes["creation-mode"].value, "new-workspace"); assert.equal(app.nodes["timeout-fields"].hidden, true);
  assert.equal(app.nodes["submit-button"].textContent, "Create and publish");
});

test("timeout Save does not adopt unrelated address values from response", async () => {
  const app = start((url, options) => options.method === "PUT"
    ? reply(timeoutVersion(JSON.parse(options.body), { listener: { host: "external.example", port: 9999 } })) : timeoutRoutes(url, options));
  await timeoutEditor(app); await editTimeout(app); await app.submit(); await app.select("draft-section", "listener");
  assert.equal(app.nodes["listener-host"].value, "127.0.0.1"); assert.equal(app.nodes["listener-port"].value, "8080");
});

test("address Save does not adopt unrelated timeout values from response", async () => {
  const app = start((url, options) => options.method === "PUT"
    ? reply(timeoutVersion({ ...timeoutValues, idleSeconds: 888 }, { listener: JSON.parse(options.body) })) : timeoutRoutes(url, options));
  await editor(app); await editHost(app); await app.submit(); await app.select("draft-section", "timeouts");
  assert.equal(app.nodes["timeout-idle"].value, "90");
});

for (const timeouts of [null, {}, { ...timeoutValues, readSeconds: -1 }, { ...timeoutValues, idleSeconds: 86401 }, { ...timeoutValues, writeSeconds: 1.5 }]) test("invalid timeout GET data cannot open Draft or enable Save", async () => {
  const app = start((url, options) => url === versionsURL ? reply([timeoutVersion(timeouts)]) : timeoutRoutes(url, options));
  await editor(app, false); await app.submit(); assert.equal(mutations(app).length, 0); assert.equal(app.nodes["submit-button"].disabled, true);
  assert.equal(app.nodes["existing-draft"].children.length, 1);
});

const invalidListenerHosts = [
  "127.0.0.1.4.5.6.7.8.9.0", "192.168.1.999", "127.0.0.1.", "1..2.3", "127.00.0.1",
  "http://localhost", "localhost:8080", "[::1]", "fe80::1%eth0", "::ffff:192.0.02.128", "-bad.host", "bad-.host",
  "2001:::1", "1::2::3", "1:2:3:4:5:6::192.0.2.128", "1:2:3:4:5:6:7", "1:2:3:4:5:6:7:",
  "foo\n.bar", "\ufefflocalhost", "ä.example", "a".repeat(64) + ".example", "a".repeat(256), "", " \t ",
];
const invalidListenerPorts = ["8777777777777777777777", "0", "000", "65536", "99999", "-1", "+1", "1.5", "1e3", "0x50", " 80 ", "80\n", "", "８０"];

async function listenerInput(app, field, value, event = "input") {
  const input = app.nodes[`listener-${field}`];
  input.value = value;
  await input.fire(event, { inputType: "insertFromPaste" });
  return input;
}

for (const [flow, setup, route] of [
  ["new workspace", async () => {}, routes],
  ["existing workspace", async (app) => {
    await app.select("creation-mode", "new-configuration");
    await app.select("existing-workspace", "1");
  }, routes],
  ["existing Draft", editor, editorRoutes],
]) test(`${flow} rejects invalid Listener host and port before any mutation`, async () => {
  const app = start(route);
  await setup(app);
  for (const [field, invalid, valid] of [
    ["host", invalidListenerHosts, "localhost"],
    ["port", invalidListenerPorts, "8081"],
  ]) {
    for (const value of invalid) {
      const input = await listenerInput(app, field, value);
      const error = app.nodes[`listener-${field}-error`];
      assert.notEqual(error.textContent, "", JSON.stringify(value));
      assert.equal(input.attributes["aria-invalid"], "true");
      await input.fire("blur");
      assert.notEqual(error.textContent, "", JSON.stringify(value));
      input.focused = false;
      await app.submit();
      assert.equal(input.focused, true);
      assert.equal(input.value, value);
      assert.equal(mutations(app).length, 0);
      await listenerInput(app, field, valid);
      assert.equal(error.textContent, "");
      assert.equal(input.attributes["aria-invalid"], undefined);
      assert.doesNotMatch(app.nodes["flow-status"].textContent, /highlighted Listener fields/);
    }
  }
});

test("blur validates a field and submit focuses host before port when both are invalid", async () => {
  const app = start(routes);
  const host = await listenerInput(app, "host", "", "blur");
  const port = await listenerInput(app, "port", "", "blur");
  assert.match(app.nodes["listener-host-error"].textContent, /host/);
  assert.match(app.nodes["listener-port-error"].textContent, /port/);
  await app.submit();
  assert.equal(host.focused, true);
  assert.equal(port.focused, undefined);
  assert.equal(mutations(app).length, 0);
});

for (const [host, rawPort, port, canonicalHost = host] of [
  ["127.0.0.1", "1", 1], ["0.0.0.0", "65535", 65535], ["::", "8081", 8081],
  ["2001:DB8::1", "8081", 8081], ["::ffff:192.0.2.128", "8081", 8081],
  ["1:2:3:4:5:6:192.0.2.128", "8081", 8081], ["1:2:3:4:5:6:7:8", "8081", 8081],
  ["localhost", "0".repeat(30) + "8081", 8081], ["App-1.example.", "8081", 8081],
  ["label.123", "8081", 8081], ["123", "8081", 8081], ["\u0085 localhost \u0085", "8081", 8081, "localhost"],
]) test(`same Draft accepts valid Listener address ${JSON.stringify(host)} and port ${port}`, async () => {
  const app = start(editorRoutes);
  await editor(app);
  await listenerInput(app, "host", host);
  await listenerInput(app, "port", rawPort);
  await app.submit();
  const sent = mutations(app);
  assert.equal(sent.length, 1);
  assert.equal(sent[0].url, `${versionsURL}/3/listener`);
  assert.equal(sent[0].options.method, "PUT");
  assert.deepEqual(JSON.parse(sent[0].options.body), { host: canonicalHost, port });
  assert.match(app.nodes["flow-status"].textContent, /saved to the existing Draft/);
});

test("configuration creation accepts IPv6 and keeps its original publish sequence", async () => {
  let listener;
  const app = start((url, options) => {
    if (url.endsWith("/listener")) {
      listener = JSON.parse(options.body);
      return reply(version("Draft", { listener }));
    }
    if (url.endsWith("/publish")) return reply(version("Published", { listener }));
    return routes(url, options);
  });
  await listenerInput(app, "host", "2001:db8::1");
  await listenerInput(app, "port", "65535");
  await app.submit();
  assert.deepEqual(listener, { host: "2001:db8::1", port: 65535 });
  assert.deepEqual(mutations(app).map(({ options }) => options.method), ["POST", "POST", "POST", "PUT", "POST"]);
  assert.equal(app.nodes["result-url"].textContent, "ws://[2001:db8::1]:65535/ws");
});

function fieldRejection(field, message) {
  return reply({ error: { code: "validation_failed", message: `${field}: ${message}` } }, 400);
}

test("Draft API field rejection survives blur and preserves input until correction", async () => {
  const app = start((url, options) => options.method === "PUT" && JSON.parse(options.body).host === "localhost"
    ? fieldRejection("host", "must be a valid IP address or hostname") : editorRoutes(url, options));
  await editor(app);
  const host = await listenerInput(app, "host", "localhost");
  await app.submit();
  assert.equal(host.value, "localhost");
  assert.equal(host.focused, true);
  assert.equal(host.disabled, false);
  assert.equal(host.attributes["aria-invalid"], "true");
  assert.equal(app.nodes["listener-host-error"].textContent, "must be a valid IP address or hostname");
  assert.match(app.nodes["draft-status"].textContent, /changes are retained/);
  await host.fire("blur");
  await app.submit();
  assert.equal(mutations(app).length, 1);
  assert.equal(app.nodes["listener-host-error"].textContent, "must be a valid IP address or hostname");
  await listenerInput(app, "host", "example.test");
  assert.equal(app.nodes["listener-host-error"].textContent, "");
  await app.submit();
  assert.deepEqual(mutations(app).map((entry) => entry.url), [`${versionsURL}/3/listener`, `${versionsURL}/3/listener`]);
  assert.equal(app.nodes["existing-draft"].value, "3");
});

for (const reuseWorkspace of [false, true]) test(`creation reports confirmed partial writes after Listener validation rejection (${reuseWorkspace ? "existing" : "new"} Workspace)`, async () => {
  const app = start((url, options) => url.endsWith("/listener")
    ? fieldRejection("port", "must be between 1 and 65535") : routes(url, options));
  if (reuseWorkspace) {
    await app.select("creation-mode", "new-configuration");
    await app.select("existing-workspace", "1");
  }
  const port = await listenerInput(app, "port", "8081");
  await app.submit();
  assert.equal(port.value, "8081");
  assert.equal(app.nodes["listener-port-error"].textContent, "must be between 1 and 65535");
  assert.equal(app.nodes["listener-host-error"].textContent, "");
  assert.equal(app.nodes["flow-status"].focused, true);
  assert.match(app.nodes["flow-status"].textContent, reuseWorkspace
    ? /Created Configuration #2 in Workspace #1 and Draft #3/ : /Created Workspace #1, Configuration #2 and Draft #3/);
  assert.match(app.nodes["flow-status"].textContent, /Studio did not publish the Draft.*Inspect Dashboard/);
  assert.deepEqual(mutations(app).map(({ options }) => options.method), reuseWorkspace ? ["POST", "POST", "PUT"] : ["POST", "POST", "POST", "PUT"]);
  await app.submit();
  assert.equal(mutations(app).length, reuseWorkspace ? 3 : 4);
  await app.again();
  assert.equal(app.nodes["listener-port-error"].textContent, "");
});

test("timeout validation rejection does not project Listener field errors", async () => {
  const app = start((url, options) => url.endsWith("/listener/timeouts")
    ? fieldRejection("port", "not a Listener address reply") : timeoutRoutes(url, options));
  await timeoutEditor(app);
  await editTimeout(app);
  await app.submit();
  assert.match(app.nodes["flow-status"].textContent, /HTTP 400/);
  assert.equal(app.nodes["listener-port-error"].textContent, "");
});

test("invalid Listener message survives cancelled navigation and clears after confirmed section change", async () => {
  const app = start(editorRoutes);
  await editor(app);
  await listenerInput(app, "host", "127.0.0.1.");
  app.decisions.push(false);
  await app.select("draft-section", "timeouts");
  assert.equal(app.nodes["listener-host"].value, "127.0.0.1.");
  assert.match(app.nodes["listener-host-error"].textContent, /IPv4/);
  await app.select("draft-section", "timeouts");
  assert.equal(app.nodes["listener-host"].value, "127.0.0.1");
  assert.equal(app.nodes["listener-host-error"].textContent, "");
});

test("legacy dotted numeric host stays readable while only address writes require correction", async () => {
  const legacyHost = "127.0.0.1.4.5.6.7.8.9.0";
  const app = start((url, options) => {
    if (options.method === "GET" && url === versionsURL) return reply([
      timeoutVersion(timeoutValues, { listener: { host: legacyHost } }),
    ]);
    if (options.method === "PUT") return reply(
      timeoutVersion(JSON.parse(options.body), { listener: { host: legacyHost } }),
    );
    return timeoutRoutes(url, options);
  });
  await editor(app);
  assert.equal(app.nodes["listener-host"].value, legacyHost);
  await listenerInput(app, "port", "8081");
  await app.submit();
  assert.match(app.nodes["listener-host-error"].textContent, /IPv4/);
  assert.equal(mutations(app).length, 0);
  await app.select("draft-section", "timeouts");
  await editTimeout(app);
  await app.submit();
  const [saved] = mutations(app);
  assert.equal(mutations(app).length, 1);
  assert.equal(saved.url, `${versionsURL}/3/listener/timeouts`);
  assert.deepEqual(JSON.parse(saved.options.body), { ...timeoutValues, handshakeSeconds: 45 });
  assert.match(app.nodes["flow-status"].textContent, /Timeouts saved to the existing Draft/);
});
