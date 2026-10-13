import { api, esc, twd, jpy, ymd, md, pct } from "./util.js";
import { valuation, badgeHTML, isTracked, PENDING_ICON, HEART, OUT_ICON } from "./card.js";
import { distHTML } from "./dist.js";

const CONF = { high: "高信心", medium: "中信心", low: "低信心" };
const METHOD = { "same-grade": "同店同等級中位數", "rank-adjusted": "跨等級換算" };
const CONF_DD = { high: "高 · 同等級價格集中", medium: "中", low: "低" };

// Listing detail (商品頁) at /listing/{id}. Every figure comes from the server
// already in TWD; the page converts nothing. A sold listing keeps its CTA (the
// shop page may still be useful) but drops the valuation, tracking and 行情估價,
// since none of them describe something that can still be bought.
export async function renderDetail(main, ov, rawId) {
  let id;
  try {
    id = decodeURIComponent(rawId);
  } catch {
    throw Object.assign(new Error("bad listing id"), { status: 404 });
  }
  const { item: it, model, scheme } = await api(`/api/listings/${encodeURIComponent(id)}`);
  // The distribution is secondary: its failure must not blank the page.
  const group = model && it.group_key
    ? await api(`/api/models/${encodeURIComponent(it.brand)}/${encodeURIComponent(it.group_key)}`).catch(() => null)
    : null;

  const name = it.model_name || it.title;
  const sold = it.status !== "active";
  const v = valuation(it);

  main.innerHTML = `<div class="k-detail${sold ? " k-detail--sold" : ""}">
  ${crumbHTML(it, name)}
  <div class="k-detail__top">
    ${photoHTML(it, name, sold)}
    <div class="k-detail__info">
      <div class="k-detail__id">
        <span class="k-detail__brand">${esc(it.brand)}${it.brand_zh ? ` <span>${esc(it.brand_zh)}</span>` : ""}</span>
        <h1>${esc(name)}</h1>
        <span class="k-detail__sub">${esc([it.category_zh, it.model_code].filter(Boolean).join(" · "))}</span>
      </div>
      ${pricingHTML(it, v, sold)}
      ${actionsHTML(it, sold)}
      ${tagHTML(it, scheme)}
    </div>
  </div>
  <div class="k-detail__col">
    ${specsHTML(it, sold)}
    ${sold ? "" : fairHTML(it, v, name, ov.method)}
    ${group ? distHTML({ stats: group.model, items: group.items, schemes: group.schemes, thisId: it.id }) : ""}
    <section class="k-dsec k-dsec--last" aria-labelledby="hist-h"><h2 id="hist-h">售價變化</h2>${historyHTML(it, sold)}</section>
  </div>
</div>`;
  return `${name} · ${it.brand} · Kura 蔵`;
}

function crumbHTML(it, name) {
  const sep = `<span aria-hidden="true">/</span>`;
  const model = it.group_key
    ? `<a class="k-link k-crumb__model" href="/model/${encodeURIComponent(it.brand)}/${encodeURIComponent(it.group_key)}">${esc(it.brand)} ${esc(name)}</a>`
    : `<span class="k-crumb__model">${esc(it.brand)} ${esc(name)}</span>`;
  const tail = [it.model_code, it.condition?.label].filter(Boolean).join(" · ");
  return `<nav aria-label="位置" class="k-crumb"><a href="/" class="k-link">首頁</a>${sep}${model}${tail ? `${sep}<span>${esc(tail)}</span>` : ""}</nav>`;
}

