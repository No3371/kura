# Patch: Kura canvas foundation + Home (guest)

> **Status:** Complete
> **Sprint:** 2610130524-kura-interactive-mock-sprint.md — iteration 1
> **Author:** Claude Opus 5.5 (patch executor, sprint subagent)
> **Parent:** 2610130535-kura-canvas-foundation-home-plan.md
> **Directive:** Execute 2610130535-kura-canvas-foundation-home-plan.md in full (patch-projex plan-execution mode; sprint `patch!` step)
> **Source Plan:** 2610130535-kura-canvas-foundation-home-plan.md
> **Related Projex:** 2610130524-kura-interactive-mock-sprint.md | 2610130449-kura-mock-def.md | 2610130420-kura-roadmap-nav.md
> **Result:** Success
> **Commits:** `0e472c0` feat(radar) (+ legacy file moves) | `169953a` feat(web)

---

## Summary

Plan Steps 1–7 executed on sprint branch `projex/sprint/2610130524-kura-interactive-mock` (worktree). `/` now serves canvas shell (tokens, fonts, light/dark, guest header, footer w/ live FX) + Home on real BRAND OFF data (最新上架 5 | 價格最美 3, `trends: hidden`); canvas listing card in all states; `/dev/cards.html` fixture page; legacy UI at `/legacy/`. Server: SPA fallback + explicit `.js`/`.css` MIME. Data: `source_name`, `photo_pending`, `drop.from_twd/to_twd`. All S1–S11 ✓.

---

## Changes

### Server routing + MIME (Step 1)

**File:** `internal/web/server.go` — Modified
- `Handler`: `GET /` → `spaFallback(sub)` (L43)
- `init` registers `.js` → `text/javascript; charset=utf-8`, `.css` → `text/css; charset=utf-8` (L56–61)
- `spaFallback` (L63–85): existing embedded file → `FileServerFS`; `/api/*` or path with extension → 404; else `index.html`

**File:** `internal/web/server_test.go` — Created. `TestStaticRoutes`: plan's 8-row table (`/`, `/browse`, `/listing/brandoff:1`, `/js/app.js`, `/kura.css`, `/legacy/`, `/js/missing.js` 404, `/api/nope` 404).

**Files:** `internal/web/static/{index.html,app.js,style.css}` → `internal/web/static/legacy/` — Moved (`move-n-stage`, content unchanged; relative asset refs work from `/legacy/`).

**Why:** history-API client routes need server fallback without masking asset/API 404s; A7 Windows registry MIME.

### Card data fields (Step 2)

- `internal/listing/condition.go` L25: `Scheme.Shop`
- `internal/source/brandoff/condition.go` L11: `Shop: "BRAND OFF"`
- `internal/radar/radar.go`: `Drop.FromTWD/ToTWD` (L45–46, set L219 via `e.FX.TWD`); `Item.PhotoPending` (L56–57, `photoPending` L226–230), `Item.SourceName` (L59, from `scheme.Shop` L197). gofmt realigned the `Item` field block (whitespace-only lines in diff).
- `internal/radar/radar_test.go`: test scheme `Shop: "テスト店"` (L20); `TestDropsAndStats` asserts TWD 24000→20000 + `SourceName` (L154–159); new `TestPhotoPending` (L187).

**Why:** card needs shop label, drop original price in TWD from server FX (one source), real pending-photo state (26 live `NowPrinting_L.jpg` rows).

### Static UI (Steps 3–6)

| File | Content |
|---|---|
| `internal/web/static/kura.css` (213 L) | tokens (`kura-tokens.css` values + `--k-chart-mark` light `#7A6B58` / dark `#857868`; `:root,[data-theme=light]` / OS-dark `:root:not([data-theme=light])` / `[data-theme=dark]`), base, tokens-file components, header, footer, home sections, card, 1023px (3 cols) + 767px (canvas mobile) breakpoints |
| `internal/web/static/index.html` | shell: theme-param script before CSS, Google Fonts link, guest header, `<main id="main">`, footer (A4 routes, `#fx`) |
| `internal/web/static/js/util.js` | `esc num twd pct ymd api` |
| `internal/web/static/js/app.js` | route table, not-found view, memoised `/api/overview`, link interception (skips `/legacy` `/dev` `/api`), `popstate`, footer FX, per-view title |
| `internal/web/static/js/card.js` | `valuation` (±3% near), `BADGE`, `cardHTML` (stack/row; pending/sold/drop/track/shop), one document-level track listener |
| `internal/web/static/js/home.js` | empty state when `active === 0`; 最新上架 + 價格最美 (`best.total` count, section omitted at 0) |
| `internal/web/static/dev/cards.html` + `dev/cards.js` | 9 states × light/dark rows + 1 row-layout card; HANDOFF §5 samples + canvas disclaimer |

---

## Verification

**Method:** `go vet ./...` + `go test ./...`; built binary in scratchpad, `serve -db kura.db -addr 127.0.0.1:8090` (real DB) and `-db <scratchpad>/empty.db -addr 127.0.0.1:8091`; built-in browser matrix; hygiene grep.

**Result:**
```
go vet ./...            → clean
go test ./...           → ok advisor | radar | source/brandoff | store | web
grep -rnE "[0-9]{10}-[a-z0-9-]+-(plan|sprint|def|nav)|projex" internal cmd → no hits
```

