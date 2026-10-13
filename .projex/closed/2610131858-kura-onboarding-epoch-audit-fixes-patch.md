# Patch: Kura onboarding-epoch audit fixes (F1–F4)

> **Status:** Complete
> **Sprint:** 2610131822-kura-def-coverage-sprint.md — iteration 1
> **Author:** Claude Opus 5.5 (patcher, sprint subagent)
> **Parent:** 2610131852-kura-multi-source-onboarding-epoch-audit.md
> **Directive:** Address findings F1–F4 of 2610131852-kura-multi-source-onboarding-epoch-audit.md (§ Findings → Minor). I1–I3 out of scope.
> **Source Plan:** Direct (audit findings; audited plan 2610131831-kura-multi-source-onboarding-epoch-plan.md, executed by 2610131843-kura-multi-source-onboarding-epoch-patch.md)
> **Related Projex:** 2610131852-kura-multi-source-onboarding-epoch-audit.md (findings source; F1–F4 marked) | 2610131831-kura-multi-source-onboarding-epoch-plan.md | 2610131843-kura-multi-source-onboarding-epoch-patch.md | 2610131822-kura-def-coverage-sprint.md (not edited)
> **Result:** Success
> **Commit:** `3aa09fa` (code + README) on `projex/sprint/2610131822-kura-def-coverage`

---

## Summary

4 minor audit findings closed in one code commit: strict-`After` onboarding boundary now asserted (kills surviving mutant M3); page-shift hazard of first complete sweep stated in `watched` comment + README (overclaim removed); per-source detail-budget rationale promoted to `Crawler.DetailBudget` comment + README; `setup` source-list validation extracted to `checkSources` with table test. No behavior change.

---

## Changes

### F1 — onboarding boundary test

**File:** `internal/radar/radar_test.go`
**Change Type:** Modified
**What Changed:**
- `TestWatchedPerSourceOnboarding`: new row `u:ib` with `FirstSeen` == `{u, LV}` onboarding time (`day(10)+1h`), commented (ln 310–312)
- Expected newest order gains `u:ib` among initial-crawl rows after `u:i2` (ln 324)

**Why:** strict `After` (equal second → not watched) was relied on but unasserted; mutant `!FirstSeen.Before(at)` survived.

---

### F2 — page-shift overclaim

**File:** `internal/radar/radar.go`
**Change Type:** Modified
**What Changed:**
- `watched` comment: `Hazard:` paragraph — listing pushed up past already-read pages during the first complete sweep is first seen next sweep and reads watched; rejected grace period after onboarding (would hide listings genuinely posted then) (ln 509–513)

**File:** `README.md`
**Change Type:** Modified
**What Changed:**
- § Newest listings: "existing stock never reads as new" → "does not flood the order as new" + "One gap remains" sentence naming the page-shift miss (ln 105–108)

**Why:** README overstated; code lacked the don't-fix note (Rule 6).

---

### F3 — per-source detail budget rationale

**File:** `internal/crawl/crawler.go`
**Change Type:** Modified
**What Changed:**
- `Crawler.DetailBudget` comment: rejected shared budget (first-crawl backlog starves other shops' details); total fetches scale N × `DetailBudget` (ln 38–41)

**File:** `README.md`
**Change Type:** Modified
**What Changed:**
- Fetcher bullet `-details`: cap per source, not shared, + fetch total grows with shop count (ln 63–65)

**Why:** rationale lived only in plan Step 3 → promoted to code + shipped doc.

---

### F4 — `setup` refusals tested

**File:** `cmd/kura/main.go`
**Change Type:** Modified
**What Changed:**
- Scheme + duplicate-name loop moved from `setup` into `checkSources(sources []crawl.Source) error` (ln 144–162); `setup` calls it (ln 137–139). Comments, messages, call position unchanged.

**File:** `cmd/kura/main_test.go`
**Change Type:** Created
**What Changed:**
- `TestCheckSources` table: brandoff alone ✓ | none ✓ | unregistered `nosuch` → scheme error | second `brandoff` → duplicate error. Stub `named` embeds `crawl.Source`, answers only `Name`.

**Why:** both refusal paths had no test (`cmd/kura` had none).

---

## Verification

**Method:** go1.26.0 in worktree — `go build ./...` | `go vet ./...` | `go test -count=1 ./...`; gofmt on LF copies of 5 changed `.go` (checkout is CRLF, `core.autocrlf=true`); 3 mutants on scratch copy (removed); hygiene grep on diff `+` lines (`projex`, `\d{10}-`, `Step \d`, `R1`, `F\d`, `:\d+`, banners, deliberate/by design, changelog words).

**Result:**
```
ok  	github.com/No3371/kura/cmd/kura
ok  	github.com/No3371/kura/internal/advisor
ok  	github.com/No3371/kura/internal/crawl
?   	github.com/No3371/kura/internal/fx	[no test files]
ok  	github.com/No3371/kura/internal/listing
ok  	github.com/No3371/kura/internal/radar
ok  	github.com/No3371/kura/internal/source/brandoff
ok  	github.com/No3371/kura/internal/store
ok  	github.com/No3371/kura/internal/web

M3  (watched: !FirstSeen.Before(at)):
    radar_test.go:325: newest = [u:w u:ib t:w t:i u:i1 u:i2 u:x], want [u:w t:w t:i u:i1 u:i2 u:ib u:x]  → killed
MS1 (checkSources: scheme check off):
    main_test.go:35: no scheme: error = <nil>, want "source \"nosuch\" has no registered condition scheme"  → killed
MS2 (checkSources: duplicate check off):
    main_test.go:35: duplicate name: error = <nil>, want "duplicate source \"brandoff\""  → killed
control (unmutated): ok cmd/kura, ok internal/radar
gofmt -l (LF): none | hygiene grep: no hits
```

**Status:** PASS

No live crawl; `kura.db` untouched (no DB access). Worktree clean after commit.

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|----------|-------------|-------------|
| 2610131852-kura-multi-source-onboarding-epoch-audit.md | Parent; findings source | F1–F4 marked `[PATCHED]` → this doc; Related Projex link added |
| 2610131831-kura-multi-source-onboarding-epoch-plan.md | Audited plan (closed) | None — patch does not execute plan objectives; I3 (header bold key) out of scope |
| 2610131822-kura-def-coverage-sprint.md | Sprint nav | Not edited (orchestrator owns) |

---

## Notes

- `checkSources` still runs after `store.Open` + `BackfillGrades` (position kept to stay a pure extract); a refused source list exits the process, so the open store is not a leak in practice.
- Out of scope, still open: I1 (global `Since` / `SoldStatsReady` epoch, step (c)) | I2 (roadmap nav R1 text) | I3 (plan header key).
