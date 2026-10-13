import { esc, num, twd } from "./util.js";

// 同型號價格分布: one column per shop grade (best first), one dot per listing
// of the model group at its TWD price. Medians and the table come from the
// server's by_grade stats (in-sale listings) rather than being recomputed
// here, so the chart, the table and the advisor quote the same figures.
//
// Dots are hover-only spans, not buttons: a large group (400+ listings) would
// otherwise add hundreds of tab stops. The plot is one role="img" with a
// summary label, and 以表格檢視 gives the same numbers as a table.

// Jitter lanes within a grade column, as fractions of the outermost lane's
// offset. That offset is 26px, shrunk to keep a dot's 7px radius (10px mark +
// 2px halo) 1px inside its column: fixed px lanes overflow narrow columns
// (5 grades at 375px wide leave ~60px each) and dots of neighbouring grades
// overlap. Scaling every lane, rather than clamping each offset, keeps the
// five lanes distinct. Pure CSS, so a resize needs no re-render.
const LANES = [0, -0.5, 0.5, -1, 1];
const LANE_MAX = 26; // px

function niceTicks(lo, hi, n = 4) {
  const span = hi - lo || hi || 1;
  const step0 = span / n, mag = 10 ** Math.floor(Math.log10(step0));
  const step = [1, 2, 2.5, 5, 10].map((m) => m * mag).find((s) => s >= step0);
  const t = [];
  for (let v = Math.ceil(lo / step) * step; v <= hi + 1e-9; v += step) t.push(v);
  return t;
}

const shortTWD = (v) => (v >= 10000 ? (v / 10000).toFixed(v % 10000 ? 1 : 0) + " 萬" : num(v));

const gradeName = (g) => g.label || g.code;

