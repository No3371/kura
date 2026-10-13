PRAGMA journal_mode = WAL;
PRAGMA foreign_keys = ON;

-- One row per item ever seen. Rows are never deleted: a listing that disappears
-- from its source becomes status='gone', which is what lets the radar estimate
-- how fast a model sells.
CREATE TABLE IF NOT EXISTS listings (
    id            TEXT PRIMARY KEY,        -- "<source>:<source_id>"
    source        TEXT NOT NULL,
    source_id     TEXT NOT NULL,
    target        TEXT NOT NULL,           -- crawl target (brand) that discovered it
    url           TEXT NOT NULL,
    image_url     TEXT NOT NULL DEFAULT '',
    brand         TEXT NOT NULL,
    title         TEXT NOT NULL,
    model_name    TEXT NOT NULL DEFAULT '',
    model_code    TEXT NOT NULL DEFAULT '',
    group_key     TEXT NOT NULL DEFAULT '',
    category      TEXT NOT NULL DEFAULT 'bag',
    gender        TEXT NOT NULL DEFAULT '',
    grade         TEXT NOT NULL DEFAULT '',        -- the source's own condition grade
    rank          TEXT NOT NULL DEFAULT '',        -- internal scale derived from grade
    badges        TEXT NOT NULL DEFAULT '',
    price_jpy     INTEGER NOT NULL,
    was_price_jpy INTEGER NOT NULL DEFAULT 0,
    color         TEXT NOT NULL DEFAULT '',
    material      TEXT NOT NULL DEFAULT '',
    size          TEXT NOT NULL DEFAULT '',
    accessories   TEXT NOT NULL DEFAULT '',
    notes         TEXT NOT NULL DEFAULT '',
    status        TEXT NOT NULL DEFAULT 'active',  -- active | gone
    first_seen    INTEGER NOT NULL,                -- unix seconds
    last_seen     INTEGER NOT NULL,
    gone_at       INTEGER,
    missed_sweeps INTEGER NOT NULL DEFAULT 0,
    detail_at     INTEGER
);
CREATE INDEX IF NOT EXISTS listings_group  ON listings(brand, group_key);
CREATE INDEX IF NOT EXISTS listings_target ON listings(target, status);
CREATE INDEX IF NOT EXISTS listings_detail ON listings(detail_at, status);

-- Append-only price history: a row is written on first sight and whenever the
-- asking price changes, so price drops are first-class events.
CREATE TABLE IF NOT EXISTS price_points (
    listing_id  TEXT NOT NULL REFERENCES listings(id),
    observed_at INTEGER NOT NULL,
    price_jpy   INTEGER NOT NULL,
    PRIMARY KEY (listing_id, observed_at)
);

-- A listing's product-page photo gallery, in the shop's order; when present
-- the detail page shows it in place of listings.image_url (the list-page
-- photo). Written by whatever collects them (SaveDetail on each product-page
-- read, or SetImages); the service only reads what is here. AUTOINCREMENT
-- keeps id growing across replacements so Version sees a re-shot gallery
-- with the same photo count.
CREATE TABLE IF NOT EXISTS listing_images (
    id          INTEGER PRIMARY KEY AUTOINCREMENT,
    listing_id  TEXT NOT NULL REFERENCES listings(id),
    pos         INTEGER NOT NULL,
    url         TEXT NOT NULL,
    UNIQUE (listing_id, pos)
);

CREATE TABLE IF NOT EXISTS crawl_runs (
    id            INTEGER PRIMARY KEY AUTOINCREMENT,
    source        TEXT NOT NULL,
    target        TEXT NOT NULL,
    started_at    INTEGER NOT NULL,
    finished_at   INTEGER,
    pages         INTEGER NOT NULL DEFAULT 0,
    seen          INTEGER NOT NULL DEFAULT 0,
    new_items     INTEGER NOT NULL DEFAULT 0,
    price_changes INTEGER NOT NULL DEFAULT 0,
    gone          INTEGER NOT NULL DEFAULT 0,
    complete      INTEGER NOT NULL DEFAULT 0,  -- 1 if every page was swept
    error         TEXT NOT NULL DEFAULT ''
);

-- 每日一品: Home's featured listing for each Taipei calendar day. The first
-- request of a day picks it; keeping the pick is what lets a listing that sells
-- during its day stay up, marked sold, instead of being silently replaced.
CREATE TABLE IF NOT EXISTS daily_picks (
    day        TEXT PRIMARY KEY,                       -- YYYY-MM-DD, Asia/Taipei
    listing_id TEXT NOT NULL REFERENCES listings(id),
    picked_at  INTEGER NOT NULL
);

CREATE TABLE IF NOT EXISTS fx_rates (
    base       TEXT NOT NULL,
    quote      TEXT NOT NULL,
    rate       REAL NOT NULL,
    as_of      INTEGER NOT NULL,
    provider   TEXT NOT NULL,
    PRIMARY KEY (base, quote, as_of)
);
