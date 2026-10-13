# Patch: Kura Vetting Record — Audit Fixes F2–F8

> **Sprint:** 2610131822-kura-def-coverage-sprint.md — iteration 2
> **Status:** Complete
> **Author:** Claude Opus 5.5 ([patch], sprint iteration 2)
> **Parent:** 2610131935-kura-provider-2-vetting-audit.md
> **Directive:** Address findings F2–F8 of 2610131935-kura-provider-2-vetting-audit.md (optional `[patch]` step, uid `2610131822-kura-def-coverage#2`). F1 (README BRAND OFF disclosure) out of scope — held for human owner; README not edited.
> **Source Plan:** Direct (audit findings; original plan 2610131913-kura-provider-2-vetting-plan.md already closed)
> **Related Projex:** 2610131935-kura-provider-2-vetting-audit.md | 2610131927-kura-provider-2-vetting-patch.md | 2610131913-kura-provider-2-vetting-plan.md | 2610131822-kura-def-coverage-sprint.md (not edited) | 2610130449-kura-mock-def.md (L3 gallery fallback cited)
> **Result:** Success

---

## Summary

Text-only fixes to `docs/provider-vetting.md` from evidence already in the record + audit. No network requests, no code, no README. Verdicts unchanged (ALLU Fail, BRAND OFF Fail, no Pass); framing, locators and disclosure tightened so the record can go to the owner as decision evidence.

| Finding | Disposition |
|---------|-------------|
| F1 README asymmetry | **Out of scope** — held for owner (public README wording on BRAND OFF terms) |
| F2 Outcome understates blocker | Fixed — GALLERY RARE relabelled; Systemic finding + (iii) state positive-permission reading |
| F3 ALLU log over budget, silent | Fixed — Budget line: 8 GETs (9 counting faq redirect hop), plan-time, before rules applied; patch "8 (5+3)" vs doc "6+3" reconciled |
| F4 ALLU single observation | Fixed — Evidence-strength line: corroborated vs not corroborated |
| F5 BRAND OFF locator | Fixed — 前文 / 第1条 / 第2条 split |
| F6 ALLU/BRAND OFF asymmetry + def L3 | Fixed — 営利目的 "noted, not decisive" on ALLU; def L3 fallback cited in ALLU record + Outcome (iii) |
| F7 Outcome (ii) overgeneralises | Fixed — "likely … none examined in this round" |
| F8 RAGTAG wording | Fixed — "R1: blocked at the first page" |

---

## Changes

### Vetting record

**File:** `docs/provider-vetting.md`
**Change Type:** Modified (commit `154bb4b`, +34/−16)
**What Changed:**
- L66–74 ALLU R2 (F6, F4): 営利目的の利用 → "non-commercial portfolio demo … noted, not decisive" (mirrors BRAND OFF L202); facts-only sentence now names def L3 店家照片準備中 fallback — dropping photos still leaves titles + verbatim definitions; dropping definitions breaks per-shop rule → not an option. New "Evidence strength" bullet: one plan-time read, no copy kept; 第18条 (5) + site-policy ban corroborated by search-index snapshot 2026-10-13 (no request to ALLU); 第18条 (12), 最終改定 date, 知的財産権 section not corroborated; owner can confirm in browser.
- L82–89 ALLU log (F3): header gains "KuraBot UA, ≥3 s apart; times not recorded" (UA/spacing ← plan § ALLU evidence header). New "Budget: over the ≤5-GET rule" bullet: made during plan research before this file's rules applied; 8 GETs on `allu-official.com` (5 named + 3 guessed 404), 9 if faq redirect's first hop counts on that host; no further ALLU requests incl. plan's 2-citation spot-check.
- L99 RAGTAG (F8): "R1 Acquisition: FAIL at the first page" → "blocked at the first page" (rubric: 403 → Blocked; verdict already Blocked).
- L115 GALLERY RARE heading (F2): "Fail (terms not locatable within budget)" → "Fail (display: no site-wide terms, no license)" — cause is absence of terms, not budget (audit verified footer of 2 pages).
- L195–196 BRAND OFF R2 (F5): 「承諾したものとみなします」 located in 前文; 第1条 defines 利用者 = everyone who accesses; 第2条 applies terms to all 利用者.
- L224 Outcome (F2): GALLERY RARE label matches heading.
- L234–237 Systemic finding (F2): blocker wider than clause — R2 asks for terms that *allow* display, so no clause + no license fails too (GALLERY RARE); as written R2 fails every shop without an explicit license.
- L242–244 Outcome (ii) (F7): "These need" → "These likely need …; no such program was examined in this round".
- L245–250 Outcome (iii) (F2, F6): owner also to say whether silent terms (default copyright, no clause) count as allowing display; def L3 gallery fallback not enough alone (titles + verbatim definitions still reproduced).

**Why:** Audit Accept-with-Conditions; F2/F3 are sign-off conditions, F4–F8 minor. All fixes reword or annotate from evidence in plan/patch/audit — no new facts gathered.

---

## Verification

**Method:** `git diff` hunk review; grep for removed/added phrases; label consistency across heading + Outcome; anchor check (`#outcome-2026-10-13` heading untouched; nothing links to GALLERY RARE heading anchor); README untouched; no code paths in diff.

**Result:**
```
git diff --stat (pre-commit): docs/provider-vetting.md | 50 +++++---- (34+, 16-)
grep "not locatable|FAIL at the first|第1条, 第2条|These need a" → none remaining
grep "provider-vetting.md#" → only audit text (no live link to changed heading)
Commit 154bb4b: docs(sources): … trailer Projex: 2610131940-kura-vetting-audit-fixes
```

**Status:** PASS

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|----------|-------------|-------------|
| 2610131935-kura-provider-2-vetting-audit.md | Parent (findings source) | Conditions F2, F3 checked w/ link to this patch; F1 left open (owner) |
| 2610131927-kura-provider-2-vetting-patch.md | Prior patch (point-in-time) | None — its "Fail (terms not locatable within budget)" label + "8 (5+3)" count stay as historical record; superseded wording lives in shipped doc |
| 2610131822-kura-def-coverage-sprint.md | Sprint nav | Not edited (per directive) |

---

## Notes

- **F1 still open:** README "Choosing a source" rejects ALLU on terms, silent on BRAND OFF's same-class 第13条 2項; Limitations bullet "confirm usage terms" stale. Owner to approve public wording. Audit sign-off condition F1 unmet → record not yet fully cleared for owner hand-off via README; `docs/provider-vetting.md` itself is now complete.
- Budget wording "before this file's request rules applied" ← audit F3 recommendation + patch 2610131927 Notes (ALLU evidence gathered at plan research; ≤5 rule set by that plan).
- Owner decision (iii) now explicitly covers silent-terms case (GALLERY RARE), per audit Recommendations § Future.
