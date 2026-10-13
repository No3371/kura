import { api, esc, num, twd, md } from "./util.js";
import { cardHTML, isTracked } from "./card.js";
import { distHTML } from "./dist.js";

// Model page (型號頁) at /model/{brand}/{group_key}. Every figure comes from
// the server (model stats, overview) already in TWD; the page converts and
// recomputes nothing. 「這週 N 人在找」 is seeded demand from the same server
// function as Home's 大家都在找 (wanted.go), so a model reads one number on
// both. The crumb names this model rather than the 大家都在找 shelf: most
// visits arrive from a listing page, and the shelf would state a path nobody
// took.

const CAP = 5; // one desktop row of cards, as the canvas draws 在售的 {model}

// aria-hidden, not the canvas's role="img" aria-label="型號照片": the icon is a
// generic bag drawn when no listing has a photo, so announcing it as the
// model's photo would claim an image that does not exist; the h1 names the
// model.
const MODEL_ICON = `<svg aria-hidden="true" viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="0.9"><path d="M12 24h40l-3 30H15z"/><path d="M23 24v-5a9 9 0 0 1 18 0v5"/><path d="M28 31h8v5h-8z"/></svg>`;
const BELL = `<svg viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="1.5" stroke-linejoin="round" aria-hidden="true"><path d="M6 16.5V11a6 6 0 0 1 12 0v5.5l1.5 2H4.5z"/><path d="M10 20.5a2 2 0 0 0 4 0"/></svg>`;

export async function renderModel(main, ov, rawBrand, rawKey) {
  let brand, key;
  try {
    brand = decodeURIComponent(rawBrand);
    key = decodeURIComponent(rawKey);
  } catch {
    throw Object.assign(new Error("bad model path"), { status: 404 });
  }
  const { model: s, items, schemes, seekers } = await api(`/api/models/${encodeURIComponent(brand)}/${encodeURIComponent(key)}`);

  const name = s.model_name || s.model_code || key;
  const brandZH = ov.brands?.find((b) => b.name === s.brand)?.zh;
  const catZH = ov.categories?.find((c) => c.key === s.category)?.zh;
  const sub = [catZH, s.model_code].filter(Boolean).join(" · ");
  const active = items.filter((it) => it.status === "active");
  // The stats' image_url is the first-inserted listing's and can be the shop's
  // "now printing" placeholder; any listing with a real photo shows the model.
  const hasPhoto = (it) => !it.photo_pending && it.image_url;
  const photo = active.find(hasPhoto) ?? items.find(hasPhoto);
  // Shares the tracked set with listings, whose keys are "{source}:{id}". The
  // "model:" prefix cannot collide with them while no source is named
  // "model"; a separate store per kind would split one 追蹤 list in two.
  const trackKey = `model:${s.brand}/${s.group_key}`;
  const on = isTracked(trackKey);
  const dist = distHTML({ stats: s, items, schemes });

  main.innerHTML = `<div class="k-model">
  <nav aria-label="位置" class="k-crumb"><a href="/" class="k-link">首頁</a><span aria-hidden="true">/</span><span class="k-crumb__model" aria-current="page">${esc(s.brand)} ${esc(name)}</span></nav>
  <div class="k-model__top">
    <div class="k-model__mat">${photo ? `<img src="${esc(photo.image_url)}" alt="${esc(name)} 型號照片">` : MODEL_ICON}</div>
    <div class="k-model__info">
      <div class="k-detail__id">
        <span class="k-detail__brand">${esc(s.brand)}${brandZH ? ` <span>${esc(brandZH)}</span>` : ""}</span>
        <h1>${esc(name)}</h1>
        ${sub ? `<span class="k-detail__sub">${esc(sub)}</span>` : ""}
      </div>
      <p class="k-model__heat">這週 ${num(seekers)} 人在找</p>
      <dl class="k-mstats">${tilesHTML(s, ov)}</dl>
      <div class="k-mtrack">
        <button type="button" class="k-dtrack k-dtrack--model" data-track="${esc(trackKey)}" aria-pressed="${on}">${BELL}<span data-track-text="追蹤這個型號">${on ? "追蹤中" : "追蹤這個型號"}</span></button>
        <span>有新上架或低於行情時通知你</span>
      </div>
    </div>
  </div>
  ${dist ? `<div class="k-model__dist">${dist}</div>` : ""}
  ${active.length ? saleHTML(s, name, active) : ""}
</div>`;
  // 型號 keeps this title apart from the listing page's, which otherwise has
  // the same {name} · {brand} shape in tabs and history.
  return `${name} 型號 · ${s.brand} · Kura 蔵`;
}

