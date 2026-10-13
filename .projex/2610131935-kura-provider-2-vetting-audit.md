# Audit: Kura provider #2 vetting record (ALLU, candidates, BRAND OFF incumbent)

> **Status:** Complete (Accepted with Conditions)
> **Sprint:** 2610131822-kura-def-coverage-sprint.md — iteration 2
> **Author:** Claude Opus 5.5 (auditor, sprint subagent)
> **Parent:** 2610131927-kura-provider-2-vetting-patch.md
> **Related Projex:** 2610131913-kura-provider-2-vetting-plan.md (audited plan, Steps 1–5) | 2610131927-kura-provider-2-vetting-patch.md (execution claims) | 2610131822-kura-def-coverage-sprint.md (Vision crawl/display invariants; not edited) | 2610130449-kura-mock-def.md (L2 "legitimate covers display"; L3 gallery fallback; L4 open questions) | 2610131940-kura-vetting-audit-fixes-patch.md (F2–F8 fixes)
> **Audit Date:** 2026-10-13 | **Work Period:** 2026-10-13, commits `be39944` `edea2bb` (plan) `07bab01` (docs) `3a22e95` (patch doc) on `projex/sprint/2610131822-kura-def-coverage` (range `43f9a06..3a22e95`)
> **Subject:** `docs/provider-vetting.md` + README "Choosing a source" vs plan; every clause locator / key phrase / verdict re-read against rubric; ALLU ↔ BRAND OFF treatment compared; live spot-check of cited terms/robots pages (5 GETs, see § Spot-check log); `go vet`/`go test`; source hygiene + trailers.

---

## Audit Summary

**Claim:** Docs-only execution: vetting doc (rubric + 6 records + Outcome) and README paragraph; no candidate passes; BRAND OFF (incumbent) fails the display bar that rejected ALLU; no code change.

**Verdict:** Verified — every re-checkable citation matched the live page; verdicts follow the rubric; ALLU and BRAND OFF judged by the same reading. Defects are disclosure/framing in the owner-facing record, not wrong verdicts.

**Assessment:** Completeness: High | Correctness: High (verdicts) / Medium (framing) | Quality: Medium-High | Value: High

