# Audit: Kura footer IA + 行情怎麼算

> **Status:** Complete (Accepted)
> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 7
> **Author:** Claude Opus 5.5 (auditor, sprint subagent)
> **Parent:** 2610131121-kura-methodology-footer-patch.md
> **Related Projex:** 2610131104-kura-methodology-footer-plan.md (audited plan, S1–S8, A1–A12) | 2610131121-kura-methodology-footer-patch.md (execution claims) | 2610130449-kura-mock-def.md (Mocked set; § Canvas gaps revision in `ca661cd`; L2 F19, L4 grade mapping / freshness open) | 2610130524-kura-interactive-mock-sprint.md | 2610131143-kura-methodology-copy-accuracy-patch.md (fixes F1–F4, F7, F8)
> **Re-audit:** 2026-10-13 of fix commit `ef3d531` + doc `abc50ee` (range `78c45bd..abc50ee`) → § Re-audit; conditions F1–F4 met, F7 + F8 fixed
> **Audit Date:** 2026-10-13 | **Work Period:** 2026-10-13, commits `c19a8a7` (code) `ca661cd` (patch doc + plan close + def revision) on `projex/sprint/2610130524-kura-interactive-mock` (range `436aba6..ca661cd`)
> **Subject:** plan S1–S8 vs code, tests and the live app (worktree `kura.db` served on 127.0.0.1:18737 by a scratchpad build, no crawl flags, stopped by PID; read-only DB copy in scratchpad for SQL probes, deleted after); every factual sentence on `/methodology` checked against the code path it describes; canvas `project/Main.dc.html` re-read (served 1791845608-ba84) for footer / 行情怎麼算 copy

---

## Audit Summary

**Claim:** `/methodology` (行情怎麼算) explains Kura's numbers with every threshold read from one source; five other footer slugs open an honest pending view; detail 行情估價 link live; detail prose + advisor prompt stop repeating thresholds as literals; nothing open decided.

