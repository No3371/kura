# Audit: Kura Provider #3 — ALLU (ALLU公式) End to End

> **Sprint:** 2610131822-kura-def-coverage-sprint.md — iteration 4
> **Status:** Complete (Accepted with Conditions)
> **Author:** Claude Opus 5.5 (auditor, sprint subagent)
> **Parent:** 2610132325-kura-allu-provider-3-patch.md
> **Related Projex:** 2610132307-kura-allu-provider-3-plan.md (audited plan, Steps 1–5, success criteria) | 2610132325-kura-allu-provider-3-patch.md (execution claims) | 2610131822-kura-def-coverage-sprint.md (§ Current Position, remainder (d); not edited) | 2610130449-kura-mock-def.md (§ Is acquisition bar, display decision "Accept as demo") | 2610132219-kura-grare-provider-2-audit.md (F1 never-complete sweeps, precedent) | 2610132349-kura-allu-audit-fixes-patch.md (F1/F2/F3 doc parts, F5, F6, F8–F10 fixed)
> **Audit Date:** 2026-10-13 | **Work Period:** 2026-10-13, range `f1f184f..24fe7c6` on `projex/sprint/2610131822-kura-def-coverage` (`89d56a1` vetting, `471fbfb` adapter + wiring, `316b3e1` README + radar comment, `24fe7c6` patch doc)
> **Subject:** `internal/source/allu` (decoder, scheme, parsers, sweeper, fixtures, tests) | `cmd/kura` wiring | `internal/radar/radar.go` comment | README | `docs/provider-vetting.md` § ALLU + Outcome | executor's recorded evidence (crawl transport log, scratch DB, plan-time ToS/policy/FAQ copies, probe payloads — all in the session scratchpad, outside the repo)
> **Method:** No request to any ALLU host (`allu-official.com`, `faq.allu-official.com`, `image.nanboya.com`). `go vet` + `go test -count=1` on HEAD; recorded evidence decoded offline (devalue decoder written in the scratchpad); own `kura serve` on a copy of the executor's scratch DB (`127.0.0.1:8095`, `-sources allu -advisor mock`), JSON API only — no browser, so no image hotlink request; stopped by PID; worktree left clean.

---

## Audit Summary

**Claim:** ALLU onboarded as provider 3 of 3: vetting → Pass on a logged ToS read; `internal/source/allu` with own 9-rank scheme (verbatim), ALLU公式-only / in-stock / no-J filters, capped newest-first sweep, sold marker `button_type == 6`; `-sources` default `brandoff,grare,allu`; README + radar comment; scratch crawl 167 rows in 12 GETs; browser matrix pass.

**Verdict:** Partial — all 7 success criteria hold on code, tests and recorded evidence; one stated filter (ToS 第四章 partner consignments "excluded") is not implemented and its factual basis is contradicted by ALLU's own payload.

**Assessment:** Completeness: High | Correctness: Medium | Quality: High | Value: Medium

**Top Issues:**
1. **Consigned items stored as ALLU-graded stock** — vetting R3 says ToS 第四章 partner consignments are excluded; the code cannot see them on list rows and `ParseDetail` ignores the flag. The probe item read as "ALLU-held" carries `item_type_key: "consigned_item"` (fixtures: `owned_inventory_item`). (F1)
2. **25% of ALLU rows get no 行情** — 42 / 167 scratch rows have neither `model` nor a code → empty `group_key` → never valued; the patch says those brands "group by model name". (F2)
3. **R1 rests on Kura's reading of 第18条 (12)** — sprint invariant "terms prohibiting bots/automated access fail acquisition"; the reading is textually sound but is an owner call, and (15) / (18) were not addressed in the re-vet. (F3)

---

## Claims vs Evidence

