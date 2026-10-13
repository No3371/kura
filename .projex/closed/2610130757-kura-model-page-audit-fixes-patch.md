# Patch: Kura model page audit fixes (F1 F2 F3 H1)

> **Status:** Complete
> **Author:** Claude Opus 5.5 (patch, sprint subagent)
> **Parent:** 2610130748-kura-model-page-audit.md
> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 3
> **Directive:** Orchestrator (direct): fix audit F1 (mobile `.k-detail__top` stretch), F2 (model title gains 型號), F3 (3 rule-6 comments in `model.js`), H1 (replace global `soldKnown` with server-owned 14-day history-span gate, provisional pending def L4; test; scratch-DB verify). Leave H2/H3/H4.
> **Source Plan:** Direct (findings of 2610130748-kura-model-page-audit.md)
> **Result:** Success
> **Related Projex:** 2610130748-kura-model-page-audit.md (findings source) | 2610130737-kura-model-page-patch.md (patched code) | 2610130449-kura-mock-def.md (real-derived class, L4 Day-1 appearance — still open) | 2610130524-kura-interactive-mock-sprint.md
> **Commits:** `b5182c4` (code)

---

## Summary

- F1: mobile listing detail top stretches → photo-less listing mat full width at 375.
- F2: model page title `{model} 型號 · {brand} · Kura 蔵`; detail title unchanged → tab/history distinguish listing vs model.
- F3: rationale comments at `MODEL_ICON`, `trackKey`, 平均找到新主人 tile, each naming the rejected alternative.
- H1: radar owns the gate. `Engine` tracks crawl history span (earliest `first_seen` → latest observation of `first_seen | last_seen | gone_at`); `MinSoldHistory = 14 days` (provisional). Overview exposes `sold_stats_ready`; model page shows 已售出 + 平均找到新主人 only when it is true. Advisor (H2) can reuse `Engine.SoldStatsReady()` — one source per stat.

---

## Changes

### Radar history-span gate (H1)

