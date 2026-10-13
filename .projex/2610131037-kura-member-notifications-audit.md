# Audit: Kura member state + notifications

> **Status:** Complete (Accepted)
> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 6
> **Author:** Claude Opus 5.5 (auditor, sprint subagent)
> **Parent:** 2610131027-kura-member-notifications-patch.md
> **Re-audit:** 2026-10-13 of fix commits `a88cee9` `4d68086` + doc `92f1d17` (range `cadc56f..92f1d17`) → § Re-audit; conditions F1, F2 met
> **Related Projex:** 2610131010-kura-member-notifications-plan.md (audited plan, S1–S9, A1–A14) | 2610131027-kura-member-notifications-patch.md (execution claims) | 2610131048-kura-newest-order-member-feed-patch.md (F1–F4, F6 fixes, re-audited) | 2610130524-kura-interactive-mock-sprint.md (blocking map: member entry, data constraint) | 2610130449-kura-mock-def.md (Mocked = feed + persona watchlists; L3 personas / notification types / static vs reactive; L4 day-1 appearance)
> **Audit Date:** 2026-10-13 | **Work Period:** 2026-10-13, commits `f92eea3` (code + README) `e76ca3c` (patch doc) on `projex/sprint/2610130524-kura-interactive-mock` (range `63644b7..e76ca3c`)
> **Subject:** plan S1–S9 vs code, tests, live app (worktree `kura.db` served on 127.0.0.1:8093, no crawl flags; read-only copy of the DB in scratchpad for SQL probes, deleted after), canvas `project/Main.dc.html` `isMember` block + `KURA_NOTIFS` read from served version 1791845608-ba84 (sprint nav: identical to pin 1791839497-46ec)

---

## Audit Summary

**Claim:** Header 登入 → demo member 林小姐 in-app; member header + 通知 dropdown per canvas; one session watchlist store behind every 追蹤 toggle; `GET /api/member` feed "mocked in who tracks what, real in what happened" — every notification an event the crawl observed, carrying the listing page's own radar `Item`.

**Verdict:** Verified — S1–S9 hold as written; the S6 premise ("an event the crawl actually observed") does not hold for the only events the demo DB produces (F1).

**Assessment:** Completeness: High | Correctness: Medium | Quality: High | Value: Medium-High

**Top Issues:**
1. **F1** — 新上架 on day-1 data is not an observed event, and the pick is inverted: "newest first_seen" within the single sweep = the *last-crawled* listing; BRAND OFF is paged newest-first (`listURL` `sort=Number1`), so the notified listing is among the model's *oldest* on the shop. Comments / README / commit claim "observed".
2. **F2** — un-tracking a model leaves 「你追蹤的型號 …」 in the feed; on the model page the dropdown and the unpressed 追蹤這個型號 contradict each other on one screen.
3. **F3** — pre-existing, same root as F1 (not this iteration): Home 最新上架 + Browse `sort=newest` order by `first_seen` → on day-1 data they surface the last-crawled (oldest-listed) items as 最新上架.

---

## Claims vs Evidence

