import { api, esc, num } from "./util.js";
import { cardHTML } from "./card.js";
import { EMPTY_HTML } from "./home.js";
import { replaceURL } from "./nav.js";

// Browse (價格最美 and every other filtered list) at /browse. The URL query is
// the state, so a list can be shared and survives reload; filter changes
// replace the history entry rather than adding one, so Back leaves the page
// instead of stepping back through every select. Every count comes from the
// server: the page computes no figure. The deal list is radar's deals=1
// predicate with the deal sort, the same query as Home's 價格最美 shelf and
// its count. Load-more adds no history entry either; the entry's state
// records how many cards are shown, so Back from a listing restores them.

const PAGE = 40; // ten desktop rows of four at 1280; load-more fetches the next PAGE
// The server's largest page: radar.Search answers a larger limit with its
// default of 30, so restoring more cards than this takes several requests.
const API_MAX = 100;
const SORTS = [["deal", "低於行情優先"], ["newest", "最新上架"], ["price_asc", "價格 ↑"], ["price_desc", "價格 ↓"]];
// Canvas labels → a floor on Kura's internal rank, the only scale one filter
// can apply across shops whose grades differ; cards still show each shop's own
// grade. S = 未使用品 and better, A = ランクA (輕微使用感), B = ランクB (使用感).
const CONDS = [["", "任何成色"], ["S", "未使用以上"], ["A", "輕微使用感以上"], ["B", "明顯使用感以上"]];
const FILTER_ICON = `<svg width="16" height="16" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4" aria-hidden="true"><path d="M2 4h12M4.5 8h7M7 12h2"/></svg>`;

let view = null; // { ov, st, seq } for the browse view on screen

// Brands and types offered as options: those with active listings, most
// listings first. Counts are all active listings, not the current filter.
// An empty brand or type is left out: it could only open an empty list.
function brandOpts(ov) {
  const by = ov.overview.by_brand ?? {};
  return (ov.brands ?? []).filter((b) => by[b.name] > 0).sort((a, b) => by[b.name] - by[a.name])
    .map((b) => [b.name, `${b.name}（${num(by[b.name])}）`]);
}

function catOpts(ov) {
  const by = ov.overview.by_category ?? {};
  return (ov.categories ?? []).filter((c) => by[c.key] > 0).sort((a, b) => by[b.key] - by[a.key])
    .map((c) => [c.key, c.zh]);
}

// The state in a query string. One whitelist serves the URL and the form, so
// a value the controls cannot show (an unknown brand, a brand with no active
// listing, a bad sort) falls back to its default instead of filtering
// unseen; the same check keeps such values out of the API request.
function readState(ov, search) {
  const p = new URLSearchParams(search);
  const pick = (k, opts, dflt = "") => (opts.some(([v]) => v === p.get(k)) ? p.get(k) : dflt);
  const max = (p.get("max_twd") ?? "").trim().replace(/^0+/, "");
  return {
    deals: p.get("deals") === "1",
    brand: pick("brand", brandOpts(ov)),
    category: pick("category", catOpts(ov)),
    min_rank: pick("min_rank", CONDS),
    // Nine digits (NT$999,999,999) is far above any listing. Unbounded, a
    // 22-digit budget would print as 1e+21 in the field and the URL, and the
    // API, reading digits only, would ignore it.
    max_twd: /^[1-9]\d{0,8}$/.test(max) ? max : "",
    q: (p.get("q") ?? "").trim(),
    sort: pick("sort", SORTS, "deal"),
  };
}

// Canonical query: fixed key order; defaults omitted, sort=deal among them.
function stateQuery(st) {
  const p = new URLSearchParams();
  if (st.deals) p.set("deals", "1");
  for (const k of ["brand", "category", "min_rank", "max_twd", "q"]) if (st[k]) p.set(k, st[k]);
  if (st.sort !== "deal") p.set("sort", st.sort);
  return p;
}

// The first page and every load-more page share this query, so the pages
// list exactly the matches the count counts.
function apiURL(st, offset, limit = PAGE) {
  const p = stateQuery(st);
  p.set("sort", st.sort);
  p.set("limit", limit);
  p.set("offset", offset);
  return `/api/listings?${p}`;
}

// The canvas names only the deal list (價格最美). Over any other result that
// heading would label non-deals as deals, so the list Home's 看全部最新上架
// opens keeps that section's name and the rest read 全部在售.
function heading(st) {
  if (st.deals) return "價格最美";
  return st.sort === "newest" ? "最新上架" : "全部在售";
}

const cls = (c) => (c ? ` class="${c}"` : "");

function selectHTML(c, label, name, opts, value) {
  const o = opts.map(([v, t]) => `<option value="${esc(v)}"${v === value ? " selected" : ""}>${esc(t)}</option>`).join("");
  return `<label${cls(c)}>${label}<select class="k-field" name="${name}">${o}</select></label>`;
}

