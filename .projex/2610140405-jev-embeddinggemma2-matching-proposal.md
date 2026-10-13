# Jev + EmbeddingGemma 2 for Kura: same-bag matching, semantic search, photo search

> **Status:** Draft
> **Author:** Claude Opus 5.5 (Agent)
> **Parent:** User
> **Related Projex:** 2610130420-kura-roadmap-nav.md (Phase 2 exit "no duplicate items across channels" + "cross-provider comps labelled"; Phase 5 market depth) | 2610130449-kura-mock-def.md (goal spec) | 2610131822-kura-def-coverage-sprint.md (residual: 25% ALLU rows ungrouped, title fallback) | 2610140136-kura-gallery-collector-plan.md (product-page photo URLs that photo search would embed)

---

## Summary

Kura's radar can only value a listing when it finds ≥3 comps of "the same bag", and its idea of the same bag is an exact model code or an exact Japanese model-name string. Today that leaves **35% of active listings with no valuation** and only **104 of 5,641 groups spanning >1 shop**. The recommendation is to use **EmbeddingGemma 2 (EG2)** to *find* candidate same-bag matches, use **Jev** to *decide* each candidate with calibrated confidence, and then reuse the same EG2 vectors for zh→ja semantic search in the advisor and, later, photo search (以圖找包). Recommended path: Option B (EG2 recall + Jev verify, run offline), then Option D's semantic advisor search once a hosting sidecar is settled.

---

## Problem Statement

### Current State

- **Grouping:** `listing.GroupKey` (`internal/listing/listing.go`) = upper-cased model code, else `"name:" + model name with whitespace removed`. Cross-shop matches happen only on identical codes or identical katakana strings.
- **Valuation:** `Engine.value` (`internal/radar/radar.go:315`) needs `MinComps = 3`. Name-only groups are capped at `low` confidence. A wide IQR/median (>0.4 / >0.6) lowers confidence because "the group mixes leathers, sizes, limited editions".
- **Search:** `matchAny` (`radar.go:486`) does a lowercase substring match over title|model|code|color|material. The advisor's `search_listings` tool relies on Claude translating 「通勤 A4 托特」 into Japanese keywords that literally appear in titles.
- **Measured on `kura.db` (2026-10-14, read-only):**

| Metric | Value |
|---|---|
| Active listings | 15,733 |
| Active groups | 5,641 |
| Singleton groups | 3,854 (68%) |
| Name-only groups (`name:` key → always low confidence) | 2,362 |
| Groups with >1 source | 104 |
| Active listings in groups <3 (no fair value) | 5,482 (35%) |
| Rows with no model code: brandoff / grare / allu | 7,075 of 15,367 / 61 of 200 / 131 of 166 |

Example misses: `エブリンTPM 111cm` (grare, no code) and `ピコタンロック18 PM` (allu, no code) each form their own `name:` group. Any other shop's spelling (`ピコタン ロック PM`, `Picotin Lock 18`) lands in a separate group.

### Gap / Need / Opportunity

1. **Fragmented comps:** spelling variants split the same bag into singletons, so there is no valuation and no deal signal. Users see 尚無行情估價 on a third of the catalog.
2. **Over-merged comps:** a single name such as `マトラッセ` spans leathers and sizes, which raises dispersion and lowers confidence. Finer attributes (size, leather, hardware) exist in titles but aren't used.
3. **Cross-shop dedupe:** the roadmap Phase 2 exit requires "no duplicate items across channels" + "cross-provider comps labelled". Nothing does this today.
4. **Advisor recall:** zh query → ja title is keyword luck. A Taiwanese buyer's phrasing (「水桶包」「可放 A4」) rarely matches katakana titles.
5. **Job fit:** 104 job 925vi lists a JP pre-owned **database**, an AI bag advisor and an AI price radar. Product matching/entity resolution is the core data problem of that database, so it is a stronger portfolio signal than another UI surface.

### Why Now?

