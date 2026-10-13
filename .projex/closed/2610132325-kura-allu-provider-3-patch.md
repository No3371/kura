# Patch: Kura Provider #3 — ALLU End to End

> **Sprint:** 2610131822-kura-def-coverage-sprint.md — iteration 4
> **Status:** Complete
> **Author:** Claude Opus 5.5 (patch!, sprint iteration 4)
> **Parent:** 2610132307-kura-allu-provider-3-plan.md
> **Directive:** Execute plan 2610132307-kura-allu-provider-3-plan.md (plan-execution mode; `patch!` step of strategy `plan → patch! → audit → [patch]`, uid `2610131822-kura-def-coverage#4`)
> **Source Plan:** 2610132307-kura-allu-provider-3-plan.md (Steps 1–5, all)
> **Related Projex:** 2610131822-kura-def-coverage-sprint.md | 2610130449-kura-mock-def.md (§ Is: ≥3 providers, display decision "Accept as demo") | 2610132204-kura-grare-provider-2-patch.md (adapter precedent) | 2610132219-kura-grare-provider-2-audit.md (F1 never-complete sweeps — ALLU joins it)
> **Result:** Success

---

## Summary

ALLU (`allu-official.com`, ALLU公式 stock only) onboarded as Kura's 3rd provider: vetting record → Pass with clause evidence + both plan-time request logs; new `internal/source/allu` (devalue payload decoder, list/product parsers, own 9-rank `listing.Scheme` verbatim, capped newest-first sweeper, fixtures + tests); `-sources` default `brandoff,grare,allu`; README + radar comment. Scratch-DB crawl landed 167 `source=allu` rows in 12 ALLU GETs; browser check passed. One plan item resolved by fixture: sold marker = product payload `button_type == 6`.

Commits (sprint branch `projex/sprint/2610131822-kura-def-coverage`):
- `89d56a1` docs(vetting): pass ALLU (ALLU公式 only) on a logged ToS read, record both request logs
- `471fbfb` feat(allu): add ALLU as third source (ALLU公式 stock, own 9-rank scheme, capped newest-first sweep)
- `316b3e1` docs: name ALLU in README and the radar's never-complete-sweep note

---

## Changes

### Vetting record (Step 1)

