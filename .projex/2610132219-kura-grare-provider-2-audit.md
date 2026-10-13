# Audit: Kura provider #2 — GALLERY RARE end to end

> **Status:** Complete (Accepted with Conditions)
> **Sprint:** 2610131822-kura-def-coverage-sprint.md — iteration 3
> **Author:** Claude Opus 5.5 (auditor, sprint subagent)
> **Parent:** 2610132204-kura-grare-provider-2-patch.md
> **Related Projex:** 2610132140-kura-grare-provider-2-plan.md (audited plan, Steps 1–7) | 2610132204-kura-grare-provider-2-patch.md (execution claims) | 2610131822-kura-def-coverage-sprint.md (§ Current Position iteration 3, § Deferred; not edited) | 2610130449-kura-mock-def.md (§ Is display decision "Accept as demo") | 2610131935-kura-provider-2-vetting-audit.md (F1 README) | 2610132229-kura-grare-audit-fixes-patch.md (F1 small part, F2–F5, F8 fixed; F6, F7 accepted)
> **Audit Date:** 2026-10-13 | **Work Period:** 2026-10-13, commits `90813c4` (plan) `249fe3c` (code + docs) `5c180a6` (patch doc) on `projex/sprint/2610131822-kura-def-coverage` (range `90813c4..5c180a6`)
> **Subject:** `internal/source/grare` (scheme, parsers, sweeper, fixtures, tests) | `cmd/kura` `-sources` wiring | `/api/models` `schemes` + `dist.js` per-shop ladders | advisor `{{shops}}` | README + `docs/provider-vetting.md` | executor scratch DB (own copy) | own browser pass on HEAD build. No live requests to g-rare.com / brandoff-store.com / allu-official.com (see § Visual pass for one incidental image-hotlink disclosure).

---

## Audit Summary

**Claim:** GALLERY RARE onboarded as provider 2 under the restated display bar: own 9-rank `listing.Scheme` with verbatim JA definitions; fixture-tested adapter; `-sources` flag; `/api/models` returns one scheme per shop and the distribution draws one ladder per shop; advisor prompt names registered shops; README + vetting doc follow the def display decision; scratch crawl landed 200 rows, grades only from the scheme, no `?` URL.

**Verdict:** Partial — all 7 steps delivered and their checkable claims verified; one shipped claim (GALLERY RARE sold items surface through a product-page 404) is contradicted by the crawler's own detail selection and by the shop's saved product-page template.

**Assessment:** Completeness: High | Correctness: Medium | Quality: High | Value: High

**Top Issues:**
1. F1 — GALLERY RARE listings can effectively never be marked sold: sweeps never complete, each product page is fetched once, and sold product pages very likely stay 200 with an out-of-stock notice. README / `Sweep` doc state otherwise.
2. F2 — Fair-value copy hides that rank-bridged comps now come from the other shop (seen: `grare:11539001` "低於行情 35%" from 86 comps, 85 of them BRAND OFF).
3. F3 — Ladder caption repeats the shop name ("BRAND OFF BRAND OFF 商品ランク · 店家鑑定").

---

## Claims vs Evidence