- EG2 released 2026-10-06: open weights, **Apache 2.0**, 100+ languages, text+image (+video/audio) in **one 768-d space**, MRL to 512/256/128, 8,192-token context, 270M text-only config. It runs locally (Ollama / llama.cpp / vLLM / LiteRT listed) with no per-call cost.
- Jev GA 2026-09-20 (`jev-1.13.0`, alias `jev-latest`). It returns typed Choice / Score / Noul answers with calibrated probabilities + confidence, costs **$0.042/MTok input, output free**, has a 64K context, and is text-only. It is built for exactly "is A the same as B?" decisions that code branches on.
- Kura now has 3 providers and real crawl data. The fragmentation shows in the numbers above, not in theory.

---

## Proposed Change

### Overview

Split the job by what each model does well:

| Role | Model | Why this model |
|---|---|---|
| **Recall:** "which listings might be the same bag?" | EG2 (local) | Cheap and multilingual (ja/zh/en spellings, katakana vs romaji); vectors are reusable for search |
| **Precision:** "are these two the same model + size + material?" | Jev Noul / Choice | Calibrated probability + confidence lets code set a merge threshold and route uncertain pairs elsewhere |
| **Merge policy** | Go code | Deterministic: threshold, dispersion guard, provenance, reversible |

The output is a `model_cluster` id per listing. The radar groups by cluster instead of the raw `GroupKey`, and the `GroupKey` stays as the cluster seed and fallback.

Jev and EG2 jobs mapped onto Kura:

| # | Use | Model(s) | Kura surface |
|---|---|---|---|
| U1 | Same-bag clustering across spelling variants + shops | EG2 kNN → Jev Noul | radar comps, fair value, 同款行情, Phase 2 dedupe |
| U2 | Attribute split (size / leather / hardware) inside over-wide groups | Jev Choice (options per model line) | dispersion ↓, confidence ↑ |
| U3 | Comp hygiene: flag non-bags (charms, straps, pouches in bag cats), exotic/limited editions | Jev Noul | `Valuation.Anomaly`-style exclusion from comps |
| U4 | Semantic advisor search (zh query → ja listings) | EG2 query embedding | new `semantic_search` tool, or a `search_listings` fallback when keywords return 0 |
| U5 | 以圖找包 (photo → similar listings) and "visually similar" shelf | EG2 image embeddings | listing detail, advisor image input |
| U6 | Advisor pre-rank: score candidates against the user's need before Claude | Jev Score | fewer Claude tokens; Claude keeps the explanation |

### Approach Options

#### Option A: Jev-only attribute normalisation

- **Description:** No embeddings. For each listing, Jev Choice questions extract `line` (from per-brand canonical lists such as LV: Neverfull | Speedy | …), `size`, `material` and `hardware`. A deterministic `GroupKey` is then built from the extracted attributes.
- **Pros:** No model to host. Explicit and debuggable keys. Fixes U2 directly.
- **Cons:** Needs hand-curated option lists per brand × line, which is a maintenance treadmill and misses unseen lines. Every listing pays a Jev call (15.7k × ~400 tok ≈ 6.3M tok ≈ **$0.27**, which is cheap; the cost is curation). Jev's docs say CJK is handled "not equally well", so every answer depends on Jev's weakest language. Gives no search or photo benefit.
- **Effort:** Medium (curation-heavy).

#### Option B: EG2 recall + Jev verify, offline batch (recommended core)

- **Description:**
  1. `kura embed`: an EG2 text embedding per listing, with the doc prefix `title: {brand} {model_name} | text: {title} {model_code} {material} {size}`, plus symmetric similarity-task prefixes for the dedupe pass. Store MRL-256 float32 in a SQLite BLOB: 15.7k × 1 KB ≈ 16 MB, and brute-force cosine in Go in-memory (the radar index is already in-memory).
  2. Candidate pairs: same brand + same category + cosine ≥ τ_recall, top-k=10, and not already sharing a group key.
  3. Jev, one request per pair with several questions (they run in parallel and the extra ones are almost free): Noul "same model line", Noul "same size", Noul "same material/leather". State = both listings' structured fields + titles.
  4. Merge when all three noul values ≥ τ_merge **and** the merged group's IQR/median ≤ the pre-merge max (dispersion guard). Union-find → `model_cluster`. Persist every decision (pair, probabilities, model version) for audit and rollback.
