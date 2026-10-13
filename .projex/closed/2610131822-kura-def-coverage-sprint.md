# Sprint: Kura Def Coverage

> **Status:** Complete (Goal Reached)
> **Goal:** Complete coverage of the def (2610130449-kura-mock-def.md): every def clause is built, or deferred with a named blocker that needs human input. Providers via open alternatives or simple crawler services (user: "Find open alternatives or build simple crawler services."). Def open questions deferred (user: "Defer them").
> **Base Branch:** master
> **Step:** standard
> **Strategy:** plan → patch! → audit → [patch]
> **Sprint Branch:** projex/sprint/2610131822-kura-def-coverage
> **Worktree Path:** kura/.projexwt/2610131822-kura-def-coverage
> **Parent:** 2610130449-kura-mock-def.md
> **Started:** 2026-10-13 18:22

## Vision

Done = every clause of 2610130449-kura-mock-def.md (Identity, Boundaries § Is / Is not / content kinds / Canvas deviations / Canvas gaps) maps to either built + verified behaviour on the sprint branch, or a Deferred row naming the human input it waits on. Baseline: master `8ab31cb` (canvas UI mock from sprint 2610130524-kura-interactive-mock-sprint.md + seeded 大家都在找 + OpenAI advisor mode).

**User scope decisions (2026-10-13, at setup):**
- **Providers (def: ≥3 = BRAND OFF + ≥2):** reach ≥3 via **open alternatives** (open / keyless data sources) **or simple crawler services**. No marketplace-API keys (Rakuten / Yahoo need user-registered app IDs → not this sprint's channel). Crawl invariants unchanged: honestly identified bot UA, never spoofed; robots.txt + Crawl-delay obeyed; no CAPTCHA/WAF evasion; no C2C scraping against ToS; a provider whose terms forbid crawling or the display Kura makes (def: "legitimate" covers display — attribution, image use, comparative labelling) is dropped, not worked around. Each provider registers its own `listing.Scheme` (own grades + verbatim definitions, 店家鑑定 vs 賣家自評) — never relabelled into another shop's scale.
- **Open questions L2–L4 + roadmap-nav open questions: deferred.** A build item an open question blocks → Deferred with that question's id; never decided in-sprint. Already-shipped provisional choices stay as they are.

**Human-input blockers known at setup** (→ Deferred, not improvised): hosting deploy (user account) | real advisor run (API key) | canvas pass 2 design (def: canvas gaps are designed before built) | crawl-history-dependent values (calendar time: `radar.MinSoldHistory` 14 d, 30-day trend) | marketplace-API providers (user-registered keys).

**Verification:**
- Clause-by-clause map def → evidence (code path / test / browser check) or Deferred row with blocker
- `go vet ./...` + `go test ./...` green on the sprint branch
- App starts from the worktree; new/changed surfaces checked in the browser (light/dark × desktop/mobile where UI changes)
- Each new provider: robots.txt + terms check recorded, identified UA, fixtures + parser tests, its own Scheme

Out of sprint: deploying to a host, entering credentials, canvas pass 2 design, answering def/nav open questions, pushing to any remote.

## Current Position

**As of 2026-10-14 (iteration 6) — goal reached:**

Sprint branch head `0844630` = iterations 1–5. At derive: `go vet ./...` clean, `go test -count=1 ./...` green in all 10 test packages (cmd/kura, advisor, crawl, listing, radar, source/allu, source/brandoff, source/grare, store, web). The app builds and starts from the worktree: `kura serve` on a scratch DB, `/` 200, `/api/models` 200, advisor in mock mode (no key). Iteration 5 changed no UI surface, so no new browser check is due. The last one was iteration 4: light/dark × desktop/375.

*Iteration 5 closed (d′).* All three providers now have a sold signal:
- BRAND OFF: complete sweeps → `SweepMissing`, plus `SOLDOUT` list tiles → `ListPage.Sold`.
- ALLU: list rows with `button_type` 6 → `sold`, plus budgeted product-page re-checks.
- GALLERY RARE: budgeted product-page re-checks (`-recheck-after` 24h, from what `-details` leaves).

Each feeds `store.MarkGone` (active only). The Real content kind's "listings turn gone" now holds per provider. The cadence that runs it is Deferred (Realtime row).

**Clause map, def 2610130449 → built (B) or Deferred (D):**

| Def clause | State | Evidence / blocker |
|---|---|---|
| L1 interactive mock of Kura (radar + advisor) | B | baseline `8ab31cb` (sprint 2610130524); `internal/web`, `internal/radar`, `internal/advisor` |
| L1 hosted demo, link in application | D | Deferred row 1 (hosting account) |
| L1 real data, BRAND OFF + more, crawled | B | `internal/source/{brandoff,grare,allu}`, `-sources` default `brandoff,grare,allu`; iterations 3–4 scratch crawls |
| L1 realtime (hourly / daily full refresh, kept current while hosted) | D | Deferred rows 1 + Realtime cadence (owner scope "only work on providers"). Per-provider sold observation it relies on: B (iteration 5) |
| L1 demo personas | B provisional / D | 林小姐 (`/api/member`); rest → Deferred member-layer row (def L3) |
| L1 canvas = goal; takeaway product thinking + design | B / D | canvas areas built (prior sprint); canvas pass 2 → Deferred rows |
| L2 canvas precedence, deviations logged | B | def § Canvas deviations (9 entries, each with pointer) |
| L2 Is: ≥3 providers, legitimate acquisition | B | `docs/provider-vetting.md` Outcome 3/3. Robots.txt + terms recorded per shop, identified UA, own `listing.Scheme` with verbatim definitions, fixtures + parser tests per adapter |
| L2 Is: display (attribution + link-out, owner "Accept as demo") | B | owner decision (Open Questions, resolved); every listing links to its shop page |
| L2 Is: comparative labelling within provider terms | B | vetting (d): no comparison/ranking clause, recorded for BRAND OFF, GALLERY RARE and ALLU |
| L2 Is: link-out only for commerce | B | listing URLs to shop pages; no cart/payment code |
| L2 Is: canvas areas Home+Browse / Detail+Model / Advisor / Watchlist+notifications | B; gallery D; real advisor run D | prior sprint; Deferred rows gallery, advisor key, advisor cap/error |
| L2 Is: footer pages reachable; 行情怎麼算 disclosure | B; bodies D | `/methodology`; Deferred footer-bodies row |
| L2 content kinds: Real | B (cadence D) | 3 adapters + sold observation (iteration 5) |
| L2 content kinds: Real-derived | B gated / D values | `radar.MinSoldHistory` 資料不足 states; Deferred real-values row (calendar time) + per-provider gate row (def L4 Day-1) |
| L2 content kinds: Mocked, unlabelled, demo badge exception, one source per stat | B | prior sprint (`wanted.go`, `member.go`, advisor `SoldStatsReady`) |
| L2 Is not (LINE/EDM, real auth, user-data service, shop) | B | none built, by construction |
| L2 Canvas deviations | B | each built as recorded in the def |
| L2 Canvas gaps | D, except those the def marks built | Deferred multi-provider / member / advisor / pages / HANDOFF §4.A rows. Built per def: 行情怎麼算, 尚無行情估價 badge (wording → L4 Grade ordering), theme × layout check |
| L2 Canvas pass 2 candidates | B provisional | def lists them as built and awaiting pass 2 (Deferred: canvas pass 2) |
| Def Open Questions L2–L4 | D | user: "Defer them". Shipped provisional answers stay. Per-provider gate → Deferred with L4 Day-1 id |

All clauses are B or D. No D row is missing a named human blocker.

*Not goal-blocking (residuals for finalize):*
- **(e) ALLU F2 code**: 42/167 ALLU rows (25%) have no model or code, so they get no 行情. No def clause requires every listing to be valued. The def's built 「尚無行情估價」 state covers no-group listings, and README discloses the share. This is quality work, not coverage. It is dropped from the buildable remainder.
- **N1 / N2** (iteration-5 re-audit, Low, "fold into … finalize clean-up"): fix one README sentence (`README.md:130` "only after … two consecutive complete sweeps"), and pin BRAND OFF's `offset+rows >= total` branch with a stated hit-count assumption. Offline, about 10 lines. A good candidate for finalize's optional sprint-scope audit or clean-up. It does not change any clause's B/D state.
- **F5**: failing re-checks stay first in `RecheckDue` order. Needs design, Low.
- `sale_status` / stock drift is silent (informational).
- Owner-held items stay in Open Questions: GALLERY RARE paging, ALLU consignments (F1), 第18条 (12) reading (F3), stop-on-403 (F7). If F3 is overruled, ALLU drops and the ≥3 clause reopens. That is a human decision after this sprint.

**Closing assessment (finalize, 2026-10-14):** goal reached — accepted at orchestrator review of derive 6. 6 iterations: 4 Success (1, 3, 4, 5), 1 Failure (2, vetting record retained → unblocked by owner display decision), 1 goal-reached derive (6). Def coverage: every clause B or D (table above); 11 Deferred rows, each naming its human input. Providers 1 → 3 (BRAND OFF + GALLERY RARE + ALLU公式), each with own `listing.Scheme`, vetting record, fixtures + parser tests, and a sold signal. Sprint-scope audit: skipped — every iteration carries its own audit (iterations 4–5 also a re-audit after [patch]) and derive 6 re-ran vet/tests + clause map. Residuals (not goal-blocking): N1, N2, F5, (e) ALLU title fallback; 4 owner questions open below — ALLU F3 overrule would reopen the ≥3 clause. Close: squash into `master` (human-confirmed 2026-10-14).


## Iterations

| # | Objective | Result | Artifacts |
|---|-----------|--------|-----------|
| 1 | Multi-source crawl → radar path (crawler runs N `crawl.Source`s, per-source details budget, backfill every Scheme) + R1 fix: per-(source, target) onboarding epoch from first complete sweep in `crawl_runs` | Success — audit Accept (no Critical/Significant); F1–F4 fixed in [patch]; I1 (global `Since`/`SoldStatsReady` for later providers) + I2 (roadmap nav R1 entry stale) open, informational | 2610131831-kura-multi-source-onboarding-epoch-plan.md, 2610131843-kura-multi-source-onboarding-epoch-patch.md, 2610131852-kura-multi-source-onboarding-epoch-audit.md, 2610131858-kura-onboarding-epoch-audit-fixes-patch.md |
| 2 | Onboard provider #2 (ALLU) end to end, or record its rejection | Failure — ALLU fails display (ToS 第18条 (5) + site policy); split to child 1 = vetting record: no candidate passes (Brandear / RAGTAG blocked, GALLERY RARE no license, RECLO maintenance); incumbent BRAND OFF fails the same bar (利用規約 第13条 2項). Audit Accept w/ conditions; F2–F8 fixed; F1 (README wording on BRAND OFF terms) held for owner. ≥3 providers now waits on owner decision (`docs/provider-vetting.md` § Outcome i–iii) | 2610131913-kura-provider-2-vetting-plan.md, 2610131927-kura-provider-2-vetting-patch.md, 2610131935-kura-provider-2-vetting-audit.md, 2610131940-kura-vetting-audit-fixes-patch.md |
| 3 | Onboard provider #2 = GALLERY RARE end to end under the restated display bar | Success — `internal/source/grare` (own 9-rank Scheme, verbatim defs), `-sources` flag, per-shop ladders in distribution (`/api/models` `schemes`), advisor shop line from schemes, README + vetting doc re-read; scratch-DB crawl: 200 listings, no `?` URLs. Audit Accept w/ conditions; [patch] fixed F1 small part (out-of-stock notice → sold) + F2–F5, F8; F6/F7 accepted. Open: F1 product-page re-check for never-complete sweeps; page-1-only depth (`/*?` courtesy limit, owner question) | 2610132140-kura-grare-provider-2-plan.md, 2610132204-kura-grare-provider-2-patch.md, 2610132219-kura-grare-provider-2-audit.md, 2610132229-kura-grare-audit-fixes-patch.md |
| 4 | Onboard provider #3 = ALLU (ALLU公式 only) end to end, or record its Fail with ToS evidence | Success — ToS re-read (round 2, 3 GETs): 第18条 (12) read as BOT-to-manipulate, not access ban → R1 Pass (Kura's reading); R2 (d) no comparison clause → Pass; vetting Outcome 3/3. `internal/source/allu` (Nuxt devalue decoder, own 9-rank N…J Scheme, verbatim defs, sold = `button_type` 6, 10-page newest-first cap), `-sources` default `brandoff,grare,allu`; scratch crawl 12 GETs all 200 → 167 rows, cross-shop groups with BRAND OFF; browser check light/dark × desktop/375. Audit Accept w/ conditions; [patch] fixed F1–F3 doc parts + F5, F6, F8–F10; re-audit Accept. Held for owner: F1 keep/exclude partner consignments (listed under ALLU公式, not flagged on list rows), F3 confirm (12) reading, F7 stop-on-403 scope. Open: F2 code (25% ALLU rows ungrouped — title-derived model fallback), F4 (sold rows on list pages → plan with remainder (d)) | 2610132307-kura-allu-provider-3-plan.md, 2610132325-kura-allu-provider-3-patch.md, 2610132339-kura-allu-provider-3-audit.md, 2610132349-kura-allu-audit-fixes-patch.md, 2610132357-kura-allu-audit-fixes-audit.md |
| 5 | Never-complete sources (GALLERY RARE, ALLU) observe sales: product-page re-check cadence inside each source's detail budget (first fetches first) + ALLU list-visible sold rows → `store.MarkGone` via `crawl.Source` contract addition | Success — `MarkGone` active-only `(bool, error)`; `RecheckDue` (oldest read first, skip rows seen this crawl); `Sweep` `visit` gains `sold []string`, `run.Gone +=`; `-recheck-after` (24h, 0 = off), re-checks only for targets whose sweep did not complete, leftover of `-details`; no shop requests (fixtures only). Audit Accept w/ conditions — F1 (BRAND OFF `SOLDOUT` tiles stored as active, 7/90 in fixture) disproved plan Assumption 1; [patch] fixed F1 (SOLDOUT → `ListPage.Sold`; raw tile count kept for last-page check — avoids false gone past page 1) + F2 (re-check volume stated: up to `-details` per source per crawl) + F3 + F4; re-audit Accept. All 3 providers now carry a sold signal. Open (Low): N1 README "only after … two consecutive complete sweeps" still literal; N2 `offset+rows >= total` branch untested + unstated SOLDOUT-in-count assumption; F5 failing re-checks stay first (needs design). | 2610140010-kura-sold-observation-plan.md, 2610140023-kura-sold-observation-patch.md, 2610140037-kura-sold-observation-audit.md, 2610140046-kura-sold-observation-audit-fixes-patch.md, 2610140055-kura-sold-observation-audit-fixes-audit.md |
| 6 | — (derive: goal reached) | Goal reached — clause map def → B/D complete, every D row names its human blocker; vet + tests green, app starts from worktree. Finalize: squash-close | — |

## Deferred

| Item | Def clause | Blocker (human input) |
|------|-----------|------------------------|
| Hosting + deploy; scheduled refresh running while hosted | L1 hosted demo / link; § Is hosted; L1 Realtime (running part) | User hosting account + deploy decision (roadmap Phase 6) |
| Real advisor run (Claude / OpenAI) | § Is canvas areas — Advisor (`claude` mode) | API key (user) |
| Advisor spend cap + cap-reached / error designed states | Canvas gaps § Advisor | Def L3 hosted-mode policy + canvas pass 2 + hosting |
| Persona count / switching / sign-in-as-persona / 加入會員, watchlist page, account menu, 通知設定 | L1 personas; Canvas gaps § Member layer | Def L3 personas + canvas pass 2 |
| Multi-provider provenance UI (dealer + channel), per-shop ladders as designed, cross-shop comp labels, per-channel CTA copy, "also on", grade-without-definition state | Canvas gaps § Multi-provider | Canvas pass 2 (def: designed before built). Data-correctness fixes (R1, dist.js per-source dots) are **not** deferred — in Current Position |
| Detail multi-photo gallery | § Is Detail (gallery) | Def L3 gallery question + roadmap image-policy question; provisional single photo stands (Vision) |
| Real values for real-derived stats (已售出, 平均找到新主人, 30-day trend) | Content kinds § Real-derived | Calendar time — `radar.MinSoldHistory` 14 d / 30 d of crawl history; sprint ships gated states only |
| Footer page bodies (關於 Kura, 常見問題, 聯絡我們, 隱私權政策, 服務條款); 資料管線 page; HANDOFF §4.A status line + 近期降價 | Canvas gaps § Pages, § HANDOFF §4.A | Owner contact/legal facts + def L2 (pipeline page) + L4 (freshness) + canvas pass 2 |
| Marketplace-API providers (Rakuten / Yahoo) | § Is ≥3 providers (one channel) | User-registered app IDs; out of this sprint's channel (Vision) |
| Realtime two-tier cadence (hourly new + price changes; daily full refresh; second-crawl drop / sold observation) | L1 Realtime | Owner scope decision 2026-10-13 "only work on providers" (sprint Revision Log 2→3); running it continuously also needs hosting (row 1) |
| Per-provider history gate: `Overview.Since` / `SoldStatsReady` per source (iteration-1 audit I1); per-model sold stats mixing shops onboarded on different dates | Content kinds § Real-derived; def L4 open question **Day-1 appearance** | Def L4 Day-1 appearance open question ("per provider; per-model stats mixing providers onboarded on different dates"). Its shipped provisional answer is a global 14 d gate (not per-shop), which stays. Live on the base `kura.db`: it holds BRAND OFF only from 2026-10-12, so GALLERY RARE and ALLU onboard later |

## Open Questions

- [x] **Display bar vs shop terms** (iteration 2) — resolved 2026-10-13 by owner: "Accept as demo" — the portfolio demo may show shop titles, photos and verbatim grade definitions with attribution + link-out despite no-reproduction clauses / silent terms; owner's decision, not a legal judgment by Kura. Acquisition bar unchanged (identified UA, robots.txt + Crawl-delay, no evasion; terms that prohibit bots/automated access still fail). README F1 wording follows this decision.

- [ ] **GALLERY RARE paging** (iteration 3): KuraBot reads only page 1 (≤40 bags/brand) — the vetting record keeps a `/*?` courtesy limit because robots.txt bars `?` URLs for the bots it names (not `*`). Owner: may KuraBot page with `?page=`? Until then: shallow coverage, sweeps never complete.
- [x] **Sold re-check for never-complete sweeps** — resolved by finalize: built in iteration 5 (`-recheck-after`, `RecheckDue`, ALLU + BRAND OFF list-visible sold rows) (iteration 3 audit F1, open part): GALLERY RARE product pages are fetched once, so later sales stay 在售. Needs a periodic product-page re-check (plan-sized) — next derive, or Deferred if the paging answer removes the need.
  - *Derive 5:* taken as the iteration-5 objective, together with ALLU's list-visible sold rows (F4). The paging answer cannot remove the need: ALLU's 10-page cap is also never-complete. No owner input needed; left open for finalize to tick.
- [ ] **ALLU partner consignments: keep or exclude** (iteration-4 audit F1). ALLU公式 includes ToS 第四章 consigned items (`item_type_key: consigned_item`), which are visible only on the product page. They are stored as ALLU-graded 店家鑑定 and disclosed in README + vetting R3. Owner: keep with disclosure (current) or exclude? Exclusion is plan-sized: a detail-time exclusion path that `MarkGone` must not count as sold.
- [ ] **ALLU ToS 第18条 (12) reading** (iteration-4 audit F3). Kura reads (12) as banning BOT use to 不正に操作 the service, not read-only access, so R1 is Pass. Owner: confirm or overrule. Overrule → ALLU fails, the ≥3 clause reopens (→ Deferred: permission / API key / new candidate).
- [ ] **Stop-on-403 scope** (iteration-4 audit F7, pre-existing, all sources). The crawler keeps going after a 403: `FetchDetails` spends the rest of the budget and `RunAll` moves to the next brand. The acquisition bar's stop is honoured only by manual runs. Owner: crawler-wide per-source stop, or provider-local sticky stop, or leave as is (scope is "only work on providers").

## Revision Log

| Date | Iter | Summary |
|------|------|---------|
| 2026-10-13 | — | Sprint started from master `8ab31cb`; scope decisions recorded in Vision (providers via open alternatives / simple crawlers; open questions deferred) |
| 2026-10-13 | 1 | Derive: clause map def → built / missing / deferred; 9 Deferred rows (hosting, advisor key, advisor cap/error, member layer, multi-provider canvas gaps, gallery, real-derived values, footer/pipeline/status-line, marketplace APIs); buildable remainder ordered (a)–(e); multi-provider correctness hazards recorded (R1 global epoch, dist.js cross-source dots, advisor prompt shop line). Objective: (a) multi-source ingestion + per-(source, target) onboarding epoch |
| 2026-10-13 | 2 | Derive: (a) marked done (iteration 1); position updated — remaining single-shop assumptions (advisor prompt, dist.js dots) listed as live with provider #2; I1 (global `Since`/`SoldStatsReady`) added as hazard (4), sequenced after provider #2 (same-day onboarding) and before any later-onboarded source/hosting; provider probe (robots + top pages, identified UA): ALLU allowed + served, RAGTAG robots timed out; remainder re-ordered (b+c)…(f). Objective: provider #2 = ALLU end to end (vetting record + adapter + own Scheme + fixtures + wiring) with dist.js `{source, code}` + advisor prompt shop-list fixes |
| 2026-10-13 | 2→3 | **Human:** display bar → "Accept as demo" (Open Questions); scope narrowed — user: "only work on providers": remaining iterations target the ≥3-providers clause only (adapters, schemes, per-provider correctness incl. what onboarding a provider requires); non-provider buildables (two-tier scheduler, etc.) → not this sprint |
| 2026-10-13 | 3 | Derive: position re-read under the restated display bar — BRAND OFF counts; GALLERY RARE (R1/R3/R4 pass, R2 failed on silence only) = provider #2 lead; ALLU (ALLU公式 only) = provider #3; `docs/provider-vetting.md` Outcome stale (predates decision). Remainder narrowed to providers: (b) GALLERY RARE → (c) ALLU → (d) I1. Deferred += realtime two-tier cadence (scope redirect). Second preparatory-only iteration flagged as the thing to avoid. Objective: onboard GALLERY RARE end to end (vetting re-read, adapter + own Scheme + fixtures + tests + wiring + scratch crawl, dist.js `{source, code}` key, advisor prompt shop list, README source text) |
| 2026-10-13 | 3 | **Human:** "wrap it up after this iter, will continue in a new session" → halted at iteration 3 boundary (Stop signal: Human). Sprint stays In Progress, worktree + branch kept for resume (`/sprint-projex.md @2610131822-kura-def-coverage-sprint.md`); next = derive iteration 4. Counters: consecutive failures 0, consecutive nothing 0 |
| 2026-10-13 | 4 | Derive (resume): head `63e03ba`, vet + tests green; iteration-2 hazards 1–2 (dist.js per-shop key, advisor shop line) + README source text marked done (iteration 3); providers 2 of 3; ALLU (ALLU公式) = only unblocked candidate — open: (d) comparison clause + 第18条 (12) BOT clause, both need ToS text (vetting budget already over → plan records the route); fallback if ALLU fails = ≥3 clause → Deferred (human: permission / API key / new candidate); remainder (c) ALLU → (d) I1 + never-complete-source detail re-check (grare F1 follow-up) merged. Objective: onboard ALLU (ALLU公式 only) end to end or record its Fail with ToS evidence |
| 2026-10-14 | 5 | Derive (resume): head `a3fa2d8`; vet + tests green; **providers 3 of 3 built** (ALLU, iteration 4). Owner calls F1 (consignments), F3 ((12) reading) and F7 (stop-on-403) added to Open Questions, provisional as shipped. Remainder (d) split. **I1 per-source gate → Deferred**: it is def L4 Day-1 open question ("per provider; per-model stats mixing providers onboarded on different dates"; shipped provisional = global 14 d), live on the base DB (BRAND OFF only, since 10-12). Sold observation for never-complete sources (GALLERY RARE, ALLU) is kept as buildable per-provider correctness: listings never turn gone today. (e) ALLU F2 title-derived model fallback is optional, judged next derive. Objective: never-complete sources observe sales (product-page re-check cadence within each source's detail budget + ALLU list-visible sold rows → gone) |
| 2026-10-14 | 6 | Derive (resume): head `0844630`; vet + tests green (10 packages); app builds + starts from worktree (scratch DB, `/` + `/api/models` 200, mock advisor). (d′) done (iteration 5): every provider has a sold signal into `MarkGone`. Full clause map written in Current Position, each clause B or D with a named blocker. (e) ALLU title-derived model fallback judged quality, not coverage (def's 尚無行情估價 state covers no-group listings; README discloses 25%) → dropped from remainder. N1/N2 (Low) + F5 left as finalize residuals; owner-held F1/F3/F7 + GALLERY RARE paging unchanged. **Outcome: goal reached** |
| 2026-10-14 | — | **Finalize:** goal reached accepted; sprint-scope audit skipped (per-iteration audits suffice); no external nav bound; Status → Complete (Goal Reached), nav → `.projex/closed/`; sold re-check open question ticked; owner questions (GALLERY RARE paging, ALLU F1/F3, F7) carried open. Branch squash-closed into `master` (human-confirmed) |
