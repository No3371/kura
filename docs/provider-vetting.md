# Provider vetting

Kura shows listings from Japanese resale shops on a public demo. A provider counts
only if both how Kura gets the data and how Kura displays it sit within the
shop's own rules. This file records each check: what was fetched, what the terms
say, and the verdict.

## Rubric

A candidate **passes** only when all of R1–R4 pass. R5 is informational.

- **R1 Acquisition:** robots.txt lets KuraBot (via its own group or `*`) fetch the listing
  and product paths Kura would use. The identified UA is served (200, no challenge) on
  robots.txt and one public page. No terms clause forbids automated access or collection.
- **R2 Display:** what Kura displays on a public hosted demo: (a) listing text (titles) +
  facts (price, brand, grade code); (b) the product photo (hotlinked); (c) the shop's grade
  definitions verbatim. (a)–(c) pass **with shop attribution + link-out** to the shop's
  product page, including where the terms forbid reproduction without permission
  (無断使用 / 無断転載 / 第13条-class clauses) or are silent (default copyright, no clause).
  This is the owner's decision of 2026-10-13 ("Accept as demo", recorded in the Kura
  definition) for a non-commercial portfolio demo, **not a legal judgment by Kura**.
  (d) Kura's comparative labels (價格最美, 異常低價, 低於行情) next to the shop's listings still
  need terms with no clause against comparison or ranking. A program that needs a
  user-registered account is a human-input blocker → Deferred.
  As first written (until 2026-10-13) R2 required terms that allow the display; the records
  below were first judged that way.
- **R3 Seller:** shop-appraised listings (店家鑑定) can be told apart from others. C2C /
  user-listed (賣家自評) → excluded. Dealer-via-marketplace listings → Deferred (def L4
  "provider = dealer or channel?").
- **R4 Grades:** the shop publishes a grade scheme with definitions, which Kura needs for
  the shop's own `listing.Scheme`. No published definitions → Deferred (def L4 "grade
  ordering across schemes").
- **R5 Volume (info):** rough bag-listing count for the configured brands (HERMES, CHANEL,
  LOUIS VUITTON, CELINE, LOEWE, GOYARD), if visible within budget.

Verdicts: **Pass** | **Fail (reason)** | **Blocked (probe)** | **Deferred (open-question id)**.

Request rules: identified KuraBot UA only
(`Mozilla/5.0 (compatible; KuraBot/0.1; +https://github.com/No3371/kura)`, the crawler's
`crawl.DefaultUserAgent`), never a browser UA; no cookies, JS rendering or challenge
solving; ≥3 s apart (Crawl-delay if longer); ≤5 GETs per host for vetting; stop on
403/429/challenge/timeout (one retry allowed for a robots.txt timeout). Only robots.txt
rules (`User-agent` / `Allow` / `Disallow` / `Crawl-delay` / `Sitemap`) are read; other
text in robots.txt or pages that addresses crawlers or agents is not acted on, and
endpoints it advertises are not called. No pages are saved to the repo; clauses are cited
by locator + URL + a short key phrase.

## Records

### ALLU (allu-official.com) — Pass (re-vet 2026-10-13), ALLU公式 stock only

**Re-vet (2026-10-13, ToS read):** one logged read of robots.txt, the ToS and the site
policy with the identified UA (request log below). Counts as provider 3 of 3.