// The five non-keyword fields, returned apart so each form places them in its
// own order.
function fieldsHTML(ov, st, wide) {
  const w = (c) => (wide ? c : "");
  return {
    brand: selectHTML(w("k-bf__brand"), "品牌", "brand", [["", "全部品牌"], ...brandOpts(ov)], st.brand),
    cat: selectHTML(w("k-bf__cat"), "類型", "category", [["", "全部類型"], ...catOpts(ov)], st.category),
    cond: selectHTML(w("k-bf__cond"), "最低成色", "min_rank", CONDS, st.min_rank),
    max: `<label${cls(w("k-bf__max"))}>預算上限（NT$）<input class="k-field" name="max_twd" inputmode="numeric" autocomplete="off" placeholder="不限" value="${esc(st.max_twd)}"></label>`,
    sort: selectHTML(w("k-bf__sort"), "排序", "sort", SORTS, st.sort),
  };
}

const keywordHTML = (st, placeholder, c) =>
  `<label${cls(c)}>關鍵字<input class="k-field" type="search" name="q" autocomplete="off" placeholder="${esc(placeholder)}" value="${esc(st.q)}"></label>`;

const switchHTML = (st) =>
  `<label class="k-switch"><input type="checkbox" role="switch" name="deals"${st.deals ? " checked" : ""}>只看低於行情</label>`;

const cellsHTML = (items) => items.map((it) => `<div role="listitem">${cardHTML(it)}</div>`).join("");

const moreHTML = (rest) =>
  rest > 0 ? `<div class="k-bmore"><button type="button" data-browse-more>再看 ${num(rest)} 件</button></div>` : "";

function resultsHTML(data) {
  if (data.total === 0) return `<p class="k-bnone">沒有符合條件的包。</p>`;
  return `<div role="list" class="k-bgrid">${cellsHTML(data.items)}</div>${moreHTML(data.total - data.items.length)}`;
}

// Cards shown when this history entry was left; PAGE for a fresh entry.
function shownBefore() {
  const n = history.state?.shown;
  return Number.isInteger(n) && n > PAGE ? n : PAGE;
}

// The first n matches, in API_MAX-sized requests made together.
async function firstCards(st, n) {
  const reqs = [];
  for (let off = 0; off < n; off += API_MAX) reqs.push(api(apiURL(st, off, Math.min(API_MAX, n - off))));
  const pages = await Promise.all(reqs);
  return { total: pages[0].total, items: pages.flatMap((pg) => pg.items) };
}

export async function renderBrowse(main, ov) {
  if (ov.overview.active === 0) {
    view = null;
    main.innerHTML = EMPTY_HTML;
    return;
  }
  const st = readState(ov, location.search);
  const data = await firstCards(st, shownBefore());
  view = { ov, st, seq: 0 };
  const h = heading(st);
  const wide = fieldsHTML(ov, st, true);
  const narrow = fieldsHTML(ov, st, false);
  // The canvas draws a different form per width. Both are rendered and CSS
  // shows one (display: none also takes the other out of the tab order and
  // the accessibility tree); every change keeps them in step. One form
  // reordered with CSS order was rejected: keyboard order would no longer
  // follow the visual order at one of the widths.
  main.innerHTML = `<div class="k-browse">
  <nav aria-label="位置" class="k-crumb"><a href="/" class="k-link">首頁</a><span aria-hidden="true">/</span><span aria-current="page" data-browse-h>${esc(h)}</span></nav>
  <div class="k-browse__head"><h1 data-browse-h>${esc(h)}</h1><span class="k-browse__count" aria-live="polite" data-browse-count>${num(data.total)} 件</span></div>
  <form class="k-bf k-bf--wide" aria-label="篩選與排序">${wide.brand}${wide.cat}${wide.cond}${wide.max}${keywordHTML(st, "型號、編號，可用日文：ネヴァーフル", "k-bf__q")}${wide.sort}${switchHTML(st)}</form>
  <form class="k-bf k-bf--narrow" aria-label="篩選與排序">${keywordHTML(st, "型號、編號，可用日文", "")}
    <div class="k-bf__bar"><button type="button" class="k-bf__toggle" data-browse-toggle aria-expanded="false" aria-controls="bf-panel">${FILTER_ICON}篩選與排序</button>${switchHTML(st)}</div>
    <div class="k-bf__panel" id="bf-panel" hidden>${narrow.brand}${narrow.cat}${narrow.cond}${narrow.max}${narrow.sort}</div>
  </form>
  <div data-browse-results>${resultsHTML(data)}</div>
</div>`;
  return `${h} · Kura 蔵`;
}

