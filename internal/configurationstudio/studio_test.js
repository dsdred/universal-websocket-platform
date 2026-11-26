"use strict";
// Run: node --test internal/configurationstudio/studio_test.js (no npm).
const test = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const script = fs.readFileSync(path.join(__dirname, "assets/studio.js"), "utf8");
const languageScript = fs.readFileSync(path.join(__dirname, "assets/control-language.js"), "utf8");
const base = "/api/v1/workspaces";
const timeoutIDs = ["timeout-handshake", "timeout-read", "timeout-write", "timeout-idle"];
class Element {
  constructor() {
    this.value = ""; this.textContent = ""; this.hidden = false; this.disabled = false;
    this.listeners = {}; this.children = []; this.attributes = {};
  }
  addEventListener(event, callback) { (this.listeners[event] ||= []).push(callback); }
  setAttribute(name, value) { this.attributes[name] = value; }
  removeAttribute(name) { delete this.attributes[name]; }
  focus() { this.focused = true; }
  getClientRects() { return this.hidden ? [] : [{}]; }
  append(node) { this.children.push(node); }
  replaceChildren(...nodes) { this.children = nodes; this.value = ""; }
  fire(event, details = {}) {
    let result;
    for (const callback of this.listeners[event]) result = callback({ currentTarget: this, preventDefault() { this.defaultPrevented = true; }, ...details });
    return result;
  }
}
const reply = (payload, status = 200) => ({ status, json: async () => payload });
const deferred = () => { let resolve; const promise = new Promise((yes) => { resolve = yes; }); return { resolve, promise }; };
async function settle() { for (let i = 0; i < 8; i++) await new Promise(setImmediate); }
function start(route, storage = new Map()) {
  const ids = ["studio-confirmation", "confirmation-cancel", "confirmation-accept", "confirmation-title", "confirmation-message", "confirmation-warning", "navigation-toggle", "control-language", "control-language-code", "configuration-form", "creation-mode", "existing-workspace", "existing-configuration", "submit-button", "flow-status", "result", "result-heading", "result-url-warning", "result-state", "result-version", "result-url", "reload-lists", "workspace-name", "configuration-name", "listener-host", "listener-host-error", "listener-port", "listener-port-error", "list-status", "existing-resources", "configuration-choice", "workspace-fields", "configuration-fields", "listener-fields", "result-url-row", "start-another", "existing-draft", "draft-fields", "draft-status", "reload-draft", "publish-draft", "draft-section", "timeout-fields", "timeout-handshake", "timeout-read", "timeout-write", "timeout-idle"];
  ids.push(...timeoutIDs.map((id) => `${id}-error`));
  ids.push("open-created-draft");
  const nodes = Object.fromEntries(ids.map((id) => [id, Object.assign(new Element(), { id })]));
  nodes["open-created-draft"].textContent = "Open created Draft";
  nodes["open-created-draft"].setAttribute("data-i18n", "");
  Object.assign(nodes["creation-mode"], { value: "new-workspace" });
  nodes["workspace-name"].value = "Workspace"; nodes["configuration-name"].value = "Configuration";
  nodes["listener-host"].value = "127.0.0.1"; nodes["listener-port"].value = "8080";
  const requests = [], timers = new Map(), confirms = [], windowEvents = {}; let timerID = 0;
  const decisions = [], documentEvents = {}, navigationAttempts = [];
  const navigation = { fail: false };
  const confirmation = nodes["studio-confirmation"];
  confirmation.open = false;
  nodes["navigation-toggle"].hidden = true;
  confirmation.showModal = () => {
    assert.equal(confirmation.open, false);
    confirmation.open = true;
    confirms.push(nodes["confirmation-message"].textContent);
    if (navigation.autoConfirm !== false) {
      const choice = decisions.length ? decisions.shift() : true;
      queueMicrotask(() => nodes[choice ? "confirmation-accept" : "confirmation-cancel"].fire("click"));
    }
  };
  confirmation.close = () => { confirmation.open = false; };
  const environment = vm.createContext({
    URL,
    window: {
      location: { href: "http://localhost/", origin: "http://localhost", pathname: "/", search: "", assign(url) {
        navigationAttempts.push(url);
        if (navigation.fail) throw new Error("navigation failed");
      } },
      addEventListener: (event, callback) => { windowEvents[event] = callback; },
    },
    document: {
      documentElement: {},
      addEventListener: (event, callback) => { documentEvents[event] = callback; },
      getElementById: (id) => nodes[id],
      createElement: () => new Element(),
      querySelectorAll(selector) {
        const visit = (node) => [node, ...node.children.flatMap(visit)];
        return Object.values(nodes).flatMap(visit).filter((node) => Object.hasOwn(node.attributes, selector.slice(1, -1)));
      },
    },
    localStorage: { getItem: (key) => storage.get(key), setItem: (key, value) => storage.set(key, value) },
    AbortController,
    setTimeout: (callback) => { const id = ++timerID; timers.set(id, callback); return id; },
    clearTimeout: (id) => timers.delete(id),
    fetch: async (url, options) => { requests.push({ url, options }); return route(url, options); },
  });
  vm.runInContext(languageScript, environment, { filename: "control-language.js" });
  vm.runInContext(script, environment, { filename: "studio.js" });
  return {
    nodes, requests, document: environment.document, timers, confirms, decisions, windowEvents, navigation, navigationAttempts,
    follow(href, options = {}) {
      const link = new Element();
      Object.assign(link, { href, target: "", closest: () => link, hasAttribute: () => false }, options.link);
      const { link: _, ...eventOptions } = options;
      const event = {
        button: 0, target: link, defaultPrevented: false, preventDefault() { this.defaultPrevented = true; }, ...eventOptions,
      };
      const result = documentEvents.click(event);
      return { link, event, result };
    },
    select: async (id, value) => { nodes[id].value = value; await nodes[id].fire("change"); },
    submit: () => nodes["configuration-form"].fire("submit"),
    again: () => nodes["start-another"].fire("click"),
    reload: () => nodes["reload-lists"].fire("click"),
  };
}
function version(state = "Draft", extra = {}) {
  return { id: 3, configurationId: 2, number: 1, state, createdAt: "2026-01-01T12:00:00Z", updatedAt: "2026-01-01T12:00:00Z",
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

function publicationServer() {
  const versions = `${base}/1/configurations/2/versions`;
  const server = {
    items: [version("Draft", { listener: { tls: { enabled: true } } }), version("Published", { id: 4, number: 2 })],
    attempts: 0,
    refreshError: false,
    route(url, options) {
      if (options.method === "GET" && url === versions) {
        if (server.refreshError) return reply({}, 503);
        return reply(structuredClone(server.items));
      }
      if (options.method === "PUT" && url === `${versions}/3/listener`) {
        const draft = server.items.find(({ id }) => id === 3);
        Object.assign(draft.listener, JSON.parse(options.body));
        draft.updatedAt = "2026-01-01T12:00:01.123456789Z";
        return reply(structuredClone(draft));
      }
      if (options.method === "POST" && url === `${versions}/3/publish`) {
        server.attempts++;
        const draft = server.items.find(({ id }) => id === 3);
        if (server.onPublish) return server.onPublish(draft);
        if (draft.state !== "Draft") return apiError(409, "version_not_publishable");
        for (const item of server.items) if (item.state === "Published") item.state = "Archived";
        draft.state = "Published";
        return reply(structuredClone(draft));
      }
      assert.equal(options.method, "GET", `Unexpected mutation ${url}`);
      assert.ok(url === base || url === `${base}/1/configurations`, `Unexpected read ${url}`);
      return routes(url, options);
    },
  };
  return server;
}

test("Publish Draft confirms scope, publishes a TLS-enabled Draft once and refreshes versions", async () => {
  const server = publicationServer(), app = start(server.route);
  await editor(app);
  const { nodes } = app;
  const button = nodes["publish-draft"];
  assert.equal(button.disabled, false);
  const before = app.requests.length;
  app.navigation.autoConfirm = false;
  button.fire("click"); button.fire("click");
  assert.equal(app.confirms.length, 1);
  assert.match(app.confirms[0], /Draft v1 \(#3\).*Configuration Configuration \(#2\).*archived/);
  assert.equal(app.requests.length, before);
  nodes["confirmation-accept"].fire("click");
  nodes["confirmation-accept"].fire("click"); button.fire("click");
  await settle();
  assert.deepEqual(app.requests.slice(before).map(({ url, options }) => [options.method, url.split("/versions").at(-1)]),
    [["GET", ""], ["POST", "/3/publish"], ["GET", ""]]);
  assert.equal(app.requests[before + 1].options.body, "{}");
  assert.equal(server.items[1].state, "Archived");
  assert.equal(nodes["existing-draft"].value, "");
  assert.equal(nodes["existing-draft"].children.length, 1);
  assert.equal(nodes["result-state"].textContent, "Published");
  assert.equal(nodes["result-url-row"].hidden, true);
  assert.equal(nodes["flow-status"].focused, true);
  assert.equal(button.disabled, true);
});

test("Publish Draft Cancel and Escape restore focus, and changed scope cancels the mutation", async () => {
  const server = publicationServer(), app = start(server.route);
  await editor(app);
  const { nodes } = app, button = nodes["publish-draft"];
  const before = app.requests.length;
  app.navigation.autoConfirm = false;
  button.fire("click"); nodes["confirmation-cancel"].fire("click");
  assert.equal(button.focused, true);
  button.focused = false;
  button.fire("click"); nodes["studio-confirmation"].fire("cancel");
  assert.equal(button.focused, true);
  button.fire("click"); nodes["existing-draft"].value = "9";
  nodes["confirmation-accept"].fire("click"); await settle();
  assert.equal(app.requests.length, before);
  assert.equal(server.attempts, 0);
});

test("Dirty listener, dirty timeouts and uncertain Save block Publish Draft", async () => {
  const server = publicationServer();
  const app = start((url, options) => options.method === "PUT" ? reply({}, 503) : server.route(url, options));
  await editor(app);
  const { nodes } = app, button = nodes["publish-draft"];
  nodes["listener-port"].value = "9090"; nodes["listener-port"].fire("input");
  assert.equal(button.disabled, true); button.fire("click");
  await app.select("draft-section", "timeouts");
  assert.equal(button.disabled, false);
  nodes["timeout-read"].value = "17"; nodes["timeout-read"].fire("input");
  assert.equal(button.disabled, true); button.fire("click");
  await app.submit();
  nodes["timeout-read"].value = "0"; nodes["timeout-read"].fire("input");
  assert.equal(button.disabled, true); button.fire("click");
  assert.match(nodes["draft-status"].textContent, /Save outcome unknown/);
  assert.equal(server.attempts, 0);
  await nodes["reload-draft"].fire("click");
  assert.equal(button.disabled, false);
});

test("another Studio tab can change a Draft and stale content requires reload and reconfirmation", async () => {
  const server = publicationServer(), first = start(server.route), second = start(server.route);
  await editor(first); await editor(second);
  first.navigation.autoConfirm = false;
  first.nodes["publish-draft"].fire("click");
  second.nodes["listener-port"].value = "9090";
  await second.submit();
  first.nodes["confirmation-accept"].fire("click"); await settle();
  assert.equal(server.attempts, 0);
  assert.match(first.nodes["draft-status"].textContent, /stored Draft changed/);
  assert.equal(first.nodes["publish-draft"].disabled, true);
  assert.equal(first.nodes["listener-port"].value, "8080");
  await first.nodes["reload-draft"].fire("click");
  assert.equal(first.nodes["listener-port"].value, "9090");
  assert.equal(first.nodes["publish-draft"].disabled, false);
  first.nodes["publish-draft"].fire("click");
  first.nodes["confirmation-accept"].fire("click"); await settle();
  assert.equal(server.attempts, 1);
  assert.equal(server.items[0].listener.port, 9090);
});

test("timestamp-only change, terminal state and missing Draft all prevent a publish POST", async () => {
  for (const change of [
    ({ items }) => { items[0].updatedAt = "2026-01-01T12:00:00.000000001Z"; },
    ({ items }) => { items[0].state = "Archived"; },
    ({ items }) => { items.shift(); },
  ]) {
    const server = publicationServer(), app = start(server.route);
    await editor(app); change(server);
    await app.nodes["publish-draft"].fire("click");
    assert.equal(server.attempts, 0);
    assert.equal(app.nodes["publish-draft"].disabled, true);
  }
});

test("reload of an archived Draft preserves local values and invalidates publish eligibility", async () => {
  const server = publicationServer(), app = start(server.route);
  await editor(app);
  server.items[0].state = "Archived";
  await app.nodes["reload-draft"].fire("click");
  assert.equal(app.nodes["listener-port"].value, "8080");
  assert.equal(app.nodes["publish-draft"].disabled, true);
  app.nodes["publish-draft"].fire("click");
  assert.equal(app.confirms.length, 0);
});

test("a successful section Save enables publication unless the other section changed on the server", async () => {
  for (const remoteChange of [false, true]) {
    const server = publicationServer(), app = start(server.route);
    await editor(app);
    if (remoteChange) server.items[0].listener.timeouts.readSeconds = 15;
    app.nodes["listener-port"].value = "9090";
    await app.submit();
    assert.equal(app.nodes["timeout-read"].value, "0");
    assert.equal(app.nodes["publish-draft"].disabled, remoteChange);
    if (remoteChange) await app.nodes["reload-draft"].fire("click");
    await app.nodes["publish-draft"].fire("click");
    assert.equal(server.attempts, 1);
    assert.equal(server.items[0].listener.port, 9090);
  }
});

test("a typed publish conflict after the fresh read preserves the exact Draft, even when refresh fails", async () => {
  const server = publicationServer(), app = start(server.route);
  await editor(app);
  server.items.push(version("Draft", { id: 6, number: 3 }));
  server.onPublish = (draft) => {
    draft.state = "Archived";
    server.refreshError = true;
    return apiError(409, "version_not_publishable");
  };
  await app.nodes["publish-draft"].fire("click");
  assert.match(app.nodes["flow-status"].textContent, /Version refresh failed/);
  assert.match(app.nodes["draft-status"].textContent, /no longer a Draft/);
  assert.equal(app.nodes["flow-status"].focused, true);
  await assertStaleActionsBlocked(app, 1);
  server.refreshError = false;
  await app.nodes["reload-draft"].fire("click");
  assert.equal(app.requests.at(-1).options.method, "GET");
  assert.match(app.nodes["flow-status"].textContent, /no longer a Draft/);
  assert.deepEqual(app.nodes["existing-draft"].children.map((option) => option.textContent),
    ["Select a Draft", "Draft v3 (#6)", "Unavailable Draft v1 (#3)"]);
  await editHost(app);
  await assertStaleActionsBlocked(app, 1);
});

test("an uncertain publish needs explicit reads, and observed lifecycle alone gives no success", async () => {
  for (const state of ["Published", "Archived", "Draft"]) {
    const server = publicationServer(), app = start(server.route);
    await editor(app);
    server.onPublish = (draft) => {
      draft.state = state;
      if (state === "Published") throw new Error("lost response");
      if (state === "Archived") return reply({}, 503);
      return reply(version("Published", { id: 99 }));
    };
    const before = app.requests.length;
    await app.nodes["publish-draft"].fire("click");
    assert.deepEqual(app.requests.slice(before).map(({ options }) => options.method), ["GET", "POST"]);
    assert.match(app.nodes["flow-status"].textContent, /Publish outcome unknown/);
    assert.equal(app.nodes["submit-button"].disabled, true);
    assert.equal(app.nodes["publish-draft"].disabled, true);
    server.refreshError = true;
    await app.nodes["reload-draft"].fire("click");
    assert.equal(app.nodes["publish-draft"].disabled, true);
    server.refreshError = false;
    await app.nodes["reload-draft"].fire("click");
    assert.equal(server.attempts, 1);
    assert.equal(app.nodes.result.hidden, true);
    assert.match(app.nodes["flow-status"].textContent, new RegExp(`outcome unknown.*${state}`));
    assert.equal(app.nodes["publish-draft"].disabled, state !== "Draft");
    if (state === "Draft") {
      delete server.onPublish;
      app.navigation.autoConfirm = false;
      app.nodes["publish-draft"].fire("click");
      assert.equal(server.attempts, 1);
      app.nodes["confirmation-accept"].fire("click"); await settle();
      assert.equal(server.attempts, 2);
    }
  }
});

test("acknowledged publication stays successful when refresh fails and retry is read-only", async () => {
  const server = publicationServer(), app = start(server.route);
  server.items.pop();
  await editor(app);
  server.onPublish = (draft) => { server.refreshError = true; draft.state = "Published"; return reply(draft); };
  await app.nodes["publish-draft"].fire("click");
  assert.equal(app.nodes["result-state"].textContent, "Published");
  assert.match(app.nodes["flow-status"].textContent, /Draft published.*refresh failed/);
  assert.equal(app.nodes["publish-draft"].disabled, true);
  server.refreshError = false;
  await app.nodes["reload-draft"].fire("click");
  assert.equal(server.attempts, 1);
  assert.equal(app.nodes["existing-draft"].children.length, 1);
});

test("a confirmed publication shows Archived if another version supersedes it before refresh", async () => {
  const server = publicationServer(), app = start(server.route);
  await editor(app);
  server.onPublish = (draft) => {
    draft.state = "Archived";
    return reply({ ...draft, state: "Published" });
  };
  await app.nodes["publish-draft"].fire("click");
  assert.equal(app.nodes["result-state"].textContent, "Archived");
  assert.match(app.nodes["flow-status"].textContent, /Draft published.*Archived/);
  assert.equal(app.nodes["existing-draft"].children.length, 1);
});

test("Publish confirmation and outcome follow the active language", async () => {
  const server = publicationServer(), app = start(server.route);
  await editor(app);
  app.nodes["control-language"].fire("click");
  app.navigation.autoConfirm = false;
  app.nodes["publish-draft"].fire("click");
  assert.equal(app.nodes["confirmation-title"].textContent, "Опубликовать черновик?");
  assert.match(app.nodes["confirmation-message"].textContent, /v1 \(#3\).*Configuration \(#2\).*архивирована/);
  assert.equal(app.nodes["confirmation-accept"].textContent, "Опубликовать черновик");
  app.nodes["confirmation-accept"].fire("click"); await settle();
  assert.match(app.nodes["flow-status"].textContent, /опубликована/);
});
const mutations = (app) => app.requests.filter(({ options }) => options.method !== "GET");
test("original new parent flow remains complete and same-origin", async () => {
  const app = start(routes); await app.submit();
  assert.deepEqual(mutations(app).map(({ options }) => options.method), ["POST", "POST", "POST", "PUT", "POST"]);
  assert.equal(app.nodes["result-state"].textContent, "Published");
  assert.equal(app.nodes["open-created-draft"].hidden, true);
  for (const { options } of app.requests) {
    assert.equal(options.mode, "same-origin"); assert.equal(options.redirect, "error"); assert.equal(options.cache, "no-store");
  }
});
test("existing Configuration creates only Draft without parent mutations or publish", async () => {
  const app = start(routes); await existing(app); await app.submit();
  assert.deepEqual(mutations(app).map(({ url }) => url), [`${base}/1/configurations/2/versions`]);
  assert.equal(app.nodes["result-state"].textContent, "Draft");
  assert.equal(app.nodes["result-url-row"].hidden, true);
  assert.equal(app.nodes["open-created-draft"].hidden, false);
  assert.equal(app.nodes["open-created-draft"].focused, true);
});

test("Open created Draft uses one fresh scoped GET and opens the exact Draft with fresh values", async () => {
  const pending = deferred();
  const versionsPath = `${base}/4/configurations/2/versions`;
  const created = version("Draft", { id: 15, number: 2 });
  const app = start((url, options) => {
    if (url === versionsPath) return options.method === "POST" ? reply(created, 201) : pending.promise;
    if (url === `${versionsPath}/15/listener` && options.method === "PUT") {
      return reply({ ...created, listener: { ...created.listener, ...JSON.parse(options.body) } });
    }
    return routes(url, options);
  });
  await app.select("creation-mode", "existing-configuration");
  await app.select("existing-workspace", "4");
  await app.select("existing-configuration", "2");
  await app.submit();
  const button = app.nodes["open-created-draft"];
  const before = app.requests.length;
  const opening = button.fire("click");
  assert.equal(button.disabled, true);
  assert.equal(app.nodes["start-another"].disabled, true);
  await button.fire("click");
  await app.again();
  await app.submit();
  await app.select("existing-workspace", "1");
  pending.resolve(reply([version(), { ...created, listener: { ...created.listener, port: 9015 } }, version("Draft", { id: 17, number: 3 })]));
  await opening;
  assert.deepEqual(app.requests.slice(before).map(({ url, options }) => [options.method, url, options.cache]), [["GET", versionsPath, "no-store"]]);
  assert.equal(app.nodes["creation-mode"].value, "existing-draft");
  assert.equal(app.nodes["existing-workspace"].value, "4");
  assert.equal(app.nodes["existing-configuration"].value, "2");
  assert.equal(app.nodes["existing-draft"].value, "15");
  assert.equal(app.nodes["existing-draft"].focused, true);
  assert.equal(app.nodes["listener-port"].value, "9015");
  assert.equal(app.nodes["listener-port"].disabled, false);
  assert.equal(app.nodes["submit-button"].disabled, true);
  assert.equal(app.nodes.result.hidden, true);
  assert.equal(app.nodes["start-another"].hidden, true);
  await button.fire("click");
  assert.equal(app.requests.length, before + 1);
  await editHost(app, "localhost");
  await app.submit();
  assert.deepEqual(mutations(app).map(({ url, options }) => [options.method, url]), [
    ["POST", versionsPath], ["PUT", `${versionsPath}/15/listener`],
  ]);
  assert.equal(app.nodes["submit-button"].disabled, true);
});

for (const [scenario, replacement] of [
  ["Published", version("Published")], ["Archived", version("Archived")], ["deleted", null],
  ["replaced after ID reuse", version("Draft", { createdAt: "2026-01-01T12:00:00.000000001Z" })],
]) {
  test(`Open created Draft keeps the result when its exact target is ${scenario}`, async () => {
    const versions = [version("Draft", { id: 5, number: 2 })];
    if (replacement) versions.push(replacement);
    const app = start((url, options) => url.endsWith("/versions") && options.method === "GET" ? reply(versions) : routes(url, options));
    await existing(app);
    await app.submit();
    const button = app.nodes["open-created-draft"];
    await button.fire("click");
    assert.equal(app.nodes["creation-mode"].value, "existing-configuration");
    assert.equal(app.nodes["existing-draft"].value, "");
    assert.equal(app.nodes.result.hidden, false);
    assert.equal(button.disabled, false);
    assert.equal(app.nodes["flow-status"].textContent, "The created version is no longer available as a Draft.");
    assert.equal(app.nodes["flow-status"].focused, true);
    assert.equal(mutations(app).length, 1);
  });
}

test("Open created Draft reports a localized read failure and can retry without creating again", async () => {
  let failed = true;
  const app = start((url, options) => {
    if (options.method !== "GET" || !url.endsWith("/versions")) return routes(url, options);
    if (failed) throw new TypeError("network error");
    return reply([version()]);
  });
  await existing(app);
  await app.submit();
  app.nodes["control-language"].fire("click");
  assert.equal(app.nodes["open-created-draft"].textContent, "Открыть созданный черновик");
  await app.nodes["open-created-draft"].fire("click");
  assert.equal(app.nodes["flow-status"].textContent, "Не удалось открыть созданный черновик. Попробуйте открыть его снова.");
  assert.equal(app.nodes["existing-workspace"].value, "1");
  assert.equal(app.nodes["existing-configuration"].value, "2");
  assert.equal(app.nodes["result-version"].textContent, "1");
  assert.equal(app.nodes.result.hidden, false);
  failed = false;
  await app.nodes["open-created-draft"].fire("click");
  assert.equal(app.nodes["existing-draft"].value, "3");
  assert.equal(mutations(app).length, 1);
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
  assert.equal(app.nodes["open-created-draft"].hidden, true);
  await app.nodes["open-created-draft"].fire("click");
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
  assert.equal(app.nodes["open-created-draft"].hidden, true);
  const before = app.requests.length;
  await app.nodes["open-created-draft"].fire("click");
  assert.equal(app.requests.length, before);
  assert.match(app.nodes["start-another"].textContent, /I checked Dashboard/);
  await app.again();
  assert.equal(app.nodes["existing-workspace"].value, "");
  await app.submit(); assert.equal(mutations(app).length, 1);
});
test("malformed Draft scope cannot become success", async () => {
  const app = start((url, options) => options.method === "POST" ? reply(version("Draft", { configurationId: 9 }), 201) : routes(url, options));
  await existing(app); await app.submit();
  assert.equal(app.nodes.result.hidden, true); assert.equal(mutations(app).length, 1);
  assert.equal(app.nodes["open-created-draft"].hidden, true);
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
  const departure = app.follow("http://localhost/dashboard");
  assert.equal(departure.event.defaultPrevented, true);
  assert.equal(app.confirms.length, 0);
  assert.deepEqual(app.navigationAttempts, []);
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
  assert.match(app.nodes[`${id}-error`].textContent, /timeout must be a whole number/);
  assert.equal(app.nodes[id].attributes["aria-invalid"], "true");
});

test("Save validates every timeout, focuses the first invalid field and clears only corrected errors", async () => {
  const app = start(timeoutRoutes);
  await timeoutEditor(app);
  await editTimeout(app, "301");
  app.nodes["timeout-read"].value = "0x10";
  app.nodes["timeout-write"].value = "0";
  app.nodes["timeout-idle"].value = " 1 ";
  await app.submit();
  assert.equal(mutations(app).length, 0);
  assert.deepEqual(timeoutIDs.map((id) => Boolean(app.nodes[id].focused)), [true, false, false, false]);
  for (const [id, label] of timeoutIDs.map((id, index) => [id, ["Handshake", "Read", "Write", "Idle"][index]])) {
    assert.match(app.nodes[`${id}-error`].textContent, new RegExp(`^${label} timeout`));
    assert.equal(app.nodes[id].attributes["aria-invalid"], "true");
  }
  await editTimeout(app, "0", "timeout-read");
  assert.equal(app.nodes["timeout-read-error"].textContent, "");
  assert.equal(app.nodes["timeout-read"].attributes["aria-invalid"], undefined);
  for (const id of ["timeout-handshake", "timeout-write", "timeout-idle"]) {
    assert.notEqual(app.nodes[`${id}-error`].textContent, "");
    assert.equal(app.nodes[id].attributes["aria-invalid"], "true");
  }
  await editTimeout(app, "300");
  await editTimeout(app, "1", "timeout-write");
  await editTimeout(app, "0", "timeout-idle");
  assert.equal(app.nodes["flow-status"].textContent, "Ready.");
  await app.submit();
  assert.deepEqual(mutations(app).map(({ url, options }) => [url, options.method, JSON.parse(options.body)]), [
    [`${versionsURL}/3/listener/timeouts`, "PUT", { handshakeSeconds: 300, readSeconds: 0, writeSeconds: 1, idleSeconds: 0 }],
  ]);
  assert.equal(app.nodes["existing-draft"].value, "3");
});

test("timeout errors survive cancelled Discard and clear when stored values are restored", async () => {
  const app = start(timeoutRoutes);
  await timeoutEditor(app);
  await editTimeout(app, "301");
  await editTimeout(app, "86401", "timeout-idle");
  await app.submit();
  const errors = () => timeoutIDs.map((id) => app.nodes[`${id}-error`].textContent);
  const invalid = errors();
  app.decisions.push(false);
  await app.select("draft-section", "listener");
  assert.equal(app.nodes["draft-section"].value, "timeouts");
  assert.deepEqual(errors(), invalid);
  assert.equal(app.nodes["timeout-idle"].value, "86401");
  await app.nodes["reload-draft"].fire("click");
  assert.deepEqual(errors(), ["", "", "", ""]);
  assert.equal(app.nodes["timeout-idle"].value, String(timeoutValues.idleSeconds));
  assert.ok(timeoutIDs.every((id) => app.nodes[id].attributes["aria-invalid"] === undefined));
  assert.equal(mutations(app).length, 0);
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

test("language changes preserve active Draft, local timeout validation and translated Save result", async () => {
  const storage = new Map();
  const app = start(timeoutRoutes, storage);
  assert.equal(app.document.documentElement.lang, "en");
  assert.equal(app.nodes["control-language-code"].textContent, "EN");
  assert.equal(app.nodes["control-language"].attributes["aria-label"], "EN (English). Switch to Russian.");
  await timeoutEditor(app);
  await editTimeout(app, "301");
  await editTimeout(app, "-1", "timeout-idle");
  await app.submit();
  const values = () => ["creation-mode", "existing-workspace", "existing-configuration", "existing-draft",
    "draft-section", "timeout-handshake"].map((id) => app.nodes[id].value);
  const before = values();
  const reads = app.requests.length;
  for (const locale of ["ru", "en", "ru"]) {
    await app.nodes["control-language"].fire("click");
    assert.equal(app.document.documentElement.lang, locale);
    assert.equal(app.nodes["control-language-code"].textContent, locale.toUpperCase());
    assert.equal(app.nodes["control-language"].attributes["aria-label"], locale === "ru"
      ? "RU (русский). Переключить на английский." : "EN (English). Switch to Russian.");
    assert.match(app.nodes["timeout-handshake-error"].textContent, locale === "ru" ? /Таймаут рукопожатия.*1.*300/ : /Handshake timeout.*1.*300/);
    assert.match(app.nodes["timeout-idle-error"].textContent, locale === "ru" ? /Таймаут простоя.*0.*86400/ : /Idle timeout.*0.*86400/);
    assert.deepEqual(values(), before);
    assert.equal(app.requests.length, reads);
  }
  await editTimeout(app, "35");
  await editTimeout(app, String(timeoutValues.idleSeconds), "timeout-idle");
  await app.submit();
  assert.match(app.nodes["flow-status"].textContent, /Таймауты сохранены/);
  await app.nodes["control-language"].fire("click");
  assert.match(app.nodes["flow-status"].textContent, /Timeouts saved/);
  await app.nodes["control-language"].fire("click");
  assert.match(app.nodes["flow-status"].textContent, /Таймауты сохранены/);
  assert.deepEqual(values().slice(0, -1), before.slice(0, -1));
  assert.equal(app.nodes["timeout-handshake"].value, "35");
  assert.equal(app.nodes["submit-button"].disabled, true);
  assert.deepEqual(mutations(app).map(({ url, options }) => [url, JSON.parse(options.body)]), [
    [`${versionsURL}/3/listener/timeouts`, { ...timeoutValues, handshakeSeconds: 35 }],
  ]);
  assert.equal(app.requests.length, reads + 1);
  assert.equal(app.confirms.length, 0);
  assert.equal(start(timeoutRoutes, storage).document.documentElement.lang, "ru");
});

test("translated listener rejection stays pinned until the rejected value is corrected", async () => {
  const app = start((url, options) => options.method === "PUT" && JSON.parse(options.body).host === "localhost"
    ? fieldRejection("host", "must be a valid IP address or hostname") : editorRoutes(url, options));
  await editor(app);
  const input = await listenerInput(app, "host", "localhost");
  await app.submit();
  await app.nodes["control-language"].fire("click");
  assert.match(app.nodes["listener-host-error"].textContent, /корректный IP-адрес/);
  assert.match(app.nodes["flow-status"].textContent, /выделенные поля/);
  assert.equal(input.attributes["aria-invalid"], "true");
  await input.fire("blur");
  await app.submit();
  assert.equal(mutations(app).length, 1);
  await listenerInput(app, "host", "example.test");
  assert.equal(app.nodes["listener-host-error"].textContent, "");
  assert.equal(app.nodes["flow-status"].textContent, "Готово.");
});

test("missing translation uses English and both pages have a Russian entry for all marked content", () => {
  const nodes = [new Element(), new Element()];
  const code = new Element();
  const unlisted = "Unlisted message {0}";
  for (const file of ["index.html", "dashboard.html"]) {
    const html = fs.readFileSync(path.join(__dirname, "assets", file), "utf8");
    for (const [, copy, label] of html.matchAll(/\bdata-i18n>([^<]+)|\bdata-i18n-label aria-label="([^"]+)"/g)) {
      const key = copy || label;
      if (key) {
        const node = new Element();
        node.textContent = key;
        nodes.push(node);
      }
    }
  }
  nodes[1].textContent = unlisted;
  const env = vm.createContext({
    window: { addEventListener() {} },
    document: { documentElement: {}, getElementById: (id) => id === "control-language" ? nodes[0] : code,
      querySelectorAll: (selector) => selector === "[data-i18n]" ? nodes.slice(1) : [] },
    localStorage: { getItem: () => "unsupported", setItem: () => { throw new Error("storage unavailable"); } },
  });
  vm.runInContext(languageScript, env);
  assert.equal(env.document.documentElement.lang, "en");
  nodes[0].fire("click");
  assert.equal(code.textContent, "RU");
  assert.equal(env.document.documentElement.lang, "ru");
  assert.equal(vm.runInContext('ControlLanguage.text(ControlLanguage.message("Unlisted message {0}", 1))', env), "Unlisted message 1");
  for (const node of nodes.slice(2)) assert.match(node.textContent, /[А-Яа-яЁё]/);
});

test("dirty internal navigation opens one dialog and Escape preserves the draft", async () => {
  const app = start(editorRoutes);
  await editor(app);
  await editHost(app);
  app.navigation.autoConfirm = false;
  const original = app.follow("http://localhost/dashboard#workspaces");
  const duplicate = app.follow("http://localhost/dashboard#configurations");
  assert.equal(original.event.defaultPrevented, true);
  assert.equal(duplicate.event.defaultPrevented, true);
  assert.equal(app.confirms.length, 1);
  const before = app.requests.length;
  await app.nodes["studio-confirmation"].fire("cancel");
  await original.result;
  assert.equal(app.requests.length, before);
  assert.deepEqual(app.navigationAttempts, []);
  assert.equal(app.nodes["listener-host"].value, "localhost");
  assert.equal(app.nodes["existing-draft"].value, "3");
  assert.equal(app.nodes["draft-section"].value, "listener");
  assert.equal(original.link.focused, true);
  assert.equal(app.nodes["studio-confirmation"].attributes["aria-describedby"], "confirmation-message");
});

test("confirmed departure attempts the captured route once and later edits still trigger native unload", async () => {
  const app = start(editorRoutes);
  await editor(app);
  await editHost(app);
  app.navigation.autoConfirm = false;
  app.navigation.fail = true;
  const leaving = app.follow("http://localhost/dashboard#workspaces");
  await app.nodes["confirmation-accept"].fire("click");
  await app.nodes["confirmation-accept"].fire("click");
  await leaving.result;
  assert.deepEqual(app.navigationAttempts, ["http://localhost/dashboard#workspaces"]);
  assert.equal(app.nodes["listener-host"].value, "127.0.0.1");
  assert.equal(app.nodes["existing-draft"].value, "3");
  assert.match(app.nodes["flow-status"].textContent, /Could not open the requested page/);
  await editHost(app);
  const unload = { preventDefault() { this.prevented = true; } };
  app.windowEvents.beforeunload(unload);
  assert.equal(unload.prevented, true);
});

test("dirty desktop self reload is guarded while anchors, external and new-tab links retain browser behavior", async () => {
  const app = start(editorRoutes);
  await editor(app);
  await editHost(app);
  app.decisions.push(false);
  const same = app.follow("http://localhost/");
  assert.equal(same.event.defaultPrevented, true);
  await same.result;
  for (const [url, opts] of [
    ["http://localhost/#studio"], ["https://example.test/"],
    ["http://localhost/dashboard", { ctrlKey: true }],
    ["http://localhost/dashboard", { link: { target: "_blank" } }],
    ["http://localhost/dashboard", { link: { hasAttribute: () => true } }],
  ]) assert.equal(app.follow(url, opts).event.defaultPrevented, false);
  assert.equal(app.confirms.length, 1);
});

test("unknown save departure keeps its warning and requires a fresh read before Save on return", async () => {
  const app = start((url, options) => options.method === "PUT" ? reply({}, 500) : editorRoutes(url, options));
  await editor(app);
  await editHost(app);
  await app.submit();
  app.navigation.autoConfirm = false;
  const pending = app.follow("http://localhost/dashboard");
  assert.equal(app.nodes["confirmation-warning"].hidden, false);
  assert.equal(app.nodes["studio-confirmation"].attributes["aria-describedby"], "confirmation-message confirmation-warning");
  await app.nodes["confirmation-accept"].fire("click");
  await pending.result;
  assert.deepEqual(app.navigationAttempts, ["http://localhost/dashboard"]);
  assert.equal(app.nodes["existing-draft"].value, "");
  assert.equal(app.nodes["submit-button"].disabled, true);
  const count = app.requests.length;
  await app.select("existing-draft", "3");
  assert.equal(app.requests.length, count + 1);
  assert.equal(app.requests.at(-1).options.method, "GET");
});

function apiError(status, code) {
  return reply({ error: { code, message: "server detail" } }, status);
}

async function assertStaleActionsBlocked(app, mutationCount) {
  assert.equal(app.nodes["existing-draft"].value, "3");
  assert.equal(app.nodes["submit-button"].disabled, true);
  assert.equal(app.nodes["publish-draft"].disabled, true);
  await app.submit();
  await app.submit();
  await app.nodes["publish-draft"].fire("click");
  assert.equal(mutations(app).length, mutationCount);
}

for (const [section, field, original, local] of [
  ["listener", "listener-host", "127.0.0.1", "localhost"],
  ["timeouts", "timeout-handshake", "10", "45"],
]) test(`typed lifecycle rejection blocks ${section} Save and Publish with retained input and focus`, async () => {
  const app = start((url, options) => options.method === "PUT"
    ? apiError(409, "version_not_editable") : editorRoutes(url, options));
  await editor(app);
  await app.select("draft-section", section);
  const input = app.nodes[field];
  input.value = local;
  await input.fire("input");
  await app.submit();
  assert.equal(input.value, local);
  assert.match(app.nodes["flow-status"].textContent, /no longer a Draft.*still in the form/);
  assert.equal(app.nodes["flow-status"].focused, true);
  assert.equal(app.nodes["existing-draft"].children.some((option) => option.value === "6"), true);
  await assertStaleActionsBlocked(app, 1);

  input.value = original;
  await input.fire("input");
  await assertStaleActionsBlocked(app, 1);
  input.value = local;
  await input.fire("input");
  await app.nodes["reload-draft"].fire("click"); // A contradictory read cannot unlock this editor.
  assert.equal(input.value, local);
  assert.equal(app.confirms.length, 0);
  assert.match(app.nodes["draft-status"].textContent, /no longer a Draft/);
  await assertStaleActionsBlocked(app, 1);

  app.decisions.push(false);
  const leave = app.follow("http://localhost/dashboard");
  assert.equal(leave.event.defaultPrevented, true);
  await leave.result;
  assert.equal(input.value, local);
  app.decisions.push(false);
  await app.select("existing-draft", "6");
  await assertStaleActionsBlocked(app, 1);
  app.decisions.push(true);
  await app.select("existing-draft", "6");
  assert.equal(app.nodes["existing-draft"].value, "6");
  assert.equal(app.nodes["publish-draft"].disabled, false);
});

for (const [name, result, expected] of [
  ["Published", () => reply([version("Published"), version("Draft", { id: 6, number: 2 })]), /no longer a Draft/],
  ["Archived", () => reply([version("Archived")]), /no longer a Draft/],
  ["missing Draft", () => reply([version("Draft", { id: 6, number: 2 })]), /Draft is unavailable/],
  ["reused ID", () => reply([version("Draft", { createdAt: "2026-01-02T12:00:00Z" })]), /Draft is unavailable/],
  ["missing Configuration", () => apiError(404, "configuration_not_found"), /Draft is unavailable/],
]) test(`Reload confirms ${name} and preserves dirty local values`, async () => {
  let changed = false;
  const app = start((url, options) => changed && url === versionsURL ? result() : editorRoutes(url, options));
  await editor(app);
  await editHost(app);
  changed = true;
  await app.nodes["reload-draft"].fire("click");
  assert.equal(app.nodes["listener-host"].value, "localhost");
  assert.match(app.nodes["draft-status"].textContent, expected);
  await assertStaleActionsBlocked(app, 0);
});

for (const code of ["configuration_not_found", "version_not_found"]) test(`Save reports scoped ${code} in RU`, async () => {
  const app = start((url, options) => options.method === "PUT" ? apiError(404, code) : editorRoutes(url, options),
    new Map([["uwp-control-language", "ru"]]));
  await editor(app);
  await editHost(app);
  await app.submit();
  assert.equal(app.nodes["listener-host"].value, "localhost");
  assert.match(app.nodes["draft-status"].textContent, /Этот черновик недоступен.*правки остались/);
  assert.equal(app.nodes["flow-status"].focused, true);
  await assertStaleActionsBlocked(app, 1);
});

test("ambiguous HTTP conflicts do not prove lifecycle state", async () => {
  for (const response of [reply({}, 409), apiError(409, "version_not_publishable"), reply({}, 404),
    { status: 409, json: async () => { throw new SyntaxError("malformed"); } }]) {
    const app = start((url, options) => options.method === "PUT" ? response : editorRoutes(url, options));
    await editor(app);
    await editHost(app);
    await app.submit();
    assert.equal(app.nodes["listener-host"].value, "localhost");
    assert.doesNotMatch(app.nodes["draft-status"].textContent, /no longer a Draft|Draft is unavailable/);
    assert.equal(app.nodes["submit-button"].disabled, false);
  }

  const server = publicationServer(), app = start(server.route);
  server.onPublish = () => apiError(409, "version_not_editable");
  await editor(app);
  await app.nodes["publish-draft"].fire("click");
  assert.match(app.nodes["flow-status"].textContent, /Publish outcome unknown/);
  await app.nodes["reload-draft"].fire("click");
  assert.equal(app.nodes["publish-draft"].disabled, false);
  assert.equal(server.attempts, 1);
});

test("Publish preflight distinguishes scoped absence from an untrusted status or failed read", async () => {
  let failure;
  const app = start((url, options) => {
    if (url === versionsURL && failure) return failure();
    return editorRoutes(url, options);
  });
  await editor(app);
  for (const read of [() => { throw new TypeError("offline"); }, () => reply({}, 503),
    () => reply({}, 404), () => apiError(404, "version_not_found"),
    () => reply([version("Archived", { configurationId: 99 })]),
    () => reply([version("Draft", { createdAt: "2026" })])]) {
    failure = read;
    await app.nodes["publish-draft"].fire("click");
    assert.match(app.nodes["flow-status"].textContent, /Could not check the stored Draft/);
    assert.equal(app.nodes["existing-draft"].value, "3");
    assert.equal(app.nodes["publish-draft"].disabled, false);
  }
  failure = () => apiError(404, "configuration_not_found");
  await app.nodes["publish-draft"].fire("click");
  assert.match(app.nodes["draft-status"].textContent, /Draft is unavailable/);
  await assertStaleActionsBlocked(app, 0);
});

test("malformed creation timestamp on Reload preserves the active editor without marking it stale", async () => {
  let malformed = false;
  const app = start((url, options) => malformed && url === versionsURL
    ? reply([version("Draft", { createdAt: "not a timestamp" })]) : editorRoutes(url, options));
  await editor(app);
  await editHost(app);
  malformed = true;
  await app.nodes["reload-draft"].fire("click");
  assert.equal(app.nodes["existing-draft"].value, "3");
  assert.equal(app.nodes["listener-host"].value, "localhost");
  assert.equal(app.nodes["submit-button"].disabled, false);
  assert.match(app.nodes["flow-status"].textContent, /Could not open Drafts/);
  assert.equal(mutations(app).length, 0);
});

test("late scoped 404 from Reload cannot invalidate a Draft in a newer editor selection", async () => {
  const oldRead = deferred();
  let delay = false;
  const app = start((url, options) => delay && url === versionsURL ? oldRead.promise : editorRoutes(url, options));
  await editor(app);
  delay = true;
  const pending = app.nodes["reload-draft"].fire("click");
  await settle();
  await app.select("creation-mode", "new-workspace");
  delay = false;
  await editor(app, false);
  await app.select("existing-draft", "6");
  oldRead.resolve(apiError(404, "configuration_not_found"));
  await pending;
  assert.equal(app.nodes["existing-draft"].value, "6");
  assert.equal(app.nodes["publish-draft"].disabled, false);
  assert.doesNotMatch(app.nodes["draft-status"].textContent, /unavailable/);
});