- **Pros:** Recall without curation, because EG2 handles katakana/romaji/zh variants. Precision is tunable from Jev's calibrated outputs. Everything runs offline, so the hosted mock needs no live model for valuation. Reversible because clusters are a derived table and the raw `GroupKey` is untouched. The vectors are reused by Option D.
- **Cons:** Needs the EG2 runtime at batch time: a Python/Ollama sidecar, since Kura is pure-Go/no-cgo. A wrong merge pollutes fair values, which is worse than a singleton, so it needs a gold set + threshold tuning. Jev is a new proprietary dependency (startup, SDK churn) but is used only offline, so an outage only delays re-clustering.
- **Cost:** pairs ≈ 5.6k groups × ≤10 ≈ ≤50k × ~500 tok ≈ 25M tok ≈ **≤ $1.05** per full run. Incremental runs only embed and pair new listings.
- **Effort:** Medium.

#### Option C: EG2-only clustering (no Jev)

- **Description:** Same as B with a pure cosine threshold (or HDBSCAN-like clustering) and no verifier.
- **Pros:** Fully local, open and free, with no proprietary dependency.
- **Cons:** Cosine can't separate "Neverfull PM" from "Neverfull MM". Size and leather tokens are a tiny part of the vector, which is exactly the dispersion problem. A single threshold either over-merges or under-merges, and there's no per-decision confidence to show in 行情怎麼算.
- **Effort:** Low–Medium.

#### Option D: Runtime semantic layer (U4 + U5), additive to B

- **Description:** The EG2 sidecar is reachable from `kura serve`. The advisor gains `semantic_search(query, filters)`: it embeds the query with the `task: search result | query:` prefix and runs kNN over listing vectors, then applies the existing structured filters (brand / price / rank). Photo search embeds an uploaded image with no prefix against **image vectors from listing photos**.
- **Pros:** Biggest demo value: 「預算 8 萬，通勤放得下 A4」 finds bags without exact keywords, and 以圖找包 is a visible "wow" in a 104 review.
- **Cons:** A hosted mock (Phase 6) needs a model process. Text-only 270M (~191 MB quantized RAM per Google) is fine; image vectors need the 440M config. Image embedding needs every listing photo **downloaded**, which is new load on shop CDNs and a new reproduction question (see Risks). It also widens the advisor tool surface, so prompt + tests change.
- **Effort:** Medium (text), High (image incl. politeness + ToS review).

### Recommended Approach

**B first, then D-text, with D-image gated on a ToS read.**

- B attacks the measured problem (35% unvalued, 104 multi-source groups) and serves the Phase 2 exit criteria. It runs offline, so hosting is unaffected.
- Use Jev only where calibrated confidence matters (merge decisions). Don't use it for extraction across the whole catalog (Option A), because CJK quality is its documented weak spot and B's pairwise questions carry the structured fields too.
- U3 (comp hygiene) rides on B's Jev calls almost for free (extra questions in the same request).
- D-text next: the vectors already exist, so it only adds query-time embedding.
- D-image is the stretch goal. It needs the photo-fetch policy settled first.
- U6 (Jev pre-rank for the advisor) is deferred. The advisor is the Claude showcase, and pre-ranking adds a hop for marginal token savings.

---

## Impact Analysis

### Affected Areas