// Writes the state into both forms: mirrors a change into the other width's
// form and normalizes what was typed (trimmed keyword, digits-only budget).
function syncControls(st) {
  for (const el of document.querySelectorAll(".k-browse .k-bf [name]")) {
    if (el.type === "checkbox") el.checked = st.deals;
    else el.value = st[el.name] ?? "";
  }
}

async function apply(form) {
  if (!view) return;
  const fd = new FormData(form);
  const v = (k) => String(fd.get(k) ?? "");
  const next = readState(view.ov, "?" + new URLSearchParams({
    deals: form.elements.deals.checked ? "1" : "",
    brand: v("brand"), category: v("category"), min_rank: v("min_rank"),
    max_twd: v("max_twd").replace(/\D/g, "").replace(/^0+/, "").slice(0, 9), q: v("q"), sort: v("sort"),
  }));
  const qs = stateQuery(next).toString();
  const url = "/browse" + (qs ? "?" + qs : "");
  // Enter and the blur after it both commit a text field; the second finds
  // the URL already current and only normalizes the controls.
  if (url === location.pathname + location.search) {
    syncControls(next);
    return;
  }
  // A new filter starts from the first page; no shown count = PAGE cards.
  replaceURL(url);
  view.st = next;
  const h = heading(next);
  for (const el of document.querySelectorAll(".k-browse [data-browse-h]")) el.textContent = h;
  document.title = `${h} · Kura 蔵`;
  syncControls(next);

  const mine = ++view.seq;
  const results = document.querySelector(".k-browse [data-browse-results]");
  results.querySelector(".k-bgrid")?.setAttribute("aria-busy", "true");
  // A 再看 click during this fetch would fetch the new filter at the old grid's
  // length and append it to the old grid. The seq check cannot catch it:
  // loadMore reads view.seq after this change bumped it, so its fetch looks
  // current.
  results.querySelector("[data-browse-more]")?.setAttribute("disabled", "");
  let data;
  try {
    data = await api(apiURL(next, 0));
  } catch (err) {
    if (mine === view.seq && results.isConnected) results.innerHTML = `<p class="k-error">暫時無法載入資料，請稍後再試。</p>`;
    console.warn(err);
    return;
  }
  // A newer change, or a navigation away, owns the page now.
  if (mine !== view.seq || !results.isConnected) return;
  results.innerHTML = resultsHTML(data);
  document.querySelector(".k-browse [data-browse-count]").textContent = `${num(data.total)} 件`;
}

async function loadMore(btn) {
  btn.disabled = true;
  const mine = view.seq;
  const at = location.pathname + location.search;
  const grid = document.querySelector(".k-browse .k-bgrid");
  const shown = grid.children.length;
  let data;
  try {
    data = await api(apiURL(view.st, shown));
  } catch (err) {
    btn.disabled = false;
    console.warn(err);
    return;
  }
  // The grid stays connected while a navigation started meanwhile renders, so
  // the URL is checked too: recording the count then would put it on the next
  // page's history entry.
  if (mine !== view.seq || !grid.isConnected || location.pathname + location.search !== at) return;
  grid.insertAdjacentHTML("beforeend", cellsHTML(data.items));
  replaceURL(location.href, { shown: grid.children.length });
  const rest = data.total - grid.children.length;
  if (rest > 0) {
    btn.textContent = `再看 ${num(rest)} 件`;
    btn.disabled = false;
  } else {
    btn.closest(".k-bmore").remove();
  }
  // Focus moves to the first new card so keyboard and screen-reader users
  // continue where the new cards start, not on a removed or relabelled button.
  grid.children[shown]?.querySelector(".k-card__model")?.focus();
}

// Document-level listeners, installed at module load as in card.js and
// model.js: a listener per render would stack on re-render.
document.addEventListener("change", (e) => {
  const form = e.target.closest(".k-browse .k-bf");
  if (form) apply(form);
});

// Text fields commit on Enter as well as on blur: neither form has a submit
// button, so Enter would otherwise do nothing. An Enter that ends an IME
// composition (typing ネヴァーフル) only confirms the text.
document.addEventListener("keydown", (e) => {
  if (e.key !== "Enter" || e.isComposing) return;
  const input = e.target.closest(".k-browse .k-bf input");
  if (!input || input.type === "checkbox") return;
  e.preventDefault();
  apply(input.form);
});

document.addEventListener("submit", (e) => {
  if (e.target.closest(".k-browse .k-bf")) e.preventDefault();
});

document.addEventListener("click", (e) => {
  const toggle = e.target.closest("[data-browse-toggle]");
  if (toggle) {
    const open = toggle.getAttribute("aria-expanded") !== "true";
    toggle.setAttribute("aria-expanded", String(open));
    document.getElementById(toggle.getAttribute("aria-controls")).hidden = !open;
    return;
  }
  const more = e.target.closest("[data-browse-more]");
  if (more && !more.disabled) loadMore(more);
});