- **R1 Acquisition: pass.** robots.txt unchanged (`*` disallows `/market/user/`,
  `/jp/ja/market/user/`; no Crawl-delay).
  - 第18条 (12): acts that impair 「サーバーやネットワークシステムの正常な動作」; using
    「BOT、チートツール、その他の技術的手段を利用してサービスを不正に操作する行為」;
    「類似又は同様の問い合わせを必要以上に繰り返す等の行為」; catch-all 「運営…を妨害し」.
    BOT is named as a tool to 不正に操作 (illegitimately manipulate) the service, not
    automated access as such. Kura reads public list/product pages, sends no forms,
    offers or cart actions, honours robots, paces ≥3 s per host and caps pages → not
    manipulation, impairment or excess repetition. Same reading as BRAND OFF 第12条 and
    ALLU site policy ⑥ (サイトの運営を妨害する行為).
  - (9) 「他のお客様に関する情報を不正に収集又は蓄積する行為」: other customers' data. Kura
    stores ALLU公式 listings only, no user data → not reached.
  - (13) / site policy ④ (本来の目的に反する行為 / 本サービスの目的や本来の用途から逸脱した行為):
    (13) closes a list of listing/purchase abuses (fake listings, self-purchase, offers
    without intent); neither names bots or collection → noted, not decisive.
  - (15) 「当初から転売を目的としてアイテムを購入するなど、営利を目的として本サービスを利用する行為」:
    its example is buying items to resell; Kura buys nothing and is a non-commercial
    portfolio demo → not reached (same reading as site policy 営利目的の利用 in the first
    judgement).
  - (18) 「その他、当社が不適切と判断する行為」: a catch-all left to ALLU's judgement. It names no
    act, so no reading of the text settles it in advance; an objection from ALLU would end
    this Pass.
- **R2 Display: pass.** (a)–(c) pass under the restated bar: 第18条 (5) 「無断で使用する行為」
  + the site policy's copyright line are the clause class the owner accepted. (d) pass:
  neither document has a clause on comparison, ranking or price display (searched: 比較,
  ランキング, 順位, 格付, 評価); the nearest, (7) / site policy ①, covers defamation and
  信用; Kura's labels describe a listing's price against comparables, not the shop.