| Claim | Evidence | Status | Notes |
|-------|----------|--------|-------|
| `IsDeal` extracted verbatim, `Search` unchanged | `radar.go` diff | ✓ | `!IsDeal` ≡ old disjunction (De Morgan); no constant touched |
| `Gone(limit)` most recent first, copies | `radar.go` `Gone`, `TestGone` | ✓ | tie by id; mirrors `Drops` |
| `/api/member` shape + A2 pick | live `curl` :8093 | ✓ | watchlist `[model:LOUIS VUITTON/M51130, model:CHANEL/A01112, brandoff:2101220216016, brandoff:2101220285869]` |
| Event item = `/api/listings/{id}` item | script over both events, keys `price_twd status fair drop first_seen condition model_name image_url` | ✓ | SAME ×2 |
| "each notification is something the crawl actually observed" (`member.go` header, README row, commit body, plan Summary) | DB probe: LV sweep run 4 00:36:31→00:43:20, 99 pages newest-first; M51130 newest `first_seen` 00:42:32 (`brandoff:2101219356990`), earliest 00:36:35 (`brandoff:2101220324117`); CHANEL A01112 00:36:28 vs run 3 end 00:36:31 | ✗ | No 新上架 was observed — the whole catalogue was first seen in one sweep; the pick = deepest page = oldest on shop → F1 |
| `go vet` / `go test` green | re-run | ✓ | advisor, radar, brandoff, store, web ok |
| Hygiene grep none | re-run | ✓ | exit 1 |
| N1 sign-in, no reload, focus bell, badge 2 | browser (fresh session) | ✓ | `__probe` kept; aria-label `通知，2 則未讀` |
| N3 popover geometry + copy | browser 1280 | ✓ | 382 wide (380 + borders), right = bell right + 8, top = bell bottom + 10; focus h2; bodies `新的一件 ランクB，NT$14,129，低於行情 13%` / `… ランクA，NT$181,653，低於行情 25%` |
| N4 no contradiction with Kura page | browser: click item 1 | ✓ | `/listing/brandoff%3A2101219356990` in-app; NT$14,129, 低於行情 13%, 上架 2026/10/13 · 在架第 1 天, 「上架以來還沒有調價」; badge → 1 |
| N8 feed static after toggles (A8) | browser: model page 追蹤這個型號 off | ⚠ | as designed, but the feed still says 「你追蹤的型號 アルマPM モノグラム」 beside an unpressed toggle → F2 |
| N11 375 dark | browser 375×812 dark | ✓ | panel fixed 8/64/367; heart + name `display:none`; `scrollWidth 375`; 0 console errors |
| Canvas copy / markup | `Main.dc.html` `isMember`, `KURA_NOTIFS` | ✓ | templates, dims, colours, footer match; dot `aria-label` → sr-only 未讀 (equivalent) |
| README row (deviation 1) | diff | ⚠ | accurate except "the … new listings the crawl observed" (F1) |

---

## Objective Verification

### S1 Entry / S2 Header / S3 Dropdown / S4 Read state
**Evidence:** `notify.js`, `index.html`, `kura.css`; browser N1/N3/N4/N11 re-run; canvas markup.
**Findings:** all criteria met; Escape / outside / popstate / item close paths present in code (patch N7 evidence not re-run). `[hidden] { display:none !important }` keeps `.k-member`'s `display:flex` from defeating `hidden`.
**Verification:** ✓ Verified

### S5 One watchlist store
**Evidence:** `member.js`, `card.js` `paintTrack`/`paintTracks`/`onChange`; browser model-page toggle → 追蹤清單（3）.
**Findings:** one store, one paint path; `isTracked` re-export keeps detail/model/advisor untouched; `model.js` `trackKey` = `member.go` `modelKey` form (model page showed pressed before toggle).
**Verification:** ✓ Verified

### S6 No contradiction
**Evidence:** `TestMemberAPI` (drop, gone, new invariant + empty store); live SAME ×2; N4.
**Findings:** figures identical to the linked Kura page; 降價 only from history `Drop`, 已找到新主人 only for gone (detail page also says 已找到新主人 for gone). The literal rule "新上架 only for a tracked model's newest active listing" holds. The premise does not: on the demo DB the 新上架 events are crawl-order artefacts, inverted vs the shop (F1).
**Verification:** ✓ (as written) / ⚠ (premise)

### S7 Server / S9 checks
**Evidence:** diff, tests re-run.
**Verification:** ✓ Verified

### S8 Responsive / themes
**Evidence:** 375 dark spot-check, 1280 light; targets: readall + footer links `min-height: 24px`, icon buttons 40 px.
**Verification:** ✓ Verified (spot-check; full N11/N12 matrix from patch not re-run)

---

## Code/Implementation Inspection

### `internal/web/member.go`
**Actual:** as planned; typed `memberView`; `newestActive` = max `FirstSeen` (tie smaller id). Quality: High except the newest rule.
**Issues Found:**
- `newestActive` ignores the radar's own precedent: `radar.go` `Engine.stats` counts days-to-sell only for items with `FirstSeen.After(e.epoch.Add(24*time.Hour))` ("Only items we watched appear …"). The member pick has no such gate, so first-sweep listings qualify as 新上架 — Severity: Significant (F1)
- Header comment "every notification is something the crawl actually observed" — false on first-sweep data — Severity: Significant (F1)