// The stored gallery when there is one, else the list-page photo alone.
// Thumbnails appear only for two or more photos; one photo needs no picker.
function photoHTML(it, name, sold) {
  const photos = it.images?.length ? it.images : it.photo_pending ? [] : [it.image_url];
  const mat = (inner) => `<div class="k-detail__mat${sold ? " is-sold" : ""}">${inner}</div>`;
  if (!photos.length) return mat(`${PENDING_ICON}<span class="k-card__pending">店家照片準備中</span>`);
  const alt = (i) => `${esc(name)} 商品照片${photos.length > 1 ? ` ${i + 1}／${photos.length}` : "（店家白底照）"}`;
  const main = mat(`<img src="${esc(photos[0])}" alt="${alt(0)}" data-gallery-main>`);
  if (photos.length < 2) return main;
  const thumbs = photos
    .map((u, i) => `<li><button type="button" class="k-gallery__thumb" data-gallery-thumb="${i}" data-src="${esc(u)}" data-alt="${alt(i)}" aria-label="第 ${i + 1} 張照片" aria-current="${i === 0}"><img src="${esc(u)}" alt="" loading="lazy"></button></li>`)
    .join("");
  return `<div class="k-gallery">${main}<ul class="k-gallery__thumbs${sold ? " is-sold" : ""}" aria-label="商品照片">${thumbs}</ul></div>`;
}

// Document-level, as in browse.js: a listener per render would stack.
document.addEventListener("click", (e) => {
  const b = e.target.closest("[data-gallery-thumb]");
  if (!b) return;
  const g = b.closest(".k-gallery");
  const img = g.querySelector("[data-gallery-main]");
  img.src = b.dataset.src;
  img.alt = b.dataset.alt;
  for (const t of g.querySelectorAll("[data-gallery-thumb]")) t.setAttribute("aria-current", String(t === b));
});

function pricingHTML(it, v, sold) {
  const price = `<div class="k-detail__price">${sold ? `<span class="k-detail__lastsold">最後售價</span>` : ""}<span class="k-price">${twd(it.price_twd)}</span><span class="k-detail__jpy">${jpy(it.price_jpy)}</span></div>`;
  const f = it.fair;
  const val = sold
    ? `<span class="k-detail__sold">已找到新主人</span>`
    : badgeHTML(v) + (f ? `<a href="#fair" class="k-detail__fairlink k-link">行情 ${twd(f.fair_twd)} · ${CONF[f.confidence] ?? esc(f.confidence)}</a>` : "");
  return `<div class="k-detail__pricing">${price}<div class="k-detail__val">${val}</div></div>`;
}

function actionsHTML(it, sold) {
  const shop = esc(it.source_name || it.source);
  const cta = `<a class="k-cta" href="${esc(it.url)}" target="_blank" rel="noopener noreferrer">到 ${shop} 商品頁<span class="k-sr">（另開視窗）</span>${OUT_ICON}</a>`;
  const on = isTracked(it.id);
  const track = sold
    ? ""
    : `<button type="button" class="k-dtrack" data-track="${esc(it.id)}" aria-pressed="${on}">${HEART}<span data-track-text>${on ? "追蹤中" : "追蹤這件"}</span></button>`;
  const fine = sold
    ? `Kura 不販售商品。購買、付款與寄送都在 ${shop} 完成。`
    : `追蹤後，降價或售出時會通知你。<br>Kura 不販售商品。購買、付款與寄送都在 ${shop} 完成。`;
  return `<div class="k-detail__actions">${cta}${track}</div><p class="k-detail__fine">${fine}</p>`;
}

