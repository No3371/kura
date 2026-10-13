# Kura 蔵 — Japanese pre-owned luxury price radar + AI bag advisor

> 日本二手精品行情雷達與 AI 找包顧問：爬取日本二手精品店的在售包款，計算同型號同成色的行情、降價與銷售速度，並讓 Claude 只根據真實在售商品幫台灣買家找包。

Kura crawls live listings from Japanese luxury resale stores (BRAND OFF, GALLERY RARE, ALLU),
normalises them into a price database, and builds two things on top:

- **Price radar (行情雷達)** — fair value per listing from like-for-like comparables,
  deal scores with honest confidence, price-drop tracking, sell-through and
  time-to-sell per model.
- **AI bag advisor (AI 找包顧問)** — a Claude tool-use agent that answers questions like
  *「預算 8 萬，想找通勤可以放 A4 的 LV 托特包」* by querying the radar, and recommends only
  listings that actually exist, with TWD prices and the reasoning behind each pick.

It is a single Go binary (crawler + API + embedded UI) on SQLite.

```
kura crawl            # sweep the configured brands once
kura serve            # http://127.0.0.1:8080  (UI + API + advisor)
```

---

## Architecture

```mermaid
flowchart LR
    subgraph Ingest
      F[Polite fetcher<br/>robots.txt · Crawl-delay · per-host rate limit<br/>jitter · backoff · honest UA] --> S[Source adapters<br/>brandoff · grare · allu: list + product parsers]
      S --> C[Crawler<br/>sweeps · dedupe · detail budget]
    end
    C --> DB[(SQLite<br/>listings · price_points<br/>crawl_runs · fx_rates)]
    FX[JPY→TWD rate] --> DB
    DB --> R[Radar engine<br/>in-memory index, rebuilt on change<br/>comparables · fair value · drops · stats]
    R --> API[JSON API]
    R --> T[Advisor tools<br/>search_listings · model_price_stats<br/>get_listing · market_overview]
    T --> A[Claude agent loop<br/>streaming · tool use · prompt caching]
    A -- SSE --> UI[Web UI]
    API --> UI
```

| Package | Responsibility |
|---|---|
| `internal/crawl` | Polite HTTP fetcher; crawler that records sweeps, price changes, sold items |
| `internal/source/brandoff` | BRAND OFF adapter: list/product parsers (fixture-tested) and paginated sweeper |
| `internal/source/grare` | GALLERY RARE adapter: list/product parsers (fixture-tested), first-page sweeper |
| `internal/source/allu` | ALLU adapter: Nuxt page-payload parsers (fixture-tested), ALLU公式-only, capped newest-first sweeper |
| `internal/listing` | Source-independent model: condition ranks, categories, brand aliases (EN/中/日) |
| `internal/store` | SQLite schema and queries (pure-Go driver, no cgo) |
| `internal/radar` | Fair value, deal score, confidence, drops, per-model stats |
| `internal/advisor` | Claude and OpenAI agents, tool implementations, offline mock |
| `internal/web` | API, SSE chat endpoint, embedded vanilla-JS UI |

## Design decisions worth reading

### Choosing a source: respect the "no"
The first candidate (KOMEHYO) allows everything in robots.txt but returns errors to any
client that identifies itself as a bot — it only answers browser user agents. Spoofing a
browser to get around that would be evading their bot policy, so Kura doesn't. It uses
BRAND OFF, GALLERY RARE and ALLU (ALLU公式 stock only), which serve an honestly identified crawler
(`KuraBot/0.1 (+repo URL)`).

