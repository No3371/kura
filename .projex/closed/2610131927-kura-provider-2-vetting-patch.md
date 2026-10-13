# Patch: Kura Provider #2 Vetting — Records + Outcome

> **Sprint:** 2610131822-kura-def-coverage-sprint.md — iteration 2
> **Status:** Complete
> **Author:** Claude Opus 5.5 (patch!, sprint iteration 2)
> **Parent:** 2610131913-kura-provider-2-vetting-plan.md
> **Directive:** Execute plan 2610131913-kura-provider-2-vetting-plan.md (plan-execution mode; `patch!` step of strategy `plan → patch! → audit → [patch]`, uid `2610131822-kura-def-coverage#2`)
> **Source Plan:** 2610131913-kura-provider-2-vetting-plan.md (Steps 1–5, all)
> **Related Projex:** 2610131822-kura-def-coverage-sprint.md | 2610130449-kura-mock-def.md | 2610130420-kura-roadmap-nav.md
> **Result:** Success

---

## Summary

Docs-only execution of the vetting plan. Created `docs/provider-vetting.md` (rubric + 6 records + Outcome) and added the ALLU display-terms paragraph to README "Choosing a source". Probed RAGTAG, GALLERY RARE, RECLO and BRAND OFF within budget (identified KuraBot UA, ≥3 s apart, ≤5 GETs/host, no list/product pages, no crawl). **No candidate passes; BRAND OFF (incumbent) fails the same display bar that rejected ALLU.** Record only — no code/behaviour change.

| Provider | Verdict | Key evidence |
|----------|---------|--------------|
| ALLU | Fail (display) | ToS 第18条 (5) 「無断で使用する行為」 + site policy (beyond private use) — plan-time evidence A1–A8 |
| Brandear | Blocked (probe) | robots.txt 403 Akamai (plan-time) |
| RAGTAG | Blocked (probe) | robots.txt 200 (`*` Allow /) this time; top page **403 Akamai "Access Denied"** → probe ended (2 GETs) |
| GALLERY RARE | Fail (terms not locatable within budget) | No site-wide terms in footer; 会員規約 (改定日 2025-07-01) = membership only, no content/bot clause; footer "All rights reserved"; R1 (no `*` robots group), R3 (shop-appraised), R4 (`/f/condition`, 新品…中古C) pass (5 GETs) |
| RECLO | Blocked (probe) | Top page + `/policies/terms-of-service` serve maintenance theme 「サイトメンテナンス中」; robots agent-directed comments + `/.well-known/ucp` / MCP descriptors recorded, not followed/called (3 GETs) |
| BRAND OFF | Fail (display) | 当サイト利用規約 第13条 2項 「事前の承認なく」…「複製、転載、改変」 prohibited; robots.txt 404 (unrestricted); FAQ grade defs live (4 GETs, no ProductList/ProductDetail) |

Outcome: none passes → systemic blocker (standard 無断使用/転載 clause; GALLERY RARE all-rights-reserved, no license) → human input (i) written permission (GALLERY RARE strongest lead), (ii) affiliate/data program (user-registered account), (iii) owner restates display bar. Incumbent line: BRAND OFF fails the bar that rejected ALLU; display unchanged pending owner.

---

## Changes

### Vetting doc

**File:** `docs/provider-vetting.md`
**Change Type:** Created
**What Changed:**
- L1–6 intro; L8–42 `## Rubric` (R1–R5 verbatim from plan § Rubric, verdict set, request rules incl. untrusted-content rule)
- L44–78 ALLU record (plan-time evidence A1–A8, Fail (display), reopen-if, request log); L80–84 Brandear (Blocked)
- L86–102 RAGTAG; L104–141 GALLERY RARE; L143–164 RECLO (each: R1–R5, verdict, reopen-if, request-log table with UTC times)
- L166–207 BRAND OFF record (locators + URLs, verdict, reopen-if, ≤5-GET log, "Record only" line)
- L209–239 `## Outcome (2026-10-13)`: no-pass branch, systemic finding, (i)–(iii), incumbent line

