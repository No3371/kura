# Kura member state + notifications

> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 6
> **Status:** Complete
> **Author:** Claude Opus 5.5 (planner, sprint subagent)
> **Parent:** 2610130524-kura-interactive-mock-sprint.md
> **Source:** Sprint derive, iteration 6 (sprint nav § Current Position + Revision Log, commit `780303d`)
> **Related Projex:** 2610130524-kura-interactive-mock-sprint.md (blocking map: member entry decided, data constraint) | 2610130449-kura-mock-def.md (Mocked = notification feed + persona watchlists; one source per stat; L3 personas / notification types / static vs reactive open) | 2610130420-kura-roadmap-nav.md (Phase 4 member layer) | 2610130913-kura-browse-page-plan.md (`nav.js` `#` handling, header search) | 2610130722-kura-model-page-plan.md (model track key) | 2610130815-kura-advisor-panel-plan.md (advisor 追蹤 button)
> Partial Execution: Steps 1–4 (all) completed via 2610131027-kura-member-notifications-patch.md (commit `f92eea3`)
> **Worktree:** Yes — existing sprint worktree `kura/.projexwt/2610130524-kura-interactive-mock`, branch `projex/sprint/2610130524-kura-interactive-mock`; execute via patch-projex plan-execution mode, no new branch

---

## Summary

Make the canvas member state walkable (pinned 1791839497-46ec; `project/Main.dc.html` `isMember` header block; artboards `Notifications` 1280 + `NotificationsMobile` 375, `user: member`, `notif: open`). Header 登入 → demo member 林小姐 (provisional entry): member header (bell + unread badge, desktop heart 追蹤清單 link, 我的帳號 pill), 通知 dropdown (desktop popover / mobile fixed sheet), one session-scoped watchlist store behind every existing 追蹤 toggle. The feed is **mocked in who tracks what, real in what happened**: the persona's watchlist is picked server-side from real listings/models, and each notification is an event the crawl actually observed on a watchlist entry (新上架 / 降價 / 已找到新主人), carrying the same radar `Item` the listing page renders — so a notification can never contradict its linked page.

**Scope:** `internal/radar` (exported deal predicate, `Gone`), `internal/web` (`GET /api/member` + tests; static shell: `index.html`, `kura.css`, new `member.js` + `notify.js`, `card.js`, `util.js`, `app.js`, `nav.js` comments). One `.projex` scope (repo root).
**Estimated Changes:** 11 files — `radar.go`, `radar_test.go`, `server.go`, `server_test.go`, `index.html`, `kura.css`, `card.js`, `util.js`, `app.js`, `nav.js` edited; `member.go`, `member.js`, `notify.js` added.

---

## Objective

### Problem / Gap / Need
The member flow is canvas-designed (header variant + 2 artboards) but unreachable: header 登入 / 加入會員 are `#` dead ends; tracking toggles (`card.js` in-memory `tracked` Set, shared by card / detail 追蹤這件 / model 追蹤這個型號 / advisor 追蹤) have no member-scoped store and vanish on reload. The DB holds one sweep (15,367 active, 0 gone, 1 price point per listing, 0 `was_price_jpy`) → no real 降價 / 已找到新主人 exists; canvas `KURA_NOTIFS` are layout samples (降價 NT$27,854 → NT$26,239 …) whose numbers would contradict any real listing page they linked to.

### Success Criteria
- [x] **S1 Entry:** clicking header 登入 (desktop + 375) swaps the guest block for the member header without a page load; focus lands on the bell; 加入會員 stays `#` (inert). Reload keeps member state (same tab); a new tab starts as guest.
- [x] **S2 Member header** (canvas `isMember`): bell button `aria-expanded`, `aria-controls`, `title="通知"`, aria-label `通知，N 則未讀` (N > 0) / `通知` (N = 0), accent badge with N (hidden at 0); desktop-only heart 追蹤清單 icon link (`#`, inert); 我的帳號 pill (button, initial 林 in brass ring, name 林小姐 desktop-only, chevron; no menu — inert).
- [x] **S3 Dropdown** (canvas `notifOpen`): `role="dialog"` `aria-label="通知"`; h2 通知 + 全部標為已讀; `<ol>` items (unread dot + sr 未讀, 48 px mat thumb, type in accent-ink / ink-3 for 已找到新主人, time, title ellipsis, body; whole item = link to `/listing/{id}`); footer 追蹤清單（N） (N = store size) + 通知設定 (both `#`, inert). Desktop: absolute popover `right -8px; top calc(100% + 10px); width 380px`. ≤ 767: `position: fixed; left 8px; right 8px; top 64px`. Opens/closes on bell; closes on Escape (focus → bell), outside pointerdown, item click, popstate.
- [x] **S4 Read state:** at sign-in the two newest notifications are unread, the rest read; item click marks it read; 全部標為已讀 marks all; badge + aria-label + dots update live; persists with the session.
- [x] **S5 One watchlist store:** card heart, detail 追蹤這件, model 追蹤這個型號, advisor 追蹤 read/write one store; sign-in adds the persona's watchlist to whatever the guest already tracked; every visible `[data-track]` control repaints on any change (incl. sign-in); count in 追蹤清單（N） follows.
- [x] **S6 No contradiction (data constraint):** every notification's figures (price, 降價 from → to, 低於行情 %, grade, status) come from the radar `Item` in `/api/member`, identical to `/api/listings/{id}` for that id; wording maps through `card.js` `valuation()` (the badge function); a 降價 appears only for a listing whose own history shows that cut, 已找到新主人 only for a gone listing, 新上架 only for a tracked model's newest active listing. On the worktree DB: feed = 2 × 新上架 (LV `M51130` アルマPM, CHANEL `A01112` — plan-time probe), both unread, badge 2.
- [x] **S7 Server:** `GET /api/member` → `{name, initial, watchlist[], events[]}` (watchlist ≤ 4 keys; events newest first); radar deal predicate exported and used by `Search` (behaviour unchanged); `Engine.Gone(limit)` most recently gone first. Tests cover both + the S6 invariant.
- [x] **S8 Responsive/themes:** 1280 + 375 × light/dark (dark token-driven, no artboard); no horizontal page scroll at 375 (member header fits one row); targets ≥ 24 px; visible focus ring; 0 console errors.
- [x] **S9** `go vet ./...` + `go test ./...` green; hygiene grep clean.

