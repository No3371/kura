# Patch: Kura Provider #2 — GALLERY RARE End to End

> **Sprint:** 2610131822-kura-def-coverage-sprint.md — iteration 3
> **Status:** Complete
> **Author:** Claude Opus 5.5 (patch!, sprint iteration 3)
> **Parent:** 2610132140-kura-grare-provider-2-plan.md
> **Directive:** Execute plan 2610132140-kura-grare-provider-2-plan.md (plan-execution mode; `patch!` step of strategy `plan → patch! → audit → [patch]`, uid `2610131822-kura-def-coverage#3`)
> **Source Plan:** 2610132140-kura-grare-provider-2-plan.md (Steps 1–7, all)
> **Related Projex:** 2610131822-kura-def-coverage-sprint.md | 2610130449-kura-mock-def.md (§ Boundaries Is: display decision "Accept as demo") | 2610131935-kura-provider-2-vetting-audit.md (F1 README → resolved here)
> **Result:** Success

---

## Summary

GALLERY RARE (`www.g-rare.com`) onboarded as Kura's 2nd provider. Vetting record re-read under the owner's display bar (BRAND OFF + GALLERY RARE Pass; ALLU = provider #3 candidate) → `internal/source/grare` (9-rank scheme, list/product parsers on 2 saved pages, first-page sweeper) → `-sources` flag + setup registers both shops → `/api/models` returns `schemes` (one per shop) and the distribution draws one ladder per shop → advisor prompt names registered shops → README. Scratch-DB crawl landed 200 GALLERY RARE listings (5 brands × 40), grades only from the 9-rank scheme, no `?` request URL, no 403/429. Browser-checked model + detail pages light/dark × desktop/375 px.

Code commit: `249fe3c1e15ffa80c4843dedabc4b64eabc80955` — `feat(source): add GALLERY RARE as second provider with per-shop grade ladders` (trailer `Projex: 2610132204-kura-grare-provider-2`).

---

## Changes

### Vetting record (Step 1)

**File:** `docs/provider-vetting.md`
**Change Type:** Modified
**What Changed:**
- § Rubric R2 restated: (a)–(c) pass with attribution + link-out incl. 無断使用/無断転載/第13条-class or silent terms (owner decision 2026-10-13, not a legal judgment); (d) still needs no clause against comparison; "As first written…" line kept.
- ALLU heading → `Fail (display) superseded; re-vet pending (provider #3 candidate)`; R2 re-read bullet ((d) not established; R1 第18条 (12) + R3 stand); old R2 marked "as first judged". No ALLU request.
- GALLERY RARE heading → `Pass (display re-read 2026-10-13)`; R2 re-read, R5 (12,617 / HERMES 5,391), crawl-limit line; old verdict kept under "First judged"; new sub-heading "Request log (adapter structure probe, 2026-10-13)" with the plan's 6-row table + purpose/#4/budget note.
- BRAND OFF heading → `Pass (display re-read 2026-10-13)`; R2 re-read ((a)–(c) pass, 第13条 2項 accepted class; (d) pass; provider 1 of 3); original FAIL paragraph kept as "As first judged".
- New `## Outcome (2026-10-13, after owner decision)`; old heading renamed `… before owner decision) — superseded`, text kept.

**Why:** Plan Step 1 — re-read, not re-fetch.

### GALLERY RARE adapter (Steps 2–3)

**Files:** `internal/source/grare/condition.go`, `parse.go`, `source.go`, `parse_test.go`, `testdata/list_lv.html`, `testdata/detail_bag.html` (all Created)
**What Changed:**
- `condition.go:9` `Scheme` — 9 grades, codes = list-page labels, DescJA verbatim from plan table, DescZH; `中古S → RankSA` with rule-6 rationale comment (`condition.go:22`); registered in `init` (`:52`).
- `parse.go:45` `ParseList` — cards `article.fs-c-productListItem`; URL query/fragment dropped; SourceID = last path segment; price/grade/NEW badge per plan; total read from `.fs-c-productList .fs-c-listControl__status__total`, missing → error. `parse.go:118` `applyTitle` per plan rules (+2 refinements, see Notes). `parse.go:172` `brandModel`. `parse.go:183` `ParseDetail` (`table.detailSpec` rows; 型番/色/素材/サイズ/付属品/状態).
- `source.go:17` `bagPaths` (5 brands, no GOYARD); `:37` `Targets`; `:56` `Sweep` (page 1 only; doc comment states hazard, rejected `?page=` + sitemap, consequences); `:84` `Detail` (404 → gone).
- `parse_test.go`: `TestParseList` (40 cards of 3729, incomplete; per-card checks; exact first card `11545330`; slug-id card `m46203` present), `TestParseListNoCount`, `TestApplyTitle` (5 plan rows + `カーフレザー/…` + 2 `CHANEL 22` rows + gender), `TestParseDetail` (exact `listing.Detail`), `TestListPageComplete`, `TestTargets` (in package, as plan allows).

**File:** `cmd/kura/main.go`, `cmd/kura/main_test.go` (Modified)
- `config.sources`; `-sources` flag default `brandoff,grare` (`main.go:57`); `setup` builds from `available` map (`main.go:137`), unknown → `unknown source %q`. Tests: `brandoff + grare` pass; `grare twice` → duplicate error.

**Why:** Plan Steps 2–3.

### Per-shop ladders (Step 4)

**Files:** `internal/web/server.go`, `server_test.go`, `static/js/dist.js`, `model.js`, `detail.js`, `kura.css` (Modified)
- `server.go:237-249` `model`: `schemes` (one per source in group, sorted by source) replaces `items[0]` `scheme`; initialised `[]` so JSON is never `null` (JS default param only covers `undefined`).
- `server_test.go`: `modelBody.Schemes`; C1 assertion on `Schemes[0]`; new `TestModelSchemesPerShop` (`:226`, shops `[店X 店Y]`).
- `dist.js:44` `distHTML({stats, items, schemes, thisId})`: ladders per scheme (viewed shop first), columns + off-ladder codes per shop, stat/dots matched on `{source, code}`; ladders with nothing drawn dropped; shared y-scale; per-ladder `.k-dist__chart` with `h3.k-dist__shop` (only when >1), own `role="img"` summary; one table with `scope="colgroup"` shop rows when >1. Old single-shop bug comment removed; new rationale comment names both rejected layouts. Toggle hides/shows all charts and tips (`:145`).
- `model.js` / `detail.js`: pass `schemes` (detail: `group.schemes`; listing endpoint `scheme` still feeds `tagHTML`).
- `kura.css`: `.k-dist__shop` (+ `small`), `.k-dist__chart + .k-dist__chart`, `.k-dist__table th[scope="colgroup"]` — tokens only.

### Advisor (Step 5)

**Files:** `internal/advisor/prompt.go`, `agent_test.go` (Modified)
- Prompt line → `{{shops}}`; `buildSystemPrompt` fills unique `Shop` of `listing.Schemes()` joined `、`, none → `（尚無）`. Test registers `ax`/`ay` (店甲/店乙) and asserts both named.

### README (Step 6)

**File:** `README.md` (Modified)
- Intro → "from Japanese luxury resale stores (BRAND OFF, GALLERY RARE)"; mermaid adapter node names both; architecture row `internal/source/grare`.
- § Choosing a source: "BRAND OFF and GALLERY RARE, which serve…"; ALLU paragraph replaced by display-decision paragraph (attribution + link-out; 第13条 2項 / silent terms; owner's decision, not legal judgment; acquisition unchanged; ALLU = third-shop candidate; link to vetting doc).
- § Turning messy listings: GALLERY RARE paragraph (text mark grade, title code / 型番, page 1 only + why, shop order, no GOYARD).
- § Knowing when something sold / § Newest listings: one sentence each.
- § Running it: `-sources grare` example + flag row. § Tests: both testdata dirs. § Limitations: "Two sources" (+ thin-sample note), image bullet per plan.

---

## Verification

**Method:** go vet/test; plan Step 3 negative check; Step 7 scratch crawl + SQL + browser.

**Result:**
```
go vet ./...                      → exit 0
go test -count=1 ./...            → ok cmd/kura, advisor, crawl, listing, radar, source/brandoff,
                                    source/grare, store, web (fx: no tests)
kura crawl -db <scratch>/nosuch.db -sources nosuch
                                  → level=ERROR msg="crawl failed" err="unknown source \"nosuch\"" (exit 1; no request)
grep -n "ALLU is not used\|a Japanese luxury resale store\|One source" README.md → no hits
```

**Step 7 crawl** (`kura crawl -db <scratch>/kura-grare.db -sources grare -details 5 -v`; scratch = read-only `.backup` of the main checkout's `kura.db`, outside the repo):
```
sweep done source=grare target=HERMES        pages=1 seen=40 new=40 gone=0 complete=false
sweep done source=grare target=CHANEL        pages=1 seen=40 new=40 gone=0 complete=false
sweep done source=grare target="LOUIS VUITTON" pages=1 seen=40 new=40 gone=0 complete=false
sweep done source=grare target=CELINE        pages=1 seen=40 new=40 gone=0 complete=false
sweep done source=grare target=LOEWE         pages=1 seen=40 new=40 gone=0 complete=false
details done source=grare fetched=5 gone=0 failed=0
```
All 5 brand-bag paths answered 200 (no `bagPaths` removal). SQL:
```
grade|rank|n : 中古A|A|111  中古S|SA|40  未使用|S|31  中古AB|AB|9  新品|N|8  中古B|B|1   (all ∈ 9-rank scheme)
grare listings with '?' in url : 0
cross-shop groups (2 sources)  : HERMES name:ピコタンロックPM, name:バーキン30, CHANEL A01112, A67086, LV M47270, …
```

**Request log — g-rare.com, this patch** (KuraBot UA = `crawl.DefaultUserAgent`; ≥3 s apart; times UTC). Fixtures via `curl -A`; crawl rows reconstructed from the crawl log (sweep timestamps) and DB `detail_at` — the fetcher does not log URLs, so times are response-complete times; statuses inferred (a list ≠200 or any ≥400 other than 404 errors the sweep/detail; none did; `gone=0` → no 404).

| # | Time | URL | Status |
|---|------|-----|--------|
| F1 | 13:46:27 | `/c/fashion/louisvuitton/louisvuitton-bag` (fixture) | 200 |
| F2 | 13:47:20 | `/c/fashion/louisvuitton/louisvuitton-bag/monogram-bag/monogram-other-bag/11533177` (fixture) | 200 |
| C1 | ~13:57:31 | `/robots.txt` (fetcher, first contact) | 200 (no KuraBot / `*` group) |
| C2 | 13:57:35 | `/c/fashion/hermes/hermes-bag` | 200 |
| C3 | 13:57:39 | `/c/fashion/chanel/chanel-bag` | 200 |
| C4 | 13:57:43 | `/c/fashion/louisvuitton/louisvuitton-bag` | 200 |
| C5 | 13:57:47 | `/c/fashion/celine/celine-bag` | 200 |
| C6 | 13:57:51 | `/c/fashion/loewe/loewe-bag` | 200 |
| C7 | 13:57:54 | `/c/fashion/loewe/loewe-bag/11396673` | 200 |
| C8 | 13:57:58 | `/c/fashion/celine/celine-bag/11403685` | 200 |
| C9 | 13:58:02 | `/c/fashion/louisvuitton/louisvuitton-bag/other-bag/11526247` | 200 |
| C10 | 13:58:06 | `/c/fashion/chanel/chanel-bag/chanel-shoulder-bag/11254010` | 200 |
| C11 | 13:58:10 | `/c/fashion/hermes/hermes-bag/birkin/birkin-black/11376439` | 200 |
| R1 | ~14:00:30 | `/robots.txt` (re-crawl, new process) | 200 |
| R2 | 14:00:35 | `/c/fashion/chanel/chanel-bag` (re-crawl after parser fix, `-brands CHANEL -details 0`) | 200 |

15 GETs on g-rare.com; none with `?`. Plan budget for Step 7 (≤11) met by C1–C11; F1–F2 = the plan's 2 fixture GETs; R1–R2 = re-run after a Step-2 fix (see Notes). Other hosts: `open.er-api.com/v6/latest/JPY` (existing FX refresh, once per crawl run). BRAND OFF: no request.

**Browser** (built-in browser pane, `kura serve -db <scratch> -sources grare -addr 127.0.0.1:8093 -advisor mock`, stopped by PID 20196):
- `/model/CHANEL/A01112` (BRAND OFF 85 + GALLERY RARE 2): two ladders, captions `BRAND OFF BRAND OFF 商品ランク · 店家鑑定` / `GALLERY RARE …`; BRAND OFF dots only under ランクA (80) / ランクB (5); GALLERY RARE dots only under 中古A / 中古AB; shared y-scale; per-ladder aria-label; table 16 rows (2 shop group rows + 5 + 9); toggle hides both charts → table, and back. GALLERY RARE card shop links read 「GALLERY RARE（另開視窗）」 → `https://www.g-rare.com/c/fashion/chanel/…/11539001` `target=_blank` (href checked, not clicked — clicking would send a browser-UA request to the shop).
- `/listing/grare:11539001`: GALLERY RARE ladder first, `is-this` on 中古A, 此件 dot in GALLERY RARE plot; 鑑定卡 shows GALLERY RARE 中古A + ZH + JA verbatim + 9-rank ladder + 店家鑑定 · GALLERY RARE 商品ランク; CTA 「到 GALLERY RARE 商品頁」.
- Visual (screenshots): model page light desktop ✓; detail dark desktop ✓; detail dark 375 px ✓ (9 columns fit, labels wrap 中古/S, no horizontal scroll: scrollWidth 375); model light 375 px ✓.
- `/methodology` (行情怎麼算): "目前的資料來自 BRAND OFF、GALLERY RARE"; GALLERY RARE 商品ランク ladder listed ✓.
- Single-shop regression: `distHTML` on `/api/models/LOUIS VUITTON/N41359` (brandoff only) → 1 chart, no h3, no colgroup rows, 12 dots = 12 items ✓.
- Console errors: none.
- Not checked visually: model page dark (desktop/mobile) and detail light (desktop/mobile) were checked by DOM only (`prefers-color-scheme` emulation confirmed, body/ink tokens resolve); tooltip hover not exercised.

**Source hygiene:** new/changed comments name symbols, no projex refs / line numbers; rationale comments name rejected alternatives (`condition.go` 中古S, `source.go` `Sweep`, `dist.js` `distHTML`, `parse.go` helpers + `/`/brand-token guard).

**Status:** PASS

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|----------|-------------|-------------|
| 2610132140-kura-grare-provider-2-plan.md | Source plan | Steps 1–7 `[PATCHED]`; criteria checked; Status Complete; moved to `.projex/closed/` |
| 2610131935-kura-provider-2-vetting-audit.md | Audit (F1 README held for owner) | None to the audit doc; F1 resolved by README § Choosing a source (this patch) — for sprint absorb |
| 2610131822-kura-def-coverage-sprint.md | Sprint nav | Not edited (orchestrator absorb) |
| 2610130449-kura-mock-def.md | Def (≥3 providers) | None — 2 of 3 now; ALLU candidate for #3 |

---

## Notes

Deviations from the plan's letter (all in the owning step's files, found by fixtures or Step 7):
- **Total selector scoped** to `.fs-c-productList`: category pages carry a sub-category block with its own `fs-c-listControl__status__total` ("4件中"); the plan's bare selector would read 4 and call the LV sweep complete (→ wrongly mark items gone).
- **Product id rule** `^[0-9A-Za-z]+$` instead of "all digits": one LV card's URL ends in `m46203` (a model code, not a 商品番号); digits-only would drop it, giving 39/40 and making a ≤40-bag brand never complete. Category slugs carry hyphens, so they still fail. Plan's `len(Cards) = 40` holds.
- **applyTitle guards:** a `/` token, an English brand token or an all-caps token ends the name only once the name has a token (`カーフレザー/モノグラム スピーディ…` and `シャネル ハンドバッグ CHANEL 22 …` gave empty names); code-shaped tokens prefixed by a brand name (`CHANEL22`, `CHANEL25`) are not model codes (`brandModel`) — first crawl stored `CHANEL22` as the code of an AS3980 bag. All plan table rows still pass. Fix landed before the commit; scratch DB's 40 GALLERY RARE CHANEL rows were deleted and re-crawled (R1–R2) because `RecordSeen` does not update title-derived columns on existing rows.
- **Detail fixture** is the list fixture's **last** card (`11533177`), not the first: card order was misread from a sorted grep. Content is a regular bag product page; no second fetch made.
- Fixtures are committed raw; the repo's `core.autocrlf=true` may normalise their line endings in the index (parsing unaffected).

Known limitations (recorded, no change):
- LOEWE codes with dots (`322.30.U95`, `387.30WS35`) do not match the code pattern → stay in the model name (LOEWE 22/41 coded). HERMES names carry size/strap tokens (`エブリンTPM 111cm`).
- Sold detection assumption (product 404) not exercised: all 5 detail fetches were live items (200).
- `crawl.Fetcher` does not log request URLs; the request log above is reconstructed (see Verification). `GODEBUG=http2debug=2` was tried, but g-rare.com answered over HTTP/1.1.
- Per-shop ladder layout is provisional (designed multi-provider UI deferred); at 375 px GALLERY RARE's 9 labels wrap to two lines.

Scratch (outside repo, kept for the audit/visual pass): `<session scratchpad>/grare-patch/` — `kura-grare.db` (+ `kura.exe`, `crawl.log`, `crawl2.log`, `serve.log`, helper scripts). Original `kura.db` untouched (opened `-readonly` for `.backup`; mtime unchanged). `new-projex.sh --projex-dir .projex/closed` scaffolded into `.projex/closed/closed/` → moved.
