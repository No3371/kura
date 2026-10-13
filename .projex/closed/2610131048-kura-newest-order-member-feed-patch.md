# Patch: Kura one 最新上架 order + member feed fixes

> **Status:** Complete
> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 6
> **Author:** Claude Opus 5.5 (patch, sprint subagent)
> **Parent:** 2610131037-kura-member-notifications-audit.md
> **Directive:** Orchestrator (direct): fix audit F1 (option b), F3, F2, F4, F6; leave F5 (canvas pass 2 copy).
> **Source Plan:** Direct (findings of 2610131037-kura-member-notifications-audit.md)
> **Result:** Success
> **Related Projex:** 2610131037-kura-member-notifications-audit.md (findings F1–F4, F6) | 2610131027-kura-member-notifications-patch.md (patched work) | 2610131010-kura-member-notifications-plan.md (S6 premise) | 2610130449-kura-mock-def.md (feed = Mocked; L4 day-1 appearance) | 2610130524-kura-interactive-mock-sprint.md
> **Commits:** `a88cee9` fix(radar) — F1 + F3 | `4d68086` fix(web) — F2 + F4 + F6

---

## Summary

First-sweep `first_seen` is crawl order, and sources sweep list pages newest first (`crawl.Source.Sweep`; BRAND OFF `sort=Number1`), so "latest `first_seen`" picked each model's *oldest* listing as 新上架 (F1) and filled Home 最新上架 / Browse `sort=newest` with the deepest pages (F3). One server-side order now serves both: listings watched appearing (`FirstSeen.After(epoch+24h)`, the existing days-to-sell rule) first, latest first; initial-crawl listings after, earliest `first_seen` first; ties → id. Member feed stays Mocked; comment + README now say what is true. Feed UI: shows/counts only still-tracked entries (F2); no 「目前沒有通知。」 before the reload refetch answers, failure line on error (F4); `nav.js` comment re-flowed (F6).

F3 size: small — one comparator swap in `sortItems`, no schema change; Home/Browse/advisor all go through `Engine.Search`.

---

## Changes

### Radar order (F1 + F3)

**File:** `internal/radar/radar.go` — Modified
- `watched` (l.480): the epoch+24h rule, now also used by `stats` days-to-sell (l.673) — one source.
- `newer` (l.495): the one 最新上架 comparator; rationale comment names the rejected "latest first_seen throughout" and the lost in-page order.
- `NewestActive` (l.511): exported pick for `member.go` (RLock; `newer` is lock-free for `Search`, which already holds RLock).
- `sortItems` → method; `newest` uses `e.newer` (l.530). Callers: `Search` only (l.428) → Home shelf, Browse, advisor `sort: newest`.

**File:** `internal/web/member.go` — Modified
- Header comment: 降價 / 已找到新主人 = observed by the crawl; 新上架 = model's newest active listing in radar 最新上架 order, watched-appear if any, else the shop's newest from the initial crawl "which Kura never saw appear". Removed "every notification is something the crawl actually observed".
- `newestActive` deleted → `s.Radar.NewestActive(items)` (l.63).

**File:** `README.md` — Modified
- New `### Newest listings` (l.96): the order rule (rationale promotion). `GET /api/member` row (l.177): "price cuts and sales the crawl observed … and each tracked model's newest listing".

**Tests:**
- `internal/radar/radar_test.go` `TestNewestOrder` (l.242): mixed data → `[w2 w1 i1 i1b i2]` (watched latest-first, initial earliest-first, tie by id), `NewestActive` skips gone; first-sweep only → `NewestActive` = `i1`, `newest` limit 1 = `i1`. Old rule fails both.
- `internal/web/server_test.go` `TestMemberAPI`: t0 moved to now−30h; m1's `wm:new` at t0+25h (post-epoch case); new HERMES model h1 all in the initial window — `hnew` (t0, first crawled) vs `hold` (t0+20m, also a deal) → pick must be `wm:hnew` (l.209, l.257). Old rule picked `hold`.

### Feed UI (F2 + F4 + F6)