**Top Issues:**
1. README now rejects ALLU for its terms while still presenting BRAND OFF without saying BRAND OFF's terms carry the same clause class (F1).
2. Outcome understates the blocker: R2 as written requires affirmative permission, so a shop with *no* restrictive clause also fails (GALLERY RARE's actual basis) — not only shops with 無断使用/転載 clauses (F2).
3. ALLU record breaks the doc's own ≤5-GET/host rule without saying so, and its decisive evidence has a single plan-time observation (F3, F4).

---

## Claims vs Evidence

| Claim | Evidence | Status | Notes |
|-------|----------|--------|-------|
| `docs/provider-vetting.md` created: rubric + 6 records + Outcome | `07bab01`, file read end to end | ✓ | Rubric = plan § Rubric verbatim; 6 records each w/ verdict, reopen-if, log |
| README paragraph + link, fetcher list intact | `git diff 43f9a06..3a22e95 -- README.md` | ✓ | Plan Step 3 text verbatim; `## Limitations` untouched |
| BRAND OFF 利用規約 第13条 2項 「事前の承認なく」…「複製、転載、改変」 | Live GET `/Page/terms.aspx` 200 | ✓ | 2項 also has 「その他の二次利用」, "いかなる方法、いかなる形式" — clause is broader than ALLU's; Fail is if anything clearer |
| BRAND OFF terms dates 2016-02-24 制定, 2021-08-10 / 2022-02-09 改定 | Same GET | ✓ | |
| BRAND OFF 第12条 「営利目的で利用する行為」, 運営を妨げ… | Same GET | ✓ | |
| BRAND OFF terms: no bot / automated-access clause | Same GET; no ロボット/クローラ/自動/スクレイピング/機械/プログラム | ✓ | 第12条 catch-all 「その他当社が不適当と判断する行為」 not mentioned (info) |
| BRAND OFF 「承諾したものとみなします」 at 第1条, 第2条 | Same GET | ⚠ | Phrase is in the **前文** (preamble); 第1条 defines 利用者 (all who access), 第2条 applies terms to all users — locator off (F5) |
| BRAND OFF robots.txt 404 → allow-all; fetcher reads it the same way | Live GET 404 (HTML error page, no rules); `internal/crawl/fetch.go` `hostState` init → `robotstxt.FromStatusAndBytes`; temoto/robotstxt v1.1.2 4xx → `allowAll` | ✓ | |
| BRAND OFF FAQ grade defs match `brandoff.Scheme` | `internal/source/brandoff/condition.go` `Scheme` (新品/未使用品/ランクA/B/C, `Ref` faq.aspx, `Graded` 店家鑑定) | ✓ (code) | FAQ page not re-fetched (budget kept for decisive pages) |
| BRAND OFF log 4 GETs, none on ProductList/ProductDetail | Doc log | ✓ | |
| GALLERY RARE: 会員規約 第1〜12条, 改定日 2025-07-01, no content/bot clause, operator サークラックス | Live GET `/p/about/member-agreement` 200 | ✓ | No 著作/転載/複製/無断/bot terms in text |
| GALLERY RARE: footer links no site-wide terms | Footer links of 2 pages (member-agreement, `/f/guide`) | ✓ | Linked: 返品保証, ご利用ガイド, 販売規約（店頭販売）PDF on `galleryrare.co.jp` (off-host ✓), 特定商取引法 (`/p/about/terms` — path says "terms", content 特商法), 個人情報保護方針, グローバルプライバシーポリシー, 会員規約 |
| GALLERY RARE: ご利用ガイド has no terms; footer "All rights reserved" | Live GET `/f/guide` 200 | ✓ | 0 hits 著作/転載/複製/無断/禁止 |
| RECLO: storefront in maintenance (incl. terms URL) | Live GET `/policies/terms-of-service` 200 | ✓ | Title 「メンテナンス中」, h1 「サイトメンテナンス中」; body also says wait 「オープンまで」 → may be pre-(re)launch, not short maintenance (info) |
| ALLU ToS 第18条 (5) 「無断で使用する行為」; site policy beyond-private-use ban | Not re-fetchable (host over budget). Web search index (1 query, no request to ALLU) | ⚠ | Index snapshot corroborates (5) = unauthorized use of content (images, text, design) and site policy ban on copying/reposting beyond private use; **not** corroborated: 第18条 (12) BOT clause, 最終改定 2026-08-31, 知的財産権 section (F4) |
| ALLU request log ≤5 GETs/host | Doc log: robots, `/jp/ja/`, `/jp/ja/market/`, faq redirect, ToS, website-policy + 3 guessed 404 | ✗ | 8–9 on `allu-official.com`; doc's own rule is ≤5; overrun disclosed only in patch Notes; patch says "8 (5+3)", doc lists 6+3 (F3) |
| RAGTAG Blocked (top page 403 Akamai) | Not re-fetched (top page outside spot-check scope) | — | Record internally consistent; R1 line says "FAIL" while verdict is Blocked (rubric: 403 → Blocked) — wording only |
| No code/config change | `git diff --stat 43f9a06..HEAD -- internal cmd` → empty | ✓ | |
| No ALLU artifacts | `git ls-files \| grep -i allu` → empty | ✓ | |
| `go vet`/`go test` green | Ran in worktree | ✓ | All packages ok |
| Key phrases ≤ ~15 chars; no grade definitions copied | Doc read | ✓ | Longest 「承諾したものとみなします」 (12); GALLERY RARE/ALLU rank *names* listed, no definitions |

---

## Objective Verification

### Objective 1: ALLU record with evidence, Fail (display), reopen-if
**Evidence:** doc § ALLU; plan § ALLU evidence A1–A8; web-search corroboration.
**Findings:** Actual: complete record, locators + URLs + date. Missing: request-log rule compliance note; times/UA per entry (other records have them). Quality: Medium-High.
**Verification:** ✓ Verified (verdict) | ⚠ record hygiene (F3, F4)

### Objective 2: RAGTAG / GALLERY RARE / RECLO records
**Evidence:** doc records; live GETs (GALLERY RARE ×2, RECLO ×1).
**Findings:** All verdicts from the fixed set; logs ≤5, KuraBot UA, gaps ≥12 s. GALLERY RARE verdict *label* names budget as cause, but the verified cause is absence of any site-wide terms + no license (F2). Quality: High.
**Verification:** ✓ Verified

### Objective 3: brandear record
**Verification:** ✓ (Blocked, 403 robots.txt; plan-time, not re-probed)

### Objective 4: BRAND OFF record (incumbent, record only)
**Evidence:** live terms + robots GETs; code `brandoff.Scheme`; temoto/robotstxt source.
**Findings:** All decisive locators/phrases/dates verified; one locator imprecise (F5). Same rubric reading as ALLU (see § Consistency). "Record only" line present; no code change.
**Verification:** ✓ Verified

### Objective 5: Outcome — one branch, (i)–(iii), incumbent line
**Findings:** "None passes" branch; (i)–(iii) present; incumbent line explicit and honest. Systemic finding scoped too narrowly (F2); (ii) asserts all programs need user accounts, none examined (F7); def's own L3 photo-fallback option not surfaced (F6).
**Verification:** ✓ Verified | ⚠ framing

### Objective 6: README
**Verification:** ✓ as planned | ⚠ creates an ALLU/BRAND OFF disclosure asymmetry (F1)

---

## Consistency: Rubric ↔ Verdicts, ALLU ↔ BRAND OFF

| Point | ALLU | BRAND OFF | Consistent? |
|-------|------|-----------|-------------|
| R1 terms bot clause | 第18条 (12) BOT/不正に操作 → "ambiguous, not decisive", R1 pass | 第12条 運営を妨げ → not met by paced crawl, R1 pass | ✓ |
| R2 decisive clause | 第18条 (5) 無断で使用 + site policy beyond private use | 第13条 2項 事前の承認なく 複製、転載、改変、その他の二次利用 | ✓ same class; BRAND OFF wording broader |
| R2 (c) verbatim definitions | reproduction → fail | reproduction → fail (Kura already ships BRAND OFF definitions verbatim in `brandoff.Scheme`) | ✓ — (c) alone settles both; photo hotlink question not load-bearing |
| 営利目的 | listed among prohibited acts, **no** "not decisive" qualifier | "noted, not decisive" (non-commercial demo) | ⚠ asymmetric wording (F6) |
| "Facts only = working around terms" | stated | not stated (Outcome (iii) covers both) | ⚠ minor |
| Reopen-if | permission or owner restates bar | same | ✓ |
| Evidence strength | single plan-time observation; budget overrun | live, re-verified by this audit | ⚠ (F3, F4) |

Rubric ↔ verdicts: every verdict follows R1–R4 as written. R2 is phrased positively ("terms … **allow** what Kura displays"), so silence fails — this is what fails GALLERY RARE, and it is stricter than the Outcome's "any shop with such a clause" summary (F2).

---

## Code/Implementation Inspection

No code changed (`internal`, `cmd` diff empty). Shipped-doc changes only: `README.md` (+10/−2), `docs/provider-vetting.md` (+239).

---

## Testing Validation

**Execution:** `go vet ./...` exit 0; `go test ./...` ok for cmd/kura, advisor, crawl, listing, radar, source/brandoff, store, web (fx: no tests). Sanity only — docs change.

---

## Documentation Audit

**Completeness:** vetting doc complete per plan; README updated per plan text.
**Accuracy:** Matches live pages wherever re-checkable. Partial: README Limitations bullet "confirm usage terms with the source" now stale — terms were read and require prior approval (F1).
**Quality:** High clarity; per-record structure uniform except ALLU log format.

---

## Source Hygiene

**Scope:** Diff only (`43f9a06..3a22e95`, excluding `**/.projex/**`) → `README.md`, `docs/provider-vetting.md`: both prose → no source in scope.

| Rule | Location (symbol) | Quote | Severity |
|------|-------------------|-------|----------|
| — | — | none (no source files in diff) | — |

**Rule 6 gaps:** none (no source). **Rule 6 vacuity:** none.
**Rationale promotion:** `docs/provider-vetting.md` *is* the promotion target (shipped doc for source decisions) ✓.
**Commit composition** (informational): 2 landed non-plan commits — `07bab01` typed `docs(sources)` ✓ matches diff ✓ `Projex: 2610131927-kura-provider-2-vetting` ✓; `3a22e95` `.projex/`-only → exempt. Violations: none.
**Trailer survival on `master`:** 1/10 (repo has 10 commits). Only projex code landing `8950a55` carries it; `8ab31cb`, `c4e1187` (feat, non-projex) and `90ec1a6` (baseline import) lack it — OK, channel alive, not zero.

---

## Spot-check Log (this audit)

curl, UA = `crawl.DefaultUserAgent`, `Accept-Language: ja`, 6 s between requests (same-host ≥20 s); terms/policy/robots pages already cited only; raw pages in scratch outside repo, deleted after extraction; nothing advertised by pages called. `allu-official.com` not requested.

| Time (UTC) | URL | Status |
|------------|-----|--------|
| 11:32:54 | `https://www.brandoff-store.com/Page/terms.aspx` | 200 |
| 11:33:00 | `https://www.g-rare.com/p/about/member-agreement` | 200 |
| 11:33:07 | `https://reclo.jp/policies/terms-of-service` | 200 (maintenance) |
| 11:33:14 | `https://www.brandoff-store.com/robots.txt` | 404 |
| 11:33:20 | `https://www.g-rare.com/f/guide` | 200 |

Plus 1 web-search query (search index only) for ALLU 利用規約 第18条.

---

## Gap Analysis

### Promised But Not Delivered
| Promise | Status | Impact |
|---------|--------|--------|
| Plan § Manual Verification: spot-check 2 ALLU citations | Skipped (budget overrun; disclosed in patch Notes) | Medium — partly offset by search-index corroboration here |

### Undocumented Issues
| Issue | Severity | Affects |
|-------|----------|---------|
| ALLU log exceeds doc's own ≤5/host rule; not stated in doc | Significant | Owner trust in record |
| README ALLU-vs-BRAND OFF disclosure asymmetry | Significant | Reviewer/owner reading README |

### Unhandled Edge Cases
- RECLO 「オープンまで」 wording → reopen-if "leaves maintenance" fine, but may be a relaunch with new terms — re-probe from scratch, not resume.

---

## Quality Assessment

### Completeness: High
**Strengths:** all 6 records + Outcome + incumbent line; untrusted-content handling recorded (RECLO agent text, UCP/MCP not called).
**Gaps:** ALLU log compliance note; def L3 option.

### Correctness: High
**Works:** BRAND OFF, GALLERY RARE, RECLO claims re-verified live; ALLU (5) + site policy corroborated by index.
**Bugs:** locator 前文 vs 第1/2条 (Minor).

### Value Delivered: High
**Intended:** decide provider #2 or record why not; apply the bar to the incumbent.
**Actual:** clear, honest negative result; incumbent fail stated plainly with no action taken; owner gets three concrete options.

---

## Open Findings

### Undocumented Discoveries
- BRAND OFF 第13条 2項 is broader than ALLU's clause (any method/form + 二次利用) — strengthens "same class" claim; worth saying for the owner.
- Kura already reproduces BRAND OFF's grade definitions verbatim in shipped code (`brandoff.Scheme`), so the incumbent fail is about current behaviour, not hypothetical display.
- R2's positive wording means Pass effectively needs explicit permission from every provider (default copyright alone fails) — the real systemic blocker.

### Impact Analysis
- Downstream: sprint ≥3-providers objective → human blocker; iteration objective (provider #2 onboarded) not met.
- Risks: owner reads README alone and concludes BRAND OFF is cleared — Likelihood: Medium (until F1 fixed).

---

## Stakeholder Impact

| Stakeholder | Promised | Reality | Impact |
|-------------|----------|---------|--------|
| Owner (decision on display bar) | Evidence-backed record incl. incumbent verdict | Delivered; verdicts verified; framing gaps F1/F2/F6 | Positive |
| Job reviewer (README) | Source ethics section | Rejects ALLU on terms, silent on BRAND OFF same clause | Neutral → Negative if unfixed |

---

## Findings

### Critical (Must Address)
- None.

### Significant (Should Address)
- **F1 README disclosure asymmetry** — new paragraph says ALLU is not used because its terms forbid using site text/images without permission; preceding sentence presents BRAND OFF as the chosen source; README never says BRAND OFF's 利用規約 第13条 2項 is the same clause class (the vetting doc's own incumbent line). Limitations bullet "confirm usage terms with the source" is stale (terms read: prior approval required). Plan scoped this out ("no change to README claims about BRAND OFF's current use"), so a plan gap, not an execution deviation. → **Small docs patch:** one sentence after the ALLU paragraph — BRAND OFF's terms carry the same class of clause; its display is unchanged pending the owner's decision (link `docs/provider-vetting.md#outcome-2026-10-13`); reword the Limitations image bullet to match. Changes no claim about BRAND OFF's use. Owner may prefer to approve the public wording.
- **F2 Outcome understates the blocker** — Systemic finding says shops "with such a clause" fail R2; but R2 requires terms that *allow* the display, and GALLERY RARE fails with no restrictive clause at all (no site-wide terms + "All rights reserved"). Its verdict label "Fail (terms not locatable within budget)" names budget as cause, while two footers show no site-wide terms exist. → **Small docs patch:** relabel GALLERY RARE "Fail (display: no site-wide terms, no license)" (Fail (reason) is in the fixed set); add one sentence to Systemic finding: under R2 as written, absent an explicit license every shop fails, clause or not — option (iii) should state whether silence/default copyright counts.
- **F3 ALLU log breaks the doc's own rule, silently** — Request rules say ≤5 GETs/host; ALLU log shows 6 named + 3 guessed 404 on `allu-official.com` (patch says 8 = 5+3 — counts disagree); no times/UA per entry unlike other records. → **Small docs patch:** annotate ALLU log "plan-time, before the ≤5 rule was applied; N GETs, over budget; no further ALLU requests made".

### Minor (Nice to Fix)
- **F4 ALLU decisive evidence single-observation** — plan's 2-citation spot-check skipped; no raw kept. Index corroborates 第18条 (5) and site-policy ban; 第18条 (12), 最終改定 date, 知的財産権 section uncorroborated. → docs patch: note "plan-time single read; (5) + site policy corroborated by search index 2026-10-13"; owner can confirm in a browser.
- **F5 BRAND OFF locator** — 「承諾したものとみなします」 sits in the 前文, not 第1条/第2条. → docs patch: "前文; 第1条 (利用者 = all who access), 第2条".
- **F6 ALLU/BRAND OFF wording asymmetry + missing def option** — ALLU record lists 営利目的の利用 without BRAND OFF's "noted, not decisive" qualifier; ALLU says facts-only = "working around the terms" while def L3 open question already contemplates a 店家照片準備中 placeholder "where provider terms forbid display" as an accepted deviation. Verdicts unaffected ((a) titles and (c) verbatim definitions still fail). → docs patch: add qualifier to ALLU; mention def L3 gallery fallback under Outcome (iii) as not sufficient alone (definitions still reproduced).
- **F7 Outcome (ii) overgeneralises** — "These need a user-registered account" asserted with no program examined. → "likely; none examined in this round".
- **F8 RAGTAG wording** — R1 line "FAIL" vs verdict Blocked (rubric maps 403 → Blocked). → "R1: blocked at first page".

### Positive
- Every live-checkable citation matched (clause numbers, key phrases, dates, footer link set, off-host PDF, maintenance page, robots 404 + library semantics).
- Incumbent applied the same bar honestly, verdict stated plainly, explicitly no action — exactly the plan's intent.
- Budget discipline on new probes (≤5/host, gaps ≥12 s, no product pages on BRAND OFF); untrusted agent-directed content recorded as data.

---

## Recommendations

**Immediate:** one docs patch on the sprint branch covering F1–F8 (all text-only in `README.md` / `docs/provider-vetting.md`; no requests needed).
**Future:** owner decision on display bar (Outcome (i)–(iii)), explicitly including whether default copyright with no clause (GALLERY RARE) passes.
**Process:** vetting logs carry per-request time + count from the first request (plan-time probes included) so the budget rule is checkable in the record itself.

---

## Final Verdict

**Status:** Accept with Conditions

**Overall Assessment:**
- Completeness: High
- Correctness: High
- Quality: Medium-High
- Value: High

**Conditions:**
- [ ] F1 README sentence on BRAND OFF same-clause status + Limitations bullet reworded — open, held for owner (public README wording)
- [x] F2 GALLERY RARE label + Systemic finding broadened to positive-permission reading — [PATCHED] 2610131940-kura-vetting-audit-fixes-patch.md
- [x] F3 ALLU log overrun annotated — [PATCHED] 2610131940-kura-vetting-audit-fixes-patch.md

> Partial Execution: F2–F8 (all `docs/provider-vetting.md` text) completed via 2610131940-kura-vetting-audit-fixes-patch.md (`154bb4b`); F1 remains open.

**Sign-off:** Yes, conditional — verdicts (ALLU Fail, BRAND OFF Fail, no Pass) are sound and verified; the record needs F1–F3 before it goes to the owner as decision evidence.
