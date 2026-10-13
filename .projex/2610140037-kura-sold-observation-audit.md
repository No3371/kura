# Audit: Kura Sold Observation for Never-Complete Sources

> **Sprint:** 2610131822-kura-def-coverage-sprint.md — iteration 5
> **Status:** Complete (Accept with Conditions)
> **Author:** Claude Opus 5.5 (auditor, sprint subagent, uid `2610131822-kura-def-coverage#5`)
> **Parent:** 2610140023-kura-sold-observation-patch.md
> **Related Projex:** 2610140010-kura-sold-observation-plan.md (audited plan, success criteria) | 2610140023-kura-sold-observation-patch.md (audited execution) | 2610131822-kura-def-coverage-sprint.md (not edited) | 2610132219-kura-grare-provider-2-audit.md (F1 source) | 2610132339-kura-allu-provider-3-audit.md (F4 source)
> **Audit Date:** 2026-10-14 | **Work Period:** 2026-10-14, range `ff3d9b7..e9a2dd4` on `projex/sprint/2610131822-kura-def-coverage` (`9003350` code + tests, `90e04ea` README + vetting, `e9a2dd4` patch doc + plan close)
> **Subject:** `internal/store` (`MarkGone`, `RecheckDue`, `scanTodos`) | `internal/crawl/crawler.go` (`Source.Sweep` contract, `SweepTarget`, `RunAll`, `FetchDetails`, `RecheckAfter`) | `internal/source/{allu,grare,brandoff}` | `cmd/kura` `-recheck-after` | `README.md` | `docs/provider-vetting.md`
> **Method:** No request to any shop host. Diff read in full. `go vet ./...` + `go test -count=1 ./...` + `-v` on new/updated tests in the worktree (read-only; Go cache only). gofmt on committed content of all 43 tracked `.go` files (`git show HEAD:<f> | gofmt -l`; worktree is CRLF). Tree of `e9a2dd4` exported (`git archive`) to the session scratchpad: 14 mutation probes run there, plus a scratch build for `kura crawl -h`. Saved list fixtures of all three shops scanned for sold markers. Scratch removed; worktree left clean.

---

## Audit Summary

**Claim:** never-complete GALLERY RARE / ALLU listings get marked sold through (1) budgeted product-page re-checks of unseen listings in incomplete targets, ≤ 1 per `RecheckAfter` (24 h), within `DetailBudget`; (2) ALLU sold list rows (`stock ≤ 0 && sale_status == 2`) via a new `sold []string` param on `visit`. `MarkGone` idempotent, active-only, no `detail_at` write.

**Verdict:** Verified — all 9 plan success criteria hold in code and tests; tests are strong (14/14 mutations killed). Sprint-level gap found outside the plan's scope (F1).

**Assessment:** Completeness: High | Correctness: High | Quality: High | Value: Medium (BRAND OFF sold gap, F1, limits the ≥3-providers sold story)

**Top Issues:**
1. **F1** BRAND OFF list pages show `SOLDOUT` cards; `brandoff.ParseList` stores them as active listings. Fixture: 7 of 90 cards (6 bags). Falsifies plan Assumption 1; pre-existing, not a regression.
2. **F2** "no more product-page requests than before" / "(no added requests)" conflates the unchanged cap with actual volume, which rises to the cap. Feeds the owner's 第18条 (12) reading.
3. **F3** `-max-pages` help, README flags table and README § Knowing when something sold still say capped sweeps never mark anything sold/gone — no longer true.

---

## Claims vs Evidence

