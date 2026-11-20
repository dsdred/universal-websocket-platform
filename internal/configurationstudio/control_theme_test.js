"use strict";
const { readFileSync } = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const assert = require("node:assert/strict");
const test = require("node:test");

const script = readFileSync(path.join(__dirname, "assets/control-theme.js"), "utf8");
const preferenceKey = "uwp-control-theme";

function boot({ saved = null, osDark = false, readBlocked = false, writeBlocked = false } = {}) {
  const select = new EventTarget();
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
    return select;
  };
  media.matches = osDark;
  window.matchMedia = (query) => {
    assert.equal(query, "(prefers-color-scheme: dark)");
    return media;
  };
  window.localStorage = localStorage;
  vm.runInNewContext(script, { window, document });

  return {
    select, dataset, writes,
    ready() { emit(document, "DOMContentLoaded"); },
    choose(value) { select.value = value; emit(select, "change"); },
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

test("boot applies saved or System colors before the control or styles initialize", () => {
  for (const [saved, osDark, effective, choice] of [
    [null, false, "light", "system"],
    ["system", true, "dark", "system"],
    ["dark", false, "dark", "dark"],
    ["light", true, "light", "light"],
    ["invalid", true, "dark", "system"],
  ]) {
    const app = boot({ saved, osDark });
    assert.equal(app.dataset.controlTheme, effective);
    assert.equal(app.select.value, undefined);
    app.ready();
    assert.equal(app.select.value, choice);
    assert.deepEqual(app.writes, []);
  }
});

test("selections persist immediately and only System follows OS changes", () => {
  const app = boot();
  app.ready();
  app.os(true);
  assert.equal(app.dataset.controlTheme, "dark");
  app.choose("light");
  app.os(false);
  app.os(true);
  assert.equal(app.dataset.controlTheme, "light");
  app.choose("dark");
  app.os(false);
  assert.equal(app.dataset.controlTheme, "dark");
  app.choose("system");
  assert.equal(app.select.value, "system");
  assert.equal(app.dataset.controlTheme, "light");
  app.os(true);
  assert.equal(app.dataset.controlTheme, "dark");
  assert.deepEqual(app.writes, ["light", "dark", "system"]);
});

test("cross-tab storage changes refresh the current choice without writing back", () => {
  const app = boot({ osDark: true });
  app.external("light");
  app.ready();
  assert.equal(app.select.value, "light");
  app.external("dark", "uwp-control-theme", {});
  app.external("dark", "uwp-control-language");
  assert.equal(app.select.value, "light");
  app.external(null, null);
  assert.equal(app.select.value, "system");
  assert.equal(app.dataset.controlTheme, "dark");
  assert.deepEqual(app.writes, []);
});

test("BFCache restoration rechecks storage and System media state", () => {
  const app = boot({ saved: "light" });
  app.ready();
  app.restore("dark", false);
  assert.equal(app.select.value, "light");
  app.restore("dark");
  assert.equal(app.select.value, "dark");
  app.os(true, false);
  app.restore("system");
  assert.equal(app.select.value, "system");
  assert.equal(app.dataset.controlTheme, "dark");
  app.os(false);
  assert.equal(app.dataset.controlTheme, "light");
});

test("blocked storage preserves the page choice and leaves System usable", () => {
  for (const readBlocked of [false, true]) {
    const app = boot({ saved: "system", readBlocked, writeBlocked: true });
    app.ready();
    app.choose("dark");
    app.restore("system");
    assert.equal(app.select.value, "dark");
    assert.equal(app.dataset.controlTheme, "dark");
    app.choose("system");
    app.os(true);
    app.os(false);
    assert.equal(app.dataset.controlTheme, "light");
    assert.deepEqual(app.writes, []);
  }
});