- `internal/store`: new tables `listing_vectors(id, model, dim, vec BLOB, embedded_at)`, `match_decisions(a, b, probs JSON, jev_model, decided_at, merged)`, `model_clusters(id, cluster)`.
- `internal/radar`: `build` groups by cluster (fallback `GroupKey`). `value` treats `cluster`-merged comps like code-matched (not `name:`) only when the cluster has Jev-verified provenance. Cross-source comps get a label (Phase 2 exit).
- `cmd/kura`: new `kura embed` / `kura match` subcommands (batch, resumable).
- New `internal/embed` (HTTP client to the sidecar: Ollama `/api/embed` or a tiny Python server) + `internal/judge` (raw HTTP to `POST https://api.typesafe.ai/v1/systemone`, since there is no Go SDK).
- `internal/advisor` (Option D): new tool, prompt update, fake-API tests.
- `internal/web` + UI (Option D-image): upload endpoint, 以圖找包 entry. 行情怎麼算 copy gains "how same bag is decided".
- README: architecture diagram + design-decision section ("why recall/verify split").

### Dependencies

- EG2 weights (HF `google/embeddinggemma-2`, Apache 2.0) + a runtime: Ollama or Python sentence-transformers sidecar.
- `TYPESAFE_API_KEY` (Jev). Pin `jev-1.13.0` for reproducible merge decisions (an alias drift would silently change clusters).
- A gold set: ~200 hand-labelled pairs (same / different / same-line-different-size) for threshold tuning and an eval number for the README.
- Depends on: R1 fix (roadmap blocker) only if new crawls run before matching. Option D-image depends on gallery collector backfill (2610140136-kura-gallery-collector-plan.md).
- Enables: roadmap Phase 2 exit (dedupe + labelled cross-provider comps), Phase 5 market depth (more groups reach `MinSoldHistory`).

### Risks

| Risk | Likelihood | Impact | Mitigation |
|---|---|---|---|
| Wrong merge pollutes fair value / deal flags | Med | High | Dispersion guard; τ_merge tuned on gold set for precision ≥ 0.95; per-merge provenance; `kura match --undo`; clusters are a derived table |
| Jev CJK accuracy weaker (its own docs say so) | Med | Med | State includes romanised/structured fields (code, size, material) not only katakana; gate on `confidence`; measure on gold set before trusting |
| Jev vendor risk (startup, early SDK churn, self-run evals) | Med | Low | Offline-only use; pinned version; decisions cached so outage = no new merges, not broken site; Option C as fallback path |
| Hosting a model process for Phase 6 | Med | Med | B needs none at serve time; D-text uses 270M text-only config; D can degrade to keyword search when sidecar is down |
| Photo download for image vectors = extra load + reproduction concern (BRAND OFF 利用規約 第13条 2項, ALLU 第18条 (5)) | High if D-image | Med | Do not do D-image until owner decides; if yes: same polite fetcher (robots, ≥3 s/host), store vectors only (discard image bytes), record in `docs/provider-vetting.md` |
| Search relevance regressions in advisor | Low | Med | Semantic search as a separate tool / zero-result fallback, not a replacement; keep keyword path |
| Cosine τ_recall too tight → misses variants | Med | Low | Recall measured on gold set; k-NN top-k rather than pure threshold |

### Breaking Changes

None to stored data: `GroupKey` and `listings` are unchanged, and clusters are additive. Visible behaviour changes on purpose: more listings get fair values, and some confidences and deal flags move. Deal counts on Home / Browse shift, so the def's provisional F1/F6 numbers should be re-read after the first match run.

---

## Open Questions

- [ ] **Scope:** B only, B + D-text, or B + D (with image)?
- [ ] **Sidecar runtime:** Ollama (simplest, needs EG2 support confirmed in the installed version) vs a small Python sentence-transformers server (reference impl, multimodal certain)?
- [ ] **Photo embedding (D-image):** acceptable under the shops' terms in a non-commercial demo? This is an owner decision like the existing display-bar decision. Fetch is gated on it.
- [ ] **Jev account:** OK to add a proprietary paid dependency (≈$1/full run) to a portfolio repo? Alternative: Option C for a zero-key path, with Jev as an opt-in flag like the OpenAI advisor mode.
- [ ] **Merged-comp confidence:** should a Jev-verified cluster lift `name:` groups out of the forced-`low` cap (`radar.go` "never vouch for it")? Proposed: yes to `medium` max, `high` only with a model code.
- [ ] **Gold set labelling:** who labels the ~200 pairs? Claude drafts, owner spot-checks?
- [ ] **UI exposure:** show "同款判定：AI 比對 (信心 0.93)" on cross-shop comps (product thinking signal), or keep it silent?