| Claim | Evidence | Status | Notes |
|-------|----------|--------|-------|
| `MarkGone` active-only, `(bool, error)`, first `gone_at` kept, unknown id no-op, no `detail_at` | `store.MarkGone`; `TestMarkGoneOnlyActive`; mutations M5 (drop `status='active'`), M8b (re-add `detail_at`) killed | ✓ | |
| `RecheckDue` selection/order/limit/empty targets | `store.RecheckDue`; `TestRecheckDue`; M4 (order), M7 (`<=` on `detail_at`), M14 (`<=` on `last_seen`) killed | ✓ | `target IN (?…)` built from count only, values bound — no injection |
| `visit` gains `sold []string`, 3 adapters + fake compile | `crawl.Source.Sweep`; `allu`/`grare`/`brandoff` `Sweep`; build green | ✓ | BRAND OFF passes `nil` though its list shows sold cards (F1) |
| Sold id marks held listing gone, `run.Gone +=`, unknown id inserts nothing | `SweepTarget`; `TestListSoldRowsMarkGone`, `TestListSoldRowsAddToSweepMissing`; M10 (`=`), M11 (ignore sold) killed | ✓ | |
| ALLU `Sold == ["2117814"]`; only ALLU公式 non-partner `stock ≤ 0 && 2` | `allu.ParseList` (seller check precedes); `TestParseList`, `TestParseListFilters`; M6 (any non-1 status) killed; M12 (sold not passed) killed by `TestSweepComplete` | ✓ | |
| First fetches first; re-checks only incomplete targets, unseen since crawl start, ≥ `RecheckAfter` old; 0 = off | `RunAll` `incomplete`, `FetchDetails` `left`; `TestRecheckIncompleteTargetsOnly`, `TestRecheckCadenceAndPriority`; M1, M2b, M3, M9 killed | ✓ | failed sweeps (`!run.Complete`) count as incomplete — documented in `RecheckAfter` |
| Re-check gone → gone; live → `SaveDetail` bumps `detail_at` | `FetchDetails` switch; M13 killed; cadence test rotation `HERMES-0`→`HERMES-1` | ✓ | |
| `-recheck-after` default 24h wired | `cmd/kura` flag + `setup`; scratch `kura crawl -h` → `(default 24h0m0s)` | ✓ | |
| README + vetting state behaviour; no "fetched once per listing" / "Nothing re-checks" | `git grep` over tracked files | ⚠ | target phrases gone; adjacent stale claims remain (F3); volume wording (F2) |
| gofmt / vet / test green | re-run | ✓ | 43/43 committed `.go` files gofmt-clean; vet exit 0; 10 packages ok |
| Per-run product-page requests ≤ `DetailBudget` per source | `FetchDetails`: `PendingDetails(…, DetailBudget)` + `RecheckDue(…, DetailBudget-first)`; M9 killed | ✓ | crawl invariant met; volume claim wording is F2 |
| `HasDetail` read only inside `store` | `git grep HasDetail` → `store.go` def + scan only | ✓ | |
| Deviation: superseded vetting Outcome has no "one product-page fetch per listing" | `docs/provider-vetting.md` `## Outcome (2026-10-13, after owner decision) — superseded`, GALLERY RARE bullet | ✗ | text exists, line-wrapped ("one product-page" / "fetch per listing"); single-line grep missed it. Outcome (left untouched) is what the plan wanted (F6) |
| Mutation check `+=` → `=` | M10 reproduced | ✓ | |

---

## Objective Verification

### Step 1 — idempotent `MarkGone`, `RecheckDue`
**Evidence:** `store.MarkGone`, `store.RecheckDue`, `store.scanTodos`; `TestMarkGoneOnlyActive`, `TestRecheckDue`; call sites `crawl.FetchDetails`, `web` `server_test`/`wanted_test`, `advisor` `tools_test`.
**Findings:** Actual: as planned; two extra test call sites (documented deviation). `Store.Version` still bumps on a gone mark (`MAX(gone_at)` term). Quality: High.
**Verification:** ✓ Verified

### Step 2 — contract + ALLU sold rows
**Evidence:** `crawl.Source.Sweep` doc/signature; `SweepTarget` sold loop; `allu.soldStatus`, `ListPage.Sold`; tests above.
**Findings:** Actual: as planned. Missing: BRAND OFF list sold cards exist and go unsignalled (F1) — outside plan scope, but the new contract text "(nil when the shop's list pages show none)" is false for BRAND OFF. Quality: High.
**Verification:** ✓ Verified (plan scope) | ⚠ contract truth for BRAND OFF (F1)

### Step 3 — re-check cadence + flag
**Evidence:** `Crawler.RecheckAfter`, `RunAll`, `FetchDetails`, `cmd/kura`; cadence/scope tests.
**Findings:** Budget arithmetic correct; `last_seen < floor(crawlStart)` excludes listings seen this crawl (M1/M14 killed). Persistent-failure head-of-line (F5) as plan Risks state. Quality: High.
**Verification:** ✓ Verified

