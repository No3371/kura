# Patch: Kura multi-source crawl + per-pair onboarding epoch

> **Status:** Complete
> **Sprint:** 2610131822-kura-def-coverage-sprint.md — iteration 1
> **Author:** Claude Opus 5.5 (patch, sprint subagent)
> **Parent:** 2610131831-kura-multi-source-onboarding-epoch-plan.md
> **Directive:** Orchestrator: run patch in plan-execution mode on 2610131831-kura-multi-source-onboarding-epoch-plan.md (sprint `patch!` step, uid `2610131822-kura-def-coverage#1`).
> **Source Plan:** 2610131831-kura-multi-source-onboarding-epoch-plan.md (Steps 1–5, full)
> **Result:** Success
> **Related Projex:** 2610131831-kura-multi-source-onboarding-epoch-plan.md (executed) | 2610131822-kura-def-coverage-sprint.md | 2610130449-kura-mock-def.md (最新上架 provisional line annotated) | 2610131037-kura-member-notifications-audit.md (R1 origin) | 2610130420-kura-roadmap-nav.md (Known Blockers R1 — now resolved, not edited)
> **Commits:** `072c15d` feat(crawl) — all code, tests, README

---

## Summary

Plan executed as written, one code commit. (1) `crawl.Crawler.Source` → `Sources []Source`: `RunAll` sweeps every target of every source (failure isolated, logged), then per-source detail budget; `cmd/kura` `setup` builds the source list (BRAND OFF only), rejects unregistered-scheme / duplicate names, backfills `listing.Schemes()...`. (2) R1 fix: watched = `first_seen` strictly after own (source, target) first complete sweep `finished_at` (`store.Onboarded`), replacing global `epoch + 24 h`. `newer` (最新上架) + `stats` (days-to-sell) read one per-item bit set in `Engine.build`. `Overview.Since` / `SoldStatsReady` stay on global epoch (out of scope, deferred).

No deviations from plan code. Additions beyond plan text: README fetcher bullet (l.63) `-details` → "per source per run" (same stale phrase as flags table); `newer` doc comment re-wrapped.

---

## Changes

### Store — onboarding + version (Step 1)

**File:** `internal/store/store.go` — Modified
- `Version` (l.288): + `(SELECT COUNT(*) FROM crawl_runs WHERE complete=1)` → format `n/last/pts/runs`; doc: changes when sweep completes.
- + `Pair` (l.349), + `Onboarded` (l.354): `MIN(finished_at)` over `complete=1 AND finished_at IS NOT NULL` grouped by source, target.

**File:** `internal/store/store_test.go` — Modified
- + `TestOnboarded` (l.183): empty → incomplete ignored + version unchanged → complete day 2 recorded + version bumps → later complete keeps MIN → unfinished (`StartRun` only) ignored → other source separate key.

### Radar — per-pair watched (Step 2)

**File:** `internal/radar/radar.go` — Modified
- `Item` + unexported `watched bool` (l.128); no JSON change.
- `Refresh` (l.187) loads `Store.Onboarded`; `build` (l.212) takes `onboarded`, sets `it.watched = watched(r, onboarded)` (l.221).
- `watched` (l.509): plain func over `store.Row` + map; comment names rejected global epoch + fail-closed no-baseline rule.
- `newer` (l.525) reads `a.watched`/`b.watched`; `stats` (l.703) `it.watched`. `epoch` field comment (l.175), `MedianDaysToSell` comment (l.158) updated.

**File:** `internal/radar/radar_test.go` — Modified
- `engine` (l.38) → `engineAt` (l.42) with `{Source: "t"}: t0+24h` — reproduces old boundary; existing assertions untouched.
- + `TestWatchedPerSourceOnboarding` (l.287): newest = `u:w, t:w, t:i, u:i1, u:i2, u:x`; `NewestActive` = `u:w`; `MedianDaysToSell` = 3.

### Crawler — several sources (Step 3)

