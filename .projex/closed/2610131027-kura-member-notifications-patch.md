# Patch: Kura member state + notifications

> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 6
> **Status:** Complete
> **Author:** Claude Opus 5.5 (executor, sprint subagent)
> **Parent:** 2610131010-kura-member-notifications-plan.md
> **Directive:** Execute 2610131010-kura-member-notifications-plan.md in full (patch-projex plan-execution mode, sprint iteration 6)
> **Source Plan:** 2610131010-kura-member-notifications-plan.md (Steps 1–4, all)
> **Result:** Success
> **Commit:** `f92eea3` (code + README) on `projex/sprint/2610130524-kura-interactive-mock`
> **Related Projex:** 2610131010-kura-member-notifications-plan.md (source) | 2610130524-kura-interactive-mock-sprint.md (iteration 6) | 2610130449-kura-mock-def.md (Mocked: feed + watchlist; L3 open)

---

## Summary

Header 登入 → demo member 林小姐 in-app (no page load): member header (bell + unread badge, desktop heart 追蹤清單, 我的帳號 pill), 通知 dropdown (desktop popover / ≤767 fixed sheet), one sessionStorage watchlist store behind every 追蹤 toggle. `GET /api/member` picks the watchlist server-side (A2) and emits only crawl-observed events on it (A3), each carrying the radar `Item` identical to `/api/listings/{id}`. Worktree DB feed = 2 × 新上架 (LV `M51130` アルマPM モノグラム, CHANEL `A01112` マトラッセ Wフラップ), both unread, badge 2 — as plan-time probe predicted.

---

## Changes

### Radar

