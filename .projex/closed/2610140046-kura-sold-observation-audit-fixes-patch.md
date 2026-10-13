# Patch: Kura Sold Observation Audit Fixes (F1–F4)

> **Sprint:** 2610131822-kura-def-coverage-sprint.md — iteration 5
> **Status:** Complete
> **Author:** Claude Opus 5.5 (patch, sprint subagent, uid `2610131822-kura-def-coverage#5`)
> **Parent:** 2610140037-kura-sold-observation-audit.md
> **Directive:** `/patch-projex.md` fix audit findings F1–F4 from `2610140037-kura-sold-observation-audit.md` (patch step after audit, uid `2610131822-kura-def-coverage#5`). Out of scope: F5 (design decision), F6/F7 (informational), owner-held decisions. No requests to any shop host; fixtures only.
> **Source Plan:** Direct (audit findings; audited plan 2610140010-kura-sold-observation-plan.md stays Complete)
> **Related Projex:** 2610140037-kura-sold-observation-audit.md (findings source) | 2610140023-kura-sold-observation-patch.md (audited execution) | 2610140010-kura-sold-observation-plan.md (Assumption 1 falsified by F1) | 2610131822-kura-def-coverage-sprint.md (not edited)
> **Result:** Success
> **Rationale promoted:** F1 sold-tile + row-count rationale → `brandoff.ListPage.Sold`, `brandoff.Source.Sweep` comments and README § Knowing when something sold | F2 volume → `crawl.Crawler.RecheckAfter`, README, vetting | F4 flap hazard → `store.MarkGone` comment

---

## Summary

| Finding | Outcome | Commit |
|---|---|---|
| F1 BRAND OFF `SOLDOUT` tiles stored as active listings | fixed | `61299b0` (code + tests), `78831b4` (README) |
| F2 re-check request-volume wording | fixed | `78831b4` |
| F3 stale capped-sweep claims | fixed | `78831b4` |
| F4 unpromoted `MarkGone` flap hazard | fixed | `78831b4` |

Re-verified against code + fixture before acting: `list_lv.html` holds exactly 7 `.product__item--soldout` tiles (`SOLDOUT` ×7) among 90 tiles; pids match the audit's list (6 bags cat `100xxx` + `2121400023012` non-bag). `brandoff.ParseList` had no soldout check; `Sweep` passed `nil`. No other `brandoff.ParseList` / `ListPage` consumer outside the package.

One hazard beyond the auditor's suggested fix, handled: moving sold tiles out of `Cards` makes a full 90-tile page look short (83) to `Sweep`'s last-page test → sweep reports complete after page 1 → `SweepMissing` counts every listing past page 1 missed (would mark them gone after two crawls). Fix counts `len(Cards)+len(Sold)` in both the `< pageSize` and `>= total` tests; a new Sweep test pins it.

---

## Changes

### F1 — BRAND OFF sold tiles (`61299b0`)

**File:** `internal/source/brandoff/parse.go`
**Change Type:** Modified
**What Changed:**
- `ListPage.Sold []Card` added (L39–44) with rationale: rejected keeping sold tiles in `Cards` (shown 在售, active comps, never missing for `SweepMissing`).
- `ParseList`: tile with `.product__item--soldout` → `page.Sold`, else `page.Cards` (L108–111). Tile still fully parsed so `Sweep` can filter by `IsBag`.

**File:** `internal/source/brandoff/source.go`
**Change Type:** Modified
**What Changed:**
- `baseURL` const → var (L19–21), test-only reason stated (same pattern as `allu.baseURL`).
- `Sweep` doc comment (L52–57): sold tiles → sold ids (bags only — only bags held); row count includes sold tiles, with the rejected alternative (cards-only count → early complete → `SweepMissing` false misses). Fixes the audit's rule-6 gap on `brandoff.Source.Sweep`.
- `Sweep` body (L95–105): bag sold ids → `visit(page, bags, sold)`; `rows := len(lp.Cards)+len(lp.Sold)` replaces `len(lp.Cards)` in the last-page test.

**File:** `internal/source/brandoff/parse_test.go`
**Change Type:** Modified
**What Changed:**
- `TestParseList`: 90 cards → 83 cards + 7 sold (L32–35).
- `TestParseListSold` (new, L83–111): the 7 fixture ids, `IsBag` per id (6 bags, `2121400023012` not), no sold id among `Cards`.

**File:** `internal/source/brandoff/source_test.go`
**Change Type:** Created
**What Changed:**
- `TestSweepSoldTiles`: httptest serves `list_lv.html` as page 1, `list_empty.html` past it; asserts pages `[1, 2]`, complete, sold == the 6 bag ids, no id both listed and sold.

**Why:** shop's own `SOLDOUT` label in the saved fixture is the evidence; `crawl.Source.Sweep` contract ("nil when the shop's list pages show none") was false for BRAND OFF and is now true. Held sold bags go gone via `SweepTarget`'s existing sold loop (`store.MarkGone`, active-only — repeats are no-ops, no flap). No new requests: same list pages read.

### F2 — request-volume wording (`78831b4`)

