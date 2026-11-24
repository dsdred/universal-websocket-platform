"use strict";
// Run with node --test internal/configurationstudio/dashboard_test.js. No npm.
const test = require("node:test");
const assert = require("node:assert/strict");
const fs = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const script = fs.readFileSync(path.join(__dirname, "assets/dashboard.js"), "utf8");
const languageScript = fs.readFileSync(path.join(__dirname, "assets/control-language.js"), "utf8");
const base = "/api/v1/workspaces";

class Element {
  constructor(tag = "div") {
    this.tag = tag; this.children = []; this.attributes = {}; this.listeners = {};
    this.disabled = false; this._text = ""; this._value = "";
  }
  get textContent() { return this._text + this.children.map((c) => c.textContent).join(""); }
  set textContent(value) { this._text = String(value); this.children = []; }
  get options() { return this.children; }
  get value() { return this._value; }
  set value(value) {
    this._value = this.tag !== "select" || this.children.some((c) => c.value === String(value)) ? String(value) : "";
  }
  append(child) { this.children.push(child); }
  replaceChildren(...children) { this.children = children; this._text = ""; this._value = ""; }
  getAttribute(key) { return this.attributes[key] ?? null; }
  setAttribute(key, value) { this.attributes[key] = value; }
  addEventListener(event, callback, { signal } = {}) {
    this.listeners[event] = callback;
    signal?.addEventListener("abort", () => {
      if (this.listeners[event] === callback) delete this.listeners[event];
    }, { once: true });
  }
  async fire(event, details) { await this.listeners[event]?.(details); }
}
const response = (data, status = 200, type = "application/json") => ({
  ok: status >= 200 && status < 300, status,
  headers: { get: () => type }, json: async () => data,
});
const deferred = () => {
  let resolve, reject;
  const promise = new Promise((yes, no) => { resolve = yes; reject = no; });
  return { promise, resolve, reject };
};
async function settle() { for (let i = 0; i < 8; i++) await new Promise(setImmediate); }
function start(routes = {}, storage = new Map()) {
  const nodes = Object.fromEntries(["control-language", "control-language-code", "workspace", "configuration", "versions", "version-count", "workspace-meta", "configuration-meta", "health-status", "health-time", "browser", "refresh", "data-status"].map((id) => [id, new Element(["workspace", "configuration"].includes(id) ? "select" : "div")]));
  nodes.refresh.setAttribute("aria-label", "Refresh");
  nodes.refresh.setAttribute("data-i18n-label", "");
  nodes.refresh.setAttribute("data-tooltip", "Refresh");
  const body = new Element("body"), document = new EventTarget(), window = new EventTarget();
  const documentEvent = (type, target) => {
    const event = new Event(type);
    Object.defineProperty(event, "target", { value: target });
    document.dispatchEvent(event);
  };
  Object.assign(document, {
    body, activeElement: body, documentElement: {},
    hasFocus: () => true,
    getElementById: (id) => nodes[id],
    createElement: (tag) => new Element(tag),
    querySelectorAll(selector) {
      const visit = (node) => [node, ...node.children.flatMap(visit)];
      return Object.values(nodes).flatMap(visit).filter((node) => Object.hasOwn(node.attributes, selector.slice(1, -1)));
    },
  });
  for (const node of Object.values(nodes)) {
    node.focus = () => {
      if (node.disabled) return;
      document.activeElement = node;
      documentEvent("focusin", node);
    };
  }
  nodes.refresh.contains = (node) => node === nodes.refresh;
  let disabled = false;
  Object.defineProperty(nodes.refresh, "disabled", {
    get: () => disabled,
    set: (value) => {
      disabled = value;
      if (value && document.activeElement === nodes.refresh && !document.delayDisabledBlur) document.activeElement = body;
    },
  });
  const requests = [], timers = new Map(), windowEvents = {}; let timerID = 0;
  const addWindowListener = window.addEventListener.bind(window);
  window.addEventListener = (event, callback, options) => {
    if (event === "pageshow") windowEvents[event] = callback;
    addWindowListener(event, callback, options);
  };
  const environment = vm.createContext({
    window, document,
    localStorage: { getItem: (key) => storage.get(key), setItem: (key, value) => storage.set(key, value) },
    AbortController,
    setTimeout(callback) { const key = ++timerID; timers.set(key, callback); return key; },
    clearTimeout(key) { timers.delete(key); },
    fetch: async (url, options) => {
      requests.push({ url, options });
      const route = Object.hasOwn(routes, url) ? routes[url] : url === "/health" ? { status: "ok" } : [];
      if (route instanceof Error) throw route;
      const result = typeof route === "function" ? await route(options) : route;
      return result && typeof result.json === "function" ? result : response(result);
    },
  });
  vm.runInContext(languageScript, environment, { filename: "control-language.js" });
  vm.runInContext(script, environment, { filename: "dashboard.js" });
  return {
    nodes, requests, routes, timers, windowEvents, document, documentEvent,
    text: () => Object.values(nodes).map((n) => n.textContent).join("\n"),
    select: async (id, value) => { nodes[id].value = value; await nodes[id].fire("change"); },
    refresh: () => nodes.refresh.disabled ? Promise.resolve() : nodes.refresh.fire("click"),
  };
}
function populated(extra = {}) {
  return start({
    [base]: [{ id: 1, name: "workspace" }, { id: 2, name: "other" }],
    [`${base}/1/configurations`]: [{ id: 3, workspaceId: 1, name: "listener" }],
    [`${base}/1/configurations/3/versions`]: [{ id: 4, configurationId: 3, number: 2, state: "Published", listener: { host: "::1", port: 8080 }, updatedAt: "2026-11-15T08:00:00Z" }],
    ...extra,
  });
}

