# Patch: Kura ALLU — Audit Fixes (F1/F2/F3 doc parts, F5, F6, F8, F9, F10)

> **Sprint:** 2610131822-kura-def-coverage-sprint.md — iteration 4
> **Status:** Complete
> **Author:** Claude Opus 5.5 ([patch], sprint iteration 4)
> **Parent:** 2610132339-kura-allu-provider-3-audit.md
> **Directive:** Address findings of 2610132339-kura-allu-provider-3-audit.md (`[patch]` step, uid `2610131822-kura-def-coverage#4`): F1 doc part only (vetting R3 + README state what the adapter keeps re consigned items; no keep/exclude change), F2 doc part only (README states the no-group-key gap; title-derived fallback out of scope), F3 doc part only (add 第18条 (15)/(18) to the ALLU re-vet from recorded ToS text; verdict unchanged), F5, F6, F8, F9, F10. Out of scope: F2 code fallback, F4, F7, any consignment-behaviour or verdict change. No requests to any ALLU host.
> **Source Plan:** Direct (audit findings; plan 2610132307-kura-allu-provider-3-plan.md already closed)
> **Related Projex:** 2610132339-kura-allu-provider-3-audit.md | 2610132325-kura-allu-provider-3-patch.md | 2610132307-kura-allu-provider-3-plan.md | 2610131822-kura-def-coverage-sprint.md (not edited)
> **Result:** Success

---

## Summary

Docs now say what the ALLU adapter keeps: ALLU-listed partner consignments pass the ALLU公式 seller check and are stored as ALLU公式 stock (vetting R3 rewritten with a correction note, Outcome line, README § Choosing a source + Limitations, `ParseList` comment). README states the 25% ungrouped share (no 行情). Vetting re-vet block gains 第18条 (15) and (18), quoted from the recorded ToS copy. Code: devalue `"null"` keys read as strings (F5); `ParseList` errors when no row of a non-empty page passes the seller check (F6); `Source.Detail` and `Sweep` end paths tested, dropped-row counter mirrors every filter (F8); top-category re-check rationale (F9); LV 49,388 named as the largest brand (F10). Verdicts, consignment behaviour, crawl invariants and `listing.Scheme`s unchanged. No network requests.

| Finding | Disposition | Commit |
|---------|-------------|--------|
| F1 doc part | Fixed — R3, Outcome, README ×2, `ParseList` comment | `ecadf0b`, `5a8a71f` (comment) |
| F1 keep vs exclude | Out of scope — owner decision, recorded open in R3 | — |
| F2 doc part | Fixed — README § Turning messy listings + Limitations | `ecadf0b` |
| F2 code fallback | Out of scope | — |
| F3 doc part | Fixed — (15)/(18) in R1 re-vet; Pass unchanged | `ecadf0b` |
| F3 (12) reading | Out of scope — owner decision | — |
| F5 `"null"` keys | Fixed + test | `5a8a71f` |
| F6 seller drift | Fixed + test | `5a8a71f` |
| F8 test gaps | Fixed (tests only) | `5a8a71f` |
| F9 rule-6 gap | Fixed (comment) | `5a8a71f` |
| F10 largest brand | Fixed — `maxListPages` comment + R5 | `5a8a71f`, `ecadf0b` |
| F4, F7 | Out of scope (plan / owner scope call) | — |

---

## Changes

### F5 — devalue null-prototype keys

