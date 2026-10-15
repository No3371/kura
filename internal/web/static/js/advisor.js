import { api, esc, num, twd } from "./util.js";
import { valuation, badgeHTML, isTracked, HEART, OUT_ICON, PENDING_ICON } from "./card.js";

// Tool → [helper label, in-progress copy]. Progress lines come from real tool
// events only; the canvas's third helper (風格) is not drawn, because its
// stated grounds (大家都在找) are mocked demand, not data. A tool missing here
// gets no line rather than a made-up label.
const STEP = {
  search_listings: ["行情", "搜尋在售商品"],
  model_price_stats: ["行情", "查詢型號行情"],
  market_overview: ["行情", "查看市場概況"],
  get_listing: ["成色", "讀取商品詳情"],
};

// The prompt asks for 2–4 picks, but the offline mock lists up to six and a
// model may cite more; a figure for every reference would bury the answer
// under cards, so later references stay chips.
const MAX_FIGURES = 4;
const REF = /\[\[([a-z]+:[\w-]+)\]\]/g;
const BULLET = /^\s*(?:[-*•]|\d+\.)\s+(.*)$/;
const NET_ERR = "暫時無法載入資料，請稍後再試。";
const CURSOR = `<span class="k-adv__cursor" aria-hidden="true"></span>`;
// Only follow the stream when the reader is already near the end; someone
// scrolled up to reread an earlier answer is not yanked back down.
const PIN_SLACK = 120;

const dock = document.querySelector("[data-adv-open]");
const dlg = document.querySelector("dialog.k-adv");
const scroller = dlg.querySelector(".k-adv__body");
const log = dlg.querySelector(".k-adv__log");
const form = dlg.querySelector(".k-adv__form");
const input = form.elements.q;
const sendBtn = form.querySelector(".k-adv__send");
const maxBtn = dlg.querySelector("[data-adv-max]");

// One session per page load, not per browser tab (sessionStorage): after a
// reload the panel starts empty, and a reused id would let the advisor answer
// from earlier turns the user can no longer see.
const session = crypto.randomUUID();

// Listings the tools returned, by id, for the page lifetime like the session:
// a later answer may cite a listing an earlier turn found without the tools
// returning it again, and its chip and figure still need the data.
const known = new Map();

// A narrow or touch screen opens the panel without the composer focused:
// focusing it there raises the virtual keyboard over the intro and prompts,
// which the canvas's phone panel shows keyboard-free. The heading takes focus
// instead, so keyboard and screen-reader users still start inside the panel.
// Rejected: the textarea's autofocus on every screen, the desktop behaviour.
const noComposerFocus = matchMedia("(max-width: 767px), (pointer: coarse)");
const heading = dlg.querySelector(".k-adv__title h2");

let busy = false;
// Set while the panel closes for an in-app link: the router then focuses the
// new page's heading, and the close handler must not pull focus back to the
// dock — the close event is a queued task that can run after a quick render.
let leaving = false;

export function initAdvisor(overviewP) {
  overviewP.then((ov) => {
    dlg.querySelector(".k-adv__demo").hidden = ov.advisor !== "mock";
  }).catch(() => {});
}

function setFull(full) {
  dlg.classList.toggle("is-full", full);
  const label = full ? "縮小成視窗" : "放大成整頁";
  maxBtn.setAttribute("aria-label", label);
  maxBtn.title = label;
}

dock.addEventListener("click", () => {
  setFull(false);
  dlg.showModal();
  (noComposerFocus.matches ? heading : input).focus();
});
maxBtn.addEventListener("click", () => setFull(!dlg.classList.contains("is-full")));
dlg.querySelector("[data-adv-close]").addEventListener("click", () => dlg.close());

// Esc fires cancel → close, so every way out lands here. Reopening shows the
// floating panel, as on the canvas.
dlg.addEventListener("close", () => {
  setFull(false);
  if (!leaving) dock.focus();
  leaving = false;
});

