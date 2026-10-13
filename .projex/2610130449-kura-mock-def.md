# Definition: Kura Mock

> **Created:** 2026-10-13 | **Last Revised:** 2026-10-13
> **Author:** Claude Opus 5.5 (definer)
> **Parent:** User
> **Scope:** The Kura mock as a deliverable — what it is, for whom, what it shows. Goal source of truth: Claude Design canvas "Kura 蔵 redesign — first pass" (https://claude.ai/artifact/6iAvGmwsajiW4X9XCD6THw), **pinned at version 1791839497-46ec** (re-read 2026-10-13; def first written against 1791835276-abf4). A newer canvas version binds only through a def revision. Canvas vs def split: Boundaries § Canvas precedence.
> **Status:** In Progress
> **Related Projex:** 2610130420-kura-roadmap-nav.md (how we get there) | 2610130513-kura-def-nav-vs-design-canvas-redteam.md (canvas-drift findings this def was revised against) | 2610130524-kura-interactive-mock-sprint.md (built the mock against this def; provisional choices + canvas pass 2 candidates recorded here)
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

**Canvas deviations** (def or AA bar overrides canvas, or build deviates for a stated reason; canvas 1791839497-46ec):
- Providers: canvas designs 1 shop; def requires ≥3
- Data: canvas sample values (大家都在找 shelf, model stats, distribution, 示意通知) → replaced per content kinds above; real-derived stats real or in the hidden / 資料不足 state
- Grade filter: canvas 最低成色 (任何成色 / 未使用以上 / 輕微使用感以上 / 明顯使用感以上) works across shops only via Kura's internal rank → L4 invariant
- Photo-mat caption 「店家照片準備中」 in dark: canvas `project/Card.dc.html` hard-codes `#7D7262` for both themes → 3.76:1 on the dark mat `#ECE5D7` (AA needs 4.5:1 at 12px); built as token `--k-mat-note` = `#6B6055` in dark (4.89:1; the canvas's own light `--k-ink-3`), light unchanged (4.71:1 on `#FFFFFF`) — canvas pass 2 to confirm (deviation D1, 2610131207-kura-closing-walk-a11y-patch.md)
- Model page crumb `首頁 / {BRAND} {model}` replaces canvas `首頁 / 大家都在找` (shelf hidden; most visits arrive from detail) — restore when the shelf ships (2610130722-kura-model-page-plan.md A2; audit 2610130748-kura-model-page-audit.md H3)
- Model page 看全部 reveals the remaining 在售 cards in place, not → Browse: Browse cannot filter by model group, so its count would not match (model plan A9; 2610130913-kura-browse-page-plan.md A9)
- Model 平均找到新主人 note 「自 M/D 起記錄」 replaces canvas 「近 30 天」 (record start, not a 30-day window) (model plan A7)
- Browse 類型 options = 10 non-empty real categories vs canvas 3 (canvas 3 = top 3) (browse plan A4; audit 2610130938-kura-browse-page-audit.md F6)

**Canvas gaps** (def/nav need these; absent on canvas 1791839497-46ec):
- Multi-provider: card/detail provenance (dealer + channel); same-model distribution with one grade ladder per shop (canvas axis = one shop's ladder); cross-shop comp labelling; CTA + footnote copy per channel (canvas hard-codes 「到 BRAND OFF 商品頁」); dealer-via-marketplace attribution; same item on several channels; condition state for a grade with no published definition
- Member layer: watchlist page (`追蹤清單（4）` → `#`), account / persona menu, sign-in-as-persona (guest header shows 登入 / 加入會員), 通知設定 (notification dropdown footer link → `#`)
- Advisor: error state, spend-cap-reached state (HANDOFF §4.D asked for error + streaming)
- Pages: 行情怎麼算 content — built (provisional: copy plan-authored, look composed from existing canvas parts; canvas pass 2 copy/look review); 關於 Kura / 常見問題 / 聯絡我們 / 隱私權政策 / 服務條款 content — reachable as a pending view (「這一頁的內容還在準備中。」), body needs owner facts + canvas pass 2; 資料管線 pipeline page (HANDOFF §4.E)
- HANDOFF §4.A not delivered: header status line (FX + date + advisor mode); side column 近期降價
- Card 「尚無行情估價」 state: `project/Card.dc.html` computes `isNone` but renders nothing (Tokens sheet designs the badge) — built (card badge from the Tokens sheet); its wording for no-group / unmappable-grade listings remains open — iter-7 F5 (2610131134-kura-methodology-footer-audit.md), tied to L4 Grade ordering
- Theme × layout combos with no artboard: model, browse, advisor in dark; detail mobile in light → token-driven — checked: closing walk + a11y pass at light/dark × 1280/375 on every flow (2610131159-kura-closing-walk-a11y-plan.md and its patch 2610131207-kura-closing-walk-a11y-patch.md); findings fixed or logged there

**Canvas pass 2 candidates** (built provisionally in sprint 2610130524-kura-interactive-mock-sprint.md where the canvas draws no state or copy; recorded, not decided — pass 2 confirms or redesigns each; pointer = artifact that raised it):
- *Undesigned states:*
  - Detail **sold state** — built from card sold vocabulary (已找到新主人, 最後售價, faded mat); 0 real instances (2610130624-kura-listing-detail-plan.md A5)
  - Detail **multi-point 售價變化** — step line per HANDOFF §4.B/§6 in the strip's vocabulary; canvas draws no-change only; 0 real instances (detail plan A6)
  - Model **資料不足 tile** — same tile, value 資料不足 in `--k-ink-3` (HANDOFF §4.C asked; canvas draws samples only) (2610130722-kura-model-page-plan.md A11) | **0-active model** state (model plan A10)
  - Advisor **dark scrim** — `::backdrop` 60% `--k-paper-sunk` (canvas literal `rgba(30,25,21,.18)` is light-only) | no composer autofocus ≤767px / coarse pointer (2610130846-kura-advisor-panel-audit.md F6 / F7; 2610130854-kura-advisor-panel-audit-fixes-patch.md)
  - Browse **mobile 篩選與排序 open state** — in-place disclosure of the five desktop fields (canvas draws the button only) | **zero-result** line 「沒有符合條件的包。」 | non-deal headings 全部在售 / 最新上架 (canvas draws 價格最美 only) (2610130913-kura-browse-page-plan.md A7 / A8 / A2; audit 2610130938-kura-browse-page-audit.md F6)
  - 通知 **empty feed** 「目前沒有通知。」 | time-label rule (分鐘前 / 小時前 / 今天 HH:MM / 昨天 / M/D) | real photo thumb in the 48px mat (2610131010-kura-member-notifications-plan.md A12 / A6 / A11)
  - Dark × browse / notifications / advisor — token-driven, checked (theme × layout line above) (browse plan A12; member plan A14)
- *Copy the canvas doesn't give:*
  - Detail valuation prose — rank-adjusted, near / above sentences, low-confidence line, 信心 中/低, 方法 「跨等級換算」 (detail plan A4) | sold + single-point 售價變化 keeps forward-looking 「店家調價時，會畫在這條線上。」 (2610130634-kura-listing-detail-patch.md Notes; 2610130657-kura-listing-detail-audit.md)
  - Model **single-listing 中位數** — one price labelled `{grade} 中位數`, note `1 件比較` (2610130748-kura-model-page-audit.md H4); advisor position bar skips count < 2 for the same reason (2610130815-kura-advisor-panel-plan.md A5)
  - Model 平均找到新主人 label vs median value (disclosed on 行情怎麼算) | 「不到 1 天」 (model plan A7; 2610131104-kura-methodology-footer-plan.md A11)
  - 通知 **新上架 title naming** — item `model_name` vs model page heading name (2610131037-kura-member-notifications-audit.md F5) | drop / new body without the 低於行情 clause when not below (member plan A5)
  - 行情怎麼算 page copy + look (Canvas gaps § Pages above; methodology plan A4)
  - Detail distribution table **「店家定義」 header** — canvas copy verbatim; content is the shop's definition in Kura's translation (2610131134-kura-methodology-footer-audit.md, adjacent finding)
  - **示範模式 tooltip** 「未設定 API 金鑰，顧問以範例資料回答」 — canvas copy verbatim; Mock mode does rule-based search over real data, not sample data (methodology audit F6; also L3 advisor item below)

---

## Open Questions

*Sprint provisional* lines = what 2610130524-kura-interactive-mock-sprint.md shipped against a question still open (sprint Vision: answers out of sprint). Recorded, not answers — every question below keeps its checkbox state.

### L2 — Boundaries (remaining)
- [x] **Extra pages** — resolved 2026-10-13 by canvas precedence: canvas footer pages in scope; 行情怎麼算 hosts real-vs-mocked disclosure. Still open: 資料管線 pipeline page (HANDOFF §4.E, not on canvas) | reviewer intro → design/plan time (user: no preference)
  - *Sprint provisional:* neither built — 資料管線 = sprint Deferred (blocker: this question + canvas pass 2); no reviewer intro. Other five footer pages = pending view (2610131104-kura-methodology-footer-plan.md A1)
- [ ] Which concrete providers fill the ≥3 — decided by roadmap Phase 2 vetting, not here
- [ ] **Unlabelled mocks — confirm with user** (red team F19): keep fully unlabelled, or add a discreet marker (hover/footnote) on the remaining mocked numbers? Mocked set is now only user-dependent surfaces.
  - *Sprint provisional:* shipped unlabelled (current def rule); mocked set = member feed / watchlist / demand only, disclosed on `/methodology` 哪些是示範 (2610131134-kura-methodology-footer-audit.md; sprint open human item)
- [ ] **Confirm Real-derived reclassification with user** — days-to-sell + trends moved from Mocked (user-agreed list) to Real-derived on canvas intent + code evidence; user's criterion ("needs a running service with real users") unchanged.
  - *Sprint provisional:* built as Real-derived — 已售出 / 平均找到新主人 → 資料不足 until `radar.MinSoldHistory` (see L4 Day-1); advisor withholds sold stats on the same `SoldStatsReady()` (2610130846-kura-advisor-panel-audit.md H2 close). 30-day trend not built (lives in the hidden 大家都在找 tile)

### L3 — Structure (not entered; canvas evidence to confirm)
- [ ] Personas: canvas shows guest + one member 「林小姐」 with watchlist (追蹤清單（4）) and 2 unread notifications. How many personas, what each demonstrates? Persona entry point undesigned (Canvas gaps).
  - *Sprint provisional:* one persona = canvas 林小姐; 登入 signs in directly, no picker / form / switching (2610131010-kura-member-notifications-plan.md A1; sprint Open Question member-state entry, decided derive 6). Watchlist mocked: 2 models + 2 listings from canvas sample brands (member plan A2); session store `sessionStorage` `kura.watch`, guest picks merged at sign-in (member plan A9 / A10)
- [ ] Canvas header for guest shows 登入 / 加入會員 — with personas, does sign-up exist or only "sign in as persona" (e.g. 以示範身分登入)?
  - *Sprint provisional:* 登入 = sign in as 林小姐 (canvas wording kept); 加入會員 stays `#` (sprint Deferred sign-in-as-persona row)
- [ ] Tracking granularity on canvas: 追蹤這件 (listing) and 追蹤這個型號 (model). Both?
  - *Sprint provisional:* both built, one session store, guest-visible (2610130535-kura-canvas-foundation-home-plan.md A2; 2610130722-kura-model-page-plan.md A8; member plan A9)
- [ ] Notification types on canvas: 降價 | 新上架 (tracked model) | 已找到新主人 (tracked item sold). Complete set? Model track copy 「有新上架或低於行情時通知你」 implies a 低於行情 type too. Mocked feed static per persona, or does tracking an item add a synthetic entry?
  - *Sprint provisional:* three types (drop | new | gone) from observed facts on watchlist entries; 低於行情 = clause on drop / new body when below, not a type (member plan A3 / A5); feed static — tracking changes watchlist + count, not feed (member plan A8); un-tracked events hidden; 新上架 picks follow the shop's newest-first order within the initial sweep (2610131037-kura-member-notifications-audit.md F1 option b, F2; 2610131048-kura-newest-order-member-feed-patch.md)
- [ ] Advisor on canvas: modes `claude` | `demo`; "helpers" progress (行情 / 成色 / 風格) — real tool steps or scripted? Real Claude calls in the mock? Constraint: no helper presents mocked data as grounds — 風格 line reads 「對照「大家都在找」」 (mocked demand) → re-ground on real-derived signals (sell-through / days-to-sell) or drop the line. Hosted mode policy: spend cap → demo fallback; demo-badge copy is developer-facing (「未設定 API 金鑰，顧問以範例資料回答」) → reviewer-facing wording.
  - *Sprint provisional:* helpers = real tool steps only, no 風格 line (2610130815-kura-advisor-panel-plan.md A1 / A2); demo pill + tooltip verbatim canvas (advisor plan A10) — tooltip says sample data, Mock mode searches real data (2610131134-kura-methodology-footer-audit.md F6); error = functional fallback only (advisor plan A9); spend cap not built (sprint Deferred). Real Claude run still unverified (no API key)
- [ ] Detail page gallery: 5 photos (正面 背面 側面 底部 內裡) — required from every provider? Fallback where provider terms forbid display: single photo | 店家照片準備中 placeholder, as an accepted per-provider deviation.
  - *Sprint provisional:* single `image_url` per listing in the canvas mat (detail, advisor figure, 通知 thumb); `photo_pending` → 店家照片準備中; photos hotlinked from BRAND OFF (home plan A6; advisor plan A4)
- [ ] Home sections 大家都在找 / 最新上架 / 價格最美 — definitions of each ranking? (大家都在找 = mocked demand)
  - *Sprint provisional:* 價格最美 = `deal_pct ≥ 0.05` (radar `OnlyDeals`) sorted by deal — shelf, count, Browse, advisor share it; canvas sample count = ≥10% (2610130602-kura-canvas-foundation-home-audit.md F1, open) | shelf led by items at the anomaly cutoff (`deal_pct 0.500`) (same audit F6, open) | 最新上架 = `Engine.newer` (shop newest-first within initial sweep; R1 new-listing-sorts-last before hosting, R2 brand clustering — member audit) | 大家都在找 = canvas `trends: hidden` (home plan A1). Same question governs model-page 「這週 N 人在找」 — not rendered (model plan A4)
- [x] Theme (light/dark) + layout (desktop/mobile) both required — resolved 2026-10-13: yes. HANDOFF §2 ("Dark mode … is required"), §3.5 (desktop ~1280 + mobile 375), §4 ("desktop and mobile artboards, each in light and dark"). Uncovered combos → Canvas gaps.

### L4 — Detail (not entered)
- [ ] Invariants carried from current Kura (per-provider grades shown verbatim; same-grade = same provider; honest valuation states; attribution + link-out) — confirm they bind the mock
- [ ] **Grade ordering across schemes** — established (user, prior session; code `internal/listing/listing.go` `Rank`): Kura's internal N–C rank is a bridge only — min-condition filter + cross-grade estimates (跨等級換算); never shown as a grade label; every listing still shows its own shop's grade. Open: disclose the mapping on 行情怎麼算 as Kura's interpretation? | grades with no defensible mapping (賣家自評 free text, parsed grade with no published definition) — excluded from the filter?
  - *Sprint provisional:* Browse 最低成色 → rank floor `S` / `A` / `B` (legacy mapping); the 1 unranked listing drops out under any floor (2610130913-kura-browse-page-plan.md A5; browse audit F6) | `/methodology` describes the bridge, not the mapping (methodology plan A5) | 「尚無行情估價」 also covers no-group / unmappable-grade listings — its wording open (methodology audit F5) | missing-grade listing: 鑑定卡 + ladder omitted, excluded from distribution (2610130624-kura-listing-detail-plan.md A11) | ladder short labels 新品 / 未使用 / A / B / C = canvas copy via `Grade.Short`, full shop labels kept elsewhere (detail plan A3)
- [ ] Behavior when a provider listing disappears after the mock is sent — sold detection is daily only, so a CTA can land on a sold/404 page for up to ~24 h. Options: last-seen line near CTA (「上次確認 N 小時前」) | tighter check for listings on home
  - *Sprint provisional:* nothing built (no crawl during sprint). 已找到新主人 disclosed on `/methodology` as Kura's inference (methodology plan A7)
- [ ] "Provider" = dealer or channel? For marketplace-API listings (Rakuten/Yahoo) define the attribution line (dealer via marketplace, required credits) + same item on several channels (canonical listing + "also on")
- [ ] Freshness visible? Def L1 makes realtime defining, canvas shows no data age; FX is a static footer date. HANDOFF §4.A asked for a header status line (FX + date + mode)
  - *Sprint provisional:* no data age shown; `/methodology` states record start `since` + whether history reached `min_sold_history_days`, no last-crawl time (methodology plan A8); status line = sprint Deferred
- [ ] Day-1 appearance: what real-derived surfaces show before ~2 weeks of history (hidden / 資料不足), per provider; per-model stats mixing providers onboarded on different dates (canvas caption 「自 10/13 起記錄」)
  - *Sprint provisional:* `radar.MinSoldHistory` = 14 d crawl-history span, global (not per-shop), no min-sample rule; gates model 已售出 + 平均找到新主人 and advisor sold stats (2610130748-kura-model-page-audit.md H1; 2610130757-kura-model-page-audit-fixes-patch.md) | 上架 = `first_seen` → initial-sweep listings read 在架第 1 天 (detail plan A7) | 新上架 events on day-1 data = newest *seen* (member plan A4)

---

## Revision Log

| Date | Summary |
|------|---------|
| 2026-10-13 | Initial definition: L1 Identity from user statements; L1 confirmations + L2–L4 open, seeded from canvas evidence. |
| 2026-10-13 | L1 round: hosted demo w/ realtime provider data; hosted link delivery; takeaway = product thinking + design. Identity rewritten; freshness question moved to L1 ("realtime" staleness). |
| 2026-10-13 | L1 agreed. Realtime = hourly new/price changes + daily full refresh. L2 partial: all 4 canvas areas in scope; real vs mocked content split (service-dependent surfaces mocked). Boundaries section added. Status Draft → In Progress. |
| 2026-10-13 | L2 round: mocked surfaces unlabelled (explained elsewhere); link-out only, Kura never sells; ≥3 providers; extra pages = no preference → open. L2 resolved pending restatement agreement. |
| 2026-10-13 | **Revise (canvas drift):** canvas pinned at 1791839497-46ec (was 1791835276-abf4) + precedence rule (canvas = look/IA/states; def = data semantics/scope); content split into Real / Real-derived / Mocked — days-to-sell + trends no longer mocked, shown real or in canvas hidden/資料不足 state; one source per stat (UI = advisor); "legitimate" covers display terms; Canvas deviations + Canvas gaps lists added; extra pages resolved by canvas footer (行情怎麼算 hosts disclosure); theme × layout resolved (both required, HANDOFF §2/§3.5/§4); L3/L4 gained notification 低於行情, 風格 helper grounding, advisor cap/badge policy, gallery fallback, grade-ordering (internal rank, established), stale link-outs, provider = dealer vs channel, freshness, day-1 appearance; two user confirmations queued (L2). — trigger: 2610130513-kura-def-nav-vs-design-canvas-redteam.md F1–F22, re-verified against canvas 1791839497-46ec (`project/Main.dc.html`, `project/Card.dc.html`, `project/canvas.json`), `internal/radar/radar.go`, `internal/advisor/tools.go`, `docs/design-handoff/HANDOFF.md`. |
| 2026-10-13 | Canvas gaps § Pages: 行情怎麼算 built (provisional) + other five footer pages reachable as pending view. No L2–L4 question decided (F19 unlabelled-mock confirmation, grade-mapping disclosure, freshness stay open). — trigger: 2610131121-kura-methodology-footer-patch.md (Sprint 2610130524-kura-interactive-mock-sprint.md — iteration 7) |
| 2026-10-13 | Canvas gaps: theme × layout → checked (walk + a11y pass, 4 combos); 「尚無行情估價」 → built; Member layer += 通知設定 (inventory: link → `#`, as sprint Deferred). Canvas deviations += D1 dark photo-mat caption colour (AA). No L2–L4 question decided. — trigger: 2610131207-kura-closing-walk-a11y-patch.md (Sprint 2610130524-kura-interactive-mock-sprint.md — iteration 8) |
| 2026-10-13 | Canvas gaps 「尚無行情估價」 built line += pointer: wording for no-group / unmappable-grade listings open (iter-7 F5). Record only; nothing decided. — trigger: 2610131253-kura-closing-walk-a11y-audit.md Minor #1 via 2610131307-kura-back-scroll-restore-patch.md (Sprint 2610130524-kura-interactive-mock-sprint.md — iteration 8) |
| 2026-10-13 | **Revise (sprint finalize):** record only — no question decided, no checkbox changed. Open Questions L2–L4 gain *Sprint provisional* lines (what the sprint shipped against each open question, with pointers); Boundaries += **Canvas pass 2 candidates** (undesigned states + copy the canvas doesn't give, built provisionally: detail sold state, multi-point 售價變化, 資料不足 tile, 0-active model, dark advisor scrim, Browse mobile filter open state / zero-result / headings, empty 通知 feed, valuation prose, single-listing 中位數, 平均 vs median, 新上架 title naming, 「店家定義」 header, 示範模式 tooltip); Canvas deviations += model crumb, model 看全部 in place, 「自 M/D 起記錄」, Browse 類型 set; deviations header reworded to cover D1 (AA bar). — trigger: sprint 2610130524-kura-interactive-mock-sprint.md finalize (Iterations 1–8 + Open Questions + Deferred), verified against iteration plans' A-items and audits (2610130657-kura-listing-detail-audit.md | 2610130748-kura-model-page-audit.md H3 H4 | 2610130846-kura-advisor-panel-audit.md F6 F7 | 2610130938-kura-browse-page-audit.md F6 | 2610131037-kura-member-notifications-audit.md F5 | 2610131134-kura-methodology-footer-audit.md F5 F6 | 2610131253-kura-closing-walk-a11y-audit.md Minor #2) |