### `internal/web/static/js/notify.js`
**Actual:** as planned. Quality: High.
**Issues Found:**
- `paint()` renders every event regardless of `isTracked(ev.key)` → F2 — Severity: Minor/Significant
- Reload as member: `paint()` runs with `events = []` until the refetch resolves — badge hidden, 「目前沒有通知。」 if opened; on a failed refetch the empty state stays and reads as "no notifications" — Severity: Minor (F4)

### `internal/web/static/js/member.js`, `card.js`, `util.js`, `index.html`, `kura.css`
**Actual:** as planned; rationale comments present (localStorage rejected, replace-on-sign-in rejected, inset ring, mobile max-height guard). Quality: High. No issues beyond F2/F4.

### `internal/radar/radar.go`
**Actual:** `IsDeal`, `Gone` exactly as claimed. Quality: High.

**Undocumented:** none beyond the patch's declared deviations (README row, MarkGone test phase, static-route body probes, typed struct, 24 px min-heights) — all verified.

---

## Testing Validation

**Execution:** all pass (re-run `-count=1`).
**Coverage:** `TestIsDeal` 6 cases; `TestGone` order + limit; `TestMemberAPI` drop → gone transition + invariant + empty store; static routes.
**Missing:** no test pins the 新上架 semantics against first-sweep data (a fixture where all rows share the first sweep would have exposed F1); invariant omits `condition`/`model_name` (checked live: SAME).

---

## Documentation Audit

**Accuracy:** README `GET /api/member` row: "the drops, sales and new listings the crawl observed on it" — overstated on day-1 data (F1). Patch doc claims otherwise accurate (line refs, N-matrix results consistent with re-runs).

---

## Source Hygiene

**Scope:** Diff only (`63644b7..e76ca3c`, `:(exclude,glob)**/.projex/**`)

| Rule | Location (symbol) | Quote | Severity |
|------|-------------------|-------|----------|
| — | — | no projex IDs, no `file:ln`, no step banners, no changelog narration | — |

**Rule 6 gaps:**
- `member.go` → `newestActive` — choosing max `FirstSeen` as "newest" is a trap on first-sweep data (crawl order is shop newest-first); no comment names it, and the existing `epoch + 24h` gate in `radar.go` is not reused (F1).

**Rule 6 vacuity:** none — `member.js`, `notify.js`, `kura.css` comments name rejected alternatives / constraints.

**Accuracy (not a numbered rule, same channel):** `member.go` package-level comment and `notify.js` `copy` comment are correct about figures; the `member.go` "actually observed" sentence is not (F1).

**Rationale promotion:** README API row added; A9 / A10 rationale lives in `member.js` comments — none further needed.

**Commit composition** (informational): 1 landed code commit — typed subject 1/1 | type matches diff 1/1 | `Projex:` trailer 1/1. `e76ca3c` doc-only (exempt, has trailer). Commit body repeats the "observed" claim (history — fix forward only).

**Trailer survival on `master`:** 0/6 — N/A, not Critical: `master` holds only the baseline import + pre-sprint doc commits (sprint unmerged). Sprint branch: every code commit carries `Projex:`.

---

## Gap Analysis

### Promised But Not Delivered
| Promise | Status | Impact |
|---------|--------|--------|
| Feed = only crawl-observed events (plan Summary, A3) | Partial — 新上架 on first-sweep data is unobserved and inverted | Medium-High (reviewer-facing credibility) |

### Undocumented Issues
| Issue | Severity | Affects |
|-------|----------|---------|
| F1 inverted 新上架 pick | Significant | feed, `member.go`, README |
| F2 untracked model still "你追蹤的型號" | Minor/Significant | feed vs model page |
| F3 Home 最新上架 / Browse newest = last-crawled on day-1 data (pre-existing) | Significant (pre-existing) | Home shelf, Browse sort |
| F4 reload flash / failed refetch reads as empty | Minor | header badge, dropdown |

