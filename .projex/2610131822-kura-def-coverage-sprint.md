# Sprint: Kura Def Coverage

> **Status:** In Progress
> **Goal:** Complete coverage of the def (2610130449-kura-mock-def.md): every def clause is built, or deferred with a named blocker that needs human input. Providers via open alternatives or simple crawler services (user: "Find open alternatives or build simple crawler services."). Def open questions deferred (user: "Defer them").
> **Base Branch:** master
> **Step:** standard
> **Strategy:** plan → patch! → audit → [patch]
> **Sprint Branch:** projex/sprint/2610131822-kura-def-coverage
> **Worktree Path:** kura/.projexwt/2610131822-kura-def-coverage
> **Parent:** 2610130449-kura-mock-def.md
> **Started:** 2026-10-13 18:22

## Vision

Done = every clause of 2610130449-kura-mock-def.md (Identity, Boundaries § Is / Is not / content kinds / Canvas deviations / Canvas gaps) maps to either built + verified behaviour on the sprint branch, or a Deferred row naming the human input it waits on. Baseline: master `8ab31cb` (canvas UI mock from sprint 2610130524-kura-interactive-mock-sprint.md + seeded 大家都在找 + OpenAI advisor mode).

**User scope decisions (2026-10-13, at setup):**
- **Providers (def: ≥3 = BRAND OFF + ≥2):** reach ≥3 via **open alternatives** (open / keyless data sources) **or simple crawler services**. No marketplace-API keys (Rakuten / Yahoo need user-registered app IDs → not this sprint's channel). Crawl invariants unchanged: honestly identified bot UA, never spoofed; robots.txt + Crawl-delay obeyed; no CAPTCHA/WAF evasion; no C2C scraping against ToS; a provider whose terms forbid crawling or the display Kura makes (def: "legitimate" covers display — attribution, image use, comparative labelling) is dropped, not worked around. Each provider registers its own `listing.Scheme` (own grades + verbatim definitions, 店家鑑定 vs 賣家自評) — never relabelled into another shop's scale.
- **Open questions L2–L4 + roadmap-nav open questions: deferred.** A build item an open question blocks → Deferred with that question's id; never decided in-sprint. Already-shipped provisional choices stay as they are.

**Human-input blockers known at setup** (→ Deferred, not improvised): hosting deploy (user account) | real advisor run (API key) | canvas pass 2 design (def: canvas gaps are designed before built) | crawl-history-dependent values (calendar time: `radar.MinSoldHistory` 14 d, 30-day trend) | marketplace-API providers (user-registered keys).

**Verification:**
- Clause-by-clause map def → evidence (code path / test / browser check) or Deferred row with blocker
- `go vet ./...` + `go test ./...` green on the sprint branch
- App starts from the worktree; new/changed surfaces checked in the browser (light/dark × desktop/mobile where UI changes)
- Each new provider: robots.txt + terms check recorded, identified UA, fixtures + parser tests, its own Scheme

Out of sprint: deploying to a host, entering credentials, canvas pass 2 design, answering def/nav open questions, pushing to any remote.

## Current Position

**As of 2026-10-13 (setup):** not yet derived — iteration 1 derive sets position.

## Iterations

| # | Objective | Result | Artifacts |
|---|-----------|--------|-----------|

## Deferred

| Item | Def clause | Blocker (human input) |
|------|-----------|------------------------|

## Open Questions

- [ ] (none yet — def/nav open questions are deferred by user decision, not re-listed here)

## Revision Log

| Date | Iter | Summary |
|------|------|---------|
| 2026-10-13 | — | Sprint started from master `8ab31cb`; scope decisions recorded in Vision (providers via open alternatives / simple crawlers; open questions deferred) |