test("initial load is GET-only, empty is explicit, no polling, Refresh performs new reads", async () => {
  const app = start();
  assert.match(app.nodes["data-status"].textContent, /Loading/);
  assert.equal(app.nodes.browser.attributes["aria-busy"], "true");
  await settle();
  assert.equal(app.nodes["health-status"].textContent, "Available");
  assert.match(app.nodes["data-status"].textContent, /No workspaces/);
  assert.equal(app.nodes.workspace.disabled, true);
  assert.equal(app.nodes.refresh.disabled, false);
  assert.equal(app.timers.size, 0);
  assert.equal(app.requests.length, 2);
  await settle(); assert.equal(app.requests.length, 2);
  await app.refresh(); assert.equal(app.requests.length, 4);
  for (const request of app.requests) {
    assert.equal(request.options.method, "GET");
    assert.equal(request.options.mode, "same-origin");
    assert.equal(request.options.cache, "no-store");
    assert.equal(request.options.redirect, "error");
    assert.equal(request.options.body, undefined);
  }
});

test("nested selection and Refresh preserve IDs, show actual versions, never display secrets or raw JSON", async () => {
  const app = populated({
    [`${base}/1/configurations/3/versions`]: [{ id: 4, configurationId: 3, number: 2, state: "Published", listener: { host: "<svg onload=alert(1)>", port: 8080, tls: { certificateRef: "private-certificate", privateKeyRef: "private-key" } }, authentication: { secretRef: "private-auth" }, routing: { handlerRef: "private-route" }, updatedAt: "2026-11-15T08:00:00Z" }],
  });
  await settle();
  await app.select("workspace", "1"); await app.select("configuration", "3");
  assert.equal(app.nodes.configuration.value, "3");
  assert.match(app.nodes.versions.textContent, /v2 · ID 4Published<svg onload=alert\(1\)> \/ 8080/);
  assert.match(app.nodes["version-count"].textContent, /^1 versions/);
  assert.doesNotMatch(app.text(), /private-(certificate|key|auth|route)/);
  assert.equal(app.nodes.versions.children[0].children[2].children.length, 0, "host is only text, never parsed markup");
  await app.refresh();
  assert.equal(app.nodes.workspace.value, "1"); assert.equal(app.nodes.configuration.value, "3");
  assert.equal(app.requests.at(-1).url, `${base}/1/configurations/3/versions`);
  assert.ok(app.requests.every((r) => r.options.method === "GET"));
  app.routes[`${base}/1/configurations`] = [];
  await app.refresh();
  assert.equal(app.nodes.configuration.disabled, true);
  assert.equal(app.nodes.configuration.value, "");
  assert.doesNotMatch(app.nodes.versions.textContent, /Published/);
});

