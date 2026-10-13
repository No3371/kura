// Card-state fixtures. Listings and photos are real BRAND OFF samples from the
// design handoff; drop, sold and photo-pending have no real instance yet, so
// this page is the only place those states can be checked. It is kept off the
// product pages, which show live data only.
import { cardHTML } from "/js/card.js";

const PHOTO = "https://www.brandoff-store.com/Contents/ProductImages/0/";
const SHOP = "https://www.brandoff-store.com/";

const grade = {
  A: { code: "ランクA", label: "ランクA", desc_ja: "多少のキズ・汚れ・若干の使用感等あり", desc_zh: "有少許刮痕、髒污與輕微使用感", rank: "A" },
  B: { code: "ランクB", label: "ランクB", desc_ja: "全体的にキズ・汚れ・使用感等あり", desc_zh: "整體有刮痕、髒污與使用感", rank: "B" },
};

const base = (id, brand, model, code, g, jpy, twd, photo) => ({
  id: "sample:" + id, url: SHOP, image_url: PHOTO + photo, source: "brandoff", source_name: "BRAND OFF",
  brand, title: model, model_name: model, model_code: code, condition: grade[g],
  price_jpy: jpy, price_twd: twd, status: "active",
});

const pico = base("pico", "HERMES", "ピコタンロックPM", "056289CK", "A", 520000, 104955, "2101220285869_L.jpg");
const matra = base("matra", "CHANEL", "マトラッセ", "A33814", "A", 240000, 48441, "2120900081669_L.jpg");
const never = base("never", "LOUIS VUITTON", "ネヴァーフルMM", "M40156", "B", 130000, 26239, "2101220126346_L.jpg");
const garden = base("garden", "HERMES", "ガーデンパーティPM", "", "A", 240000, 48441, "2101220247300_L.jpg");
const alma = base("alma", "LOUIS VUITTON", "アルマPM モノグラム", "M51130", "B", 22000, 4440, "2118700085789_L.jpg");

const fair = (deal_pct, confidence, extra = {}) => ({ fair: { deal_pct, confidence, anomaly: false, ...extra } });

const STATES = [
  ["預設 · 接近行情", { ...never, ...fair(0.01, "high") }],
  ["低於行情 · 高信心", { ...pico, ...fair(0.31, "high") }],
  ["可能低於行情 · 低信心", { ...garden, ...fair(0.47, "low") }],
  ["異常低價", { ...alma, ...fair(0.72, "high", { anomaly: true }) }],
  ["高於行情", { ...never, ...fair(-0.12, "high") }],
  ["尚無行情估價", { ...matra }],
  ["降價", { ...never, ...fair(0.32, "high"), drop: { from_twd: 27854, to_twd: 26239 } }],
  ["已找到新主人", { ...pico, ...fair(0.31, "high"), status: "gone" }],
  ["照片準備中", { ...matra, ...fair(0.49, "high"), image_url: PHOTO + "NowPrinting_L.jpg", photo_pending: true }],
];

for (const grid of document.querySelectorAll("[data-rows] .dev-grid")) {
  grid.innerHTML = STATES.map(([label, it]) =>
    `<div class="dev-cell"><span class="dev-label">${label}</span>${cardHTML(it)}</div>`).join("");
}
document.getElementById("row-card").innerHTML = cardHTML({ ...pico, ...fair(0.31, "high") }, { layout: "row" });
