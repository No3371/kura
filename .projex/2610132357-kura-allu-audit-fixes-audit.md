# Audit: Kura ALLU — Audit Fixes Patch (F1/F2/F3 doc parts, F5, F6, F8, F9, F10)

> **Sprint:** 2610131822-kura-def-coverage-sprint.md — iteration 4
> **Status:** Complete (Accepted)
> **Author:** Claude Opus 5.5 (auditor, sprint subagent; re-audit after `[patch]`, uid `2610131822-kura-def-coverage#4`)
> **Parent:** 2610132349-kura-allu-audit-fixes-patch.md
> **Related Projex:** 2610132349-kura-allu-audit-fixes-patch.md (audited patch) | 2610132339-kura-allu-provider-3-audit.md (findings source F1–F10) | 2610132325-kura-allu-provider-3-patch.md (original ALLU execution) | 2610131822-kura-def-coverage-sprint.md (not edited)
> **Audit Date:** 2026-10-13 | **Work Period:** 2026-10-13, range `10fe0ad..a5d9ca2` on `projex/sprint/2610131822-kura-def-coverage` (`5a8a71f` code + tests, `ecadf0b` README + vetting, `a5d9ca2` patch doc + parent-audit conditions)
> **Subject:** `internal/source/allu` (`payload.go`, `parse.go`, `source.go`, tests) | `README.md` | `docs/provider-vetting.md` § ALLU + Outcome | parent audit's Conditions edit
> **Scope:** In — F1/F2/F3 documentation parts; F5, F6, F8, F9, F10 in full. Out (not counted against the patch) — F2 code fallback, F4, F7, any consignment keep/exclude change, any vetting verdict change.
> **Method:** No request to any ALLU host. Tree of `a5d9ca2` exported (`git archive`) to the session scratchpad; `go vet ./...` + `go test -count=1 ./...` + `-v` on `internal/source/allu` there; a throwaway scratch test decoded `testdata/list_lv.html` and the recorded HERMES probe page with the package's own decoder (deleted after). `gofmt -l` on the worktree (read-only). Recorded ToS copy and product probe from the earlier ALLU rounds re-read offline; ungrouped counts re-derived from a copy of the executor's scratch DB. Worktree left clean.

---

## Audit Summary

**Claim:** Docs now say ALLU公式 includes kept partner consignments (R3, Outcome, README ×2, `ParseList` comment) and that 42/167 ALLU rows are ungrouped; (15)/(18) added to the R1 re-vet; devalue `"null"` keys read as strings; seller value drift errors; `Source.Detail` + `Sweep` end paths tested, dropped counter mirrors every filter; top-category rationale; LV 49,388 named largest. Verdicts and consignment behaviour unchanged.

**Verdict:** Verified

**Assessment:** Completeness: High | Correctness: High | Quality: High | Value: High