**Files:** `internal/crawl/crawler.go` (`Crawler.RecheckAfter`, L54–57) | `README.md` § Knowing when something sold (L143–148) | `docs/provider-vetting.md` GR crawl limits (L215–217), ALLU Outcome bullet (L352–354)
**Change Type:** Modified
**What Changed:**
- Dropped "no more product-page requests than before" / "(no added requests)" (rule 3 + inaccurate).
- Now states cap + steady state: stays within `-details` per source per crawl; while re-check candidates exist a crawl can spend the whole cap (150 at default, up to 600/day at `-crawl-every 6h`), not only new listings' pages. ALLU Outcome bullet gets the same volume line (owner's 第18条 (12) reading reads it).

**Why:** cap unchanged; actual volume rises toward the cap. Wording only — no budget change.

### F3 — stale capped-sweep claims (`78831b4`)

**Files:** `cmd/kura/main.go` `-max-pages` help (L60) | `README.md` flags row (L238), § Knowing when something sold (L130–131, L136–137) | `internal/crawl/crawler.go` `Crawler.MaxPages` (L36–39)
**Change Type:** Modified
**What Changed:**
- "capped sweeps never mark items sold" → "capped sweeps never count unseen items toward sold" (help + flags row).
- README: capped sweeps never count a missing item toward gone "(the sold signals below still apply)"; never-complete sweeps: "missing from a sweep never marks their items sold. Other signals do:" (was "sweeps never mark their items sold. Two other signals do:" — now three signals).
- `Crawler.MaxPages` comment carried the same stale claim (not named by the audit, same finding): now says capped sweeps skip `SweepMissing`; list-page sold ids and re-checks still mark gone.

**Why:** capped ALLU sweeps mark sold rows gone, capped targets are re-checked, and BRAND OFF capped sweeps now pass sold tiles.

### F1 docs + F4 hazard (`78831b4`)

**File:** `README.md` § Knowing when something sold (L152–155)
**What Changed:** new bullet "SOLDOUT tiles on BRAND OFF list pages" — marks held listing gone, never stored, counts toward page size.

**File:** `internal/store/store.go` `MarkGone` doc (L258–264)
**What Changed:** hazard promoted from 2610140023-kura-sold-observation-patch.md: first-read-gone listing that a list page keeps showing as for sale flaps (RecordSeen reactivates → PendingDetails refetches → gone_at moves forward, one first fetch per crawl). States that adapters route shop-marked sold tiles to sold ids (removes the known trigger) and that an unmarked sold item still flaps.

**Why:** F4 lived only in a projex doc (rule 1 promotion; rule 6 don't-fix trap).

---

## Verification

**Method:** `go vet ./...`; `go test -count=1 ./...`; `go test -count=1 -v ./internal/source/brandoff/`; gofmt on committed content of all 44 tracked `.go` files (`git show HEAD:<f> | gofmt -l`, worktree is CRLF/autocrlf); 4 mutation probes on a scratch copy outside the repo (removed after).

**Result:**
```
go vet ./...            → exit 0
go test -count=1 ./...  → exit 0
ok  github.com/No3371/kura/cmd/kura
ok  github.com/No3371/kura/internal/advisor
ok  github.com/No3371/kura/internal/crawl
?   github.com/No3371/kura/internal/fx  [no test files]
ok  github.com/No3371/kura/internal/listing
ok  github.com/No3371/kura/internal/radar
ok  github.com/No3371/kura/internal/source/allu
ok  github.com/No3371/kura/internal/source/brandoff
ok  github.com/No3371/kura/internal/source/grare
ok  github.com/No3371/kura/internal/store
ok  github.com/No3371/kura/internal/web
brandoff -v: TestParseList | TestParseListSold | TestParseListPastEnd | TestApplyTitle | TestParseDetail | TestSweepSoldTiles — PASS
gofmt (committed HEAD): 44/44 clean

Mutations (all killed):
M1 rows := len(lp.Cards)            → FAIL TestSweepSoldTiles (stops after page 1)
M2 visit(page, bags, nil)           → FAIL TestSweepSoldTiles
M3 bag filter on sold dropped       → FAIL TestSweepSoldTiles (non-bag id passed)
M4 soldout split disabled           → FAIL TestParseList, TestParseListSold, TestSweepSoldTiles
```

**Status:** PASS

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|----------|-------------|-------------|
| 2610140037-kura-sold-observation-audit.md | Parent (findings F1–F4) | none — audit is a point-in-time record; this patch answers it |
| 2610140010-kura-sold-observation-plan.md | Audited plan; Assumption 1 falsified by F1 | none — closed record; F1 recorded here |
| 2610131822-kura-def-coverage-sprint.md | Sprint | not edited (orchestrator's absorb owns nav). For absorb: BRAND OFF sold tiles now observed → all 3 providers have a sold signal (BRAND OFF: complete-sweep `SweepMissing` + SOLDOUT tiles + re-checks when capped/failed; GR/ALLU: re-checks; ALLU: sold rows). Audit condition 1 (F1) and condition 2 (F2) met. |

---

## Notes

- Constraints held: crawl invariants unchanged; no new request paths (sold tiles come from list pages already read); `-details` budget and `RecheckAfter` logic untouched; each provider's `listing.Scheme` untouched; no request to any shop host.
- Not done (out of scope): F5 (re-check failure head-of-line — needs failure-backoff / `checked_at` design), F6/F7 (informational), owner-held items (ALLU consignments, 第18条 (12) reading, stop-on-403, GALLERY RARE paging).
- Unknown offline: whether BRAND OFF's product page for a SOLDOUT item 404s or still renders — `brandoff.Source.Detail` unchanged; the list tile alone now marks it gone.
