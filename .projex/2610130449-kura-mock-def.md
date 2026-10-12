# Definition: Kura Mock

> **Created:** 2026-10-13 | **Last Revised:** 2026-10-13
> **Author:** Claude Opus 5.5 (definer)
> **Parent:** User
> **Scope:** The Kura mock as a deliverable — what it is, for whom, what it shows. Goal source of truth: Claude Design canvas "Kura 蔵 redesign — first pass" (https://claude.ai/artifact/6iAvGmwsajiW4X9XCD6THw), **pinned at version 1791839497-46ec** (re-read 2026-10-13; def first written against 1791835276-abf4). A newer canvas version binds only through a def revision. Canvas vs def split: Boundaries § Canvas precedence.
> **Status:** In Progress
> **Related Projex:** 2610130420-kura-roadmap-nav.md (how we get there) | 2610130513-kura-def-nav-vs-design-canvas-redteam.md (canvas-drift findings this def was revised against)
> **Mode:** Mixed — Discovery (canvas + existing `kura/` code are readable evidence) | Origination (scope decisions are the user's)

---

## Identity

*Layer L1 — every statement user-established 2026-10-13 unless marked.*

Kura mock: an **interactive mock** of Kura (蔵), a product that tracks pre-owned luxury handbags listed by Japanese resale providers and helps Taiwanese buyers judge price and condition (price radar + Claude-powered bag advisor). It is a **hosted demo**, delivered as a **link sent with the user's job application** (104 job 925vi); the reviewer opens it to evaluate the user's work. Its provider data is **real and realtime** — from BRAND OFF and more providers, **crawled or not**, kept current while hosted. Everything user-side is **demo-grade** ("just a mock"), entered through **pre-made demo personas**. Its intended shape is the Claude Design canvas above (**canvas = latest goal**; operationalized as the pinned version + precedence rule — revise 2026-10-13). Primary takeaway for the reviewer: **product thinking + design**.

**Realtime (user):** new listings + price changes picked up **hourly**; **full refresh daily** (catches sold items). Cadence always inside each provider's terms.

*L1 agreed by user 2026-10-13.*

---

## Boundaries

*Layer L2 — resolved 2026-10-13 except items under Open Questions L2. Revised 2026-10-13 against canvas 1791839497-46ec.*

**Canvas precedence:**
- **Canvas governs:** look, IA (screens, header, footer links), copy, designed component states (`trends: sample|hidden`, `data: full|empty`, card variants, advisor `claude|demo`).
- **Def governs:** data semantics (real | real-derived | mocked), scope (providers, areas), invariants.
- Conflict → logged under **Canvas deviations** below; never settled silently during build.
- Canvas is a self-declared "first pass" and single-provider (every sample `shop: 'BRAND OFF'`). Surfaces the def requires but the canvas lacks → **Canvas gaps** below → designed (canvas pass 2, nav Phase 3) before they are built, not improvised.

**Is:**
- Hosted, reviewer-facing web app at a link in the application
- Multi-provider: **≥3 providers** (BRAND OFF + ≥2 more), any legitimate channel. "Legitimate" covers **display** too — attribution, image use, comparative labelling (價格最美 ranking, 異常低價) within each provider's terms — not only data acquisition.
- **Link-out only** for commerce: every listing links to its provider's own page; buying, payment, shipping happen at the provider
- Canvas areas (user, all in scope): **Home + Browse** (大家都在找 / 最新上架 / 價格最美, filters, empty state) | **Detail + Model pages** (鑑定卡, valuation prose, gallery, price distribution, price history) | **Advisor** (floating / centered / full-page panel, helper progress lines) | **Watchlist + notifications** (track listing / model, notification dropdown, personas)
- Footer pages linked on the canvas (`project/Main.dc.html` footer: 關於 Kura / 行情怎麼算 / 常見問題 / 聯絡我們 / 隱私權政策 / 服務條款) are canvas IA → reachable. **行情怎麼算** (also linked from detail valuation) hosts methodology + real-vs-mocked disclosure.
- Three kinds of content:
  - **Real** — provider listings, prices, grades, photos, price history; refreshed per Realtime cadence
  - **Real-derived** — computed from accrued crawl history, needs no users: median days-to-sell, sell-through, 已售出 count, 30-day trend, same-model price distribution. Radar already computes sell-through + median days-to-sell (`internal/radar/radar.go` `SellThrough`, `MedianDaysToSell`); advisor tool returns them (`internal/advisor/tools.go`). Until history suffices → the canvas's own not-yet states (`trends: hidden`, 資料不足 per HANDOFF §4.C), never sample values. Canvas intent: samples are "for layout only; real values need ~2 weeks of history" (`project/Main.dc.html` model-page sample block).
  - **Mocked** — needs real users or a user-data service: demand (大家都在找 shelf, 「這週 N 人在找」), notification feed, persona watchlists. User: "Mock. We don't have the service yet."
  - ~~Mocked: demand, trends, days-to-sell, notification feed, persona watchlists~~ — superseded 2026-10-13: trends + days-to-sell need history, not users → Real-derived.
  - **One source per stat:** any number shown in both UI and advisor output comes from the same computation (no mocked tile beside a real advisor answer).
  - Mocked surfaces are **unlabelled** in the UI (user) — no per-surface 示意數據 mark; real-vs-mocked explained on 行情怎麼算. Risk on record: a probing reviewer can read an unlabelled mocked number as fabrication (red team F19) → Open Questions L2.
  - Exception: advisor demo mode is labelled (canvas 示範模式 badge).

**Is not:**
- LINE/EDM marketing automation
- Real accounts / production auth (pre-made personas only)
- A real user-data service: no real demand tracking, no real notification delivery
- A shop: Kura never sells, holds carts, takes payment or ships

**Canvas deviations** (def overrides canvas; canvas 1791839497-46ec):
- Providers: canvas designs 1 shop; def requires ≥3
- Data: canvas sample values (大家都在找 shelf, model stats, distribution, 示意通知) → replaced per content kinds above; real-derived stats real or in the hidden / 資料不足 state
- Grade filter: canvas 最低成色 (任何成色 / 未使用以上 / 輕微使用感以上 / 明顯使用感以上) works across shops only via Kura's internal rank → L4 invariant

**Canvas gaps** (def/nav need these; absent on canvas 1791839497-46ec):
- Multi-provider: card/detail provenance (dealer + channel); same-model distribution with one grade ladder per shop (canvas axis = one shop's ladder); cross-shop comp labelling; CTA + footnote copy per channel (canvas hard-codes 「到 BRAND OFF 商品頁」); dealer-via-marketplace attribution; same item on several channels; condition state for a grade with no published definition
- Member layer: watchlist page (`追蹤清單（4）` → `#`), account / persona menu, sign-in-as-persona (guest header shows 登入 / 加入會員)
- Advisor: error state, spend-cap-reached state (HANDOFF §4.D asked for error + streaming)
- Pages: 行情怎麼算 content; 資料管線 pipeline page (HANDOFF §4.E)
- HANDOFF §4.A not delivered: header status line (FX + date + advisor mode); side column 近期降價
- Card 「尚無行情估價」 state: `project/Card.dc.html` computes `isNone` but renders nothing (Tokens sheet designs the badge)
- Theme × layout combos with no artboard: model, browse, advisor in dark; detail mobile in light → token-driven, checked in a11y pass

---

## Open Questions

### L2 — Boundaries (remaining)
- [x] **Extra pages** — resolved 2026-10-13 by canvas precedence: canvas footer pages in scope; 行情怎麼算 hosts real-vs-mocked disclosure. Still open: 資料管線 pipeline page (HANDOFF §4.E, not on canvas) | reviewer intro → design/plan time (user: no preference)
- [ ] Which concrete providers fill the ≥3 — decided by roadmap Phase 2 vetting, not here
- [ ] **Unlabelled mocks — confirm with user** (red team F19): keep fully unlabelled, or add a discreet marker (hover/footnote) on the remaining mocked numbers? Mocked set is now only user-dependent surfaces.
- [ ] **Confirm Real-derived reclassification with user** — days-to-sell + trends moved from Mocked (user-agreed list) to Real-derived on canvas intent + code evidence; user's criterion ("needs a running service with real users") unchanged.

### L3 — Structure (not entered; canvas evidence to confirm)
- [ ] Personas: canvas shows guest + one member 「林小姐」 with watchlist (追蹤清單（4）) and 2 unread notifications. How many personas, what each demonstrates? Persona entry point undesigned (Canvas gaps).
- [ ] Canvas header for guest shows 登入 / 加入會員 — with personas, does sign-up exist or only "sign in as persona" (e.g. 以示範身分登入)?
- [ ] Tracking granularity on canvas: 追蹤這件 (listing) and 追蹤這個型號 (model). Both?
- [ ] Notification types on canvas: 降價 | 新上架 (tracked model) | 已找到新主人 (tracked item sold). Complete set? Model track copy 「有新上架或低於行情時通知你」 implies a 低於行情 type too. Mocked feed static per persona, or does tracking an item add a synthetic entry?
- [ ] Advisor on canvas: modes `claude` | `demo`; "helpers" progress (行情 / 成色 / 風格) — real tool steps or scripted? Real Claude calls in the mock? Constraint: no helper presents mocked data as grounds — 風格 line reads 「對照「大家都在找」」 (mocked demand) → re-ground on real-derived signals (sell-through / days-to-sell) or drop the line. Hosted mode policy: spend cap → demo fallback; demo-badge copy is developer-facing (「未設定 API 金鑰，顧問以範例資料回答」) → reviewer-facing wording.
- [ ] Detail page gallery: 5 photos (正面 背面 側面 底部 內裡) — required from every provider? Fallback where provider terms forbid display: single photo | 店家照片準備中 placeholder, as an accepted per-provider deviation.
- [ ] Home sections 大家都在找 / 最新上架 / 價格最美 — definitions of each ranking? (大家都在找 = mocked demand)
- [x] Theme (light/dark) + layout (desktop/mobile) both required — resolved 2026-10-13: yes. HANDOFF §2 ("Dark mode … is required"), §3.5 (desktop ~1280 + mobile 375), §4 ("desktop and mobile artboards, each in light and dark"). Uncovered combos → Canvas gaps.

### L4 — Detail (not entered)
- [ ] Invariants carried from current Kura (per-provider grades shown verbatim; same-grade = same provider; honest valuation states; attribution + link-out) — confirm they bind the mock
- [ ] **Grade ordering across schemes** — established (user, prior session; code `internal/listing/listing.go` `Rank`): Kura's internal N–C rank is a bridge only — min-condition filter + cross-grade estimates (跨等級換算); never shown as a grade label; every listing still shows its own shop's grade. Open: disclose the mapping on 行情怎麼算 as Kura's interpretation? | grades with no defensible mapping (賣家自評 free text, parsed grade with no published definition) — excluded from the filter?
- [ ] Behavior when a provider listing disappears after the mock is sent — sold detection is daily only, so a CTA can land on a sold/404 page for up to ~24 h. Options: last-seen line near CTA (「上次確認 N 小時前」) | tighter check for listings on home
- [ ] "Provider" = dealer or channel? For marketplace-API listings (Rakuten/Yahoo) define the attribution line (dealer via marketplace, required credits) + same item on several channels (canonical listing + "also on")
- [ ] Freshness visible? Def L1 makes realtime defining, canvas shows no data age; FX is a static footer date. HANDOFF §4.A asked for a header status line (FX + date + mode)
- [ ] Day-1 appearance: what real-derived surfaces show before ~2 weeks of history (hidden / 資料不足), per provider; per-model stats mixing providers onboarded on different dates (canvas caption 「自 10/13 起記錄」)

---

## Revision Log

| Date | Summary |
|------|---------|
| 2026-10-13 | Initial definition: L1 Identity from user statements; L1 confirmations + L2–L4 open, seeded from canvas evidence. |
| 2026-10-13 | L1 round: hosted demo w/ realtime provider data; hosted link delivery; takeaway = product thinking + design. Identity rewritten; freshness question moved to L1 ("realtime" staleness). |
| 2026-10-13 | L1 agreed. Realtime = hourly new/price changes + daily full refresh. L2 partial: all 4 canvas areas in scope; real vs mocked content split (service-dependent surfaces mocked). Boundaries section added. Status Draft → In Progress. |
| 2026-10-13 | L2 round: mocked surfaces unlabelled (explained elsewhere); link-out only, Kura never sells; ≥3 providers; extra pages = no preference → open. L2 resolved pending restatement agreement. |
| 2026-10-13 | **Revise (canvas drift):** canvas pinned at 1791839497-46ec (was 1791835276-abf4) + precedence rule (canvas = look/IA/states; def = data semantics/scope); content split into Real / Real-derived / Mocked — days-to-sell + trends no longer mocked, shown real or in canvas hidden/資料不足 state; one source per stat (UI = advisor); "legitimate" covers display terms; Canvas deviations + Canvas gaps lists added; extra pages resolved by canvas footer (行情怎麼算 hosts disclosure); theme × layout resolved (both required, HANDOFF §2/§3.5/§4); L3/L4 gained notification 低於行情, 風格 helper grounding, advisor cap/badge policy, gallery fallback, grade-ordering (internal rank, established), stale link-outs, provider = dealer vs channel, freshness, day-1 appearance; two user confirmations queued (L2). — trigger: 2610130513-kura-def-nav-vs-design-canvas-redteam.md F1–F22, re-verified against canvas 1791839497-46ec (`project/Main.dc.html`, `project/Card.dc.html`, `project/canvas.json`), `internal/radar/radar.go`, `internal/advisor/tools.go`, `docs/design-handoff/HANDOFF.md`. |
