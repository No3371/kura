import { api, num } from "./util.js";
import { cardHTML } from "./card.js";

const row = (items) =>
  `<div role="list" class="k-row">${items.map((it) => `<div role="listitem">${cardHTML(it)}</div>`).join("")}</div>`;

// The canvas's site-empty state; it stands in for every page's body while no
// listing is active.
export const EMPTY_HTML = `<section class="k-empty">
  <h1>還在準備中。</h1>
  <p>店裡正在整理，新的包很快就會上架。</p>
</section>`;

// Home (行情雷達). 大家都在找 is not rendered: its demand figures have no real
// source yet, and showing sample values on the product page would pass them
// off as data. 最新上架 therefore opens the page.
export async function renderHome(main, ov) {
  if (ov.overview.active === 0) {
    main.innerHTML = EMPTY_HTML;
    return;
  }
  const [fresh, best] = await Promise.all([
    api("/api/listings?sort=newest&limit=5"),
    api("/api/listings?deals=1&sort=deal&limit=3"),
  ]);
  let html = `<h1 class="k-sr">Kura 行情雷達</h1>
  <section aria-labelledby="new-h" class="k-sec k-sec--first">
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
