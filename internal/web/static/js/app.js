import { api, ymd } from "./util.js";
import { renderHome } from "./home.js";
import { renderDetail } from "./detail.js";
import { renderModel } from "./model.js";
import { renderBrowse } from "./browse.js";
import { renderMethodology, renderPending } from "./pages.js";
import { go, keepViewState, markRendered, moved, rememberScroll, setRenderer, viewShown } from "./nav.js";
import { initAdvisor } from "./advisor.js";
import { initNotify } from "./notify.js";
import { initTheme } from "./theme.js";

// [path pattern, view, default document title]. A view is called as
// view(el, overview, ...pathParams) with the pattern's capture groups still
// URI-encoded; it renders into el, may await data, and may return a document
// title that replaces the default. The router swaps el into <main> and sets
// the title only if no newer navigation started meanwhile. A view that throws
// an error with status 404 gets the not-found view.
const routes = [
  [/^\/$/, renderHome, "Kura 蔵"],
  [/^\/browse$/, renderBrowse, "價格最美 · Kura 蔵"],
  [/^\/listing\/([^/]+)$/, renderDetail, "商品頁 · Kura 蔵"],
  [/^\/model\/([^/]+)\/([^/]+)$/, renderModel, "型號頁 · Kura 蔵"],
  [/^\/methodology$/, renderMethodology, "行情怎麼算 · Kura 蔵"],
  [/^\/(about|faq|contact|privacy|terms)$/, renderPending, "Kura 蔵"],
];

const NOT_FOUND_TITLE = "找不到這個頁面 · Kura 蔵";

const main = document.getElementById("main");

// The first load shares one request among the FX line, the advisor and the
// first view; each later navigation asks again (fresh), since a crawl running
// while the page is open changes the counts.
let overviewP;
function overview(fresh = false) {
  if (fresh || !overviewP) {
    const p = api("/api/overview").catch((err) => {
      if (overviewP === p) overviewP = undefined;
      throw err;
    });
    overviewP = p;
  }
  return overviewP;
}

function renderNotFound(el) {
  el.innerHTML = `<section class="k-notfound">
    <h1>找不到這個頁面</h1>
    <p><a href="/" class="k-link">回到首頁</a></p>
  </section>`;
}

// After an in-app navigation the element that had focus is gone with the old
// view, so focus falls to <body>: the next Tab restarts at the logo and a
// screen reader announces nothing. Focus moves to the new view's heading (the
// load-failure view has none: its message). preventScroll: the router has
// just put the view at the top or, for Back and Forward, where the entry was
// left. The first render is left alone, where the browser's start position
// is what a visitor expects.
// Rejected: focusing <main> — a screen reader then reads the whole view
// instead of its heading.
// Rejected: leaving focus on <body> when Back restores a position below the
// top. The next Tab then starts at the view's first link (the crumb) all the
// same, scrolling back up, and a screen reader announces nothing.
function focusView() {
  const target = main.querySelector("h1, .k-error");
  if (!target) return;
  target.setAttribute("tabindex", "-1");
  target.focus({ preventScroll: true });
}

// The footer link for the page on screen carries aria-current, as the crumbs do.
const footerLinks = document.querySelectorAll(".k-footer nav a");
function markFooter() {
  for (const a of footerLinks) {
    if (new URL(a.href).pathname === location.pathname) a.setAttribute("aria-current", "page");
    else a.removeAttribute("aria-current");
  }
}

let seq = 0;

async function render() {
  const mine = ++seq;
  const key = markRendered();
  let m = null;
  const route = routes.find(([re]) => (m = re.exec(location.pathname)));
  const out = document.createElement("div");
  let title = route ? route[2] : NOT_FOUND_TITLE;
  try {
    const ov = await overview(mine > 1);
    if (route) title = (await route[1](out, ov, ...m.slice(1))) || title;
    else renderNotFound(out);
  } catch (err) {
    if (err.status === 404) {
      renderNotFound(out);
      title = NOT_FOUND_TITLE;
    } else {
      console.warn(err);
      out.innerHTML = `<p class="k-error">暫時無法載入資料，請稍後再試。</p>`;
    }
  }
  // A stale view must neither replace nor retitle a newer page.
  if (mine === seq) {
    main.replaceChildren(...out.childNodes);
    document.title = title;
    markFooter();
    viewShown(key, mine > 1);
    if (mine > 1) focusView();
  }
}

// Paths the server answers with their own files, not the shell: intercepting
// them would render the not-found view instead of the page.
const EXTERNAL = /^\/(legacy|dev|api)(\/|$)/;

setRenderer(render);

document.addEventListener("click", (e) => {
  if (e.defaultPrevented || e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
  const a = e.target.closest("a[href^='/']");
  if (!a || a.target || a.hasAttribute("download")) return;
  const url = new URL(a.href);
  if (url.origin !== location.origin || EXTERNAL.test(url.pathname)) return;
  e.preventDefault();
  go(url.pathname + url.search + url.hash);
});

// The header search opens Browse in-app. A native GET would reload the page
// and drop the advisor conversation, which lives only in memory; the form
// keeps action="/browse" so it still works before this module runs. A search
// starts a fresh Browse state.
document.addEventListener("submit", (e) => {
  if (!e.target.matches("form.k-search")) return;
  e.preventDefault();
  const q = String(new FormData(e.target).get("q") ?? "").trim();
  go(q ? `/browse?q=${encodeURIComponent(q)}` : "/browse");
});

addEventListener("popstate", () => {
  if (moved()) {
    rememberScroll();
    render();
  } else keepViewState();
});

overview().then((ov) => {
  const fx = ov.fx;
  const asOf = fx.Provider === "fallback" ? "預設匯率" : ymd(fx.AsOf);
  document.getElementById("fx").textContent = ` · 匯率 1 JPY = ${fx.Rate.toFixed(4)} TWD（${asOf}）`;
}).catch(() => {});

initAdvisor(overview());
initNotify();
initTheme();

render();
