"use strict";
(() => {
  const menu = document.getElementById("navigation-toggle");
  const nav = document.getElementById("control-navigation");
  const language = document.getElementById("control-language");
  const desktopHeader = language.parentElement;
  const mobileLanguageSlot = document.getElementById("mobile-language-slot");
  const mobile = window.matchMedia("(max-width: 700px)");
  const expanded = () => menu.getAttribute("aria-expanded") === "true";

  function setExpanded(open) {
    menu.setAttribute("aria-expanded", String(open));
    nav.hidden = mobile.matches && !open;
  }

  function closeMenu() {
    if (nav.contains(document.activeElement)) menu.focus();
    setExpanded(false);
  }

  function syncLayout() {
    const focused = document.activeElement;
    const destination = mobile.matches ? mobileLanguageSlot : desktopHeader;
    if (!mobile.matches) desktopHeader.hidden = false;
    if (language.parentElement !== destination) destination.append(language);
    desktopHeader.hidden = mobile.matches;
    if (mobile.matches) {
      menu.hidden = false;
      closeMenu();
    } else {
      setExpanded(false);
      if (focused === menu) nav.querySelector("a").focus();
      menu.hidden = true;
    }
    if (focused === language && document.activeElement !== language) language.focus();
  }

  menu.addEventListener("click", () => {
    if (!mobile.matches) return;
    if (expanded()) closeMenu();
    else setExpanded(true);
  });
  document.addEventListener("keydown", (event) => {
    if (event.key !== "Escape" || event.defaultPrevented || !expanded()) return;
    event.preventDefault();
    menu.focus();
    closeMenu();
  });
  nav.addEventListener("click", (event) => {
    if (!mobile.matches || event.defaultPrevented || event.button !== 0 ||
        event.altKey || event.ctrlKey || event.metaKey || event.shiftKey) return;
    const link = event.target.closest("a[href]");
    if (!link) return;
    if (link.href === window.location.href) event.preventDefault();
    closeMenu();
  });
  mobile.addEventListener("change", syncLayout);
  window.addEventListener("pageshow", (event) => {
    if (event.persisted) syncLayout();
  });
  syncLayout();
})();