// Whether sold-derived figures (已售出, 平均找到新主人) may be stated is the
// server's call (overview sold_stats_ready: enough crawl history), so every
// surface showing them shares one gate. Before it opens they read 資料不足:
// a model's 0 sold is unknown, not zero, and would read as "never sells".
const soldKnown = (ov) => ov.overview.sold_stats_ready === true;

function tilesHTML(s, ov) {
  const tile = (k, v, note, na = false) =>
    `<div><dt>${esc(k)}</dt><dd class="k-mstat__v${na ? " is-na" : ""}">${esc(v)}</dd><dd class="k-mstat__note">${esc(note)}</dd></div>`;
  const NA = "資料不足";
  const rec = `自 ${md(ov.overview.since)} 起記錄`;
  // The labelled grade with the most listings in sale; strict > keeps the
  // earlier (better) grade on a tie, since by_grade is ordered best first.
  const g = (s.by_grade ?? [])
    .filter((x) => x.grade?.label)
    .reduce((best, x) => (!best || x.count > best.count ? x : best), null);
  const d = s.median_days_to_sell;

  const known = soldKnown(ov);

  let html = tile("在售", `${num(s.active)} 件`, "各等級合計");
  html += known ? tile("已售出", `${num(s.sold)} 件`, rec) : tile("已售出", NA, rec, true);
  if (g) html += tile(`${g.grade.label} 中位數`, twd(g.median_twd), `${num(g.count)} 件比較`);
  // Label kept verbatim from the canvas although the value is the median days
  // to sell, not a mean: the median resists one listing that sat for months.
  // Relabelling is a copy decision for the canvas, not this page.
  html += known && d != null
    ? tile("平均找到新主人", d < 1 ? "不到 1 天" : `${Math.round(d)} 天`, rec)
    : tile("平均找到新主人", NA, rec, true);
  return html;
}

function saleHTML(s, name, active) {
  const more = active.length > CAP
    ? `<button type="button" class="k-sec__more k-link k-more" data-model-more>看全部 ${num(s.active)} 件 →</button>`
    : "";
  const cards = active.map((it, i) => `<div role="listitem"${i >= CAP ? " hidden" : ""}>${cardHTML(it)}</div>`).join("");
  return `<section class="k-msale" aria-labelledby="ml-h">
    <div class="k-msale__head"><h2 id="ml-h">在售的 <span>${esc(name)}</span></h2>${more}</div>
    <div role="list" class="k-row">${cards}</div>
  </section>`;
}

// 看全部 reveals the cards already rendered rather than linking to Browse:
// Browse filters by brand, type, condition, budget and keyword, not by model
// group, so no Browse URL lists exactly this group and the link's count
// would not match the page it opens. One document-level listener, installed
// at module load, as in card.js: a listener per render would stack on
// re-render.
document.addEventListener("click", (e) => {
  const btn = e.target.closest("[data-model-more]");
  if (!btn) return;
  const rest = btn.closest(".k-msale").querySelectorAll("[role=list] > [hidden]");
  for (const li of rest) li.hidden = false;
  btn.remove();
  rest[0]?.querySelector(".k-card__model")?.focus();
});
