# Red Team: Kura def + nav vs the design canvas

> **Status:** In Progress
> **Lead:** Claude Opus 5.5 (red team)
> **Parent:** Orchestrator
> **Subject:** Do 2610130449-kura-mock-def.md (Definition) and 2610130420-kura-roadmap-nav.md (Navigation) honor the Claude Design canvas "Kura 蔵 redesign — first pass" (https://claude.ai/artifact/6iAvGmwsajiW4X9XCD6THw)? | **Related:** 2610130449-kura-mock-def.md | 2610130420-kura-roadmap-nav.md | non-projex: `docs/design-handoff/HANDOFF.md`
> **Evidence base:** canvas version **1791839497-46ec** (19 artboards; `project/canvas.json`, `project/Main.dc.html`, `project/Card.dc.html`, `project/CardStates.dc.html`, `project/Tokens.dc.html`, 15 thin wrapper artboards that `<dc-import name="Main">` with props) | `internal/radar/radar.go` | `internal/advisor/tools.go`. Locators `project/*.dc.html:N` = line in that canvas file at that version.

---

## Bottom Line

**Verdict:** Fix Issues

The look and IA carry over well: tokens, the 4 canvas areas, 鑑定卡 with 店家鑑定/賣家自評 variants, link-out copy, the notification types and personas all match the canvas. The gaps are in data semantics and multi-provider scope. In those areas the def overrides or goes past the canvas while still calling the canvas "the goal", and nothing records where the def takes precedence.

**Top Vulnerabilities:**
1. **Def marks real, computable signals as mocked and leaves them unlabelled, against the canvas's intent** (F2, F19, F20). The canvas marks 大家都在找/trends/days-to-sell as layout samples that "need ~2 weeks of history" and has a `trends: hidden` state. The radar already computes `MedianDaysToSell` and `SellThrough`, and the advisor returns them. Under the def, the reviewer would see fabricated stats next to real ones, and the advisor would contradict the UI. That breaks the honesty thesis the design is built on.
2. **The canvas covers one provider; the def requires ≥3** (F3, F4, F14). Every canvas surface is BRAND OFF-shaped: same-shop valuation copy, a chart axis built on one shop's grade ladder, the CTA "到 BRAND OFF 商品頁", and a cross-shop 最低成色 filter. That filter needs an implicit house scale, which the established grading rule forbids. Neither doc flags the design gap or plans a design pass for it.
3. **"Canvas = latest goal" is an unpinned, moving, self-declared first pass** (F1, F5, F12). The def pins version 1791835276-abf4, but the live version is 1791839497-46ec. The nav pins nothing. Screens the roadmap needs (watchlist page, persona switcher, real-vs-mocked page, pipeline page, advisor error/cap states) are not on the canvas. Phase 3's exit criterion ("every canvas screen reachable") therefore cannot catch them.

---

## Stakeholder Roles

| Wave | Role | Cares About | Pain Points | Critical Assumptions |
|------|------|-------------|-------------|---------------------|
| 1 | Hiring reviewer (104 job 925vi) | Product thinking + design shown by the mock | Fake-looking data, dead links, unclear what is real | What they see is the intended product |
| 1 | Implementer (executes nav Phases 2–6 from def + canvas) | Unambiguous spec; one source of truth | Def and canvas disagree with no precedence rule | Def + canvas together are complete and consistent |
| 1 | Design owner (the canvas's intent, i.e. the user as designer) | Design decisions carried through intact | Intent silently overridden; first-pass gaps frozen as "goal" | Docs cite the canvas → docs honor the canvas |
| 2 | Providers / marketplace operators (BRAND OFF, Rakuten/Yahoo dealers) | Attribution, image/data terms, fair portrayal | UI that misattributes, hotlinks, or ranks them as "cheap/anomalous" | Canvas link-out pattern fits every channel |
| 2 | Hosted pipeline operator (hourly/daily refresh) | Producing the data each canvas surface needs | Surfaces requiring data the cadence can't give (sold state, history) | Realtime cadence covers every canvas data need |
| 2 | Advisor runtime (real Claude + spend cap) | Truthful tool-grounded answers | Helper steps with no real tool; cap → mode switch mid-demo | Canvas helper lines map to real tools |
| 3 | Skeptical evaluator (adversarial: a reviewer actively probing for fabrication) | Catching overclaim or fakery | n/a — succeeds on any inconsistency | Unlabelled numbers will be taken at face value |
| 3 | Provider terms / compliance (accountable) | Display rights, attribution, no disparagement | Public ranking + "異常低價" on a named dealer's goods | "Legitimate channel" for fetching = right to display |
| 3 | Applicant as accountable owner (the user) | Personal credibility of the application | Mock misrepresents capability or data provenance | The mock can be sent whenever it is built |

### Wave Derivation

- **Wave 1 → 2:** F3/F4 (multi-provider surfaces undesigned, grade-scale tension) → **Providers**. F2/F8 (signals reclassified; freshness invisible) → **Pipeline operator**, who must produce or withhold them. F7 (風格 helper reads mocked demand) → **Advisor runtime**.
- **Wave 2 → 3:** F2 + F17 + F18 (fabricated stats, stale link-outs, mode flips) → **Skeptical evaluator**, who turns any inconsistency into "the data is fake". F13/F15 (attribution, images) + Main.dc.html's public ranking and 異常低價 labels → **Provider compliance**. F2/F19 (unlabelled fabrication) + F22 (history timing) → **Applicant**, who answers for the mock personally.

## Roles Not Attacked

| Role | Surfaced by | What would have been asked |
|------|-------------|---------------------------|
| Taiwanese buyer persona (the fictional product audience, HANDOFF §1) | F19: the honesty thesis is aimed at them | Does the canvas copy (已找到新主人, 行情怎麼算) still read as honest once mocked numbers sit beside it? |
| Accessibility auditor | HANDOFF §3.7 vs canvas `role="switch"` checkbox, hover-only chart tooltip | Do the def or nav carry the a11y bar (nav Phase 3 says "a11y pass" with no criteria)? |
| Hosting / font CDN vendor | Phase 6 + canvas Google Fonts (5 families) | CJK font weight on mobile first load for a one-shot reviewer visit? |

---

## Attack Surface (Per Role)

**Hiring reviewer:**
- Claims to this role: hosted link; real realtime provider data; mocked user-side features; product thinking + design (def Identity).
- Assumptions: numbers on screen are either real or clearly hypothetical.
- Dependencies: real-vs-mocked explanation surface (def L2: "where: open"); freshness; working link-outs.

**Implementer:**
- Claims to this role: def = precise spec, canvas = target shape (nav Vision); Phase 3 exit "every canvas screen reachable … real where def says real, mocked where def says mocked".
- Assumptions: canvas and def never conflict; canvas covers every screen the roadmap needs.
- Dependencies: a stable canvas version; a precedence rule; designs for multi-provider and member surfaces.

**Design owner:**
- Claims to this role: def Scope says "Goal source of truth: Claude Design canvas"; nav says "user: canvas = latest goal".
- Assumptions: data-state intent encoded in the canvas (`trends` prop, `data: empty`, 示意數據 artboard titles, code comments) is carried over.
- Dependencies: the docs read the canvas source, not only its screenshots.

---

## Critical Findings

### Wave 1

### F1: Canvas version drift; "latest goal" has no pin
**Severity:** Medium | **Likelihood:** High
**Affects Roles:** Implementer, Design owner
**Attack Vector:** Def Scope pins `version 1791835276-abf4`. The live canvas is `1791839497-46ec`, so it has been republished since the def was written. The nav cites the canvas with no version. The canvas can be edited (comments, room, type-instructions allow hand edits), and the def says "canvas = latest goal", so the goal changes whenever someone saves.
**Role-Specific Impact:**
- **Implementer:** cannot tell which version an L3 answer or a Phase 3 milestone was checked against.
- **Design owner:** a later canvas edit becomes binding with no review, or is silently ignored.
**Blast Radius:** Every def L3 question seeded from "canvas evidence"; nav Phase 3 exit.
**Remediation:** Each def revision records the canvas version it read. The nav cites the same version. "Latest goal" becomes "goal as of version X; newer versions need a def revision".

### F2: Def mocks signals that the canvas meant to be real and the radar already computes
**Severity:** High | **Likelihood:** High
**Affects Roles:** Reviewer, Design owner, Implementer
**Attack Vector:** Def L2 lists "demand (「這週 N 人在找」), trends, days-to-sell, notification feed, persona watchlists" as **Mocked** and **unlabelled**. Nav Phase 5 says "goal flipped from 'replace 示意數據' to 'gallery real, signals mocked'". The canvas says otherwise:
- `project/Main.dc.html:636`: `示意數據 — sample values for layout only; real values need ~2 weeks of history`. The samples are placeholders for real values.
- `project/Main.dc.html:106-107,625`: `<!-- 大家都在找: hidden until demand data exists -->` plus a `trends` prop with a `hidden` option. The canvas designs a "not yet" state rather than permanent fake data.
- Artboard titles in `project/canvas.json` flag samples explicitly: "大家都在找 = 示意數據", "分布圖 = 示意數據", "型號頁 (示意數據)", "示意通知".
- HANDOFF §4.C: model stats show "資料不足 when there isn't enough data".
- Code: `internal/radar/radar.go:108-112,538-542` computes `SellThrough` and `MedianDaysToSell`. `internal/advisor/tools.go:263` returns `median_days_to_sell` to Claude. Nav Recent Progress itself lists "sell-through".
Only "N 人在找" (needs real users) and the notification feed truly need a service. Days-to-sell, sell-through, 30-day trend, 已售出 counts and the price distribution all derive from crawl history.
**Role-Specific Impact:**
- **Reviewer:** sees invented market stats with no label on a product whose pitch is "honesty as a design value" (HANDOFF §2).
- **Design owner:** the designed `hidden` and 資料不足 states are thrown away.
- **Implementer:** builds a fake-data path for stats the backend already produces.
**Blast Radius:** Home 大家都在找 shelf, model page stats (`project/Main.dc.html:658-663`), distribution chart, advisor answers (F20).
**Remediation:** Revise def L2 into three classes. **Real-derived**: days-to-sell, sell-through, trend, 已售出, distribution, each with the canvas's `hidden`/資料不足 state until history suffices. **Mocked, needs users**: N 人在找, notification feed, persona watchlists. **Mocked, needs service**. Revise nav Phase 5 to match.

### F3: Canvas covers one provider; def requires ≥3; no design pass planned
**Severity:** High | **Likelihood:** High
**Affects Roles:** Implementer, Design owner, Reviewer
**Attack Vector:** Def Boundaries require "≥3 providers" and also "canvas = goal". Every canvas surface assumes one shop:
- All sample items have `shop: 'BRAND OFF'` (`project/Main.dc.html:627-631`, `project/Card.dc.html` KURA_SAMPLES).
- Valuation prose: "同一家店、同樣是 ランクA …" (`project/Main.dc.html:320`). The method is "同店同等級中位數".
- 同型號價格分布: x-axis = one shop's ladder 新品/未使用品/ランクA/B/C (`project/Main.dc.html:645-651`). With 3 shops, one model has 3 incompatible ladders and the chart has no designed layout for that.
- Detail CTA "到 BRAND OFF 商品頁" and footnote "購買、付款與寄送都在 BRAND OFF 完成" (`project/Main.dc.html:271,274`).
- Browse brand counts sum to ~15,367, i.e. the BRAND OFF crawl.
Nav Phase 2 exit adds UI requirements the canvas lacks: "every listing shows provider + **channel** + its own grading; **cross-provider comps labelled**". Nav Phase 3 exit checks only "every canvas screen reachable", so it passes even though multi-provider surfaces were never designed.
**Role-Specific Impact:**
- **Implementer:** invents the multi-provider chart, model page and card provenance ad hoc during Phase 3/2 overlap (nav runs them in parallel).
- **Design owner:** the most product-thinking-heavy part of the mock (comparing shops with different grading) gets no design.
- **Reviewer:** sees either a single-shop UI over multi-shop data or improvised screens that don't match the polished canvas.
**Blast Radius:** Detail, Model, Browse, Card, advisor recommendation card.
**Remediation:** Add a nav milestone before Phase 3 IA work: "Canvas pass 2: multi-provider surfaces". It covers the card provenance line, per-shop distribution (facet per shop, or columns grouped by shop), cross-shop comp labelling, CTA copy per channel, and the model page with mixed schemes. Alternatively, the def explicitly declares how single-shop patterns extend.

### F4: The 最低成色 filter and 跨等級換算 need an implicit house scale
**Severity:** High | **Likelihood:** Medium
**Affects Roles:** Design owner, Implementer
**Attack Vector:** Established rule (user; HANDOFF §3.1): "Never relabel a shop's grade into a Kura grade." The canvas Browse filter offers `任何成色 / 未使用以上 / 輕微使用感以上 / 明顯使用感以上` (`project/Main.dc.html:186-188`; Tokens sheet shows "最低成色 輕微使用感以上"). The Tokens valuation copy says "中信心 · 跨等級換算". With more than one scheme, both features need an ordinal mapping of every shop's grades onto shared tiers, which is a house scale even if never shown as a label. Nav Phase 2 "Learned rank factors … once grades come from several schemes" also assumes cross-scheme comparability. Def L4 only asks to confirm "per-provider grades shown verbatim", which covers display and leaves filtering and conversion open.
**Role-Specific Impact:**
- **Design owner:** the grading principle is followed on the surface while a house scale sits underneath.
- **Implementer:** must invent a mapping table with no spec. For 賣家自評 and free-text grades (F14) there may be no defensible mapping.
**Blast Radius:** Browse filter, medium-confidence valuations, advisor 成色 helper.
**Remediation:** Def L4 decides this explicitly. Options: (a) per-provider grade filter, with the filter changing by selected shop; (b) a documented Kura mapping, disclosed as Kura's interpretation on 行情怎麼算, never shown as a grade label; (c) drop the cross-shop filter. Record which one the canvas filter becomes.

### F5: Member-layer screens the nav builds are not on the canvas
**Severity:** Medium | **Likelihood:** High
**Affects Roles:** Reviewer, Implementer
**Attack Vector:** Def lists "Watchlist + notifications" as a canvas area. The canvas has only the track toggles, a notification dropdown, `追蹤清單（4）` → `href="#"`, `通知設定` → `#`, and a `我的帳號` button with no menu (`project/Main.dc.html:77-78,84,88`). There is no watchlist page, persona switcher or sign-in-as-persona flow. Guests see `登入 / 加入會員` (`:44-45`), which suggests real accounts, while the def says personas only. Nav Phase 4 exit ("reviewer switches persona → sees watchlist") and Phase 6 ("how to switch personas") depend on screens that do not exist on the canvas.
**Role-Specific Impact:**
- **Reviewer:** the persona entry point is the first thing they touch (Phase 6 entry), and it is the one undesigned surface.
- **Implementer:** improvises the most-used demo control.
**Remediation:** Add these to a canvas pass 2: watchlist page, account/persona menu, and replacing 加入會員 with "以示範身分登入" (or equivalent). Def L3 persona questions get answered against that design.

### F6: Real-vs-mocked explanation surface undesigned; canvas footer already commits to pages the def calls "candidates"
**Severity:** Medium | **Likelihood:** High
**Affects Roles:** Reviewer, Design owner
**Attack Vector:** The def requires real-vs-mocked to be "explained elsewhere" and leaves the location open. Its candidate list treats About/FAQ/legal as possible extras. The canvas footer already links `關於 Kura / 行情怎麼算 / 常見問題 / 聯絡我們 / 隱私權政策 / 服務條款` (`project/Main.dc.html:484-489`), and detail links `行情怎麼算` again (`:325`). The pipeline page (HANDOFF §4.E, "our promise" page) was requested but not delivered by the canvas. Neither doc schedules 行情怎麼算, which is the canvas's own methodology explainer and the natural home for both the real/mock disclosure and the F4 mapping.
**Remediation:** Def resolves the Extra-pages question as: footer pages on the canvas are in scope; 行情怎麼算 (or 關於 Kura) hosts the real/mock disclosure. Nav Phase 3 gains a design + build milestone for it.

### F7: Advisor 風格 helper consults mocked demand
**Severity:** Medium | **Likelihood:** High
**Affects Roles:** Implementer, Reviewer
**Attack Vector:** Canvas helper line: `風格 — 對照「大家都在找」，看哪些款式耐看、好轉手⋯` (`project/Main.dc.html:550`). Nav Phase 3: "step lines ('helpers') mapped to real tool calls". Under the def, 大家都在找 demand is mocked, so either no real tool backs 風格, or a real Claude reasons over fabricated demand and presents it as grounds for a recommendation.
**Remediation:** Re-ground 風格 on real signals (sell-through/days-to-sell; see F2), or drop that helper line. Record the mapping helper → tool in def L3.

### F8: "Realtime" is invisible on the canvas
**Severity:** Medium | **Likelihood:** High
**Affects Roles:** Reviewer
**Attack Vector:** Def Identity makes hourly/daily realtime a defining property. The canvas has no "updated N minutes ago", crawl status or data-age indicator. The FX line is static: `匯率 1 JPY = 0.2018 TWD（2026/10/12）` (`project/Main.dc.html:493`). HANDOFF §4.A asked for a header status line with FX + date + advisor mode; the canvas moved FX to the footer and dropped mode. Neither doc adds a freshness surface or an FX refresh source.
**Role-Specific Impact:**
- **Reviewer:** cannot see the hardest engineering part of the mock (live multi-source ingestion), so the product-thinking signal is lost.
**Remediation:** Def adds a freshness invariant (data age + FX date visible). Nav Phase 6 adds an FX source/refresh. A canvas pass places the indicator.

### F9: Notification semantics: canvas promises a type the docs omit; mocked feed doesn't react to tracking
**Severity:** Low | **Likelihood:** Medium
**Affects Roles:** Implementer, Reviewer
**Attack Vector:** Model track copy: `有新上架或低於行情時通知你` (`project/Main.dc.html:358`). That implies a "低於行情" notification, but the def/nav type set is `降價 | 新上架 | 已找到新主人`. Separately, tracking toggles are interactive (nav Phase 4) while the feed is a fixed per-persona mock. If a reviewer tracks a new item, nothing in the feed ever reflects it.
**Remediation:** Def L3 fixes the type set (add 低於行情 or change the copy). Def states that the mocked feed is static, or derives a synthetic entry when tracking changes.

### F10: Demo-mode badge copy is for developers
**Severity:** Low | **Likelihood:** Medium
**Affects Roles:** Reviewer
**Attack Vector:** `示範模式` badge title: `未設定 API 金鑰，顧問以範例資料回答` (`project/Main.dc.html:518`). Nav Phase 6 "fallback to mock mode" when the spend cap hits will show this to a reviewer. It is also the one labelled mock surface, against the def's unlabelled policy, and no doc reconciles the two.
**Remediation:** Def lists the advisor demo mode as an explicitly labelled exception. Change the copy to reviewer-facing wording (e.g. "示範額度已用完，改用範例回答").

### F11: Nav pre-decides an open def question
**Severity:** Low | **Likelihood:** High
**Affects Roles:** Implementer
**Attack Vector:** Def L3 leaves open "Theme (light/dark) + layout (desktop/mobile) both required?". Nav Vision and Phase 3 already commit to "light/dark; desktop/mobile" and "Light/dark, mobile, a11y pass". The canvas also covers only some combinations: detail mobile exists only in dark; model, browse and advisor have no dark artboards.
**Remediation:** Close def L3 (HANDOFF §2 says dark mode "is required"), or mark the nav item as pending the def.

### F12: Def promotes a self-declared "first pass" to source of truth; first-pass gaps go untracked
**Severity:** Medium | **Likelihood:** High
**Affects Roles:** Design owner, Implementer
**Attack Vector:** Canvas title: "Kura 蔵 redesign — first pass". HANDOFF items not delivered and not tracked by either doc:
- pipeline page (§4.E)
- advisor error and streaming states (§4.D)
- side column 近期降價 (§4.A)
- header status line (§4.A)
- the "尚無行情估價" card state: `project/Card.dc.html` computes `isNone` but has no `<sc-if>` that renders it, so a no-estimate card shows no state (Tokens sheet does design that badge)
- 資料不足 model stats (§4.C)
Nav Phase 0 marks "Design brief + first design pass" done, and Phase 3's exit tests against the canvas, so these gaps can never fail a check.
**Remediation:** Nav adds a "canvas gap list" milestone (HANDOFF §4 vs canvas) under Phase 3, or a canvas pass 2. Def records which HANDOFF requirements the canvas supersedes and which still bind.

### Wave 2

### F13: The link-out/attribution pattern doesn't fit marketplace channels
**Severity:** Medium | **Likelihood:** Medium
**Affects Roles:** Providers, Implementer
**Attack Vector:** The canvas has one party per listing: shop name + "到 {shop} 商品頁". Nav Phase 2 prefers the Rakuten/Yahoo APIs, which have two parties (dealer + marketplace), possible affiliate URLs and mandatory credits (nav Open Q). The def says "every listing links to its provider's own page", but for an API listing that is the marketplace page, not the dealer's.
**Remediation:** Def defines "provider" (dealer vs channel) and the attribution line format. Canvas pass 2 adds the dealer-via-marketplace variant and any required credit.

### F14: 鑑定卡 assumes a published, ordered grade ladder
**Severity:** Medium | **Likelihood:** Medium
**Affects Roles:** Providers, Design owner
**Attack Vector:** The 鑑定卡 renders the label, zh meaning, verbatim JP definition and a ladder strip `新品 / 未使用 / A / B / C` (`project/Main.dc.html:280-297`). Nav Phase 2: on marketplace APIs, the "condition grade lives in free-text item name/caption → per-shop parsing". A parsed label may have no published definition and no known ladder. The canvas's 賣家自評 variant (Tokens) covers marketplace self-grading, but not "dealer grade extracted from free text, definition unknown".
**Remediation:** Def L4 defines a third condition state, "dealer grade, definition unavailable": label verbatim, no ladder, an honest note. Canvas pass 2 draws it.

### F15: The gallery depends on an unresolved image policy
**Severity:** Medium | **Likelihood:** Medium
**Affects Roles:** Providers, Reviewer
**Attack Vector:** The canvas detail and advisor show 5 photos per listing (`正面 背面 側面 底部 內裡`, `project/Main.dc.html:243,785`). HANDOFF says shop photos are "for layout realism only". Nav Open Q (hotlink vs cache) and def L3 (5 photos required?) are both open. If a provider's terms forbid display, the gallery falls back to the "店家照片準備中" placeholder permanently for that provider.
**Remediation:** Resolve before Phase 5. Def states the fallback (single photo / placeholder) as an accepted canvas deviation per provider.

### F16: Per-model stats mix observation windows across providers
**Severity:** Low | **Likelihood:** Medium
**Affects Roles:** Pipeline operator, Reviewer
**Attack Vector:** Canvas model stats read "已售出 2 件 · 自 10/13 起記錄" (`project/Main.dc.html:660`). Providers onboarded later start their history later. A per-model sold count or days-to-sell summed across shops then mixes windows, and the single "自 10/13" caption becomes false.
**Remediation:** Show the window per provider, or start the caption at the latest provider's start date.

### F17: Link-outs can go stale between refreshes
**Severity:** Medium | **Likelihood:** High
**Affects Roles:** Pipeline operator, Reviewer
**Attack Vector:** Def/nav: sold detection happens only on the **daily** full refresh; hourly sweeps "never mark items sold". For up to ~24 h a listing can show as for sale on Kura while the CTA lands on a sold or 404 page. The canvas has no "may have sold" state. Def L4 lists "provider listing disappears" as open.
**Remediation:** Def L4 resolves this. Options: a last-seen time on the card/detail, or a "上次確認 N 小時前" line near the CTA. Nav Phase 1 incremental mode records last-seen.

### F18: Spend cap → mode flip mid-session; no designed state
**Severity:** Low | **Likelihood:** Medium
**Affects Roles:** Advisor runtime, Reviewer
**Attack Vector:** The canvas advisor appears in `mode="claude"` (desktop) and `mode="demo"` (mobile artboards) as static choices. Nav Phase 6 cap → fallback means a live session can switch from real to scripted answers with no transition design, and no error state exists (HANDOFF §4.D asked for "error (gentle)").
**Remediation:** Canvas pass 2 adds advisor error and cap-reached states. Def L3 records the advisor mode policy for the hosted mock.

### Wave 3

### F19: The skeptical evaluator turns unlabelled fabrication into "the data is fake"
**Severity:** Critical | **Likelihood:** Medium
**Affects Roles:** Skeptical evaluator (gains), Reviewer, Applicant
**Attack Vector:** Capability: one question, "where does 這週 41 人在找 come from on a demo with no users?" The answer sits on the 行情怎麼算/about page, which the evaluator may never open, or in public source. The canvas brand language ("honesty as a design value", "uncertainty shown gracefully", HANDOFF §2) makes the contradiction sharper.
**Cost asymmetry:** seconds of probing undo the honesty narrative for the whole mock, including the parts that are real.
**Remediation:** Same as F2. Additionally, every remaining mocked number carries a discreet hover/footnote marker, or uses the canvas's `hidden` state. Revisit the def's "unlabelled" decision with the user and record this risk in the def.

### F20: Advisor and UI disagree on the same statistic
**Severity:** High | **Likelihood:** High (if F2 not fixed)
**Affects Roles:** Skeptical evaluator, Reviewer, Advisor runtime
**Attack Vector:** The real advisor tool returns `median_days_to_sell` from the radar (`internal/advisor/tools.go:263`). The UI model page shows mocked "平均 6 天找到新主人" (`project/Main.dc.html:638,662`). Asking the advisor "ピコタンロック 多快賣掉?" gives a different number from the tile beside it. The evaluator does not need source access to spot this.
**Remediation:** A single source for any stat shown in both places: real-derived (F2) or withheld.

### F21: Public ranking and anomaly labels on named dealers' goods
**Severity:** Medium | **Likelihood:** Low
**Affects Roles:** Provider compliance, Applicant
**Attack Vector:** The canvas publicly ranks a named shop's items under "價格最美" and stamps "異常低價 · 先確認狀況" (implying undisclosed defects; Tokens copy) on that shop's listings, alongside its photos. The def's "any legitimate channel" covers *acquiring* data, not *display* or comparative portrayal rights. Nav Open Q covers attribution/caching only.
**Remediation:** Phase 2's vetting rubric adds "display + comparative-labelling terms". Def Boundaries state display-rights compliance as part of "legitimate".

### F22: Send timing vs history accrual is undefined
**Severity:** Medium | **Likelihood:** Medium
**Affects Roles:** Applicant, Reviewer
**Attack Vector:** The canvas says real signals "need ~2 weeks of history" (`project/Main.dc.html:636`). The nav calls history "the long pole" but gives no send date or readiness gate. The def does not define what a day-1 mock shows: the canvas `trends: hidden`, 資料不足 or mocks. If the link is sent early, real-derived surfaces are empty, and the pull toward mocking them (F2) grows.
**Remediation:** Nav Phase 6 exit adds a readiness gate: minimum history days, or explicit acceptance of 資料不足 states. Def records the day-1 appearance.

---

## Role-Based Assumption Challenges

### Implementer: "Def and canvas together are a consistent spec"
**Challenge:** The def overrides canvas data semantics (F2) and pre-empts canvas scope (F3, F5) without a precedence rule.
**Counter-Evidence:** `project/Main.dc.html:106,636` vs def L2 Mocked list. Single-shop canvas vs def "≥3 providers".
**If Wrong:** Phase 3 builds to the canvas and contradicts the def, or the reverse, with no arbiter.
**Action:** Reject. Add a precedence clause: canvas governs look/IA; def governs data semantics and scope; conflicts are logged in the def.

### Design owner: "Citing the canvas = honoring it"
**Challenge:** The docs read the canvas's screens and areas, but not its encoded states (`trends`, `data`, `mode` props, source comments).
**Counter-Evidence:** Def L2 ignores the `hidden` variant; nav Phase 5 "flipped" the goal away from the canvas's stated intent.
**If Wrong:** Design intent lost silently.
**Action:** Validate. The def lists each canvas prop and state and its fate.

### Reviewer: "Numbers on screen are real or clearly hypothetical"
**Challenge:** The def intentionally leaves mocked numbers unlabelled.
**Counter-Evidence:** Def L2, "Mocked surfaces are unlabelled in the UI".
**If Wrong:** F19/F20.
**Action:** Relax. Keep the clean UI, but cut the mocked set to what truly needs users (F2) and add a non-intrusive marker.

---

## Role-Specific Edge Cases & Failures

### Reviewer: opens link on day 1 of a new provider
**Trigger:** Provider #3 onboarded days before sending.
**Role Experience:** Its listings have no estimate (few comps), no history and no distribution, and the card's "尚無行情估價" state is not rendered (F12).
**Recovery:** Possible.
**Mitigation:** Render the `none` card state; per-provider readiness gate (F22).

### Reviewer: clicks CTA on an item sold overnight
**Trigger:** Daily-only sold detection (F17).
**Role Experience:** Lands on the shop's sold or 404 page.
**Recovery:** Difficult (first impression).
**Mitigation:** last-seen line; tighter sold check for listings shown on home.

### Implementer: same bag on dealer site + Rakuten
**Trigger:** Nav Phase 2 cross-channel dedupe.
**Role Experience:** Which CTA, which grade wording and which 鑑定卡 issuer to show? The canvas has one slot.
**Recovery:** Possible.
**Mitigation:** Def rule: canonical listing + "also on" link (canvas pass 2).

### Reviewer: mobile dark detail vs light-only mobile model/browse
**Trigger:** Toggling theme on mobile (F11).
**Role Experience:** Undesigned combinations rendered by the implementer.
**Recovery:** Possible.
**Mitigation:** Token-driven theme; verify each combination in Phase 3 a11y pass.

---

## What's Hidden (Per Role)

**Omissions per role:**
- **Reviewer:** which numbers are fabricated (by design); data age (F8).
- **Implementer:** the canvas version honored (F1); the multi-provider design (F3); the grade-ordering rule (F4).
- **Design owner:** that the def reversed the canvas's `hidden`-until-real intent (F2, nav Phase 5 log "goal flipped").

**Tradeoffs per role:**
- **Reviewer:** a clean UI was chosen over transparency.
- **Design owner:** single-shop polish vs multi-shop truth. The canvas optimized the former; the def demands the latter.

**What holds (acknowledged):**
- Def areas = canvas rows A/B/C + advisor + 通知 ✓
- Link-out-only + "Kura 不販售商品" ✓ (`project/Main.dc.html:274`)
- Notification types match `KURA_NOTIFS` ✓
- Persona 林小姐 + 追蹤清單（4）+ 2 unread ✓
- Advisor `claude|demo` modes captured in def L3 ✓
- Nav "19 artboards, `--k-*` tokens" ✓ (`project/canvas.json`)
- Nav Phase 3 "condition tag both variants" ✓ matches the Tokens 店家鑑定/賣家自評 sheet
- LINE/EDM absent from canvas and excluded by both docs ✓
- Per-source verbatim grading on display: canvas complies (label + zh + JP original + issuer) ✓

---

## Scale & Stress (Role Impact)

**At 10x (≥3 providers, ~150k listings):**
- **Implementer:** the single-ladder distribution chart (F3) and cross-shop 最低成色 filter (F4) break first.
- **Reviewer:** the 大家都在找 shelf of 5 fixed mocked models stays static over 10x live data and looks frozen.

**At 100x (many marketplace dealers via API):**
- **Providers:** per-dealer free-text grade parsing (F14) means hundreds of schemes, and the 鑑定卡 ladder becomes meaningless.
- **Pipeline operator:** hourly sweeps across many channels inside each provider's rate terms. The freshness promise needs a per-provider cadence, and the UI shows none (F8).

---

## Remediation

### Must Fix (Before Proceeding)
- **Reclassify mocked vs real-derived signals** (affects: Reviewer, Design owner, Evaluator). Revise def L2 + nav Phase 5: days-to-sell/sell-through/trend/sold/distribution → real-derived with the canvas `hidden`/資料不足 states; mocked = only user-dependent surfaces. **Verify:** no stat appears in both advisor output and UI with different sources (F2, F19, F20).
- **Precedence + version pin** (affects: Implementer, Design owner). Def clause: canvas governs look/IA at version X; def governs data semantics/scope; conflicts logged. Nav cites the same version (F1).
- **Multi-provider design gap** (affects: Implementer, Design owner, Providers). Nav milestone "canvas pass 2" before Phase 3 IA: card provenance, per-shop distribution, CTA/attribution per channel, mixed-scheme model page, dedupe display (F3, F13, F14).
- **Grade-ordering decision** (affects: Design owner, Implementer). Def L4 chooses per-provider filter | disclosed mapping | drop filter, and states how 跨等級換算 squares with "no house scale" (F4).

### Should Fix (Before Production)
- **Member-layer screens** (Reviewer, Implementer) → watchlist page, persona menu, guest header copy designed (F5).
- **Disclosure + methodology page** (Reviewer) → 行情怎麼算/關於 scheduled; hosts the real/mock note (F6).
- **風格 helper grounding** (Implementer, Advisor runtime) → real signal or remove (F7).
- **Freshness + FX source** (Reviewer) → visible data age; FX refresh (F8).
- **Stale link-out handling** (Reviewer, Pipeline operator) → last-seen line (F17).
- **Canvas first-pass gap list** (Implementer) → HANDOFF §4 vs canvas tracked; render the `none` card state (F12).
- **Advisor states + copy** (Reviewer) → error/cap states; reviewer-facing demo badge (F10, F18).
- **Display-rights vetting** (Provider compliance) → add to Phase 2 rubric (F21).

### Monitor
- **Canvas edits after pin** (Design owner): on each canvas republish → def revision.
- **History vs send date** (Applicant): re-check at Phase 6. Gate on minimum history (F22).
- **Notification type set / reactive feed** (Implementer): at Phase 4 start (F9).
- **Theme × layout coverage** (Implementer): at Phase 3 a11y pass (F11).

---

## Final Assessment

**Soundness:** Fixable
**Risk:** Medium (High on the honesty axis until F2 is resolved)
**Readiness:** Needs Work

**Per-Role Readiness:**
- **Hiring reviewer:** Not Ready. Unlabelled fabricated stats contradict real data and the design's honesty thesis.
- **Implementer:** Not Ready. No precedence rule, no version pin, no multi-provider or member-layer designs.
- **Design owner:** Ready with Fixes. Visual/IA intent carried. Data-state intent (`hidden`, 資料不足) dropped.
- **Providers / compliance:** Needs Work. Attribution for marketplace channels and display rights undefined.

**Conditions for Approval:**
- [ ] Def L2 mocked list reduced to user/service-dependent surfaces; derived stats real or hidden (for Reviewer, Evaluator)
- [ ] Canvas version pinned in def + nav; precedence clause written (for Implementer, Design owner)
- [ ] Multi-provider design pass scheduled in nav before Phase 3 IA (for Implementer)
- [ ] Def L4 decision on cross-scheme grade ordering (for Design owner)

**No-Go If:**
- [ ] The mock ships with a stat shown in both the UI (mocked) and advisor output (real) with different values (impacts Reviewer, Applicant)
- [ ] Multi-provider data goes live on the single-shop chart/filter design unchanged (impacts Reviewer, Design owner)