// 鑑定卡: the shop's own grade, its definition, and where it sits on the shop's
// ladder. Each part renders only when the shop supplied it. Home's hero hangs
// the same tag without a scheme, so without the ladder.
export function tagHTML(it, scheme) {
  const c = it.condition;
  if (!c?.label) return "";
  const grades = scheme?.grades ?? [];
  const ladder = grades.some((g) => g.code === c.code)
    ? `<div class="k-ladder" role="img" aria-label="${esc("店家等級：" + grades.map((g) => g.label + (g.code === c.code ? "（此件）" : "")).join("、"))}">${grades
        .map((g) => `<span${g.code === c.code ? ` class="is-this"` : ""}>${esc(g.short || g.label)}</span>`)
        .join("")}</div>`
    : "";
  const by = it.condition_by || it.condition_scheme
    ? `<div class="k-tag__foot"><div class="k-tag__by">${it.condition_by ? `<strong>${esc(it.condition_by)}</strong>` : ""}${it.condition_scheme ? `<span>${esc(it.condition_scheme)}</span>` : ""}</div>${it.condition_by === "店家鑑定" ? `<span class="k-tag__stamp" aria-hidden="true">店鑑</span>` : ""}</div>`
    : "";
  return `<div class="k-tag"><div class="k-tag__card">
    <span class="k-tag__hole" aria-hidden="true"></span>
    <div class="k-tag__head"><span>鑑定卡</span><span>${esc(it.source_name || it.source)}</span></div>
    <span class="k-tag__label">${esc(c.label)}</span>
    ${c.desc_zh ? `<span class="k-tag__zh">${esc(c.desc_zh)}</span>` : ""}
    ${c.desc_ja ? `<span class="k-tag__ja">「${esc(c.desc_ja)}」</span>` : ""}
    ${ladder}
    ${by}
  </div></div>`;
}

function specsHTML(it, sold) {
  const listed = sold ? ymd(it.first_seen) : `${ymd(it.first_seen)} · 在架第 ${it.days_listed + 1} 天`;
  const rows = [
    ["型號編號", it.model_code], ["類型", it.category_zh], ["顏色", it.color], ["材質", it.material],
    ["尺寸", it.size], ["付屬品", it.accessories], ["店家備註", it.notes], ["上架", listed],
  ].filter(([, val]) => val);
  return `<section class="k-dsec" aria-labelledby="spec-h"><h2 id="spec-h">商品資料</h2><dl class="k-specs">${rows
    .map(([k, val]) => `<div><dt>${k}</dt><dd>${esc(val)}</dd></div>`)
    .join("")}</dl></section>`;
}

// m: the radar's thresholds from the overview, so this prose and the
// 行情怎麼算 page it links to state one value.
function fairHTML(it, v, name, m) {
  const f = it.fair;
  const how = `<div><a href="/methodology" class="k-link">行情怎麼算</a></div>`;
  let body;
  if (!f) {
    body = `<p>同型號的比較樣本不足 ${m.min_comps} 件，暫不估計行情。</p><dl>${how}</dl>`;
  } else {
    const g = esc(it.condition?.label);
    const n = esc(name);
    let s = f.method === "same-grade"
      ? `同一家店、同樣是 ${g} 的 ${n}，目前有 ${f.comps} 件可以比較，中位數是 <strong>${twd(f.fair_twd)}</strong>。`
      : `同一家店、同樣是 ${g} 的 ${n} 不到 ${m.min_comps} 件，所以把同型號在各家店、各等級的 ${f.comps} 件換算到 ${g} 來比較，估計行情是 <strong>${twd(f.fair_twd)}</strong>。`;
    if (v.state === "anomaly") s += `價格比同型號同成色的行情低超過 ${pct(m.anomaly_pct)}%，通常代表有標題沒寫的瑕疵、缺件或版本差異，購買前請務必確認商品頁的照片與備註。`;
    else if (v.state === "below" || v.state === "low") s += `這件比它低 ${v.pct}%，約省 ${twd(f.fair_twd - it.price_twd)}。`;
    else if (v.state === "above") s += `這件比它高 ${v.pct}%。`;
    else if (v.state === "near") s += "這件的售價接近行情。";
    if (f.confidence === "low") {
      s += f.dispersion > 0.4
        ? "同型號的價格非常分散（可能混有不同皮革、尺寸或限定版），這個估計僅供參考。"
        : "這個估計僅供參考。";
    }
    body = `<p>${s}</p><dl>
      <div><dt>方法</dt><dd>${METHOD[f.method] ?? esc(f.method)}</dd></div>
      <div><dt>樣本</dt><dd>${f.comps} 件</dd></div>
      <div><dt>信心</dt><dd>${CONF_DD[f.confidence] ?? esc(f.confidence)}</dd></div>
      ${how}
    </dl>`;
  }
  return `<section id="fair" class="k-dsec k-fair" aria-labelledby="fair-h"><h2 id="fair-h">行情估價</h2>${body}</section>`;
}

