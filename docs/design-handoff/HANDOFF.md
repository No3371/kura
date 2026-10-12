# Kura 蔵: design handoff

**For:** Claude Design
**Ask:** Redesign Kura's web UI. The current build works but looks like a generic dashboard (see `current-*.jpg`). It should feel like a small, trusted luxury consignment house: warm, human and quietly elegant. The goods are used but fine, cared for, and waiting for their next owner. It must not look like a discount marketplace or a crypto-style price tracker.

---

## 1. What Kura is

Kura (蔵, "storehouse") tracks pre-owned luxury handbags for sale at Japanese resale shops and helps Taiwanese buyers decide what is worth buying.

- **行情雷達 (price radar).** Every listing gets a fair-value estimate from comparable listings (same model, same shop grade). It shows whether the price sits below market, how sure Kura is, price drops, and how fast a model sells.
- **AI 找包顧問 (AI bag advisor).** A chat where Claude searches the real listings and recommends specific bags, explaining its reasoning. Replies stream in, and the steps it takes are shown as it searches.
- **資料管線 (pipeline).** A transparent page explaining where the data comes from and how values are computed, plus a crawl log.

**Audience.** Taiwanese buyers, 25–45, who know the brands, are price-literate and are wary of fakes and hidden damage. They want reassurance and expertise, not hype.

**Language.** The UI is Traditional Chinese (zh-TW). Japanese appears inline: model names, the shop's grade labels and the shop's original wording. Brand names are in Latin capitals.

## 2. The feeling

| We want | We avoid |
|---|---|
| A consignment boutique with an appraiser who tells you the truth | A SaaS analytics dashboard, KPI tiles everywhere |
| Warmth of paper, linen, ink, leather, handwriting | Black-and-gold bling, glossy gradients, "luxury" clichés |
| Wear treated with respect: 「用過，但被好好愛過」 ("used, but well loved") | Red SALE stickers, countdowns, urgency |
| Restraint, generous whitespace, editorial rhythm | Dense grids of identical cards with no hierarchy |
| Honesty as a design value: uncertainty shown gracefully | Overclaiming ("BEST DEAL!!"), fake precision |