Acquisition is only half of it: what Kura shows has to be fair to the shop too. What Kura
shows — titles, hotlinked photos, each shop's grade definitions verbatim — always carries
the shop's name and a link to its product page. BRAND OFF's terms forbid reproduction
without prior approval (利用規約 第13条 2項), ALLU's forbid unauthorized use of site content
(利用規約 第18条 (5)), the same class of clause, and GALLERY RARE's are silent (no site-wide
terms). Showing them this way in a non-commercial portfolio demo is the owner's decision,
not a legal judgment. Kura's acquisition rules are unchanged, and a shop whose terms forbid
bots is not crawled. ALLU's terms bar bots used to manipulate the service and requests
repeated beyond need (第18条 (12)); Kura reads only public pages, paced and capped, which it
reads as outside that clause. ALLU's marketplace also carries partner and user listings,
which Kura leaves out. ALLU公式 itself includes partner-owned consigned items that ALLU lists
(利用規約 第四章; they may be inspected only after purchase). Only the product page marks
them, so Kura currently keeps them as ALLU公式 stock, graded by ALLU's scheme. Each
candidate's check is recorded in [docs/provider-vetting.md](docs/provider-vetting.md).

The fetcher also:
- reads robots.txt per host and honours `Crawl-delay` when it's longer than the default,
- runs one request at a time per host at ≥ 3 s plus 0–25% jitter,
- backs off exponentially on 429/5xx and honours `Retry-After`,
- caps product-page fetches per source per run (`-details`); list pages already carry what the
  radar needs. The cap is per source rather than shared so one shop's first-crawl backlog
  can't starve another's details; total fetches per run grow with the number of shops.
  Re-checks of product pages already read (see *Knowing when something sold*) come out of
  the same cap, after first fetches.

### Turning messy listings into comparable data
BRAND OFF titles are Japanese free text with the condition shown only as an icon image.
The adapter:
- maps icon numbers to ranks (新品→N, 未使用品→S, A/B/C),
- splits titles into brand / model name / bag type / gender / model code
  (`ルイ・ヴィトン キーポル55 ボストンバッグ バッグ メンズ レディース M41424`),
- keeps only bag categories (the brand filter returns clothing too; category ids `100xxx` are bags),
- takes the model code from the product page's spec table when available, since that's authoritative.

GALLERY RARE shows the grade as a text mark on each listing (中古A, 未使用, …), which is read
as is. The model code comes from the title token (`M46222`), or the product page's 型番 when
fetched. Kura reads only the first page (40 bags) of each brand's bag category: pagination
and sort exist only as query strings, and Kura leaves those alone (the shop's robots.txt bars
them for the bots it names). The page is in the shop's own order, not newest first. GOYARD
has no bag category there, so it is not swept on GALLERY RARE.

ALLU's pages carry the listing data as Nuxt state (`__NUXT_DATA__`), so there is no title
splitting: model name and model code (型番, when given) are fields, and a code-shaped title
token stands in when 型番 is empty. Many rows have neither: in a 167-row ALLU crawl, 42
(25%) carried no model name, no 型番 and no code token (CHANEL 20 of 27, GOYARD 11 of 29,
CELINE 9 of 27, LOEWE 2 of 28; HERMES and LOUIS VUITTON 0), against 0.7% of BRAND OFF rows.
Those listings have no group, so they get no comparables, fair value or model page. Kura
keeps ALLU公式 rows that are in stock and skips grade J (使用困難): priced against C
comparables, a bag that is hard to use would read as a deal.

### Condition grades stay the shop's
Every shop grades differently. Brand Off uses 新品 / 未使用品 / ランクA / B / C, others add
SA / AB / BC, ALLU uses 9 ranks N … J, and C2C marketplaces like Mercari let sellers grade their own items. Kura does
not translate these into a house scale:
- Each source registers its **own scheme**: labels, its definitions verbatim, and who grades
  (店家鑑定 or 賣家自評). The UI shows the shop's label, a plain-text meaning, and the shop's
  original wording.
- **"Same condition" means same shop, same grade.** Another shop's "A" only counts through
  an internal rank bridge, at lower confidence.
- The internal N–C scale exists only for cross-grade estimates and the "at least this good"
  filter, and is never displayed in place of a grade.