### Out of Scope
- 加入會員, persona switch/menu, sign-out, account menu content, watchlist page, 通知設定 — sprint Deferred (def § Canvas gaps / L3); controls stay `#` / inert as canvas draws them
- Def L3 decisions: persona count; sign-up vs sign-in-as-persona wording; notification type set (no 低於行情 type built); static vs reactive feed (feed is static per sign-in — canvas evidence); all choices here marked provisional
- Server-side persistence / real notification delivery (def Is-not)
- Advisor awareness of the member (advisor unchanged)
- F1 / F6 deal threshold + anomaly ordering — the exported predicate keeps today's constants
- Footer pages incl. 行情怎麼算 (next derive)

---

## Context

### Current State
- **Header** (`index.html`): `.k-guest` = `<a href="#" class="k-link">登入</a>` + `<a href="#" class="k-join">加入會員</a>`; mobile CSS hides `.k-join` and `.k-search` (≤ 767). No member markup.
- **Tracking** (`card.js`): module-level `const tracked = new Set()` ("in-memory only; persistence belongs to the member layer"); `isTracked(id)` exported (imported by `detail.js`, `model.js`, `advisor.js`); one document click listener toggles `[data-track]` and repaints every copy of that key (icon-only → aria-label/title; text-labelled → `[data-track-text]` text, off label = `dataset.trackText || "追蹤這件"`). Model key `model:${s.brand}/${s.group_key}` (`model.js`, comment: one store for both kinds).
- **Router/nav:** `app.js` document click → in-app `go()` for `a[href^='/']`; header search intercepted; comment says a native GET would "drop what lives only in memory (tracked listings, the advisor conversation)". `nav.js` header comment: popstate fires for `#` steps "(the header's 登入 and 加入會員 are href="#")".
- **Server:** `server.go` routes `/api/overview|listings|listings/{id}|drops|models|models/{brand}/{key}|crawl|chat`; `s.api` refreshes the radar then JSON-encodes. Radar: `Search` inlines the deal predicate `Fair != nil && DealPct >= 0.05 && Confidence != "low" && !Anomaly`; `Drops(brand, limit)` = active with `Drop` (last price < previous), newest cut first; `Models(brand, limit)` most-listed first; `Model(brand, key)` stats + items; `Get(id)` item with history. No gone listing accessor. `Item.FirstSeen`, `GoneAt`, `Drop{FromTWD, ToTWD, At}`.
- **Data** (worktree `kura.db`, plan-time probe via local serve, read-only): 15,367 active, first_seen 2026-10-13 00:25–01:41 (+08), `/api/drops` 0, gone 0. Newest active listing per top-10 model passing the deal predicate: LV — `M51130` アルマPM モノグラム (newest `brandoff:2101219356990`, ランクB, NT$14,129, deal 0.125 high), `M51165`, `M51136`, `M51145`; HERMES — none (every name-keyed group is low confidence); CHANEL — `A01112` マトラッセ Wフラップ クルーズコレクション (newest `brandoff:2107601067062`, ランクA, NT$181,653, 0.25 high), `A67086`, `A01113`. Top deal per brand (`deals=1&limit=1`): LV `brandoff:2101220216016` カルトシエールGM 0.50 | HERMES `brandoff:2101220285869` ピコタンロックPM 0.307 | CHANEL `brandoff:2120900081669` マトラッセ 0.495.
- **Canvas** (`Main.dc.html`): member header markup quoted in the nav blocking map; `KURA_NOTIFS` 3 items (降價 unread '2 小時前' · 新上架 unread '今天 09:12' · 已找到新主人 read '昨天'), title forms `ネヴァーフルMM · M40156` / `你追蹤的型號 ピコタンロックPM` / `マトラッセ · A33814`, bodies `NT$27,854 → NT$26,239，現在低於行情 32%` / `新的一件 ランクA，NT$104,955，低於行情 31%` / `你追蹤的這件已經賣出了。`; thumb = bag icon on mat; unread row bg `--k-paper`, read transparent; dot accent / transparent. Badge `2`, `aria-label="通知，2 則未讀"` hard-coded. Detail fine print 「追蹤後，降價或售出時會通知你。」; model 「有新上架或低於行情時通知你」.