test("empty nested lists and missing listener/timestamp are honest unavailable states", async () => {
  const app = populated({ [`${base}/1/configurations/3/versions`]: [] });
  await settle(); await app.select("workspace", "2");
  assert.match(app.nodes.configuration.textContent, /No configurations/);
  await app.select("workspace", "1"); await app.select("configuration", "3");
  assert.match(app.nodes.versions.textContent, /No versions/);
  app.routes[`${base}/1/configurations/3/versions`] = [{ id: 4, configurationId: 3, number: 1, state: "Draft" }];
  await app.refresh();
  assert.match(app.nodes.versions.textContent, /DraftUnavailableUnavailable/);
});

test("health failures never imply runtime availability and do not hide successful workspace reads", async () => {
  for (const bad of [new TypeError("private-health-failure"), { status: "running", secretRef: "private-health-data" }, response({}, 503)]) {
    const app = populated({ "/health": bad }); await settle();
    assert.notEqual(app.nodes["health-status"].textContent, "Available");
    assert.match(app.nodes.workspace.textContent, /workspace · ID 1/);
    assert.doesNotMatch(app.text(), /private-health|running/);
    assert.match(app.nodes["workspace-meta"].textContent, /2 workspaces/);
  }
});

test("a hanging health check cannot turn completed workspace data into a data timeout", async () => {
  const app = populated({ "/health": ({ signal }) => new Promise((_, reject) => signal.addEventListener("abort", () => reject(new Error("aborted")))) });
  await settle();
  assert.match(app.nodes["data-status"].textContent, /Data loaded/);
  assert.equal(app.timers.size, 1, "completed data cycle has cancelled its deadline; only health remains");
  for (const callback of [...app.timers.values()]) callback();
  await settle();
  assert.match(app.nodes["health-status"].textContent, /Not confirmed/);
  assert.match(app.nodes["data-status"].textContent, /Data loaded/);
  assert.match(app.nodes.workspace.textContent, /workspace · ID 1/);
  assert.equal(app.nodes.refresh.disabled, false);
});

test("duplicate IDs reject the whole entity response without partial rendering", async () => {
  const app = start({ [base]: [{ id: 1, name: "first-duplicate" }, { id: 1, name: "second-duplicate" }] });
  await settle();
  assert.match(app.nodes["data-status"].textContent, /Unsupported/);
  assert.doesNotMatch(app.text(), /first-duplicate|second-duplicate/);
  assert.equal(app.nodes.workspace.disabled, true);
});

for (const [label, bad, message] of [
  ["offline", new TypeError("do-not-display-private-network-error"), /Offline \/ unreachable/],
  ["HTTP error", response({ secretRef: "do-not-display-secret" }, 500), /HTTP 500/],
  ["unsupported shape", { secretRef: "do-not-display-secret" }, /Unsupported response/],
  ["unsupported content type", response([], 200, "text/html"), /Unsupported response/],
  ["malformed JSON", { ok: true, status: 200, headers: { get: () => "application/json" }, json: async () => { throw new Error("do-not-display-private-json"); } }, /Unsupported response/],
]) {
  test(`${label} clears previous values and does not reveal raw errors`, async () => {
    const app = populated(); await settle(); await app.select("workspace", "1"); await app.select("configuration", "3");
    app.routes[base] = bad; await app.refresh();
    assert.match(app.nodes["data-status"].textContent, message);
    assert.doesNotMatch(app.text(), /do-not-display|Published|listener · ID/);
    assert.equal(app.nodes.workspace.disabled, true);
    assert.equal(app.nodes.configuration.disabled, true);
    assert.equal(app.nodes.refresh.disabled, false);
    assert.equal(app.nodes["health-status"].textContent, "Available", "independent health is not Runtime/data status");
  });
}

