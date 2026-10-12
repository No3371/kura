// Package store persists listings, price history, crawl runs and FX rates in
// SQLite (pure-Go driver, no cgo).
package store

import (
	"context"
	"database/sql"
	_ "embed"
	"errors"
	"fmt"
	"strings"
	"time"

	_ "modernc.org/sqlite"

	"github.com/No3371/kura/internal/listing"
)

//go:embed schema.sql
var schema string

// GoneAfterMissedSweeps is how many consecutive complete sweeps a listing must
// be absent from before it is considered sold. One miss can be a pagination
// artefact (an item shifting across a page boundary mid-sweep); two is not.
const GoneAfterMissedSweeps = 2

type Store struct{ db *sql.DB }

func Open(path string) (*Store, error) {
	db, err := sql.Open("sqlite", "file:"+path+"?_pragma=busy_timeout(5000)")
	if err != nil {
		return nil, err
	}
	db.SetMaxOpenConns(1) // SQLite has one writer; serialise instead of SQLITE_BUSY
	s := &Store{db: db}
	if err := s.migrate(); err != nil {
		db.Close()
		return nil, fmt.Errorf("store: migrate: %w", err)
	}
	return s, nil
}

func (s *Store) migrate() error {
	if _, err := s.db.Exec(schema); err != nil {
		return err
	}
	// Databases created before per-source grades lack the column.
	var n int
	if err := s.db.QueryRow(`SELECT COUNT(*) FROM pragma_table_info('listings') WHERE name='grade'`).Scan(&n); err != nil {
		return err
	}
	if n == 0 {
		if _, err := s.db.Exec(`ALTER TABLE listings ADD COLUMN grade TEXT NOT NULL DEFAULT ''`); err != nil {
			return err
		}
	}
	return nil
}

// BackfillGrades fills missing grade codes from the internal rank, for
// sources whose scheme maps ranks to grades one-to-one. It is idempotent and
// only touches rows that predate per-source grades.
func (s *Store) BackfillGrades(ctx context.Context, schemes ...listing.Scheme) (int64, error) {
	var total int64
	for _, sc := range schemes {
		byRank := map[listing.Rank][]string{}
		for _, g := range sc.Grades {
			byRank[g.Rank] = append(byRank[g.Rank], g.Code)
		}
		for rank, codes := range byRank {
			if len(codes) != 1 {
				continue // ambiguous: leave blank rather than guess
			}
			res, err := s.db.ExecContext(ctx, `UPDATE listings SET grade=? WHERE source=? AND rank=? AND grade=''`,
				codes[0], sc.Source, string(rank))
			if err != nil {
				return total, err
			}
			n, _ := res.RowsAffected()
			total += n
		}
	}
	return total, nil
}

func (s *Store) Close() error { return s.db.Close() }

// Row is a stored listing with its lifecycle fields.
type Row struct {
	listing.Listing
	Color, Material, Size, Accessories, Notes string

	ID           string
	Target       string
	Status       string
	FirstSeen    time.Time
	LastSeen     time.Time
	GoneAt       time.Time // zero while active
	HasDetail    bool
	PriceHistory []PricePoint
}

type PricePoint struct {
	At       time.Time
	PriceJPY int64
}

// SeenResult reports what an observation changed.
type SeenResult struct {
	New      bool
	OldPrice int64 // previous asking price when it changed, else 0
}