### Key Files

| File | Role | Change Summary |
|------|------|----------------|
| `internal/radar/radar.go` | index | `IsDeal` (extracted predicate, used by `Search`); `Gone(limit)` |
| `internal/radar/radar_test.go` | tests | `IsDeal` cases; `Gone` order/limit |
| `internal/web/member.go` | new | demo persona: watchlist pick + real events; `member` handler |
| `internal/web/server.go` | API | route `GET /api/member` |
| `internal/web/server_test.go` | tests | `/api/member` shape + no-contradiction invariant; static routes for new JS |
| `internal/web/static/index.html` | shell | `data-signin` on 登入; `.k-member` header block + 通知 panel |
| `internal/web/static/kura.css` | styles | member header, badge, 通知 panel (desktop/mobile) |
| `internal/web/static/js/member.js` | new | the one watchlist store (sessionStorage) |
| `internal/web/static/js/notify.js` | new | sign-in, member header, dropdown |
| `internal/web/static/js/card.js` | card | store-backed tracking; repaint on store change |
| `internal/web/static/js/util.js` | utils | `ago()` notification time |
| `internal/web/static/js/app.js` | router | `initNotify()`; header-search comment |
| `internal/web/static/js/nav.js` | nav | comment (登入 no longer `#` step) |

### Dependencies
- **Requires:** sprint branch HEAD `780303d` (iterations 1–5).
- **Blocks:** nothing in-sprint. A later def L3 answer (persona count, sign-in wording, reactive feed) replaces the entry / persona pick without touching the member surfaces.

### Constraints
- HANDOFF §3.6 no build step → native ES modules; §3.5 375 + 16 px gutter; §3.7 targets ≥ 24 px, visible focus.
- Canvas governs look/IA/copy/states; def governs data: feed + watchlist = Mocked (unlabelled), but no mocked figure — every number shown is the radar's (one source per stat).
- Def L3 items stay open; provisional choices named in Assumptions, out-of-scope controls inert.
- Deal predicate constants unchanged (F1/F6 open).
- Source Hygiene: no projex IDs in comments; non-obvious decisions name the rejected alternative; present tense.
- No crawls (serve without crawl flags; never `POST /api/crawl`); commits on the sprint branch in the worktree only.