**Human touch ideas** (pick and refine; don't use them all):
- **The appraisal tag (鑑定卡).** Present a listing's condition like the paper tag a shop ties to an item: the shop's grade, its definition in the shop's words, and who graded it. The tag might use a handwritten or stamped accent such as a hanko-style seal, but the type stays legible.
- **Margin notes.** Kura's commentary ("價格比同款同等級低 31%" / "31% below the same model at the same grade") reads like a pencilled note from a knowledgeable friend, not a system badge.
- **"Found a new home."** A sold listing reads 已找到新主人 ("found a new owner"), not just 已售出 ("sold"). It's gentle and a little sentimental, which fits pre-loved goods.
- **Wabi-sabi, lightly.** A subtle paper grain or deckle edge on one or two surfaces at most. Never on data or charts.
- **Hand-set numerals** for prices: a refined serif with old-style or tabular figures as appropriate.

**Type direction.** An elegant serif for display and prices, with CJK support (e.g. Noto Serif TC or Shippori Mincho for Japanese). Pair it with a calm sans for UI and dense data. Charts use the sans.

**Colour direction.** Warm off-white paper, ink near-black, and one deep accent (the current oxblood `#6b1f2a` is a fine starting point). Add a muted secondary like aged brass or tea, used only as hairlines. Dark mode is a "night shop", warm charcoal rather than pure black, and is required.

## 3. Hard constraints

1. **Respect each shop's condition grading.** Each source keeps its own grading system. Never relabel a shop's grade into a Kura grade. Always show:
   - the shop's label exactly (e.g. `ランクA`),
   - its meaning in plain zh-TW text (e.g. 有少許刮痕、髒污與輕微使用感),
   - the shop's original wording, smaller (e.g. 「多少のキズ・汚れ・若干の使用感等あり」),
   - who graded it (`BRAND OFF 商品ランク，店家鑑定`).
   Future sources may be self-graded marketplaces (賣家自評), and that difference must be visible. Design the condition component for both 店家鑑定 (graded by the shop) and 賣家自評 (graded by the seller).
2. **Honest valuation states.** Every fair-value display has one of these states (copy in §5):
   - *Below market, high or medium confidence*: the only state allowed to feel positive.
   - *Possibly below market, low confidence*: neutral, with an explanation.
   - *Abnormally low price (anomaly)*: a gentle caution, never celebrated.
   - *Near market* / *above market*: quiet.
   - *Not enough data*: quiet.
   Colour must never be the only signal; pair it with words or an icon.
3. **Source attribution.** Every listing links to the shop's product page and shows the shop name. Use no shop logos and no branding that suggests Kura is the shop.
4. **Product photos** are the shop's white-background studio shots in a square, of uneven quality. The design must make white-background photos look intentional (a mat or frame treatment) in both light and dark mode.
5. **Responsive.** Desktop is about 1280 px. Mobile is 375 px with a 16 px gutter and no horizontal scroll, except deliberate carousels.
6. **Implementation reality.** The front end is plain HTML, CSS and JS embedded in a Go binary, with no build step and no framework. Google Fonts is fine. Prefer CSS tokens (custom properties) and simple components. Charts are hand-drawn inline SVG.
7. **Accessibility.** WCAG AA text contrast, visible focus, and hover targets of at least 24 px. Charts have a table view.

## 4. Screens and components to design

Please produce desktop and mobile artboards, each in light and dark.

### A. 行情雷達 (home)
- **Header:** the Kura 蔵 wordmark, three tabs, and a status line with the FX rate and its date (e.g. `1 JPY = 0.2018 TWD · 2026/10/12`) and the advisor mode. Mode is `Claude` or `示範模式` (demo mode); demo mode should look calm, not like an error.
- **Market summary.** Currently five KPI tiles; please rethink this. Data available: 15,367 在售 (for sale), 0 已售出 (sold), 9,130 有行情估價 (have an estimate), 1,020 低於行情 ≥10% (≥10% below market), 0 近期降價 (recent price drops), tracking since 2026/10/13, and per-brand counts (LOUIS VUITTON 6,030 · CHANEL 3,979 · HERMES 2,931 · CELINE 1,432 · LOEWE 897 · GOYARD 98). An editorial sentence might work better than tiles, e.g. 「今天追蹤 15,367 只包，其中 1,020 只低於行情。」 ("Tracking 15,367 bags today, 1,020 of them below market.")
- **Filters (one row):** brand, bag type, minimum condition (任何成色 / 未使用以上 / 輕微使用感以上 / 明顯使用感以上, i.e. any / unused or better / light wear or better / visible wear or better), budget in TWD, a keyword field (often Japanese, e.g. ネヴァーフル), sort (低於行情優先 / 最新上架 / 價格↑ / 價格↓: below market first, newest, price up, price down), and a 只看低於行情 ("below market only") toggle.
- **Listing card,** the most important component. Fields:
  - photo, brand (Latin), model name (Japanese), bag type (zh), model code,
  - condition: shop label plus the zh meaning,
  - price in TWD (large) and JPY (small), valuation state, price-drop state.
  States to show: default, below market (high confidence), low confidence, anomaly, price drop, sold, and photo pending (the shop's "NowPrinting" placeholder).
- **Side column:** 熱門型號 (popular models: name, brand, code, median TWD, counts) and 近期降價 (recent drops: from → to).
- **Empty state:** no data yet, with the next step being to run a crawl.

### B. Listing detail (currently a right-hand drawer; could be a page)
- Large photo, brand (Latin + zh), model, price (TWD / JPY), valuation state.
- **The appraisal tag** (condition component, §3.1).
- Specs: model code, bag type, colour, material, size, accessories (付屬品), shop notes (備註), and listed date with days on market.
- **Fair-value explanation:** estimate, sample count, method (same shop grade vs. converted across grades), confidence in words, and the saving in TWD. Include the anomaly and low-confidence wording.
- **售價變化 (price history):** a step line. Usually there's only one point so far, so design the "no change since first seen" state well.
- **同型號價格分布 (price distribution for the same model):** dots grouped by shop grade (one column per grade, best first) on a TWD axis. Filled dots are for sale, hollow rings are sold, the current item is highlighted, and a tick marks the median. There's a legend and a 以表格檢視 ("view as table") toggle that lists grade, definition, count, median, min and max.
- A link out to the shop's product page.

### C. Model view
Summary for a model (for sale, sold, median price, median days to sell, or 資料不足 when there isn't enough data), the distribution chart, and a grid of its listings.

### D. AI 找包顧問 (advisor chat)
- Opening message from the advisor and four suggested prompts.
- User bubble, then the advisor's reply in this order:
  1. **Search steps** while working (e.g. 「⋯ 搜尋在售商品」 searching for-sale listings, then 「✓ 找到 23 件，回傳 6 件」 found 23, returned 6). These should feel like an appraiser checking the stockroom, not a log.
  2. **Streamed text** with simple formatting (paragraphs, bullets, bold).
  3. Inline listing references `[[id]]`, rendered as small chips, and a horizontal strip of recommended cards.
- States: streaming (cursor), error (gentle), and demo mode.
- Composer with Enter to send and Shift+Enter for a new line.

### E. 資料管線 (pipeline)
Two explainer columns ("where the data comes from" / "how values are computed") and a crawl log table: start time, brand, pages, bags, new listings, price changes, marked sold, and status (✓ 完整 complete / 部分 partial / ⚠ error). Make the explanation feel like a shop's "our promise" page, with the table as an appendix.

### Component sheet
Tokens for colour (light and dark), type scale, spacing, radius and shadow. Components:
- buttons, selects and inputs,
- the condition tag (shop-graded and seller-graded variants),
- valuation badges (every state),
- card variants, the chat bubbles, the search-step line, a chart frame with legend and tooltip,
- the drawer or page shell.

## 5. Copy (zh-TW) and real sample data

**Valuation states** (the current wording; feel free to refine the tone):

| State | Badge | Detail line |
|---|---|---|
| Below market | ▼ 低於行情 31% | 行情估計 NT$151,378 · 高信心 · 11 件樣本 · 同店同等級中位數。售價比行情低 31%，約省 NT$46,423。 |
| Low confidence | 可能低於行情 47% · 低信心 | 同型號的價格非常分散（可能混有不同皮革、尺寸或限定版），這個估計僅供參考。 |
| Anomaly | ⚠ 異常低價 · 請先確認狀況 | 價格比同型號同成色的行情低一半以上，通常代表有標題沒寫的瑕疵、缺件或版本差異，購買前請務必確認商品頁的照片與備註。 |
| Near / above market | 接近行情 / 高於行情 12% | — |
| No estimate | — | 同型號的比較樣本不足 3 件，暫不估計行情。 |

**BRAND OFF grading** (the shop's words, then our translation):

| Label | Shop's definition | zh-TW |
|---|---|---|
| 新品 | 新品未使用品です | 全新未使用品 |
| 未使用品 | 未使用品にキズなどが少なく全体的に綺麗な状態 | 未使用，幾乎沒有刮痕，整體狀態漂亮 |
| ランクA | 多少のキズ・汚れ・若干の使用感等あり | 有少許刮痕、髒污與輕微使用感 |
| ランクB | 全体的にキズ・汚れ・使用感等あり | 整體有刮痕、髒污與使用感 |
| ランクC | かなりのキズ・汚れ・使用感等あり | 有相當多的刮痕、髒污與使用感 |

**Real listings for mockups** (photos are the shop's; use them for layout realism only):

| Case | Brand | Model / code | Type | Grade | Price | Fair value / state |
|---|---|---|---|---|---|---|
| High-confidence deal | HERMES 愛馬仕 | ピコタンロックPM · 056289CK | 手提包 | ランクA | NT$104,955 (¥520,000) | NT$151,378 · −31% · high · 11 comps |
| Chanel deal | CHANEL 香奈兒 | マトラッセ · A33814 | 肩背包 | ランクA | NT$48,441 (¥240,000) | NT$95,873 · −49% · high · 6 comps |
| Everyday classic | LOUIS VUITTON 路易威登 | ネヴァーフルMM · M40156 | 托特包 | ランクB | NT$26,239 (¥130,000) | NT$38,349 · −32% · high · 17 comps |
| Low confidence | HERMES 愛馬仕 | ガーデンパーティPM (no code) | 托特包 | ランクA | NT$48,441 (¥240,000) | NT$90,827 · −47% · low · 37 comps |
| Anomaly | LOUIS VUITTON 路易威登 | アルマPM モノグラム · M51130 | 手提包 | ランクB | NT$4,440 (¥22,000) | NT$16,147 · anomaly · 46 comps |

Photos, in the same order:
`https://www.brandoff-store.com/Contents/ProductImages/0/2101220285869_L.jpg` ·
`…/2120900081669_L.jpg` · `…/2101220126346_L.jpg` · `…/2101220247300_L.jpg` · `…/2118700085789_L.jpg`

**Advisor prompts:** 預算 8 萬，通勤可放 A4 的 LV 托特包 (budget NT$80k, an LV tote for commuting that fits A4) · 哪些愛馬仕包明顯低於行情？ (which Hermès bags are clearly below market?) · 香奈兒菱格紋鏈條包的行情 (market prices for Chanel quilted chain bags) · 5 萬以內的入門經典款 (entry-level classics under NT$50k).

**Standing footnote:** 價格未含國際運費與進口稅費 (prices exclude international shipping and import duties) · 真偽與商品狀況以賣家說明為準 (authenticity and condition are as described by the seller).

## 6. Data viz rules to keep

- Price distribution: one hue for "for sale" (filled) and the same or a second hue for "sold" (hollow ring). Shape carries the difference, not just colour. Use a 2 px surface ring on overlapping dots.
- Price history: a step line, 2 px wide.
- Use a single y-axis, quiet gridlines, a legend whenever there are two or more series, and a hover tooltip.
- Text in charts uses ink colours, never the series colour.

## 7. What to hand back

1. Artboards: A–E, desktop and mobile, light and dark.
2. A component sheet with tokens, written as CSS custom properties I can paste into `internal/web/static/style.css`.
3. Short notes on any copy changes and the reasoning behind the human-touch choices, so they can be implemented consistently.

**Files in this folder:** `current-radar.jpg`, `current-detail.jpg`, `current-chart.jpg`, `current-advisor.jpg`, `current-mobile.jpg`. These were captured before per-shop grades: their badges show `A 級` / `B 級`, where the build now shows `ランクA` / `ランクB` with the definition text.
