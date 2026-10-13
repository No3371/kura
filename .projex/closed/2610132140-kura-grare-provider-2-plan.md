# Kura Provider #2 — GALLERY RARE End to End

> **Sprint:** 2610131822-kura-def-coverage-sprint.md — iteration 3
> **Status:** Complete
> **Author:** Claude Opus 5.5 (plan, sprint iteration 3)
> **Parent:** 2610131822-kura-def-coverage-sprint.md
> **Source:** Direct request — sprint derive, iteration 3 (uid `2610131822-kura-def-coverage#3`) + orchestrator amendment (no ALLU fall-through)
> **Related Projex:** 2610131822-kura-def-coverage-sprint.md (§ Current Position: hazards 1, 2, 4; buildable remainder (b)) | 2610130449-kura-mock-def.md (§ Boundaries Is: acquisition bar + display decision "Accept as demo") | 2610131935-kura-provider-2-vetting-audit.md (F1 README, held for owner → answered) | 2610131913-kura-provider-2-vetting-plan.md (vetting record) | 2610131831-kura-multi-source-onboarding-epoch-plan.md, 2610131852-kura-multi-source-onboarding-epoch-audit.md (multi-source seam; I1)
> Partial Execution: Steps 1–7 (all) completed via 2610132204-kura-grare-provider-2-patch.md
> **Worktree:** No — runs on the sprint worktree `kura/.projexwt/2610131822-kura-def-coverage`, branch `projex/sprint/2610131822-kura-def-coverage` (strategy `plan → patch! → audit → [patch]`); no further worktree/branch

---

## Summary

Onboard GALLERY RARE (`www.g-rare.com`, 株式会社サークラックス) as Kura's 2nd provider: vetting record re-read under the owner's display bar → `internal/source/grare` adapter (own `listing.Scheme`, 9 ranks verbatim, fixtures + parser tests) → `setup` wiring → per-shop grade ladders in the model distribution → advisor prompt names registered shops → README source text → scratch-DB crawl + browser check.

**Scope:** GALLERY RARE provider only + the correctness fixes a 2nd scheme makes live (dist.js, `model` endpoint scheme, advisor shop line, README). ALLU, I1, designed multi-provider UI = out.
**Estimated Changes:** ~16 files — 1 new Go package (4 source files + 2 fixtures + 1 test), `cmd/kura/main.go` (+test), `internal/web/server.go` (+test), 3 JS files + `kura.css`, `internal/advisor/prompt.go` (+test), `README.md`, `docs/provider-vetting.md`.

---

## Objective

### Problem / Gap / Need
Def § Is: ≥3 providers. Kura has 1 (BRAND OFF). Owner's 2026-10-13 display decision ("Accept as demo") clears GALLERY RARE's only failed check (R2, failed on silence). `docs/provider-vetting.md` still reads "no candidate passes". Every UI/advisor path that assumed one shop becomes wrong the day a 2nd scheme registers: `dist.js` matches dots by grade code across shops; `model` endpoint returns `items[0]`'s scheme only; advisor prompt names BRAND OFF as the only source; README presents one source and rejects ALLU on terms.

### Success Criteria
- [x] `go vet ./...` + `go test ./...` green, incl. new `internal/source/grare` parser tests on saved fixtures
- [x] `setup` registers `brandoff` + `grare` (default `-sources`), each with its own scheme; `checkSources` passes; `-sources grare` limits a run to GALLERY RARE
- [x] Scratch-DB crawl (`-sources grare`) lands GALLERY RARE listings with `source=grare`, grade codes from its 9-rank scheme, no request URL containing `?`
- [x] `/api/models/{brand}/{key}` returns `schemes` (one per shop in the group); model + detail distribution draws each shop's dots only under its own ladder (browser-checked, light/dark × desktop/mobile)
- [x] Advisor system prompt names BRAND OFF and GALLERY RARE (from registered schemes, test-asserted)
- [x] `docs/provider-vetting.md`: R2 restated per owner bar; BRAND OFF counts; GALLERY RARE Pass; ALLU line re-read from the existing record (no new ALLU requests); new dated Outcome
- [x] README source text follows the def display decision (resolves iteration-2 audit F1)