| Claim | Evidence | Status | Notes |
|-------|----------|--------|-------|
| Vetting: ALLU → Pass, (12)/(d) settled with clause text, 3 + 7 logged GETs, Outcome 3 of 3 | `docs/provider-vetting.md` diff; plan-time ToS/policy copies (scratchpad) | ⚠ | (12) quote verbatim ✓; (d) search terms 比較/ランキング/順位/格付/評価 absent from both docs ✓; logs 3 + 7 rows ✓; Outcome ✓. R3 consignment text inaccurate (F1) |
| `go vet` + `go test` green, 15 allu tests | own run, `-count=1` | ✓ | All packages ok; 15 allu tests pass, names match patch |
| `-sources` default `brandoff,grare,allu`; `checkSources`; unknown source fails early | `main.go` diff; `kura crawl -h`; `kura crawl -sources nosuch` | ✓ | `unknown source "nosuch"`, exit 1, before FX refresh or any request |
| Stored rows ALLU公式, non-partner, in stock, grade ∈ scheme \ J | `ParseList` filters; `TestParseListFilters`; scratch DB | ✓ | Grades A 28 · AB 74 · B 60 · S 1 · SA 4; 0 bad URLs. Field-level true; consignment caveat (F1) |
| Scratch crawl: `source=allu` rows, ≤13 GETs, no 403/429, no `/market/user/` | executor's `crawl.log` (`GODEBUG=http2debug=2`), scratch DB | ✓ | 12 ALLU requests (robots + 6 lists + 5 products), all `KuraBot/0.1` UA, all `:status 200`, ~4 s apart; 1 extra `open.er-api.com` (FX, `Go-http-client/2.0`, not a shop). 167 rows; 6 `sweep done … complete=false`; `details done fetched=5 gone=0 failed=0` |
| Cross-shop group exists | scratch DB SQL | ✓ | CHANEL A01112 … AP0250, LV M22576 … M45595 (brandoff + allu); HERMES name-groups join BRAND OFF too (`allu:2116870` → 13 comps, rank-adjusted) |
| Browser: 鑑定卡 9 ladder + JA def; link-out; model page own ladder; 行情怎麼算 3 ladders | patch table; own API re-check | ⚠ | Not re-run in a browser (would hotlink `image.nanboya.com`). API: `/api/overview` `schemes` = allu 9 (N…J) · brandoff 5 · grare 9; `/api/listings/allu:2116870` condition `AB：ややきれい`, `desc_ja` verbatim, `condition_by` 店家鑑定, URL `…/market/items/2116870/`; `/api/models/LOUIS VUITTON/M45321` schemes [allu, brandoff]. Consistent with patch |
| README: three shops, ALLU filter, cap, limits | README diff; plan grep | ✓ | Plan grep → no hits. "model name and model code … are fields" omits F2 gap |
| Scheme definitions verbatim from 状態ランク一覧 | `condition.go` vs plan-time FAQ copy | ✓ | 9/9 labels + DescJA match character for character |
| Sold marker `button_type` 1 vs 6 | fixtures decoded offline | ✓ | `detail_bag` 1, `detail_sold` 6; conservative "== 6" choice with rejected alternative commented |
| Patch `file:ln` locators | `cat -n` of each file | ✓ | All 14 locators exact |
| Fixtures raw, sizes 448,626 / 118,372 / 117,424 | `ls`, price text | ✓ | Sizes match; `¥1,360,000` / `198,000` present in list HTML. UA of fixture GETs not independently provable (no transport log) |

---

## Objective Verification

### Step 1: Vetting record

**Evidence:** `docs/provider-vetting.md` diff; plan-time `tos.html` / `policy.html` copies.

**Findings:**
- Actual: heading Pass; re-vet block; first-judgement evidence line updated; two request-log tables; new Outcome, old one marked superseded. 第18条 (12) key phrases match the recorded ToS verbatim. Neither document contains 比較 / ランキング / 順位 / 格付 / 評価.
- Missing: (15) 「営利を目的として本サービスを利用する行為」 and (18) 「その他、当社が不適切と判断する行為」 not discussed in the re-vet (first judgement covered site policy ④ 営利目的 only). R3 consignment sentence wrong (F1).
- Quality: High

**Verification:** ⚠ Partial — **Issues:** F1, F3.

### Step 2: Decoder, scheme, parsers, fixtures, tests

**Evidence:** `payload.go`, `condition.go`, `parse.go`, tests, fixtures decoded offline.

