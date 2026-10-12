# Kura Roadmap — interactive mock on real multi-provider data, sent with the application

> **Status:** In Progress
> **Created:** 2026-10-13 | **Last Revised:** 2026-10-13
> **Author:** Claude Opus 5.5 (navigator)
> **Parent:** User
> **Scope:** Whole `kura/` repo — ingestion (all providers), radar, advisor, web UI, delivery of the mock. Excludes LINE/EDM marketing automation (user: out of scope).
> **Parent Navigation:** none
> **Related Projex:** 2610130449-kura-mock-def.md (what the mock IS — goal spec) | Non-projex inputs: `docs/design-handoff/HANDOFF.md` (design brief) | Claude Design canvas "Kura 蔵 redesign — first pass" (https://claude.ai/artifact/6iAvGmwsajiW4X9XCD6THw) — user: canvas = latest goal

---

## Vision

A **hosted interactive mock** of the Kura product, linked from the user's application (104 job 925vi — AI engineer, luxury cross-border e-commerce). User-side features are "just a mock"; provider data is real and realtime. Reviewer takeaway: product thinking + design. Target shape = the Claude Design canvas; precise spec lives in 2610130449-kura-mock-def.md.

- **Real, realtime data, many providers, crawled or not** — listings from BRAND OFF + additional Japanese resale providers via official APIs, feeds/imports or polite crawling, kept current while hosted; each provider keeps its own condition grading (店家鑑定 vs 賣家自評).
- **Redesigned UI** — canvas look: home (大家都在找 / 最新上架 / 價格最美), browse, detail w/ 鑑定卡, model page, advisor panel; light/dark; desktop/mobile.
- **Member layer, all demo-grade** — pre-made demo personas; 追蹤清單 (watchlist) + notifications (price drop | new listing of tracked model | sold).
- **Market depth** — photo gallery, demand signals (大家都在找), trends, days-to-sell.
- **Claude advisor** inside the mock.

---

## Current Position

**As of 2026-10-13:**

Working local app, single provider, old UI. Repo has **zero commits**; no prior projex.

### Recent Progress
- Crawler + polite fetcher (robots.txt, Crawl-delay, honest `KuraBot` UA, backoff) — BRAND OFF adapter, fixture tests — 2026-10-13
- Full crawl: 15,366 bags / 6 brands, all sweeps complete — 2026-10-13
- Radar: same-grade comps, dispersion-based confidence, anomaly flag, drops, sell-through — 2026-10-13
- Advisor: streaming Claude tool loop + mock mode; verified via fake-API tests only — 2026-10-13
- Per-source condition schemes (shop label + verbatim definition + zh text; same-grade = same source) — 2026-10-13
- Design handoff written; Claude Design first pass delivered (19 artboards, CSS tokens `--k-*`) — 2026-10-13

### Active Work
- none (this roadmap is the first projex)

### Known Blockers
- No commits → projex execute cycle (needs committed plans/base branch) unavailable until baseline commit — user must request commit.
- Real advisor never run: no `ANTHROPIC_API_KEY` in dev env.
- KOMEHYO excluded: rejects self-identified bots; same vetting applies to every new provider.

---

## Roadmap

### Phase 0: Foundation — Done

**Goal:** single-provider pipeline, radar, advisor, honest valuation.

**Milestones:**
- [x] BRAND OFF ingestion + store + radar + advisor + legacy UI
  - Ideation: — (pre-projex)
  - Execution: — (uncommitted working tree)
- [x] Per-source condition grading
  - Execution: — (uncommitted)
- [x] Design brief + first design pass
  - Ideation: `docs/design-handoff/HANDOFF.md`, design canvas

**Exit Criteria:** ✓ met.

---

### Phase 1: Baseline & Verification — Current

**Goal:** trustworthy starting point for everything after.

**Milestones:**
- [ ] Baseline commit of current tree (user-requested)
- [ ] Real Claude advisor run end-to-end w/ API key; fix whatever the live API surfaces
- [ ] Second+ crawls → confirm price-drop + sold detection on real data
- [ ] Incremental refresh mode: hourly "first pages only" sweep that never marks items sold (current crawler only has full sweeps)

**Exit Criteria:** committed baseline; advisor answered ≥3 real queries correctly; ≥1 real drop and ≥1 real sold item observed.

---

### Phase 2: More Providers (crawled or not) — Next

**Goal:** real data from ≥2 additional providers, via whichever **legitimate channel** each offers. Crawling is one channel, not the default.

**Channels (preference order):**
1. **Official API** — e.g. Rakuten Ichiba Item Search API, Yahoo! Shopping Item Search API (v3): keyed, terms-governed; many JP resale dealers (incl. KOMEHYO, which blocks crawlers) list stock there. Condition grade lives in free-text item name/caption → per-shop parsing.
2. **Partner / affiliate feed or file import** — dealer-provided CSV/JSON, affiliate datafeeds, one-off snapshots.
3. **Polite crawl** — only where robots.txt allows **and** an honestly identified bot is served (current BRAND OFF path).
- Never: UA spoofing, CAPTCHA/WAF evasion, C2C scraping against ToS (Mercari etc. only via official route).

**Milestones:**
- [ ] Channel/provider vetting rubric + shortlist — per provider: available channels, terms (display/attribution/affiliate rules, caching limits), grading scheme, bag volume, freshness. Leads: KOMEHYO / BRAND OFF / other dealers via Rakuten + Yahoo APIs; ALLU (robots allows market pages); RAGTAG (robots allows; fashion-heavy). Probe-blocked: 2nd STREET, Daikokuya.
- [ ] Generalize ingestion seam: `crawl.Source` → channel-agnostic provider interface (sweep = API pagination | file import | crawl); provenance per listing (provider, channel, fetched-at)
- [ ] Marketplace-API adapter (first: Rakuten or Yahoo) + per-shop `listing.Scheme` + free-text grade parser w/ fixtures
- [ ] File-import adapter (CSV/JSON) for non-crawled feeds
- [ ] Cross-source identity: same model across providers (model-code normalization) + **same item listed on multiple channels** (dealer's own site + marketplace) → dedupe
- [ ] Learned rank factors from data (B/A measured 0.75 vs hard-coded 0.85) — needed once grades come from several schemes

**Exit Criteria:** ≥3 providers on schedule via ≥2 channel types; every listing shows provider + channel + its own grading; cross-provider comps labelled; no duplicate items across channels.

---

### Phase 3: Redesign Implementation — Next (parallel w/ Phase 2)

**Goal:** ship the Claude Design look as the real UI.

**Milestones:**
- [ ] Tokens + type + base components (`--k-*`, fonts, mat frame, valuation badges, condition tag both variants)
- [ ] New IA: home sections, Browse (價格最美), detail page w/ 鑑定卡 + prose valuation, model page
- [ ] Advisor as floating panel ↔ full page; step lines ("helpers") mapped to real tool calls
- [ ] Light/dark, mobile, a11y pass
- [ ] Real-vs-mocked explanation surface (def: mocked surfaces unlabelled; page choice open — pipeline | about | reviewer intro)

**Exit Criteria:** every canvas screen reachable in the running app; real data where def says real, mocked where def says mocked.

---

### Phase 4: Member Layer (demo-grade) — Future

**Goal:** watchlist + notifications usable through pre-made demo personas.

**Milestones:**
- [ ] Pre-made demo personas (no real auth; persona switch / sign-in as persona)
- [ ] 追蹤清單: track listing | track model (interactive, persona-scoped)
- [ ] Notification feed: **mocked** per persona (降價 | 新上架 | 已找到新主人) — def: "we don't have the service yet"

**Exit Criteria:** reviewer switches persona → sees that persona's watchlist + mocked notifications; tracking toggles work.

---

### Phase 5: Market Depth — Future

**Goal:** canvas market surfaces present: gallery real; service-dependent signals mocked.

**Milestones:**
- [ ] Photo gallery: all product images per listing (real; detail-page / API image fields; budgeted)
- [ ] Mocked market signals: 大家都在找 demand, 30-day trends, days-to-sell (def: mocked — no service yet; unlabelled in UI, real-vs-mocked explained elsewhere)

**Exit Criteria:** every canvas market surface renders; real vs mocked split matches 2610130449-kura-mock-def.md.

---

### Phase 6: Hosted Mock — Future (infra can start earlier so data stays realtime)

**Goal:** a hosted link the reviewer opens unaided, showing realtime provider data.

**Milestones:**
- [ ] Host + deploy; scheduled server-side refresh: **hourly** new listings + price changes (first pages / API delta), **daily** full refresh (sold detection) — within each provider's terms
- [ ] Advisor spend bounded for reviewer use (simple cap / fallback to mock mode)
- [ ] Attribution + link-out per listing; image + data display within each provider's terms
- [ ] Application-ready entry point: what the reviewer sees first, how to switch personas

**Exit Criteria:** reviewer with only the link walks every canvas flow; data age stays within the agreed staleness.

---

## Priorities

**Current focus:** Phase 1 — baseline commit + real advisor run + start accruing crawl history (history is the long pole for Phases 4–5).

**Next up:** Phase 2 provider vetting and Phase 3 redesign in parallel — independent code areas (`internal/source/*` vs `internal/web/static`).

**Deferred:** LINE/EDM marketing automation — user decision, out of scope. C2C marketplaces — only if an official data route exists.

---

## Open Questions

- [x] "Interactive mock" boundary → all member features demo-grade (user, 2026-10-13)
- [x] Accounts → pre-made demo personas (user, 2026-10-13)
- [x] Delivery → hosted link w/ realtime provider data; "just a mock" = user-side features only (user, 2026-10-13)
- [x] "Realtime" → hourly new/price changes + daily full refresh (user, 2026-10-13)
- [ ] Image policy where the mock is shown: hotlink provider images vs cached thumbnails — check each provider's terms
- [ ] Marketplace-API terms: attribution/affiliate-link display and caching limits may constrain the UI (e.g. mandatory credit, price freshness) — vet before adapter work
- [ ] API keys needed (Rakuten app ID / Yahoo client ID) — user to register; never committed

---

## Revision Log

| Date | Summary of Changes |
|------|--------------------|
| 2026-10-13 | Initial roadmap. User inputs: goal = interactive mock, hosted live demo; features = watchlist+notifications, trends, accounts, gallery; real data from BRAND OFF + more providers; LINE/EDM out of scope. |
| 2026-10-13 | User: providers "crawled or not" → Phase 2 reframed around channels (official API > feed/import > polite crawl); added ingestion-seam generalization, file import, cross-channel dedupe, API-terms + key questions. |
| 2026-10-13 | User: all demo-grade, pre-made personas, "just a mock", sent with application (not a demo to operate). Vision reframed; Phase 4 → personas; Phase 6 "Hosted Live Demo" → "Mock Delivery" (dropped unattended-uptime, Prometheus); 3 open questions resolved; goal spec delegated to 2610130449-kura-mock-def.md. |
| 2026-10-13 | User correction (def L1 round): hosted demo w/ realtime provider data, hosted link. Phase 6 → "Hosted Mock" w/ scheduled refresh; vision updated. |
| 2026-10-13 | Def L2 round: realtime = hourly + daily full; service-dependent surfaces (demand, trends, days-to-sell, notification feed) mocked. Phase 1 gains incremental-refresh milestone (replaces local continuous crawl); Phase 4 notifications → mocked per persona; Phase 5 goal flipped from "replace 示意數據" to "gallery real, signals mocked". |
| 2026-10-13 | Def L2 round: mocked surfaces unlabelled → Phase 3 adds real-vs-mocked explanation surface, drops 示意數據 labels; link-out only; ≥3 providers (matches Phase 2 exit). |