- Older databases are migrated in place, and missing grades are backfilled from each scheme
  where the mapping is unambiguous.

### Knowing when something sold
A listing that vanishes might have sold, or might just have shifted across a page
boundary during a newest-first sweep. Kura marks an item **gone** only after it is missing
from **two consecutive complete sweeps**. Capped sweeps (`-max-pages`) never count a missing
item toward gone (the sold signals below still apply), and reappearing items are reactivated. Time-to-sell only counts items Kura actually
saw being listed (first seen after Kura's first complete sweep of that shop's brand), because
older stock has an unknown listing date.

GALLERY RARE sweeps of a brand with more than 40 bags, and ALLU sweeps (at most 10 pages per
brand, newest first) of a brand with more ALLU公式 bags, are never complete, so missing from
a sweep never marks their items sold. Other signals do:
- **Product-page re-check.** Kura reads a listing's product page first in the crawl that
  first sees it, budget permitting. For a brand whose sweep did not complete, it re-reads the
  page of an active listing that the crawl's list pages no longer show, at most once per
  `-recheck-after` (24 h) per listing, least recently read first. A page that answers 404,
  shows GALLERY RARE's out-of-stock notice or carries ALLU's sold button type marks the item
  gone. Re-checks use only what first fetches leave of `-details`, so a crawl stays within
  `-details` product pages per shop; while re-check candidates exist it can use the whole
  cap (150 product pages per shop per crawl at the default, up to 600 a day at
  `-crawl-every 6h`), not only what new listings need. When listings outnumber that remainder, each is re-checked
  less often than daily. A BRAND OFF brand whose sweep was capped or failed gets the same
  re-checks.
- **Sold rows on ALLU list pages.** ALLU keeps sold items in its list (`stock 0`,
  `sale_status 2`). Such a row in the pages read marks the listing gone if Kura holds it; it
  is never stored as a new listing.
- **SOLDOUT tiles on BRAND OFF list pages.** BRAND OFF keeps sold items on its list pages
  for a while, marked `SOLDOUT`. Such a bag tile marks the listing gone if Kura holds it and
  is never stored as a new listing; it still counts toward the page's size, so a full page
  with sold tiles is not read as the last page.

An item that sells after leaving the pages read is marked gone at its next re-check, not at
the moment of sale, and its price changes are seen only while it stays in the pages read.

### Newest listings
最新上架 (Home's shelf, Browse's newest sort, the advisor's `newest`, the demo member's
新上架) is one order. Listings Kura saw appear — first seen after the first complete sweep
of their shop and brand — come first, latest first. The initial crawl's own first-seen times
are only crawl order, and list pages are swept newest first, so among those listings the
*earliest* first seen is the shop's newest (GALLERY RARE's page is in the shop's own order, so
its listings keep only crawl order, and its first crawl's stock stays ahead of later
arrivals). Listings first seen in the same second (about one list page) fall back to id
order. A shop or brand added later starts with its own initial sweep, so its existing stock
does not flood the order as new; until a shop's brand has had
one complete sweep, none of its listings count as seen appearing. One gap remains: if items
sell while that first sweep runs, later pages shift up and a listing can slip past unread; it
is first seen on the next sweep and reads as newly listed. GALLERY RARE brands with more than
40 bags never get a complete sweep, so their listings never count as seen appearing. The same
holds for ALLU brands with more ALLU公式 bags than its 10-page cap.