**Findings:**
- Actual: devalue hydrator per plan (memo, cycle → nil, specials, typed tags); `node[T]` single-prefix rule; filters as specified; `category`, `modelCode`, `ParseDetail` + `soldButton`; J rationale names both rejected alternatives.
- Missing: `"null"` tag decodes keys as indices (F5); no guard when value drift filters every row (F6).
- Quality: High

**Verification:** ✓ Verified (with Minor F5, F6, F8).

### Step 3: Sweeper + wiring

**Evidence:** `source.go`, `main.go`, `main_test.go`, `TestSweepCap`.

**Findings:**
- Actual: cap 10, lowered by `-max-pages`; non-200 → error; empty page before total → error; 404 → gone; `baseURL` var with rationale; registered in `available`; default + help updated.
- Missing: `Source.Detail` branches and `Sweep`'s complete / empty-page paths untested (F8).
- Quality: High

**Verification:** ✓ Verified.

### Step 4: README + radar comment

**Evidence:** README diff; `radar.go` `newer` comment.

**Findings:** All plan bullets present; comment present tense. README silent on F1 consignments and F2 ungrouped share.

**Verification:** ✓ Verified (F1/F2 doc follow-ups).

### Step 5: Scratch crawl + browser

**Evidence:** transport log, scratch DB, own API pass.

**Findings:** Request count, UA, statuses, rows, grades, URLs, cross-shop groups all reproduce from evidence. Browser matrix taken on the executor's word plus API consistency.

**Verification:** ✓ Verified (browser ⚠ by proxy).

---

## Code/Implementation Inspection

### `internal/source/allu` → `ParseList` / `ParseDetail` (consignments)