**File:** `internal/source/allu/payload.go`
**Change Type:** Modified
**What Changed:**
- `hydrator.typed` (L129–144): `"Map"` and `"null"` split. `"null"` (L135) reads `rest[k]` as a literal string key, value via `h.ref`; comment states why (keys are strings, unlike Map's indices; resolving them would collapse every pair under `<nil>`).

**File:** `internal/source/allu/payload_test.go`
**Change Type:** Modified
**What Changed:**
- `TestHydrate` (L57–63): `["null","a",3,"b",1]` → `{"a": 2.5, "b": "x"}`.

**Why:** devalue serialises null-prototype objects as `["null", key, idx, …]`; latent (no `"null"` tag in fixtures — re-checked: tags present are Reactive, Set, ShallowReactive).

### F6 — seller value drift is an error

**File:** `internal/source/allu/parse.go`
**Change Type:** Modified
**What Changed:**
- `ParseList`: `sellerOK` counter (L87, L98) of rows passing `AppCustomer.Name == official && !IsAlluPartner`; after the loop (L148–157), `Rows > 0 && sellerOK == 0` → `allu: page N: none of M rows has seller ALLU公式 (seller values changed?)`. Comment names the consequence (stale brand, or every listing gone on a completing brand) and the rejected alternative (silent drop like the other filters, whose all-dropped pages are normal — a page of sold rows).
- Import `fmt` (L9).

**File:** `internal/source/allu/parse_test.go`
**Change Type:** Modified
**What Changed:**
- New `TestParseListSellerDrift` (L194): renamed seller + partner-only page → seller error; sold ALLU公式 row + renamed row → no error, 0 cards; empty last page → no error, complete.

**Why:** key drift already errored; value drift returned `Cards == 0` silently. Check scoped to the seller only: other filters legitimately drop whole pages. `Sweep` propagates the error (brand sweep stops; `RunAll` continues) — fewer requests, never more.

### F8 — test gaps

**File:** `internal/source/allu/source_test.go`
**Change Type:** Modified
**What Changed:**
- `serveBase` helper (L70–78): httptest server + `baseURL` swap with cleanup.
- `TestSweepComplete` (L80): two 1-row pages, `total_count 2` → `complete=true` after pages [1, 2], each page's card visited.
- `TestSweepEmptyPage` (L105): `total 60`, 0 rows on page 1 → `empty page` error, not complete.
- `TestDetail` (L120): `bag` 200 → detail (`M15242`), not gone | `sold` 200 → gone | 404 → gone | 403 → error | 202 → error (`Detail`'s own non-200 branch) | 500 → error (`Retries = 0`; ordered last because the fetcher still schedules its 2 s backoff).

**File:** `internal/source/allu/parse_test.go`
**Change Type:** Modified
**What Changed:**
- `TestParseList` dropped counter (L50–53) adds `!graded` (grade not in `Scheme`) and `DisplayPriceMin <= 0`, so it mirrors every `ParseList` filter.

### F9 — top-category re-check rationale

**File:** `internal/source/allu/parse.go`
**Change Type:** Modified
**What Changed:**
- Comment above the `TopCategory != "バッグ"` check (L104–108): URL scope (`c-58`) is a request query Kura cannot verify from the response; a wallet/accessory stored as a bag would join a bag's comparables. Rejected: trusting the URL alone.

### F1 — consignments: what the adapter keeps

**File:** `docs/provider-vetting.md`
**Change Type:** Modified
**What Changed:**
- ALLU R3 (L83–99): ALLUパートナー → excluded (Deferred), ユーザー → excluded; new sub-bullets: consigned items are kept — ToS 第四章 (ALLU lists partner-owned 委託アイテム itself, ALLU公式 seller; may not be in ALLU's hands at listing, then 状態確認及び真贋鑑定 after purchase; item page states it); no list-row field; product payload marks them (`is_consignment`, `item_type_key: "consigned_item"` vs `owned_inventory_item` on both fixtures), Kura does not read them → stored as ALLU公式 stock, ALLU scheme, 店家鑑定. Correction note: probe #5 item was misread as ALLU-held; its `item_type_key` is `"consigned_item"`. Open owner decision: keep with disclosure or exclude (detail-time check, not built); Pass stands until then.
- Outcome (L348–350): ALLU line notes kept consignments, keep vs exclude open (R3).

**File:** `README.md`
**Change Type:** Modified
**What Changed:**
- § Choosing a source (L73–76): ALLU公式 includes ALLU-listed partner consignments (第四章, may be inspected only after purchase), marked only on the product page; Kura currently keeps them as ALLU公式 stock.
- § Limitations, Three sources bullet (L267–268): consigned items kept; telling them apart needs a product-page check.

**File:** `internal/source/allu/parse.go`
**Change Type:** Modified
**What Changed:**
- Seller-check comment (L89–94): drops "partner consignments may be inspected only after purchase" as a reason for the filter (implied exclusion); states consigned items pass this check and only the product payload marks them, unread by `ParseDetail`.

**Why:** audit F1 — the record said excluded; the code never could. Evidence re-checked offline: recorded ToS copy 第四章 text; probe payload (`market:asset:2118219`) contains `"consigned_item"`; fixtures decoded → `is_consignment: false`, `item_type_key: owned_inventory_item` (both).

### F2 — ungrouped share

**File:** `README.md`
**Change Type:** Modified
**What Changed:**
- § Turning messy listings (L103–109): code-shaped title token stands in when 型番 is empty; 42 of 167 ALLU rows (25%) have no model name, 型番 or code token (CHANEL 20/27, GOYARD 11/29, CELINE 9/27, LOEWE 2/28; HERMES, LV 0) vs 0.7% of BRAND OFF rows → no group, so no comparables, fair value or model page.
- § Limitations, Three sources bullet (L265–266): about a quarter of ALLU rows unvalued; title-derived model name would close part of it.

**Why:** audit F2. Counts re-derived read-only from the executor's scratch DB (`group_key = ''`: allu 42/167, brandoff 101/15,367). The patch-doc claim "those group by model name" does not appear in any shipped doc (grep of `*.md` outside `.projex/`: no hit), so nothing else to correct.

### F3 — 第18条 (15), (18) in the re-vet

**File:** `docs/provider-vetting.md`
**Change Type:** Modified
**What Changed:**
- ALLU R1 re-vet (L70–76): (15) quoted in full (example: buying items to resell; Kura buys nothing, non-commercial demo → not reached, same reading as site policy 営利目的の利用); (18) quoted (catch-all at ALLU's judgement, names no act, so no reading settles it in advance; an ALLU objection would end the Pass).

**Why:** audit F3 doc part. Text from the recorded ToS copy (re-vet round, session scratchpad), no new request. R1 verdict and the (12) reading untouched — owner decision.

### F10 — largest brand

**File:** `internal/source/allu/source.go`
**Change Type:** Modified
**What Changed:**
- `maxListPages` comment (L22–24): LV ~49,000 ALLU公式 bag rows (>1,600 pages) and HERMES ~22,000 (>700), sold rows included.

**File:** `docs/provider-vetting.md`
**Change Type:** Modified
**What Changed:**
- R5 (L101–102): 49,388 LV (largest; list fixture page 1) | 22,135 HERMES (probe #4).

**Why:** `testdata/list_lv.html` payload `total_count` 49,388 (decoded offline; key `market:searchItems:/jp/ja/market/b-21/c-58/…?seller=1&page=1`).

---

## Verification

**Method:** `gofmt -l` (clean), `go vet ./...`, `go test -count=1 ./...`, `-v` on `internal/source/allu` for the new tests. Doc facts re-derived offline from fixtures, recorded ToS/probe copies and the scratch DB (read-only). No request to any ALLU host.

**Result:**
```
go vet ./...            → no output
go test -count=1 ./...
ok  	github.com/No3371/kura/cmd/kura
ok  	github.com/No3371/kura/internal/advisor
ok  	github.com/No3371/kura/internal/crawl
?   	github.com/No3371/kura/internal/fx	[no test files]
ok  	github.com/No3371/kura/internal/listing
ok  	github.com/No3371/kura/internal/radar
ok  	github.com/No3371/kura/internal/source/allu
ok  	github.com/No3371/kura/internal/source/brandoff
ok  	github.com/No3371/kura/internal/source/grare
ok  	github.com/No3371/kura/internal/store
ok  	github.com/No3371/kura/internal/web
allu -v: 19 tests PASS incl. TestParseListSellerDrift, TestSweepComplete, TestSweepEmptyPage, TestDetail
```

**Status:** PASS

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|----------|-------------|-------------|
| 2610132339-kura-allu-provider-3-audit.md | Parent (findings source) | Related Projex link; Conditions split — doc parts [PATCHED], owner decisions left open; Partial Execution line |
| 2610131822-kura-def-coverage-sprint.md | Sprint | Not edited (orchestrator absorbs) |

---

## Notes

- Owner decisions still open, surfaced unchanged: F1 keep vs exclude consignments (exclusion is plan-sized: detail-time path in `crawl`/`store`; `MarkGone` would miscount them as sold) | F3 confirm or overrule the 第18条 (12) reading | F7 stop-on-403 scope.
- `new-projex.sh --type patch --projex-dir .projex/closed` scaffolded into `.projex/closed/closed/` (the script already places patches under `closed/`); file moved to `.projex/closed/` before commit, stray dir removed.
- F6 check counts the seller check only; `sale_status`/stock value drift still drops silently, because a fully sold page is a legitimate state.
