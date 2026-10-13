# Patch: Kura 行情怎麼算 copy accuracy (audit F1–F4, F7, F8)

> **Status:** Complete
> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 7
> **Author:** Claude Opus 5.5 (patch, sprint subagent)
> **Parent:** 2610131134-kura-methodology-footer-audit.md
> **Directive:** Fix audit F1–F4 (four `/methodology` sentences false of the code in the general data state), F7 (`since` guard rationale), F8 (`Schemes()` order test). F5, F6, anomaly boundary, F19 left to the owner.
> **Source Plan:** Direct (findings of 2610131134-kura-methodology-footer-audit.md)
> **Related Projex:** 2610131134-kura-methodology-footer-audit.md (findings source; conditions ticked) | 2610131121-kura-methodology-footer-patch.md (introduced the copy) | 2610131104-kura-methodology-footer-plan.md (authored the copy) | 2610130524-kura-interactive-mock-sprint.md
> **Result:** Success
> **Commits:** `ef3d531` (code + test)

---

## Summary

Four `/methodology` sentences held only because the DB had seen one sweep (0 gone, 1 price point per listing) or misattributed text. Each is rewritten to what its code path does after a crawl marks listings gone or sees price changes; verified against the code and against a scratch DB copy carrying one gone listing and one price cut. Added the `since` guard rationale and a `Schemes()` ordering test.

---

## Changes

### `/methodology` copy (F1–F4) + `since` rationale (F7)

**File:** `internal/web/static/js/pages.js` — `renderMethodology`
**Change Type:** Modified

| # | Section | Before | After | Code path checked |
|---|---|---|---|---|
| F4 | `#source` (ln 29) | 商品、售價、照片、成色等級與等級說明，都照店家刊登的內容呈現。 | 商品、售價、照片與成色等級，都照店家刊登的內容呈現；等級說明的中文是 Kura 的翻譯，「」內是店家的原文。 | `listing.Grade.DescZH` = "Traditional Chinese translation of DescJA"; `brandoff.Scheme` `DescJA` quoted from the shop FAQ; ladder renders `desc_zh` + 「`desc_ja`」 |
| F1 | `#fair` (ln 67) | …畫的是同型號在架商品依店家等級分組的售價；中位數、表格與顧問引用的是同一組數字。 | …畫的是同型號在售與已售出商品依店家等級分組的售價，已售出的商品標在 Kura 最後記錄到的售價。中位數與表格只算在售商品，和顧問引用的是同一組數字。 | `Engine.Model` returns every group item (active first, no window); `dist.js` `distHTML` plots each, `k-dot--sold` for non-active; medians + table (在售/中位數/最低/最高) read `by_grade`, built in `Engine.stats` from active items only; advisor `model_price_stats` reads `ByGrade`; gone item's `price_jpy` = last `RecordSeen` price |
| F3 | `#history` (ln 82) | 「售價變化」是 Kura 每次查看店家時記下的售價；店家調價，就會畫在這條線上。 | 「售價變化」從 Kura 第一次看到這件商品時的售價畫起；之後查看店家時發現調價，才記下新的售價，價格沒變的查看不另外記錄。線畫到今天；已找到新主人的商品，畫到 Kura 把它標為已找到新主人的那天。 | `store.RecordSeen` inserts `price_points` only when `isNew || prev != l.PriceJPY`; `detail.js` `historyHTML` ends at `Date.now()` / 「今天」, or `gone_at` when sold |
| F2 | `#demo` 追蹤清單 (ln 89) | …是 Kura 從目前在架的商品裡挑出來的； | …是 Kura 從店家的商品裡挑出來的，可能包含已找到新主人的商品； | `Server.member` listing candidates = `Radar.Drops` + `Radar.Gone` + per-brand top deal; models = model whose `NewestActive` is a deal |
| F7 | `renderMethodology` `since` | no comment | 4-line comment: Go encodes unset epoch as `0001-01-01T00:00:00Z`; rejected alternative = truthiness check on the string → would print 「自 0001/01/01 起記錄」; `Date.parse` → negative / NaN, `> 0` rejects both | `Overview.Since` = `e.epoch` (zero on empty DB) |