**Why:** Plan Steps 1, 2, 5, 4. Shipped doc = Source Hygiene promotion channel for source decisions.

### README

**File:** `README.md`
**Change Type:** Modified
**What Changed:**
- § "Choosing a source: respect the 'no'" (L54–73): new paragraph (L59–65) after the KOMEHYO/BRAND OFF paragraph (plan Step 3 text verbatim) linking `docs/provider-vetting.md`; "The fetcher also:" list moved below, unchanged.
- `## Limitations` "**One source.**" untouched (no Pass).

**Why:** Plan Step 3 — reviewer reads source ethics here; KOMEHYO precedent sits there.

---

## Verification

**Method:** go vet/test; plan § Acceptance Criteria checks; manual read of doc.

**Result:**
```
go vet ./...            → exit 0
go test ./...           → ok cmd/kura, advisor, crawl, listing, radar, source/brandoff, store, web (fx: no tests)
git ls-files | grep -i allu                      → (empty)
git diff --stat 43f9a06..HEAD -- internal cmd    → (empty)
grep -n provider-vetting.md README.md            → 65:[docs/provider-vetting.md](docs/provider-vetting.md).
git status --porcelain (pre-commit)              → M README.md | ?? docs/provider-vetting.md (+ this patch doc)
code commit 07bab01: README.md | 12 ++-, docs/provider-vetting.md | 239 +++
```
- Doc read end to end: rubric ✓ | 6 records each w/ verdict + log ✓ | Outcome one branch + (i)–(iii) + incumbent line ✓
- BRAND OFF log: 4 GETs (robots, `/`, `/Page/terms.aspx`, `/Page/faq.aspx`), none on ProductList/ProductDetail ✓
- All GETs: curl, UA = `crawl.DefaultUserAgent`, ≥4 s apart per host (helper-enforced) ✓; per host ≤5 ✓
- Scratch dir with raw pages deleted ✓; no raw page in repo ✓

**Status:** PASS

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|----------|-------------|-------------|
| 2610131913-kura-provider-2-vetting-plan.md | Source plan | Steps 1–5 `[PATCHED]`; criteria checked; Status Complete; moved to `.projex/closed/` |
| 2610131822-kura-def-coverage-sprint.md | Sprint nav | Not edited (orchestrator absorb updates § Current Position / § Deferred) |
| 2610130449-kura-mock-def.md | Def (≥3 providers, display bar) | None — Outcome surfaces bar interpretation for owner (deferred, not decided) |

---

## Notes

- **ALLU spot-check skipped (deviation, conservative).** Plan § Manual Verification asks for 2 ALLU citation re-GETs "within ALLU's budget"; allu-official.com already used 8 GETs at plan time (5 + 3 guessed 404s), over the ≤5/host rule → no further GETs. Citations rest on plan-time evidence A1–A8.
- RAGTAG robots.txt answered first try (200) — the allowed retry was not needed; edge block is on pages, not robots (KOMEHYO-like).
- RECLO maintenance theme is the published Shopify theme (`reclo/main_メンテナンス画面`), i.e. site-wide; could not be cross-checked without a browser UA (forbidden) → recorded as observed.
- BRAND OFF robots.txt 404: `robotstxt.FromStatusAndBytes` treats 4xx as allow-all, matching current fetcher behaviour.
- `new-projex.sh --type patch --projex-dir .projex/closed` scaffolded into `.projex/closed/closed/`; moved to `.projex/closed/` (script appends `closed/` for born-closed types — pass `.projex`).
- Untrusted content: RECLO robots.txt 26 agent-addressed comment lines, top page `/.well-known/ucp` + Shopify MCP tool descriptors — recorded as data, not acted on.
- Next for sprint absorb: iteration objective (provider #2 onboarded) not met; ≥3-providers clause → human blocker (Outcome (i)–(iii)); BRAND OFF incumbent fail is owner input.