**Top Issues:**
1. None in scope. Every in-scope finding is resolved as the parent audit prescribed; no regression found.
2. Carried open (out of scope, unchanged): F1 keep vs exclude consignments and F3 第18条 (12) reading (owner); F7 stop-on-403 scope (owner); F4 and F2 code fallback (later plan/patch).
3. Informational: F6 guards the seller value only — `sale_status` / stock value drift still drops silently (patch Notes say so; parent audit's prescribed fix was seller-scoped).

---

## Claims vs Evidence

| Claim | Evidence | Status | Notes |
|-------|----------|--------|-------|
| F5: `"null"` keys literal strings, values via `ref`; test | `payload.go` → `hydrator.typed`; `TestHydrate` | ✓ | Matches devalue's null-prototype encoding (`["null", key, idx, …]`) vs `Map` (`[idx, idx, …]`). Test `["null","a",3,"b",1]` → `{a: 2.5, b: "x"}` pins both halves. No `"null"` tag in any fixture or recorded list/product page (re-checked) → latent fix |
| F6: page with rows but no ALLU公式 non-partner row → error | `ParseList` `sellerOK`; `TestParseListSellerDrift` | ✓ | Premise "every row of a `?seller=1` page is ALLU公式" holds on 60/60 recorded rows (LV fixture 30/30, HERMES probe 30/30, all `ALLU公式`, `is_allu_partner false`). Sold rows count as seller-OK → a fully sold page stays error-free (tested). Empty page skipped by `Rows > 0` → `Sweep`'s own empty-page / complete logic unchanged (tested) |
| F6: error stops brand sweep, `RunAll` continues, fewer requests | `crawl.Crawler.RunAll` / `SweepTarget` | ✓ | Error → `run.Complete = false` → no `SweepMissing`; run recorded with error; `RunAll` logs and moves to next target. No path issues more requests |
| F8: `Detail` 200 bag / 200 sold / 404 / 403 / 202 / 500; `Sweep` complete + empty page; dropped counter all classes | `source_test.go` `serveBase`, `TestSweepComplete`, `TestSweepEmptyPage`, `TestDetail`; `TestParseList` | ✓ | All six `Detail` branches asserted; 202 reaches `Detail`'s own non-200 branch (fetcher passes 2xx). "500 last because backoff still scheduled" true — `Fetcher.Get` calls `h.schedule(backoff)` after the final retryable failure. `serveBase` swaps package var `baseURL` with cleanup; no `t.Parallel` in package → no race. robots.txt fetched from the httptest server (404 → allow) → no outside request |
| F9: top-category rationale, rejected alternative named | `ParseList` comment above `TopCategory` check | ✓ | Names constraint (request scope unverifiable from rows), consequence (wallet joins bag comparables), rejected alternative (trust URL) — rule 6 satisfied |
| F10: LV largest, 49,388 | `maxListPages` comment; vetting R5; fixture decode | ✓ | `list_lv.html` key `market:searchItems:/jp/ja/market/b-21/c-58/…?seller=1&page=1`, `total_count` 49,388 (>1,646 pages @30); HERMES probe 22,135 ✓ |
| F1 doc: R3 + Outcome + README ×2 + comment state consignments kept | diffs; recorded ToS 第四章; product probe; fixtures | ✓ | 第四章 heading 「当社が委託アイテムを出品する場合の特則」 verbatim; "individual page states it is a consignment" and "inspection after purchase when not in ALLU's control" match ToS text. Probe page contains `consigned_item`; both product fixtures `owned_inventory_item`. `apiAsset` has no consignment field → "Kura does not read them" true. No stale "excluded"/"ALLU-held" claim left in shipped files (grep) except inside the explicit Correction note |
| F1: keep/exclude behaviour unchanged | `ParseList` seller check diff | ✓ | Condition `Name != official \|\| IsAlluPartner` untouched; comment-only change there |
| F2 doc: 42/167 (25%); CHANEL 20/27, GOYARD 11/29, CELINE 9/27, LOEWE 2/28, HERMES/LV 0; BRAND OFF 0.7% | scratch DB copy, SQL on `group_key = ''` | ✓ | Reproduced exactly; BRAND OFF 101/15,367 = 0.66%. "code-shaped title token stands in when 型番 is empty" matches `modelCode` |
| F3 doc: (15), (18) quoted from recorded ToS; Pass unchanged | vetting R1 re-vet; recorded `tos.html` | ✓ | (15) full text 「当初から転売を目的としてアイテムを購入するなど、営利を目的として本サービスを利用する行為」 verbatim; (18) 「その他、当社が不適切と判断する行為」 verbatim. (18) reading honest (no advance settlement; ALLU objection ends Pass). R1 heading still Pass; (12) reading untouched |
| `go vet`, `go test` green; 19 allu tests | own run on exported tree | ✓ | All packages ok; `-v` → 19 PASS incl. 4 new. `gofmt -l internal/source/allu/` on worktree clean (exported copy reports CRLF only — `core.autocrlf=true` on archive, not a source issue) |
| Parent audit: Conditions split, Partial Execution line | `a5d9ca2` diff | ✓ | Doc parts `[x] [PATCHED]`; owner decisions left `[ ]`; F7/F4/F2-code listed open |
| Patch `file:ln` locators | `sed -n` per locator | ✓ | All land on the named code; `payload_test.go` L57–63 ends one line past the block (blank) — immaterial |
| Out-of-scope items untouched | diff file list | ✓ | No `crawl`, `store`, `cmd/kura`, `available`/`-sources`, or verdict line changed |

---

## Objective Verification

### F5 — devalue null-prototype keys

**Evidence:** `payload.go` → `hydrator.typed` `"null"` case; `TestHydrate`.

**Findings:**
- Actual: `Map` and `"null"` split; `"null"` reads `rest[k]` as string; comment states why (string keys vs Map indices; collapse under `<nil>`).
- Missing: nothing. A non-string key would map to `""` silently — devalue never emits one, so no case to handle.
- Quality: High

**Verification:** ✓ Verified

### F6 — seller value drift

**Evidence:** `ParseList`; `TestParseListSellerDrift`; crawler call path.

**Findings:**
- Actual: counter placed immediately after the seller check, before every other filter → only seller drift trips it; sold, J, priced-range and non-bag rows still drop silently as before. Error text `allu: page N: none of M rows has seller ALLU公式 (seller values changed?)`.
- Missing: `sale_status` / stock drift still silent — stated in patch Notes; legitimately ambiguous with a sold-out page; not in the parent audit's prescribed fix.
- Quality: High

**Verification:** ✓ Verified

### F8 — test gaps

**Evidence:** `source_test.go`, `parse_test.go`; `go test -v`.

**Findings:**
- Actual: every branch the parent audit listed is now asserted; `TestParseList` counter adds `!graded` and `DisplayPriceMin <= 0`, so it mirrors `ParseList` exactly (seller | stock/status | top cat | J | graded | price).
- Missing: none in scope (consignment payload fixture still absent — tied to out-of-scope F1 behaviour).
- Quality: High

**Verification:** ✓ Verified

### F9 — rule-6 gap

**Evidence:** `ParseList` comment. **Verification:** ✓ Verified

### F10 — largest brand

**Evidence:** `maxListPages` comment; vetting R5; fixture decode. **Verification:** ✓ Verified

### F1 doc part — consignments kept

**Evidence:** vetting R3 + Outcome; README § Choosing a source + Limitations; `ParseList` seller comment; ToS copy; probe; fixtures.

**Findings:**
- Actual: record now matches code: kept, stored as ALLU公式 under ALLU's scheme, 店家鑑定; Correction note owns the earlier misreading; owner decision recorded open, Pass stands meanwhile.
- Missing: request-log row #5 still lists `is_consignment` true without `item_type_key` — a point-in-time log row, and R3 now carries the type key; not a defect.
- Quality: High

**Verification:** ✓ Verified

### F2 doc part — ungrouped share

**Evidence:** README § Turning messy listings + Limitations; scratch DB. **Verification:** ✓ Verified

### F3 doc part — (15), (18)

**Evidence:** vetting R1 re-vet; ToS copy. **Verification:** ✓ Verified

---

## Code/Implementation Inspection

### `internal/source/allu` → `payload.go`

**Claimed:** `"null"` keys as strings. **Actual:** as claimed — Quality: High
**Issues Found:** none
**Undocumented:** none

### `internal/source/allu` → `parse.go`

**Claimed:** `sellerOK` error; F1 + F9 comments; `fmt` import. **Actual:** as claimed — Quality: High
**Issues Found:** none. Filter semantics unchanged for any page with ≥1 ALLU公式 row — the diff adds only a counter and a post-loop check, no filter condition changed (HEAD: LV fixture 29 cards, HERMES probe 27).
**Undocumented:** none

### `internal/source/allu` → `source.go`

**Claimed:** `maxListPages` comment only. **Actual:** comment only; `maxListPages = 10` unchanged — Quality: High

### Tests

**Claimed:** 4 new tests + counter. **Actual:** as claimed; `serveBase` reusable, cleanup restores `baseURL` — Quality: High

**Metrics:**
- Readability/Maintainability: High
- Test Coverage: `Source.Detail` and `Sweep` end paths now covered; decoder `"null"` covered
- Technical Debt: None added

---

## Testing Validation

**Coverage:** decoder incl. `"null"` | list fixture + filter table + seller drift (3 cases) | detail fixtures | `Sweep` cap / complete / empty page / unknown brand | `Detail` 6 status branches.
**Execution:** All pass? Yes (`go test -count=1 ./...`, 11 packages ok, `fx` no tests) | Flaky tests: none seen (`TestDetail` 0.01 s — the 500 case's backoff never waited on)
**Missing:** consignment product payload (no fixture; belongs with the F1 owner decision).
**Quality Issues:** none.

---

## Documentation Audit

**Completeness:** README — Complete for F1/F2 disclosure | Vetting record — Complete for F1/F3/F10 | Migration/Rollback n/a.
**Accuracy:** Matches implementation? Yes | Quotes verbatim? Yes (第四章 heading, (15), (18)) | Numbers reproduce? Yes (42/167, per-brand, 0.7%, 49,388, 22,135).
**Quality:** Clarity High | Completeness High | Usability High

---

## Source Hygiene

**Scope:** Diff only (`10fe0ad..a5d9ca2`, excluding `**/.projex/**`)

| Rule | Location (symbol) | Quote | Severity |
|------|-------------------|-------|----------|
| — | — | No rule 1–5 violation: no projex ID/finding label (F1…F10) in comments, no `file:ln`, present tense, no banners, no bare reassurance | — |

**Rule 6 gaps:** none — the one non-obvious change (`sellerOK` error) carries consequence + rejected alternative; `"null"` case states the trap.

**Rule 6 vacuity:** none — `hydrator.typed` `"null"`, `ParseList` seller comment, top-category comment, drift comment, `maxListPages` each name a constraint, trap or rejected alternative. Test comments (`f.Retries = 0`, 500-last ordering) state their constraint.

**Rationale promotion:** none needed — supported by diff; consignment and ungrouped-share rationale shipped in README + `docs/provider-vetting.md`.

**Commit composition** (informational — fix forward, never rewrite history): 2 landed commits outside `.projex/` checked — typed subject 2/2 (`fix(allu)`, `docs(allu)`) | type matches diff 2/2 (`5a8a71f` code + tests; `ecadf0b` README + vetting only) | `Projex:` trailer 2/2 (`2610132349-kura-allu-audit-fixes`; `a5d9ca2` doc-only, also trailed)
**Violations:** None

**Trailer survival on `master`:** 1/10 recent commits carry `Projex:` (`8950a55`, the only projex code landing; rest doc-only / baseline / pre-convention). Channel alive — OK (unchanged since the parent audit).

---

## Gap Analysis

### Promised But Not Delivered
| Promise | Status | Impact |
|---------|--------|--------|
| — (all in-scope items delivered) | — | — |

### Undocumented Issues
| Issue | Severity | Affects |
|-------|----------|---------|
| None found | — | — |

### Unhandled Edge Cases
- `?seller=1` stops filtering server-side and a page holds 0 ALLU公式 rows → now an error for that brand's sweep — Impact: none harmful; fewer requests, run logged with error, no gone-marking. Desired.
- `sale_status` / stock value drift → still silent empty pages — Impact: never-complete brands go stale unsignalled; a completing brand would mark gone after two sweeps. Pre-existing, named in patch Notes; Low, informational.

---

## Quality Assessment

### Completeness: High
**Strengths:** every in-scope finding closed at the depth the parent audit prescribed; doc facts re-derived from evidence, not copied from the audit.
**Gaps:** none in scope.

### Correctness: High
**Works:** `"null"` decode: Yes | seller-drift error without false positives on sold/empty pages: Yes | `Detail` branches: Yes | docs vs code: Yes
**Bugs:** none found

### Code Quality: High
**Positive:** minimal diff; error placed so only the intended drift trips it; tests use httptest + package var swap with cleanup, no network.
**Concerns:** none.
**Tech Debt:** none added.

### Value Delivered: High
**Intended:** record that no longer overstates the seller filter; disclosed coverage gap; latent decoder bug and silent drift closed; untested paths covered.
**Actual:** all of it; owner now has an accurate R3 / R1 to decide F1 and F3 on.
**Impact:** User: Positive (README tells visitors which ALLU rows lack 行情 and that consignments are included) | Business: def ≥3-provider clause rests on a truthful record, pending owner calls.

---

## Open Findings

### Undocumented Discoveries
- Changes: none outside the patch's file list.
- Problems: none inferred.
- Workarounds: patch scaffold landed in `.projex/closed/closed/` and was moved — no stray directory remains (checked).

### Impact Analysis
- Downstream: `crawl` — a seller-drift page now yields a recorded failed run instead of a silent empty one; operators see `sweep failed … seller values changed?` in logs.
- Future: Enabled: F1 owner decision can be implemented against an accurate record | Blocked: nothing.
- Risks: false-positive seller error if ALLU starts mixing partner rows into `?seller=1` pages — Likelihood: Low (60/60 recorded rows clean; failure mode is a skipped brand, not bad data).

### Improvements
- Could be better: extend drift detection to `sale_status` codes (e.g. error when no row on a page has any known status value) — only if a distinguishing signal exists.
- Would make excellent: a consignment product fixture once the F1 decision is made.
- Missed: none.

---

## Stakeholder Impact

| Stakeholder | Promised | Reality | Impact |
|-------------|----------|---------|--------|
| Owner | Truthful R3/R1 to decide F1/F3 | Delivered; decisions recorded open | Positive |
| Demo visitor | Knows which ALLU rows lack 行情; knows consignments included | README states both | Positive |
| ALLU | No change in request volume | No request-path change; drift error can only reduce requests | Neutral |

---

## Findings

### Critical (Must Address)
- None.

### Significant (Should Address)
- None in scope.

### Minor (Nice to Fix)
- **`sale_status` / stock drift still silent** (informational; outside F6's prescribed seller scope, named in patch Notes) → consider only with a signal that separates drift from a sold-out page.
- **Patch locator `payload_test.go` L57–63 one line long** → immaterial; no action.

### Carried open (out of scope — not counted against the patch)
- **F1 keep vs exclude consignments** — owner decision; exclusion is plan-sized.
- **F3 第18条 (12) reading** — owner confirms or overrules.
- **F7 stop-on-403 scope** — owner scope call.
- **F4 list-visible sold rows**, **F2 title-derived model-name fallback** — later plan / patch.

### Positive
- Doc numbers and quotes reproduce exactly from evidence (42/167 and per-brand split, 0.66%, 49,388, 22,135, 第四章 heading, (15), (18)).
- F6 check placed so sold-out pages stay legitimate; premise backed by every recorded `?seller=1` row.
- `TestDetail` covers all six status branches and documents the fetcher's post-final-failure backoff.
- Correction note in R3 owns the earlier misreading instead of silently rewriting it.

---

## Recommendations

**Immediate:** none for code. Surface F1 (keep/exclude), F3 ((12) reading), F7 (scope) to the owner, unchanged.
**Future:** F4 and F2 code fallback with remainder (d) / a follow-up patch; consignment fixture with any F1 exclusion work.
**Process:** none — the patch re-derived each doc fact from evidence rather than copying the audit's numbers; keep that.

---

## Final Verdict

**Status:** Accept

**Overall Assessment:**
- Completeness: High
- Correctness: High
- Quality: High
- Value: High

**Conditions:**
- [x] In-scope findings F1/F2/F3 doc parts, F5, F6, F8, F9, F10 resolved without regression
- [ ] (carried, owner) F1 keep vs exclude consignments
- [ ] (carried, owner) F3 第18条 (12) reading
- [ ] (carried, owner) F7 scope

**Sign-off:** Yes — 2610132349-kura-allu-audit-fixes-patch.md resolves every in-scope finding of 2610132339-kura-allu-provider-3-audit.md as prescribed; vet, tests and gofmt clean; no regression in filter semantics, request behaviour or verdicts. Remaining items are owner decisions and later-plan work explicitly outside this patch.