### Unhandled Edge Cases
- Untracking a listing with a 降價/已找到新主人 event → same as F2 for listing events (not reachable on today's DB).
- Watchlist drift after a crawl (plan Risk) — recomputed pick on reload adds events for keys not in the stored watchlist; logged by plan, not re-tested.

---

## Quality Assessment

### Completeness: High
**Strengths:** every S-criterion met; canvas parity; tests extend beyond plan (gone path).
**Gaps:** first-sweep semantics of 新上架.

### Correctness: Medium
**Works:** sign-in, store, repaint, read state, no-figure-contradiction: Yes.
**Bugs:** F1 (Significant), F2 (Minor/Significant), F4 (Minor).

### Code Quality: High
**Positive:** small modules mirroring advisor pattern; one store/one paint path; typed API; rationale comments.
**Tech Debt:** model-key format in two places (plan Risk, verified consistent) — Low.

### Value Delivered: Medium-High
**Intended:** walkable designed member flow, honest data.
**Actual:** flow fully walkable and canvas-faithful; the two visible notifications are honest about figures but not about "new".
**Impact:** User: Positive, with a credibility risk a probing reviewer can find (def F19-type risk) if they open BRAND OFF's newest-first list.

---

## Open Findings

### Undocumented Discoveries
- F3 — `radar.go` `sortItems` `newest` + Home `api("/api/listings?sort=newest&limit=5")`: on day-1 data the only genuinely new item is the one CHANEL listing first seen in run 9 (01:41); the rest of the shelf are the last pages of the LOEWE/CELINE sweeps (oldest-listed). Pre-existing since iteration 1 / 5; same root and same remedy family as F1 (def L4 day-1 appearance).

### Impact Analysis
- Future: a later def L3 answer (personas, reactive feed) replaces `member` + `signIn` only — claim holds; README row would need the persona wording revised.
- Risks: F1 visible to any reviewer who cross-checks the shop — Likelihood: Low-Medium; Impact: High for a portfolio demo whose takeaway is product thinking.

---

## Stakeholder Impact

| Stakeholder | Promised | Reality | Impact |
|-------------|----------|---------|--------|
| Reviewer (104 job) | walkable member flow, no fabricated data | walkable; figures real; 新上架 label unobserved/inverted | Positive with risk |
| Def owner (user) | L3 questions stay open | A1/A2/A3/A8/A9 provisional, named; nothing pre-empted beyond those logged | Neutral |

---

## Findings

### Critical (Must Address)
- None.

### Significant (Should Address)
- **F1 — 新上架 on first-sweep data is unobserved and inverted** — `member.go` `newestActive` takes max `first_seen`; the demo DB is one sweep paged newest-first, so the notified listing (`brandoff:2101219356990`, 00:42:32 of a 00:36–00:43 sweep) is among the model's oldest on BRAND OFF, and the "observed" claim in `member.go`, README and plan is false for it. Kura's own page does not contradict it (both say 上架 2026/10/13), so the def rule holds; the fabrication risk is reviewer-side. **Fixable now (mechanism):** gate `new` events with the radar's existing watched-appear rule (`FirstSeen.After(epoch + 24h)`, exported via a small accessor) and correct the comment/README. **Needs a decision (outcome):** on today's DB that empties the feed (「目前沒有通知。」, no badge) and loses the designed 2-unread state. Alternatives: (a) honest empty feed; (b) keep a Mocked 新上架 but pick the shop's newest (min `first_seen` within the first sweep) and reword comments/README to "mocked on first-sweep data"; (c) keep as is, log as def L4 day-1 deviation. Orchestrator can choose under def "feed = Mocked"; (c) is the weakest.
- **F2 — feed still says 你追蹤的型號 after un-tracking** — `notify.js` `paint()` renders all events; verified live: model page shows 追蹤這個型號 unpressed while the open dropdown on that page lists 「你追蹤的型號 アルマPM モノグラム」. **Fixable now:** render/count only events whose `ev.key` is tracked (hides, never adds — does not answer def L3 "does tracking add a synthetic entry"). Orchestrator decision because A8 (static feed) is provisional against L3; recommend the filter.

### Minor (Nice to Fix)
- **F4 — reload as member shows no badge / empty state until refetch; a failed refetch reads as "目前沒有通知。"** → keep the last feed in the session store, or leave the list unpainted (no empty line) until loaded; fixable now.
- **F5 — 新上架 title uses item `model_name` (`マトラッセ Wフラップ`) vs model page longest name (`… クルーズコレクション`)** (patch-noted) → canvas pass 2 copy review; no action now.
- **F6 — `nav.js` header comment re-wrap leaves a dangling "The" line** → cosmetic re-flow.

### Pre-existing (out of iteration scope, for orchestrator)
- **F3 — Home 最新上架 / Browse `sort=newest` = last-crawled on day-1 data** → same decision as F1 (def L4 day-1 appearance); if F1 takes option (b), apply the same ordering to `sortItems` `newest`.

### Positive
- Server-side pick + same-`Item` events make figure contradiction structurally impossible; tested including the gone transition.
- Canvas parity is exact (dims, copy templates, colours, mobile sheet); a11y beyond canvas (sr-only 未讀, live aria-label, 24 px floors, inset ring).
- Provisional choices are named and confined to `member.go` / `signIn`; def L3 items are not pre-empted.

---

## Recommendations

**Immediate:** F1 mechanism fix + comment/README correction once outcome (a)/(b)/(c) is chosen; F2 filter; F4 small.
**Future:** F3 with F1's decision; def L4 day-1 appearance → record that first-sweep `first_seen` is crawl order, inverted vs shop order.
**Process:** plans that derive "new"/"newest" from `first_seen` should state the sweep order of the source.

---

## Final Verdict

**Status:** Accept with Conditions → **Accept** after re-audit (conditions met)

**Overall Assessment:**
- Completeness: High
- Correctness: Medium
- Quality: High
- Value: Medium-High

**Conditions:**
- [x] F1 — choose outcome (a)/(b)/(c); apply mechanism + correct `member.go` comment and README row → orchestrator chose (b); `a88cee9`, re-audited ✓
- [x] F2 — filter feed to tracked keys (or log as accepted L3 provisional) → `4d68086`, re-audited ✓

**Sign-off:** Yes, conditional — criteria met and the def's no-contradiction rule holds; F1 is a credibility defect in exactly the area the sprint nav flagged. **After re-audit: Yes** — conditions met; R1–R2 (below) carried to orchestrator, none makes a shown fact false.

---

## Re-audit (fix patch 2610131048-kura-newest-order-member-feed-patch.md)

**Scope:** `a88cee9` fix(radar) (`watched`, `newer`, `NewestActive`, `sortItems` → method, `stats` reuse; `member.go`; README *Newest listings* + `/api/member` row; `TestNewestOrder`, `TestMemberAPI`), `4d68086` fix(web) (`notify.js` `shown()` + `feed` state, `nav.js` comment), `92f1d17` doc. Live: patched build on the worktree `kura.db` at 127.0.0.1:8097 (no crawl flags), read-only scratch DB copy for SQL (deleted), built-in browser.

**Build / IDE diagnostics:** `go build ./... && go vet ./... && go test -count=1 ./...` → clean; advisor | radar | brandoff | store | web ok (go1.26.0). The reported `radar_test.go` errors are **refuted** — around line 251 it is `gone.Status, gone.GoneAt = "gone", gone.FirstSeen.Add(time.Hour)` (`store.Row.GoneAt` is `time.Time`, no `&LastSeen` anywhere), and `"slices"` is imported at the top of the file. The IDE was showing stale diagnostics.

| Finding | Evidence | Result |
|---|---|---|
| F1 pick | `/api/member` → `model:LOUIS VUITTON/M51146` → `brandoff:2101220315146` (00:36:43, 12 s into LV run 4), `model:CHANEL/A67085` → `brandoff:2106800652536` (00:31:50, 3 s into CHANEL run 3); SQL `ORDER BY first_seen, id` per model (active): both == pick, unique earliest (no tie). Both items == `/api/listings/{id}` (8 keys SAME). LV moved off M51130 because its shop-newest listing is not a deal (loop as A2). | ✓ |
| F1 claims | `member.go` header now separates crawl-observed 降價 / 已找到新主人 from 新上架 "which Kura never saw appear"; `newer` comment names the rejected rule; README *Newest listings* + `/api/member` row are accurate. Commit `f92eea3` body still says "observed" (history, fix forward). UI copy unchanged: 「新上架 · 今天 00:36」 — the time is the crawl's first sight, not the shop's listing time; the detail page's 上架 date uses the same `first_seen`, so they agree. That is the Mocked-feed outcome the orchestrator accepted under (b). | ✓ |
| F3 | Home 最新上架 DOM ids == `/api/listings?sort=newest&limit=5` == Browse `/browse?sort=newest` first 5 == SQL earliest 5 (GOYARD run 1, 00:25:25 = epoch). | ✓ (see R2) |
| F2 | Model page M51146, dropdown open, toggle via keyboard-path `click()` (no pointerdown): panel stays open, items 2 → 1, badge 2 → 1, aria `通知，1 則未讀`, 追蹤清單（3）, button 追蹤這個型號; re-track → 2 / badge 2. Hide-only (nothing added). Also hides drifted-pick events whose key is not in the stored watchlist (plan Risk). | ✓ |
| F4 | Fresh `notify.js` instance with `/api/member` rejected after 800 ms: during load list `""`, badge hidden, aria `通知`; after → 「暫時無法載入資料，請稍後再試。」 (the existing generic error copy, plan A9 of iter 4). | ✓ |
| F6 | `nav.js` comment re-flowed. | ✓ |
| Regressions | `price_asc` / `price_desc` / deal order unchanged (comparators untouched; deal top `brandoff:2101220216016` as before); detail page of the new pick: NT$24,220 / 低於行情 14% / 上架 2026/10/13 · 在架第 1 天 / no-change history (matches feed); days-to-sell gate now `e.watched` (same expression); advisor `sort: newest` goes through `Search` → same order as Home/Browse (no advisor copy describes the order); `NewestActive` takes its own RLock outside `Search` (no nested lock). 0 console errors. | ✓ |

**Source hygiene (fix diff):** no projex refs / `file:ln` / banners; `newer`, `watched`, `notify.js` header and `feed` comments name the rejected alternative or constraint. `a88cee9` + `4d68086` typed subjects with `Projex:` trailers; `92f1d17` doc-only.

### Patch agent's two product calls — truthful vs suboptimal

- **R1 — a genuinely new listing inside the epoch + 24 h window sorts last.** On today's DB this is the one CHANEL listing first seen in run 9 (`brandoff:2101220126889`, 01:41:20): it is the newest listing in the DB and is position 15,367 of 15,367 in 最新上架. **Not visible, so nothing a reviewer sees is false today** (the card and detail facts stay true). It becomes reviewer-visible once crawling resumes: for the first 24 h of hourly crawls (hosted realtime, roadmap Phase 6), every genuine new arrival would sink to the bottom of 最新上架 and could never be a member's 新上架. **Minor now; must fix before any further crawl or hosting.** Fix: compute `watched` from each target's first *complete* sweep end (`crawl_runs`), not from a global epoch + 24 h — radar ← store run data. Log under def L4 day-1 appearance.
- **R2 — Home 最新上架 is all GOYARD; Browse `sort=newest` pages 1–3 are all 98 GOYARD listings, then HERMES.** Home's five are GOYARD's top five on the shop's newest-first list. Each is individually defensible, but none is shown as shop-wide newest — **suboptimal, not untruthful**. Browse is weaker: GOYARD's page-2 stock (its oldest) is ordered ahead of every other brand's newest. That cross-brand order comes from the crawl schedule (GOYARD swept first), not from shop truth. No date or figure displayed is false (every card's 上架 is 2026/10/13), but the ordering implies a recency the data cannot support. **Minor, orchestrator call.** Cheap fix inside radar, no store change: in the initial window, rank by offset from the item's own target's first `first_seen` (≈ page depth, since every brand pages at the same `pageSize`), so each brand's newest interleave at the top. The alternative is to log it as def L4 day-1 appearance.

### Re-audit verdict

F1 (b), F2, F3, F4, F6 **Verified**. Audit status → **Complete (Accepted)**. Still open:
- **F5** — canvas pass 2 (unchanged).
- **R1** — must fix before further crawling or hosting.
- **R2** — orchestrator/def call.
- **Plan provisional items** (A1, A2, A8, A9 …) — still against def L3, for the orchestrator to log.
