import { esc, ymd, pct } from "./util.js";
import { badgeHTML, NEAR_BAND } from "./card.js";

// 行情怎麼算: what each kind of number on Kura is and where it comes from.
//
// Every threshold this page states is read from the overview's method block
// (the radar's own constants) or card.js NEAR_BAND, never typed into the copy:
// a typed number would keep describing the old rule after the code changed.
// The page names no internal rank and makes no claim about how fresh the data
// is — how grades map across shops and what freshness to promise are not
// settled, and stating either here would read as a commitment.
export function renderMethodology(el, ov) {
  const m = ov.method;
  const band = pct(NEAR_BAND);
  const deal = pct(m.deal_min_pct);
  const anomaly = pct(m.anomaly_pct);
  const n = m.min_comps;
  const days = m.min_sold_history_days;
  const schemes = ov.schemes || [];
  const shops = [...new Set(schemes.map((s) => s.shop))];
  // An empty DB has no epoch, and Go encodes the unset time as
  // "0001-01-01T00:00:00Z": a truthiness check on the string would accept it
  // and print 「自 0001/01/01 起記錄」. Date.parse gives that a negative value
  // and a missing field NaN, so `> 0` rejects both.
  const since = Date.parse(ov.overview?.since) > 0 ? ov.overview.since : null;

  const source = [
    shops.length ? `目前的資料來自 ${shops.map(esc).join("、")} 的線上商店。` : "",
    "商品、售價、照片與成色等級，都照店家刊登的內容呈現；等級說明的中文是 Kura 的翻譯，「」內是店家的原文。",
    since ? `Kura 自 ${ymd(since)} 起記錄。` : "",
  ].join("");

  const ladders = schemes.map((s) => {
    const ref = s.ref
      ? ` · <a class="k-link" href="${esc(s.ref)}" target="_blank" rel="noopener noreferrer">店家的等級說明<span class="k-sr">（另開視窗）</span></a>`
      : "";
    const rows = (s.grades || []).map((g) => `<div><dt><span class="k-grade">${esc(g.label)}</span></dt><dd>${esc(g.desc_zh)}${
      g.desc_ja ? `<span class="k-method__ja" lang="ja">「${esc(g.desc_ja)}」</span>` : ""
    }</dd></div>`).join("");
    return `<p>${esc(s.name)}（${esc(s.graded)}）${ref}</p><dl class="k-method__dl">${rows}</dl>`;
  }).join("");

  const badges = [
    [{ state: "below", pct: "N" }, `售價比行情低 ${band}% 以上；N 是實際低多少。`],
    [{ state: "low" }, `也低了 ${band}% 以上，但估價信心低，僅供參考。`],
    [{ state: "near" }, `和行情相差不到 ${band}%。`],
    [{ state: "above", pct: "N" }, `比行情高 ${band}% 以上。`],
    [{ state: "anomaly" }, `比行情低超過 ${anomaly}%。低這麼多，通常代表標題沒寫的瑕疵、缺件或版本差異，購買前請先看店家的照片與備註。`],
    [{ state: "none" }, `同型號可以比較的商品不到 ${n} 件，暫不估計。`],
  ].map(([v, text]) => `<div><dt>${badgeHTML(v)}</dt><dd>${text}</dd></div>`).join("");

  el.innerHTML = `<div class="k-method">
  <nav aria-label="位置" class="k-crumb"><a href="/" class="k-link">首頁</a><span aria-hidden="true">/</span><span aria-current="page">行情怎麼算</span></nav>
  <h1>行情怎麼算</h1>
  <p class="k-method__lead">Kura 上的數字分三種：店家刊登的資料、由這些資料算出來的行情與統計，以及為了展示設計而做的示範內容。這一頁說明每一種怎麼來。</p>

  <section class="k-dsec" id="source" aria-labelledby="source-h"><h2 id="source-h">店家資料</h2>
    <p>${source}</p>
    <p>售價以日圓刊登，Kura 用頁尾標示的匯率換算成新台幣，未含國際運費與進口稅費。Kura 不販售商品：每件商品都連到店家自己的商品頁，購買、付款與寄送都在店家完成。</p>
    ${ladders}
  </section>

  <section class="k-dsec" id="fair" aria-labelledby="fair-h"><h2 id="fair-h">行情估價</h2>
    <p>行情是同品牌、同型號（有型號編號時以編號為準）的在架與近期下架商品售價的中位數。比較時優先找同一家店、同一個等級的商品：同店同等級有 ${n} 件以上，就取它們的中位數（同店同等級中位數）。</p>
    <p>不到 ${n} 件時，Kura 把同型號各等級的價格換算到這件的等級，再取中位數（跨等級換算）。這個換算只是估價用的參考，不會取代店家的等級：頁面上的等級一律是店家自己的等級與說明。同型號可以比較的商品不到 ${n} 件時，不估計行情。</p>
    <p>信心表示這個估計有多可靠。同店同等級的樣本越多、價格越集中，信心越高；跨等級換算的估計最高是「中」。價格很分散（可能混有不同皮革、尺寸或限定版）時信心會調降；沒有型號編號、只能用名稱分組的估計一律是「低」，僅供參考。</p>
    <p>商品頁與型號頁的「同型號價格分布」，畫的是同型號在售與已售出商品依店家等級分組的售價，已售出的商品標在 Kura 最後記錄到的售價。中位數與表格只算在售商品，和顧問引用的是同一組數字。</p>
  </section>

  <section class="k-dsec" id="badges" aria-labelledby="badges-h"><h2 id="badges-h">標籤怎麼看</h2>
    <dl class="k-method__dl">${badges}</dl>
    <p>「價格最美」收錄比行情低 ${deal}% 以上、估價信心不是「低」、也不是異常低價的在售商品，依低於行情的幅度排序；信心「中」的估計排序時會打折。${
      band < deal ? `所以標示「低於行情」但低不到 ${deal}% 的商品，不會出現在價格最美。` : ""
    }</p>
  </section>

  <section class="k-dsec" id="history" aria-labelledby="history-h"><h2 id="history-h">需要時間累積的數字</h2>
    <p>已售出件數、找到新主人的天數與售出率，要看商品從上架到售出的過程，需要累積記錄。Kura 的記錄滿 ${days} 天之前，這些數字一律顯示「資料不足」：不是 0，也不代表賣不掉。${
      ov.overview?.sold_stats_ready ? `目前的記錄已滿 ${days} 天，這些數字已經開始計算。` : `目前的記錄還不到 ${days} 天。`
    }</p>
    <p>商品從店家的頁面下架時，Kura 把它標為「已找到新主人」；店家通常不公開成交，這是 Kura 的推定。找到新主人的天數只計算 Kura 看著它上架的商品：開始記錄時已經在架上的商品，不知道真正的上架日，不列入。型號頁的「平均找到新主人」顯示的是這些天數的中位數，不會被少數放了很久的商品拉高。</p>
    <p>商品頁的「售價變化」從 Kura 第一次看到這件商品時的售價畫起；之後查看店家時發現調價，才記下新的售價，價格沒變的查看不另外記錄。線畫到今天；已找到新主人的商品，畫到 Kura 把它標為已找到新主人的那天。</p>
  </section>

  <section class="k-dsec k-dsec--last" id="demo" aria-labelledby="demo-h"><h2 id="demo-h">哪些是示範</h2>
    <p>Kura 還沒有真的會員服務。以下內容是為了展示設計而做的示範，不是真實使用者的資料：</p>
    <ul>
      <li><strong>登入與示範會員</strong>：沒有帳號系統。按「登入」會直接以示範會員進入，狀態只保存在目前的瀏覽器分頁，新開的分頁會回到訪客。</li>
      <li><strong>追蹤清單</strong>：示範會員追蹤的型號與商品，是 Kura 從店家的商品裡挑出來的，可能包含已找到新主人的商品；你按的「追蹤」也只保存在這個分頁。</li>
      <li><strong>通知</strong>：誰在追蹤是示範的，但每則通知說的事和數字，都和連到的商品頁一致：新上架是這個型號在店家最新的一件，降價與已找到新主人只在 Kura 真的記錄到時出現。通知不會寄出。</li>
      <li><strong>大家都在找</strong>：需要真實使用者的瀏覽與追蹤資料。Kura 還沒有這些資料，所以目前不顯示。</li>
    </ul>
    <p>除此之外，頁面上的數字都來自店家資料，或由它算出。顧問回答裡的價格、行情與件數，和頁面用的是同一套計算；顧問標示「示範模式」時，回答是依規則搜尋的結果，沒有經過 AI 分析。</p>
  </section>
</div>`;
}

// 關於 Kura, 常見問題, 聯絡我們, 隱私權政策, 服務條款: their content — what Kura
// is, contact details, legal terms — is the owner's to write, and invented text
// here would read as real. Each slug still opens a page because the footer
// promises one; the not-found view would read as a broken link. The heading is
// the footer link's own text, so the canvas label lives once, in index.html,
// rather than in a second copy here.
export function renderPending(el, ov, slug) {
  const label = document.querySelector(`.k-footer nav a[href="/${slug}"]`)?.textContent.trim() || "Kura 蔵";
  el.innerHTML = `<section class="k-empty k-pending">
    <h1>${esc(label)}</h1>
    <p>這一頁的內容還在準備中。想知道 Kura 的數字怎麼來，可以先看<a href="/methodology" class="k-link">行情怎麼算</a>。</p>
  </section>`;
  return `${label} · Kura 蔵`;
}
