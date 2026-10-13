import { api, esc, num, twd } from "./util.js";
import { cardHTML } from "./card.js";
import { tagHTML } from "./detail.js";

const row = (items) =>
  `<div role="list" class="k-row">${items.map((it) => `<div role="listitem">${cardHTML(it)}</div>`).join("")}</div>`;

// The canvas's site-empty state; it stands in for every page's body while no
// listing is active.
export const EMPTY_HTML = `<section class="k-empty">
  <h1>還在準備中。</h1>
  <p>店裡正在整理，新的包很快就會上架。</p>
</section>`;

// Home: the 每日一品 hero, then 大家都在找, 最新上架, 價格最美. 大家都在找's 「這週 N 人在找」
// is seeded demand (server wanted.go; disclosed on 行情怎麼算 哪些是示範);
// its 行情, 在售 and days to a new owner are real. The canvas's 30-day trend
// is not drawn: it needs crawl history Kura has not built yet. The canvas's
// 看更多型號 link is left out — there is no model list for it to open.
export async function renderHome(main, ov) {
  if (ov.overview.active === 0) {
    main.innerHTML = EMPTY_HTML;
    return;
  }
  const [daily, wanted, fresh, best] = await Promise.all([
    api("/api/daily"),
    api("/api/wanted"),
    api("/api/listings?sort=newest&limit=5"),
    api("/api/listings?deals=1&sort=deal&limit=3"),
  ]);
  const shelf = wanted.length > 0;
  let html = heroHTML(ov.overview.active, daily.item);
  if (shelf) {
    html += `<section aria-labelledby="want-h" class="k-sec k-sec--ruled">
    <div class="k-sec__head"><h2 id="want-h">大家都在找</h2></div>
    <ol class="k-want">${wanted.map(rowHTML).join("")}</ol>
  </section>`;
  }
  html += `<section aria-labelledby="new-h" class="k-sec k-sec--ruled">
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
  main.querySelector("[data-hero-adv]").addEventListener("click", () => document.querySelector("[data-adv-open]").click());
}

// The hero (canvas concept B): the headline, today's real active count, and
// the day's 每日一品 (server daily.go: best condition, then lowest price,
// picked once per Taipei day) on its mat with the shop's 鑑定卡 hanging off
// it. A pick that sold during its day stays, marked as the cards mark a sale.
// With no pick, the headline stands alone.
function heroHTML(active, it) {
  const sold = it && it.status !== "active";
  const href = it && `/listing/${encodeURIComponent(it.id)}`;
  const fig = it
    ? `<figure class="k-hero__fig${sold ? " is-sold" : ""}">
      <div class="k-hero__stage">
        <a class="k-hero__mat" href="${href}" tabindex="-1" aria-hidden="true"><img src="${esc(it.image_url)}" alt=""></a>
        <span class="k-hero__daily">每日一品</span>
        ${sold ? `<div class="k-card__sold">已找到新主人</div>` : ""}
        <div class="k-hero__tag">${tagHTML(it)}</div>
      </div>
      <figcaption class="k-hero__cap">
        <span class="k-hero__brand">${esc(it.brand)}${it.model_code ? ` · ${esc(it.model_code)}` : ""}</span>
        <a class="k-hero__model k-link" href="${href}">${esc(it.model_name || it.title)}</a>
      </figcaption>
    </figure>`
    : "";
  return `<section class="k-hero${it ? "" : " k-hero--solo"}" aria-labelledby="hero-h">
    <div class="k-hero__text">
      <h1 id="hero-h">它與前任錯過，<br>只為與你相逢。</h1>
      <p class="k-hero__lede">Kura 陪你一起找到想要的日本二手精品包。<br>今天架上有 <span class="k-hero__count">${num(active)}</span>&nbsp;只。</p>
      <div class="k-hero__cta">
        <a class="k-hero__go" href="/browse">逛逛架上精品</a>
        <button type="button" class="k-hero__ask" data-hero-adv><span class="k-adv__seal" aria-hidden="true">蔵</span>問找包顧問</button>
      </div>
    </div>
    ${fig}
  </section>`;
}

const BAG = `<svg aria-hidden="true" viewBox="0 0 64 64" fill="none" stroke="currentColor" stroke-width="1.1"><path d="M12 24h40l-3 30H15z"/><path d="M23 24v-5a9 9 0 0 1 18 0v5"/><path d="M28 31h8v5h-8z"/></svg>`;

// One 大家都在找 row (canvas concept A's index): a large rank, a small mat,
// brand and model, the seeded demand line, then real figures. The rank is the
// <ol>'s own order, so the numeral is aria-hidden rather than read twice.
function rowHTML(m) {
  const name = m.model_name || m.model_code || m.group_key;
  const href = `/model/${encodeURIComponent(m.brand)}/${encodeURIComponent(m.group_key)}`;
  const d = m.median_days_to_sell;
  const speed = d != null ? `${d < 1 ? "不到 1 天" : `平均 ${Math.round(d)} 天`}找到新主人 · ` : "";
  return `<li><a class="k-want__row k-link" href="${href}">
    <span class="k-want__rank" aria-hidden="true">${m.rank}</span>
    <span class="k-want__mat">${m.image_url ? `<img src="${esc(m.image_url)}" alt="">` : BAG}</span>
    <span class="k-want__id"><span class="k-want__brand">${esc(m.brand)}</span><span class="k-want__model">${esc(name)}</span></span>
    <span class="k-want__heat">這週 ${num(m.seekers)} 人在找</span>
    <span class="k-want__facts"><span>行情 <span class="k-want__price">${twd(m.median_twd)}</span></span><span>${speed}在售 ${num(m.active)} 件</span></span>
  </a></li>`;
}