**File:** `internal/crawl/crawler.go` — Modified
- `Sources []Source` (l.30) + sequential-sweep rationale; `DetailBudget` doc per source.
- `RunAll` (l.51): nested source/target loop, then `FetchDetails` per source. `SweepTarget(ctx, src, target)` (l.71), `FetchDetails(ctx, src)` (l.132); logs + error text carry source.

**File:** `internal/crawl/crawler_test.go` — Created
- `fakeSource` (l.19); `TestRunAllSweepsEverySource` (l.70): 3 runs (fa/CHANEL failed, fb still runs), listings both sources, details own-source only, `Onboarded` = fa/HERMES + fb/HERMES. `TestDetailBudgetPerSource` (l.140): 2 + 2.

### Wiring (Step 4)

**File:** `cmd/kura/main.go` — Modified
- `-details` help (l.57); `BackfillGrades(ctx, listing.Schemes()...)` (l.117); `sources` list (l.134) + `SchemeFor` check (l.142) + duplicate check (l.148); `crawlOnce` comment (l.156).

### Web fixture + README (Step 5)

**File:** `internal/web/server_test.go` — Modified
- `TestMemberAPI`: `onboard` (l.265) records complete `wm` runs for LOUIS VUITTON + HERMES finishing t0+30m; t0 comment (l.234). Without it the test failed as expected (M1 watchlist + `new` event missing) — confirms fixture is load-bearing.

**File:** `README.md` — Modified
- Fetcher bullet (l.63) + flags table (l.168): `-details` per source. *Knowing when something sold* (l.94) + *Newest listings* (l.99): first complete sweep of shop's brand; later-added shop/brand + no-complete-sweep sentence.

---

## Verification

**Method:** `go vet ./...`; `go test -count=1 ./...`; gofmt on LF-normalised copies of changed `.go` (clean); hygiene grep (`projex|\d{10}-|Step \d|R1|F1`) on diff + new test (clean). Manual: `sqlite3 -readonly kura.db ".backup <scratch>"` (original db/wal mtimes unchanged after); baseline binary from `git archive 526ada6`, new binary from work-root; each served its own db copy on 127.0.0.1:8091 / 8092 with `-advisor mock` (no crawl flags), full `sort=newest` order paged (limit 100) + plan spot checks; both stopped by PID. No live crawl.

**Result:**
```
ok  internal/advisor | internal/crawl | internal/listing | internal/radar
ok  internal/source/brandoff | internal/store | internal/web
total base/new 15367 15367
base first5 [2100301255937 2101219731339 2101219811840 2101219850221 2101220005108]
new  first5 [2101220126889 2100301255937 2101219731339 2101219811840 2101219850221]
base last2  [2101218806076 2101220126889]
new  last2  [2101218805895 2101218806076]
R1 position base 15367 -> new 1
new == [R1] + base-minus-R1: True
since 2026-10-13T00:25:25+08:00 (both) | sold_stats_ready false (both) | active 15367 (both)
```
(ids prefixed `brandoff:`)

**Status:** PASS

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|----------|-------------|-------------|
| 2610131831-kura-multi-source-onboarding-epoch-plan.md | Source plan | Steps 1–5 `[PATCHED]`, criteria + verification checked, Status Complete, moved to `closed/` |
| 2610130449-kura-mock-def.md | Def (最新上架 provisional line) | R1 marked fixed with rule + this patch |
| 2610130420-kura-roadmap-nav.md | Known Blockers R1, Phase 6 prerequisite | Not edited (nav not referenced by this patch) — orchestrator: R1 resolved |
| 2610131822-kura-def-coverage-sprint.md | Sprint nav | Not edited (orchestrator-owned) |

---

## Notes

- Fail-closed: pair without a complete sweep → nothing watched (capped `-max-pages`-only runs lose 新上架 / days-to-sell). Sprint step (e) scheduler must keep periodic full sweeps.
- Plan's "Every listing first seen ≤ its pair's first complete `finished_at` except `brandoff:2101220126889`" confirmed on the copy, incl. HERMES second complete sweep (run 8): full-order diff shows R1 as the only move.
- Line endings: repo `core.autocrlf=true`; working copies kept CRLF to match.