**Verdict:** Partial — mechanics, routing, single-sourcing, tests and look all verified; four sentences on `/methodology` are not true of the code (three hold on today's DB only because it has no gone listings / no price change yet, one misattributes the Chinese grade text).

**Assessment:** Completeness: High | Correctness: Medium | Quality: High | Value: High

**Top Issues:**
1. 「同型號價格分布」 sentence says the chart plots in-sale listings — it plots sold listings too (only the median is in-sale)
2. 追蹤清單 sentence says persona picks come from listings on sale — `member` also picks gone listings
3. 售價變化 sentence says a price is recorded at every visit — only first sighting + each change are recorded; and `#source` presents Kura's Chinese grade translation as shop-published content

---

## Claims vs Evidence

| Claim | Evidence | Status | Notes |
|-------|----------|--------|-------|
| S1 routes + titles, SPA fallback, Back/Forward, `/nope` | Real mouse clicks from Home on all six footer links (window marker kept = no reload); direct loads; history nav | ✓ | Titles `{label} · Kura 蔵` / `行情怎麼算 · Kura 蔵`; `/nope` + `/methodology/` → not-found (same as other routes) |
| S2 detail link | Real click Home → card カルトシエールGM → 行情估價 「行情怎麼算」 | ✓ | In-app, scrollY 0, title set |
| S3 content structure | `get_page_text`; DOM | ✓ | crumb, h1, lead, 5 sections w/ ids; BRAND OFF ladder 5 grades with 中文 + 「日文」; 6 badges via `badgeHTML` |
| S4 disclosure = def mocked set | `#demo` vs def § Boundaries Mocked | ✓ | sign-in/member, watchlist, 通知, 大家都在找; no new label elsewhere (grep) |
| S5 one source per threshold | `TestOverviewMethod`, `TestPromptStatesThresholds` pass; `/api/overview` `method` = `{0.05, 0.5, 3, 14}`; grep | ✓ | Page reads 3 / 5 / 50 / 14 and 3 件; no literal in `pages.js` / `detail.js` |
| S6 not stated | page text; grep `rank|更新|最後` | ✓ | no rank codes / data age / legal; constants unchanged |
| S7 look | 1280 + 375 × light/dark DOM metrics + screenshots that rendered | ✓ | sw 1265 ≤ 1280, 375 = 375; badges single-line (dt 216 px, widest 170 px); 375 dl 343 px single column, h1 26 px; targets 24 px; `:focus-visible` 2px solid; 0 console errors |
| S8 build | `go vet ./...` clean; `go test -count=1 ./...` ok | ✓ | |
| Step 4.4 "content audit: each sentence's claim matches the code path" | Sentence-by-sentence re-check (§ Content Audit) | ✗ | 4 sentences false of the code (F1–F4) |
| Patch "nothing remains open" / def revision claims only what was built | `ca661cd` def diff | ✓ | Canvas gaps line + Revision Log row; explicitly decides no L2–L4 question |

---

## Content Audit — `/methodology` sentence by sentence

Checked against code at `ca661cd`. ✓ = true of code and today's data; ✗ = false of code; ⚠ = true today, imprecise.

| Section | Sentence (abridged) | Code path | Status |
|---|---|---|---|
| lead | numbers are shop data / computed / demo | — | ✓ |
| `#source` | 資料來自 BRAND OFF | `listing.Schemes()` → one scheme | ✓ |
| `#source` | 商品、售價、照片、成色等級與**等級說明**，都照店家刊登的內容呈現 | `listing.Grade.DescZH` = "Traditional Chinese translation of DescJA" (`brandoff` scheme) — only the 「」 Japanese is the shop's text | ✗ **F4** |
| `#source` | Kura 自 2026/10/13 起記錄 | `Overview.Since` = epoch 2026-10-13T00:25+08 | ✓ |
| `#source` | 頁尾匯率換算，未含運費稅費 | `fx.Converter.Current` same rate as footer `#fx` | ✓ |
| `#source` | Kura 不販售；連到店家商品頁 | detail CTA | ✓ |
| `#fair` | 同品牌同型號（編號優先）在架與近期下架售價的中位數 | `value`: group = brand + `GroupKey`; gone within `compWindow` 180 d | ✓ |
| `#fair` | 同店同等級 ≥ N → 中位數 | `same` = same source + grade code, `MinComps` | ✓ |
| `#fair` | < N → 各等級換算到這件的等級再取中位數；< N 件不估計 | `adjusted` / `rank.Factor()`; `default: return nil` | ✓ (also no estimate for an unmappable grade — see F5) |
| `#fair` | 信心：樣本多、集中 → 高；跨等級最高「中」；分散調降；名稱分組一律「低」 | level 1/2 at 3/6; rank-adjusted ≤ 1; dispersion > 0.4 / 0.6; `name:` → 0 | ✓ |
| `#fair` | 「同型號價格分布」畫的是同型號**在架**商品依等級分組的售價；中位數、表格與顧問同一組數字 | `dist.js` plots every item it gets with `k-dot--sold` + legend 已售出; `model.js` passes all `items`; `/api/models` returns gone items (`Radar.Model`); medians/table = `by_grade` (active only), advisor `model_price_stats` reads `ByGrade` | ✗ **F1** (true today only: 0 gone rows) |
| `#badges` | below / low ≥ band; near < band; above ≥ band; anomaly > AnomalyPct | `card.js` `valuation` | ✓ |
| `#badges` | none = 可比較商品不到 N 件 | `!f` — also no group key or rank `''` | ⚠ F5 |
| `#badges` | 價格最美 = ≥ DealMinPct, 非低信心, 非異常, 在售；依幅度排序，「中」打折 | `IsDeal`; `deals=1&sort=deal`; `dealScore` × 0.7 | ✓ |
| `#history` | 已售出/天數/售出率 滿 N 天前「資料不足」；目前不到 14 天 | `SoldStatsReady` (latest − epoch); `model.js` NA; advisor `insufficient_history` | ✓ (售出率 is advisor-only; no UI tile) |
| `#history` | 下架 → 「已找到新主人」，Kura 的推定 | `SweepMissing` / `MarkGone` → `gone`; card/detail/notify label | ✓ |
| `#history` | 天數只計 Kura 看著上架的商品；型號頁「平均」= 中位數 | `watched` (epoch + 24 h); `median(days)` | ✓ |
| `#history` | 「售價變化」是 Kura **每次查看店家時**記下的售價 | `store.RecordSeen` inserts a `price_points` row only when new or price changed; `historyHTML` steps to today | ✗ **F3** |
| `#demo` | 登入直接進示範會員；狀態只在此分頁 | `member.js` sessionStorage | ✓ |
| `#demo` | 追蹤清單是 Kura **從目前在架的商品**裡挑出來的 | `Server.member` listing candidates = `Radar.Drops` + **`Radar.Gone`** + top deals | ✗ **F2** (true today only: 0 gone rows) |
| `#demo` | 通知：誰在追蹤示範；事與數字一致；新上架 = 最新一件；降價/新主人只在記錄到時 | `member`: `NewestActive`, `Drops`, `Gone`, `bare(Item)` | ✓ |
| `#demo` | 大家都在找 不顯示 | `home.js`, `model.js` (「這週 N 人在找」 not rendered) | ✓ |
| `#demo` | 顧問與頁面同一套計算；示範模式 = 規則式搜尋，未經 AI | advisor tools → radar; `mock.go` banner | ✓ (but canvas-copy header tooltip disagrees — F6) |

---

## Objective Verification

### Step 1: constants, overview `method` + `schemes`, advisor prompt
**Evidence:** `c19a8a7` diff; `TestOverviewMethod`, `TestPromptStatesThresholds`, `TestIsDeal` pass; live `/api/overview`.
**Findings:** `MinComps`/`DealMinPct`/`AnomalyPct` exported, values unchanged; `IsDeal`/`value` use them; `Schemes()` sorted by source; prompt `{{anomaly_pct}}` replaced, `only_deals` built from `pctText(radar.DealMinPct)`. Quality: High. **Verification:** ✓

### Step 2: `/methodology` + pending pages; router; styles
**Evidence:** `pages.js`, `app.js`, `card.js`, `kura.css`; browser.
**Findings:** Structure, interpolation and styling as planned; copy is the plan's verbatim — F1–F4 are plan-authored copy errors the executor's Step 4.4 content audit passed. Quality: Medium (copy). **Verification:** ⚠ Partial

### Step 3: detail prose from `ov.method`
**Evidence:** `detail.js` diff; live detail (16-comp same-grade listing).
**Findings:** 「不足 ${m.min_comps} 件」, 「不到 ${m.min_comps} 件」, 「低超過 ${pct(m.anomaly_pct)}%」; no literal left. **Verification:** ✓

### Step 4: end-to-end verification
**Findings:** Matrix re-run with real clicks (patch used `.click()`/router calls for S2) — all mechanical claims hold. Step 4.4 sentence audit claim ✗ (F1–F4). **Verification:** ⚠ Partial

---

## Code/Implementation Inspection

### `pages.js` (new)
**Claimed:** page + pending view, no threshold literal. **Actual:** as claimed; `esc()` on every interpolated string from data; `rank` never rendered; `since` guard `Date.parse(...) > 0` handles Go zero time (`0001-01-01` → negative). Quality: High.
**Issues Found:** copy F1–F4 — Severity: Medium each. **Undocumented:** none.

### `card.js` / `detail.js` / `app.js` / `kura.css`
**Actual:** `NEAR_BAND` exported, both comparisons use it; routes appended; CSS three extra declarations disclosed in patch. Quality: High.

### `radar.go` / `condition.go` / `server.go` / `prompt.go`
**Actual:** rename-only for values; anomaly rationale comment moved to the constant. `Schemes()` takes the registry read lock. Quality: High.
**Metrics:** Readability High | Tech debt Low (two formatters for one fraction→percent: Go `pctText`, JS `pct` — both `round(f*100)`, consistent).

---

## Testing Validation

**Execution:** `go vet ./...` clean; `go test -count=1 ./...` → ok advisor | radar | source/brandoff | store | web.
**Coverage:** `method` values, scheme membership, prompt threshold text, static routes for `/methodology` `/about` `/js/pages.js`.
**Missing:** `Schemes()` source ordering untested (`internal/listing` has no tests; membership-only assertion by design since the registry is process-wide) — Low. No automated check that page copy matches behaviour (inherent; Risks § Copy accuracy drift in plan).

---

## Documentation Audit

**Patch doc accuracy:** Mechanical claims (files, line ranges, test names, live values, S1–S3/S5–S8 evidence) match. Overstated: S3/S4/Step 4.4 imply sentence-level accuracy; F1–F4 contradict. S2 evidence was programmatic — now confirmed by real clicks.
**Def revision (`ca661cd`):** § Canvas gaps Pages line states 行情怎麼算 built (provisional) + five pending pages; Revision Log row says no L2–L4 question decided — accurate; it marks nothing resolved and adds no Mocked item. ✓
**Plan close:** all steps `[PATCHED]`, Status Complete, criteria ticked — ticks for S3/S4 inherit the content gap.

---

## Source Hygiene

**Scope:** Diff only (`git show c19a8a7 -- . ':(exclude,glob)**/.projex/**'`)

| Rule | Location (symbol) | Quote | Severity |
|------|-------------------|-------|----------|
| — | — | no rule 1–5 violation (grep `projex|[0-9]{10}-|// Step|====|:NN` on added lines → none) | — |

**Rule 6 gaps:**
- `pages.js` → `renderMethodology` `since` guard — `Date.parse(...) > 0` rejects Go's zero time (`0001-01-01T00:00:00Z`, empty DB); not obvious why `> 0` rather than truthiness — Low.

**Rule 6 vacuity:** none — `radar` const block, `card.js` `NEAR_BAND`, `pages.js` headers, `renderPending` each name the rejected alternative (typed copy goes stale; server constant needs overview in every card; not-found reads broken; JS label map = second copy).

**Rationale promotion:** none needed — the load-bearing rationale (why constants are single-sourced, why band is client-owned, why pending pages exist) lives in the code comments, not only in projex.

**Commit composition** (informational): 2 landed commits — `c19a8a7` `feat(web): …` typed, type matches diff, `Projex:` trailer ✓; `ca661cd` doc-only, trailer ✓. **Violations:** None.

**Trailer survival on `master`:** 0/6 — N/A, not Critical: `master` holds only the baseline import + pre-sprint commits (sprint unmerged). Sprint branch: every code commit carries `Projex:`.

---

## Gap Analysis

### Promised But Not Delivered
| Promise | Status | Impact |
|---------|--------|--------|
| Every `/methodology` sentence true of the code (plan Rationale, Step 4.4) | Partial (F1–F4) | Medium — the page is the def's disclosure channel; a reviewer who sees a sold dot or a gone tracked item next to the page's claim reads it as inaccurate |

### Undocumented Issues
| Issue | Severity | Affects |
|-------|----------|---------|
| F1–F3 become false at the first crawl that marks a listing gone or changes a price (hosted demo crawls hourly per def L1) | Medium | `/methodology` credibility after first crawl |
| Advisor 示範模式 tooltip (`index.html` `.k-adv__demo` title, canvas copy) says 「顧問以範例資料回答」; the page (and `mock.go`) say rule-based search over real data | Low | advisor disclosure consistency (F6, human/canvas) |

### Unhandled Edge Cases
- Listing with unmappable grade (rank `''`, 1 row today) or no group key → 「尚無行情估價」 legend / detail say 「不到 N 件」 — Impact: wording imprecise for those rows; describing unmappable grades is def L4 open.
- `/methodology/` (trailing slash) → not-found — consistent with all other routes; none.

---

## Quality Assessment

### Completeness: High
**Strengths:** all six footer links + detail link reachable; single-source thresholds on server, client and advisor; tests pin the API to the constants. **Gaps:** copy accuracy (F1–F4).

### Correctness: Medium
**Works:** routing, titles, history, SPA fallback, interpolated values, badges, a11y, theming: Yes. **Bugs:** F1–F4 copy — Medium.

### Code Quality: High
**Positive:** minimal, rename-only server change; escaped output; labels read from `index.html` once. **Tech Debt:** none new.

### Value Delivered: High
**Intended:** last unreachable canvas surfaces + the def's disclosure. **Actual:** delivered; disclosure needs four sentence fixes to be fully true. **Impact:** User: Positive.

---

## Open Findings

### Undocumented Discoveries
- Changes: none beyond patch doc.
- Problems: F1–F3 latent on today's DB (status: 15367 active, 0 gone; every listing has exactly one price point) — evidence: SQL on DB copy.
- Workarounds: none.

### Impact Analysis
- Downstream: iteration 8 goal walk — every canvas link now reachable ✓.
- Future: Enabled: an F1/F6 constant answer flows to page, detail, advisor with no copy edit. Blocked: none.
- Risks: qualitative sentences drift from code — Likelihood: Medium (already happened once, F1–F3).

### Improvements
- Could be better: sentences that depend on data state (gone rows, price changes) phrased for the general case, not the current DB.
- Missed: Step 4.4 checked sentences against plan § Context, which itself omitted sold dots, gone picks and change-only price points.

---

## Stakeholder Impact

| Stakeholder | Promised | Reality | Impact |
|-------------|----------|---------|--------|
| Reviewer (job portfolio) | honest methodology + disclosure page | page reachable, numbers live; 4 sentences inaccurate in general case | Positive (Neutral until F1–F4 fixed) |
| Owner | no invented company/legal facts; open questions untouched | pending pages say only 「還在準備中」; def/L2–L4 untouched | Positive |

---

## Findings

### Critical (Must Address)
- None.

### Significant (Should Address) — fixable now (copy in `pages.js` `renderMethodology`; no decision needed)
- **F1 「同型號價格分布」 claims in-sale only** — chart plots sold listings as `k-dot--sold` (legend 已售出); only medians are in-sale. False once any listing goes. → e.g. 「畫的是同型號在售與已售出商品依店家等級分組的售價；中位數只算在售商品，中位數、表格與顧問引用的是同一組數字。」
- **F2 追蹤清單 「從目前在架的商品裡挑出來」** — `Server.member` also tracks `Radar.Gone` listings (to surface 已找到新主人). False once any listing goes. → e.g. 「是 Kura 從店家的商品裡挑出來的，可能包含已找到新主人的商品」.
- **F3 售價變化 「每次查看店家時記下的售價」** — `store.RecordSeen` stores the first-seen price and each change only. → e.g. 「是 Kura 第一次看到這件時的售價，以及之後查看店家時記到的每次調價」.
- **F4 等級說明 「照店家刊登的內容呈現」** — the Chinese text (`DescZH`) is Kura's translation; only 「」 is the shop's original. → add 「中文說明是 Kura 的翻譯，「」內是店家的原文。」 (or exclude 等級說明 from the "as published" list and say so beside the ladder).

### Minor (Nice to Fix)
- **F7 Rule-6 gap** — `renderMethodology` `since` guard lacks the Go-zero-time rationale → one-line comment.
- **F8 `Schemes()` order untested** → optional `internal/listing` test (register two sources, assert order among them).

### Needs Human Decision (not fixable by the agent)
- **F5 「尚無行情估價」 meaning** — legend + detail say 「可比較商品不到 N 件」 but the state also covers listings with no model group or an unmappable grade; how to state unmappable grades is def L4 (grade mapping disclosure / grades with no mapping) — owner.
- **Anomaly boundary display** — brandoff:2101220216016 reads 「低 50%」 without anomaly (DealPct ≤ 0.5 rounds to 50) while the legend says anomaly is 「低超過 50%」: consistent with strict `>`, visually identical at the boundary → F6 (anomaly-boundary ordering) — owner.
- **F6 示範模式 tooltip vs page** — `index.html` `.k-adv__demo` title 「未設定 API 金鑰，顧問以範例資料回答」 says sample data; the page and `mock.go` say rule-based search over real data. The tooltip is canvas copy verbatim (`project/Main.dc.html` advisor header, served 1791845608-ba84; landed in `912a354`), and canvas governs copy → canvas pass 2 / owner: reword the tooltip, or log it as a canvas deviation. Pre-existing; iteration 7 made the conflict visible.
- **F19** — whether unlabelled mocks are acceptable with this page as the only disclosure stays open (def L2); the page does not settle it.

### Positive
- Real-click re-verification from Home matches every mechanical claim in the patch; no reload on any in-app hop.
- Thresholds now have exactly one definition each across page, detail prose, card band and advisor prompt; tests pin the API to the constants.
- Pending pages invent nothing and keep canvas labels in one place.

---

## Recommendations

**Immediate:** patch F1–F4 (four sentences in `pages.js`); optionally F7 comment, F8 test.
**Future:** F5 wording once L4 decides how unmappable grades are disclosed; F6 tooltip at canvas pass 2.
**Process:** content audits of explanatory copy should check each sentence against the code path (store, member, chart), not against the plan's Context summary, and in the general data state (gone rows, price changes), not only today's DB.

---

## Final Verdict

**Status:** Accept with Conditions

**Overall Assessment:**
- Completeness: High
- Correctness: Medium
- Quality: High
- Value: High

**Conditions:**
- [x] F1 — 同型號價格分布 sentence covers sold dots / in-sale median — 2610131143-kura-methodology-copy-accuracy-patch.md
- [x] F2 — 追蹤清單 sentence no longer says picks are on sale only — 2610131143-kura-methodology-copy-accuracy-patch.md
- [x] F3 — 售價變化 sentence states first price + changes — 2610131143-kura-methodology-copy-accuracy-patch.md
- [x] F4 — Chinese grade text disclosed as Kura's translation — 2610131143-kura-methodology-copy-accuracy-patch.md (also F7 comment, F8 test)

**Sign-off:** Yes, conditionally — routing, single-sourcing, tests and look verified; the page's purpose is accuracy, so the four false sentences should be corrected before iteration 8's goal walk.

**Re-audit verdict:** Accept — conditions F1–F4 earned (verified below, not taken from the patch doc's ticks); F7, F8 fixed. Open items are human decisions only.

---

## Re-audit

> **Scope:** `ef3d531` (`pages.js` copy + `since` comment; new `internal/listing/condition_test.go`) and `abc50ee` (patch doc 2610131143-kura-methodology-copy-accuracy-patch.md; ticks on this audit's conditions). **Method:** each corrected sentence re-checked against its code path; scratchpad build served a **mutated scratch copy** of `kura.db` (outside the repo, deleted after; 127.0.0.1:18738, no crawl flags, stopped by PID) with one listing marked gone (brandoff:2101218767803, same model M51252) and one price cut recorded (brandoff:2101220216016, ¥50,000 → ¥45,000, new `price_points` row); real click footer → `/methodology` (window marker kept); `go vet` + `go test -count=1 ./...`; mutation tests of F8.

| Finding | Corrected sentence (abridged) | Evidence | Status |
|---|---|---|---|
| F1 | 分布畫在售與已售出商品…已售出標在 Kura 最後記錄到的售價；中位數與表格只算在售，和顧問同一組 | Mutated DB: detail dist shows one `k-dot--sold` 「已售出 NT$14,129」 (= ¥70,000 last stored `price_jpy` × rate); table ランクA 在售 8 excludes the gone row; legend 中位數（在售）; advisor `model_price_stats` ← `ByGrade` | ✓ |
| F2 | 追蹤清單…從店家的商品裡挑出來的，可能包含已找到新主人的商品 | Mutated DB `/api/member`: watchlist includes gone brandoff:2101218767803 with a `gone` event | ✓ |
| F3 | 售價變化從第一次看到時的售價畫起；發現調價才記；沒變不記；畫到今天／已找到新主人那天 | `store.RecordSeen` (insert on new or changed price only); mutated DB: detail history NT$10,092 → NT$9,083 ending 「今天」; gone listing's line ends at its gone date | ✓ |
| F4 | 商品、售價、照片與成色等級照店家刊登；等級說明的中文是 Kura 的翻譯，「」內是店家原文 | `listing.Grade.DescZH` doc; ladder rows render 「desc_ja」 | ✓ |
| F7 | `since` guard comment | names the rejected truthiness check and its consequence (「自 0001/01/01 起記錄」), and NaN for a missing field | ✓ |
| F8 | `TestSchemesSortedBySource` | passes as committed (×3); comparator reversed → FAIL; sort removed → FAIL 10/10 runs (file restored by byte copy, `cmp` identical) | ✓ |

**Regressions:** none — `/methodology` 5 sections, 6 badges, title, no h-scroll at 375 (sw 375); `/faq` pending + `/nope` not-found unchanged; detail fair prose interpolates (mutated listing now anomaly → 「低超過 50%」); 0 console errors; vet clean; all packages ok (now incl. `internal/listing`).

**Source hygiene (`ef3d531`):** no rule 1–5 hit on added lines; F8 test comment states the trap (map iteration order passing by luck) — rule 6 satisfied. Commits: `ef3d531` `fix(web): …` typed, matches diff, `Projex:` ✓; `abc50ee` doc-only, `Projex:` ✓.

**Adjacent — dist table 「店家定義」 column** (flagged by the patch agent): header is canvas copy verbatim (`project/Main.dc.html` table head, served 1791845608-ba84; canvas data carries `gradeZh`), content is `desc_zh` — the shop's definition in Kura's translation. **Minor, canvas copy, not a condition:** authorship of the definition is the shop's, the language is Kura's, and `/methodology` `#source` now discloses that translation once for every surface (card, detail tag, dist table, advisor, ladder). Same reading covers `#fair` 「頁面上的等級一律是店家自己的等級與說明」. Rewording the header (e.g. 「店家定義（中譯）」) is a canvas pass 2 copy call → owner.

**Still open (human decisions, unchanged):** F5 「尚無行情估價」 meaning for no-group / unmappable-grade rows (def L4) | F6 示範模式 tooltip 「以範例資料回答」 (canvas copy) vs rule-based-over-real-data behaviour | anomaly boundary display (「低 50%」 non-anomaly vs legend 「低超過 50%」, sprint F6) | F19 unlabelled mocks (def L2) | 「店家定義」 header wording (canvas pass 2).
