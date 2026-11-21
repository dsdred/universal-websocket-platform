"use strict";
const { readFileSync } = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const assert = require("node:assert/strict");
const test = require("node:test");

const script = readFileSync(path.join(__dirname, "assets/control-theme.js"), "utf8");
const preferenceKey = "uwp-control-theme";

function boot({ saved = null, osDark = false, readBlocked = false, writeBlocked = false } = {}) {
  const button = new EventTarget();
  let label;
  button.setAttribute = (name, value) => {
    assert.equal(name, "aria-labelledby");
    label = value;
  };
  const document = new EventTarget();
  const window = new EventTarget();
  const media = new EventTarget();
  const dataset = {};
  const writes = [];
  const localStorage = {
    getItem(key) {
      assert.equal(key, preferenceKey);
      if (readBlocked) throw new Error("Storage access denied");
      return saved;
    },
    setItem(key, value) {
      assert.equal(key, preferenceKey);
      if (writeBlocked) throw new Error("Storage write denied");
      writes.push(value);
      saved = value;
    },
  };
  const emit = (target, type, properties) => target.dispatchEvent(Object.assign(new Event(type), properties));
  document.documentElement = { dataset };
  document.getElementById = (id) => {
    assert.equal(id, "control-theme");
    return button;
  };
  media.matches = osDark;
  window.matchMedia = (query) => {
    assert.equal(query, "(prefers-color-scheme: dark)");
    return media;
  };
  window.localStorage = localStorage;
  vm.runInNewContext(script, { window, document });

  return {
    dataset, writes,
    label() { return label; },
    saved() { return saved; },
    ready() { emit(document, "DOMContentLoaded"); },
    toggle() { emit(button, "click"); },
    os(dark, notify = true) {
      media.matches = dark;
      if (notify) emit(media, "change");
    },
    external(value, key = preferenceKey, area = localStorage) {
      if (area === localStorage && (key === preferenceKey || key === null)) saved = value;
      emit(window, "storage", { key, storageArea: area });
    },
    restore(value, persisted = true) {
      saved = value;
      emit(window, "pageshow", { persisted });
    },
  };
}

function assertTheme(app, palette, action) {
  assert.equal(app.dataset.controlTheme, palette);
  assert.equal(app.label(), `theme-switch-${action}`);
}

test("boot applies saved or System colors before the button or styles initialize", () => {
  for (const [saved, osDark, effective, action] of [
    [null, false, "light", "dark"],
    ["system", true, "dark", "light"],
    ["dark", false, "dark", "light"],
    ["light", true, "light", "dark"],
    ["invalid", true, "dark", "light"],
  ]) {
    const app = boot({ saved, osDark });
    assert.equal(app.dataset.controlTheme, effective);
    assert.equal(app.label(), undefined);
    app.ready();
    assertTheme(app, effective, action);
    assert.equal(app.saved(), saved);
    assert.deepEqual(app.writes, []);
  }
});

test("first toggle pins the opposite of the current OS palette and keeps legacy System until clicked", () => {
  for (const saved of [null, "system", "invalid"]) {
    const app = boot({ saved });
    app.ready();
    assertTheme(app, "light", "dark");
    app.os(true);
    assertTheme(app, "dark", "light");
    assert.equal(app.saved(), saved);
    assert.deepEqual(app.writes, []);
    app.toggle();
    assertTheme(app, "light", "dark");
    app.os(false);
    app.os(true);
    assertTheme(app, "light", "dark");
    app.toggle();
    assertTheme(app, "dark", "light");
    app.os(false);
    assertTheme(app, "dark", "light");
    assert.deepEqual(app.writes, ["light", "dark"]);
  }
});

test("cross-tab storage changes refresh the current choice without writing back", () => {
  const app = boot({ osDark: true });
  app.external("light");
  app.ready();
  assertTheme(app, "light", "dark");
  app.external("dark", "uwp-control-theme", {});
  app.external("dark", "uwp-control-language");
  assertTheme(app, "light", "dark");
  app.external("dark");
  assertTheme(app, "dark", "light");
  app.external("system");
  app.os(false);
  assertTheme(app, "light", "dark");
  app.os(true);
  app.external(null, null);
  assertTheme(app, "dark", "light");
  assert.deepEqual(app.writes, []);
});

test("BFCache restoration rechecks storage and System media state", () => {
  const app = boot({ saved: "light" });
  app.ready();
  app.restore("dark", false);
  assertTheme(app, "light", "dark");
  app.restore("dark");
  assertTheme(app, "dark", "light");
  app.os(true, false);
  app.restore("system");
  assertTheme(app, "dark", "light");
  app.os(false);
  assertTheme(app, "light", "dark");
});

test("blocked storage preserves each toggle and keeps the page choice on BFCache restore", () => {
  for (const readBlocked of [false, true]) {
    const app = boot({ saved: "system", readBlocked, writeBlocked: true });
    app.ready();
    app.toggle();
    app.restore("system");
    assertTheme(app, "dark", "light");
    app.os(true);
    app.os(false);
    assertTheme(app, "dark", "light");
    app.toggle();
    assertTheme(app, "light", "dark");
    assert.deepEqual(app.writes, []);
  }
});