**File:** `internal/radar/radar.go` — Modified
- `MinSoldHistory` const (`:31-42`): 14 days; comment names rejected "any listing gone" gate + its consequences, provenance (design's ~2-weeks note), and that it is provisional pending the undecided day-1 / per-shop / mixed-onboarding rule (def L4, not named in source per hygiene rule 1).
- `Engine.latest` field (`:156`); `build` computes it as max of `FirstSeen | LastSeen | GoneAt` over rows (`:193-209`), stored with `epoch` (`:218`).
- `Engine.SoldStatsReady()` + unlocked `soldStatsReady()` (`:222-235`): `latest − epoch ≥ MinSoldHistory`, false on empty index. Comment names rejected `now − epoch` span (counts days the crawler did not run, e.g. copied DB).
- `Overview.SoldStatsReady` → JSON `sold_stats_ready` (`:631-639`).

**Why:** audit H1 — count-based global gate flips all 5,472 models on the first sweep marking any listing gone; def classes sold-derived stats 資料不足 until history suffices. Server owns it so advisor iteration (H2) reuses one function.

**File:** `internal/radar/radar_test.go` — Modified
- `TestSoldStatsReady` (`:210`): sold row present in every case; span `14d−1h` → false, `14d` → true, `15d` → true (method + `Overview`); a `gone_at` alone extending span counts; empty engine → false.

### Model page (F2, F3, H1 client)

**File:** `internal/web/static/js/model.js` — Modified
- `MODEL_ICON` comment (`:14-17`): `aria-hidden` vs canvas `role="img" aria-label="型號照片"` — generic icon is not the model's photo; h1 names the model.
- `trackKey` comment (`:40-42`): `model:` prefix vs `{source}:{id}` keys; collision-free while no source named `model`; rejected separate store per kind.
- Title (`:67-69`): `${name} 型號 · ${s.brand} · Kura 蔵` + comment.
- `soldKnown` (`:72-76`): reads `ov.overview.sold_stats_ready === true`; comment states server owns the gate.
- `tilesHTML` (`:90-100`): 已售出 and 平均找到新主人 both gated on `known`; 平均 tile comment: label verbatim from canvas though value is median — relabel is a canvas copy decision.

### Detail mobile layout (F1)

**File:** `internal/web/static/kura.css` — Modified
- Mobile `.k-detail__top` gains `align-items: stretch` (`:386`); shared comment above it (`:383-385`); `.k-model__top` comment now points to it (`:389`).

**Why:** desktop grid's `align-items: start` leaked into the mobile flex column → photo-less mat 220 of 343 (audit F1); canvas `detailTop` mobile omits `align-items` (= stretch).

---

## Verification

**Method:** `go vet ./...` + `go test ./...`; hygiene grep; binary built to scratchpad, served a scratch copy of worktree `kura.db` (outside repo, deleted afterwards) at `127.0.0.1:8791`, built-in browser.

**Result:**
```
go vet ./...  → clean
go test ./... → ok advisor | radar | brandoff | store | web
go test ./internal/radar -run TestSoldStatsReady -v → --- PASS
grep -rnE "[0-9]{10}-[a-z0-9-]+" internal cmd → exit 1 (none)

Scratch DB phase A — HERMES 027633CC: 2 rows set gone (first_seen epoch+2d, gone_at epoch+5d); span 5.0 d
  /api/overview → sold 2, sold_stats_ready false; /api/models/HERMES/027633CC → sold 2, median_days_to_sell 3
  page tiles: 在售 3 件 | 已售出 資料不足 | ランクA 中位數 NT$736,705 · 2 件比較 | 平均找到新主人 資料不足
  title: "バーキン30 ルージュヴィフ 型號 · HERMES · Kura 蔵"
Scratch DB phase B — one active row last_seen epoch+14d+1h; span 14.04 d; server restarted
  /api/overview → sold_stats_ready true
  027633CC tiles: 已售出 2 件 | 平均找到新主人 3 天
  CHANEL AS4906 tiles: 已售出 0 件 | 平均找到新主人 資料不足 (no observed listing → still 資料不足)

F1 @375 (emulated, scratch server):
  photo-less /listing/brandoff%3A2120100030382 → .k-detail__top flex, align-items stretch, mat 343 / info 343, scrollWidth 375 (was mat 220)
  photo /listing/brandoff%3A2101220016234 → stretch, mat 343 / info 343, img present
  model /model/CHANEL/AS4906 → .k-model__top stretch, mat 343 / info 343
  @1280 photo-less listing → .k-detail__top grid, align-items start (desktop unchanged)
F2: detail title "バーキン30 ルージュヴィフ · HERMES · Kura 蔵" vs model "バーキン30 ルージュヴィフ 型號 · HERMES · Kura 蔵"
Real worktree kura.db (read-only query): span 2026-10-12 16:25 → 17:44 UTC (~1.3 h) → gate closed today.
```

**Status:** PASS

Cleanup: servers stopped, viewport reset, tab closed, scratch DB + binary deleted, worktree clean after commits.

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|----------|-------------|-------------|
| 2610130748-kura-model-page-audit.md | Findings source | F1 condition ticked; F1/F2/F3/H1 marked `[PATCHED]` → this patch; Status untouched (H2–H4 logging remains orchestrator's) |
| 2610130449-kura-mock-def.md | L4 Day-1 appearance | None — question stays open; `MinSoldHistory` is a placeholder for its answer |
| 2610130524-kura-interactive-mock-sprint.md | Sprint | None (nav not edited per directive) |

---

## Notes

- Not done (out of directive): min-sample rule for days-to-sell — once the gate opens, 平均找到新主人 still shows a value from 1–2 sales (027633CC: 3 天 from 2). Audit H1 suggested pairing a min-sample rule; belongs with the L4 answer.
- Gate is global, not per shop/per model — L4's per-provider / mixed-onboarding question unresolved; changing it only touches `Engine.soldStatsReady`.
- Advisor still returns `sold_observed` / `median_days_to_sell` ungated (H2) — `Engine.SoldStatsReady()` is ready for it.
- Overview cached per client session (`overviewP`) → a long-lived tab won't see the gate flip until reload. Pre-existing (audit edge case), Low.
- Trailer uses `{yymmddhhmm}-{patch-name}` (no `-patch` suffix), matching patch-projex.md.
