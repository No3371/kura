// 外觀 (footer): 跟系統 / 日間 / 夜間. The choice lives in localStorage and the
// inline script in index.html applies it before first paint; this only wires
// the switch. 跟系統 removes data-theme so kura.css's prefers-color-scheme rule
// decides. A ?theme= link still wins for that page view, and the switch shows it.
const KEY = "kura-theme";
const root = document.documentElement;

export function initTheme() {
  const group = document.querySelector(".k-theme");
  const current = root.dataset.theme === "light" || root.dataset.theme === "dark" ? root.dataset.theme : "system";
  group.querySelector(`input[value="${current}"]`).checked = true;
  group.addEventListener("change", (e) => {
    const v = e.target.value;
    if (v === "system") delete root.dataset.theme;
    else root.dataset.theme = v;
    // Storage can be blocked (private mode, site data off): the switch still
    // works for this page, it just isn't remembered.
    try {
      if (v === "system") localStorage.removeItem(KEY);
      else localStorage.setItem(KEY, v);
    } catch {}
  });
}