test("unsafe uint64 IDs reject entire affected list before display or dependent requests", async () => {
  const unsafe = JSON.parse("18446744073709551615");
  const app = start({ [base]: [{ id: 1, name: "valid-but-no-partial-render" }, { id: unsafe, name: "unsafe-name" }] });
  await settle();
  assert.match(app.nodes["data-status"].textContent, /Unsupported/);
  assert.doesNotMatch(app.text(), /unsafe-name|valid-but-no-partial-render|184467/);
  assert.equal(app.requests.filter((r) => r.url.startsWith(base)).length, 1);
  for (const [key, value] of [["id", unsafe], ["workspaceId", unsafe], ["workspaceId", 2]]) {
    app.routes[base] = [{ id: 1, name: "workspace" }]; await app.refresh();
    app.routes[`${base}/1/configurations`] = [{ id: 3, workspaceId: 1, name: "bad-configuration", [key]: value }];
    await app.select("workspace", "1");
    assert.match(app.nodes["data-status"].textContent, /Unsupported/);
    assert.doesNotMatch(app.text(), /bad-configuration/);
    assert.ok(!app.requests.some((r) => r.url.includes("/versions")));
  }
});

test("invalid versions, parent mismatches and unsafe IDs never produce partial tables", async () => {
  const app = populated(); await settle(); await app.select("workspace", "1");
  for (const override of [{ id: Number.MAX_SAFE_INTEGER + 1 }, { configurationId: 99 }, { state: "Running" }, { number: 4294967296 }, { listener: { host: "host", port: -1 } }]) {
    app.routes[`${base}/1/configurations/3/versions`] = [
      { id: 4, configurationId: 3, number: 1, state: "Draft" },
      { id: 5, configurationId: 3, number: 2, state: "Published", ...override },
    ];
    await app.select("configuration", "3");
    assert.match(app.nodes["data-status"].textContent, /Unsupported/);
    assert.doesNotMatch(app.nodes.versions.textContent, /Draft|Published|Running/);
  }
});

test("late response for old selection cannot replace new workspace data", async () => {
  const late = deferred();
  const app = populated({ [`${base}/1/configurations`]: () => late.promise, [`${base}/2/configurations`]: [{ id: 8, workspaceId: 2, name: "current" }] });
  await settle();
  app.nodes.workspace.value = "1";
  const old = app.nodes.workspace.fire("change");
  await settle(); await app.select("workspace", "2");
  late.resolve([{ id: 3, workspaceId: 1, name: "stale" }]); await old;
  assert.match(app.nodes.configuration.textContent, /current/);
  assert.doesNotMatch(app.nodes.configuration.textContent, /stale/);
  assert.equal(app.nodes.refresh.disabled, false);
});

test("focused manual Refresh restores focus after disabling and ignores duplicate activation", async () => {
  const app = start();
  await settle();
  const { document, nodes, requests } = app;
  const response = deferred();
  app.routes[base] = () => response.promise;
  nodes.refresh.focus();
  const before = requests.length;
  const pending = app.refresh();
  assert.equal(document.activeElement, document.body);
  assert.equal(nodes.refresh.disabled, true);
  await app.refresh();
  assert.equal(requests.length - before, 2);
  response.resolve([]);
  await pending;
  assert.equal(document.activeElement, nodes.refresh);
  assert.equal(nodes.refresh.disabled, false);
});

test("Refresh restores after a browser defers the blur caused by disabling", async () => {
  const app = start();
  await settle();
  const { document, nodes } = app;
  const response = deferred();
  document.delayDisabledBlur = true;
  app.routes[base] = () => response.promise;
  nodes.refresh.focus();
  const pending = app.refresh();
  assert.equal(document.activeElement, nodes.refresh);
  document.activeElement = document.body;
  await nodes.refresh.fire("blur", { relatedTarget: null });
  response.resolve([]);
  await pending;
  assert.equal(document.activeElement, nodes.refresh);
});

test("Refresh keeps intervening focus and does not restore after a later blur or outside pointer", async () => {
  for (const exit of ["focus", "focus-and-blur", "pointer"]) {
    const app = start();
    await settle();
    const { document, nodes, documentEvent } = app;
    const response = deferred();
    app.routes[base] = () => response.promise;
    nodes.refresh.focus();
    const pending = app.refresh();
    if (exit === "pointer") {
      documentEvent("pointerdown", document.body);
    } else {
      nodes["control-language"].focus();
      if (exit === "focus-and-blur") document.activeElement = document.body;
    }
    response.resolve([]);
    await pending;
    assert.equal(document.activeElement, exit === "focus" ? nodes["control-language"] : document.body, exit);
  }
});