// RecordSeen upserts a listing observed during a sweep at time now.
func (s *Store) RecordSeen(ctx context.Context, target string, l listing.Listing, now time.Time) (SeenResult, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return SeenResult{}, err
	}
	defer tx.Rollback()

	var prev int64
	err = tx.QueryRowContext(ctx, `SELECT price_jpy FROM listings WHERE id = ?`, l.ID()).Scan(&prev)
	isNew := errors.Is(err, sql.ErrNoRows)
	if err != nil && !isNew {
		return SeenResult{}, err
	}
	ts := now.Unix()
	badges := strings.Join(l.Badges, ",")
	if isNew {
		_, err = tx.ExecContext(ctx, `INSERT INTO listings
			(id, source, source_id, target, url, image_url, brand, title, model_name, model_code, group_key,
			 category, gender, grade, rank, badges, price_jpy, was_price_jpy, first_seen, last_seen)
			VALUES (?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
			l.ID(), l.Source, l.SourceID, target, l.URL, l.ImageURL, l.Brand, l.Title, l.ModelName, l.ModelCode,
			l.GroupKey(), l.Category, l.Gender, l.Grade, string(l.Rank), badges, l.PriceJPY, l.WasPriceJPY, ts, ts)
	} else {
		// A listing that reappears after being marked gone is active again.
		// Shops occasionally regrade an item; the latest grade wins.
		_, err = tx.ExecContext(ctx, `UPDATE listings SET
			url=?, image_url=?, title=?, grade=?, rank=?, badges=?, price_jpy=?, was_price_jpy=?,
			last_seen=?, status='active', gone_at=NULL, missed_sweeps=0
			WHERE id=?`,
			l.URL, l.ImageURL, l.Title, l.Grade, string(l.Rank), badges, l.PriceJPY, l.WasPriceJPY, ts, l.ID())
	}
	if err != nil {
		return SeenResult{}, err
	}
	res := SeenResult{New: isNew}
	if isNew || prev != l.PriceJPY {
		if !isNew {
			res.OldPrice = prev
		}
		if _, err := tx.ExecContext(ctx, `INSERT OR REPLACE INTO price_points (listing_id, observed_at, price_jpy) VALUES (?,?,?)`,
			l.ID(), ts, l.PriceJPY); err != nil {
			return SeenResult{}, err
		}
	}
	return res, tx.Commit()
}

// SweepMissing is called after a complete sweep of target that started at
// sweepStart. Active listings not seen during it get a missed sweep; those
// reaching GoneAfterMissedSweeps are marked gone. Returns how many went gone.
func (s *Store) SweepMissing(ctx context.Context, source, target string, sweepStart, now time.Time) (int, error) {
	if _, err := s.db.ExecContext(ctx, `UPDATE listings SET missed_sweeps = missed_sweeps + 1
		WHERE source=? AND target=? AND status='active' AND last_seen < ?`, source, target, sweepStart.Unix()); err != nil {
		return 0, err
	}
	res, err := s.db.ExecContext(ctx, `UPDATE listings SET status='gone', gone_at=?
		WHERE source=? AND target=? AND status='active' AND missed_sweeps >= ?`,
		now.Unix(), source, target, GoneAfterMissedSweeps)
	if err != nil {
		return 0, err
	}
	n, _ := res.RowsAffected()
	return int(n), nil
}

// DetailTodo is a listing that still needs its product page fetched.
type DetailTodo struct{ ID, URL string }

// PendingDetails returns active listings without detail. Listings lacking a
// model code come first (the product page is what can supply one, moving the
// item from fuzzy name-only comparables to like-for-like ones), then targets
// are interleaved so one brand doesn't consume the whole budget.
func (s *Store) PendingDetails(ctx context.Context, source string, limit int) ([]DetailTodo, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, url FROM (
			SELECT id, url, model_code = '' AS no_code, first_seen,
			       ROW_NUMBER() OVER (PARTITION BY target, model_code = '' ORDER BY first_seen DESC, id) AS turn
			FROM listings
			WHERE source=? AND status='active' AND detail_at IS NULL)
		ORDER BY no_code DESC, turn, first_seen DESC, id LIMIT ?`, source, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []DetailTodo
	for rows.Next() {
		var t DetailTodo
		if err := rows.Scan(&t.ID, &t.URL); err != nil {
			return nil, err
		}
		out = append(out, t)
	}
	return out, rows.Err()
}

// SaveDetail stores product-page fields. A model code from the spec table is
// authoritative and replaces the one parsed from the title.
func (s *Store) SaveDetail(ctx context.Context, id string, d listing.Detail, now time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE listings SET
		color=?, material=?, size=?, accessories=?, notes=?, detail_at=?,
		model_code = CASE WHEN ?<>'' THEN ? ELSE model_code END,
		group_key  = CASE WHEN ?<>'' THEN UPPER(?) ELSE group_key END
		WHERE id=?`,
		d.Color, d.Material, d.Size, d.Accessories, d.Notes, now.Unix(),
		d.ModelCode, d.ModelCode, d.ModelCode, d.ModelCode, id)
	return err
}

// MarkGone records that a product page no longer exists.
func (s *Store) MarkGone(ctx context.Context, id string, now time.Time) error {
	_, err := s.db.ExecContext(ctx, `UPDATE listings SET status='gone', gone_at=?, detail_at=? WHERE id=?`,
		now.Unix(), now.Unix(), id)
	return err
}

// AllRows loads every listing with its price history. The dataset is small
// (thousands of rows) so the radar indexes it in memory.
func (s *Store) AllRows(ctx context.Context) ([]Row, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, source, source_id, target, url, image_url, brand, title,
		model_name, model_code, category, gender, grade, rank, badges, price_jpy, was_price_jpy,
		color, material, size, accessories, notes, status, first_seen, last_seen, gone_at, detail_at
		FROM listings`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Row
	idx := map[string]int{}
	for rows.Next() {
		var r Row
		var rank, badges string
		var first, last int64
		var gone, detail sql.NullInt64
		if err := rows.Scan(&r.ID, &r.Source, &r.SourceID, &r.Target, &r.URL, &r.ImageURL, &r.Brand, &r.Title,
			&r.ModelName, &r.ModelCode, &r.Category, &r.Gender, &r.Grade, &rank, &badges, &r.PriceJPY, &r.WasPriceJPY,
			&r.Color, &r.Material, &r.Size, &r.Accessories, &r.Notes, &r.Status, &first, &last, &gone, &detail); err != nil {
			return nil, err
		}
		r.Rank = listing.Rank(rank)
		if badges != "" {
			r.Badges = strings.Split(badges, ",")
		}
		r.FirstSeen, r.LastSeen = time.Unix(first, 0), time.Unix(last, 0)
		if gone.Valid {
			r.GoneAt = time.Unix(gone.Int64, 0)
		}
		r.HasDetail = detail.Valid
		idx[r.ID] = len(out)
		out = append(out, r)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	pts, err := s.db.QueryContext(ctx, `SELECT listing_id, observed_at, price_jpy FROM price_points ORDER BY observed_at`)
	if err != nil {
		return nil, err
	}
	defer pts.Close()
	for pts.Next() {
		var id string
		var at, p int64
		if err := pts.Scan(&id, &at, &p); err != nil {
			return nil, err
		}
		if i, ok := idx[id]; ok {
			out[i].PriceHistory = append(out[i].PriceHistory, PricePoint{At: time.Unix(at, 0), PriceJPY: p})
		}
	}
	return out, pts.Err()
}

// Version changes whenever listing data changes; the radar uses it to decide
// when to rebuild its in-memory index.
func (s *Store) Version(ctx context.Context) (string, error) {
	var n, last, pts int64
	err := s.db.QueryRowContext(ctx, `SELECT COUNT(*), COALESCE(MAX(last_seen),0) +
		COALESCE(MAX(gone_at),0) + COALESCE(MAX(detail_at),0), (SELECT COUNT(*) FROM price_points) FROM listings`).
		Scan(&n, &last, &pts)
	return fmt.Sprintf("%d/%d/%d", n, last, pts), err
}

// Run is one sweep of one crawl target.
type Run struct {
	ID                                   int64
	Source, Target                       string
	StartedAt, FinishedAt                time.Time
	Pages, Seen, New, PriceChanges, Gone int
	Complete                             bool
	Error                                string
}

func (s *Store) StartRun(ctx context.Context, source, target string, at time.Time) (int64, error) {
	res, err := s.db.ExecContext(ctx, `INSERT INTO crawl_runs (source, target, started_at) VALUES (?,?,?)`,
		source, target, at.Unix())
	if err != nil {
		return 0, err
	}
	return res.LastInsertId()
}

func (s *Store) FinishRun(ctx context.Context, r Run) error {
	errText := r.Error
	_, err := s.db.ExecContext(ctx, `UPDATE crawl_runs SET finished_at=?, pages=?, seen=?, new_items=?,
		price_changes=?, gone=?, complete=?, error=? WHERE id=?`,
		r.FinishedAt.Unix(), r.Pages, r.Seen, r.New, r.PriceChanges, r.Gone, r.Complete, errText, r.ID)
	return err
}

func (s *Store) RecentRuns(ctx context.Context, limit int) ([]Run, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT id, source, target, started_at, COALESCE(finished_at,0), pages, seen,
		new_items, price_changes, gone, complete, error FROM crawl_runs ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var out []Run
	for rows.Next() {
		var r Run
		var st, fin int64
		if err := rows.Scan(&r.ID, &r.Source, &r.Target, &st, &fin, &r.Pages, &r.Seen, &r.New, &r.PriceChanges,
			&r.Gone, &r.Complete, &r.Error); err != nil {
			return nil, err
		}
		r.StartedAt = time.Unix(st, 0)
		if fin > 0 {
			r.FinishedAt = time.Unix(fin, 0)
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// Rate is an exchange rate quote.
type Rate struct {
	Base, Quote, Provider string
	Rate                  float64
	AsOf                  time.Time
}

func (s *Store) SaveRate(ctx context.Context, r Rate) error {
	_, err := s.db.ExecContext(ctx, `INSERT OR REPLACE INTO fx_rates (base, quote, rate, as_of, provider) VALUES (?,?,?,?,?)`,
		r.Base, r.Quote, r.Rate, r.AsOf.Unix(), r.Provider)
	return err
}

// LatestRate returns the most recent stored rate, or sql.ErrNoRows.
func (s *Store) LatestRate(ctx context.Context, base, quote string) (Rate, error) {
	r := Rate{Base: base, Quote: quote}
	var asOf int64
	err := s.db.QueryRowContext(ctx, `SELECT rate, as_of, provider FROM fx_rates WHERE base=? AND quote=?
		ORDER BY as_of DESC LIMIT 1`, base, quote).Scan(&r.Rate, &asOf, &r.Provider)
	r.AsOf = time.Unix(asOf, 0)
	return r, err
}