**Why:** the page is the def's disclosure channel; each sentence must hold in the general data state, not only today's single-sweep DB.

---

### `Schemes()` order test (F8)

**File:** `internal/listing/condition_test.go`
**Change Type:** Created
**What Changed:** `TestSchemesSortedBySource` registers six schemes in reverse order, filters `Schemes()` to them, asserts ascending source order. Six entries so a missing sort fails on nearly every run rather than passing by map-iteration luck (comment states this).

**Why:** `internal/listing` had no tests; `Schemes()` iterates a map, so the sort alone keeps the ladder order stable.

---

## Verification

**Method:**
1. `go vet ./...`; `go test -count=1 ./...`.
2. Mutation check for F8: comparator temporarily replaced with `return 0` (file restored from copy, `git status` clean of it) → test run 5×.
3. Scratch DB: worktree `kura.db` (+wal/shm) copied to the session scratchpad, checkpointed, then one simulated second sweep at T2 applied by SQL mirroring `RecordSeen`/`SweepMissing`: all active rows `last_seen = T2` (no new price point), `brandoff:2101219505503` ランクA ¥170000 → ¥155000 + one `price_points` row, `brandoff:2101219436272` ランクA ¥190000 → `status='gone'`, `gone_at = T2`. Served by a scratchpad build on 127.0.0.1:18741 (no crawl flags; log has 0 crawl lines; mock advisor), APIs + built-in browser read, server stopped by PID, scratch DB/binary/log deleted. Worktree `kura.db*` untouched (mtimes unchanged).

**Result:**
```
go vet ./...            clean
go test -count=1 ./...  ok advisor | listing | radar | source/brandoff | store | web
F8 mutation (5 runs):   FAIL ×5, e.g. Schemes() order = [sch-f sch-e sch-d sch-c sch-b sch-a]

/api/overview           active 15366, sold 1, drops 1
F1 /api/models/LOUIS VUITTON/M40008
   items: 9 active + 436272:gone:ランクA:190000
   by_grade: ランクA n=3 med 160000 max 160000 | ランクB n=6 (gone item excluded)
   detail page DOM: .k-dot--sold "ランクA|已售出|38349"; table ランクA 在售 3, 最高 NT$32,294
F2 /api/member watchlist: [model:LOUIS VUITTON/M51146, model:CHANEL/A67085,
   brandoff:2101219505503, brandoff:2101219436272]  ← gone listing tracked
   events: gone:…436272 | drop:…505503 | new | new
F3 505503 history: [00:41 ¥170000, 04:44 ¥155000]; end label 今天
   909455 re-seen at T2, history: [00:40 ¥160000]  ← unchanged sighting not recorded
   436272 (gone) end label 10/13 = gone_at, not 今天
   (also pinned by store TestPriceHistoryAndGoneLifecycle: day1 100, day2 90, day3 90 → 2 points)
F4 /api/overview schemes ランクA: desc_ja 多少のキズ・汚れ・若干の使用感等あり (shop FAQ), desc_zh 有少許刮痕… (Kura)
/methodology (browser): all four new sentences rendered as above; 5 「」 originals; title 行情怎麼算 · Kura 蔵;
   375 px: scrollWidth 375 = clientWidth
```

**Status:** PASS

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|----------|-------------|-------------|
| 2610131134-kura-methodology-footer-audit.md | Findings source | Conditions F1–F4 ticked with this patch; patch added to Related Projex |
| 2610130524-kura-interactive-mock-sprint.md | Sprint | none (orchestrator owns nav) |

---

## Notes

- Not touched (owner decisions per directive): F5 「尚無行情估價」 meaning, F6 示範模式 tooltip, anomaly boundary display, F19.
- Adjacent, left alone: the distribution table's column header 「店家定義」 (`dist.js`) shows `desc_zh`, i.e. Kura's translation of the shop definition — same attribution question as F4, but it is canvas copy; `#fair` 「頁面上的等級一律是店家自己的等級與說明」 now reads consistently with the F4 translation disclosure (definition is the shop's, in Kura's translation).
- F1 wording uses 已售出 to match the chart legend; the rest of the page uses 已找到新主人 for the same state.