### Out of Scope
- ALLU (provider #3) — later iteration; no ALLU requests in this one (orchestrator amendment: no fall-through)
- I1 per-source `Since` / `SoldStatsReady` (sprint remainder (d))
- Designed multi-provider UI (canvas pass 2, sprint Deferred row) — this plan's per-shop ladder is a minimal data-correctness rendering
- Query-string pagination / sitemap enumeration of GALLERY RARE (see Constraints)
- Live crawl of BRAND OFF; any write to `kura.db` in the main checkout; pushing

---

## Context

### Current State
- `crawl.Source` (`internal/crawl/crawler.go`): `Name/Targets/Sweep/Detail`; crawler runs sources sequentially; capped sweeps (`complete=false`) never mark gone; detail 404 → `MarkGone`.
- `cmd/kura/main.go` `setup`: one hard-coded `brandoff.Source`; `checkSources` requires a registered scheme + unique names. `crawl` sweeps every source → no way to crawl one shop.
- `internal/source/brandoff`: pattern to mirror (`source.go` sweeper, `parse.go` goquery parsers, `condition.go` scheme registered in `init`, `testdata/` raw pages).
- `listing.Rank`: N, S(未使用品), SA, A, AB, B, BC, C. `GradeFor`/`SchemeFor` per source.
- `internal/web/server.go` `model`: `res["scheme"] = SchemeFor(items[0].Source)` (comment admits single-shop). `listing` endpoint returns the item's own scheme (鑑定卡) — correct, unchanged.
- `internal/web/static/js/dist.js` `distHTML({stats, items, scheme, thisId})`: columns = one ladder; `col.pts` filters items by `condition.code` only (comment lines 44–48 flags the bug). Callers: `model.js` (`scheme` from `/api/models`), `detail.js` (`scheme` from `/api/listings/{id}`, items from `/api/models`).
- `internal/advisor/prompt.go:16`: 「目前的資料來源是日本二手精品店 BRAND OFF 的線上商店」. Prompt built once per advisor (`agent.go:53`, `openai.go:68`), after source `init`s.
- Attribution + link-out (def display conditions) already generic: `card.js` shop link `it.source_name`, `detail.js` CTA 「到 {shop} 商品頁」 (`radar.Item.SourceName` = `scheme.Shop`).

### GALLERY RARE site facts (plan-time probe, 2026-10-13, KuraBot UA, ≥3 s apart)

| # | Time (UTC) | URL | Status | Finding |
|---|------------|-----|--------|---------|
| 1 | 13:33:03 | `/robots.txt` | 200 | Unchanged vs record: groups only for named bots; they disallow `/*?`, `/*?page=`, `/*?sort=`, `/p/search`; no `*` group; no Crawl-delay; `Sitemap: /s/sitemap/sitemap.xml` (not fetched) |
| 2 | 13:33:17 | `/c/fashion/bag` | 200 | FutureShop, server-rendered; 40 cards/page; total 12,617 bags (R5); pagination only `?page=N&sort=latest` |
| 3 | 13:35:07 | `/c/fashion/hermes/hermes-bag` | 200 | Brand-bag list; total 5,391; default order `sort=priority` (shop's own order, not newest) |
| 4 | 13:35:14 | `/c/fashion/louisvuitton/louisvuitton-bag/11547174` | 404 | Guessed path; product URL must be the full category path |
| 5 | 13:35:19 | `/f/condition` | 200 | 商品ランクガイド: 9 ranks + definitions (Step 2 table) |
| 6 | 13:36:07 | `/c/fashion/louisvuitton/louisvuitton-bag/monogram-bag/monogram-hand-bag/11547174` | 200 | Product page: `table.detailSpec` th/td spec rows |

Markup (list card `article.fs-c-productListItem`): link + title `h2.fs-c-productName a > span.fs-c-productName__name`; image `img.fs-c-productListItem__image__image[data-layzr]` (`src` is a loading.svg; images on `galleryrare.itembox.cloud`); price `.fs-c-productPrice--selling .fs-c-price__value` (「3,280,000」, 税込); grade = `.fs-c-productMark__mark.label_top .fs-c-productMark__label` text (seen: 新品 未使用 中古S 中古A 中古AB); other marks `label_bottom` (NEW, 送料無料, …). Total: `.fs-c-listControl__status__total` leading int. Product id = last URL path segment (= 商品番号 on the product page). Titles: `{brandJA} {category?} {model…} {colour/金具…} {code?} {BRAND EN} {extras}`, e.g. `ルイヴィトン ハンドバッグ モノグラム スピーディ・バンドリエール20 M46222 LOUIS VUITTON 2way`; HERMES titles carry no code. Brand-bag paths seen in product URLs: `hermes/hermes-bag`, `chanel/chanel-bag`, `louisvuitton/louisvuitton-bag`, `celine/celine-bag`, `loewe/loewe-bag`; GOYARD products sit directly under `/c/fashion/goyard/` (no bag sub-category). Product page spec keys: 商品番号 ブランド ランク 状態 型番 色 素材 金具 サイズ 重量 仕様 付属品 生産国 返品.

### Key Files

| File | Role | Change Summary |
|------|------|----------------|
| `docs/provider-vetting.md` | Vetting record | R2 restated; BRAND OFF/GALLERY RARE/ALLU R2 re-read; GALLERY RARE Pass + probe log; new Outcome |
| `internal/source/grare/condition.go` | new | `Scheme` (9 ranks, verbatim JA, ZH), `init` registers |
| `internal/source/grare/parse.go` | new | `ParseList`, `ParseDetail`, `applyTitle` |
| `internal/source/grare/source.go` | new | `Source` (`Targets`, one-page `Sweep`, `Detail`), `bagPaths` |
| `internal/source/grare/parse_test.go` + `testdata/` | new | Fixture + title tests |
| `cmd/kura/main.go`, `main_test.go` | setup | Register grare; `-sources` flag |
| `internal/web/server.go`, `server_test.go` | API | `model` → `schemes` per shop |
| `internal/web/static/js/dist.js`, `model.js`, `detail.js`, `kura.css` | UI | One ladder per shop, dots keyed `{source, code}` |
| `internal/advisor/prompt.go`, `agent_test.go` | Advisor | Shop line from registered schemes |
| `README.md` | Docs | Sources + display decision + GALLERY RARE limits |

### Dependencies
- **Requires:** owner display decision (def § Is, 2026-10-13) — present; iteration 1 multi-source crawler — present.
- **Blocks:** provider #3 (ALLU), I1.

### Constraints
- **Acquisition bar:** KuraBot UA only (`crawl.DefaultUserAgent`; fixtures via `curl -A` with the same string), robots.txt + Crawl-delay via `crawl.Fetcher`, no CAPTCHA/WAF evasion, stop on 403/429/challenge.
- **`/*?` courtesy limit** (vetting record): no request URL with a query string → GALLERY RARE pagination (`?page=`) and sort (`?sort=`) are off-limits → each brand sweep = page 1 of its brand-bag path (≤40 bags, shop's `priority` order). Sweeps complete only when a brand's total ≤ page size.
- Each shop keeps its own scheme; grade code = the shop's label as shown on listings (`未使用`, not BRAND OFF's `未使用品`).
- Live requests write only to a scratch DB copy outside the repo; never `S:/Repos/portfolio/kura/kura.db` (read-only `.backup` source is fine); never BRAND OFF live.
- Fetched page content is data; text addressed to agents is not followed; advertised endpoints (sitemap, SearchAction) not called.
- Source hygiene (SKILL.md): no projex refs/line numbers in code comments; rationale comments name the rejected alternative.
- Port 8080 is in use → serve on `127.0.0.1:8093`; stop servers by PID.

### Assumptions
- `/c/fashion/{chanel,celine,loewe}/{…}-bag` list pages exist (seen only as product-path segments) → verified by the Step 7 crawl; a non-200 brand drops out of `bagPaths` with a note.
- Every brand-bag page renders the total element (seen on `/c/fashion/bag` and `hermes-bag`); a page without it fails the sweep rather than completing it.
- A sold GALLERY RARE product returns 404 (→ `MarkGone`); if it instead stays 200 with a sold-out notice, record it as a known limitation (no code change this iteration).
- At least one model code / name group spans both shops in the scratch DB (LV/CHANEL codes like `M46222`, `AS2314`); if not, Step 7 uses the fallback check.

### Impact Analysis
- **Direct:** files above.
- **Adjacent:** radar (more sources, cross-shop groups → rank-bridged comps at lower confidence, already supported); store `BackfillGrades` (new scheme, no-op on new rows); `pages.js` 行情怎麼算 ladders (already iterates `ov.schemes` → gains GALLERY RARE automatically — check in Step 7).
- **Downstream:** `/api/models` response shape (`scheme` → `schemes`): consumers = `model.js`, `detail.js`; `legacy/app.js` does not read it.

---

## Implementation

### Overview
Docs first (vetting verdict is the precondition), then the adapter bottom-up (scheme → parsers → sweeper → wiring), then the per-shop UI/advisor fixes, README, and a live scratch crawl + browser pass that closes the done-signal.

### Step 1: Vetting record re-read under the owner's display bar [PATCHED] → 2610132204-kura-grare-provider-2-patch.md

**Objective:** `docs/provider-vetting.md` reflects the restated bar; GALLERY RARE Pass; no new requests except the plan-time probe already logged above.
**Confidence:** High
**Depends on:** None
**Do-Projex:** Encouraged

**Files:** `docs/provider-vetting.md`

**Changes:**
- **Rubric R2** → restate: (a) titles + facts, (b) hotlinked photo, (c) verbatim grade definitions pass **with shop attribution + link-out**, including where terms forbid reproduction without permission (無断使用 / 無断転載 / 第13条-class clauses) or are silent (default copyright) — owner decision 2026-10-13 ("Accept as demo"), recorded in the def; owner's decision for a non-commercial portfolio demo, **not a legal judgment by Kura**. (d) comparative labels still need terms with no clause against them. Keep one line: "As first written (until 2026-10-13) R2 required terms that allow the display; the records below were first judged that way."
- **BRAND OFF** heading → `— Pass (display re-read 2026-10-13)`. Add "R2 re-read (2026-10-13): (a)–(c) pass under the restated bar (第13条 2項 is the accepted clause class); (d) pass — the terms read carry no clause on comparison or ranking. Counts as provider 1 of 3." Keep the original R2 FAIL paragraph, prefixed "As first judged:".
- **GALLERY RARE** heading → `— Pass (display re-read 2026-10-13)`. R2 re-read: silent terms + "All rights reserved" = the "silent" case the owner accepted → (a)–(c) pass; (d) pass — no site-wide terms, so no clause against comparison. R5: 12,617 bags in `/c/fashion/bag`, HERMES 5,391 (probe #2/#3). Append the 6-row adapter-probe log (table in this plan's Context) under a "Request log (adapter structure probe, 2026-10-13)" sub-heading, noting: purpose = list/product structure for the adapter, not vetting; #4 was a guessed path (404); the 5-GET vetting budget is per vetting round, this is a separate round. Add "Crawl limits: `/*?` kept as courtesy → page 1 per brand-bag path; adapter `internal/source/grare`."
- **ALLU** → add "R2 re-read (2026-10-13, from this record, no new requests): (a)–(c) pass under the restated bar; (d) not established — the record does not say whether the ToS addresses comparison; re-read in the provider #3 round. R1 ambiguity (第18条 (12)) and R3 (ALLU公式 only) stand." Heading → `— Fail (display) superseded; re-vet pending (provider #3 candidate)`.
- **Outcome**: insert a new `## Outcome (2026-10-13, after owner decision)` above the old one: BRAND OFF Pass; GALLERY RARE Pass → adapter `internal/source/grare` (2 of 3); ALLU (ALLU公式 only) = provider #3 candidate, pending (d) + R3 filter; RAGTAG / Brandear / RECLO blocked (acquisition, unchanged). Rename the old heading to `## Outcome (2026-10-13, before owner decision) — superseded` and keep its text.

**Rationale:** Re-read, not re-fetch: R2 changed, the evidence did not (audit F-notes on ALLU budget).
**Verification:** Every heading verdict uses the fixed set (Pass | Fail | Blocked | Deferred) or states what supersedes it; no ALLU URL fetched; `git diff --stat` = this file only.
**If this fails:** `git checkout -- docs/provider-vetting.md`.

---

### Step 2: `internal/source/grare` — scheme, parsers, fixtures, tests [PATCHED] → 2610132204-kura-grare-provider-2-patch.md

**Objective:** Pure parsing layer for GALLERY RARE with its own scheme, tested on real saved pages.
**Confidence:** High (markup seen in probe); Medium for title splitting heuristics
**Depends on:** None
**Verify-Projex:** Encouraged

**Files:** `internal/source/grare/condition.go`, `parse.go`, `parse_test.go`, `testdata/list_lv.html`, `testdata/detail_bag.html`

**Fixtures** (2 GETs, KuraBot UA, ≥3 s apart, raw bytes, no edits): `https://www.g-rare.com/c/fashion/louisvuitton/louisvuitton-bag` → `list_lv.html`; one product URL taken from that file's first card → `detail_bag.html`. Log time/URL/status in the patch doc.

**`condition.go`** — `SourceName = "grare"` (declare in `parse.go` like brandoff). Scheme, `Grades` in published order, registered in `init`:

```go
var Scheme = listing.Scheme{
	Source: SourceName,
	Name:   "GALLERY RARE 商品ランク",
	Shop:   "GALLERY RARE",
	Graded: "店家鑑定",
	Ref:    "https://www.g-rare.com/f/condition",
	Grades: []listing.Grade{ /* table below */ },
}
```

| Code = Label | Short | Rank | DescJA (verbatim, 商品ランクガイド) | DescZH |
|---|---|---|---|---|
| 新品 | 新品 | RankNew | 傷や使用感がなく、商品の製造より2年以内の商品。 | 無傷、無使用感，且製造日起 2 年內的商品。 |
| 未使用 | 未使用 | RankUnused | 傷や使用感はほぼないが、下記のいずれか、または複数に該当する商品。①製造より2年以上経過しているか、もしくは製造年が不明。②保存や、展示の際軽微な状態変化が見られる。 | 幾乎無傷、無使用感，但符合下列一項以上的商品：①製造已逾 2 年或製造年份不明；②保存或展示時出現輕微狀態變化。 |
| 中古S | S | RankSA | わずかしか使用の形跡が見られない中古品。 | 僅有極少使用痕跡的二手品。 |
| 中古SA | SA | RankSA | ジュエリー・時計など磨き上げを施したものできれいになった商品。 | 珠寶、手錶等經拋光處理而變得漂亮的商品。 |
| 中古A | A | RankA | 傷などの使用の形跡が見られるが、コンディションの良い商品。 | 可見刮痕等使用痕跡，但狀態良好的商品。 |
| 中古AB | AB | RankAB | 日常的に使用していた形跡が見られ、目立つ傷等もあるがコンディションの良い商品。 | 可見日常使用痕跡，也有明顯刮痕等，但狀態良好的商品。 |
| 中古B | B | RankB | 日常的に使用していた形跡が見られ、目立つ傷等もある中古商品。 | 可見日常使用痕跡，也有明顯刮痕等的二手商品。 |
| 中古BC | BC | RankBC | 日常的に使用していた形跡が見られ、目立つ傷が多くみられる中古商品。 | 可見日常使用痕跡，且有許多明顯刮痕的二手商品。 |
| 中古C | C | RankC | 非常に使用感や傷があり劣化の目立つ中古商品。遠くから見ても劣化のわかる商品。 | 使用感與刮痕非常明顯、劣化顯著的二手商品，遠看也看得出劣化。 |

Rationale comment required on the 中古S → `RankSA` line (rule 6): rejected `RankUnused` (that rank stands for unused stock; 中古S is a used item by the shop's own definition) and rejected a new Rank (would change `rankFactor`, the advisor's `min_rank` enum and every scheme's bridge for one shop). Two grades sharing `RankSA` is fine: Rank is only the cross-shop bridge.

**`parse.go`** (package doc mirrors brandoff's):
- `type ListPage struct { Total int; Cards []listing.Listing }`; `func (p ListPage) Complete() bool { return len(p.Cards) >= p.Total }`.
- `ParseList(r io.Reader, base *url.URL) (ListPage, error)`: for each `article.fs-c-productListItem`: href from `.fs-c-productName a` → resolve against `base`, drop query/fragment; `SourceID` = last path segment, skip card unless all digits; `Title` = `cleanText(.fs-c-productName__name)`; `ImageURL` = resolved `img.fs-c-productListItem__image__image` `data-layzr`; `PriceJPY` = first int of `.fs-c-productPrice--selling .fs-c-price__value` (skip card if 0); grade = first `.fs-c-productMark__mark.label_top .fs-c-productMark__label` whose text is a `Scheme` code → `Grade, Rank`; badges: a `label_bottom` mark `NEW` → `"NEW"`; `Source = SourceName`; `applyTitle`. Total: `.fs-c-listControl__status__total` leading int; element missing → `error("grare: no result count on list page")`, cards or not (a changed page must not read as a complete sweep: two complete sweeps that miss items mark them gone).
- `applyTitle(l *listing.Listing, title string)`: `toks := strings.Fields(title)`; drop toks[0] if `listing.BrandByName` → `l.Brand`. Over remaining tokens: `Category` = first `listing.CategoryFromJA(tok)` hit other than `"bag"`, default `"bag"`; `ModelCode` = first token matching `^[0-9A-Z][0-9A-Z-]{4,}$` that contains a digit; `Gender` from メンズ/レディース/ユニセックス (`"/"`-joined); `セール品` → badge `"SALE"`. `ModelName` = tokens from the start (after brand) up to the first stop token, skipping category and gender tokens; stop token = contains `/` | equals `ModelCode` | `BrandByName` hit | all-ASCII-uppercase letters, len ≥ 4 (e.g. `LOUIS`) | `セール品`.
- `ParseDetail(r io.Reader) (listing.Detail, error)`: rows of `table.detailSpec tr` → `th` text → `td` lines (`<br>` → `" / "`, as brandoff); map 型番→`ModelCode`, 色→`Color`, 素材→`Material`, サイズ→`Size`, 付属品→`Accessories`, 状態→`Notes`; no rows → error `"grare: no spec table on product page"`.
- `cleanText`, `firstInt` local copies (brandoff's are unexported; a shared helper package for two one-liners = rejected).

**`parse_test.go`:**
- `TestParseList` (fixture): `len(Cards)` = 40, `Total` > 40, `!Complete()`; every card has SourceID, URL without `?`, Title, ImageURL (https, not `loading.svg`), PriceJPY > 0, Grade ∈ Scheme with matching Rank, Brand `LOUIS VUITTON`; ≥1 card with non-empty ModelCode. Plus one exact `listing.Listing` assertion for a card the executor picks from the fixture.
- `TestApplyTitle` (table, literal titles from the probe):

| Title | Brand | ModelName | ModelCode | Category |
|---|---|---|---|---|
| `ルイヴィトン ハンドバッグ モノグラム スピーディ・バンドリエール20 M46222 LOUIS VUITTON 2way` | LOUIS VUITTON | `モノグラム スピーディ・バンドリエール20` | M46222 | handbag |
| `エルメス バーキン30 ブルーニュイ/シルバー金具 トゴ Y刻印 HERMES Birkin ハンドバッグ` | HERMES | `バーキン30` | "" | handbag |
| `エルメス ハンドバッグ ケリー28 内縫い エトゥープ/シルバー金具 エバーカラー G刻印 HERMES Kelly 2wayバッグ` | HERMES | `ケリー28 内縫い` | "" | handbag |
| `シャネル チェーンショルダーバッグ マトラッセ ココマーク バニティ パテント AP2873 CHANEL 黒 セール品` | CHANEL | `チェーンショルダーバッグ マトラッセ ココマーク バニティ パテント` | AP2873 | bag (+ badge SALE) |
| `セリーヌ ハンドバッグ ベルトバッグ マイクロ レザー 189153 CELINE 2wayショルダ` | CELINE | `ベルトバッグ マイクロ レザー` | 189153 | handbag |

  (Row 4: `チェーンショルダーバッグ` is not in `listing.Categories`, so it stays in the name and Category falls back to `bag`. Row 2: `ハンドバッグ` after the stop token still sets Category.)
- `TestParseDetail` (fixture): ModelCode/Color/Material/Size non-empty, exact values asserted from the fixture.
- `TestListPageComplete`: synthetic `{Total: 3, 3 cards}` → true; `{Total: 41, 40 cards}` → false; `{0, none}` → true.

**Verification:** `go test ./internal/source/grare/ -v` green; `go vet ./...`.
**If this fails:** delete `internal/source/grare/`.

---

### Step 3: Sweeper, setup wiring, `-sources` flag [PATCHED] → 2610132204-kura-grare-provider-2-patch.md

**Objective:** GALLERY RARE crawlable through the existing crawler; any run can be limited to chosen shops.
**Confidence:** High
**Depends on:** Step 2

**Files:** `internal/source/grare/source.go`, `cmd/kura/main.go`, `cmd/kura/main_test.go`

**`source.go`:**

```go
const baseURL = "https://www.g-rare.com"

// bagPaths: each brand's bag category. Brands without one (GOYARD lists
// products directly under its brand page, mixed with wallets) are not targeted.
var bagPaths = map[string]string{
	"HERMES":        "/c/fashion/hermes/hermes-bag",
	"CHANEL":        "/c/fashion/chanel/chanel-bag",
	"LOUIS VUITTON": "/c/fashion/louisvuitton/louisvuitton-bag",
	"CELINE":        "/c/fashion/celine/celine-bag",
	"LOEWE":         "/c/fashion/loewe/loewe-bag",
}

type Source struct { Fetcher *crawl.Fetcher; Brands []string }
func (s *Source) Name() string { return SourceName }
func (s *Source) Targets() []string // s.Brands in order, kept only if in bagPaths
func (s *Source) Sweep(ctx, target, maxPages, visit) (bool, error)
func (s *Source) Detail(ctx, u) (listing.Detail, bool, error)
```

- `Sweep`: unknown target → error; GET `baseURL + bagPaths[target]`; status ≠ 200 → error `grare: list %s: status %d`; `ParseList(…, resp.URL)`; fill empty `Brand` with target; `visit(1, cards)`; return `lp.Complete()`. `maxPages` unused (always one page).
- Doc comment on `Sweep` (hazard, rule 3/6): reads page 1 only — pagination and sort exist only as query strings, which Kura leaves alone (the shop's robots.txt bars them for the bots it names); rejected: `?page=` paging (outside that limit) and sitemap enumeration (one product-page fetch per listing per crawl). Consequences stated: order is the shop's default, not newest first as `crawl.Source.Sweep` asks; a brand with > 40 bags never completes a sweep, so sweeps never mark its items gone and none count as seen appearing; sold items surface only through a product-page 404.
- `Detail`: GET; `404` → `gone=true`; else `ParseDetail`.

**`main.go`:** add `sources string` to `config`; flag (both commands) `fs.StringVar(&c.sources, "sources", "brandoff,grare", "comma-separated sources to crawl (brandoff, grare)")`; in `setup` replace the literal list:

```go
available := map[string]crawl.Source{
	brandoff.SourceName: &brandoff.Source{Fetcher: crawl.NewFetcher(c.delay), Brands: brands},
	grare.SourceName:    &grare.Source{Fetcher: crawl.NewFetcher(c.delay), Brands: brands},
}
var sources []crawl.Source
for n := range strings.SplitSeq(c.sources, ",") {
	src, ok := available[strings.TrimSpace(n)]
	if !ok {
		return nil, fmt.Errorf("unknown source %q", strings.TrimSpace(n))
	}
	sources = append(sources, src)
}
```

Import `internal/source/grare` (also registers its scheme). `checkSources` unchanged (catches `-sources grare,grare`).

**`main_test.go`:** add cases `{"brandoff + grare", []crawl.Source{&brandoff.Source{}, &grare.Source{}}, ""}`; `TestGrareTargets`: `(&grare.Source{Brands: []string{"HERMES","GOYARD","LOUIS VUITTON"}}).Targets()` = `[HERMES LOUIS VUITTON]` (put in `grare` package test instead if preferred).

**Verification:** `go vet ./... && go test ./...`; `go run ./cmd/kura crawl -sources nosuch` → `unknown source "nosuch"` (exits before any request).
**If this fails:** revert `main.go`/`main_test.go`; remove `source.go`.

---

### Step 4: Model distribution — one ladder per shop [PATCHED] → 2610132204-kura-grare-provider-2-patch.md

**Objective:** each shop's dots, medians and table rows sit only under its own grade ladder.
**Confidence:** Medium (UI layout is provisional)
**Depends on:** None (testable with any two registered schemes)
**Verify-Projex:** Encouraged

**Files:** `internal/web/server.go`, `internal/web/server_test.go`, `internal/web/static/js/dist.js`, `model.js`, `detail.js`, `internal/web/static/kura.css`

**`server.go` `model`:** replace the `items[0]` block + its comment (add the `slices` import):

```go
// One ladder per shop in the group: grades are only comparable within a shop.
var schemes []listing.Scheme
seen := map[string]bool{}
for _, it := range items {
	if seen[it.Source] {
		continue
	}
	seen[it.Source] = true
	if sc, ok := listing.SchemeFor(it.Source); ok {
		schemes = append(schemes, sc)
	}
}
slices.SortFunc(schemes, func(a, b listing.Scheme) int { return strings.Compare(a.Source, b.Source) })
res["schemes"] = schemes
```

**`server_test.go`:** `modelBody.Scheme` → `Schemes []listing.Scheme`; existing C1 assertion → `len(mb.Schemes) == 1 && mb.Schemes[0].Shop == "テスト店"`. New `TestModelSchemesPerShop` (own store): register `wx` (Shop `店X`) and `wy` (Shop `店Y`), one listing each with the same ModelCode; `/api/models/HERMES/{code}` → 2 items, `schemes` shops = [店X 店Y].

**`dist.js`:** signature `distHTML({ stats, items, schemes = [], thisId = null })`.
- Ladders: `schemes` ordered with the viewed listing's shop first (`items.find(i => i.id === thisId)?.source`), else as given. Per ladder: `cols = scheme.grades.map(grade => ({ source: scheme.source, grade }))`; append codes from that shop's items outside the ladder (existing rule, now per shop); `col.stat = stats.by_grade?.find(s => s.source === col.source && s.grade.code === col.grade.code)`; `col.pts = items.filter(it => it.source === col.source && it.condition?.code === col.grade.code)`. Delete the comment block at lines 44–48 (fixed).
- Y-scale shared across ladders (prices stay comparable). Return `""` when no ladder has points/medians.
- Markup: per ladder a `<div class="k-dist__chart">` containing, only when ladders > 1, `<h3 class="k-dist__shop">${shop} <small>${scheme.name} · ${scheme.graded}</small></h3>`, then plot (own `role="img"` summary: `${model_name} ${shop} 各等級的價格分布，在售 ${n} 件` with n = that ladder's active dots; this-listing text as now) + axis. One table: per ladder, when > 1, a `<tr><th colspan="6" scope="colgroup">${shop}</th></tr>` before its rows. Section heading/legend/toggle unchanged.
- Listeners: toggle hides/shows **all** `.k-dist__chart` in the section and hides every `.k-dist__tip` (`querySelectorAll`); hover/mouseout already scoped to the dot's own plot.

**`model.js`:** destructure `schemes`, call `distHTML({ stats: s, items, schemes })`. **`detail.js`:** `distHTML({ stats: group.model, items: group.items, schemes: group.schemes, thisId: it.id })` (the listing endpoint's `scheme` still feeds `tagHTML`).

**`kura.css`:** after `.k-dist__table` rules: `.k-dist__shop` (font-size 15px, weight 600, `var(--k-ink)`, margin 18px 0 8px) with `small` (12px, `var(--k-ink-3)`, weight 400); `.k-dist__chart + .k-dist__chart { margin-top: 20px; }`; `.k-dist__table th[scope="colgroup"]` (left, `var(--k-ink-2)`, padding-top 16px). Tokens only, no new colours.

**Rationale:** Rejected: both ladders as columns of one plot — 14 columns at 375 px leave ~20 px each, below the dot-lane minimum the file's LANES comment sets. Rejected: plot only the viewed shop — hides the other shop's prices on the model page.
**Verification:** `go test ./internal/web/`; Step 7 browser checks.
**If this fails:** revert the six files.

---

### Step 5: Advisor prompt names the registered shops [PATCHED] → 2610132204-kura-grare-provider-2-patch.md

**Objective:** prompt's source line lists every registered scheme's shop.
**Confidence:** High
**Depends on:** None
**Do-Projex:** Encouraged

**Files:** `internal/advisor/prompt.go`, `internal/advisor/agent_test.go`

```go
// Before:
目前的資料來源是日本二手精品店 BRAND OFF 的線上商店，由 Kura 定期爬取並計算行情。
// After:
目前的資料來源是日本二手精品店 {{shops}} 的線上商店，由 Kura 定期爬取並計算行情。
```

`buildSystemPrompt`: `{{shops}}` → unique `Shop` of `listing.Schemes()` in order, joined `、`; none registered → `（尚無）`. Test: in `TestPromptStatesThresholds` (or a new test) register `ax` (Shop `店甲`) + `ay` (Shop `店乙`) → prompt contains `店甲` and `店乙`; existing no-`{{` check covers the placeholder.

**Verification:** `go test ./internal/advisor/`.
**If this fails:** revert both files.

---

### Step 6: README source text [PATCHED] → 2610132204-kura-grare-provider-2-patch.md

**Objective:** README matches the def display decision, the two-shop state, and GALLERY RARE's limits (closes iteration-2 audit F1).
**Confidence:** High
**Depends on:** Steps 1–3
**Do-Projex:** Encouraged

**Files:** `README.md`

- Intro line 5: "from a Japanese luxury resale store" → "from Japanese luxury resale stores (BRAND OFF, GALLERY RARE)".
- Architecture table: add `internal/source/grare` | GALLERY RARE adapter: list/product parsers (fixture-tested), first-page sweeper.
- § Choosing a source: keep KOMEHYO/acquisition paragraph; "It uses BRAND OFF, which…" → "It uses BRAND OFF and GALLERY RARE, which serve…". Replace the ALLU paragraph: what Kura shows (titles, hotlinked photos, verbatim grade definitions) always carries the shop's name and a link to its product page; BRAND OFF's terms forbid reproduction without prior approval (利用規約 第13条 2項) and GALLERY RARE's are silent (no site-wide terms); showing them this way in a non-commercial portfolio demo is the owner's decision, not a legal judgment; Kura's acquisition rules are unchanged, and a shop whose terms forbid bots is not crawled. Records: link `docs/provider-vetting.md`. ALLU: "a candidate for a third shop".
- § Turning messy listings…: new sub-paragraph GALLERY RARE: grade read from the listing's text mark; model code from the title token, product page 型番 when fetched; Kura reads only the first page (40 bags) of each brand's bag category, because pagination exists only as query strings and Kura leaves those alone (the shop's robots.txt bars them for the bots it names); no GOYARD bag category.
- § Knowing when something sold / § Newest listings: one sentence each — GALLERY RARE sweeps of brands with > 40 bags are never complete, so sweeps never mark its items sold (a product-page 404 does), and its listings never count as seen appearing.
- § Running it: mention `-sources brandoff,grare`.
- § Tests: parsers run against saved pages in `internal/source/brandoff/testdata` and `internal/source/grare/testdata`.
- § Limitations: "One source" → "Two sources" (third: ALLU candidate; arbitrage text kept); image bullet → "Product images are hotlinked from each shop with its name and a link back; a production deployment would need the shops' permission (their terms reserve reproduction) and would cache thumbnails."

**Verification:** `grep -n "ALLU is not used\|a Japanese luxury resale store\|One source" README.md` → no hits; links resolve.
**If this fails:** `git checkout -- README.md`.

---

### Step 7: Scratch-DB crawl + browser check [PATCHED] → 2610132204-kura-grare-provider-2-patch.md

**Objective:** live evidence for the done-signal; no code changes unless a check fails (then fix in the owning step's files and re-run).
**Confidence:** Medium (live site)
**Depends on:** Steps 2–5

**Commands** (`$S` = a new scratch dir outside the repo, e.g. the session scratchpad):
1. `sqlite3 S:/Repos/portfolio/kura/kura.db ".backup '$S/kura-grare.db'"` (read-only online copy; BRAND OFF rows give cross-shop groups).
2. `go run ./cmd/kura crawl -db $S/kura-grare.db -sources grare -details 5 -v 2> $S/crawl.log` → ≤ 11 GETs (robots + 5 lists + 5 products). Check the log: 5 `sweep done source=grare` lines, `complete=false` for brands with > 40 bags, no `?` in any logged URL, no 403/429. A brand path ≠ 200 → remove it from `bagPaths`, note in patch doc + README.
3. `sqlite3 $S/kura-grare.db "SELECT grade, COUNT(*) FROM listings WHERE source='grare' GROUP BY grade"` → only codes from the Step 2 table; `SELECT model_code FROM listings WHERE model_code<>'' GROUP BY brand, model_code HAVING COUNT(DISTINCT source)=2 LIMIT 5` → a cross-shop group.
4. `go run ./cmd/kura serve -db $S/kura-grare.db -sources grare -addr 127.0.0.1:8093 -advisor mock` (no crawl flags; `-sources grare` keeps any accidental 立即更新 off BRAND OFF).
5. Browser (light/dark × desktop/375 px): `/model/{brand}/{cross-shop code}` and one GALLERY RARE `/listing/grare:{id}` in that group → two ladders, captions, GALLERY RARE dots only under 中古… columns, BRAND OFF dots only under ランク…, table grouped by shop, toggle switches both charts; 鑑定卡 on the GALLERY RARE detail shows 9-rank ladder + JA definition; card shop link reads GALLERY RARE and opens `www.g-rare.com`; 行情怎麼算 lists both ladders. No console errors. **Fallback** if step 3 finds no cross-shop group: on any model page, run `(await import('/js/dist.js')).distHTML(...)` in the console with that page's data plus a second scheme and inspect the rendered HTML.
6. Stop the server by PID; delete nothing in the repo; `$S` stays outside git.

**Verification:** each check recorded (pass/fail + evidence) in the patch doc; request log (time, URL, status) copied from `crawl.log`.
**If this fails:** acquisition failure (403/challenge/robots change) → stop, record evidence in `docs/provider-vetting.md`, revert Steps 2–3 wiring, iteration ends as Failure (no ALLU fall-through).

---

## Verification Plan

### Automated Checks
- [x] `go vet ./...`
- [x] `go test ./...` (grare parsers, main `checkSources`/targets, web `schemes`, advisor prompt)
- [x] `git status` clean apart from intended files; no `*.db` or fixtures outside `internal/source/grare/testdata`

### Manual Verification
- [x] Step 7 crawl log + SQL checks
- [x] Step 7 browser matrix (light/dark × desktop/mobile) on model + detail pages
- [x] Source hygiene pass on new/changed Go/JS comments (no projex refs, no line numbers, rationale names rejected alternatives)

### Acceptance Criteria Validation
| Criterion | How to Verify | Expected Result |
|-----------|---------------|-----------------|
| Tests green | `go vet ./... && go test ./...` | exit 0 |
| Two sources, own schemes | `main_test`; `kura crawl -sources nosuch` | pass; error before any request |
| Scratch crawl lands grare | crawl.log + SQL | `source=grare` rows, grades ∈ 9 codes, no `?` URLs |
| Per-shop ladders | browser + `TestModelSchemesPerShop` | each shop's dots only under its ladder |
| Advisor shops | advisor test | prompt has both shop names |
| Vetting doc | read diff | GALLERY RARE Pass, BRAND OFF counts, ALLU re-read cited, new Outcome |
| README | grep + read | display decision, two shops, GALLERY RARE limits |

---

## Rollback Plan

1. `git -C <work-root> revert` the patch commit(s) on the sprint branch; uncommitted work → `git -C <work-root> checkout -- <files>` + remove `internal/source/grare/`.
2. Delete the scratch dir `$S` (outside the repo).

---

## Notes

### Risks
- **Thin GALLERY RARE sample** (≤40 bags/brand/crawl, shop's order): radar comps for its listings lean on BRAND OFF via the rank bridge → lower confidence; stated in README. Mitigation path (not this plan): owner decides whether `?page=` is acceptable for KuraBot (robots names other bots only).
- **Title heuristics** mis-split some names (HERMES colour words before `/`) → name-only groups fragment; model codes (LV/CHANEL/CELINE) and product-page 型番 are unaffected.
- **Sold detection** depends on product 404 (assumption) → stale "active" GALLERY RARE listings possible; recorded.
- **Markup drift** (FutureShop theme) → parser errors surface as sweep errors, not silent data (`ParseList` errors without a result count).

### Split decision
`No split — single scope, within size budget` (7 steps but ~420 lines / ~39 KB, so the size heuristic is not tripped). One provider end to end is the iteration's unit: the derive flags a second preparatory-only iteration as the smell to avoid, and Steps 4–5 are the correctness a 2nd scheme makes live, so they land with it. Sized for one patch pass: Steps 1, 5, 6 are mechanical; Steps 2–4 carry the code.

### Open Questions
- (none blocking) Whether KuraBot may page GALLERY RARE with `?page=` — owner question for a later round; this plan keeps the courtesy limit.
