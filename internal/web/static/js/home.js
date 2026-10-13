import { api, esc, num, twd } from "./util.js";
import { cardHTML } from "./card.js";

const row = (items) =>
  `<div role="list" class="k-row">${items.map((it) => `<div role="listitem">${cardHTML(it)}</div>`).join("")}</div>`;

// The canvas's site-empty state; it stands in for every page's body while no
// listing is active.
export const EMPTY_HTML = `<section class="k-empty">
  <h1>還在準備中。</h1>
  <p>店裡正在整理，新的包很快就會上架。</p>
</section>`;

// Home (行情雷達): 大家都在找, 最新上架, 價格最美. 大家都在找's 「這週 N 人在找」
// is seeded demand (server wanted.go; disclosed on 行情怎麼算 哪些是示範);
// its 行情, 在售 and days to a new owner are real. The canvas's 30-day trend
// is not drawn: it needs crawl history Kura has not built yet. The canvas's
// 看更多型號 link is left out — there is no model list for it to open.
export async function renderHome(main, ov) {
  if (ov.overview.active === 0) {
    main.innerHTML = EMPTY_HTML;
    return;
  }
  const [wanted, fresh, best] = await Promise.all([
    api("/api/wanted"),
    api("/api/listings?sort=newest&limit=5"),
    api("/api/listings?deals=1&sort=deal&limit=3"),
  ]);
  const shelf = wanted.length > 0;
  let html = `<h1 class="k-sr">Kura 行情雷達</h1>`;
  if (shelf) {
    html += `<section aria-labelledby="want-h" class="k-sec k-sec--first">
    <div class="k-sec__head"><h2 id="want-h">大家都在找</h2></div>
    <ol class="k-want">${wanted.map(tileHTML).join("")}</ol>
  </section>`;
  }
  html += `<section aria-labelledby="new-h" class="k-sec ${shelf ? "k-sec--ruled" : "k-sec--first"}">
    <div class="k-sec__head">
      <h2 id="new-h">最新上架</h2>
      <a class="k-sec__more k-link" href="/browse?sort=newest">看全部最新上架 →</a>
    </div>
    ${row(fresh.items)}
  </section>`;
  if (best.total > 0) {
    html += `<section aria-labelledby="deal-h" class="k-sec k-sec--ruled">
    <div class="k-sec__head">
      <h2 id="deal-h">價格最美</h2>
      <a class="k-sec__more k-link" href="/browse?deals=1&sort=deal">看全部價格最美 ${num(best.total)} 件 →</a>
    </div>
    ${row(best.items)}
  </section>`;
  }
  main.innerHTML = html;
}

const BAG = `<svg aria-hidden="true" viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="1.1"><path d="M12 24h40l-3 30H15z"/><path d="M23 24v-5a9 9 0 0 1 18 0v5"/><path d="M28 31h8v5h-8z"/></svg>`;

// One 大家都在找 tile, as the canvas draws it: rank on the mat, brand, model,
// the seeded demand line, then real figures. The rank is the <ol>'s own
// order, so the mat numeral is aria-hidden rather than read twice.
function tileHTML(m) {
  const name = m.model_name || m.model_code || m.group_key;
  const href = `/model/${encodeURIComponent(m.brand)}/${encodeURIComponent(m.group_key)}`;
  const d = m.median_days_to_sell;
  const speed = d != null ? `${d < 1 ? "不到 1 天" : `平均 ${Math.round(d)} 天`}找到新主人 · ` : "";
  return `<li><a class="k-want__tile k-link" href="${href}">
    <span class="k-want__mat">${m.image_url ? `<img src="${esc(m.image_url)}" alt="">` : BAG}<span class="k-want__rank" aria-hidden="true">${m.rank}</span></span>
    <span class="k-want__id"><span class="k-want__brand">${esc(m.brand)}</span><span class="k-want__model">${esc(name)}</span></span>
    <span class="k-want__heat">這週 ${num(m.seekers)} 人在找</span>
    <span class="k-want__facts"><span>行情 <span class="k-want__price">${twd(m.median_twd)}</span></span><span>${speed}在售 ${num(m.active)} 件</span></span>
  </a></li>`;
}