**File:** `docs/provider-vetting.md`
**Change Type:** Modified
**What Changed:**
- L50 heading → `### ALLU (allu-official.com) — Pass (re-vet 2026-10-13), ALLU公式 stock only`.
- "R2 re-read" bullet replaced by **Re-vet (2026-10-13, ToS read)** block: R1 pass (第18条 (12) key phrases + Kura's reading; (9); (13) / site policy ④), R2 (a)–(c) pass + (d) pass (search terms 比較, ランキング, 順位, 格付, 評価; (7)/① defamation), R3 ALLU公式 + `is_allu_partner == false` (第四章 partner consignments excluded; `is_consignment` reading), R4 unchanged, R5 22,135, crawl limits, "Kura's reading, not a legal judgment" + removal rule.
- First-judgement "Evidence strength" last sentence → 第18条 (12), 最終改定, 知的財産権 now corroborated by the re-vet read.
- L133 `#### Request log (re-vet ToS read, 2026-10-13)` (3 rows) + `#### Request log (adapter structure probe, 2026-10-13)` (7 rows), copied from the plan.
- L323 new `## Outcome (2026-10-13, ALLU re-vet)` (3 of 3); previous Outcome renamed `… — superseded`.

**Why:** The verdict changed (ambiguous → pass); the record carries its evidence.

---

### ALLU adapter (Steps 2–3)

**Files (Created):** `internal/source/allu/payload.go`, `condition.go`, `parse.go`, `source.go`, `payload_test.go`, `parse_test.go`, `source_test.go`, `testdata/list_lv.html`, `testdata/detail_bag.html`, `testdata/detail_sold.html`
**What Changed:**
- `payload.go`: `payload` (L17) reads `script#__NUXT_DATA__`, hydrates devalue (`hydrator` L45: memo for shared indices, busy set → cycle back edge = nil; negative specials, −6 → 0.0; `Reactive`/`ShallowReactive`/`Ref`/`ShallowRef`, `Date`, `Set`, `Map`, `null`; other tags → nil). `node[T]` (L142): one `data` key by prefix, >1 → error naming both, none → error.
- `condition.go`: `Scheme` (L10) — 9 ranks verbatim JA + ZH per plan table; J rationale comment names both rejected alternatives; registered in `init`.
- `parse.go`: `ParseList` (L73) — errors on missing `total_count`/`page`/`limit`; keeps only `ALLU公式` ∧ `!is_allu_partner`, `stock > 0 ∧ sale_status == 1`, top category `バッグ`, grade ≠ J and in scheme, `display_price_min == max > 0`; `category` (L147), `modelCode` (L161), `ParseDetail` (L197) with `soldButton` (L191).
- `source.go`: `baseURL` var (L17, test-only reason commented), `maxListPages = 10` (L27), `brandCodes` (L30), `Targets`, `listURL`, `Sweep` (L74: cap, lowered by `maxPages`; empty page before total → error), `Detail` (L118: 404 → gone, other non-200 → error).

**Why:** Pure parsing layer on saved pages + bounded sweeper, per plan.

---

### Wiring (Step 3)

**Files:** `cmd/kura/main.go` (Modified), `cmd/kura/main_test.go` (Modified)
**What Changed:**
- `main.go` L28 import; L58 `-sources` default `brandoff,grare,allu`, help `(brandoff, grare, allu)`; L141 `available` += `allu`.
- `main_test.go`: case `{"all three", …&allu.Source{}}`.

---

### Docs (Step 4)

**Files:** `README.md`, `internal/radar/radar.go` (comment only)
**What Changed:**
- README: intro three shops; mermaid `brandoff · grare · allu`; architecture row `internal/source/allu`; § Choosing a source (ALLU 第18条 (5) same class, 第18条 (12) reading, partner/user listings left out; "candidate for a third shop" removed); § Turning messy listings (Nuxt state, no title splitting, ALLU公式 in stock, J skipped); § Condition grades (9 ranks N … J); § Knowing when something sold + § Newest listings (10-page cap, never-complete, sold via 404 or the payload's sold button type); § Running it (`-sources` default ×2, "ALLU adds at most 10 pages per brand"); § Tests (`internal/source/allu/testdata`); § Limitations ("Three sources", never-complete note).
- `radar.go` `newer` comment: "A source that never completes a sweep of a brand (GALLERY RARE …; ALLU reads at most ten pages newest first; …)".

---

## Fixtures (Step 2) — request log

KuraBot UA (`curl -A` with `crawl.DefaultUserAgent`), raw bytes, no edits; ≥3 s apart.

| # | Time (UTC) | URL | Status | Bytes | Use |
|---|------------|-----|--------|-------|-----|
| 1 | 15:13:45 | `https://allu-official.com/jp/ja/market/b-21/c-58/?seller=1&page=1` | 200 | 448,626 | `list_lv.html` (exact `Sweep` URL) |
| 2 | 15:14:09 | `https://allu-official.com/jp/ja/market/items/2118097/` | 200 | 118,372 | `detail_bag.html` (first kept row) |
| 3 | 15:14:15 | `https://allu-official.com/jp/ja/market/items/2117814/` | 200 | 117,424 | `detail_sold.html` (first `stock == 0` row) |

Fixture facts: payload key `market:searchItems:/jp/ja/market/b-21/c-58/:5237:anonymous?seller=1&page=1`, `total_count` 49,388, 30 rows, all ALLU公式 / not partner, 1 sold (2117814) → 29 cards; every row has `item_number`.

**Sold marker (plan § "Sold marker"):** scalar diff of the two product payloads → `button_type` 1 (active) vs 6 (sold); also differing: `purchase_store` (null vs `ALLU online store`), `nft_status` 0 vs 2, `ec_shipping_date_text` set vs null. HTML carries no sold text (product body is client-rendered). Chosen: `button_type == 6` → `sold=true` (`TestParseDetailSold`). Rejected in code comment: "≠ 1 means sold" — one pair only; a misread active item would be marked gone. A sold page answers 200, not 404.

**Assumption checks:**
- `display_price_min` = card price: list HTML shows `¥1,360,000` for 2118097 (`display_price_min` 1360000) ✓.
- LV rows carry `item_number` ✓ (29/29 cards coded).
- Brand codes: b-4/b-8/b-21/b-81/b-88 confirmed by the list payload's brand table (`name_en`); b-422 not in that table — confirmed by the live crawl (GOYARD sweep returned 29 GOYARD bags).
- Middle categories seen: ショルダーバッグ, ハンドバッグ, トートバッグ, ウエストポーチ・ボディバッグ, **リュックサック** (not in plan list; `CategoryFromJA` maps it → backpack; added to `TestCategory`).

---

## Verification

**Method:** `go vet ./...`; `go test -count=1 ./...`; `kura crawl -sources nosuch`; scratch-DB crawl + SQL; browser pane matrix.

**Result:**
```
go vet ./...  → OK
ok  github.com/No3371/kura/cmd/kura
ok  github.com/No3371/kura/internal/advisor
ok  github.com/No3371/kura/internal/crawl
?   github.com/No3371/kura/internal/fx [no test files]
ok  github.com/No3371/kura/internal/listing
ok  github.com/No3371/kura/internal/radar
ok  github.com/No3371/kura/internal/source/allu   (15 tests: Hydrate, PayloadMissing, Node, ParseList, ParseListFilters, Category, ModelCode, ParseDetail, ParseDetailSold, ParseDetailNoNode, ListPageComplete, Targets, ListURL, SweepCap, SweepUnknownBrand)
ok  github.com/No3371/kura/internal/source/brandoff
ok  github.com/No3371/kura/internal/source/grare
ok  github.com/No3371/kura/internal/store
ok  github.com/No3371/kura/internal/web

kura crawl -db <scratch> -sources nosuch → level=ERROR msg="crawl failed" err="unknown source \"nosuch\"" (exit 1, no request)
```

### Scratch crawl (Step 5)

DB: `sqlite3 -readonly kura.db ".backup <scratch>/kura-allu.db"` (main DB read-only; it held BRAND OFF rows only, 15,367). Command: `kura crawl -db <scratch>/kura-allu.db -sources allu -max-pages 1 -details 5 -v`, run with `GODEBUG=http2debug=2` so the Go HTTP/2 transport logs each request's `:path`, `user-agent` and response `:status` (no code change; the crawler logs no per-request URL).

Request log (from the transport log; times UTC):

| # | Time | URL | Status |
|---|------|-----|--------|
| 1 | 15:21:44 | `https://allu-official.com/robots.txt` | 200 |
| 2 | 15:21:47 | `/jp/ja/market/b-4/c-58/?seller=1&page=1` | 200 |
| 3 | 15:21:51 | `/jp/ja/market/b-8/c-58/?seller=1&page=1` | 200 |
| 4 | 15:21:55 | `/jp/ja/market/b-21/c-58/?seller=1&page=1` | 200 |
| 5 | 15:21:59 | `/jp/ja/market/b-81/c-58/?seller=1&page=1` | 200 |
| 6 | 15:22:03 | `/jp/ja/market/b-88/c-58/?seller=1&page=1` | 200 |
| 7 | 15:22:07 | `/jp/ja/market/b-422/c-58/?seller=1&page=1` | 200 |
| 8 | 15:22:10 | `/jp/ja/market/items/2112033/` | 200 |
| 9 | 15:22:15 | `/jp/ja/market/items/2113474/` | 200 |
| 10 | 15:22:20 | `/jp/ja/market/items/2116176/` | 200 |
| 11 | 15:22:25 | `/jp/ja/market/items/2117678/` | 200 |
| 12 | 15:22:29 | `/jp/ja/market/items/2116870/` | 200 |

All 12 with the KuraBot UA. Not ALLU: the crawl's FX refresh, `open.er-api.com/v6/latest/JPY` 200 at 15:21:44 (`Go-http-client/2.0` UA; every `kura crawl` makes it).

| Check | Expected | Result |
|---|---|---|
| GETs to ALLU | ≤ 13 | 12 ✓ |
| 403 / 429 / challenge | none | none ✓ (no retries) |
| `/market/user/` URL | none | none ✓ |
| `sweep done source=allu` | 6, `complete=false` | 6, all `complete=false` (HERMES 27, CHANEL 27, LV 29, CELINE 27, LOEWE 28, GOYARD 29 seen) ✓ |
| details | 5 | `fetched=5 gone=0 failed=0` ✓ |
| grades | ⊂ {N,S,SA,A,AB,B,BC,C} | A 28 · AB 74 · B 60 · S 1 · SA 4 ✓ |
| bad URL count | 0 | 0 ✓ |
| cross-shop group | ≥1 | CHANEL A01112, A37586, A37590, A58600, A58601; LV M22576, M44071, M45321, M45555, M45595 (brandoff + allu) ✓ |

Model codes: LV 29/29, CHANEL 7/27; HERMES, CELINE, LOEWE, GOYARD 0 (no `item_number`, no code-shaped title token) → those group by model name.

### Browser (Step 5) — built-in browser pane, `kura serve -db <scratch> -sources allu -addr 127.0.0.1:8094 -advisor mock`

| Page | Light/dark × desktop 1280 / 375 px | Result |
|---|---|---|
| `/listing/allu:2116870` | light desktop (screenshot), dark desktop (screenshot), dark 375 (screenshot), light 375 (DOM) | 鑑定卡: ALLU · `AB：ややきれい` · ZH meaning · JA 「目立たない汚れや、小傷などが見られる商品」 · ladder N S SA A AB B BC C J · 店家鑑定 · ALLU 状態ランク ✓; shop button 「到 ALLU 商品頁」 → `https://allu-official.com/jp/ja/market/items/2116870/`, `target=_blank` ✓; photo from `image.nanboya.com` loaded (hotlink served) ✓; distribution: ALLU ladder (在售 1 under AB) + BRAND OFF ladder ✓; scrollWidth = 375 at 375 px ✓ |
| `/model/LOUIS VUITTON/M45321` | light 375 (screenshot + DOM), dark desktop (DOM) | two ladders: `ALLU 状態ランク · 店家鑑定` (在售 1 under AB only; all other ALLU columns —) and `BRAND OFF 商品ランク` (ランクA 17, ランクB 4) ✓; list mixes ALLU + BRAND OFF cards with own labels ✓; no horizontal scroll ✓ |
| `/methodology` (行情怎麼算) | light 375, dark desktop (DOM) | lists `ALLU 状態ランク`, `BRAND OFF 商品ランク`, `GALLERY RARE 商品ランク` (3 ladders), incl. `J：使用困難` ✓ |

Console errors: none. Some later screenshots timed out (pane not drawing while hidden); those cells were checked by DOM text/metrics. Advisor shop line not separately checked (mock); `/api/overview` `schemes` = allu (9), brandoff (5), grare (9). Server stopped by PID (19500).

**Status:** PASS

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|----------|-------------|-------------|
| 2610132307-kura-allu-provider-3-plan.md | Source plan | Steps 1–5 marked `[PATCHED]`, success criteria checked, Status → Complete, moved to `closed/` |
| 2610131822-kura-def-coverage-sprint.md | Sprint | None here (orchestrator absorbs the iteration) |

---

## Notes

### Deviations from plan
- **Sold marker** decided by fixture: `button_type == 6` (plan left the field open). Docs (README § Knowing when something sold) name it.
- **`TestParseDetail`**: plan asked Color/Material/Size non-empty; the fixture item's `material` is null, so the test asserts the exact `listing.Detail` with empty Material.
- **Extra tests** beyond the plan: `TestNode`, `TestParseDetailNoNode`, `TestSweepUnknownBrand`; `TestCategory` adds `リュックサック` and `""`; `TestSweepCap` adds `maxPages=50` (cap still 10).
- **Request-log method**: `GODEBUG=http2debug=2` transport log instead of `crawl.log` lines (the crawler logs no per-request URL). Plan's `2> crawl.log` kept.
- **FX GET**: `kura crawl` also fetches `open.er-api.com` once; not an ALLU request, not in the plan's 13-GET count, recorded above.

### Total ALLU requests this patch
3 fixture GETs + 12 crawl GETs = 15, all KuraBot UA, all 200; within the plan's budgets (≤3 fixtures, ≤13 crawl).

### For a human decision (carried from plan Risks, unchanged)
- Kura's reading of ToS 第18条 (12) (BOT = tool to manipulate, not access as such); owner may overrule → remove `allu` from `available`, mark Fail (acquisition).
- Consignment reading (`is_consignment: true`, `is_allu_partner: false` = ALLU-held stock).
- `button_type` semantics inferred from one active/sold pair; a different sold-state value would leave such items 在售 (same failure mode as no marker).

### Follow-up
- Never-complete ALLU sweeps (all 6 brands at 1 page here; HERMES/LV/CHANEL exceed 10 pages in production) join sprint remainder (d) with GALLERY RARE.

### Cleanup
Scratch DB, crawl/serve logs, binary and fixture staging kept in the session scratchpad outside the repo; worktree clean apart from committed files.