### Step 4 — shipped docs
**Evidence:** README fetcher bullet, § Knowing when something sold, flags row, § Tests, § Limitations; vetting GR crawl limits + ALLU re-vet Outcome.
**Findings:** Planned text landed verbatim. Stale adjacent claims (F3); volume wording (F2). Quality: Medium.
**Verification:** ⚠ Partial (criterion's two phrases ✓; adjacent stale text remains)

---

## Code/Implementation Inspection

### `internal/store`
**Claimed:** active-only `MarkGone`, `RecheckDue`, shared scan. **Actual:** matches — Quality: High.
**Issues Found:** F4 hazard comment gap (Low).
**Undocumented:** none.

### `internal/crawl/crawler.go`
**Claimed:** contract change, sold marks, re-check phase. **Actual:** matches — Quality: High.
**Issues Found:** `RecheckAfter` doc "so a run makes no more product-page requests than before" — F2 (Medium; also rule 3 framing).
**Undocumented:** `TestListSoldRowsAddToSweepMissing` added (documented deviation, good).

### `internal/source/brandoff`
**Claimed:** `visit(page, bags, nil)`. **Actual:** matches. **Issues Found:** F1 — `ParseList` keeps `.product__item--soldout` cards as listings; `Sweep` passes `nil` (Significant, pre-existing).

**Metrics:** Readability/Maintainability: High | Test Coverage: all new branches mutation-checked | Technical Debt: Low

---

## Testing Validation

**Coverage:** every new branch (MarkGone guard, RecheckDue predicates/order/limit, incomplete-target scoping, budget split, cadence, `RecheckAfter=0`, sold-row marking/accumulation, ALLU sold filter, ALLU sold passthrough) killed by at least one test — 14/14 mutations (M1–M14; M2 and M8 re-run as M2b/M8b after the first sed variants produced a build error / a no-op SQLite semantics).
**Execution:** All pass: Yes | Flaky: none seen.
**Missing:** BRAND OFF parser test asserts 90 cards from a fixture holding 7 `SOLDOUT` cards — it pins the F1 behaviour.
**Quality Issues:** none.

---

## Documentation Audit

**Completeness:** README + vetting — Complete for the planned scope.
**Accuracy:** Partial — F2 (volume), F3 (capped sweeps / "only after … two consecutive complete sweeps"), README § Knowing when something sold says nothing of BRAND OFF `SOLDOUT` cards (F1).
**Quality:** Clarity High | Completeness Medium | Usability High.

---

## Source Hygiene

**Scope:** Diff only (`ff3d9b7..e9a2dd4`, `.projex/` excluded)

| Rule | Location (symbol) | Quote | Severity |
|------|-------------------|-------|----------|
| 3 | `crawl.Crawler.RecheckAfter` | "so a run makes no more product-page requests than before" | Low (wording also inaccurate → F2) |

No projex ids, `file:ln`, plan-shape banners or bare reassurance in added comments. `detail_sold.html` / item `2117814` in `allu.soldStatus` cite a shipped fixture — fine.

**Rule 6 gaps:**
- `brandoff.Source.Sweep` — passes `nil` sold ids although its list pages render `SOLDOUT` cards — the contract reads `nil` as "list pages show none"; a reader takes it as a checked fact (F1).
- `store.MarkGone` — names the relist trap of stamping `detail_at` but not the opposite trap: a first-fetch gone row that a list page keeps showing is reactivated with `detail_at NULL` and refetched every crawl, moving `gone_at` forward (F4).

**Rule 6 vacuity:** none — `MarkGone`, `soldStatus`, `RecheckAfter` each name rejected alternatives.

**Rationale promotion:** patch doc has no `**Rationale promoted:**` line (F7). Diff supports promotion for behaviour (README § Knowing when something sold, vetting doc); the F4 hazard lives only in the patch doc's "For audit" note — unpromoted.

**Commit composition** (informational — fix forward, never rewrite history): 3 landed commits checked — typed subject where required 2/2 (`feat(crawl)`, `docs(crawl)`; `e9a2dd4` doc-only, exempt) | type matches diff 2/2 | `Projex:` trailer 3/3
**Violations:** None

**Trailer survival on `master`:** 1/10 (master holds 10 commits; sprint not yet landed; earlier squash landings e.g. `8950a55` carry none) — not zero, so not Critical; informational. Sprint branch: 13/30 (rest doc-only).

---

## Gap Analysis

### Promised But Not Delivered
| Promise | Status | Impact |
|---------|--------|--------|
| No live text claiming the old sold behaviour | Partial (F3) | Low |

### Undocumented Issues
| Issue | Severity | Affects |
|-------|----------|---------|
| F1 BRAND OFF `SOLDOUT` list cards stored and kept active | High | BRAND OFF listings, comps, radar deals, sprint ≥3-providers sold claim |
| F2 request-volume wording | Medium | owner's 第18条 (12) reading (F3 owner-held) |
| F6 patch deviation note wrong | Low | patch doc accuracy |

### Unhandled Edge Cases
- Persistently failing re-check reads keep `detail_at` → head of `ORDER BY detail_at` every crawl → can take the whole leftover budget (F5; plan Risks; bounded).
- List page that keeps showing a sold item + product page that reports gone → active↔gone flap each crawl (F4). BRAND OFF `Detail` reports gone only on 404 / `Error.aspx`, so whether F1 triggers it is unknown offline.

---

## Quality Assessment

### Completeness: High
**Strengths:** all four steps; extra test for the `+=` accumulation; extra call sites caught.
**Gaps:** F1 (scope-adjacent), F3.

### Correctness: High
**Works:** idempotence: Yes | budget split: Yes | cadence boundary (Unix s): Yes | incomplete-target scoping incl. failed sweeps: Yes | strict ALLU sold reading: Yes
**Bugs:** none in the delivered code. F1 is a pre-existing parser gap.

### Code Quality: High
**Positive:** minimal contract change; selection by existing `detail_at` (no schema change); rationale comments name rejected alternatives.
**Concerns:** F2 wording; F4 comment gap.
**Tech Debt:** Low.

### Value Delivered: Medium
**Intended:** sold observation for the two never-complete providers. **Actual:** delivered for both. **Impact:** User: Positive (GR/ALLU sold items stop staying 在售 forever). Business: the third provider (BRAND OFF), assumed covered by `SweepMissing`, shows sold items as cards — F1 keeps "sold observed for all 3" short of true.

---

## Open Findings

### Undocumented Discoveries
- Problems: BRAND OFF `list_lv.html` holds 7 `.product__item--soldout` cards (`SOLDOUT`) among 90: pids `2101220306854`, `2118400254263`, `2101220320522`, `2107601195543`, `2100301266889`, `2100301259737` (bags, cat `100xxx`) + `2121400023012` (cat `115100`, non-bag). `brandoff.ParseList` has no soldout check; `brandoff` `TestParseList` expects 90 cards. Each crawl `RecordSeen` keeps them active; a complete sweep sees them, so `SweepMissing` never counts them missed and they are not re-check candidates. How long BRAND OFF keeps sold cards listed: unknown offline.
- GALLERY RARE `list_lv.html`: `SOLD OUT` appears only in a page script; no card carries it, and the page's filter offers 「在庫なし商品を含む」 as an opt-in — supports the `grare.Source.Sweep` comment "List pages show no sold rows".

### Impact Analysis
- Downstream: `radar` — re-check gone dates lag sales by ≤ one re-check cycle; GR brands that complete (≤ 40 bags) are onboarded/watched but are not re-checked while complete, so time-to-sell is affected only for a brand that flips complete → incomplete (negligible). The plan's "GR/ALLU rows are never `watched`" is slightly overstated; no effect.
- Future: Enabled: BRAND OFF list-row sold signal is a one-adapter change on the new contract.
- Risks: re-check volume reaches the `-details` cap every crawl once held GR/ALLU listings exceed it — Likelihood: High (by design; F2 is about saying so).

### Improvements
- Could be better: state steady-state request volume (F2).
- Missed: BRAND OFF sold cards (F1) — the new contract made it a small fix.

---

## Stakeholder Impact

| Stakeholder | Promised | Reality | Impact |
|-------------|----------|---------|--------|
| Demo viewer | GR/ALLU sold items stop showing 在售 | Yes, within ≤ 1 re-check cycle; ALLU sold rows immediately | Positive |
| Shops (crawl politeness) | no extra requests per run beyond `-details` | Cap held; daily volume rises toward the cap | Neutral |
| Owner (F3 reading) | re-check volume stated in vetting Outcome | stated as cap with "(no added requests)" | Negative until F2 fixed |

---

## Findings

### Critical (Must Address)
- None.

### Significant (Should Address)
- **F1 — BRAND OFF `SOLDOUT` cards stored as active listings** — 7/90 fixture cards (6 bags) are sold yet `brandoff.ParseList` emits them; they stay 在售, count as active comps, can surface as radar deals, and never go missing for `SweepMissing`. Falsifies plan Assumption 1 ("A card on a list page is for sale … existing `RecordSeen` premise for GR / BRAND OFF"), which the plan asked to verify early; the patch flagged it only as hypothetical. Pre-existing, not a regression. → **Code-level, follow-up patch:** route `.product__item--soldout` cards' `pid` to a `ListPage.Sold` instead of `Cards`; keep the raw row count for `Sweep`'s last-page test (`len(lp.Cards) < pageSize`); `visit(page, bags, lp.Sold)`; update `brandoff` parser test (cards 90 → 83, `Sold` 7; non-bag sold ids are harmless — `MarkGone` on an unheld id is a no-op) and README § Knowing when something sold; fix the `brandoff.Source.Sweep` rule-6 gap. Evidence is the shop's own `SOLDOUT` label in the saved fixture; no request needed. **Sprint absorb:** the sprint's Current Position ("sold observation for two of the three providers") should record BRAND OFF's gap before finalize ticks the ≥3 clause.
- **F2 — request-volume wording** — `crawl.Crawler.RecheckAfter` comment, README § Knowing when something sold ("no more product-page requests than before") and vetting GR crawl limits ("(no added requests)") are true of the cap only. Before: product pages per crawl ≈ new listings. After: up to `-details` (150) per source per crawl whenever re-check candidates exist — e.g. up to 600/day/source at `-crawl-every 6h`. Within each source's stated budget (no constraint breach), but the owner's 第18条 (12) reading is fed this line. → **Doc-level, follow-up patch:** state cap + steady-state volume ("up to `-details` per source per crawl; re-checks fill what first fetches leave"); drop "than before" (rule 3).

### Minor (Nice to Fix)
- **F3 — stale capped-sweep claims** — `cmd/kura` `-max-pages` help and README flags row "capped sweeps never mark items sold"; README § Knowing when something sold "marks an item gone only after … two consecutive complete sweeps. Capped sweeps (`-max-pages`) never mark anything gone". Capped ALLU sweeps now mark sold rows gone, and capped targets are re-checked. → **Code/doc-level, follow-up patch.**
- **F4 — unpromoted flap hazard** — first-fetch gone no longer stamps `detail_at`; a list page that keeps showing that item reactivates it and it is refetched/re-marked each crawl (`gone_at` moves forward). Lives only in the patch doc. → **Code comment, follow-up patch** (add to `store.MarkGone`); F1's fix removes the main trigger.
- **F5 — persistent re-check failures head the queue** — failing reads leave `detail_at`, so they lead `RecheckDue` every crawl and can absorb the leftover budget. Plan Risks; bounded; stop-on-403 part is owner-held (F7) and not counted. → **Accepted risk; plan-sized if pursued** (needs a failure-backoff or `checked_at` design choice).
- **F6 — patch deviation note wrong** — the superseded vetting Outcome does contain the "one product-page fetch per listing" text (line-wrapped). Outcome correct. → informational; no code fix.
- **F7 — no `**Rationale promoted:**` line in the patch doc** — promotion itself is supported by the diff except F4. → informational.

### Positive
- 14/14 mutation probes killed — tests pin each predicate, order, boundary and budget term.
- `TestListSoldRowsAddToSweepMissing` added on the executor's own initiative and mutation-checked.
- Strict `stock ≤ 0 && sale_status == 2` reading: drift can only miss sales, never fake them.
- Crawl invariants untouched: re-checks go through each source's existing `Fetcher`; `listing.Scheme`s untouched; owner-held items untouched.

---

## Recommendations

**Immediate:** follow-up patch for F1 (BRAND OFF sold cards), F2 (volume wording), F3 (stale capped-sweep text), F4 (MarkGone hazard comment) — all offline, fixture-backed.
**Future:** F5 if re-check failures are seen in a real crawl log (`details done … failed`).
**Process:** plan Assumptions marked "verify early" need a recorded check in the patch doc (here: a fixture scan of each shop's list page for sold markers).

---

## Final Verdict

**Status:** Accept with Conditions

**Overall Assessment:**
- Completeness: High
- Correctness: High
- Quality: High
- Value: Medium

**Conditions:**
- [ ] F1 fixed by follow-up patch, or recorded in the sprint as an open gap, before finalize judges the ≥3-providers sold-observation clause
- [ ] F2 wording fixed before the owner's 第18条 (12) reading relies on the vetting Outcome's request volume

**Sign-off:** Yes — the planned work is correct, budget-safe, well tested and as documented; conditions concern a pre-existing sibling gap and doc precision.
