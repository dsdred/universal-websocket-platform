"use strict";
(() => {
  const preferenceKey = "uwp-control-theme";
  const systemDark = window.matchMedia("(prefers-color-scheme: dark)");
  const root = document.documentElement;
  const normalize = (value) => value === "light" || value === "dark" ? value : "system";

  function savedPreference() {
    try { return normalize(window.localStorage.getItem(preferenceKey)); }
    catch (_) { return null; }
  }

  let preference = savedPreference() ?? "system";
  let radios = [];
  let pageOnlyPreference = false;

  function apply() {
    const effective = preference === "system" ? (systemDark.matches ? "dark" : "light") : preference;
    root.dataset.controlTheme = effective;
    for (const radio of radios) radio.checked = radio.value === preference;
  }

  function syncPreference() {
    const saved = savedPreference();
    if (saved !== null) {
      preference = saved;
      pageOnlyPreference = false;
    }
    apply();
  }

  apply();
  document.addEventListener("DOMContentLoaded", () => {
    radios = document.getElementById("control-theme").querySelectorAll('input[type="radio"]');
    apply();
    for (const radio of radios) {
      radio.addEventListener("change", () => {
        preference = normalize(radio.value);
        try {
          window.localStorage.setItem(preferenceKey, preference);
          pageOnlyPreference = false;
        } catch (_) { pageOnlyPreference = true; }
        apply();
      });
    }
  }, { once: true });
  systemDark.addEventListener("change", () => {
    if (preference === "system") apply();
  });
  window.addEventListener("storage", (event) => {
    if (event.key !== preferenceKey && event.key !== null) return;
    try {
      if (event.storageArea !== window.localStorage) return;
    } catch (_) { return; }
    syncPreference();
  });
  window.addEventListener("pageshow", (event) => {
    if (!event.persisted) return;
    if (pageOnlyPreference) apply();
    else syncPreference();
  });
})();