### Fair value with honest confidence
- **Comparables** are the same brand and model code (or model name when there's no code),
  covering active listings plus items gone within 180 days.
- **Same-rank median** when there are ≥ 3 comparables (≥ 6 counts as high confidence).
  Otherwise prices are **rank-adjusted** to A-equivalent (S 1.15, A 1.0, B 0.85, C 0.70, …)
  and confidence drops a level.
- **Dispersion check.** The first live run surfaced "64% below market" Hermès Garden Party
  bags: the name-only group mixed canvas and leather versions. Kura now measures IQR/median
  of the comparables and downgrades confidence above 0.4 (low above 0.6). Groups with no
  model code are always low confidence.
- **Anomaly flag.** The full crawl surfaced an Alma PM (M51130, B rank) at ¥22,000. It had
  46 same-rank comparables with a median of ¥80,000, and the next cheapest was ¥50,000. That
  is almost certainly unlisted damage, not a 73% bargain. Anything more than 50% below a fair
  value is flagged 「異常低價 · 請先確認狀況」, never ranked as a deal, and the advisor is told
  to warn rather than recommend.
- **Deal ranking** is weighted by confidence, so a reliable 15% outranks a shaky 50%. The
  UI labels low-confidence cases 「可能低於行情 · 低信心」 and explains why.
- **Detail fetches go where they matter.** About 46% of listings have no model code in the
  title, so the product-page budget goes to those first (the spec table supplies the code),
  round-robin across brands.

### The advisor agent
- **Manual streaming tool loop** on the Anthropic Go SDK (`Beta.Messages.NewStreaming`).
  Text deltas, tool calls and tool results stream to the browser over SSE, so users can
  watch it search.
- **Grounded by construction.** The model only sees what the tools return. The system
  prompt requires every recommendation to cite a listing id (`[[brandoff:…]]`), which the
  UI renders as a clickable card.
- **Titles are Japanese.** The prompt teaches the model to search with katakana or model
  codes (Neverfull → ネヴァーフル); users type Chinese or English.
- **Tool inputs are validated** (unknown fields, unknown brands, bad enums). Errors go back
  as `is_error` tool results so the model corrects itself. Eager input streaming is on, and
  tools are never run after a `max_tokens` stop.
- **Robustness.** Server-side refusal fallback (`fallbacks: "default"`) and prompt caching on
  the static system prompt and tools. Conversation history is append-only, and a failed
  turn rolls back so the session stays valid.
- **OpenAI mode.** With `OPENAI_API_KEY` (or `-advisor openai`) the same prompt, tools and
  tool runner drive the OpenAI Chat Completions API over plain HTTP (`openai.go`);
  `OPENAI_BASE_URL` targets any compatible endpoint.
- **Mock mode.** With neither key, a rule-based advisor runs the same search
  tool, so the whole demo works offline.
- **Tested without a key.** `agent_test.go` and `openai_test.go` replay scripted SSE from
  fake Messages / Chat Completions APIs and assert the request shape and the tool round trip.

## Running it

Requires Go 1.26+.

```bash
go build -o kura ./cmd/kura
./kura crawl -brands "HERMES,CHANEL,LOUIS VUITTON" -details 100
./kura crawl -sources grare               # one shop only (default: -sources brandoff,grare,allu)
export ANTHROPIC_API_KEY=sk-ant-...      # optional; or OPENAI_API_KEY; with neither the advisor runs in mock mode
./kura serve -crawl-every 6h             # keeps the radar fresh; price drops and sold items appear from the 2nd crawl on
```

| Flag | Default | |
|---|---|---|
| `-db` | `kura.db` | SQLite path |
| `-brands` | `HERMES,CHANEL,LOUIS VUITTON,CELINE,LOEWE,GOYARD` | Brands to sweep (EN, 中文 or katakana) |
| `-sources` | `brandoff,grare,allu` | Shops to crawl |
| `-delay` | `3s` | Minimum gap between requests (robots.txt `Crawl-delay` wins if longer) |
| `-max-pages` | `0` (all) | Cap pages per brand; capped sweeps never count unseen items toward sold |
| `-details` | `150` | Product pages fetched per source per crawl (colour, size, accessories) |
| `-recheck-after` | `24h` | Least time between product-page re-checks of one listing, for brands whose sweep didn't complete (0 = off); uses what `-details` leaves |
| `-advisor` | `auto` (`KURA_ADVISOR`) | `claude`, `openai` or `mock`; auto picks by which API key is set (Claude first) |
| `-model` | `claude-opus-5-5` (`KURA_MODEL`) | Claude advisor model |
| `-openai-model` | `gpt-5.4` (`KURA_OPENAI_MODEL`) | OpenAI advisor model |
| `-openai-effort` | unset (`KURA_OPENAI_EFFORT`) | `reasoning_effort` for reasoning models; unset = not sent |
| `-crawl-every` | off | Background crawl interval while serving |

A full sweep of the six default brands is roughly 300 list pages, about 20 minutes at the
default delay. ALLU adds at most 10 pages per brand.

### API

| Endpoint | |
|---|---|
| `GET /api/overview` | Counts, FX rate, advisor mode, recent crawl runs |
| `GET /api/listings?brand=&category=&min_rank=&max_twd=&q=&deals=1&sort=deal` | Search |
| `GET /api/listings/{id}` | One listing with price history and model stats |
| `GET /api/models?brand=` · `GET /api/models/{brand}/{key}` | Per-model statistics |
| `GET /api/drops` | Recent price cuts |
| `GET /api/member` | Demo member 林小姐: a watchlist picked from live models and listings, the price cuts and sales the crawl observed on it, and each tracked model's newest listing (see *Newest listings*) |
| `POST /api/chat` `{session, message}` | Advisor, streamed as Server-Sent Events |
| `POST /api/crawl` | Start a crawl in the background |

## Tests

```bash
go test ./...
```

- **Parsers** run against real saved pages (`internal/source/brandoff/testdata`,
  `internal/source/grare/testdata`, `internal/source/allu/testdata`).
- **Store** covers the price-history and sold-detection lifecycle, including the
  "missed once is not sold" rule and reactivation.
- **Crawler** covers per-source detail budgets, re-check priority and cadence, and sold rows
  on list pages.
- **Radar** covers same-rank and rank-adjusted valuation, dispersion downgrade, drops,
  stats and filters.
- **Agent** runs the full tool loop against a fake streaming API.

## Limitations and next steps

- **Three sources.** BRAND OFF, GALLERY RARE and ALLU (ALLU公式). The
  `crawl.Source` interface is the seam for more resellers. More sources also means
  cross-site comparables and price arbitrage between shops. GALLERY RARE contributes at most
  40 bags per brand per crawl, so its listings' comparables lean on BRAND OFF through the
  rank bridge, at lower confidence. GALLERY RARE and ALLU never complete a sweep of a large
  brand, so sold status for their listings comes from product-page re-checks (and ALLU's
  sold list rows) and can lag the sale by a re-check cycle, and 最新上架 relies on crawl
  order. About a quarter of ALLU rows (mostly CHANEL, GOYARD, CELINE) have no model
  name or code and so no fair value; a title-derived model name would close part of that.
  ALLU's consigned items are kept as ALLU公式 stock; telling them apart needs a
  product-page check.
- **Variant-aware grouping.** Leather, colour and hardware drive Hermès/Chanel prices, but
  most titles don't carry them. Product-page materials (already stored) could split
  name-only groups.
- **Landed cost.** Prices exclude international shipping and Taiwan import duties/VAT. A
  landed-cost estimator would make TWD comparisons with local listings fair.
- **Marketing automation.** Member segments from advisor conversations, plus Claude-drafted
  LINE/EDM copy for "price dropped on a model you asked about" alerts. The data model
  (price-drop events, per-model interest) already supports it.
- **Observability.** The fetcher keeps request/byte/retry/error counters; exposing them and
  sweep durations as Prometheus metrics is a small step.
- Product images are hotlinked from each shop with its name and a link back; a production
  deployment would need the shops' permission (their terms reserve reproduction) and would
  cache thumbnails.
