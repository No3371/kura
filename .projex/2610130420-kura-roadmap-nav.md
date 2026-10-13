# Kura Roadmap — interactive mock on real multi-provider data, sent with the application

> **Status:** In Progress
> **Created:** 2026-10-13 | **Last Revised:** 2026-10-13
> **Author:** Claude Opus 5.5 (navigator)
> **Parent:** User
> **Scope:** Whole `kura/` repo — ingestion (all providers), radar, advisor, web UI, delivery of the mock. Excludes LINE/EDM marketing automation (user: out of scope).
> **Parent Navigation:** none
> **Related Projex:** 2610130449-kura-mock-def.md (what the mock IS — goal spec) | 2610130513-kura-def-nav-vs-design-canvas-redteam.md (canvas-drift findings this nav was revised against) | 2610130524-kura-interactive-mock-sprint.md (sprint: single-provider canvas build — Phase 3 + demo-grade Phase 4/5 surfaces; goal reached) | Non-projex inputs: `docs/design-handoff/HANDOFF.md` (design brief) | Claude Design canvas "Kura 蔵 redesign — first pass" (https://claude.ai/artifact/6iAvGmwsajiW4X9XCD6THw) — user: canvas = latest goal; **pinned at version 1791839497-46ec** (same pin as def; newer version binds only via def revision)

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

Working local app, single provider (BRAND OFF, one crawl sweep), **canvas UI built** — every canvas-designed flow walkable end-to-end on real data, light/dark × 1280/375, `go test ./...` green (sprint 2610130524-kura-interactive-mock-sprint.md, goal reached after 9 iterations; work on branch `projex/sprint/2610130524-kura-interactive-mock`, merge to `master` at sprint finalize). Legacy UI kept at `/legacy/`. Not hosted; advisor verified in mock mode only.

### Recent Progress
- Sprint 2610130524-kura-interactive-mock-sprint.md — goal reached — 2026-10-13: canvas foundation + Home | detail `/listing/{id}` | model `/model/{brand}/{group_key}` | advisor dock + dialogs over `/api/chat` SSE | Browse `/browse` | provisional member state (登入 → canvas member 林小姐) + mocked 通知 | footer IA + `/methodology` (行情怎麼算) + 5 pending footer pages | closing walk + a11y pass. Real-derived stats gated (資料不足 / `trends: hidden`), one source UI = advisor (`SoldStatsReady()`). Def § Canvas gaps: built or sprint Deferred w/ blocker, 1:1. Provisional choices + canvas pass 2 candidates → def (Open Questions *Sprint provisional* lines; § Canvas pass 2 candidates)
- Crawler + polite fetcher (robots.txt, Crawl-delay, honest `KuraBot` UA, backoff) — BRAND OFF adapter, fixture tests — 2026-10-13
- Full crawl: 15,366 bags / 6 brands, all sweeps complete — 2026-10-13
- Radar: same-grade comps, dispersion-based confidence, anomaly flag, drops, sell-through — 2026-10-13
- Advisor: streaming Claude tool loop + mock mode; verified via fake-API tests only — 2026-10-13
- Per-source condition schemes (shop label + verbatim definition + zh text; same-grade = same source) — 2026-10-13
- Design handoff written; Claude Design first pass delivered (19 artboards, CSS tokens `--k-*`) — 2026-10-13

### Active Work
- ~~Orchestrated chain: canvas red team → def + nav revised → sprint toward the interactive mock~~ — red team + revisions done; sprint 2610130524-kura-interactive-mock-sprint.md goal reached, in finalize (merge to `master`)

### Known Blockers
- ~~No commits → projex execute cycle unavailable until baseline commit~~ — cleared 2026-10-13 (`90ec1a6`).
- Real advisor never run: no `ANTHROPIC_API_KEY` in dev env.
- KOMEHYO excluded: rejects self-identified bots; same vetting applies to every new provider.
- **R1 — 最新上架 order:** a genuinely new listing first seen inside epoch + 24 h sorts last (`Engine.newer`, shop order within the initial sweep). Invisible on today's one-sweep DB; reviewer-visible once crawling resumes → **fix before further crawls / Phase 6 hosting** (per-brand first complete sweep from `crawl_runs`). Source: 2610131037-kura-member-notifications-audit.md R1. R2 (newest order clusters first-crawled brand, GOYARD) — Minor, orchestrator/def call.
- Human answers pending (each shipped as a provisional choice, recorded in def): see Open Questions below.

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
- [x] Baseline commit of current tree (user-requested) — `90ec1a6`
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

### Phase 3: Redesign Implementation — Current (single-provider build done; canvas pass 2 open)

**Goal:** ship the Claude Design look (canvas 1791839497-46ec) as the real UI; close the canvas gaps by design, not improvisation.

**Milestones:**
- [ ] **Canvas pass 2** (design work on the canvas, before the IA milestone builds the affected surfaces) — covers def § Canvas gaps: multi-provider (card/detail provenance, per-shop distribution ladders, cross-shop comp labels, CTA/attribution per channel incl. dealer-via-marketplace, dedupe "also on", grade-without-definition condition state) | member layer (watchlist page, account/persona menu, sign-in-as-persona replacing 加入會員, 通知設定) | advisor error + cap-reached states, reviewer-facing demo badge | 行情怎麼算 page | HANDOFF §4 leftovers (header status line, 近期降價, pipeline page if kept). Single-provider surfaces (tokens, home, browse, advisor panel) need not wait.
  - Not started. Inputs grown by the sprint: def § Canvas pass 2 candidates (states + copy built provisionally) + def § Canvas deviations (model crumb, 看全部 in place, D1 dark mat caption, …) + sprint Deferred rows (2610130524-kura-interactive-mock-sprint.md). Who designs it = Open Question.
- [x] Tokens + type + base components (`--k-*`, fonts, mat frame, valuation badges, condition tag both variants; card 「尚無行情估價」 state — canvas computes `isNone` but never renders it)
  - Execution: sprint iter 1 (2610130535-kura-canvas-foundation-home-plan.md) + iter 8 D1 dark `--k-mat-note` (2610131207-kura-closing-walk-a11y-patch.md). 「尚無行情估價」 wording for no-group / unmappable grade open (def L4)
- [x] New IA: home sections, Browse (價格最美), detail page w/ 鑑定卡 + prose valuation, model page; canvas footer pages reachable (關於 Kura / 行情怎麼算 / 常見問題 / 聯絡我們 / 隱私權政策 / 服務條款)
  - Execution: sprint iters 1 / 2 / 3 / 5 / 7. Home 大家都在找 = canvas `trends: hidden` state (ranking def open); 價格最美 predicate provisional (F1 / F6 open); 5 footer pages = pending view, content deferred (needs owner facts + canvas pass 2)
- [x] Advisor as floating panel ↔ full page; step lines ("helpers") mapped to real tool calls; 風格 helper grounded on real-derived signals (sell-through / days-to-sell) or dropped — never on mocked 大家都在找 demand (def L3)
  - Execution: sprint iter 4 (2610130815-kura-advisor-panel-plan.md) — real tool steps only, 風格 line dropped (provisional vs def L3); error = functional fallback, designed error / cap states deferred
- [x] Light/dark, mobile, a11y pass — both required (def L3, resolved); combos with no artboard (model/browse/advisor dark, detail mobile light) token-driven, checked here; HANDOFF §3.7 bar (AA contrast, visible focus, ≥24 px targets, chart table view)
  - Execution: sprint iter 8 (2610131159-kura-closing-walk-a11y-plan.md → 2610131253-kura-closing-walk-a11y-audit.md): walk matrix every flow × 4 combos; final scan 16 routes × 4 combos 0 overflow / 0 contrast fails / 0 non-excepted small targets / 0 console errors; route-change focus + router scroll restoration. Accepted Low residuals: `#` placeholders jump to top; scroll map in-memory; post-Back Tab goes to top
- [ ] 行情怎麼算 page: methodology + real / real-derived / mocked disclosure (def: mocked surfaces unlabelled in UI; disclosure lives here) + grade-ordering note per def L4
  - Built, provisional — sprint iter 7 (2610131104-kura-methodology-footer-plan.md → 2610131134-kura-methodology-footer-audit.md): methodology + 哪些是示範 disclosure, thresholds from server constants. Remaining: grade-mapping note waits on def L4 (page describes the rank bridge, not the mapping); copy + look → canvas pass 2
- ~~Real-vs-mocked explanation surface (page choice open — pipeline | about | reviewer intro)~~ — superseded 2026-10-13: page = 行情怎麼算 (canvas footer + detail link); pipeline / reviewer intro remain open in def L2

**Exit Criteria:** every canvas screen reachable in the running app; every def § Canvas gaps item designed in pass 2 and built, or explicitly deferred in the def; real / real-derived / mocked exactly as the def classes them.
- Status (sprint finalize): reachability ✓ and data classes ✓ (single provider); gap items built or deferred in sprint Deferred w/ blocker — **not** yet designed in pass 2 → exit not met.

---

### Phase 4: Member Layer (demo-grade) — In progress (canvas-designed part built; persona layer waits on canvas pass 2 + def L3)

**Goal:** watchlist + notifications usable through pre-made demo personas.

**Milestones:**
- [ ] Pre-made demo personas (no real auth; persona switch / sign-in as persona) — UI from canvas pass 2 (not on canvas 1791839497-46ec)
  - Partial — sprint iter 6 (2610131010-kura-member-notifications-plan.md): one persona = canvas member 林小姐 via provisional 登入; no picker / switching / 加入會員 (sprint Deferred: canvas pass 2 + def L3 persona count)
- [ ] 追蹤清單: track listing | track model (interactive, persona-scoped); watchlist page from canvas pass 2
  - Partial — sprint iters 1 / 3 / 6: 追蹤這件 + 追蹤這個型號 toggles work, one session store (`sessionStorage`), mocked persona watchlist (4 entries), header count. Watchlist page deferred (canvas pass 2)
- [x] Notification feed: **mocked** per persona (降價 | 新上架 | 已找到新主人; 低於行情 pending def L3 — canvas model-track copy promises it) — def: "we don't have the service yet"; static vs reacts-to-tracking pending def L3
  - Execution: sprint iter 6 (→ 2610131037-kura-member-notifications-audit.md, 2610131048-kura-newest-order-member-feed-patch.md): `GET /api/member` mocked feed tied to real listings, 通知 popover / mobile sheet, unread badge, 全部標為已讀. Provisional vs def L3: static feed; 低於行情 = body clause, not a type. 通知設定 deferred

**Exit Criteria:** reviewer switches persona → sees that persona's watchlist + mocked notifications; tracking toggles work.
- Status (sprint finalize): tracking toggles ✓, one persona's watchlist count + mocked notifications ✓; persona switching ✗ → exit not met.

---

### Phase 5: Market Depth — In progress (gated surfaces built; real values need crawl history)

**Goal:** canvas market surfaces present: gallery real; history-derived signals real (canvas hidden/資料不足 state until history suffices); only user-dependent signals mocked.

**Milestones:**
- [ ] Photo gallery: all product images per listing (real; detail-page / API image fields; budgeted); per-provider fallback (single photo | 店家照片準備中) where terms forbid display (def L3)
  - Not started — sprint ships one `image_url` per listing + 店家照片準備中 for pending photos (def L3 *Sprint provisional*)
- [ ] Real-derived market signals: median days-to-sell, sell-through, 已售出, 30-day trend, same-model distribution — radar already computes sell-through + median days-to-sell (`internal/radar/radar.go`), trend from stored price history; canvas `trends: hidden` / 資料不足 states until enough history; one source shared by UI + advisor
  - Partial — sprint iters 2–4: same-model distribution (`dist.js`, server TWD) + per-listing 售價變化 live; 已售出 / 平均找到新主人 in 資料不足 behind `radar.MinSoldHistory` 14 d (provisional, def L4), advisor withholds sold stats on the same `SoldStatsReady()` (one source ✓). 30-day trend not built (canvas places it in the hidden 大家都在找 tile). Real values need Phase 1 crawl history
- [ ] Mocked user-dependent signals: 大家都在找 shelf / 「這週 N 人在找」 demand (def: mocked — needs users; unlabelled in UI pending def L2 confirmation, disclosed on 行情怎麼算)
  - Not built — blocked on ranking definition (human; def L3 Home sections) → Home `trends: hidden`, model demand line omitted. 行情怎麼算 already discloses demand as mocked
- ~~Mocked market signals: 大家都在找 demand, 30-day trends, days-to-sell~~ — superseded 2026-10-13: trends + days-to-sell derive from crawl history (def Real-derived)

**Exit Criteria:** every canvas market surface renders; real / real-derived / mocked split matches 2610130449-kura-mock-def.md; no stat shown in both UI and advisor output from different sources.

---

### Phase 6: Hosted Mock — Future (infra can start earlier so data stays realtime)

**Goal:** a hosted link the reviewer opens unaided, showing realtime provider data.

**Milestones:**
- [ ] Host + deploy; scheduled server-side refresh: **hourly** new listings + price changes (first pages / API delta), **daily** full refresh (sold detection) — within each provider's terms
  - Prerequisite: fix R1 (最新上架 sorts a genuinely new listing last within epoch + 24 h) before scheduled crawls run — see Known Blockers
- [ ] Advisor spend bounded for reviewer use (simple cap / fallback to mock mode); cap-reached transition + reviewer-facing demo copy per canvas pass 2
- [ ] Attribution + link-out per listing; image + data display within each provider's terms
- [ ] FX rate source + refresh (canvas footer FX is a static date); data-age indicator if def L4 freshness question resolves yes
- [ ] Application-ready entry point: what the reviewer sees first, how to switch personas
- [ ] Send-readiness gate: minimum crawl-history days per provider before the link is sent, or explicit acceptance of 資料不足 / hidden states on day 1 (def L4 day-1 appearance) — canvas: real values need ~2 weeks

**Exit Criteria:** reviewer with only the link walks every canvas flow; data age stays within the agreed staleness; send-readiness gate met.

---

## Priorities

**Current focus:** Phase 1 — real advisor run (baseline commit done) + start accruing crawl history (history is the long pole for Phases 4–5; every real-derived surface now built waits on it). Fix R1 before the next crawl. Human answers on the open questions below unblock provisional choices the sprint shipped.

**Next up:** Phase 2 provider vetting + canvas pass 2 in parallel — single-provider canvas surfaces are built (sprint 2610130524-kura-interactive-mock-sprint.md); what remains of Phase 3 / 4 is pass-2-gated (multi-provider, persona layer, watchlist page, advisor error / cap states) or human-gated (def L2–L4).

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
- [ ] **價格最美 threshold** (sprint F1, 2610130602-kura-canvas-foundation-home-audit.md): shipped `deal_pct ≥ 0.05` for shelf, count, Browse, advisor; canvas sample count = ≥10%. Keep 5% as logged deviation | 10% everywhere | defer — one constant either way
- [ ] **價格最美 led by anomaly-boundary items** (sprint F6, same audit): top items sit at `deal_pct 0.500`, one tick under the anomaly cutoff — accept vs cap / penalise near the cutoff; also anomaly display 「低 50%」 vs legend 「低超過 50%」 (2610131134-kura-methodology-footer-audit.md)
- [ ] **大家都在找 ranking definition** (def L3 Home sections): blocks the shelf + model 「這週 N 人在找」 (Phase 5 mocked-demand milestone)
- [ ] **「尚無行情估價」 wording** for no-group / unmappable-grade listings (iter-7 F5, 2610131134-kura-methodology-footer-audit.md; def L4 grade ordering)
- [ ] **示範模式 tooltip** 「未設定 API 金鑰，顧問以範例資料回答」 (canvas copy) vs Mock mode's rule-based search over real data (iter-7 F6, same audit) — canvas pass 2 / def L3 reviewer-facing wording
- [ ] **R2** — 最新上架 clusters the first-crawled brand (GOYARD): suboptimal, not untruthful; orchestrator/def call (2610131037-kura-member-notifications-audit.md)
- Sprint provisional choices against def L2–L4 + canvas pass 2 candidates are recorded in 2610130449-kura-mock-def.md (Open Questions *Sprint provisional* lines; § Canvas pass 2 candidates)

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
| 2026-10-13 | **Revise (baseline committed):** Current Position, Known Blockers, Phase 1 baseline milestone, Priorities, Active Work updated — repo no longer has zero commits. — trigger: `git log` shows baseline `90ec1a6` + red team `eae09f3` on `master`. |
| 2026-10-13 | **Revise (sprint finalize):** Current Position rewritten (canvas UI built, single provider, goal-reached sprint); Recent Progress + Active Work updated; Known Blockers += R1 (fix before further crawls / Phase 6) + pending human answers. Phase 3 → Current: tokens/components, new IA, advisor panel, light/dark/mobile/a11y checked w/ execution pointers; 行情怎麼算 left open (grade-mapping note waits on def L4); canvas pass 2 not started; exit not met (gap items deferred, not designed). Phase 4 → In progress: notification feed checked; personas + watchlist partial; exit not met (no persona switch). Phase 5 → In progress: gallery not started; real-derived signals partial (gated, 30-day trend not built); mocked demand blocked on ranking definition. Phase 6 hosting milestone gains R1 prerequisite. Priorities updated. Open Questions += F1, F6, 大家都在找 ranking, 「尚無行情估價」 wording, 示範模式 tooltip, R2 + pointer to def provisional record. — trigger: 2610130524-kura-interactive-mock-sprint.md finalize (Current Position goal check, Iterations 1–8, Open Questions, Deferred), verified against iteration audits (2610130602-kura-canvas-foundation-home-audit.md, 2610131037-kura-member-notifications-audit.md R1/R2, 2610131134-kura-methodology-footer-audit.md, 2610131253-kura-closing-walk-a11y-audit.md). |