**File:** `internal/radar/radar.go` — Modified
- `IsDeal(it *Item)` (`:451`) — predicate extracted verbatim from `Search`; `Search` now `f.OnlyDeals && !IsDeal(it)` (`:420`). Constants unchanged (`0.05`, `anomalyPct`, `dealScore`, `confidenceWeight` untouched; Overview's 10% count left separate, noted in doc comment).
- `Engine.Gone(limit)` (`:534`) — non-active with `GoneAt`, `GoneAt` desc, tie by id, limit > 0 caps; copies (mirrors `Drops`).

**File:** `internal/radar/radar_test.go` — Modified
- `TestIsDeal` (`:179`): nil ✗ | 0.05 high ✓ | 0.049 ✗ | 0.2 medium ✓ | low ✗ | anomaly ✗.
- `TestGone` (`:199`): gone day 3 + day 6 + active → `Gone(10)` = [g6, g3]; `Gone(1)` = [g6].

### Server

**File:** `internal/web/member.go` — Created
- `member` handler: models — per `memberBrands` (`LOUIS VUITTON, HERMES, CHANEL`), first of `Models(b, 10)` whose newest active listing (`newestActive`: max `FirstSeen`, tie smaller id) passes `radar.IsDeal` → `model:{brand}/{group_key}` + `new` event at `FirstSeen`; stop at 2. Listings — candidates `Drops("", 2)` → `Gone(2)` → per-brand `Search{Brand, OnlyDeals, Limit 1}`, dedupe, stop at 2; `drop` event when active + `Drop`, `gone` when non-active + `GoneAt`, else none. `bare()` strips `History`. Events sorted `At` desc, then item id. Non-nil slices → JSON `[]`.
- Header comment names rejected alternative (canned sample notifications contradict real listing pages).

**File:** `internal/web/server.go` — Modified: route `GET /api/member` (`:51`).

**File:** `internal/web/server_test.go` — Modified
- `TestStaticRoutes`: `/js/member.js` (body `signIn`), `/js/notify.js` (`initNotify`), `/` contains `k-member`.
- `TestMemberAPI` (`:176`): own scheme `wm`; m1 = 3 comps 118k/120k/122k + newest 90k (+1h, deal medium); m2 listing 120k → 100k (+2h). Asserts persona; watchlist = [`model:LOUIS VUITTON/{group_key from /api/models}`, `wm:cut`, `wm:new`] (top deal watched, no event); events = [drop wm:cut, new m1] newest first; drop TWD 24000 → 20000; **each event item = `/api/listings/{id}` item** (`price_twd`, `status`, `fair`, `drop`, no history). Then `MarkGone(wm:cut)` → events = [gone wm:cut, new m1], invariant re-checked. Empty store → `"watchlist":[]`, `"events":[]`.

### Client

**File:** `internal/web/static/js/member.js` — Created: the one store. `sessionStorage` `kura.watch` = `{member, tracked[], read[]}`, try/catch (memory fallback); `onChange`, `isTracked`, `trackedCount`, `setTracked`, `member`, `eventID` (`{type}:{item.id}`), `isRead`, `markRead`, `signIn` (watchlist ∪ guest picks; events after the 2 newest pre-marked read). Rationale comments: localStorage rejected (no sign-out); replace-on-sign-in rejected (guest hearts would clear).

**File:** `internal/web/static/js/card.js` — Modified: in-memory `Set` removed; imports store, re-exports `isTracked` (detail/model/advisor untouched); `paintTrack`/`paintTracks` (`:87`, `:100`) repaint every `[data-track]` from the store, subscribed via `onChange(paintTracks)` (`:104`); click listener → `setTracked(id, !isTracked(id))` (`:112`). Existing reasoning comment kept, extended for sign-in repaint + advisor.

**File:** `internal/web/static/js/notify.js` — Created: `initNotify()` — sign-in on `[data-signin]` (preventDefault, `/api/member`, `signIn`, show header, focus bell; failure → `console.warn`, guest header kept); `paint()` on every store change (badge, bell aria-label `通知，N 則未讀` / `通知`, `追蹤清單（N）`, list or 「目前沒有通知。」); copy per A5 through `valuation()`; `setOpen` (h2 focus on open, bell focus on Escape); close on bell toggle, Escape, outside pointerdown, item click, popstate; 全部標為已讀; reload as member → show header immediately, refetch feed.

**File:** `internal/web/static/js/util.js` — Modified: `ago(iso, now)` (`:30`) per A6 (Taipei calendar dates; `HM` 24 h formatter).

**File:** `internal/web/static/js/app.js` — Modified: `initNotify()` import + call (`:8`, `:111`); header-search comment no longer claims tracked listings live only in memory.

**File:** `internal/web/static/js/nav.js` — Modified: `#` comment now names 加入會員, 追蹤清單, 通知設定 (登入 no longer a `#` step).

**File:** `internal/web/static/index.html` — Modified: `data-signin` on 登入 (`:28`); `.k-member` block (`:31`) — canvas markup (bell/heart/chevron SVG paths verbatim), styles → classes.

**File:** `internal/web/static/kura.css` — Modified: `/* Member header + 通知 */` block (`:133`) with canvas `isMember` / `notifStyle` / `notifs` values; ≤767 (`:543`): hide `.k-watch` + `.k-acct__name`, fixed sheet `left 8 right 8 top 64` + `max-height: calc(100dvh - 72px)` guard (comment names the clipped-footer consequence). Item ring inset (`outline-offset: -2px`) — panel `overflow: hidden` would clip an outside ring.

**File:** `README.md` — Modified: API table row `GET /api/member` (deviation: not in plan Key Files; added so shipped docs list every endpoint).

---

## Verification

**Method:** `go vet` / `go test`; hygiene grep; constants diff; `/api/member` vs `/api/listings/{id}` script on worktree `kura.db` (serve :8090, no crawl flags); browser pane matrix N1–N14 (tabs 1280 × 900 / 375 × 812, OS dark + `?theme=light|dark`); empty DB (scratchpad, serve :8092). Servers stopped by PID (90328, 97012); scratch DB + binary deleted.

**Result:**
```
go vet ./...            clean
go test ./...           ok advisor | radar (TestIsDeal, TestGone) | brandoff | store | web (TestMemberAPI, TestStaticRoutes)
hygiene grep            none (exit 1)
radar.go diff           only the 0.05 predicate moved into IsDeal (same comparisons, De Morgan); no constant changed
/api/member (kura.db)   watchlist [model:LOUIS VUITTON/M51130, model:CHANEL/A01112, brandoff:2101220216016, brandoff:2101220285869]
                        new M51130 brandoff:2101219356990 00:42:32 ランクB NT$14,129 deal 0.125 high  SAME as /api/listings
                        new A01112 brandoff:2107601067062 00:36:28 ランクA NT$181,653 deal 0.25 high  SAME as /api/listings
```

| # | Result | Evidence |
|---|---|---|
| N1 | ✓ | `__probe` survives; guest hidden, member shown; focus = bell; badge `2`; aria-label `通知，2 則未讀`; pill 林 / 林小姐 |
| N2 | ✓ | Home + Browse `deals=1&brand=LOUIS VUITTON`: only `brandoff:2101220216016` heart pressed (Browse: first card) |
| N3 | ✓ | popover rect 382 (380 + borders) wide, right = bell right + 8, top = bell bottom + 10; 2 unread items newest first; `新的一件 ランクB，NT$14,129，低於行情 13%` / `… ランクA，NT$181,653，低於行情 25%`; footer 追蹤清單（4）+ 通知設定; focus h2 |
| N4 | ✓ | real click → `/listing/brandoff%3A2101219356990` in-app; price NT$14,129, badge 低於行情 13%, 上架 2026/10/13 (label 今天 00:42), 售價變化 「上架以來還沒有調價」; badge → 1, panel closed. Item 2: NT$181,653 / 25%, badge hidden |
| N5 | ✓ | `/model/LOUIS%20VUITTON/M51130`, `/model/CHANEL/A01112`: 追蹤中 pressed; notified listing among 在售 |
| N6 | ✓ | real click 全部標為已讀: bg + dot transparent, no 未讀 sr, badge hidden, aria-label `通知` |
| N7 | ✓ | Escape (real key) → closed, focus bell, `aria-expanded=false`; outside click → closed; Back → closed |
| N8 | ✓ | model 追蹤 off → 3; card heart → 4, all copies pressed; detail 追蹤這件 off → 3; advisor (mock) 追蹤 → 追蹤中, 4, in store; feed stays 2 |
| N9 | ✓ | reload: member header, read state + hearts kept; new tab: 登入 / 加入會員, store null |
| N10 | ✓ | guest heart survives reload; 登入 → still pressed, 追蹤清單（5）, badge 2 |
| N11 | ✓ | 375 light/dark: header one row (logo 16–121, search 201–241, member 245–359), panel fixed 8/64/367, heart + name `display:none`, `scrollWidth 375 = innerWidth`; dark badge ring = `--k-paper` #1B1815; 1280 light popover screenshot = canvas `Notifications`; OS dark (colorScheme) run on tab 1 |
| N12 | ✓ | real keys: Enter bell → h2 focus; Tab → readall, item 1, item 2, 追蹤清單, 通知設定 — all `:focus-visible` with 2px ring (items inset); Enter on item → navigates, marks read |
| N13 | ✓ | :8092 empty DB: `{"watchlist":[],"events":[]}`; member header, no badge, 「目前沒有通知。」, 追蹤清單（0） |
| N14 | ✓ | 加入會員 (guest), 我的帳號, heart icon, 追蹤清單（N）, 通知設定: path unchanged, view not re-rendered (`#main` probe kept); in-panel `#` links close the panel (popstate) |
| 4.4 | ✓ | `read_console_messages` onlyErrors: none on all 3 tabs; page text lacks `NT$27,854`, `NT$104,955`, `A33814`, `NT$26,239` |
| extra | ✓ | `ago()` branches (1 分鐘前 / 45 分鐘前 / 4 小時前 / 今天 00:42 / 昨天 / 10/11); `/dev/cards.html` renders 19 cards on the shared store |

**Status:** PASS

### Acceptance Criteria

| Criterion | Result | Evidence |
|---|---|---|
| S1 Entry | ✓ | N1, N9, N14 |
| S2 Member header | ✓ | N1, N6, N11 |
| S3 Dropdown | ✓ | N3, N7, N11, N12 |
| S4 Read state | ✓ | N3, N4, N6, N9 |
| S5 One store | ✓ | N2, N5, N8, N10 |
| S6 No contradiction | ✓ | `TestMemberAPI` (drop + gone + new invariant); `/api/member` vs `/api/listings` SAME; N4, N5 |
| S7 Server | ✓ | `TestIsDeal`, `TestGone`, `TestMemberAPI`; `Search` callers' tests unchanged and green |
| S8 Responsive/themes | ✓ | N11, N12; 0 console errors; targets 40 px (readall / footer links min-height 24) |
| S9 | ✓ | vet / test green; hygiene grep none |

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|----------|-------------|-------------|
| 2610131010-kura-member-notifications-plan.md | Source plan (all 4 steps) | Steps marked `[PATCHED]`, Related Projex partial-execution line, Status → Complete, moved to `closed/` |
| 2610130524-kura-interactive-mock-sprint.md | Sprint nav | Not edited (orchestrator owns nav) |
| 2610130449-kura-mock-def.md | Def (L3 open items) | Not edited — provisional choices below for orchestrator to log |

---

## Notes

- **Deviations:** (1) `README.md` API row added (not in plan file list; stale-doc rule). (2) `TestMemberAPI` extends the plan's case with a `MarkGone` phase so the `gone` path and its invariant are tested, not only `drop`/`new`. (3) `TestStaticRoutes` body probes `signIn` / `initNotify` (plan named only status + type). (4) `member.go` returns a typed `memberView` struct instead of `map` — same JSON keys. (5) Readall button + footer links carry `min-height: 24px` (S8 target floor; canvas draws 4px padding only).
- **Observed, not changed:** 新上架 title uses the listing's `model_name` (`マトラッセ Wフラップ`) while the model page heading uses the group's longest name (`マトラッセ Wフラップ クルーズコレクション`) — both real, per A5 (`{model}` = item `model_name || title`); a copy pass may prefer the model page's name → def § Canvas gaps.
- **Item click order:** the panel's click handler marks read (list repaints) before the router's document listener navigates; the router resolves the link from the detached target's ancestors — verified in N4/N12; comment in `notify.js` states the dependency.
- **Provisional (for orchestrator to log, unchanged from plan Open Questions):** A1 entry, A2 watchlist composition/brands, A4 新上架 = newest *seen* on day-1 data, A5 copy for non-below states, A6 time rule, A7 two-unread, A8 static feed, A9 session scope, A11 photo thumb, A12 empty line, A14 dark × notifications token-driven.
- Guest-flash risk (plan Risks) observed as designed: member header appears after module load on reload.