// An in-app link inside the panel (chip, name, 看詳細) closes it so the page
// it opens is visible. No preventDefault: the router's document listener
// still handles the navigation. Modified clicks and new-tab links (the shop
// page) leave the panel open, as the router ignores them too. Focus then
// belongs to the new page's heading, not the dock (see `leaving`).
dlg.addEventListener("click", (e) => {
  if (e.button !== 0 || e.metaKey || e.ctrlKey || e.shiftKey || e.altKey) return;
  const a = e.target.closest("a[href^='/']");
  if (!a || a.target) return;
  leaving = true;
  dlg.close();
});

log.addEventListener("click", (e) => {
  const b = e.target.closest("[data-adv-prompt]");
  if (b) ask(b.textContent);
});
form.addEventListener("submit", (e) => {
  e.preventDefault();
  ask(input.value);
});
input.addEventListener("keydown", (e) => {
  // isComposing: Enter that confirms an IME candidate (注音, 倉頡) must not send.
  if (e.key === "Enter" && !e.shiftKey && !e.isComposing) {
    e.preventDefault();
    form.requestSubmit();
  }
});

const nearBottom = () => scroller.scrollHeight - scroller.scrollTop - scroller.clientHeight < PIN_SLACK;
const toEnd = () => { scroller.scrollTop = scroller.scrollHeight; };

async function ask(raw) {
  const text = raw.trim();
  if (!text || busy) return;
  busy = true;
  sendBtn.disabled = true;
  input.value = "";
  log.querySelector(".k-adv__prompts")?.remove();

  log.insertAdjacentHTML("beforeend", `<div class="k-adv__user">${esc(text)}</div>`);
  const turn = document.createElement("div");
  turn.className = "k-adv__turn";
  turn.innerHTML = `<ol class="k-adv__steps" aria-label="助手的進度" hidden></ol><div class="k-adv__answer" aria-live="polite" aria-busy="true">${CURSOR}</div>`;
  log.append(turn);
  const steps = turn.querySelector(".k-adv__steps");
  const answer = turn.querySelector(".k-adv__answer");
  toEnd();

  let md = "";
  let errText = "";
  try {
    for await (const ev of events(text)) {
      const pinned = nearBottom();
      switch (ev.type) {
        case "text":
          md += ev.text;
          answer.innerHTML = streamingHTML(blocks(md, known));
          break;
        case "tool_call": {
          const step = STEP[ev.tool];
          if (!step) break;
          steps.hidden = false;
          steps.insertAdjacentHTML("beforeend", `<li data-tool="${esc(ev.tool)}"><span class="k-adv__who">${step[0]}</span><span>${step[1]}⋯</span></li>`);
          break;
        }
        case "tool_result": {
          const li = [...steps.querySelectorAll("li:not(.is-done)")].filter((l) => l.dataset.tool === ev.tool).pop();
          if (!li) break;
          // A failed call goes back to the model, which retries; its raw
          // error is model-facing English, so the line goes rather than shows.
          if (ev.is_error) {
            li.remove();
            steps.hidden = !steps.children.length;
            break;
          }
          li.lastElementChild.textContent = ev.summary;
          li.insertAdjacentHTML("beforeend", `<span class="k-adv__ok" aria-label="完成">✓</span>`);
          li.classList.add("is-done");
          break;
        }
        case "listings":
          for (const it of ev.listings ?? []) known.set(it.id, it);
          break;
        case "error":
          errText = ev.text;
          break;
      }
      if (pinned) toEnd();
    }
  } catch (err) {
    console.warn(err);
    errText ||= err.shopper || NET_ERR;
  } finally {
    const pinned = nearBottom();
    // Figures render once, here: building them per text delta would refetch
    // the position-bar data on every chunk.
    answer.innerHTML = finalHTML(blocks(md, known), known);
    answer.removeAttribute("aria-busy");
    if (errText) answer.insertAdjacentHTML("beforeend", `<p class="k-adv__err">${esc(errText)}</p>`);
    fillPositions(answer);
    busy = false;
    sendBtn.disabled = false;
    if (pinned) toEnd();
  }
}

