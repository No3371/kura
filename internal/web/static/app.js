"use strict";

const $ = (s, el = document) => el.querySelector(s);
const $$ = (s, el = document) => [...el.querySelectorAll(s)];
const esc = (s) => String(s ?? "").replace(/[&<>"']/g, (c) => ({ "&": "&amp;", "<": "&lt;", ">": "&gt;", '"': "&quot;", "'": "&#39;" })[c]);
const nf = new Intl.NumberFormat("zh-TW");
const twd = (n) => "NT$" + nf.format(Math.round(n));
const jpy = (n) => "¥" + nf.format(n);
const pct = (f) => Math.round(f * 100) + "%";
// Internal cross-source scale, used only to order grades best-first.
const RANKS = ["N", "S", "SA", "A", "AB", "B", "BC", "C"];
const gradeKey = (it) => it.source + "|" + (it.condition?.code || "");
const gradeLabel = (c) => c?.label || "未標示";
const CONF_ZH = { high: "高信心", medium: "中信心", low: "低信心" };

async function api(path, opts) {
  const r = await fetch(path, opts);
  const body = await r.json().catch(() => ({}));
  if (!r.ok) throw new Error(body.error || r.statusText);
  return body;
}

let OV = null; // last /api/overview

/* ============================== tabs ============================== */
function showTab(name) {
  if (!["radar", "advisor", "pipeline"].includes(name)) name = "radar";
  $$("nav.tabs button").forEach((b) => b.setAttribute("aria-selected", String(b.dataset.tab === name)));
  $$("main > section").forEach((s) => (s.hidden = s.id !== "tab-" + name));
  if (name === "pipeline") renderRuns();
  if (name === "advisor") $("#input").focus();
}
$$("nav.tabs button").forEach((b) => b.addEventListener("click", () => (location.hash = b.dataset.tab)));
window.addEventListener("hashchange", () => showTab(location.hash.slice(1)));

/* ============================ overview ============================ */
async function loadOverview() {
  OV = await api("/api/overview");
  const o = OV.overview;
  const fx = OV.fx;
  const fxDate = fx.AsOf && !fx.AsOf.startsWith("0001") ? new Date(fx.AsOf).toLocaleDateString("zh-TW") : "預設值";
  const mock = OV.advisor === "mock";
  $("#status").innerHTML = `
    <span class="pill num">1 JPY = ${fx.Rate.toFixed(4)} TWD · ${esc(fxDate)}</span>
    <span class="pill ${mock ? "live mock" : "live"}">${mock ? "顧問：示範模式" : "顧問：" + esc(OV.advisor.replace("claude:", ""))}</span>
    ${OV.crawling ? '<span class="pill">爬取中…</span>' : ""}`;
  $("#tiles").innerHTML = [
    ["在售包款", nf.format(o.active), `${Object.keys(o.by_brand || {}).length} 個品牌`],
    ["已售出（觀察到）", nf.format(o.sold), o.since && !o.since.startsWith("0001") ? "自 " + new Date(o.since).toLocaleDateString("zh-TW") + " 起追蹤" : "尚未開始追蹤"],
    ["有行情估價", nf.format(o.valued), o.active ? pct(o.valued / o.active) + " 的在售商品" : ""],
    ["低於行情 ≥10%", nf.format(o.deals), "依同型號同成色中位數"],
    ["近期降價", nf.format(o.drops), "最近一次觀察為降價"],
  ].map(([k, v, s]) => `<div class="panel tile"><div class="k">${k}</div><div class="v num">${v}</div><div class="s">${esc(s)}</div></div>`).join("");

  const bSel = $('#filters select[name="brand"]');
  if (bSel.options.length === 1) {
    for (const b of OV.brands) bSel.add(new Option(`${b.zh} ${b.name}`, b.name));
    const cSel = $('#filters select[name="category"]');
    for (const c of OV.categories) if (c.key !== "bag") cSel.add(new Option(c.zh, c.key));
  }
  $("#crawl-btn").disabled = !OV.can_crawl || OV.crawling;
}

/* ============================== cards ============================== */
function dealTag(it) {
  if (!it.fair) return "";
  const d = it.fair.deal_pct;
  if (it.fair.anomaly)
    return `<span class="tag drop" title="比行情 ${twd(it.fair.fair_twd)} 低 ${pct(d)}，常見原因是瑕疵、缺件或版本不同">⚠ 異常低價 · 請先確認狀況</span>`;
  if (d >= 0.03 && it.fair.confidence === "low")
    return `<span class="tag" title="同型號價格分散（${it.fair.comps} 件樣本），可能混有不同皮革或版本">可能低於行情 ${pct(d)} · 低信心</span>`;
  if (d >= 0.03) return `<span class="tag deal" title="行情 ${twd(it.fair.fair_twd)}（${it.fair.comps} 件樣本，${CONF_ZH[it.fair.confidence]}）">▼ 低於行情 ${pct(d)}</span>`;
  if (d <= -0.03) return `<span class="tag over" title="行情 ${twd(it.fair.fair_twd)}">高於行情 ${pct(-d)}</span>`;
  return `<span class="tag" title="行情 ${twd(it.fair.fair_twd)}">接近行情</span>`;
}
function dropTag(it) {
  return it.drop ? `<span class="tag drop">降價 ${pct(it.drop.pct)}</span>` : "";
}
function card(it) {
  return `<button class="card" type="button" data-id="${esc(it.id)}">
    <div class="img"><img loading="lazy" src="${esc(it.image_url)}" alt="${esc(it.model_name)}">
      <span class="rank" title="${esc(it.condition_scheme)}：${esc(it.condition.desc_ja)}">${esc(gradeLabel(it.condition))}</span>${it.status !== "active" ? '<span class="sold-flag">已售出</span>' : ""}</div>
    <div class="body">
      <div class="b">${esc(it.brand)}</div>
      <div class="m">${esc(it.model_name || it.title)}</div>
      <div class="meta">${esc(it.category_zh)}${it.model_code ? " · " + esc(it.model_code) : ""}</div>
      ${it.condition.desc_zh ? `<div class="cond">${esc(it.condition.desc_zh)}</div>` : ""}
      <div class="p num">${twd(it.price_twd)}<small>${jpy(it.price_jpy)}</small></div>
      <div class="tags">${dealTag(it)}${dropTag(it)}</div>
    </div></button>`;
}
document.addEventListener("click", (e) => {
  const c = e.target.closest("[data-id]");
  if (c) openListing(c.dataset.id);
  const m = e.target.closest("[data-model]");
  if (m) openModel(m.dataset.brand, m.dataset.model);
});

/* ============================== radar ============================== */
let searchTimer;
$("#filters").addEventListener("input", () => {
  clearTimeout(searchTimer);
  searchTimer = setTimeout(loadListings, 250);
});
$("#filters").addEventListener("submit", (e) => e.preventDefault());

async function loadListings() {
  const fd = new FormData($("#filters"));
  const qs = new URLSearchParams();
  for (const [k, v] of fd) if (v) qs.set(k, v);
  qs.set("limit", "60");
  const { total, items } = await api("/api/listings?" + qs);
  $("#result-meta").textContent = total ? `共 ${nf.format(total)} 件，顯示前 ${items.length} 件` : "";
  $("#grid").innerHTML = items.length ? items.map(card).join("") :
    `<div class="empty panel">${OV && OV.overview.active === 0 ? "還沒有資料。到「資料管線」按「立即爬取」，或執行 <code>kura crawl</code>。" : "沒有符合條件的商品，試著放寬條件。"}</div>`;
  const brand = fd.get("brand") || "";
  loadSide(brand);
}

async function loadSide(brand) {
  const q = brand ? "?brand=" + encodeURIComponent(brand) : "";
  const [models, drops] = await Promise.all([api("/api/models" + (q ? q + "&" : "?") + "limit=12"), api("/api/drops" + (q ? q + "&" : "?") + "limit=8")]);
  $("#models").innerHTML = models.filter((m) => m.active > 0).map((m) => `<li><button type="button" data-model="${esc(m.group_key)}" data-brand="${esc(m.brand)}">
      <span class="n">${esc(m.model_name || m.group_key)}</span><span class="r num">${twd(m.median_jpy * OV.fx.Rate)}</span>
      <span class="c">${esc(m.brand)}${m.model_code ? " · " + esc(m.model_code) : ""}</span><span class="c r">在售 ${m.active}${m.sold ? " · 售出 " + m.sold : ""}</span>
    </button></li>`).join("") || '<li class="muted" style="font-size:13px">尚無資料</li>';
  $("#drops").innerHTML = drops.map((it) => `<li><button type="button" data-id="${esc(it.id)}">
      <span class="n">${esc(it.model_name)}</span><span class="r num">${twd(it.price_twd)}</span>
      <span class="c">${esc(it.brand)} · ${esc(gradeLabel(it.condition))}</span><span class="c r">${jpy(it.drop.from_jpy)} → ${jpy(it.drop.to_jpy)}</span>
    </button></li>`).join("") || '<li class="muted" style="font-size:13px">還沒有觀察到降價，需要多次爬取才會出現。</li>';
}

/* ============================== drawer ============================== */
function openDrawer(title, html) {
  $("#drawer-title").textContent = title;
  $("#drawer-body").innerHTML = html;
  $("#drawer").hidden = $("#scrim").hidden = false;
  $("#drawer").scrollTop = 0;
}
function closeDrawer() { $("#drawer").hidden = $("#scrim").hidden = true; }
$("#drawer-close").addEventListener("click", closeDrawer);
$("#scrim").addEventListener("click", closeDrawer);
document.addEventListener("keydown", (e) => e.key === "Escape" && closeDrawer());

async function openListing(id) {
  const { item: it, model } = await api("/api/listings/" + encodeURIComponent(id));
  const f = it.fair;
  const c = it.condition;
  const condHTML = `<span class="cond-label">${esc(gradeLabel(c))}</span>
    ${c.desc_zh ? `<span class="cond-text">${esc(c.desc_zh)}</span>` : ""}
    ${c.desc_ja ? `<span class="cond-src">店家原文：「${esc(c.desc_ja)}」— ${esc(it.condition_scheme)}，${esc(it.condition_by)}</span>` : ""}`;
  const specs = [["型號", it.model_code], ["包款", it.category_zh], ["成色", condHTML, true], ["顏色", it.color], ["材質", it.material],
    ["尺寸", it.size], ["配件", it.accessories], ["備註", it.notes], ["上架", `${new Date(it.first_seen).toLocaleDateString("zh-TW")}（已 ${it.days_listed} 天）`]]
    .filter(([, v]) => v).map(([k, v, raw]) => `<dt>${k}</dt><dd>${raw ? v : esc(v)}</dd>`).join("");
  let html = `<div class="detail-top">
      <img src="${esc(it.image_url)}" alt="${esc(it.model_name)}">
      <div>
        <div class="muted" style="letter-spacing:.08em;font-size:12px">${esc(it.brand)} ${esc(it.brand_zh)}</div>
        <h3>${esc(it.model_name || it.title)}</h3>
        <div class="price-big num">${twd(it.price_twd)} <span class="muted" style="font-size:15px;font-weight:400">${jpy(it.price_jpy)}</span></div>
        <div class="tags" style="margin:6px 0 12px">${dealTag(it)}${dropTag(it)}${it.status !== "active" ? '<span class="tag">已售出</span>' : ""}</div>
        <dl class="specs">${specs}</dl>
        <p style="margin:14px 0 0"><a class="btn" href="${esc(it.url)}" target="_blank" rel="noopener">前往 BRAND OFF 商品頁 ↗</a></p>
      </div></div>`;
  if (f) {
    html += `<div class="panel fairbox">
      <div class="line"><strong>行情估計 ${twd(f.fair_twd)}</strong><span class="muted">${CONF_ZH[f.confidence]} · ${f.comps} 件樣本 · ${f.method === "same-rank" ? "同成色中位數" : "跨成色換算"}</span></div>
      <div class="ink2" style="font-size:14px">${f.deal_pct >= 0 ? `售價比行情低 ${pct(f.deal_pct)}，約省 ${twd(f.fair_twd - it.price_twd)}。` : `售價比行情高 ${pct(-f.deal_pct)}。`}
      ${f.anomaly ? "價格比同型號同成色的行情低一半以上，通常代表有標題沒寫的瑕疵、缺件或版本差異，購買前請務必確認商品頁的照片與備註。" : ""}
      ${f.confidence === "low" ? (f.dispersion > 0.5 ? "同型號的價格非常分散（可能混有不同皮革、尺寸或限定版），這個估計僅供參考。" : "樣本較少，僅供參考。") : ""}</div></div>`;
  } else if (it.status === "active") {
    html += `<div class="panel fairbox ink2" style="font-size:14px">同型號的比較樣本不足 3 件，暫不估計行情。</div>`;
  }
  html += `<div class="panel chart" id="hist-chart"></div>`;
  if (model) html += `<div class="panel chart" id="model-chart"></div>`;
  openDrawer(it.brand + " · " + (it.model_code || it.model_name), html);
  historyChart($("#hist-chart"), it.history || []);
  if (model) {
    const m = await api(`/api/models/${encodeURIComponent(it.brand)}/${encodeURIComponent(it.group_key)}`);
    stripChart($("#model-chart"), m.model, m.items, it.id);
  }
}

async function openModel(brand, key) {
  const m = await api(`/api/models/${encodeURIComponent(brand)}/${encodeURIComponent(key)}`);
  const s = m.model;
  const days = s.median_days_to_sell != null ? `${s.median_days_to_sell.toFixed(1)} 天` : "資料不足";
  const html = `<div class="tiles" style="grid-template-columns:repeat(4,minmax(0,1fr));margin:0">
      ${[["在售", s.active], ["已售出", s.sold], ["在售中位數", twd(s.median_jpy * OV.fx.Rate)], ["售出中位天數", days]]
        .map(([k, v]) => `<div class="panel tile"><div class="k">${k}</div><div class="v num" style="font-size:20px">${v}</div></div>`).join("")}
    </div>
    <div class="panel chart" id="model-chart"></div>
    <div class="grid">${m.items.slice(0, 24).map(card).join("")}</div>`;
  openDrawer(`${s.brand} · ${s.model_name || s.group_key}`, html);
  stripChart($("#model-chart"), s, m.items, null);
}

/* ============================== charts ============================== */
// Shared tooltip positioned inside a .chart panel.
function tooltip(host) {
  const tip = document.createElement("div");
  tip.className = "tooltip";
  tip.hidden = true;
  host.appendChild(tip);
  return {
    show(html, x, y) {
      tip.innerHTML = html; tip.hidden = false;
      const w = tip.offsetWidth, hw = host.clientWidth;
      tip.style.left = Math.min(Math.max(x - w / 2, 4), hw - w - 4) + "px";
      tip.style.top = y - tip.offsetHeight - 10 + "px";
    },
    hide() { tip.hidden = true; },
  };
}
function niceTicks(lo, hi, n = 4) {
  const span = hi - lo || hi || 1;
  const step0 = span / n, mag = 10 ** Math.floor(Math.log10(step0));
  const step = [1, 2, 2.5, 5, 10].map((m) => m * mag).find((s) => s >= step0);
  const t = [];
  for (let v = Math.ceil(lo / step) * step; v <= hi + 1e-9; v += step) t.push(v);
  return t;
}
const shortTWD = (v) => (v >= 10000 ? (v / 10000).toFixed(v % 10000 ? 1 : 0) + " 萬" : nf.format(v));

// Dot strip: one column per source grade (the shop's own label), best first;
// y = price (TWD). Active listings are filled series-1 dots, sold ones hollow
// series-2 rings (shape + colour), the median of active listings is a tick.
function stripChart(host, stats, items, selfId) {
  const rate = OV.fx.Rate;
  const pts = items.filter((i) => i.condition?.code).map((i) => ({ ...i, twd: Math.round(i.price_jpy * rate), col: gradeKey(i) }));
  const colGrade = new Map(pts.map((p) => [p.col, p.condition]));
  const ranks = [...colGrade.keys()].sort((a, b) => RANKS.indexOf(colGrade.get(a).rank) - RANKS.indexOf(colGrade.get(b).rank) || a.localeCompare(b));
  host.innerHTML = `<div class="chart-title"><h4>同型號價格分布（依成色）</h4>
    <div class="legend"><span><i class="active"></i>在售</span><span><i class="sold"></i>已售出</span>${selfId ? '<span><i class="self"></i>此商品</span>' : ""}<span>— 在售中位數</span></div></div>`;
  if (!pts.length) { host.insertAdjacentHTML("beforeend", '<p class="muted">沒有資料</p>'); return; }
  const W = 680, H = 260, m = { l: 56, r: 12, t: 12, b: 28 };
  const iw = W - m.l - m.r, ih = H - m.t - m.b;
  const prices = pts.map((p) => p.twd);
  let lo = Math.min(...prices), hi = Math.max(...prices);
  const pad = (hi - lo) * 0.08 || hi * 0.1;
  lo = Math.max(0, lo - pad); hi += pad;
  const ticks = niceTicks(lo, hi);
  const y = (v) => m.t + ih - ((v - lo) / (hi - lo)) * ih;
  const band = iw / ranks.length;
  const cx = (r) => m.l + band * (ranks.indexOf(r) + 0.5);
  // Deterministic jitter so the chart is stable between renders.
  const jit = (id) => { let h = 0; for (const c of id) h = (h * 31 + c.charCodeAt(0)) | 0; return ((h % 1000) / 1000 - 0.5) * Math.min(band * 0.5, 60); };
  let svg = `<svg class="viz" viewBox="0 0 ${W} ${H}" role="img" aria-label="價格分布圖">`;
  svg += `<g class="grid">${ticks.map((t) => `<line x1="${m.l}" x2="${W - m.r}" y1="${y(t)}" y2="${y(t)}"/>`).join("")}</g>`;
  svg += ticks.map((t) => `<text x="${m.l - 8}" y="${y(t) + 4}" text-anchor="end" class="num">${shortTWD(t)}</text>`).join("");
  svg += `<line class="base" x1="${m.l}" x2="${W - m.r}" y1="${m.t + ih}" y2="${m.t + ih}"/>`;
  svg += ranks.map((r) => `<text x="${cx(r)}" y="${H - 8}" text-anchor="middle">${esc(colGrade.get(r).label)}</text>`).join("");
  for (const r of ranks) {
    const act = pts.filter((p) => p.col === r && p.status === "active").map((p) => p.twd).sort((a, b) => a - b);
    if (act.length >= 2) {
      const md = act.length % 2 ? act[(act.length - 1) / 2] : (act[act.length / 2 - 1] + act[act.length / 2]) / 2;
      svg += `<line class="median" x1="${cx(r) - band * 0.32}" x2="${cx(r) + band * 0.32}" y1="${y(md)}" y2="${y(md)}"/>`;
    }
  }
  // Sold first so active dots sit on top; the selected listing last.
  const order = [...pts].sort((a, b) => (a.id === selfId) - (b.id === selfId) || (a.status === "active") - (b.status === "active"));
  svg += order.map((p, i) => {
    const x = cx(p.col) + jit(p.id), yy = y(p.twd);
    const cls = p.status === "active" ? "dot-active" : "dot-sold";
    const self = p.id === selfId ? " dot-self" : "";
    return `<circle class="${cls}${self}" cx="${x}" cy="${yy}" r="${p.id === selfId ? 7 : 5}"/>
      <circle class="hit" data-i="${pts.indexOf(p)}" cx="${x}" cy="${yy}" r="11"/>`;
  }).join("");
  svg += "</svg>";
  host.insertAdjacentHTML("beforeend", svg);
  const tip = tooltip(host);
  const svgEl = host.querySelector("svg");
  svgEl.addEventListener("mousemove", (e) => {
    const h = e.target.closest(".hit");
    if (!h) return tip.hide();
    const p = pts[+h.dataset.i];
    const box = host.getBoundingClientRect(), c = h.getBoundingClientRect();
    tip.show(`<b class="num">${twd(p.twd)}</b> · ${esc(gradeLabel(p.condition))} · ${p.status === "active" ? "在售" : "已售出"}<br><span class="muted">${esc(p.condition.desc_zh || p.color || p.model_name)}</span>`,
      c.left - box.left + c.width / 2, c.top - box.top);
  });
  svgEl.addEventListener("mouseleave", () => tip.hide());
  svgEl.addEventListener("click", (e) => { const h = e.target.closest(".hit"); if (h && pts[+h.dataset.i].id !== selfId) openListing(pts[+h.dataset.i].id); });
  host.insertAdjacentHTML("beforeend", `<details class="table-view"><summary>以表格檢視</summary><table class="data">
    <tr><th>成色（店家評級）</th><th>店家定義</th><th class="r">在售</th><th class="r">中位數</th><th class="r">最低</th><th class="r">最高</th></tr>
    ${(stats.by_grade || []).map((r) => `<tr><td>${esc(gradeLabel(r.grade))}</td><td class="ink2">${esc(r.grade.desc_zh)}</td><td class="r num">${r.count}</td><td class="r num">${twd(r.median_jpy * rate)}</td>
      <td class="r num">${twd(r.min_jpy * rate)}</td><td class="r num">${twd(r.max_jpy * rate)}</td></tr>`).join("")}
    </table></details>`);
}

// Asking-price history as a step line (prices change discretely).
function historyChart(host, hist) {
  const rate = OV.fx.Rate;
  host.innerHTML = `<div class="chart-title"><h4>售價變化</h4><span class="muted" style="font-size:12px">以目前匯率換算</span></div>`;
  if (hist.length < 2) {
    host.insertAdjacentHTML("beforeend", `<p class="muted" style="margin:4px 0 6px;font-size:14px">自 ${hist[0] ? new Date(hist[0].At).toLocaleDateString("zh-TW") : "—"} 首次看到以來，售價沒有變動。</p>`);
    return;
  }
  const pts = [...hist, { At: new Date().toISOString(), PriceJPY: hist[hist.length - 1].PriceJPY, now: true }]
    .map((p) => ({ t: new Date(p.At).getTime(), v: Math.round(p.PriceJPY * rate), now: p.now }));
  const W = 680, H = 180, m = { l: 56, r: 16, t: 12, b: 26 };
  const iw = W - m.l - m.r, ih = H - m.t - m.b;
  const t0 = pts[0].t, t1 = pts[pts.length - 1].t || t0 + 1;
  let lo = Math.min(...pts.map((p) => p.v)), hi = Math.max(...pts.map((p) => p.v));
  const pad = (hi - lo) * 0.2 || hi * 0.05; lo -= pad; hi += pad;
  const x = (t) => m.l + ((t - t0) / (t1 - t0 || 1)) * iw;
  const y = (v) => m.t + ih - ((v - lo) / (hi - lo)) * ih;
  const ticks = niceTicks(lo, hi, 3);
  let d = `M${x(pts[0].t)},${y(pts[0].v)}`;
  for (let i = 1; i < pts.length; i++) d += ` H${x(pts[i].t)} V${y(pts[i].v)}`;
  const fmtD = (t) => new Date(t).toLocaleDateString("zh-TW", { month: "numeric", day: "numeric" });
  let svg = `<svg class="viz" viewBox="0 0 ${W} ${H}" role="img" aria-label="售價變化">`;
  svg += `<g class="grid">${ticks.map((t) => `<line x1="${m.l}" x2="${W - m.r}" y1="${y(t)}" y2="${y(t)}"/>`).join("")}</g>`;
  svg += ticks.map((t) => `<text x="${m.l - 8}" y="${y(t) + 4}" text-anchor="end" class="num">${shortTWD(t)}</text>`).join("");
  svg += `<text x="${m.l}" y="${H - 6}">${fmtD(t0)}</text><text x="${W - m.r}" y="${H - 6}" text-anchor="end">今天</text>`;
  svg += `<path class="line" d="${d}"/>`;
  svg += pts.filter((p) => !p.now).map((p, i) => `<circle class="dot-active" cx="${x(p.t)}" cy="${y(p.v)}" r="5"/><circle class="hit" data-i="${i}" cx="${x(p.t)}" cy="${y(p.v)}" r="11"/>`).join("");
  svg += "</svg>";
  host.insertAdjacentHTML("beforeend", svg);
  const tip = tooltip(host);
  const svgEl = host.querySelector("svg");
  svgEl.addEventListener("mousemove", (e) => {
    const h = e.target.closest(".hit");
    if (!h) return tip.hide();
    const p = pts[+h.dataset.i], src = hist[+h.dataset.i];
    const box = host.getBoundingClientRect(), c = h.getBoundingClientRect();
    tip.show(`<b class="num">${twd(p.v)}</b> · ${jpy(src.PriceJPY)}<br><span class="muted">${new Date(p.t).toLocaleString("zh-TW")}</span>`, c.left - box.left + c.width / 2, c.top - box.top);
  });
  svgEl.addEventListener("mouseleave", () => tip.hide());
}

/* ============================== advisor ============================== */
const session = (() => {
  try {
    let s = sessionStorage.getItem("kura-session");
    if (!s) { s = crypto.randomUUID(); sessionStorage.setItem("kura-session", s); }
    return s;
  } catch { return crypto.randomUUID(); }
})();
const TOOL_ZH = { search_listings: "搜尋在售商品", model_price_stats: "查詢型號行情", get_listing: "讀取商品詳情", market_overview: "查看市場概況" };

// Minimal, safe markdown: escape first, then **bold**, "- " bullets,
// paragraphs, and [[listing-id]] references rendered as clickable chips.
function renderMD(src, known) {
  const lines = esc(src).split("\n");
  let html = "", inList = false;
  const inline = (s) => s
    .replace(/\*\*(.+?)\*\*/g, "<b>$1</b>")
    .replace(/\[\[([a-z]+:[\w-]+)\]\]/g, (_, id) => {
      const it = known.get(id);
      return `<span class="ref" data-id="${id}" role="button" tabindex="0">${it ? "▸ " + esc(it.model_name) : "▸ 查看商品"}</span>`;
    });
  for (const ln of lines) {
    const m = ln.match(/^\s*(?:[-*•]|\d+\.)\s+(.*)$/);
    if (m) { if (!inList) { html += "<ul>"; inList = true; } html += `<li>${inline(m[1])}</li>`; continue; }
    if (inList) { html += "</ul>"; inList = false; }
    if (ln.trim()) html += `<p>${inline(ln.replace(/^#+\s*/, ""))}</p>`;
  }
  return html + (inList ? "</ul>" : "");
}

let busy = false;
async function send(text) {
  text = text.trim();
  if (!text || busy) return;
  busy = true; $("#send").disabled = true; $("#input").value = "";
  const msgs = $("#msgs");
  msgs.insertAdjacentHTML("beforeend", `<div class="msg user">${esc(text)}</div>`);
  const bot = document.createElement("div");
  bot.className = "msg bot";
  bot.innerHTML = `<div class="steps"></div><div class="text typing"></div><div class="strip-host"></div>`;
  msgs.appendChild(bot);
  const steps = $(".steps", bot), textEl = $(".text", bot), stripHost = $(".strip-host", bot);
  const known = new Map();
  let raw = "";
  const scroll = () => (msgs.scrollTop = msgs.scrollHeight);
  scroll();
  try {
    const resp = await fetch("/api/chat", { method: "POST", headers: { "Content-Type": "application/json" }, body: JSON.stringify({ session, message: text }) });
    if (!resp.ok) throw new Error((await resp.json().catch(() => ({}))).error || resp.statusText);
    const reader = resp.body.getReader(), dec = new TextDecoder();
    let buf = "";
    for (;;) {
      const { value, done } = await reader.read();
      if (done) break;
      buf += dec.decode(value, { stream: true });
      let i;
      while ((i = buf.indexOf("\n\n")) >= 0) {
        const chunk = buf.slice(0, i); buf = buf.slice(i + 2);
        if (!chunk.startsWith("data: ")) continue;
        const ev = JSON.parse(chunk.slice(6));
        if (ev.type === "text") { raw += ev.text; textEl.innerHTML = renderMD(raw, known); }
        else if (ev.type === "tool_call") steps.insertAdjacentHTML("beforeend", `<div class="step">⋯ ${TOOL_ZH[ev.tool] || esc(ev.tool)} <code>${esc(ev.input.length > 140 ? ev.input.slice(0, 140) + "…" : ev.input)}</code></div>`);
        else if (ev.type === "tool_result") steps.insertAdjacentHTML("beforeend", `<div class="step ${ev.is_error ? "err" : ""}">${ev.is_error ? "⚠" : "✓"} ${esc(ev.summary)}</div>`);
        else if (ev.type === "listings") for (const it of ev.listings) known.set(it.id, it);
        else if (ev.type === "error") textEl.insertAdjacentHTML("afterend", `<div class="err">${esc(ev.text)}</div>`);
        scroll();
      }
    }
  } catch (err) {
    textEl.insertAdjacentHTML("afterend", `<div class="err">${esc(err.message)}</div>`);
  }
  textEl.classList.remove("typing");
  textEl.innerHTML = renderMD(raw, known);
  // Show cards for the listings the advisor actually referenced (else all it saw).
  const refs = [...raw.matchAll(/\[\[([a-z]+:[\w-]+)\]\]/g)].map((m) => m[1]).filter((id, i, a) => a.indexOf(id) === i && known.has(id));
  const show = (refs.length ? refs : [...known.keys()].slice(0, 8)).map((id) => known.get(id));
  if (show.length) stripHost.innerHTML = `<div class="strip">${show.map(card).join("")}</div>`;
  busy = false; $("#send").disabled = false; scroll();
}
$("#composer").addEventListener("submit", (e) => { e.preventDefault(); send($("#input").value); });
$("#input").addEventListener("keydown", (e) => { if (e.key === "Enter" && !e.shiftKey && !e.isComposing) { e.preventDefault(); send($("#input").value); } });
$$(".suggest button").forEach((b) => b.addEventListener("click", () => send(b.dataset.q)));
document.addEventListener("keydown", (e) => { if (e.key === "Enter" && e.target.classList?.contains("ref")) openListing(e.target.dataset.id); });

/* ============================== pipeline ============================== */
async function renderRuns() {
  await loadOverview();
  const rows = OV.runs || [];
  const t = (s) => (s && !s.startsWith("0001") ? new Date(s).toLocaleString("zh-TW") : "—");
  $("#runs").innerHTML = `<tr><th>開始</th><th>品牌</th><th class="r">頁數</th><th class="r">包款</th><th class="r">新上架</th><th class="r">價格變動</th><th class="r">判定售出</th><th>狀態</th></tr>` +
    (rows.map((r) => `<tr><td class="num">${t(r.StartedAt)}</td><td>${esc(r.Target)}</td><td class="r num">${r.Pages}</td><td class="r num">${r.Seen}</td>
      <td class="r num">${r.New}</td><td class="r num">${r.PriceChanges}</td><td class="r num">${r.Gone}</td>
      <td>${r.Error ? `<span class="run-partial">⚠ ${esc(r.Error)}</span>` : !r.FinishedAt || r.FinishedAt.startsWith("0001") ? "進行中" : r.Complete ? '<span class="run-ok">✓ 完整</span>' : '<span class="run-partial">部分（未判定售出）</span>'}</td></tr>`).join("") ||
      '<tr><td colspan="8" class="muted">尚無紀錄</td></tr>');
}
$("#crawl-btn").addEventListener("click", async () => {
  try { await api("/api/crawl", { method: "POST" }); } catch (e) { alert(e.message); }
  renderRuns();
});

/* ============================== boot ============================== */
(async () => {
  showTab(location.hash.slice(1));
  await loadOverview();
  await loadListings();
  setInterval(async () => { if (OV.crawling) { await loadOverview(); } }, 15000);
})();