| Criterion | Evidence | |
|---|---|---|
| S1 | `TestStaticRoutes` passes; live `/js/app.js` → `200 text/javascript; charset=utf-8`; `/browse` direct load → shell + not-found view; `/legacy/` → tabs 行情雷達/AI 找包顧問/資料管線 + tiles, no console errors | ✓ |
| S2 | `?theme=light` body `rgb(245,240,230)` = `#F5F0E6`; `?theme=dark` `rgb(27,24,21)` = `#1B1815`, `--k-chart-mark` `#857868`; no param + colorScheme dark emulation → dark; `?theme=light` under OS dark → light; `/dev/cards.html` light/dark sections render per own `data-theme` | ✓ |
| S3 | `document.fonts` loaded: EB Garamond, Noto Sans TC, Noto Serif TC, Shippori Mincho (LXGW loads on demand — used by sold stamp only); computed families: body ui, h2 display, model ja, price num | ✓ |
| S4 | screenshots 1280/375 match canvas guest header; search submit `ネヴァーフル` → `/browse?q=ネヴァーフル`; mobile: pill + 加入會員 hidden, icon shown | ✓ |
| S5 | footer text `價格未含國際運費與進口稅費 · 真偽與商品狀況以賣家說明為準 · 匯率 1 JPY = 0.2018 TWD（2026/10/12）` / `© 2026 Kura` (API `Rate 0.201837`, `AsOf 2026-10-12T08:02:32+08:00`); 6 links → A4 routes; empty DB shows `0.2000 TWD（預設匯率）` (fallback provider) | ✓ |
| S6 | `/dev/cards.html`: 預設·接近行情 / 低於行情 31% / 可能低於行情·低信心 / 異常低價·先確認狀況 / 高於行情 12% / 尚無行情估價 / 降價 原價NT$27,854 / 已找到新主人 最後售價 / 店家照片準備中, both themes; track toggle → `aria-pressed=true`, title 追蹤中，點一下取消, colour accent-ink (light `rgb(107,31,42)`, dark `rgb(227,161,166)`), toggles back; row card present; real pending card `brandoff:2104102531481` on Home | ✓ |
| S7 | Home card ids/prices/badges identical to `/api/listings?sort=newest&limit=5` (5) and `?deals=1&sort=deal&limit=3` (3, all `deal_pct 0.5 high`); link 看全部價格最美 1,471 件 = `total 1471` | ✓ |
| S8 | :8091 1280 light + 375 dark: `main` = 還在準備中。/店裡正在整理，新的包很快就會上架。, 0 `.k-sec` | ✓ |
| S9 | 1280: wrap padding 40px, 5 grid columns, 最新上架 padding `44px 0 64px` no rule, 價格最美 head brass top rule; 375: padding 16px, header `14px 0 0`, h2 24px, rows flex 200px carousel, footer 2 cols, `scrollWidth 375 = clientWidth 375`; no `a/button/input` <24px (both widths); Tab → logo `:focus-visible` outline `2px solid #6B1F2A` | ✓ |
| S10 | trace per plan table: price ← `price_twd`, badge ← `fair.deal_pct/confidence/anomaly`, grade ← `condition.label/desc_zh`, shop ← `source_name`, 價格最美 count ← `total`, FX ← `/api/overview fx`; no fixture on `/` (fixtures only in `/dev/`) | ✓ |
| S11 | vet/test green; hygiene grep no hits | ✓ |

**Status:** PASS

---

## Deviations from Plan

1. **`.k-card { position: relative }`** (kura.css L158–160) — not in plan. At 375 the card's visually-hidden 「（另開視窗）」 span (absolute) escaped the scrolling `.k-row` and widened the document to 880px (mobile emulation then grew `innerWidth` to match, so the plan's `scrollWidth <= innerWidth` check passed falsely). Verified fixed with `scrollWidth == clientWidth == 375`.
2. **`.k-search label { align-self: stretch }`** (L123) — input was 22px tall (<24px target); now fills the 40px pill.
3. **`.k-link:hover` uses `!important`** (L95–97), as canvas does — header/footer link rules of equal specificity masked the hover colour.
4. **Track toggle syncs every copy** of the same listing id (`card.js` listener) — a listing can sit in both shelves.
5. **Router renders into a detached element and swaps only if still current** (`app.js` `render`); `popstate` ignores hash-only changes (登入/加入會員 `href="#"`).
6. **Commit grouping:** legacy file moves (staged by `move-n-stage` before the first commit) landed in `0e472c0` with the radar fields instead of `169953a`; at `0e472c0` alone `/` has no `index.html` — transient, branch-internal. `0e472c0` trailer order is `Co-Authored-By` then `Projex` (amend blocked by repo git-safety hook); `169953a` has `Co-Authored-By` last.
7. Fixture `尚無行情估價` uses マトラッセ, `高於行情` uses ネヴァーフルMM (plan named states, not bases).

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|---|---|---|
| 2610130535-kura-canvas-foundation-home-plan.md | Source plan | Status → Complete, all criteria checked, `[PATCHED]` link; moved to `closed/` |
| 2610130524-kura-interactive-mock-sprint.md | Sprint nav | none (orchestrator owns nav) |

---

## Notes

- Orchestrator follow-ups from plan Open Questions: log A1 (大家都在找 → Deferred; blocker def L3 ranking definition + crawl history) and A2–A4 decisions in sprint nav.
- `/js/`, `/dev/` directory requests show `FileServerFS` listings (no index.html) — harmless; later hardening may want a 404.
- Built-in browser screenshots intermittently timed out; DOM/computed-style checks used as primary evidence.