**File:** `internal/web/static/js/notify.js` — Modified
- `shown()` (l.71) = events with `isTracked(ev.key)`; `paint` lists + counts only these; 全部標為已讀 marks `shown()` (l.130). Hide-only — never adds. Header comment states the rule + rejected alternative (unpressed toggle beside 「你追蹤的型號 …」).
- `feed` state (l.68) `loading | ready | failed`: list blank while loading (l.90), `暫時無法載入資料，請稍後再試。` on refetch failure (l.91); sign-in path sets `ready` before `signIn`.

**File:** `internal/web/static/js/nav.js` — Modified — header comment re-flowed (no dangling "The").

---

## Verification

**Method:** `go vet ./...` + `go test -count=1 ./...`; live server (patched build, worktree `kura.db`, `127.0.0.1:8094`, no crawl flags) + read-only scratch copy for SQL; built-in browser.

**Result:**
```
go test: advisor ok | radar ok | brandoff ok | store ok | web ok
/api/member: model:LOUIS VUITTON/M51146 → brandoff:2101220315146 (00:36:43; LV run 4 began 00:36:31)
             model:CHANEL/A67085      → brandoff:2106800652536 (00:31:50; CHANEL run 3 began 00:31:47)
SQL (active, ORDER BY first_seen, id LIMIT 1): A67085 → 2106800652536 ✓ | M51146 → 2101220315146 ✓
  (previous LV pick model M51130: max first_seen 00:42:32 = deepest page; its earliest 2101220324117 is not a deal → selection moves to M51146)
Home 最新上架 (DOM): 2100301255937 2101219731339 2101219811840 2101219850221 2101220005108
  = SQL ORDER BY first_seen, id LIMIT 5 (GOYARD run 1 page 1, 00:25:25 = epoch) ✓
  = /api/listings?sort=newest&limit=30 first 5 ✓ (Home ids == Browse first ids)
F2 model page M51146, dropdown open, toggle off via keyboard (no pointerdown): panel stays open,
  items 2 → 1 (only ボーイシャネル スモール), badge 2 → 1, aria 通知，1 則未讀, 追蹤清單（3）; re-track → 2 / badge 2
F4 fresh notify.js instance, /api/member rejected after 800 ms: during load list innerHTML "" + badge hidden;
  after → 「暫時無法載入資料，請稍後再試。」 ; normal reload shows both events
Console errors: none
```

**Status:** PASS

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|----------|-------------|-------------|
| 2610131037-kura-member-notifications-audit.md | Source findings | None (point-in-time record); F1 (b), F2, F3, F4, F6 resolved here; F5 left for canvas pass 2 |
| 2610130524-kura-interactive-mock-sprint.md | Sprint | Not edited (orchestrator owns nav) |

---

## Notes

- **Residual — late-but-within-24h appearances:** the watched-appear gate is `epoch + 24h` (directive: reuse the radar rule). On today's DB the one CHANEL listing first seen in run 9 (01:41, an incomplete later sweep) is genuinely new but sits inside the window, so it sorts as the *oldest* initial-crawl listing. A per-target "first complete sweep" boundary from `crawl_runs` would fix it; it needs radar ← store run data, out of patch scope. Candidate for def L4 day-1 appearance.
- **Cross-brand order:** earliest-first across the whole initial window puts the first-crawled brand's page 1 (GOYARD) ahead of every other brand — Home 最新上架 is currently five GOYARD bags. Truthful (GOYARD's newest on the shop), but not a shop-wide newest. Interleaving by offset from each target's sweep start is the alternative; orchestrator/def call.
- In-page order is lost (≈50 listings share a second; source id is not monotonic with page order — 46% of pairs inverted in an LV sample), so ties fall back to id; README says so.
- `store.PendingDetails` still orders by `first_seen DESC` — detail-fetch priority, not a 最新上架 surface; untouched.
- Feed At for initial-crawl 新上架 = crawl first_seen, so its 「今天 00:31」 is when Kura first saw it, not the shop's listing time — unchanged from before, same as the detail page's 上架 date.