function historyHTML(it, sold) {
  const h = it.history ?? [];
  const end = sold && it.gone_at ? md(it.gone_at) : "今天";
  if (h.length < 2) {
    return `<div class="k-hist">
      <span class="k-hist__line" aria-hidden="true"></span><span class="k-hist__start" aria-hidden="true"></span><span class="k-hist__now" aria-hidden="true"></span>
      <span class="k-hist__label">${md(h[0]?.at ?? it.first_seen)} 上架</span><span class="k-hist__label k-hist__label--end">${end}</span>
      <span class="k-hist__price">${twd(h[0]?.price_twd ?? it.price_twd)}</span>
    </div><p class="k-hist__note">上架以來還沒有調價。店家調價時，會畫在這條線上。</p>`;
  }
  // Step line: a price holds until the next observation, so the line runs flat
  // and drops (or rises) vertically at each change, ending at today (or the
  // day the listing went).
  const t0 = new Date(h[0].at).getTime();
  let t1 = sold && it.gone_at ? new Date(it.gone_at).getTime() : Date.now();
  if (t1 <= t0) t1 = t0 + 1;
  // An observation stamped after the end (recorded later than gone_at, or a
  // server clock ahead of this browser's) is pinned to the right edge.
  // Stretching t1 to cover it instead would move the end of the line off the
  // day its end label names.
  const x =(t) => Math.min(100, ((t - t0) / (t1 - t0)) * 100);
  const ps = h.map((p) => p.price_twd);
  let lo = Math.min(...ps), hi = Math.max(...ps);
  const pad = (hi - lo) * 0.2 || hi * 0.05;
  lo -= pad;
  hi += pad;
  const y = (val) => ((hi - val) / (hi - lo)) * 100;
  const f2 = (n) => n.toFixed(2);

  let d = `M${f2(x(t0))} ${f2(y(h[0].price_twd))}`;
  for (const p of h.slice(1)) d += ` H${f2(x(new Date(p.at).getTime()))} V${f2(y(p.price_twd))}`;
  d += " H100";
  const pts = h.map((p) => {
    const px = x(new Date(p.at).getTime());
    const pos = `left:${f2(px)}%;top:${f2(y(p.price_twd))}%`;
    // Labels near either end align inward so they never spill past the strip.
    const edge = px < 15 ? " k-hist__ptlabel--start" : px > 85 ? " k-hist__ptlabel--end" : "";
    return `<span class="k-hist__pt" style="${pos}" title="${ymd(p.at)} ${twd(p.price_twd)}"></span><span class="k-hist__ptlabel${edge}" style="${pos}">${twd(p.price_twd)}</span>`;
  }).join("");
  const now = `<span class="k-hist__pt k-hist__pt--now" style="left:100%;top:${f2(y(h[h.length - 1].price_twd))}%"></span>`;
  return `<div class="k-hist k-hist--steps">
    <div class="k-hist__plot" aria-hidden="true"><svg viewBox="0 0 100 100" preserveAspectRatio="none"><path d="${d}" fill="none" stroke="var(--k-chart-mark)" stroke-width="2" vector-effect="non-scaling-stroke"/></svg>${pts}${now}</div>
    <span class="k-hist__label">${md(h[0].at)} 上架</span><span class="k-hist__label k-hist__label--end">${end}</span>
    <ul class="k-sr">${h.map((p) => `<li>${ymd(p.at)} ${twd(p.price_twd)}</li>`).join("")}</ul>
  </div>`;
}
