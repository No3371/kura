# Kura Roadmap — interactive mock on real multi-provider data, sent with the application

> **Status:** In Progress
> **Created:** 2026-10-13 | **Last Revised:** 2026-10-13
> **Author:** Claude Opus 5.5 (navigator)
> **Parent:** User
> **Scope:** Whole `kura/` repo — ingestion (all providers), radar, advisor, web UI, delivery of the mock. Excludes LINE/EDM marketing automation (user: out of scope).
> **Parent Navigation:** none
> **Related Projex:** 2610130449-kura-mock-def.md (what the mock IS — goal spec) | 2610130513-kura-def-nav-vs-design-canvas-redteam.md (canvas-drift findings this nav was revised against) | Non-projex inputs: `docs/design-handoff/HANDOFF.md` (design brief) | Claude Design canvas "Kura 蔵 redesign — first pass" (https://claude.ai/artifact/6iAvGmwsajiW4X9XCD6THw) — user: canvas = latest goal; **pinned at version 1791839497-46ec** (same pin as def; newer version binds only via def revision)

---

## Vision

A **hosted interactive mock** of the Kura product, linked from the user's application (104 job 925vi — AI engineer, luxury cross-border e-commerce). User-side features are "just a mock"; provider data is real and realtime. Reviewer takeaway: product thinking + design. Target shape = the Claude Design canvas (look, IA, copy, designed states); data semantics + scope = 2610130449-kura-mock-def.md (def § Canvas precedence; deviations + canvas gaps listed there).

- **Real, realtime data, many providers, crawled or not** — listings from BRAND OFF + additional Japanese resale providers via official APIs, feeds/imports or polite crawling, kept current while hosted; each provider keeps its own condition grading (店家鑑定 vs 賣家自評).
- **Redesigned UI** — canvas look: home (大家都在找 / 最新上架 / 價格最美), browse, detail w/ 鑑定卡, model page, advisor panel; light/dark; desktop/mobile.
- **Member layer, all demo-grade** — pre-made demo personas; 追蹤清單 (watchlist) + notifications (price drop | new listing of tracked model | sold).
- **Market depth** — photo gallery (real); trends, days-to-sell, sell-through, distribution (real-derived from crawl history; canvas hidden/資料不足 state until enough); demand 大家都在找 (mocked — needs users).
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
  - Note (revise 2026-10-13): pass 1 is single-provider and leaves HANDOFF §4 items undelivered → tracked as canvas gaps (def) + Phase 3 canvas pass 2

**Exit Criteria:** ✓ met.

---

### Phase 1: Baseline & Verification — Current

**Goal:** trustworthy starting point for everything after.

**Milestones:**
- [ ] Baseline commit of current tree (user-requested)
- [ ] Real Claude advisor run end-to-end w/ API key; fix whatever the live API surfaces
- [ ] Second+ crawls → confirm price-drop + sold detection on real data
- [ ] Incremental refresh mode: hourly "first pages only" sweep that never marks items sold (current crawler only has full sweeps); record last-seen per listing (feeds def L4 stale-link-out decision)

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
- [ ] Channel/provider vetting rubric + shortlist — per provider: available channels, terms (display/attribution/affiliate rules, caching limits, image use, comparative labelling such as 價格最美 ranking + 異常低價 — def: "legitimate" covers display), grading scheme, bag volume, freshness. Leads: KOMEHYO / BRAND OFF / other dealers via Rakuten + Yahoo APIs; ALLU (robots allows market pages); RAGTAG (robots allows; fashion-heavy). Probe-blocked: 2nd STREET, Daikokuya.
- [ ] Generalize ingestion seam: `crawl.Source` → channel-agnostic provider interface (sweep = API pagination | file import | crawl); provenance per listing (provider, channel, fetched-at)
- [ ] Marketplace-API adapter (first: Rakuten or Yahoo) + per-shop `listing.Scheme` + free-text grade parser w/ fixtures
- [ ] File-import adapter (CSV/JSON) for non-crawled feeds
- [ ] Cross-source identity: same model across providers (model-code normalization) + **same item listed on multiple channels** (dealer's own site + marketplace) → dedupe
- [ ] Learned rank factors from data (B/A measured 0.75 vs hard-coded 0.85) — needed once grades come from several schemes; operates on Kura's internal rank only, never shown as a grade label (def L4)

**Exit Criteria:** ≥3 providers on schedule via ≥2 channel types; every listing shows provider + channel + its own grading; cross-provider comps labelled; no duplicate items across channels. UI for provenance / comp labels / dedupe follows canvas pass 2 (Phase 3), not ad-hoc screens.

---

### Phase 3: Redesign Implementation — Next (parallel w/ Phase 2)

**Goal:** ship the Claude Design look (canvas 1791839497-46ec) as the real UI; close the canvas gaps by design, not improvisation.

**Milestones:**
- [ ] **Canvas pass 2** (design work on the canvas, before the IA milestone builds the affected surfaces) — covers def § Canvas gaps: multi-provider (card/detail provenance, per-shop distribution ladders, cross-shop comp labels, CTA/attribution per channel incl. dealer-via-marketplace, dedupe "also on", grade-without-definition condition state) | member layer (watchlist page, account/persona menu, sign-in-as-persona replacing 加入會員) | advisor error + cap-reached states, reviewer-facing demo badge | 行情怎麼算 page | HANDOFF §4 leftovers (header status line, 近期降價, pipeline page if kept). Single-provider surfaces (tokens, home, browse, advisor panel) need not wait.
- [ ] Tokens + type + base components (`--k-*`, fonts, mat frame, valuation badges, condition tag both variants; card 「尚無行情估價」 state — canvas computes `isNone` but never renders it)
- [ ] New IA: home sections, Browse (價格最美), detail page w/ 鑑定卡 + prose valuation, model page; canvas footer pages reachable (關於 Kura / 行情怎麼算 / 常見問題 / 聯絡我們 / 隱私權政策 / 服務條款)
- [ ] Advisor as floating panel ↔ full page; step lines ("helpers") mapped to real tool calls; 風格 helper grounded on real-derived signals (sell-through / days-to-sell) or dropped — never on mocked 大家都在找 demand (def L3)
- [ ] Light/dark, mobile, a11y pass — both required (def L3, resolved); combos with no artboard (model/browse/advisor dark, detail mobile light) token-driven, checked here; HANDOFF §3.7 bar (AA contrast, visible focus, ≥24 px targets, chart table view)
- [ ] 行情怎麼算 page: methodology + real / real-derived / mocked disclosure (def: mocked surfaces unlabelled in UI; disclosure lives here) + grade-ordering note per def L4
- ~~Real-vs-mocked explanation surface (page choice open — pipeline | about | reviewer intro)~~ — superseded 2026-10-13: page = 行情怎麼算 (canvas footer + detail link); pipeline / reviewer intro remain open in def L2

**Exit Criteria:** every canvas screen reachable in the running app; every def § Canvas gaps item designed in pass 2 and built, or explicitly deferred in the def; real / real-derived / mocked exactly as the def classes them.

---

### Phase 4: Member Layer (demo-grade) — Future

**Goal:** watchlist + notifications usable through pre-made demo personas.

**Milestones:**
- [ ] Pre-made demo personas (no real auth; persona switch / sign-in as persona) — UI from canvas pass 2 (not on canvas 1791839497-46ec)
- [ ] 追蹤清單: track listing | track model (interactive, persona-scoped); watchlist page from canvas pass 2
- [ ] Notification feed: **mocked** per persona (降價 | 新上架 | 已找到新主人; 低於行情 pending def L3 — canvas model-track copy promises it) — def: "we don't have the service yet"; static vs reacts-to-tracking pending def L3

**Exit Criteria:** reviewer switches persona → sees that persona's watchlist + mocked notifications; tracking toggles work.

---

### Phase 5: Market Depth — Future

**Goal:** canvas market surfaces present: gallery real; history-derived signals real (canvas hidden/資料不足 state until history suffices); only user-dependent signals mocked.

**Milestones:**
- [ ] Photo gallery: all product images per listing (real; detail-page / API image fields; budgeted); per-provider fallback (single photo | 店家照片準備中) where terms forbid display (def L3)
- [ ] Real-derived market signals: median days-to-sell, sell-through, 已售出, 30-day trend, same-model distribution — radar already computes sell-through + median days-to-sell (`internal/radar/radar.go`), trend from stored price history; canvas `trends: hidden` / 資料不足 states until enough history; one source shared by UI + advisor
- [ ] Mocked user-dependent signals: 大家都在找 shelf / 「這週 N 人在找」 demand (def: mocked — needs users; unlabelled in UI pending def L2 confirmation, disclosed on 行情怎麼算)
- ~~Mocked market signals: 大家都在找 demand, 30-day trends, days-to-sell~~ — superseded 2026-10-13: trends + days-to-sell derive from crawl history (def Real-derived)

**Exit Criteria:** every canvas market surface renders; real / real-derived / mocked split matches 2610130449-kura-mock-def.md; no stat shown in both UI and advisor output from different sources.

---

### Phase 6: Hosted Mock — Future (infra can start earlier so data stays realtime)

**Goal:** a hosted link the reviewer opens unaided, showing realtime provider data.

**Milestones:**
- [ ] Host + deploy; scheduled server-side refresh: **hourly** new listings + price changes (first pages / API delta), **daily** full refresh (sold detection) — within each provider's terms
- [ ] Advisor spend bounded for reviewer use (simple cap / fallback to mock mode); cap-reached transition + reviewer-facing demo copy per canvas pass 2
- [ ] Attribution + link-out per listing; image + data display within each provider's terms
- [ ] FX rate source + refresh (canvas footer FX is a static date); data-age indicator if def L4 freshness question resolves yes
- [ ] Application-ready entry point: what the reviewer sees first, how to switch personas
- [ ] Send-readiness gate: minimum crawl-history days per provider before the link is sent, or explicit acceptance of 資料不足 / hidden states on day 1 (def L4 day-1 appearance) — canvas: real values need ~2 weeks

**Exit Criteria:** reviewer with only the link walks every canvas flow; data age stays within the agreed staleness; send-readiness gate met.

---

## Priorities

**Current focus:** Phase 1 — baseline commit + real advisor run + start accruing crawl history (history is the long pole for Phases 4–5).

**Next up:** Phase 2 provider vetting and Phase 3 redesign in parallel — independent code areas (`internal/source/*` vs `internal/web/static`). Canvas pass 2 gates only the multi-provider + member-layer surfaces; single-provider canvas surfaces can be built first.

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
- [ ] Canvas pass 2: who designs it (user in Claude Design | agent-authored artboards for user review) and does it republish the pinned canvas (→ def revision to re-pin)?
- [ ] User confirmations queued in def L2: Real-derived reclassification (days-to-sell, trends) | unlabelled mocks vs discreet marker

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
| 2026-10-13 | **Revise (canvas drift):** canvas pinned at 1791839497-46ec (same as def) + precedence pointer; Phase 3 gains canvas pass 2 milestone (multi-provider, member layer, advisor states, 行情怎麼算, HANDOFF §4 leftovers) and exit now checks def § Canvas gaps; explanation surface → 行情怎麼算 (prior milestone superseded); 風格 helper grounding; theme × layout required; Phase 5 goal reverses the earlier flip — days-to-sell/trends real-derived w/ hidden/資料不足 states, only demand mocked (prior milestone superseded); Phase 1 records last-seen; Phase 2 rubric adds display + comparative-labelling terms; Phase 4 persona/watchlist UI from pass 2, 低於行情 type pending; Phase 6 adds FX source, cap-state copy, send-readiness gate; 2 open questions added. — trigger: 2610130513-kura-def-nav-vs-design-canvas-redteam.md (F1–F22, Must/Should Fix), re-verified against canvas 1791839497-46ec + `internal/radar/radar.go` + HANDOFF §4; mirrors 2610130449-kura-mock-def.md revision. |