---

## Next Steps

If accepted:
1. Preplan spike (`/preplan-projex`): EG2 via Ollama + 300 listings + a Jev call on 50 pairs to confirm runtime, prefixes, latency and real CJK behaviour before the plan.
2. Build the gold set (~200 pairs, stratified: code-match / name-variant / same-line-diff-size / different).
3. Plan Option B: store schema, `kura embed`, `kura match`, radar cluster grouping, eval report (before/after: % valued, multi-source groups, merge precision).
4. Re-measure the table in Problem Statement, then add a README section with the numbers.
5. Decide D-text / D-image from the Open Questions, then a separate plan.

---

## Appendix

### Research / References

- EG2 model card: https://huggingface.co/google/embeddinggemma-2. 740M total (130M backbone, 140M embedder, vision 170M, audio 300M); 768-d native, MRL 512/256/128 ("minimal quality impact down to 256d"); 8,192-token shared context; image ≈ 280 tokens; MTEB multilingual v2 61.36 (EG1 61.15), MTEB code 78.68 (EG1 68.76), MIEB lite 64.64. Prefixes: query `task: search result | query: {q}`, doc `title: {title|none} | text: {content}`; no prefix for images; interleaved text+`<|image|>` inputs return one vector.
- EG2 coverage: https://www.eesel.ai/blog/embeddinggemma-2 · https://vpsranking.com/news/ai/ai-2026-10-06-google-embeddinggemma-2/. Released 2026-10-06; 270M text / 440M text+vision / 570M text+audio / 740M all, same space; ~191 MB text-only quantized on Pixel per Google's example; Ollama / vLLM / LiteRT / MediaPipe support listed.
- Jev docs: https://docs.typesafe.ai/ · https://docs.typesafe.ai/models. Choice (`choice`, `probabilities`, `confidence`), Score (`score`, `probabilities`, `confidence`), Noul (0–1); questions evaluated in parallel against one state; "English is the primary training language… CJK scripts are handled but not equally well; test on your own content".
- Jev listing: https://www.llmreference.com/model/jev. `jev-1.13.0`, aliases `jev-latest` / `jev-preview`; 64K ctx (32K state + longest question); text only; $0.042/MTok in, output free; `POST https://api.typesafe.ai/v1/systemone`; GA 2026-09-20.
- Kura code: `internal/listing/listing.go` (`GroupKey`), `internal/radar/radar.go` (`MinComps`, `Engine.value`, `matchAny`, `Search`), `internal/advisor/tools.go` (`search_listings`).

### Alternatives Considered

- **Claude for pair verification:** works, but it costs ~50–100× more per pair at 50k pairs and gives no calibrated probability. Jev's typed output is the reason to bring it in. Claude stays the advisor voice.
- **OpenAI / Gemini hosted embeddings:** a per-call cost plus a third vendor. No image+text shared space at EG2's size, and not local. EG2 is open, so the hosted mock stays self-contained.
- **Vector DB (sqlite-vec, pgvector, Qdrant):** unnecessary at 16k × 256-d. Brute force in Go takes milliseconds, and sqlite-vec needs cgo/extension loading (Kura uses pure-Go `modernc.org/sqlite`). Revisit beyond ~500k listings.
- **Jev for listing-level extraction across the catalog (Option A):** rejected as the primary path because of curation load and CJK weakness. Kept as a possible U2 refinement inside large clusters.