test("Refresh can restore after an error but a superseded request cannot reclaim focus", async () => {
  const app = populated();
  await settle();
  const { document, nodes } = app;
  app.routes[base] = new Error("offline");
  nodes.refresh.focus();
  await app.refresh();
  assert.equal(document.activeElement, nodes.refresh);
  const response = deferred();
  app.routes[base] = () => response.promise;
  const pending = app.refresh();
  await app.select("configuration", "");
  response.resolve([]);
  await pending;
  assert.equal(document.activeElement, document.body);
});

test("request deadline aborts pending fetch and offers manual retry without automatic requests", async () => {
  const app = start({ [base]: ({ signal }) => new Promise((_, reject) => signal.addEventListener("abort", () => reject(new Error("aborted")))) });
  await settle();
  assert.equal(app.nodes.refresh.disabled, true);
  for (const callback of [...app.timers.values()]) callback();
  await settle();
  assert.match(app.nodes["data-status"].textContent, /timed out/);
  assert.doesNotMatch(app.nodes.workspace.textContent, /Loading/);
  assert.equal(app.nodes.refresh.disabled, false);
  assert.equal(app.requests.length, 2);
  app.routes[base] = [];
  await app.refresh();
  assert.match(app.nodes["data-status"].textContent, /No workspaces/);
});

test("locale updates loaded Dashboard rows, selections and feedback without network requests", async () => {
  const storage = new Map([["uwp-control-language", "ru"]]);
  const app = start({
    [base]: [{ id: 1, name: "Published" }],
    [`${base}/1/configurations`]: [{ id: 3, workspaceId: 1, name: "Draft" }],
    [`${base}/1/configurations/3/versions`]: [{ id: 4, configurationId: 3, number: 2, state: "Published",
      listener: { host: "::1", port: 8080 }, updatedAt: "2026-11-15T08:00:00Z" }],
  }, storage);
  await settle();
  await app.select("workspace", "1");
  await app.select("configuration", "3");
  const { nodes, document, requests } = app;
  assert.equal(document.documentElement.lang, "ru");
  assert.equal(nodes.refresh.getAttribute("aria-label"), "Обновить");
  assert.equal(nodes.refresh.getAttribute("data-tooltip"), "Обновить");
  const row = nodes.versions.children[0];
  const option = nodes.workspace.options[1];
  const before = requests.length;
  const russianDate = row.children[3].textContent;
  assert.match(nodes["health-status"].textContent, /Доступно/);
  assert.match(nodes["version-count"].textContent, /Версий в выбранной конфигурации: 1/);
  assert.equal(row.children[1].textContent, "Опубликовано");
  await app.nodes["control-language"].fire("click");
  assert.equal(row.children[1].textContent, "Published");
  assert.notEqual(row.children[3].textContent, russianDate);
  await app.nodes["control-language"].fire("click");
  assert.equal(row.children[3].textContent, russianDate);
  assert.equal(nodes.versions.children[0], row);
  assert.equal(nodes.workspace.options[1], option);
  assert.equal(option.textContent, "Published · ID 1");
  assert.equal(nodes.configuration.options[1].textContent, "Draft · ID 3");
  assert.equal(nodes.workspace.value, "1");
  assert.equal(nodes.configuration.value, "3");
  assert.equal(row.children[2].textContent, "::1 / 8080");
  assert.equal(requests.length, before);
  assert.equal(storage.get("uwp-control-language"), "ru");
});

test("restoring a Dashboard from the back-forward cache applies the newest preference in place", async () => {
  const storage = new Map();
  const app = start({}, storage);
  await settle();
  const count = app.requests.length;
  storage.set("uwp-control-language", "ru");
  app.windowEvents.pageshow({ persisted: true });
  assert.equal(app.document.documentElement.lang, "ru");
  assert.equal(app.nodes["control-language-code"].textContent, "RU");
  assert.match(app.nodes["data-status"].textContent, /абочих областей нет/);
  assert.equal(app.requests.length, count);
});