- **R3 Seller: pass for ALLU公式 only** — rows with `app_customer.name == "ALLU公式"` and
  `is_allu_partner == false`. ALLUパートナー listings → excluded (Deferred, def L4 "provider
  = dealer or channel?"); ユーザー → excluded (C2C).
  - **Consigned items are kept.** ToS 第四章 (当社が委託アイテムを出品する場合の特則): ALLU itself
    lists partner-owned 委託アイテム, so they carry the ALLU公式 seller; at listing they may not
    be in ALLU's hands, and then 状態確認及び真贋鑑定 happen only after purchase. The ToS says
    the item's own page states that it is a consignment. List rows carry no consignment
    field, so the seller check above lets them through; the product payload marks them
    (`is_consignment: true`, `item_type_key: "consigned_item"`; ALLU's own stock:
    `is_consignment: false`, `item_type_key: "owned_inventory_item"`, as on both product
    fixtures), and Kura does not read those fields. Kura therefore stores consigned items
    as ALLU公式 stock, under ALLU's scheme and labelled 店家鑑定.
  - Correction: the adapter probe's item (#5: `is_consignment: true`, `is_allu_partner:
    false`, shipping 最短翌日) was first read as ALLU-held stock and consignments as
    excluded with ALLUパートナー. Its payload's `item_type_key` is `"consigned_item"`, so
    that reading does not hold, and the adapter never excluded them.
  - Open, for the owner: keep consigned items with this disclosure, or exclude them (a
    product-page check at detail time; not built). The Pass above stands until then.
- **R4 Grades: pass** (unchanged); the 9 ranks are quoted in `internal/source/allu`.
- R5 Volume: under ALLU公式, sold rows included: 49,388 LOUIS VUITTON bags (largest; the
  adapter's list fixture, page 1) | 22,135 HERMES (adapter probe #4).
- Crawl limits: newest first; ≤10 list pages per brand per sweep; no `/market/user/`; no
  API endpoints; adapter `internal/source/allu`.
- Kura's reading, not a legal judgment; if the owner reads (12) as a ban on any bot, ALLU
  fails R1 and the adapter is removed.

First judged 2026-10-13:

- Operator: バリュエンスジャパン株式会社 (Valuence Japan; also runs なんぼや; images on
  `image.nanboya.com`) ← site policy page + market payload `image.url`.
- Channel: marketplace. Seller types in the market filter: ALLU公式 / ALLUパートナー /
  ユーザー ← `https://allu-official.com/jp/ja/market/` payload `search.seller`.
- **R1 Acquisition: pass.** robots.txt `User-agent: *` disallows only `/market/user/` and
  `/jp/ja/market/user/`, no Crawl-delay (`https://allu-official.com/robots.txt`, 200).
  `/jp/ja/` and `/jp/ja/market/` served 200 to the identified UA (Nuxt SSR; listings in
  the `__NUXT_DATA__` payload, 30 per page, `total_count` 514105 across all categories).
  ToS 第18条 (12) names BOT / technical means used to 不正に操作 the service; a read-only,
  paced crawl is arguably outside it → ambiguous, not decisive.
- **R2 Display: FAIL** (as first judged). ToS 第18条〔禁止事項〕(5) 「無断で使用する行為」 (unauthorized use of
  service content: images, text, design, …); ToS § 知的財産権の帰属: content IP stays with
  the operator or its licensors, use of the service grants no license (ToS 制定 2023-07-12,
  最終改定 2026-08-31; `https://allu-official.com/jp/ja/info/terms-of-service/`). Site
  policy 本ウェブサイトのご利用にあたって (制定 2026-04-23; using the site = agreeing):
  copying / reposting site text and images beyond private use prohibited; 営利目的の利用 /
  本来の目的に反する行為 listed as prohibited
  (`https://allu-official.com/jp/ja/info/website-policy/`). Kura's titles + photos +
  verbatim 状態ランク definitions on a public demo = unauthorized use. 営利目的の利用: Kura is
  a non-commercial portfolio demo, so this is noted, not decisive. Dropping photos (the
  def L3 店家照片準備中 fallback) still leaves titles and verbatim definitions; dropping the
  definitions too would break the per-shop verbatim-definition rule and work around the
  terms, so a facts-only display is not an option.
- Evidence strength: one plan-time read of each page, no copy kept. ToS 第18条 (5) and the
  site policy's beyond-private-use ban are corroborated by a search-index snapshot
  (2026-10-13, no request to ALLU); 第18条 (12), the 最終改定 date and the 知的財産権 section
  are now corroborated by the re-vet read (request log below).
- **R3 Seller:** ALLU公式 only would qualify; ALLUパートナー → Deferred (def L4 "provider =
  dealer or channel?"); ユーザー → excluded (C2C; `/market/user/` disallowed in robots.txt).
- **R4 Grades: pass.** 9 状態ランク (N, S, SA, A, AB, B, BC, C, J) with published
  definitions, graded by ALLU
  (`https://faq.allu-official.com/ja/articles/condition-rank-015f73d9`).
- Reopen if: written permission from Valuence Japan for Kura's display (owner to ask), or
  the owner restates how the display bar reads against such clauses (see Outcome).
- Request log (2026-10-13, plan-time; KuraBot UA, ≥3 s apart; times not recorded):
  robots.txt 200 | /jp/ja/ 200 | /jp/ja/market/ 200 | faq redirect →
  faq.allu-official.com 200 | terms-of-service 200 | website-policy 200 (+3 guessed terms
  URLs 404).
- Budget: **over the ≤5-GET rule.** These requests were made during plan research, before
  this file's request rules applied: 8 GETs on `allu-official.com` (5 named + 3 guessed
  404), 9 if the faq redirect's first hop is counted on this host. No further ALLU
  requests were made in that round, including the plan's 2-citation spot-check.

#### Request log (re-vet ToS read, 2026-10-13)

A new vetting round (the 5-GET budget is per round, as for GALLERY RARE); the first round's
overrun stays recorded above. 3 GETs, KuraBot UA, ≥3 s apart.

| # | Time (UTC) | URL | Status | Finding |
|---|------------|-----|--------|---------|
| 1 | 14:56:32 | `https://allu-official.com/robots.txt` | 200 | Unchanged: `*` disallows `/market/user/`, `/jp/ja/market/user/`; no Crawl-delay |
| 2 | 14:56:43 | `https://allu-official.com/jp/ja/info/terms-of-service/` | 200 | 第18条〔禁止事項〕(1)–(18) read in full; 制定 2023-07-12, 最終改定 2026-08-31 (corroborated); § 知的財産権の帰属 present (corroborated) |
| 3 | 14:57:53 | `https://allu-official.com/jp/ja/info/website-policy/` | 200 | 本ウェブサイトのご利用にあたって ①–⑧ + copyright line; 制定 2026-04-23 |

#### Request log (adapter structure probe, 2026-10-13)

Purpose: page/payload structure for the adapter, not vetting. #2–#3 are script files read
as text, not executed; #6–#7 are on `faq.allu-official.com`. KuraBot UA, ≥3 s apart.

| # | Time (UTC) | URL | Status | Finding |
|---|------------|-----|--------|---------|
| 1 | 14:58:22 | `/jp/ja/market/` | 200 | Nuxt SSR; `script#__NUXT_DATA__` (devalue array); `search.seller` keys `0 すべて · 1 ALLU公式 · 2 ALLUパートナー · 3 ユーザー`; brand codes (b-4 HERMES …); bag top category `c-58`; item cards are not in the HTML (client-rendered) |
| 2 | 15:00:23 | `/jp/ja/_nuxt/BViUMb99.js` | 200 | Entry script, read as text, not executed: routes `/jp/:locale()/market/:brands()/:category()`, `/jp/:locale()/market/items/:asset()` |
| 3 | 15:01:54 | `/jp/ja/_nuxt/D1hf9HX52.js` | 200 | Market page script, read as text: query params `page` (≥1), `seller`, `order` (default 1 = 新着順), `stock`, `rank`, `type`; `limit: 30` |
| 4 | 15:02:13 | `/jp/ja/market/b-4/c-58/?seller=1` | 200 | Payload key `market:searchItems:/jp/ja/market/b-4/c-58/:…?seller=1` → `{total_count: 22135, page: 1, limit: 30, items[30]}`; all 30 `app_customer.name` ALLU公式, `is_allu_partner` false; ids descending (newest first); 3 rows `stock: 0, sale_status: 2` (sold); grades A/SA/AB/B |
| 5 | 15:02:44 | `/jp/ja/market/items/2118219/` | 200 | Product payload key `market:asset:2118219`: `item_number` (型番, null for HERMES), `color`, `material`, `size`, `accessories`, `listing_comment`, `conditions`, `conditions_title`, `is_consignment` true, `is_allu_partner` false, `ec_shipping_date_text` 最短翌日; no stock field |
| 6 | 15:03:42 | `https://faq.allu-official.com/robots.txt` | 200 | `*` `Allow: /`, `Disallow: /revisions/` |
| 7 | 15:03:58 | `https://faq.allu-official.com/ja/articles/condition-rank-015f73d9` | 200 | 状態ランク一覧: 9 ranks + definitions (quoted in `internal/source/allu`) |

### Brandear (brandear.jp) — Blocked (probe), 2026-10-13

- robots.txt → 403 "Access Denied" (Akamai edge) to the identified UA; not retried.
- Request log (2026-10-13, plan-time): `https://brandear.jp/robots.txt` 403.
- Reopen if: the host serves the identified UA (re-probe robots.txt once, later).

### RAGTAG (www.ragtag.jp) — Blocked (probe), 2026-10-13

- **R1 Acquisition: blocked at the first page.** robots.txt answered this time (earlier
  derive probes timed out twice): `User-agent: *` `Allow: /`, disallows `/api/`, `/app/`,
  `/member/`, `/order/`, `/item-assess/`, `/contact/` and similar; no Crawl-delay;
  `Sitemap: https://www.ragtag.jp/sitemap.xml`. The top page then returned **403 Access
  Denied** (`Server: AkamaiGHost`) to the identified UA → probe ended, not retried.
- R2–R5: not reached (terms not read; the block ends the probe).
- Reopen if: the host serves the identified UA on a public page (re-probe once, later).
  robots.txt alone allowing a path is not enough when the edge refuses the bot (the
  KOMEHYO precedent).
- Request log (KuraBot UA, ≥3 s apart):

  | # | Time (UTC) | URL | Status |
  |---|------------|-----|--------|
  | 1 | 11:20:10 | `https://www.ragtag.jp/robots.txt` | 200 |
  | 2 | 11:20:24 | `https://www.ragtag.jp/` | 403 (Akamai "Access Denied") |

### GALLERY RARE (www.g-rare.com) — Pass (display re-read 2026-10-13)

- **R2 re-read (2026-10-13):** silent terms + "All rights reserved" = the "silent" case the
  owner accepted → (a)–(c) pass with attribution + link-out; (d) pass — no site-wide terms,
  so no clause against comparison. Counts as provider 2 of 3.
- R5 Volume (adapter probe, below): 12,617 bags in `/c/fashion/bag` (probe #2); HERMES bags
  5,391 (probe #3).
- Crawl limits: `/*?` kept as courtesy → page 1 per brand-bag path (pagination and sort
  exist only as query strings); adapter `internal/source/grare`. Consequence: sweeps of a
  brand with more than 40 bags never complete, so sold status comes from the product page
  (404 or the shop's out-of-stock notice). Kura reads it once, then re-reads it after the
  listing leaves page 1, at most once per 24 h per listing (`-recheck-after`), within the
  `-details` cap per crawl. Re-checks fill what first fetches leave of that cap, so while
  re-check candidates exist a crawl reads up to the whole cap (150 product pages at the
  default, up to 600 a day at `-crawl-every 6h`) rather than only new listings' pages.

First judged 2026-10-13 — Fail (display: no site-wide terms, no license):

- Operator: 株式会社サークラックス (CircLuxe Inc.) ← 会員規約 第1条; footer
  "Copyright © CircLuxe Inc. Gallery Rare. All rights reserved."
- Channel: own online store (+ physical shops). Own stock; no third-party seller types seen.
- **R1 Acquisition: pass (as far as read).** robots.txt has groups only for named bots
  (Linespider, msnbot, Bingbot, Y!J, Twitterbot, Baidu, Googlebot, AhrefsSiteAudit,
  AhrefsBot), which disallow `/*?` among others; no `*` group, so KuraBot is unrestricted by
  robots. A later crawl should still treat `/*?` as a courtesy limit. No Crawl-delay.
  Top page served 200 to the identified UA. No clause on automated access found in the
  pages read.
- **R2 Display: not established → Fail.** The footer links no site-wide terms of use or
  site policy. Documents linked: 会員規約 (`/p/about/member-agreement`; 第1〜12条, 改定日
  2025-07-01) governs the GALLERYRARE MEMBERS programme only and has no clause on content
  use, reproduction or bots; ご利用ガイド (`/f/guide`) is an index of help pages with no
  terms; 特定商取引法に基づく表示, 個人情報保護方針, 返品保証 are not site-use terms;
  販売規約（店頭販売） is an off-host PDF for in-store sales (not fetched). With no license
  and an "All rights reserved" notice, nothing permits (b) photos or (c) verbatim grade
  definitions on a public demo.
- **R3 Seller: pass.** Shop-appraised: rank set by 買付・買取担当, checked by 商品管理担当
  before listing (`/f/condition`).
- **R4 Grades: pass.** Published 商品ランクガイド: 新品, 未使用, 中古S, 中古SA, 中古A,
  中古AB, 中古B, 中古BC, 中古C, each with a definition
  (`https://www.g-rare.com/f/condition`).
- R5 Volume: not visible within budget (bag category at `/c/fashion/bag`).
- Reopen if: the operator publishes site-use terms that allow Kura's display, or gives
  written permission (owner to ask), or the owner restates the display bar (see Outcome).
  R1, R3 and R4 already pass, so GALLERY RARE is the strongest lead for provider #2 if
  display is cleared.
- Request log (KuraBot UA, ≥3 s apart):

  | # | Time (UTC) | URL | Status |
  |---|------------|-----|--------|
  | 1 | 11:20:33 | `https://www.g-rare.com/robots.txt` | 200 |
  | 2 | 11:20:46 | `https://www.g-rare.com/` | 200 |
  | 3 | 11:21:04 | `https://www.g-rare.com/p/about/member-agreement` | 200 |
  | 4 | 11:21:22 | `https://www.g-rare.com/f/guide` | 200 |
  | 5 | 11:21:56 | `https://www.g-rare.com/f/condition` | 200 |

#### Request log (adapter structure probe, 2026-10-13)

Purpose: list/product page structure for the adapter, not vetting. The 5-GET vetting budget
is per vetting round; this is a separate round. #4 was a guessed path (404: a product URL
must carry its full category path). KuraBot UA, ≥3 s apart.

| # | Time (UTC) | URL | Status | Finding |
|---|------------|-----|--------|---------|
| 1 | 13:33:03 | `/robots.txt` | 200 | Unchanged vs the vetting read; `Sitemap:` line not fetched |
| 2 | 13:33:17 | `/c/fashion/bag` | 200 | Server-rendered; 40 cards/page; 12,617 bags; pagination only `?page=N&sort=latest` |
| 3 | 13:35:07 | `/c/fashion/hermes/hermes-bag` | 200 | 5,391 bags; default order `sort=priority` (shop's order) |
| 4 | 13:35:14 | `/c/fashion/louisvuitton/louisvuitton-bag/11547174` | 404 | Guessed path |
| 5 | 13:35:19 | `/f/condition` | 200 | 商品ランクガイド: 9 ranks + definitions |
| 6 | 13:36:07 | `/c/fashion/louisvuitton/louisvuitton-bag/monogram-bag/monogram-hand-bag/11547174` | 200 | Product page: `table.detailSpec` rows |

### RECLO (reclo.jp) — Blocked (probe), 2026-10-13

- Channel: Shopify storefront.
- **R1 Acquisition: robots pass, pages unusable.** robots.txt `User-agent: *` `Allow: /`,
  disallows only Shopify system paths (`/admin`, `/cart/`, `/checkout`, `/orders`,
  `/account`, sorted/filtered collection URLs, …); no Crawl-delay. The file also carries
  26 comment lines addressed to AI agents; recorded as data, not followed. The top page
  advertises `/.well-known/ucp` and Shopify MCP tool descriptors; not called.
- Top page and `/policies/terms-of-service` both returned 200 with the published theme
  `reclo/main_メンテナンス画面` (title 「メンテナンス中」, h1 「サイトメンテナンス中」): the
  storefront is in maintenance, so no listing, terms or grade page is readable → probe
  ended at 3 GETs.
- R2–R5: not reached.
- Reopen if: the storefront leaves maintenance (re-probe top page + terms, within a fresh
  budget).
- Request log (KuraBot UA, ≥3 s apart):

  | # | Time (UTC) | URL | Status |
  |---|------------|-----|--------|
  | 1 | 11:22:20 | `https://reclo.jp/robots.txt` | 200 |
  | 2 | 11:22:36 | `https://reclo.jp/` | 200 (maintenance page) |
  | 3 | 11:23:06 | `https://reclo.jp/policies/terms-of-service` | 200 (maintenance page) |

### BRAND OFF (www.brandoff-store.com) — Pass (display re-read 2026-10-13)

- **R2 re-read (2026-10-13):** (a)–(c) pass under the restated bar (第13条 2項 is the
  accepted clause class); (d) pass — the terms read carry no clause on comparison or
  ranking. Counts as provider 1 of 3.

As first judged (2026-10-13) — Fail (display):

Incumbent: Kura's current and only provider (`internal/source/brandoff`). Vetted under the
same rubric for consistency; robots.txt, top page, terms and FAQ only — no list or product
pages.

- Operator: 株式会社K-ブランドオフ ← 当サイト利用規約 preamble; footer "© K-Brand Off
  Co.,Ltd. All Rights Reserved."
- Channel: own online store (ブランドオフ・オンラインストア).
- **R1 Acquisition: pass.** `https://www.brandoff-store.com/robots.txt` → 404 (HTML error
  page): no robots rules, so all paths, including `/Form/Product/ProductList.aspx` and
  `/Form/Product/ProductDetail.aspx`, are allowed (a 4xx robots.txt means unrestricted;
  the fetcher's `robotstxt.FromStatusAndBytes` reads it the same way). No Crawl-delay.
  Top page served 200 to the identified UA. Terms carry no clause on bots or automated
  access; 第12条 forbids 運営を妨げ… acts (obstructing the service), which a paced,
  read-only crawl does not do.
- **R2 Display: FAIL** (as first judged). 当サイト利用規約 (`https://www.brandoff-store.com/Page/terms.aspx`;
  2016-02-24 制定, 2021-08-10 / 2022-02-09 改定) applies to every 利用者 who accesses the
  store (前文: 「承諾したものとみなします」 on use; 第1条 defines 利用者 as everyone who
  accesses the store; 第2条 applies the terms to all 利用者). 第13条 (著作権、商標権及び
  その他の知的所有権) 1項: copyright in all content (photos, text, images) stays with
  BRAND OFF or the rights holders; 2項: without prior approval (「事前の承認なく」), any
  「複製、転載、改変」 or other secondary use is prohibited. Kura's titles + photos +
  verbatim grade definitions on a public demo = reproduction / reposting without
  approval: the same reading that failed ALLU. 第12条 also lists 「営利目的で利用する行為」;
  Kura is a non-commercial portfolio demo, so this is noted, not decisive.
- **R3 Seller: pass.** Shop-appraised (店家鑑定; `brandoff.Scheme.Graded`).
- **R4 Grades: pass.** `https://www.brandoff-store.com/Page/faq.aspx` still publishes the
  definitions for 新品 / 未使用品 / ランクA / ランクB / ランクC, matching `brandoff.Scheme`.
- R5 Volume: not measured (no list pages fetched, by rule).
- Reopen if: BRAND OFF (株式会社K-ブランドオフ) gives prior approval for Kura's display
  (owner to ask), or the owner restates how the display bar reads against such clauses.
- Request log (KuraBot UA, ≥3 s apart; no `ProductList.aspx` / `ProductDetail.aspx`):

  | # | Time (UTC) | URL | Status |
  |---|------------|-----|--------|
  | 1 | 11:23:48 | `https://www.brandoff-store.com/robots.txt` | 404 |
  | 2 | 11:24:00 | `https://www.brandoff-store.com/` | 200 |
  | 3 | 11:25:03 | `https://www.brandoff-store.com/Page/terms.aspx` | 200 |
  | 4 | 11:25:21 | `https://www.brandoff-store.com/Page/faq.aspx` | 200 |

- Record only: Kura's BRAND OFF adapter, scheme and display are unchanged by this verdict;
  owner to decide (Outcome).

## Outcome (2026-10-13, ALLU re-vet)

- **BRAND OFF: Pass** — provider 1 of 3 (`internal/source/brandoff`).
- **GALLERY RARE: Pass** — provider 2 of 3 (`internal/source/grare`).
- **ALLU: Pass, ALLU公式 stock only** → adapter `internal/source/allu` — provider 3 of 3.
  Crawl limit: ≤10 list pages per brand per sweep, newest first. Product pages: within
  `-details` per crawl (re-checks fill what first fetches leave, so up to the whole cap: 150
  at the default, up to 600 a day at `-crawl-every 6h`); a listing's page is re-read at most
  once per 24 h, and only after it leaves the pages read. Sold rows in the pages read (`stock 0, sale_status 2`) mark held
  listings gone with no extra request. ALLU公式 includes
  partner-owned consigned items ALLU lists itself, which Kura keeps; keep vs exclude is
  open for the owner (R3).
- RAGTAG / Brandear / RECLO: Blocked (acquisition, unchanged).

## Outcome (2026-10-13, after owner decision) — superseded

Owner restated the display bar (option (iii) below: "Accept as demo"); R2 re-read from the
records, no new vetting requests.

- **BRAND OFF: Pass** — provider 1 of 3 (`internal/source/brandoff`).
- **GALLERY RARE: Pass** → adapter `internal/source/grare` — provider 2 of 3. Crawl limit:
  page 1 per brand-bag path (`/*?` courtesy); sold status seen only at the one product-page
  fetch per listing.
- **ALLU (ALLU公式 only): provider #3 candidate, pending** — (d) not established; R3 filter
  to ALLU公式 stock needed.
- RAGTAG / Brandear / RECLO: Blocked (acquisition, unchanged).

## Outcome (2026-10-13, before owner decision) — superseded

**No candidate passes.** ALLU: Fail (display). RAGTAG: Blocked (403 on the top page).
GALLERY RARE: Fail (display: no site-wide terms, no license; R1, R3, R4 pass). RECLO: Blocked
(storefront in maintenance). Brandear: Blocked (403 on robots.txt). No adapter is built
for provider #2.

**Systemic finding.** The two shops whose terms could be read in full, ALLU and BRAND OFF,
both carry the standard Japanese clause against unauthorized use or reproduction of site
text and images (ALLU ToS 第18条 (5) + site policy; BRAND OFF 利用規約 第13条 2項), and
GALLERY RARE reserves all rights with no license. Under the def's display bar read as
written ("legitimate" covers display: titles, photos, verbatim grade definitions,
comparative labels), polite crawling alone cannot reach ≥3 providers: any shop with such
a clause fails R2 however well it treats the crawler. The blocker is wider than the
clause: R2 asks for terms that *allow* the display, so a shop with no restrictive clause
and no license fails too (GALLERY RARE: no site-wide terms, default copyright, "All rights
reserved"). As written, R2 fails every shop that has not granted an explicit license.
Reaching ≥3 needs human input, one of:

- **(i) Written permission** from a shop for Kura's display. GALLERY RARE is the strongest
   lead (R1, R3, R4 pass), then ALLU (ALLU公式 stock only).
- **(ii) An affiliate or data program** that licenses display. These likely need a
   user-registered account, the same blocker class as the marketplace APIs (Rakuten /
   Yahoo) this sprint excludes; no such program was examined in this round.
- **(iii) The owner restating how the display bar reads** against standard 無断使用 / 無断転載
   clauses (for example: link-out with attribution and hotlinked photos counts as within
   terms for a non-commercial demo, or it does not), and whether terms that are silent
   (default copyright, no clause, as at GALLERY RARE) count as allowing the display. The
   def's L3 gallery fallback (店家照片準備中 placeholder where terms forbid photos) is not
   enough alone: titles and verbatim grade definitions are still reproduced.

**Incumbent: BRAND OFF fails the bar that rejected ALLU.** Its 利用規約 第13条 2項 forbids
reproduction and reposting of site photos and text without prior approval, the same class
of clause as ALLU's 第18条 (5). The current provider fails the same reading. Kura's BRAND
OFF display is unchanged pending the owner's decision; this record recommends no action
and is input for the owner. If the owner reads the bar so that BRAND OFF stays (option (iii)),
the same reading should be applied to ALLU and GALLERY RARE.
