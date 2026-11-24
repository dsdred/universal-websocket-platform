"use strict";
const { readFileSync } = require("node:fs");
const path = require("node:path");
const vm = require("node:vm");
const assert = require("node:assert/strict");
const test = require("node:test");

const script = readFileSync(path.join(__dirname, "assets/control-navigation.js"), "utf8");

function load(matches, nativeScrollbar = 0) {
  class Target {
    listeners = new Map();
    addEventListener(type, handler) { this.listeners.set(type, handler); }
    emit(type, props = {}) {
      const event = { defaultPrevented: false, preventDefault() { this.defaultPrevented = true; }, ...props };
      this.listeners.get(type)(event);
      return event;
    }
  }
  class Element extends Target {
    attributes = new Map();
    style = new Map();
    hidden = false;
    setAttribute(name, value) { this.attributes.set(name, value); }
    getAttribute(name) { return this.attributes.get(name); }
    focus() { document.activeElement = this; }
    contains(node) {
      for (let parent = node; parent; parent = parent.parentElement) {
        if (parent === this) return true;
      }
      return false;
    }
    append(node) {
      if (node.contains(document.activeElement)) document.activeElement = null;
      node.parentElement = this;
    }
  }
  const document = new Target();
  const window = new Target();
  const media = new Target();
  media.matches = matches;
  window.matchMedia = () => media;
  window.location = { href: "http://localhost/" };
  const menu = new Element(), nav = new Element(), language = new Element(),
    themeButton = new Element(), preferences = new Element(),
    desktop = new Element(), slot = new Element(), link = new Element();
  preferences.append(language);
  preferences.append(themeButton);
  desktop.append(preferences);
  nav.append(link);
  nav.querySelector = () => link;
  const nodes = {
    "navigation-toggle": menu, "control-navigation": nav,
    "control-preferences": preferences, "mobile-preferences-slot": slot,
  };
  document.getElementById = (id) => nodes[id];
  desktop.style.setProperty = desktop.style.set.bind(desktop.style);
  desktop.style.removeProperty = desktop.style.delete.bind(desktop.style);
  window.innerWidth = 1280;
  document.documentElement = { clientWidth: window.innerWidth };
  document.body = new Element();
  document.createElement = () => ({
    style: {},
    offsetWidth: 100,
    clientWidth: 100 - nativeScrollbar,
    contains: () => false,
    remove() {},
  });
  let observeResize;
  class ResizeObserver {
    constructor(callback) { observeResize = callback; }
    observe(target) { assert.equal(target, document.documentElement); }
  }
  vm.runInNewContext(script, { document, window, ResizeObserver });
  function updateScrollbar(occupiedWidth) {
    document.documentElement.clientWidth = window.innerWidth - occupiedWidth;
    observeResize();
  }
  function resize(narrow) {
    media.matches = narrow;
    media.emit("change");
  }
  function clickLink(href, props = {}) {
    return nav.emit("click", { button: 0, target: { closest: () => ({ href }) }, ...props });
  }
  return { document, window, menu, nav, language, themeButton, preferences, desktop, slot, link, resize, clickLink, updateScrollbar };
}

test("mobile disclosure starts closed and Escape restores focus after opening", () => {
  const { document, menu, nav, preferences, desktop, slot, link } = load(true);
  assert.equal(preferences.parentElement, slot);
  assert.equal(desktop.hidden, true);
  assert.equal(menu.hidden, false);
  assert.equal(menu.getAttribute("aria-expanded"), "false");
  assert.equal(nav.hidden, true);
  menu.focus();
  menu.emit("click");
  assert.equal(menu.getAttribute("aria-expanded"), "true");
  assert.equal(nav.hidden, false);
  assert.equal(document.activeElement, menu);
  link.focus();
  assert.equal(document.emit("keydown", { key: "Escape" }).defaultPrevented, true);
  assert.equal(menu.getAttribute("aria-expanded"), "false");
  assert.equal(nav.hidden, true);
  assert.equal(document.activeElement, menu);
});

test("mobile activation closes, retaining ordinary navigation and suppressing exact self reload", () => {
  const { document, menu, nav, link, clickLink, window } = load(true);
  menu.emit("click");
  assert.equal(clickLink("http://localhost/dashboard", { ctrlKey: true }).defaultPrevented, false);
  assert.equal(nav.hidden, false);
  link.focus();
  assert.equal(clickLink("http://localhost/dashboard").defaultPrevented, false);
  assert.equal(document.activeElement, menu);
  assert.equal(nav.hidden, true);
  menu.emit("click");
  assert.equal(clickLink(window.location.href).defaultPrevented, true);
  assert.equal(nav.hidden, true);
});

test("breakpoints move the original preferences and preserve focus on either control", () => {
  const { document, window, menu, nav, language, themeButton, preferences, desktop, slot, link, resize } = load(false);
  assert.equal(menu.hidden, true);
  assert.equal(nav.hidden, false);
  language.focus();
  resize(true);
  assert.equal(preferences.parentElement, slot);
  assert.equal(document.activeElement, language);
  themeButton.focus();
  resize(false);
  assert.equal(preferences.parentElement, desktop);
  assert.equal(document.activeElement, themeButton);
  resize(true);
  assert.equal(document.activeElement, themeButton);
  menu.focus();
  resize(false);
  assert.equal(preferences.parentElement, desktop);
  assert.equal(desktop.hidden, false);
  assert.equal(menu.hidden, true);
  assert.equal(nav.hidden, false);
  assert.equal(document.activeElement, link);
  resize(true);
  assert.equal(document.activeElement, menu);
  menu.emit("click");
  window.emit("pageshow", { persisted: true });
  assert.equal(nav.hidden, true);
});

test("desktop header fills only the missing native scrollbar gutter", () => {
  const { desktop, resize, updateScrollbar } = load(false, 19);
  const inset = "--control-header-end-inset";
  assert.equal(desktop.style.get(inset), "19px");
  updateScrollbar(19);
  assert.equal(desktop.style.has(inset), false);
  updateScrollbar(0);
  assert.equal(desktop.style.get(inset), "19px");
  resize(true);
  assert.equal(desktop.style.has(inset), false);
  updateScrollbar(0);
  assert.equal(desktop.style.has(inset), false);
  resize(false);
  assert.equal(desktop.style.get(inset), "19px");
});

test("overlay scrollbars reserve no space in the desktop header", () => {
  const { desktop } = load(false);
  assert.equal(desktop.style.get("--control-header-end-inset"), "0px");
});
