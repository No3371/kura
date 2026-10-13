import { esc, twd, pct } from "./util.js";
import { isTracked, setTracked, onChange } from "./member.js";

// detail.js, model.js and advisor.js import it from here, beside the card markup.
export { isTracked };

// Card valuation state. Within NEAR_BAND of fair value a price reads as
// 接近行情, the same band as the legacy UI's badge (legacy/app.js valBadge).
// The band lives here rather than on the server because only the badge reads
// it and valuation() runs per card without the overview at hand; a server
// constant would need the overview threaded into every card render. 行情怎麼算
// (pages.js) states the band from this export.
//
// This band is not the radar's deal predicate. The 價格最美 shelf, its count
// and the advisor's only_deals filter use radar.IsDeal: deal_pct at least
// radar.DealMinPct, excluding low-confidence and anomaly prices.
// /api/overview counts deals at deal_pct >= 0.10. A card can therefore read
// 低於行情 with a pct under DealMinPct and still be absent from 價格最美. The
// thresholds are separate constants: changing one here does not move the others.
export const NEAR_BAND = 0.03;

export function valuation(it) {
  if (it.status !== "active") return { state: "sold" };
  const f = it.fair;
  if (!f) return { state: "none" };
  if (f.anomaly) return { state: "anomaly" };
  if (f.deal_pct >= NEAR_BAND) return { state: f.confidence === "low" ? "low" : "below", pct: pct(f.deal_pct) };
  if (f.deal_pct <= -NEAR_BAND) return { state: "above", pct: pct(-f.deal_pct) };
  return { state: "near" };
}

const BADGE = {
  below: (v) => `<span class="k-badge k-badge--below"><svg width="8" height="8" viewBox="0 0 10 10" aria-hidden="true"><path d="M1 2.5h8L5 8.5z" fill="currentColor"/></svg>低於行情 ${v.pct}%</span>`,
  low: () => `<span class="k-badge k-badge--low">可能低於行情 · 低信心</span>`,
  anomaly: () => `<span class="k-badge k-badge--anomaly"><svg width="11" height="11" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.6" aria-hidden="true"><path d="M8 2 1.5 13.5h13z"/><path d="M8 6.5v3.3M8 11.6v.4"/></svg>異常低價 · 先確認狀況</span>`,
  near: () => `<span class="k-badge k-badge--near">接近行情</span>`,
  above: (v) => `<span class="k-badge k-badge--above">高於行情 ${v.pct}%</span>`,
  none: () => `<span class="k-badge k-badge--none">尚無行情估價</span>`,
  sold: () => "",
};

export const badgeHTML = (v) => BADGE[v.state](v);

const TRACK_ON = "追蹤中，點一下取消";
const TRACK_OFF = "追蹤這件";

export const PENDING_ICON = `<svg width="44" height="44" viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="1.4" stroke-dasharray="3 3" aria-hidden="true"><path d="M12 24h40l-3 30H15z"/><path d="M23 24v-5a9 9 0 0 1 18 0v5"/></svg>`;
export const HEART = `<svg width="17" height="17" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.6" stroke-linejoin="round" aria-hidden="true"><path d="M12 20s-7-4.4-7-10a4 4 0 0 1 7-2.6A4 4 0 0 1 19 10c0 5.6-7 10-7 10z"/></svg>`;
const DROP_ICON = `<svg width="13" height="13" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><path d="M3 4l5 5 2-2 3 3"/><path d="M13 6.5V10H9.5"/></svg>`;
export const OUT_ICON = `<svg width="10" height="10" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.5" aria-hidden="true"><path d="M6 3h7v7M13 3 4 12"/></svg>`;

export function cardHTML(it, { layout = "stack" } = {}) {
  const v = valuation(it);
  const sold = v.state === "sold";
  const model = it.model_name || it.title;
  const on = isTracked(it.id);
  const label = on ? TRACK_ON : TRACK_OFF;

  const mat = it.photo_pending
    ? `${PENDING_ICON}<span class="k-card__pending">店家照片準備中</span>`
    : `<img src="${esc(it.image_url)}" alt="${esc(model)}" loading="lazy">`;
  const mark = sold
    ? `<div class="k-card__sold">已找到新主人</div>`
    : `<button type="button" class="k-card__track" data-track="${esc(it.id)}" aria-pressed="${on}" aria-label="${label}" title="${label}">${HEART}</button>`;
  const grade = it.condition?.label
    ? `<div class="k-card__grade"><span class="k-grade">${esc(it.condition.label)}</span><span class="k-card__gradezh">${esc(it.condition.desc_zh)}</span></div>`
    : "";
  const drop = it.drop && !sold
    ? `<div class="k-card__drop">${DROP_ICON}<span>已降價，原價</span><s>${twd(it.drop.from_twd)}</s></div>`
    : "";

  const href = `/listing/${encodeURIComponent(it.id)}`;
  // The photo repeats the model link for pointer users only, as on Home's
  // hero: one tab stop and one announced link per card.
  return `<article class="k-card${layout === "row" ? " k-card--row" : ""}${sold ? " is-sold" : ""}" data-id="${esc(it.id)}">
  <div class="k-card__media"><a class="k-card__mat" href="${href}" tabindex="-1" aria-hidden="true">${mat}</a>${mark}</div>
  <div class="k-card__body">
    <span class="k-card__brand">${esc(it.brand)}</span>
    <a class="k-card__model k-link" href="${href}">${esc(model)}</a>
    ${grade}
    <div class="k-card__price">${sold ? `<span class="k-card__lastsold">最後售價</span>` : ""}<span class="k-price">${twd(it.price_twd)}</span>${badgeHTML(v)}</div>
    ${drop}
    <a class="k-card__shop k-link" href="${esc(it.url)}" target="_blank" rel="noopener noreferrer">${esc(it.source_name || it.source)}<span class="k-sr">（另開視窗）</span>${OUT_ICON}</a>
  </div>
</article>`;
}

// Every copy of a key shows the store's state. The same listing can sit in two
// shelves, or on a card and the detail page, at once, and sign-in adds a whole
// watchlist; repainting every [data-track] on any change keeps them in step
// without re-rendering the view. Text-labelled buttons keep their visible text
// as the accessible name; icon-only ones carry it in aria-label. A text
// button's off label is its data-track-text value (empty → 追蹤這件), so the
// model page's 追蹤這個型號 and the advisor's 追蹤 share this path instead of a
// second toggle.
function paintTrack(b) {
  const on = isTracked(b.dataset.track);
  b.setAttribute("aria-pressed", String(on));
  const text = b.querySelector("[data-track-text]");
  if (text) {
    text.textContent = on ? "追蹤中" : text.dataset.trackText || "追蹤這件";
  } else {
    const label = on ? TRACK_ON : TRACK_OFF;
    b.setAttribute("aria-label", label);
    b.title = label;
  }
}

export function paintTracks() {
  for (const b of document.querySelectorAll("[data-track]")) paintTrack(b);
}

onChange(paintTracks);

// One document-level listener, installed at module load: re-rendered views
// would stack duplicate listeners (double toggle) if each render bound its own.
document.addEventListener("click", (e) => {
  const btn = e.target.closest("[data-track]");
  if (!btn) return;
  const id = btn.dataset.track;
  setTracked(id, !isTracked(id));
});
