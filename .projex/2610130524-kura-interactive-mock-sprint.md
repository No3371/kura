# Sprint: Kura Interactive Mock

> **Status:** In Progress
> **Goal:** Running `kura` locally, a reviewer walks every canvas-designed flow (canvas pinned 1791839497-46ec) end-to-end on real BRAND OFF data, with real / real-derived / mocked exactly as 2610130449-kura-mock-def.md classes them (canvas 資料不足/hidden states where history is short), in light/dark × desktop/mobile; `go test ./...` green; every def § Canvas gaps item either built from existing canvas evidence or listed under this nav's Deferred with its blocker.
> **Base Branch:** master
> **Step:** standard
> **Strategy:** plan → patch! → audit → [patch]
> **Sprint Branch:** projex/sprint/2610130524-kura-interactive-mock
> **Worktree Path:** kura/.projexwt/2610130524-kura-interactive-mock
> **Nav:** 2610130420-kura-roadmap-nav.md
> **Parent:** 2610130420-kura-roadmap-nav.md
> **Started:** 2026-10-13 05:24

## Vision

Done = the mock a reviewer would open, minus hosting. Covers roadmap Phase 3 (redesign impl), Phase 4 (member layer, demo-grade personas) and Phase 5 (market depth) to the extent the pinned canvas designs them, on the single live provider (BRAND OFF).

Verification:
- App starts from the sprint worktree; every canvas screen reachable and rendered against the pinned canvas (browser check, both themes, desktop + mobile widths)
- Each displayed stat traceable to its def class: real | real-derived (computed from crawl history; 資料不足 when short) | mocked (needs real users only)
- No stat shown in both UI and advisor from different computations
- `go test ./...` green
- Deferred list complete: each canvas-gap item not built names its blocker (canvas pass 2 | open human question in def/nav)

Out of sprint (need human credentials, approval, or design work): new providers (roadmap Phase 2 — API keys / terms vetting), public hosting + send-readiness gate (Phase 6), canvas pass 2 design, answers to def/nav Open Questions. Where an open question blocks a build item, defer it — don't decide it.

## Current Position

**As of 2026-10-13 (iteration 0):**

Not yet assessed — first derive sets position.

## Iterations

| # | Objective | Result | Artifacts |
|---|-----------|--------|-----------|

## Deferred

| Item | Blocker |
|------|---------|

## Open Questions

- [ ] (none yet)

## Revision Log

| Date | Iter | Summary |
|------|------|---------|
| 2026-10-13 | — | Sprint started — goal scoped by orchestrator from user prompt "sprint-projex to build the interactive mock end-to-end" (user AFK); hosting/new providers/canvas pass 2 excluded as needing human action |