// events streams POST /api/chat's server-sent events as parsed objects.
async function* events(message) {
  const r = await fetch("/api/chat", {
    method: "POST",
    headers: { "Content-Type": "application/json" },
    body: JSON.stringify({ session, message }),
  });
  if (!r.ok || !r.body) {
    // Refusals before the stream (rate limit, over-long message) carry a
    // shopper-facing line in {"error": …}.
    const err = new Error(`chat: HTTP ${r.status}`);
    if (r.status === 429 || r.status === 400) err.shopper = (await r.json().catch(() => null))?.error;
    throw err;
  }
  const reader = r.body.getReader();
  const dec = new TextDecoder();
  let buf = "";
  for (;;) {
    const { value, done } = await reader.read();
    if (done) return;
    buf += dec.decode(value, { stream: true });
    let i;
    while ((i = buf.indexOf("\n\n")) >= 0) {
      const chunk = buf.slice(0, i);
      buf = buf.slice(i + 2);
      if (chunk.startsWith("data: ")) yield JSON.parse(chunk.slice(6));
    }
  }
}

// blocks parses the answer's small markdown subset (paragraphs, bullet lists,
// **bold**, [[listing-id]] references) into blocks. Text is escaped before
// any markup is added. Each block keeps the ids it references, in order, so
// a figure can follow the block that first names its listing.
function blocks(src, known) {
  const out = [];
  let list = null;
  for (const line of esc(src).split("\n")) {
    const m = line.match(BULLET);
    if (m) {
      if (!list) out.push((list = { items: [], ids: [] }));
      list.items.push(inline(m[1], known, list.ids));
      continue;
    }
    list = null;
    if (line.trim()) {
      const b = { ids: [] };
      b.text = inline(line.replace(/^#+\s*/, ""), known, b.ids);
      out.push(b);
    }
  }
  return out;
}

function inline(s, known, ids) {
  return s
    .replace(/\*\*(.+?)\*\*/g, "<strong>$1</strong>")
    .replace(REF, (_, id) => {
      ids.push(id);
      return chip(id, known.get(id));
    });
}

function chip(id, it) {
  let text = "查看商品";
  if (it) text = (it.model_name || it.title) + (it.model_code ? ` · ${it.model_code}` : "");
  return `<a class="k-adv__chip" href="/listing/${encodeURIComponent(id)}">${esc(text)}</a>`;
}

// blockHTML renders one block; tail goes inside its last line (the cursor).
function blockHTML(b, tail = "") {
  if (!b.items) return `<p>${b.text}${tail}</p>`;
  const last = b.items.length - 1;
  return `<ul>${b.items.map((li, i) => `<li>${li}${i === last ? tail : ""}</li>`).join("")}</ul>`;
}

function streamingHTML(bs) {
  if (!bs.length) return CURSOR;
  return bs.map((b, i) => blockHTML(b, i === bs.length - 1 ? CURSOR : "")).join("");
}

// finalHTML adds a figure after the block holding each listing's first
// reference, for listings the tools returned and that are still for sale.
// Sold or unknown ids stay chips: there is nothing to buy behind a figure.
function finalHTML(bs, known) {
  const shown = new Set();
  let html = "";
  for (const b of bs) {
    html += blockHTML(b);
    for (const id of b.ids) {
      const it = known.get(id);
      if (shown.has(id) || shown.size >= MAX_FIGURES || it?.status !== "active") continue;
      shown.add(id);
      html += figure(it);
    }
  }
  return html;
}

function figure(it) {
  const name = it.model_name || it.title;
  const href = `/listing/${encodeURIComponent(it.id)}`;
  const shop = it.source_name || it.source;
  const on = isTracked(it.id);
  const c = it.condition ?? {};
  const grade = c.label
    ? `<span class="k-adv__grade"><span class="k-grade">${esc(c.label)}</span>${esc(c.desc_zh)}</span>`
    : "";
  return `<figure class="k-adv__rec" aria-label="推薦：${esc(name)}${it.model_code ? " " + esc(it.model_code) : ""}">
  ${galleryHTML(it, name, href)}
  <div class="k-adv__recmeta">
    <div class="k-adv__recid"><span class="k-adv__brand">${esc(it.brand)}</span><a class="k-adv__name k-link" href="${href}">${esc(name)}</a>${grade}</div>
    <div class="k-adv__recprice"><span class="k-price">${twd(it.price_twd)}</span>${badgeHTML(valuation(it))}</div>
  </div>
  <div class="k-adv__pos" data-adv-pos hidden data-brand="${esc(it.brand)}" data-group="${esc(it.group_key)}" data-source="${esc(it.source)}" data-code="${esc(c.code)}" data-label="${esc(c.label)}" data-price="${esc(it.price_twd)}"></div>
  <div class="k-adv__acts">
    <button type="button" class="k-adv__btn" data-track="${esc(it.id)}" aria-pressed="${on}">${HEART}<span data-track-text="追蹤">${on ? "追蹤中" : "追蹤"}</span></button>
    <a class="k-adv__btn" href="${href}">看詳細</a>
    <a class="k-adv__btn k-adv__btn--shop" href="${esc(it.url)}" target="_blank" rel="noopener noreferrer">${esc(shop)} 商品頁<span class="k-sr">（另開視窗）</span>${OUT_ICON}</a>
  </div>
</figure>`;
}

// galleryHTML is a figure's photo: the stored gallery when there is one, else
// the list-page photo alone, as on the detail page. Two or more photos become
// a scroll-snapped strip — swiped on touch, paged by the arrow buttons — with
// a counter; one photo stays a plain mat. Every photo links to the listing.
// A lone photo's link is hidden from assistive tech, as the name link already
// leads there; a strip's photos stay exposed, each alt saying which of n it is,
// so the arrow buttons page through something a screen reader can read.
function galleryHTML(it, name, href) {
  const photos = it.images?.length ? it.images : it.photo_pending ? [] : [it.image_url];
  if (!photos.length) {
    return `<a class="k-adv__mat" href="${href}" tabindex="-1" aria-hidden="true">${PENDING_ICON}<span class="k-card__pending">店家照片準備中</span></a>`;
  }
  if (photos.length < 2) {
    return `<a class="k-adv__mat" href="${href}" tabindex="-1" aria-hidden="true"><img src="${esc(photos[0])}" alt="${esc(name)} 商品照片（店家白底照）" loading="lazy"></a>`;
  }
  const n = photos.length;
  const slides = photos
    .map((u, i) => `<a class="k-adv__mat" href="${href}" tabindex="-1"><img src="${esc(u)}" alt="${esc(name)} 商品照片 ${i + 1}／${n}" loading="lazy"></a>`)
    .join("");
  return `<div class="k-adv__gal" data-adv-gal>
    <div class="k-adv__slides">${slides}</div>
    <button type="button" class="k-adv__nav k-adv__nav--prev" data-adv-step="-1" aria-label="上一張照片" disabled>${CHEVRON}</button>
    <button type="button" class="k-adv__nav k-adv__nav--next" data-adv-step="1" aria-label="下一張照片">${CHEVRON}</button>
    <span class="k-adv__count" aria-hidden="true"><span data-adv-at>1</span>／${n}</span>
  </div>`;
}

const CHEVRON = `<svg viewBox="0 0 16 16" aria-hidden="true"><path d="M6 3l5 5-5 5" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linecap="round" stroke-linejoin="round"/></svg>`;
const calm = matchMedia("(prefers-reduced-motion: reduce)");

// The strip's page index, from its scroll position.
const page = (strip) => Math.round(strip.scrollLeft / strip.clientWidth);

log.addEventListener("click", (e) => {
  const b = e.target.closest("[data-adv-step]");
  if (!b) return;
  const strip = b.closest("[data-adv-gal]").querySelector(".k-adv__slides");
  // A second click during a smooth scroll steps from the page it is headed
  // to; reading the half-scrolled position would round back and lose a step.
  const from = strip.dataset.to ? Number(strip.dataset.to) : page(strip);
  const to = Math.min(Math.max(from + Number(b.dataset.advStep), 0), strip.children.length - 1);
  strip.dataset.to = to;
  strip.scrollTo({ left: to * strip.clientWidth, behavior: calm.matches ? "auto" : "smooth" });
});

// scroll does not bubble; capturing it here covers every strip in the log,
// swiped or paged, and keeps the counter and arrows in step with it.
log.addEventListener("scroll", (e) => {
  const strip = e.target;
  if (!strip.classList?.contains("k-adv__slides")) return;
  const g = strip.closest("[data-adv-gal]");
  const at = page(strip);
  const last = strip.children.length - 1;
  // Settled on a page: the paged scroll is over, or a swipe took over.
  if (Math.abs(strip.scrollLeft - at * strip.clientWidth) < 1) delete strip.dataset.to;
  g.querySelector("[data-adv-at]").textContent = at + 1;
  g.querySelector(".k-adv__nav--prev").disabled = at <= 0;
  g.querySelector(".k-adv__nav--next").disabled = at >= last;
}, true);

// Model responses per brand|group_key for the page lifetime. A failed fetch
// is dropped from the cache so a later turn can retry it.
const models = new Map();
function modelData(brand, group) {
  const key = `${brand}|${group}`;
  if (!models.has(key)) {
    models.set(key, api(`/api/models/${encodeURIComponent(brand)}/${encodeURIComponent(group)}`).catch(() => {
      models.delete(key);
      return null;
    }));
  }
  return models.get(key);
}

const clamp = (v, lo, hi) => Math.min(hi, Math.max(lo, v));

// fillPositions draws each figure's 「在同款 {grade} 的 n 件裡的位置」 bar from
// the model's by_grade stats for the listing's shop and grade — the in-sale
// figures the detail and model distributions use — so every number on it has
// that one source. A grade with fewer than two listings in sale, or a single
// price, has no range to place the listing in; its bar stays hidden.
function fillPositions(root) {
  for (const el of root.querySelectorAll("[data-adv-pos]")) {
    const { brand, group, source, code, label } = el.dataset;
    const price = Number(el.dataset.price);
    if (!group || !code) continue;
    modelData(brand, group).then((res) => {
      const s = res?.model?.by_grade?.find((x) => x.source === source && x.grade.code === code);
      if (!s || s.count < 2 || !(s.max_twd > s.min_twd)) return;
      const at = (v) => clamp(((v - s.min_twd) / (s.max_twd - s.min_twd)) * 100, 0, 100);
      // The median label is centred on its tick but kept between 25% and 75%,
      // where it cannot run into the min and max labels at either end.
      const medLeft = clamp(at(s.median_twd), 25, 75);
      const medLabel = `<span class="k-adv__medlabel" style="left:${medLeft}%">中位數 ${twd(s.median_twd)}</span>`;
      const pinned = nearBottom();
      el.setAttribute("role", "img");
      el.setAttribute("aria-label", `同款 ${label} 在售價格 ${twd(s.min_twd)} 至 ${twd(s.max_twd)}，中位數 ${twd(s.median_twd)}，這件 ${twd(price)}`);
      el.innerHTML = `<span class="k-adv__poscap">在同款 ${esc(label)} 的 ${num(s.count)} 件裡的位置</span>
  <div class="k-adv__medrow">${medLabel}</div>
  <div class="k-adv__track"><i class="k-adv__med" style="left:${at(s.median_twd)}%"></i><i class="k-adv__this" style="left:${at(price)}%"></i></div>
  <div class="k-adv__ticks"><span>${twd(s.min_twd)}</span>${medLabel}<span>${twd(s.max_twd)}</span></div>`;
      el.hidden = false;
      if (pinned) toEnd();
    });
  }
}