| Claim | Evidence | Status | Notes |
|-------|----------|--------|-------|
| `go vet` + `go test` green | Re-run in worktree: vet exit 0; `go test -count=1 ./...` all ok; `-count=3` advisor/web ok | ✓ | |
| 9-rank scheme, DescJA verbatim from 商品ランクガイド | `grare.Scheme` vs saved plan-time probe `/f/condition` (session scratch, not refetched) | ✓ | All 9 labels + definitions exact (①② lines joined). Product pages carry a different, politer 商品ランク一覧 (「…中古品です。」, label 未使用品); scheme correctly quotes the guide |
| 中古S → `RankSA` with rationale | `grare.Scheme` comment names both rejected options | ✓ | |
| Total read from product list, not sub-category block | `ParseList` selector `.fs-c-productList .fs-c-listControl__status__total`; fixture has `4件中` ×2 then `3729件中`; `TestParseList` asserts 3729 | ✓ | Deviation justified; plan selector would have read 4 → complete → mark LV gone |
| Product id may be a model code (`m46203`) | Fixture card `data-product-url="m46203"`, `item_id` m46203 | ✓ | Real product, not a category |
| `-sources` flag; unknown → error before any request | `setup` `available` map; `TestCheckSources` adds `brandoff + grare`, `grare twice` | ✓ | |
| `/api/models` → `schemes` per shop | Own server: `/api/models/CHANEL/A01112` → `[brandoff 5 grades, grare 9]`; LV N41359 → `[brandoff]`; `TestModelSchemesPerShop` | ✓ | |
| Dots keyed `{source, code}` | Browser DOM: BRAND OFF plot dots ランクA 80 / ランクB 5 only; GALLERY RARE plot 中古A 1 / 中古AB 1 only | ✓ | |
| Advisor names registered shops | `buildSystemPrompt` + `TestPromptStatesThresholds` (店甲/店乙) | ✓ | |
| Scratch crawl: 200 rows, grades ∈ scheme, no `?` URL | Own copy of scratch DB: grare 200 rows; 中古A 111, 中古S 40, 未使用 31, 中古AB 9, 新品 8, 中古B 1; 0 rows with `?` or off-host URL; 0 ungraded; `crawl.log` 5 `sweep done … complete=false`, `details done fetched=5 gone=0 failed=0` | ✓ | `detail_at` set on 4 rows, not 5 — the CHANEL detail was lost when CHANEL rows were deleted + re-crawled (patch Notes); consistent |
| No request URL with `?` | Code-level: list URLs = `baseURL + bagPaths[...]` constants; product URLs from cards with `RawQuery`/`Fragment` cleared; fetcher does not log URLs, `?` hits in `crawl.log` are FX-host http2 debug lines only | ✓ | Patch request table is reconstructed (stated) |
| Cross-shop group exists | SQL: CHANEL A01112 (87), A67086 (69), LV M47270 (32), … | ✓ | |
| Sold items surface through product-page 404 (README § Knowing when something sold; `Sweep` doc) | `store.PendingDetails` selects `detail_at IS NULL` only; saved product page `detail_bag.html` template | ✗ | F1 |
| Browser-checked light/dark × desktop/mobile | Patch: model-dark + detail-light were DOM-only, tooltip not exercised; plan box ticked anyway | ⚠ | Covered by this audit's visual pass — pass |
| Vetting doc: R2 restated, 3 shops re-read, new Outcome, no new ALLU request | `git diff 90813c4..249fe3c -- docs/provider-vetting.md` | ✓ | Old verdicts kept as "first judged"; adapter-probe log appended |
| README follows def display decision (closes iteration-2 F1) | README diff § Choosing a source | ✓ | Wording matches def § Is (attribution + link-out, owner's decision, not legal judgment) |

---

## Objective Verification

### Objective: provider #2 = GALLERY RARE end to end under the restated display bar

**Evidence:** commits `249fe3c` `5c180a6`; tests; own scratch-DB copy; own browser pass.

**Findings:**
- Actual: own scheme (verbatim) + adapter + fixtures/tests + wiring + per-shop distribution + advisor shop line + README + vetting doc + scratch crawl. Display conditions (attribution + link-out) present on card (「GALLERY RARE（另開視窗）」 → `www.g-rare.com/…`, `target=_blank`, `rel=noopener noreferrer`) and detail CTA 「到 GALLERY RARE 商品頁」; 鑑定卡 shows 中古A ZH + JA verbatim + 9-step ladder + 店家鑑定 · GALLERY RARE 商品ランク; 行情怎麼算 reads 「目前的資料來自 BRAND OFF、GALLERY RARE 的線上商店」 and lists both ladders.
- Missing: working sold/price-change tracking for GALLERY RARE listings that leave page 1 (F1).
- Quality: High

**Verification:** ⚠ Partial (F1)

---

## Code/Implementation Inspection

### `internal/source/grare`

**Claimed:** scheme, `ParseList`/`applyTitle`/`ParseDetail`, first-page `Sweep`, `Detail` 404 → gone.
**Actual:** as claimed — Quality: High. Clean mirror of brandoff; heuristics documented; brand-model guard (`brandModel`) and `/`-lead guard tested on literal titles.

**Issues Found:**
- `Source.Detail` treats any non-404 page as live. Saved `detail_bag.html` carries the shop's sold-page handling: a script that looks for `.fs-c-productNotice--outOfStock` and then shows 「表記価格は最終販売した時点での価格です。」 plus a 商品お探しサービス button. A sold product page therefore most likely answers 200 with that notice, which `ParseDetail` parses as a live item — Severity: High (part of F1)
- Brand misfiles handled correctly: `grare:11532682` (title ロエベ … LOEWE) sits in the CELINE category → `brand=LOEWE`, `target=CELINE` (LOEWE 41 / CELINE 39 rows) — positive.

**Undocumented:** none beyond patch Notes.

**Metrics:** Readability: High | Test Coverage: parsers fixture + table tests; `Sweep`/`Detail` untested (network) | Tech Debt: Low

### `internal/crawl` + `internal/store` interaction (not changed, made live)

**Actual:** `store.PendingDetails` = `status='active' AND detail_at IS NULL`; sweeps mark missing items only when complete. GALLERY RARE: LV 3,729 / HERMES 5,391 bags → no brand-bag sweep completes → nothing is ever marked gone by sweeps; each listing gets one product-page fetch, usually within the crawl it is first seen (budget 150/source/crawl vs ≤200 new) — so the single 404 chance is spent while the item is still on sale. Price changes on items off page 1 are never observed either.

### `internal/web` (`server.go` `model`, `dist.js`, `model.js`, `detail.js`, `kura.css`)

**Claimed:** `schemes` per shop; per-shop ladders, shared y-scale, viewed shop first, colgroup table rows, toggle over all charts.
**Actual:** as claimed — Quality: High. `[]` initialiser keeps JSON non-null.

**Issues Found:**
- `h3.k-dist__shop` = `${shop} <small>${scheme.name} · ${graded}</small>`; both scheme names start with the shop → "BRAND OFF BRAND OFF 商品ランク · 店家鑑定" (F3) — Low
- `server.go` `model`: comment `// [] not null in JSON` names the constraint, not the trap it avoids (`distHTML`'s `schemes = []` default only replaces `undefined`; `null` would throw on `.filter`) (F5) — Low

### `internal/advisor`

**Actual:** `{{shops}}` from `listing.Schemes()`, dedup by Shop, `（尚無）` fallback; comment states why (no second list). Quality: High.
**Issues Found:** test registers `ax`/`ay` into the process-global registry with no cleanup (same in web `TestModelSchemesPerShop` `wx`/`wy`) — harmless today, leaks into later tests (F6) — Low

### `internal/radar` (not changed, made live)

- `Engine.newer` doc: "a source sweeps its list pages newest first … so an earlier first_seen is a newer listing" — no longer true for GALLERY RARE (shop `priority` order). GALLERY RARE pairs never onboard, so its listings are never watched; later-crawled GALLERY RARE arrivals sort behind its first-crawl stock in 最新上架 (F4) — Low
- `Engine.value` rank-adjusted comps include every source in the group (by design, "only comparable through the rank bridge"); with GALLERY RARE's ≤40/brand sample, its valuations are mostly BRAND OFF-priced (F2)

---

## Testing Validation

**Coverage:** grare parsers: list fixture (40 cards, per-card invariants, one exact card, `m46203`), no-count error, 8 title rows + gender row, detail fixture exact, `Complete`, `Targets`. main: `checkSources` incl. duplicate grare. web: `TestModelSchemesPerShop`. advisor: shop names.
**Execution:** All pass? Yes (`-count=1` all; `-count=3` advisor + web) | Flaky: none seen
**Missing:** `Detail` on a sold page (no fixture of one; a synthetic `.fs-c-productNotice--outOfStock` page would do); `distHTML` has no JS test (checked in browser).
**Quality Issues:** global scheme registration without cleanup (F6).

---

## Documentation Audit

**Completeness:** README (sources, display decision, limits, flag, tests) — Complete; vetting doc — Complete; rollback — plan's.
**Accuracy:** Partial — README § Knowing when something sold "a product page that answers 404 does" and `Sweep` doc "sold items surface only through a product-page 404" overstate (F1). README "GALLERY RARE contributes at most 40 bags per brand per crawl" true per crawl; rows accumulate across crawls (and stay active, F1).
**Quality:** High clarity; display-decision paragraph matches def § Is wording.

---

## Source Hygiene

**Scope:** Diff only (`90813c4..5c180a6`, excluding `**/.projex/**`)

| Rule | Location (symbol) | Quote | Severity |
|------|-------------------|-------|----------|
| — | — | No projex refs, line numbers, plan shape or bare reassurance in added comments | — |

**Rule 6 gaps:** none material. `grare.ParseList` total selector, `productID` rule, `applyTitle` lead guards, `brandModel`, `grare.Scheme` 中古S, `Source.Sweep`, `distHTML` all carry rationale naming the rejected option.

**Rule 6 vacuity:**
- `web.Server.model` — "[] not null in JSON" — names no consumer/trap (F5).

**Stale comment (rule 3, made stale by this work, not in diff):** `radar.Engine.newer` — newest-first assumption (F4).

**Rationale promotion:** README carries the page-1 / query-string limit, display decision and sold/new consequences; scheme bridge rationale lives in code (its shipped place). No projex-only load-bearing rationale found. Patch has no explicit `Rationale promoted:` line — supported by diff.

**Commit composition** (informational): 1 code commit checked — typed subject 1/1 (`feat(source)`) | type matches diff 1/1 | `Projex:` trailer 1/1 (`2610132204-kura-grare-provider-2`). `5c180a6` doc-only, exempt.
**Violations:** None

**Trailer survival on `master`:** 1/11 recent commits carry `Projex:` — the only projex code landing (`8950a55`) has it; `c4e1187`, `8ab31cb` are non-projex feature commits, rest doc-only/baseline. Channel alive — OK.

---

## Visual Pass (this audit)

Own HEAD build (`kura serve -sources grare -advisor mock`, no crawl flags) on own copy of the scratch DB, `127.0.0.1:8094`; detail/methodology via a local CSP proxy (`img-src 'self'`) on 8095 to stop shop-image hotlinks. Both stopped by PID. No shop link clicked.

| Check | Result |
|-------|--------|
| `/model/CHANEL/A01112` **dark** desktop | ✓ body `rgb(27,24,21)`, ink `rgb(238,231,219)`; 2 charts; captions in ink / ink-3 tokens; dots per ladder correct; no horizontal scroll |
| Tooltip hover (dark) | ✓ GALLERY RARE 中古A dot → 「中古A · 在售 NT$156,764」 in its own plot; other plot's tip hidden |
| Toggle (dark) with tip open | ✓ both charts hidden, both tips hidden, table 16 rows with BRAND OFF / GALLERY RARE colgroup rows; back to charts ✓ |
| Model dark 375 px | ✓ no h-scroll (375/375); 9 GALLERY RARE labels wrap to 2 lines, 「未使用」 breaks 「未使/用」 (F7, cosmetic) |
| `/listing/grare:11539001` **light** desktop | ✓ bg `rgb(245,240,230)`; GALLERY RARE ladder first, `is-this` on 中古A, 此件 dot in chart 0; per-ladder aria-labels; CTA 「到 GALLERY RARE 商品頁」; 鑑定卡 中古A ZH + JA verbatim + 9-step ladder |
| Tooltip hover (light) | ✓ 中古AB tip; moving to a BRAND OFF dot hides it and opens ランクB tip in the BRAND OFF plot |
| Detail light 375 px | ✓ no h-scroll; both ladders render |
| `/methodology` | ✓ both shops named; both ladders listed |
| Console | Only the proxy's own CSP block message; no app errors |

**Disclosure:** the first page load (model page, before the proxy) let the browser hotlink 5 product images from `www.brandoff-store.com` — the app's normal display, not a page request, but it is traffic to a shop host. Later loads blocked images. The executor's browser pass did the same. Future audits: serve through an image-blocking proxy from the first load.

---

## Gap Analysis

### Promised But Not Delivered
| Promise | Status | Impact |
|---------|--------|--------|
| Sold GALLERY RARE items leave 在售 (README / `Sweep` doc) | Missing in practice | High — stale 在售 dots, medians, advisor `cheapest_active`, comparative labels over time |
| Browser matrix fully visual | Partial in patch; completed here | None now |

### Undocumented Issues
| Issue | Severity | Affects |
|-------|----------|---------|
| Shop serves sold product pages 200 + out-of-stock notice (template evidence) → plan assumption "sold → 404" unsupported | High | `grare.Source.Detail`, README |
| Rank-adjusted comps silently cross shops (copy says 同型號各等級) | Medium | detail fair-value text, 行情怎麼算 |
| `Engine.newer` comment stale for non-newest-first sources | Low | radar maintainers, 最新上架 order |

### Unhandled Edge Cases
- Total element present but non-numeric → `Total=0` → `Complete()` true with 40 cards → marks the brand's other listings gone — Impact: Low likelihood (layout change), high blast radius; guard `Total == 0 && len(Cards) > 0` → error.
- Sold-out cards on a list page (list template ships a 'SOLD OUT' → 商品お探しサービス script; none in fixture) would be ingested as active — Impact: Low/unknown.

---

## Quality Assessment

### Completeness: High
**Strengths:** every plan step landed; deviations found by fixtures, fixed, tested, recorded.
**Gaps:** sold detection (F1).

### Correctness: Medium
**Works:** scheme/grades ✓ | parsers ✓ | per-shop ladders ✓ | advisor ✓ | `-sources` ✓ | no-query acquisition ✓
**Bugs:** F1 — Severity: Significant

### Code Quality: High
**Positive:** small, mirrored package; rationale comments name rejected options; strict total-count failure mode.
**Concerns:** F5, F6.
**Tech Debt:** GALLERY RARE freshness model (first page + one detail fetch) — Medium.

### Value Delivered: High
**Intended:** provider 2 of 3, live and correct beside BRAND OFF.
**Actual:** live, displayed correctly per shop; freshness of its rows decays after first sight.
**Impact:** User: Positive (second shop, cross-shop groups) | Portfolio: Positive (multi-scheme design visible).

---

## Open Findings

### Undocumented Discoveries
- Problems: sold-page template on product pages — Evidence: `internal/source/grare/testdata/detail_bag.html` script on `.fs-c-productNotice--outOfStock` / `.soldout-text`.
- Workarounds: executor deleted + re-crawled scratch CHANEL rows after the `brandModel` fix because `RecordSeen` does not refresh title-derived columns — acceptable for scratch; the same gap means any future `applyTitle` fix leaves old rows mis-split in a real DB (BRAND OFF shares it; pre-existing).

### Impact Analysis
- Downstream: radar valuations for GALLERY RARE lean on BRAND OFF (F2); 最新上架 never surfaces GALLERY RARE arrivals as new (README states it).
- Future: Enabled: provider #3 follows the same seam. | Harder: any freshness fix must touch `store.PendingDetails` (shared by BRAND OFF).
- Risks: stale active GALLERY RARE rows grow each crawl — Likelihood: High once crawls repeat.

### Improvements
- Could be better: per-source detail re-check for sources whose sweeps never complete (oldest `detail_at` first, within budget).
- Missed: a sold-page fixture (the shop's own template makes one synthetic case easy).

---

## Stakeholder Impact

| Stakeholder | Promised | Reality | Impact |
|-------------|----------|---------|--------|
| Owner / portfolio reviewer | 2nd shop shown with own grades + attribution | Delivered, verified visually | Positive |
| Buyer persona | 在售 GALLERY RARE listings are on sale | True at first sight; decays (F1) | Neutral → Negative over time |

---

## Findings

### Critical (Must Address)
- None.

### Significant (Should Address)
- **F1 GALLERY RARE sold detection effectively absent** — sweeps never complete (all 5 brands > 40 bags); `store.PendingDetails` fetches a product page once (`detail_at IS NULL`); saved product page shows the shop serves sold pages with `.fs-c-productNotice--outOfStock`, so even that one fetch likely reads a sold item as live. README § Knowing when something sold and `grare.Source.Sweep` doc claim a product-page 404 marks it sold. → **Small patch:** `grare.ParseDetail`/`Detail` treat `.fs-c-productNotice--outOfStock` as gone (+ synthetic test); correct README + `Sweep` doc + vetting "Crawl limits" line to say GALLERY RARE sold status is observed only at the single product-page fetch. **Follow-up (plan-sized, not a patch):** per-source detail re-check cadence for never-complete sources in `store.PendingDetails` / crawler — candidate for the I1 iteration or Deferred with blocker "design decision".

### Minor (Nice to Fix)
- **F2 Fair-value copy hides cross-shop comps** — `detail.js` rank-adjusted sentence 「把同型號各等級的 N 件換算到…」 and 行情怎麼算 paragraph omit that comps include other shops; observed `grare:11539001` 低於行情 35% on 86 comps, 85 BRAND OFF. Labels as designed stay Deferred (canvas pass 2 "cross-shop comp labels"); the copy accuracy is a small patch (e.g. 「同型號各店、各等級」). → small patch
- **F3 Caption repeats shop** — `distHTML` h3 + `scheme.name` → "BRAND OFF BRAND OFF 商品ランク". → small patch (show `scheme.name · graded` only, or strip a leading shop)
- **F4 Stale `radar.Engine.newer` comment** — newest-first assumption false for GALLERY RARE. → small patch (comment only)
- **F5 Thin rationale** — `web.Server.model` "[] not null in JSON" → name the `distHTML` default-param trap. → small patch
- **F6 Test registry leak** — advisor/web tests register schemes globally without cleanup. → small patch or accept
- **F7 375 px label break** — 「未使用」 splits 「未使/用」 in the 9-column axis; provisional layout (Deferred designed multi-provider UI). → accept or `word-break: keep-all` on axis labels
- **F8 Zero-total guard** — `ParseList` with a total element that parses to 0 but cards present reads complete. → small patch

### Positive
- Verbatim scheme matches the shop's guide exactly; product-page wording divergence correctly not used.
- Total-selector and product-id deviations caught by fixtures, explained, tested — each would have caused wrong gone-marking.
- Per-shop ladders, toggle, tooltips, a11y labels and dark/light tokens all correct on first audit pass.
- Acquisition discipline: no-query guarantee holds at code level, not just in logs.

---

## Recommendations

**Immediate:** `[patch]` for F1 small part (outOfStock → gone + doc corrections) and F2–F5, F8 (all ≤ a few lines each).
**Future:** freshness re-check for never-complete sources (F1 follow-up) before or with I1.
**Process:** browser audits block third-party images from the first load; plan assumptions about a shop's sold behaviour should be checked against saved page templates during fixture capture.

---

## Final Verdict

**Status:** Accept with Conditions

**Overall Assessment:**
- Completeness: High
- Correctness: Medium
- Quality: High
- Value: High

**Conditions:**
- [x] F1 small part: out-of-stock product page → gone; README / `Sweep` doc / vetting crawl-limit line state the real sold-detection limit — [PATCHED] 2610132229-kura-grare-audit-fixes-patch.md
- [ ] F1 follow-up recorded in the sprint nav (re-check cadence) — plan or Deferred, orchestrator's call

> Partial Execution: F1 small part, F2–F5, F8 completed via 2610132229-kura-grare-audit-fixes-patch.md (`1a95cb2`, `903ded1`); F6, F7 accepted there; F1 re-check cadence remains open (sprint Deferred / next derive).

**Sign-off:** Yes, conditional — provider 2 is live and correct at first sight; its sold/freshness claim must be made true or stated truthfully before the ≥3-providers clause is judged.