### Assumptions
- **A1 — Entry (provisional, sprint derive 6):** 登入 signs in as the canvas's one member 林小姐 directly — no persona picker, no sign-in form. Def L3 (persona count; 以示範身分登入 wording) open → orchestrator logs.
- **A2 — Watchlist pick (mocked, provisional):** ≤ 4 keys = 2 models + 2 listings (canvas 追蹤清單（4）). Persona brands `LOUIS VUITTON, HERMES, CHANEL` (the canvas sample brands). Models: per brand in order, first of `Models(brand, 10)` whose newest active listing passes `IsDeal`; stop at 2. Listings: candidates in order — most recent real drops (`Drops("", 2)`), most recently gone (`Gone(2)`), then each persona brand's top deal (`Search{Brand, OnlyDeals, Limit 1}`); dedupe; stop at 2. Worktree DB → `model:LOUIS VUITTON/M51130`, `model:CHANEL/A01112`, `brandoff:2101220216016`, `brandoff:2101220285869`. Deterministic per DB state.
- **A3 — Events = observed facts on watchlist entries:** listing key with `Drop` → `drop` at `Drop.At`; listing key gone → `gone` at `GoneAt`; model key → `new` = its newest active listing at `FirstSeen`. Sorted `At` desc, then id. Item copied without `History`. A listing that is neither dropped nor gone yields no event. Types = canvas's three; 低於行情 type not built (def L3).
- **A4 — 新上架 on day-1 data:** every listing's `first_seen` is the first crawl, so "newest of the model" is the newest *seen*. The listing page states the same 上架 date (`detail.js` specs `ymd(first_seen)`), so nothing contradicts; with an hourly crawl it becomes genuinely new. Logged risk (def L4 day-1 appearance).
- **A5 — Copy:** canvas templates verbatim, figures from the item: drop title `{model} · {code}` / body `{twd(drop.from_twd)} → {twd(drop.to_twd)}` + `，現在低於行情 N%` when `valuation()` = below; new title `你追蹤的型號 {model}` / body `新的一件 {grade}，{twd(price_twd)}` + `，低於行情 N%` when below; gone title `{model} · {code}` / body `你追蹤的這件已經賣出了。`. `{model}` = `model_name || title`; ` · {code}` omitted when no code.
- **A6 — Time label (plan-authored rule, canvas shows the forms only):** < 1 h `N 分鐘前` (N ≥ 1); < 6 h `N 小時前`; same Taipei date `今天 HH:MM`; previous Taipei date `昨天`; else `M/D`.
- **A7 — Read state:** two newest unread at sign-in (canvas: 2 of 3 unread, newest first); item click → read; 全部標為已讀 → all. Feed is fetched again on reload; ids `{type}:{item.id}` not in the read set are unread (a new real event after a crawl shows unread).
- **A8 — Static feed (def L3 open):** toggling 追蹤 after sign-in changes the watchlist and its count, not the feed (canvas draws a static list beside working toggles).
- **A9 — Session store:** `sessionStorage` key `kura.watch` holds `{member, tracked[], read[]}` for guest and member alike; read/write wrapped in try/catch (memory-only fallback). Rejected: `localStorage` — no sign-out is designed (account menu deferred), so a reviewer would stay 林小姐 in every future visit. Guest hearts now survive reload (side effect; was memory-only).
- **A10 — Merge at sign-in:** persona watchlist ∪ guest picks. Rejected: replace — hearts the visitor just set would silently clear.
- **A11 — Thumb = real photo** in the 48 px mat (`image_url`, `PENDING_ICON` when `photo_pending`), as cards/detail replaced canvas icons; `aria-hidden`.
- **A12 — Empty feed** (no qualifying entries, e.g. an empty DB): list shows 「目前沒有通知。」 (plan-authored) → canvas gap. Sign-in failure (API error) leaves the guest header, `console.warn` only (no designed error).
- **A13 — Dropdown focus:** open → focus h2 (`tabindex=-1`, the advisor's heading-focus pattern); Escape → bell. Non-modal: Tab may leave the panel; outside pointerdown closes it.
- **A14 — Dark × notifications** token-driven (no artboard) → def § Canvas gaps tally.

### Impact Analysis
- **Direct:** header gains the member variant; tracking moves to `member.js`; new `/api/member`.
- **Adjacent:** `detail.js`, `model.js`, `advisor.js` import `isTracked` from `card.js` — kept as a re-export, files untouched; `/dev/cards.html` imports `card.js` (works; reads the same store). `Search` callers (Home, Browse, advisor `only_deals`) — predicate extracted verbatim. `nav.js` `#` handling unchanged (登入 now `preventDefault`s, other `#` links unchanged).
- **Downstream:** none outside the repo.

---

## Implementation

### Overview
Server first (tested predicate + gone accessor + member endpoint with the no-contradiction invariant), then the client store under the existing toggles, then header UI + dropdown, then the browser matrix.

### Step 1: Radar `IsDeal` + `Gone`; `GET /api/member` [PATCHED] → 2610131027-kura-member-notifications-patch.md

**Objective:** S6, S7 server side. **Confidence:** High. **Depends on:** None.
**Verify-Projex:** Encouraged

**Files:** `internal/radar/radar.go`, `internal/radar/radar_test.go`, `internal/web/member.go` (new), `internal/web/server.go`, `internal/web/server_test.go`

**Changes:**

```go
// radar.go — Before (in Search):
(f.OnlyDeals && (it.Fair == nil || it.Fair.DealPct < 0.05 || it.Fair.Confidence == "low" || it.Fair.Anomaly)) {

// After:
(f.OnlyDeals && !IsDeal(it)) {

// IsDeal reports whether a listing counts as below market: valued at least 5%
// under fair value, not low confidence, not flagged as an anomaly. It is the
// 價格最美 predicate — Home's shelf and count, Browse's switch, the advisor's
// only_deals and the demo member's tracked models all use it, so one change
// here moves them together.
func IsDeal(it *Item) bool {
	return it.Fair != nil && it.Fair.DealPct >= 0.05 && it.Fair.Confidence != "low" && !it.Fair.Anomaly
}

// Gone returns listings no longer on sale, most recently gone first.
func (e *Engine) Gone(limit int) []Item   // mirrors Drops: RLock; Status != "active" && GoneAt != nil; sort GoneAt desc; cap limit (>0); copies
```

`internal/web/member.go` (new):

```go
// The demo member is the one signed-in visitor the design draws. Kura has no
// users, so what she tracks is picked here from live listings; every
// notification is something the crawl actually observed on what she tracks,
// carrying the same Item the listing page shows. Rejected: canned sample
// notifications — a price cut or sale invented for a real, active listing
// would contradict that listing's own page.

var memberBrands = []string{"LOUIS VUITTON", "HERMES", "CHANEL"} // the brands the design's samples track

type memberEvent struct {
	Type string     `json:"type"` // drop | new | gone
	Key  string     `json:"key"`  // the watchlist entry it is about
	At   time.Time  `json:"at"`
	Item radar.Item `json:"item"`
}

func modelKey(brand, group string) string { return "model:" + brand + "/" + group } // same form as the model page's 追蹤 key

func (s *Server) member(r *http.Request) (any, int, error) {
	// models: for each memberBrands b, for m in s.Radar.Models(b, 10): _, items, _ := s.Radar.Model(b, m.GroupKey);
	//   newest = active item with max FirstSeen (tie: smaller ID); if radar.IsDeal(&newest) → watch modelKey, event {new, key, newest.FirstSeen, newest}; next brand. Stop at 2.
	// listings: candidates = Drops("", 2) ids, Gone(2) ids, then per brand Search(Filter{Brand: b, OnlyDeals: true, Limit: 1}) ids; dedupe; stop at 2.
	//   for each: it, _, _ := s.Radar.Get(id); Drop != nil && Status == "active" → {drop, id, Drop.At, it}; Status != "active" && GoneAt != nil → {gone, id, *GoneAt, it}.
	// Item.History = nil on every event item. Sort events At desc, then Item.ID.
	// return map{"name": "林小姐", "initial": "林", "watchlist": models ++ listings, "events": events} (non-nil slices → JSON [])
}
```

`server.go`: `mux.HandleFunc("GET /api/member", s.api(s.member))`.

Tests:
- `radar_test.go`: `TestIsDeal` (nil fair ✗; 0.05 high ✓; 0.049 ✗; low ✗; anomaly ✗); extend `TestDropsAndStats` or add `TestGone` — two gone rows (GoneAt day 3, day 6) + active → `Gone(10)` = [day 6, day 3], `Gone(1)` len 1. Existing `TestSearchFilters` unchanged and green.
- `server_test.go` `TestMemberAPI`: temp store (pattern of `TestListingAPI`, own scheme source `wm`), brand `LOUIS VUITTON`, model code `m1`: 3 comps grade A (~118k/120k/122k) at t0 + newest grade A 90k at t0+1h (deal, non-low — tighten comps if confidence comes out low); listing in model `m2` recorded at 120k then 100k one hour later (a real drop). Assert: 200; `name == "林小姐"`; `watchlist` contains `model:LOUIS VUITTON/{group_key}` (group_key as `/api/models` reports it for `m1`) and the dropped id, `len ≤ 4`; events = `new` (item = newest m1 listing) + `drop` (dropped id), newest first; **for each event, `GET /api/listings/{id}` item has equal `price_twd`, `status`, `fair` (deal_pct, confidence, anomaly) and `drop`** (S6 invariant); no event for watched listings without drop/gone. Empty store → `watchlist: []`, `events: []`.
- `TestStaticRoutes`: add `/js/member.js` and `/js/notify.js` (200, javascript); `/` body contains `k-member`.

**Rationale:** Server picks + events keep every figure on the radar's single computation; reusing the exported predicate means the persona's "below market" models follow F1 when it is decided. `Gone` lets 已找到新主人 appear once the crawl observes a sale, with no further code.

**Verification:** `go vet ./...`; `go test ./internal/radar ./internal/web`; local serve on `kura.db` → `curl /api/member` = A2 keys, 2 `new` events (M51130 first — newest seen 00:42 vs A01112 00:36), each item = `/api/listings/{id}` item.

**If this fails:** revert the five files; the shell is unchanged so far.

---

### Step 2: One watchlist store under the existing toggles [PATCHED] → 2610131027-kura-member-notifications-patch.md

**Objective:** S5 (store + repaint), A9, A10. **Confidence:** High. **Depends on:** None (client-only; Step 3 consumes it).

**Files:** `internal/web/static/js/member.js` (new), `internal/web/static/js/card.js`, `internal/web/static/js/app.js`, `internal/web/static/js/nav.js`

**Changes:**

```js
// member.js (new) — the one watchlist store.
// Keys: listing ids ("{source}:{id}") and "model:{brand}/{group_key}".
// Session-scoped (sessionStorage "kura.watch"): a reload keeps the walk going,
// a new tab starts as a guest. Rejected: localStorage — no sign-out exists, so
// a visitor would stay signed in on every later visit.
const KEY = "kura.watch";
const UNREAD = 2; // the design shows its two newest notifications unread
let state = load(); // { member: {name, initial} | null, tracked: Set, read: Set }
const subs = [];
// load(): JSON.parse(sessionStorage.getItem(KEY)) → state, try/catch → empty state
// save(): try sessionStorage.setItem(KEY, JSON.stringify({member, tracked: [...], read: [...]})) catch {}; then call each sub
export const onChange = (fn) => { subs.push(fn); };
export const isTracked = (key) => state.tracked.has(key);
export const trackedCount = () => state.tracked.size;
export function setTracked(key, on) { /* add|delete, save() */ }
export const member = () => state.member;
export const eventID = (ev) => `${ev.type}:${ev.item.id}`;
export const isRead = (ev) => state.read.has(eventID(ev));
export function markRead(evs) { /* add eventID of each, save() */ }
// signIn adds the persona's watchlist to the guest's picks (rejected: replacing
// them — hearts just set would silently clear) and leaves the newest UNREAD
// notifications unread.
export function signIn(m) { /* member = {name, initial}; tracked ∪= m.watchlist; read = new Set(m.events.slice(UNREAD).map(eventID)); save() */ }
```

```js
// card.js — Before:
const tracked = new Set(); // in-memory only; persistence belongs to the member layer
export const isTracked = (id) => tracked.has(id);
…document click: toggle `tracked`, then repaint every [data-track="{id}"]

// After:
import { isTracked, setTracked, onChange } from "./member.js";
export { isTracked }; // detail.js, model.js, advisor.js keep importing it from here
…
function paintTrack(b) { /* existing per-button body: aria-pressed; [data-track-text] text (on → 追蹤中, off → dataset.trackText || 追蹤這件) else aria-label/title TRACK_ON/OFF — from isTracked(b.dataset.track) */ }
export function paintTracks() { for (const b of document.querySelectorAll("[data-track]")) paintTrack(b); }
onChange(paintTracks); // a toggle, another copy of the key, or sign-in adding a watchlist
document click on [data-track]: setTracked(id, !isTracked(id)) — repaint happens via onChange
```
Keep the existing comment's reasoning (same key in several places; text vs icon buttons; model page shares the listener) attached to `paintTrack`.

- `app.js` header-search comment: "drop what lives only in memory (tracked listings, the advisor conversation)" → "drop the advisor conversation, which lives only in memory".
- `nav.js` header comment: "(the header's 登入 and 加入會員 are href="#")" → "(加入會員, the member header's 追蹤清單 and 通知設定 links are href="#")".

**Rationale:** One store, one paint path: sign-in can update every visible toggle without re-rendering the view; the `card.js` re-export avoids touching three views.

**Verification:** guest: heart on Home card → reload → still 追蹤中; detail 追蹤這件 of the same id in sync; model 追蹤這個型號 toggles; `sessionStorage.getItem("kura.watch")` reflects keys; new tab → none.

**If this fails:** revert the four files (card.js back to the in-memory Set).

---

### Step 3: Member header + 通知 dropdown [PATCHED] → 2610131027-kura-member-notifications-patch.md

**Objective:** S1–S4, S8. **Confidence:** Medium (canvas pixel match, focus handling). **Depends on:** Steps 1, 2.
**Verify-Projex:** Encouraged

**Files:** `internal/web/static/index.html`, `internal/web/static/kura.css`, `internal/web/static/js/notify.js` (new), `internal/web/static/js/util.js`, `internal/web/static/js/app.js`

**Changes:**

`index.html` — 登入 gains `data-signin`; after `.k-guest` (canvas markup, styles → classes):
```html
<div class="k-member" hidden>
  <div class="k-notif">
    <button type="button" class="k-iconbtn k-bell" aria-expanded="false" aria-controls="k-notif" aria-label="通知" title="通知">{bell svg: canvas paths}<span class="k-bell__count" aria-hidden="true" hidden></span></button>
    <section id="k-notif" class="k-notif__panel" role="dialog" aria-label="通知" hidden>
      <header><h2 tabindex="-1">通知</h2><button type="button" class="k-link k-notif__readall">全部標為已讀</button></header>
      <ol class="k-notif__list"></ol>
      <footer><a href="#" class="k-link" data-watch-count>追蹤清單（0）</a><a href="#" class="k-link">通知設定</a></footer>
    </section>
  </div>
  <a href="#" class="k-iconbtn k-link k-watch" aria-label="追蹤清單" title="追蹤清單">{heart svg 20px}</a>
  <button type="button" class="k-acct" aria-label="我的帳號"><span class="k-acct__initial" aria-hidden="true"></span><span class="k-acct__name"></span>{chevron svg 12px}</button>
</div>
```

`kura.css` — new block `/* Member header + 通知 */`, values from canvas `isMember` / `notifStyle` / `notifs`: `.k-member` flex center gap 8px; `.k-notif` relative; `.k-iconbtn` 40×40 flex-center ink-2, transparent, no border, radius 50%; `.k-bell[aria-expanded="true"]` bg `--k-paper-sunk`; `.k-bell__count` absolute top 4 right 3, min-width 16, h 16, pad 0 4, radius 999, accent / on-accent, 10px/16px, `box-shadow: 0 0 0 2px var(--k-paper)`; `.k-notif__panel` absolute `right:-8px; top:calc(100% + 10px); width:380px; z-index:40`, surface, 1px rule-strong, radius 4, `--k-shadow`, overflow hidden, text-align left; header flex space-between baseline, pad 16 18 12, bottom rule; h2 display 400 17px m0; readall 12px ink-2 transparent pad 4px 0; list reset, li bottom rule; `.k-notif__item` grid `8px 48px minmax(0,1fr)` gap 12 start, pad 14 18 14 10, ink, no underline; `.is-unread` bg `--k-paper`; dot 7×7 mt 7 radius 50% (`.is-unread` → accent); mat 48×48 `--k-mat` bg, 1px rule, flex-center, `--k-mat-ink`, img contain + multiply, svg 26px; text col flex column gap 3 min-width 0; meta flex space-between gap 8 12px ink-3; type 700 accent-ink (`.k-notif__item--gone` → ink-3); title 14px 500 ellipsis nowrap; body 13px/1.5 ink-2; empty line 13px ink-2 pad 14 18; footer flex space-between pad 12 18 13px, links ink-2 no underline; `.k-acct` h40 pad 0 10 0 4 flex center gap 8, 14px ink, transparent, 1px rule-strong, radius 999; `.k-acct__initial` 30×30 round, paper-sunk, 1px brass, flex-center, display font 14px. `@media (max-width: 767px)`: `.k-watch, .k-acct__name { display: none; }`, `.k-notif__panel { position: fixed; left: 8px; right: 8px; top: 64px; width: auto; max-height: calc(100dvh - 72px); overflow-y: auto; }` (max-height: plan-authored guard for short screens).

`util.js` — `ago(iso, now = new Date())` per A6 (Taipei date via the existing `YMD` formatter; `HH:MM` via a `zh-TW`/`en-GB` 24 h formatter in Asia/Taipei).

`notify.js` (new) — `export function initNotify()`:
- Elements from the header; `TYPE = {drop: "降價", new: "新上架", gone: "已找到新主人"}`; `let events = []`.
- `document` click on `[data-signin]` → `preventDefault`; `api("/api/member")` → `events = m.events; signIn(m); show(); bell.focus()`; catch → `console.warn` (A12).
- On init: `if (member()) { show(); load(); }` where `load()` refetches events and paints (A7).
- `show()`: guest block `hidden`, member block shown, initial/name from `member()`.
- `paint()` (also `onChange(paint)`): unread = events not `isRead`; badge text/hidden; bell aria-label (S2); `[data-watch-count]` → `追蹤清單（${trackedCount()}）`; list = items (A5 copy via `valuation()` from `card.js`, `ago(ev.at)`, `esc` on every string) or the A12 empty line. Item markup: `<li><a class="k-notif__item[ is-unread][ k-notif__item--gone]" href="/listing/{encodeURIComponent(id)}" data-ev="{eventID}"><span class="k-notif__dot">{unread ? <span class="k-sr">未讀</span> : ""}</span><span class="k-notif__mat" aria-hidden="true">{img | PENDING_ICON}</span><span class="k-notif__text"><span class="k-notif__meta"><span class="k-notif__type">{TYPE}</span><span>{ago}</span></span><span class="k-notif__title">{title}</span><span class="k-notif__body">{body}</span></span></a></li>`.
- `setOpen(open, {focus})`: panel `hidden`, bell `aria-expanded`; open → h2 focus; close with `focus` → bell. Bell click toggles; `keydown` Escape while open → close + focus bell; `pointerdown` outside `.k-notif` → close; panel click on `a[data-ev]` → `markRead([ev])` + close (the router's document listener still navigates); 全部標為已讀 → `markRead(events)`; `popstate` → close.
- 我的帳號 button and `#` links: no handlers (inert, A1 / Out of Scope).

`app.js` — `import { initNotify } from "./notify.js";` and call `initNotify()` next to `initAdvisor(...)`.

**Rationale:** Static markup + small module mirrors the advisor's pattern (dialog markup in `index.html`, behaviour in its module); copy goes through `valuation()` so a notification and the badge on the linked page cannot word the same figure differently.

**Verification:** Step 4 matrix N1–N12.

**If this fails:** revert the five files; Step 2's store still serves guest tracking.

---

### Step 4: End-to-end verification [PATCHED] → 2610131027-kura-member-notifications-patch.md

**Objective:** S1–S9 evidence. **Confidence:** Medium (hotlinked photos/fonts need network). **Depends on:** Steps 1–3.
**Verify-Projex:** Encouraged

**Procedure:**
1. `go vet ./...` + `go test ./...`; hygiene grep `grep -rnE "[0-9]{10}-[a-z0-9-]+-(plan|sprint|def|nav|audit|patch)" internal cmd` → none; `git diff 780303d -- internal/radar/radar.go` changes no `0.05`, `anomalyPct`, `dealScore`, `confidenceWeight` value (S7 / F1–F6 untouched).
2. Background `go run ./cmd/kura serve -db kura.db -addr 127.0.0.1:8090` (worktree root, no crawl flags, never `POST /api/crawl`). Empty DB: `go run ./cmd/kura serve -db <scratchpad>/empty.db -addr 127.0.0.1:8092`.
3. Matrix (browser pane; fresh tab or `sessionStorage.clear()` before N1):

| # | Case | Action | Check |
|---|---|---|---|
| N1 | sign-in | 1280 `/` → click 登入 | no reload (`window.__probe` survives); member header (bell, heart, pill 林 林小姐 ⌄); focus on bell; badge `2`; bell aria-label `通知，2 則未讀` |
| N2 | hearts after sign-in | N1 → Home 價格最美 shelf / Browse `deals=1&brand=LOUIS%20VUITTON` | `brandoff:2101220216016` card heart pressed (A2 listing); others as before |
| N3 | dropdown desktop | click bell | canvas `Notifications` 1280 (popover under bell, 380 wide); 2 items, newest first, both unread (dot + paper bg); types 新上架; titles `你追蹤的型號 …`; bodies `新的一件 {grade}，NT$…，低於行情 N%`; footer 追蹤清單（4） + 通知設定; focus on h2 |
| N4 | no contradiction | click each item | lands on `/listing/{id}` in-app; page price = body NT$; page badge `低於行情 N%` = body N; 上架 date consistent with time label; 售價變化 shows no change; status active; item now read, badge `1`, panel closed |
| N5 | model consistency | open `/model/LOUIS%20VUITTON/M51130` and `/model/CHANEL/A01112` | 追蹤這個型號 shows 追蹤中; the notified listing is among 在售 |
| N6 | read all | open bell → 全部標為已讀 | dots gone, bg transparent, badge hidden, aria-label `通知` |
| N7 | close paths | open → Escape; open → click page body; open → Back | closes each time; Escape returns focus to bell; `aria-expanded` false |
| N8 | one store | toggle a card heart, detail 追蹤這件, model 追蹤這個型號, advisor recommendation 追蹤 (demo mode fine) | every copy of the key repaints; 追蹤清單（N） follows (open bell to read N); feed unchanged (A8) |
| N9 | persistence | reload as member; open new tab `/` | reload: still member, read state kept, hearts kept; new tab: guest header (登入 / 加入會員) |
| N10 | guest merge | new tab: heart a card as guest → 登入 | that heart stays pressed; 追蹤清單（5） |
| N11 | layouts/themes | N3 at 1280 + 375 × `?theme=light|dark` (+ once colorScheme emulation) | 375 = canvas `NotificationsMobile` (fixed panel 8 px inset, top 64; no heart icon; pill without name); header single row; `scrollWidth <= innerWidth`; dark tokens throughout (badge ring = paper) |
| N12 | keyboard | 1280: Tab to 登入 → Enter → Enter on bell → Tab through readall, items, footer links | visible ring on each; Enter on item navigates |
| N13 | empty DB | :8092 `/` → 登入 → bell | member header, no badge, 「目前沒有通知。」, 追蹤清單（0） |
| N14 | inert controls | click 加入會員, 我的帳號, heart icon, 追蹤清單（N）, 通知設定 | no navigation, view not re-rendered (nav.js `#` handling), 0 console errors |

4. Per view: `read_console_messages` onlyErrors empty; page text has no canvas sample values (`NT$27,854`, `NT$104,955`, `A33814`).
5. Stop servers; delete `<scratchpad>/empty.db*`; `git -C <work-root> status` clean except committed work.

---

## Verification Plan

### Automated Checks
- [x] `go vet ./...` clean
- [x] `go test ./...` green incl. `TestIsDeal`, `TestGone`, `TestMemberAPI`, extended `TestStaticRoutes`

### Manual Verification
- [x] Step 4 matrix N1–N14
- [x] Home / Browse / detail / model / advisor otherwise unchanged (guest view identical before 登入)
- [x] Source hygiene grep → none

### Acceptance Criteria Validation

| Criterion | How to Verify | Expected Result |
|---|---|---|
| S1 | N1, N9, N14 | in-app sign-in; session-scoped; 加入會員 inert |
| S2 | N1, N6, N11 | canvas member header; live aria-label + badge |
| S3 | N3, N7, N11 vs canvas artboards | dropdown markup, positions, close paths |
| S4 | N3, N4, N6, N9 | 2 unread at sign-in; item/all read; persists |
| S5 | N2, N5, N8, N10 | one store; repaint everywhere; merge |
| S6 | `TestMemberAPI`; N4, N5; step 4.4 | figures = listing API; only observed events |
| S7 | Step 1 tests; `curl /api/member` | shape; predicate unchanged; `Gone` order |
| S8 | N11, N12; console | both themes/widths, one-row header, rings, 0 errors |
| S9 | vet / test / grep | green / none |

**Data-class trace (member):** persona name/initial ← canvas (mocked) | watchlist keys ← A2 pick over real listings/models (mocked: who tracks what) | events ← radar observations on those keys (real: drop history, gone status, first_seen) | every figure in an event ← radar `Item` = `/api/listings/{id}` (real / radar valuation, one computation) | read state ← session (mocked) | 追蹤清單（N） ← store size (session). Real-derived: none shown.

---

## Rollback Plan

1. `git -C <work-root> log --oneline 780303d..HEAD` → this plan's commits.
2. `git -C <work-root> revert <sha>…` newest first (sprint branch only): header back to guest-only, tracking back to the in-memory Set, `/api/member` gone.

---

## Notes

### Split Decision
No split — single scope (one `.projex` folder; `internal/radar` + `internal/web` one repo scope, as iterations 1–5), within size budget (~400 lines, 4 steps; heuristic needs > 5 steps and > 500 lines). Fits one strategy run.

### Risks
- **Guest flash on reload as member:** the module runs after first parse, so the guest block can show for a frame before `notify.js` swaps it. Accepted for a demo; an inline head check of `sessionStorage` (like the `?theme=` script) would remove it.
- **Model key format lives in two places** (`model.js` `trackKey`, `member.go` `modelKey`) — Go tests cannot see the JS; N5 checks they agree. A mismatch would show 追蹤這個型號 unpressed for a tracked model.
- **Watchlist drifts with data:** picks are recomputed per request; a crawl between sign-in and reload can change which models qualify — the merged store keeps the sign-in keys, the feed shows events for the current pick. No crawl runs in the mock; hosted realtime (Phase 6) would pin the persona pick.
- **Persona brands fixed in code** (`memberBrands`) — if HERMES never qualifies (today: all low confidence) the watchlist leans LV/CHANEL; acceptable, logged with A2.
- Hotlinked photos/fonts unreachable → same exposure as other views.

### Open Questions
None blocking execution. For the orchestrator to log (sprint nav / def), not decided here:
- A1 entry (登入 → 林小姐 directly), A2 persona watchlist composition + brands, A7 two-unread rule, A9 session scope (no sign-out) → provisional against def L3 personas / sign-in wording.
- A3 type set = canvas three, each only when observed (no 低於行情 type; feed shows only 新上架 on today's DB) and A8 static feed → def L3 notification types / static vs reactive (still open; canvas evidence applied). Def Canvas deviation "示意通知 → replaced" now concrete: replaced by observed events on a mocked watchlist.
- A4 新上架 = newest *seen* on day-1 data → def L4 day-1 appearance.
- A5 copy for non-below states (no 低於行情 clause), A6 time-label rule, A11 photo thumb, A12 empty-feed line 「目前沒有通知。」, A14 dark × notifications (token-driven) → def § Canvas gaps / canvas pass 2 copy review.