**Claimed:** R3 — ALLU公式 with `is_allu_partner == false`; 第四章 partner consignments excluded with ALLUパートナー.
**Actual:** List rows carry no consignment field (30/30 LV rows: no `is_consignment`; `type`, `system_type`, `display_type` identical between the probe's consigned row 2118219 and owned rows). Product payload carries `is_consignment` and `item_type_key`; `apiAsset` reads neither. ToS 第四章 defines 委託アイテム as ALLUパートナー-owned items **listed by 当社** (so `app_customer.name` = ALLU公式) whose 状態確認及び真贋鑑定 happen after purchase (購入の流れ (5)). Probe item 2118219: `is_consignment: true`, `item_type_key: "consigned_item"`, seller ALLU公式, `is_allu_partner: false`; both fixtures: `is_consignment: false`, `item_type_key: "owned_inventory_item"`. The "最短翌日 → ALLU-held" reading does not survive ALLU's own type key. — Quality: Medium

**Issues Found:**
- Consigned (partner-owned, possibly pre-inspection) bags stored and labelled 店家鑑定 under ALLU's scheme; record says they are excluded — Severity: High (F1)

### `ParseList` → model keys

**Actual:** `ModelName` ← `model` only; `ModelCode` ← `item_number` or a code-shaped title token. Scratch rows with neither: CHANEL 20/27, GOYARD 11/29, CELINE 9/27, LOEWE 2/28 (42/167 = 25%) vs BRAND OFF 101/15,367 (0.7%). `radar` skips empty `GroupKey` → those listings never get fair value, deal score or model page. — Severity: Medium (F2)

### `payload.go` → `hydrator.typed`

devalue serialises a null-prototype object as `["null", "key", idx, …]` — keys are literal strings, unlike `Map` (indices). `typed` resolves both through `h.ref`, so a `"null"` key becomes `<nil>` and pairs collide. No `"null"` tag in any fixture → latent. — Severity: Low (F5)

### `ParseList` → filter drift

A value change (seller renamed, `sale_status` code change) drops every row silently: `Rows > 0`, `Cards == 0`, no error. Never-complete brands go stale without a signal; a brand small enough to complete (≤300 ALLU公式 bags) would have every listing marked gone after two such sweeps. Key drift already errors; value drift does not. — Severity: Low (F6)

### `internal/crawl` (pre-existing, inherited)

`Fetcher.do` returns 403 as a non-retryable error; `FetchDetails` counts it `failed` and continues through the detail budget (up to 150 more GETs); `RunAll` continues to the next brand. The acquisition bar's "stop on 403/429/challenge" is honoured by the executor's manual runs, not by the crawler at runtime. Same for every source; not introduced here. — Severity: Low (F7)

**Undocumented:** none beyond F1/F2; the diff matches the patch's file list exactly.

**Metrics:** Readability/Maintainability: High | Test Coverage: parsers and decoder well covered, `Source.Detail` uncovered | Technical Debt: Low

---

## Testing Validation

**Coverage:** decoder (shared index, specials, Set/Date/Map/EmptyRef, cycle), `node` ambiguity, list fixture + synthetic filter table, category / model-code tables, detail + sold fixtures, no-node error, `Complete`, targets, URL, cap (2 / 0 / 50), unknown brand.
**Execution:** All pass? Yes (`go test -count=1 ./...`) | Flaky tests: none seen
**Missing:** `Source.Detail` (404 → gone, 403/5xx → error, 200 sold → gone); `Sweep` returning `complete=true` on the last page and the empty-page error; `"null"` tag; consignment payload (no fixture).
**Quality Issues:** `TestParseList`'s dropped-row counter omits the "grade not in Scheme" and `price ≤ 0` classes, so it mirrors the filter imperfectly (passes today because the fixture has none).

---

## Documentation Audit

**Completeness:** User (README) Partial — omits F1, F2 | Vetting record Partial — R3 inaccurate (F1), (15)/(18) unaddressed (F3) | Migration/Rollback n/a (additive source).
**Accuracy:** Matches implementation? Partial (R3 consignments; patch "those group by model name") | Examples work? Yes (`-sources`, help text).
**Quality:** Clarity High | Completeness Medium | Usability High

---

## Source Hygiene

**Scope:** Diff only (`f1f184f..24fe7c6`, excluding `**/.projex/**`)

| Rule | Location (symbol) | Quote | Severity |
|------|-------------------|-------|----------|
| — | — | No rule 1–5 violation: no projex refs, no `file:ln`, present tense, no plan banners, no bare reassurance | — |

**Rule 6 gaps** (non-obvious change, no rationale comment):
- `allu` → `ParseList` top-category check — the request URL already scopes to `c-58` (バッグ); why the payload's `item_category_top_name` is re-checked is unstated (Low, F9).

**Rule 6 vacuity:** none — `hydrator`, `baseURL`, `maxListPages`, `Sweep`, `soldButton`, J grade and `cleanText` each name the rejected alternative.

**Rationale promotion:** none needed — supported by diff; load-bearing rationale (cap, J exclusion, sold marker, never-complete consequences) lives in code comments, README and the vetting record.

**Commit composition** (informational — fix forward, never rewrite history): 3 landed commits outside `.projex/` checked — typed subject 3/3 | type matches diff 3/3 (`docs` for README + comment-only `radar.go`) | `Projex:` trailer 3/3
**Violations:** None

**Trailer survival on `master`:** 1/10 recent commits carry `Projex:` — the only projex code landing (`8950a55`); rest doc-only / baseline / pre-convention features. Channel alive — OK (same as iteration-3 audit).

---

## Gap Analysis

### Promised But Not Delivered
| Promise | Status | Impact |
|---------|--------|--------|
| ToS 第四章 partner consignments excluded (vetting R3) | Missing | High |
| Browser matrix re-verifiable | Partial (executor evidence + API proxy) | Low |

### Undocumented Issues
| Issue | Severity | Affects |
|-------|----------|---------|
| 25% of ALLU rows ungroupable → no 行情 | Medium | radar, Browse, advisor |
| `"null"` devalue keys mis-decoded | Low | `payload` (latent) |
| Silent empty result on value drift | Low | sweep freshness / gone marking |
| Crawler continues after 403 | Low | acquisition bar at runtime (all sources) |

### Unhandled Edge Cases
- Sold rows visible in the 10 newest pages (`stock 0, sale_status 2`) are dropped, not used to mark a stored listing gone — Impact: the one cheap sold signal for never-complete brands is discarded (F4).
- A stored listing whose row later reads `sale_status ≠ 1` with stock (reserved/other states) → unseen → in a complete sweep, gone after two sweeps — Impact: only brands ≤300 bags; unknown whether any qualifies.
- `Brand` falls back to the crawl target when `BrandByName` misses — Impact: none seen (title prefixes match brand on all 167 rows).

---

## Quality Assessment

### Completeness: High
**Strengths:** every plan step landed with the planned files; extra tests; request evidence reproducible from the transport log.
**Gaps:** consignment exclusion; ungrouped share undocumented.

### Correctness: Medium
**Works:** filters, cap, sold marker, scheme, wiring, link-out: Yes
**Bugs:** consigned items admitted (F1) — High | `"null"` tag keys (F5) — Low

### Code Quality: High
**Positive:** small, typed decoder; errors over silent data for key drift and missing counts; every non-obvious choice commented with its rejected alternative.
**Concerns:** value-drift blind spot; `Source.Detail` untested.
**Tech Debt:** never-complete sweeps (accepted, remainder (d)) — Medium

### Value Delivered: Medium
**Intended:** ≥3 providers with own schemes and fair, attributed display.
**Actual:** third shop live with verbatim scheme and cross-shop groups; a quarter of its rows unvalued and an unknown share pre-inspection consignments.
**Impact:** User: Positive (more comps, LV/CHANEL code groups across shops) | Business: def ≥3 clause reachable, pending F1/F3 owner calls.

---

## Open Findings

### Undocumented Discoveries
- Changes: none outside the patch's file list.
- Problems: consignments admitted (F1) — Evidence: probe payload `item_type_key: "consigned_item"` + ToS 第四章 text.
- Workarounds: request log from `GODEBUG=http2debug=2` instead of crawler lines — acceptable, and stronger evidence (headers + status per request).

### Impact Analysis
- Downstream: radar — third shop in groups and rank-bridged comps (works; `allu:2116870` valued from 13 BRAND OFF comps).
- Future: Enabled: ≥3 providers judgement | Blocked: nothing; remainder (d) now spans GALLERY RARE + ALLU.
- Risks: default `-sources` now includes ALLU, so every `kura crawl`, `serve -crawl-every` and 立即更新 sends up to 1 + 60 list + 150 product GETs to ALLU — Likelihood: High (by design; bears on the (12) 「必要以上に繰り返す」 reading, F3).

### Improvements
- Could be better: product-page `is_consignment` / `item_type_key` captured at detail time (F1).
- Would make excellent: list-row sold signal feeding gone-marking (F4); title-derived model name for CHANEL/GOYARD/CELINE (F2).
- Missed: LV is the largest brand (fixture `total_count` 49,388 ≈ 1,647 pages), not HERMES (22,135) — `maxListPages` comment and R5 understate it (F10).

---

## Stakeholder Impact

| Stakeholder | Promised | Reality | Impact |
|-------------|----------|---------|--------|
| Owner | ALLU公式 stock only, ALLU's own appraisal | Also partner-owned consigned items under ALLU公式 | Negative (decision needed) |
| Demo visitor | ALLU listings with 行情 | 75% valued; 25% shown without fair value | Neutral |
| ALLU | Identified, paced, capped bot | Verified: KuraBot UA, ~4 s spacing, 12 GETs | Positive |

---

## Findings

### Critical (Must Address)
- None.

### Significant (Should Address)
- **F1 — Consigned items admitted; R3 record says excluded.** ALLU's payload types item 2118219 `consigned_item` (fixtures `owned_inventory_item`); ToS 第四章: partner-owned, listed by ALLU, inspected after purchase. List rows carry no flag, `ParseDetail` ignores `is_consignment`. → **Small doc fix (patch):** vetting R3 + README state that ALLU公式 includes consigned items, identified only on the product page (`is_consignment` / `item_type_key`), and that Kura currently keeps them. **Human decision:** keep consignments (with that disclosure) or exclude them — exclusion is plan-sized (detail-time exclusion path in `crawl`/`store`; `MarkGone` would wrongly count them as sold).
- **F2 — 42/167 ALLU rows (25%) ungroupable → no 行情.** CHANEL 20, GOYARD 11, CELINE 9, LOEWE 2 have no `model`, no `item_number`, no code token; patch says those brands "group by model name". → **Small doc fix (patch):** README § Turning messy listings + Limitations state the share. **Follow-up patch (code, moderate):** title-derived model-name fallback when `model` is empty, tested on fixture titles.
- **F3 — R1 Pass rests on Kura's reading of 第18条 (12).** BOT is named only with 不正に操作; Kura's paced read-only crawl is plausibly outside it, but the sprint invariant fails any terms prohibiting bots, (18) 「その他、当社が不適切と判断する行為」 and (15) 営利目的 were not addressed in the re-vet, and default sources now send up to ~211 GETs per crawl. → **Human decision:** owner confirms or overrules (overrule → remove `allu` from `available`, record Fail (acquisition), ≥3 → Deferred). Small doc fix: add (15)/(18) to the re-vet block.

### Minor (Nice to Fix)
- **F4 — List-visible sold rows discarded.** Rows with `stock 0, sale_status 2` in the 10 newest pages are proof of sale for listings Kura already holds, but `visit` only receives kept cards. → **Plan, with remainder (d)** (needs a `crawl.Source` / store contract change; not a provider-only patch).
- **F5 — devalue `"null"` keys decoded as indices.** → **Small code fix:** read `rest[k]` as a string for `"null"`; add a `TestHydrate` case.
- **F6 — Value drift silently empties a page.** → **Small code fix:** `ParseList` errors when `Rows > 0` and no row passes the ALLU公式 check; test.
- **F7 — Crawler continues after 403 (pre-existing, all sources).** → **Human decision on scope** ("only work on providers"): crawler-wide stop-on-403 per source, or a provider-local sticky stop in `allu.Source`.
- **F8 — Test gaps.** `Source.Detail` branches; `Sweep` complete / empty-page paths; `TestParseList` dropped-counter classes. → **Small code fix (tests only).**
- **F9 — Rule 6 gap:** top-category re-check lacks rationale. → **Small code fix (comment).**
- **F10 — Largest brand understated.** `maxListPages` comment and vetting R5 cite HERMES 22k; LV fixture `total_count` 49,388. → **Small doc/comment fix.**

### Positive
- Scheme definitions verbatim 9/9 against the recorded FAQ; (12) quoted verbatim; (d) search reproducible.
- Request discipline reproducible from a per-request transport log: 12 GETs, honest UA, all 200, under budget.
- Sold marker chosen conservatively from a fixture diff, rejected alternative recorded in code.
- Patch locators and fixture sizes exact; source hygiene clean; commits typed and trailed.

---

## Recommendations

**Immediate:** [patch] for F1/F2/F3 doc corrections + F5, F6, F8, F9, F10. Surface F1 (keep/exclude consignments), F3 (owner confirms the (12) reading) and F7 (scope) to the owner.
**Future:** F4 and F2 code fallback with remainder (d) / a follow-up patch; consignment exclusion if the owner asks.
**Process:** When a probe payload offers a typed field (`item_type_key`), prefer it over an inference from shipping text before writing a seller rule.

---

## Final Verdict

**Status:** Accept with Conditions

**Overall Assessment:**
- Completeness: High
- Correctness: Medium
- Quality: High
- Value: Medium

**Conditions:**
- [x] F1 documented truthfully (vetting R3 + README) — [PATCHED] 2610132349-kura-allu-audit-fixes-patch.md
- [ ] F1 owner decides keep vs exclude consignments
- [x] F2 limitation stated in README — [PATCHED] 2610132349-kura-allu-audit-fixes-patch.md
- [x] F3 (15)/(18) addressed in the record — [PATCHED] 2610132349-kura-allu-audit-fixes-patch.md
- [ ] F3 owner confirms (or overrules) the 第18条 (12) reading

> Partial Execution: F1/F2/F3 doc parts, F5, F6, F8, F9, F10 completed via 2610132349-kura-allu-audit-fixes-patch.md (`5a8a71f`, `ecadf0b`); owner decisions F1 (keep/exclude), F3 ((12) reading), F7 (scope) and F4 / F2 code fallback remain open.

**Sign-off:** Yes, conditional. The adapter does what the plan specified and the acquisition evidence reproduces. The open items are one record that overstates the seller filter, one undocumented coverage gap, and an owner decision that was already flagged.