// distHTML renders the section, or "" when no listing in the group has a grade.
// thisId marks one listing (detail page); null renders the plain model chart.
//
// One ladder (plot + axis) per shop in schemes: a grade code means something
// only within its shop, so each shop's dots, medians and table rows sit under
// its own ladder, keyed {source, code}. The viewed listing's shop comes first.
// Rejected: every shop's grades as columns of one plot (two ladders at 375px
// leave ~20px per column, under the dot lanes' minimum above) and plotting
// only the viewed shop (hides the other shops' prices on the model page). The
// price scale is shared so the ladders stay comparable by eye.
export function distHTML({ stats, items, schemes = [], thisId = null }) {
  const thisSource = items.find((it) => it.id === thisId)?.source;
  const ordered = thisSource
    ? [...schemes.filter((s) => s.source === thisSource), ...schemes.filter((s) => s.source !== thisSource)]
    : schemes;
  const ladders = ordered.map((scheme) => {
    const cols = (scheme.grades ?? []).map((grade) => ({ grade }));
    const own = items.filter((it) => it.source === scheme.source);
    for (const it of own) {
      const c = it.condition;
      // Grades outside the published ladder are appended after it.
      if (c?.code && !cols.some((x) => x.grade.code === c.code)) cols.push({ grade: c });
    }
    for (const col of cols) {
      col.stat = stats.by_grade?.find((s) => s.source === scheme.source && s.grade.code === col.grade.code);
      col.pts = own.filter((it) => it.condition?.code === col.grade.code).sort((a, b) => a.price_twd - b.price_twd);
    }
    return { scheme, cols };
  }).filter((l) => l.cols.some((c) => c.pts.length || c.stat));
  if (!ladders.length) return "";
  const vals = ladders.flatMap((l) => l.cols.flatMap((c) => [...c.pts.map((p) => p.price_twd), ...(c.stat ? [c.stat.median_twd] : [])]));

  let lo = Math.min(...vals), hi = Math.max(...vals);
  const pad = (hi - lo) * 0.08 || hi * 0.1;
  lo = Math.max(0, lo - pad);
  hi += pad;
  const top = (v) => (((hi - v) / (hi - lo)) * 100).toFixed(2) + "%";
  const ticks = niceTicks(lo, hi, 4).filter((t) => t > lo && t < hi);
  const grid = ticks.map((t) => `<div class="k-dist__grid" style="top:${top(t)}"></div><span class="k-dist__y" style="top:${top(t)}">${shortTWD(t)}</span>`).join("");
  const many = ladders.length > 1;
  const dash = (v) => (v == null ? "—" : twd(v));

  let charts = "", rows = "";
  for (const { scheme, cols } of ladders) {
    const shop = scheme.shop || scheme.source;
    let thisPt = null, thisCol = null;
    let dots = "", medians = "", active = 0;
    // A % in `left` resolves against the plot width, so (50 / cols)% is half a column.
    const reach = `clamp(0px, ${(50 / cols.length).toFixed(2)}% - 8px, ${LANE_MAX}px)`;
    cols.forEach((col, i) => {
      const cx = (((i + 0.5) / cols.length) * 100).toFixed(2);
      if (col.stat) medians += `<div class="k-dist__median" style="left:${cx}%;top:${top(col.stat.median_twd)}"></div>`;
      col.pts.forEach((p, j) => {
        const isThis = p.id === thisId;
        const sold = p.status !== "active";
        if (!sold) active++;
        const left = `calc(${cx}% + ${LANES[j % LANES.length]} * ${reach})`;
        const status = (isThis ? "此件 · " : "") + (sold ? "已售出" : "在售");
        const kind = isThis ? "this" : sold ? "sold" : "sale";
        const dot = `<span class="k-dot k-dot--${kind}" style="left:${left};top:${top(p.price_twd)}" data-grade="${esc(gradeName(col.grade))}" data-status="${status}" data-price="${p.price_twd}"></span>`;
        if (isThis) {
          thisPt = { p, left, top: top(p.price_twd), dot };
          thisCol = col;
        } else {
          dots += dot;
        }
      });
    });
    // This listing's dot goes last so it paints above its neighbours.
    if (thisPt) dots += thisPt.dot + `<span class="k-dist__this" style="left:${thisPt.left};top:${thisPt.top}">此件</span>`;

    let summary = `${stats.model_name} ${shop} 各等級的價格分布，在售 ${num(active)} 件`;
    if (thisCol?.stat) summary += `；${gradeName(thisCol.grade)} 中位數 ${twd(thisCol.stat.median_twd)}`;
    if (thisPt) summary += `，此件 ${twd(thisPt.p.price_twd)}`;
    summary += "。可切換為表格檢視。";

    const axis = cols.map((col) => `<div${col === thisCol ? ` class="is-this"` : ""}><b>${esc(gradeName(col.grade))}</b><small>${col.stat ? `在售 ${num(col.stat.count)}` : "—"}</small></div>`).join("");
    // Scheme names lead with the shop ("BRAND OFF 商品ランク"), which the
    // heading already names; the caption keeps only the rest.
    const name = scheme.name?.startsWith(shop) ? scheme.name.slice(shop.length).trim() : scheme.name ?? "";
    const head = many ? `<h3 class="k-dist__shop">${esc(shop)} <small>${esc(name)}${scheme.graded ? ` · ${esc(scheme.graded)}` : ""}</small></h3>` : "";
    charts += `<div class="k-dist__chart">${head}
    <div class="k-dist__plot" role="img" aria-label="${esc(summary)}">${grid}${medians}${dots}<div class="k-dist__tip" role="status" hidden></div></div>
    <div class="k-dist__axis" aria-hidden="true" style="grid-template-columns:repeat(${cols.length},minmax(0,1fr))">${axis}</div>
  </div>`;

    if (many) rows += `<tr><th colspan="6" scope="colgroup">${esc(shop)}</th></tr>`;
    rows += cols.map((col) => {
      const s = col.stat;
      return `<tr><th scope="row">${esc(gradeName(col.grade))}</th><td class="def">${esc(col.grade.desc_zh)}</td><td class="num">${s ? num(s.count) : "—"}</td><td class="num">${dash(s?.median_twd)}</td><td class="num">${dash(s?.min_twd)}</td><td class="num">${dash(s?.max_twd)}</td></tr>`;
    }).join("");
  }

  return `<section class="k-dsec k-dist" aria-labelledby="dist-h">
  <div class="k-dist__head"><h2 id="dist-h">同型號價格分布</h2><button type="button" class="k-dist__toggle" data-dist-toggle>以表格檢視</button></div>
  <div class="k-legend" aria-hidden="true"><span><i class="sale"></i>在售</span><span><i class="sold"></i>已售出</span>${thisId ? `<span><i class="this"></i>此件</span>` : ""}<span><i class="median"></i>中位數（在售）</span></div>
  ${charts}
  <div class="k-dist__table" hidden><table>
    <thead><tr><th scope="col">等級</th><th scope="col">店家定義</th><th scope="col" class="num">在售</th><th scope="col" class="num">中位數</th><th scope="col" class="num">最低</th><th scope="col" class="num">最高</th></tr></thead>
    <tbody>${rows}</tbody>
  </table></div>
</section>`;
}

// Document-level listeners, installed once at module load: re-rendered views
// would stack duplicate listeners (a toggle that flips twice) if each render
// bound its own.
document.addEventListener("click", (e) => {
  const btn = e.target.closest("[data-dist-toggle]");
  if (!btn) return;
  const sec = btn.closest(".k-dist");
  const table = sec.querySelector(".k-dist__table");
  const showTable = table.hidden;
  table.hidden = !showTable;
  for (const chart of sec.querySelectorAll(".k-dist__chart")) chart.hidden = showTable;
  for (const tip of sec.querySelectorAll(".k-dist__tip")) tip.hidden = true;
  btn.textContent = showTable ? "以圖表檢視" : "以表格檢視";
});

document.addEventListener("mouseover", (e) => {
  const dot = e.target.closest?.(".k-dot");
  if (!dot) return;
  const tip = dot.parentElement.querySelector(".k-dist__tip");
  tip.innerHTML = `<span class="ja">${esc(dot.dataset.grade)}</span> · ${esc(dot.dataset.status)}<br><span class="k-price">${twd(+dot.dataset.price)}</span>`;
  tip.style.left = dot.style.left;
  tip.style.top = dot.style.top;
  tip.hidden = false;
});

document.addEventListener("mouseout", (e) => {
  const dot = e.target.closest?.(".k-dot");
  if (dot) dot.parentElement.querySelector(".k-dist__tip").hidden = true;
});
