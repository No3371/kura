# Definition: Kura Mock

> **Created:** 2026-10-13 | **Last Revised:** 2026-10-13
> **Author:** Claude Opus 5.5 (definer)
> **Parent:** User
> **Scope:** The Kura mock as a deliverable — what it is, for whom, what it shows. Goal source of truth: Claude Design canvas "Kura 蔵 redesign — first pass" (https://claude.ai/artifact/6iAvGmwsajiW4X9XCD6THw), version 1791835276-abf4.
> **Status:** In Progress
> **Related Projex:** 2610130420-kura-roadmap-nav.md (how we get there)
> **Mode:** Mixed — Discovery (canvas + existing `kura/` code are readable evidence) | Origination (scope decisions are the user's)

---

## Identity

*Layer L1 — every statement user-established 2026-10-13 unless marked.*

Kura mock: an **interactive mock** of Kura (蔵), a product that tracks pre-owned luxury handbags listed by Japanese resale providers and helps Taiwanese buyers judge price and condition (price radar + Claude-powered bag advisor). It is a **hosted demo**, delivered as a **link sent with the user's job application** (104 job 925vi); the reviewer opens it to evaluate the user's work. Its provider data is **real and realtime** — from BRAND OFF and more providers, **crawled or not**, kept current while hosted. Everything user-side is **demo-grade** ("just a mock"), entered through **pre-made demo personas**. Its intended shape is the Claude Design canvas above (**canvas = latest goal**). Primary takeaway for the reviewer: **product thinking + design**.

**Realtime (user):** new listings + price changes picked up **hourly**; **full refresh daily** (catches sold items). Cadence always inside each provider's terms.

*L1 agreed by user 2026-10-13.*

---

## Boundaries

*Layer L2 — resolved 2026-10-13 except extra pages (user: no preference → open).*

**Is:**
- Hosted, reviewer-facing web app at a link in the application
- Multi-provider: **≥3 providers** (BRAND OFF + ≥2 more), any legitimate channel
- **Link-out only** for commerce: every listing links to its provider's own page; buying, payment, shipping happen at the provider
- Canvas areas (user, all in scope): **Home + Browse** (大家都在找 / 最新上架 / 價格最美, filters, empty state) | **Detail + Model pages** (鑑定卡, valuation prose, gallery, price distribution, price history) | **Advisor** (floating / centered / full-page panel, helper progress lines) | **Watchlist + notifications** (track listing / model, notification dropdown, personas)
- Two kinds of content:
  - **Real** — provider listings, prices, grades, price history; refreshed per Realtime cadence
  - **Mocked** — anything that needs a running service with real users: demand (「這週 N 人在找」), trends, days-to-sell, notification feed, persona watchlists. User: "Mock. We don't have the service yet."
  - Mocked surfaces are **unlabelled** in the UI (no per-surface 示意數據 mark); real-vs-mocked explained elsewhere — where: open (see Extra pages)

**Is not:**
- LINE/EDM marketing automation
- Real accounts / production auth (pre-made personas only)
- A real user-data service: no real demand tracking, no real notification delivery
- A shop: Kura never sells, holds carts, takes payment or ships

---

## Open Questions

### L2 — Boundaries (remaining)
- [ ] **Extra pages** — user: no preference (2026-10-13); decide at design/plan time. Candidates: 資料管線 pipeline page (not on canvas) | About/FAQ | legal stubs (隱私權政策, 服務條款) | reviewer intro. Constraint from L2: real-vs-mocked must be explained somewhere since surfaces are unlabelled
- [ ] Which concrete providers fill the ≥3 — decided by roadmap Phase 2 vetting, not here

### L3 — Structure (not entered; canvas evidence to confirm)
- [ ] Personas: canvas shows guest + one member 「林小姐」 with watchlist (追蹤清單（4）) and 2 unread notifications. How many personas, what each demonstrates?
- [ ] Canvas header for guest shows 登入 / 加入會員 — with personas, does sign-up exist or only "sign in as persona"?
- [ ] Tracking granularity on canvas: 追蹤這件 (listing) and 追蹤這個型號 (model). Both?
- [ ] Notification types on canvas: 降價 | 新上架 (tracked model) | 已找到新主人 (tracked item sold). Complete set?
- [ ] Advisor on canvas: modes `claude` | `demo`; "helpers" progress (行情 / 成色 / 風格) — real tool steps or scripted? Real Claude calls in the mock?
- [ ] Detail page gallery: 5 photos (正面 背面 側面 底部 內裡) — required from every provider?
- [ ] Home sections 大家都在找 / 最新上架 / 價格最美 — definitions of each ranking?
- [ ] Theme (light/dark) + layout (desktop/mobile) both required in the delivered mock?

### L4 — Detail (not entered)
- [ ] Invariants carried from current Kura (per-provider grades shown verbatim; same-grade = same provider; honest valuation states; attribution + link-out) — confirm they bind the mock
- [ ] Behavior when a provider listing disappears after the mock is sent

---

## Revision Log

| Date | Summary |
|------|---------|
| 2026-10-13 | Initial definition: L1 Identity from user statements; L1 confirmations + L2–L4 open, seeded from canvas evidence. |
| 2026-10-13 | L1 round: hosted demo w/ realtime provider data; hosted link delivery; takeaway = product thinking + design. Identity rewritten; freshness question moved to L1 ("realtime" staleness). |
| 2026-10-13 | L1 agreed. Realtime = hourly new/price changes + daily full refresh. L2 partial: all 4 canvas areas in scope; real vs mocked content split (service-dependent surfaces mocked). Boundaries section added. Status Draft → In Progress. |
| 2026-10-13 | L2 round: mocked surfaces unlabelled (explained elsewhere); link-out only, Kura never sells; ≥3 providers; extra pages = no preference → open. L2 resolved pending restatement agreement. |
