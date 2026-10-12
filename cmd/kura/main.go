// Command kura crawls Japanese pre-owned luxury listings, computes a price
// radar, and serves a web UI with a Claude-powered bag advisor.
//
//	kura crawl   # one crawl of the configured brands
//	kura serve   # web UI + API (+ optional scheduled crawling)
package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/anthropics/anthropic-sdk-go"

	"github.com/No3371/kura/internal/advisor"
	"github.com/No3371/kura/internal/crawl"
	"github.com/No3371/kura/internal/fx"
	"github.com/No3371/kura/internal/listing"
	"github.com/No3371/kura/internal/radar"
	"github.com/No3371/kura/internal/source/brandoff"
	"github.com/No3371/kura/internal/store"
	"github.com/No3371/kura/internal/web"
)

const defaultBrands = "HERMES,CHANEL,LOUIS VUITTON,CELINE,LOEWE,GOYARD"

type config struct {
	db, brands, addr, model string
	delay, every            time.Duration
	maxPages, details       int
	fxFallback              float64
	crawlOnStart, verbose   bool
}

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	cmd := os.Args[1]
	fs := flag.NewFlagSet(cmd, flag.ExitOnError)
	var c config
	fs.StringVar(&c.db, "db", "kura.db", "SQLite database path")
	fs.StringVar(&c.brands, "brands", defaultBrands, "comma-separated brands to crawl")
	fs.DurationVar(&c.delay, "delay", 3*time.Second, "minimum delay between requests to the source (robots.txt Crawl-delay wins if longer)")
	fs.IntVar(&c.maxPages, "max-pages", 0, "max list pages per brand per sweep (0 = all; capped sweeps never mark items sold)")
	fs.IntVar(&c.details, "details", 150, "max product pages to fetch per crawl for colour/size/accessories")
	fs.Float64Var(&c.fxFallback, "fx-fallback", 0.20, "JPY→TWD rate used until a live rate is fetched")
	fs.BoolVar(&c.verbose, "v", false, "debug logging")
	if cmd == "serve" {
		fs.StringVar(&c.addr, "addr", "127.0.0.1:8080", "listen address")
		fs.DurationVar(&c.every, "crawl-every", 0, "crawl on this interval while serving (e.g. 6h; 0 = never)")
		fs.BoolVar(&c.crawlOnStart, "crawl-on-start", false, "crawl once at startup")
		fs.StringVar(&c.model, "model", envOr("KURA_MODEL", "claude-opus-5-5"), "Claude model for the advisor")
	}
	fs.Parse(os.Args[2:])

	level := slog.LevelInfo
	if c.verbose {
		level = slog.LevelDebug
	}
	log := slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{Level: level}))
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	var err error
	switch cmd {
	case "crawl":
		err = runCrawl(ctx, c, log)
	case "serve":
		err = runServe(ctx, c, log)
	default:
		usage()
		os.Exit(2)
	}
	if err != nil && !errors.Is(err, context.Canceled) {
		log.Error(cmd+" failed", "err", err)
		os.Exit(1)
	}
}

func usage() {
	fmt.Fprintln(os.Stderr, "usage: kura <crawl|serve> [flags]   (kura <cmd> -h for flags)")
}

func envOr(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

type deps struct {
	store   *store.Store
	fx      *fx.Converter
	crawler *crawl.Crawler
}

func setup(ctx context.Context, c config, log *slog.Logger) (*deps, error) {
	st, err := store.Open(c.db)
	if err != nil {
		return nil, err
	}
	if n, err := st.BackfillGrades(ctx, brandoff.Scheme); err != nil {
		return nil, err
	} else if n > 0 {
		log.Info("backfilled source condition grades", "rows", n)
	}
	conv := &fx.Converter{Store: st, Fallback: c.fxFallback}
	if err := conv.Load(ctx); err != nil {
		return nil, err
	}
	var brands []string
	for b := range strings.SplitSeq(c.brands, ",") {
		br, ok := listing.BrandByName(strings.TrimSpace(b))
		if !ok {
			return nil, fmt.Errorf("unknown brand %q", b)
		}
		brands = append(brands, br.Name)
	}
	src := &brandoff.Source{Fetcher: crawl.NewFetcher(c.delay), Brands: brands}
	cr := &crawl.Crawler{Store: st, Source: src, Log: log, MaxPages: c.maxPages, DetailBudget: c.details}
	return &deps{store: st, fx: conv, crawler: cr}, nil
}

// crawlOnce refreshes FX then sweeps every brand.
func (d *deps) crawlOnce(log *slog.Logger) func(context.Context) error {
	return func(ctx context.Context) error {
		if err := d.fx.Refresh(ctx); err != nil {
			log.Warn("fx refresh failed; keeping previous rate", "err", err)
		}
		return d.crawler.RunAll(ctx)
	}
}

func runCrawl(ctx context.Context, c config, log *slog.Logger) error {
	d, err := setup(ctx, c, log)
	if err != nil {
		return err
	}
	defer d.store.Close()
	return d.crawlOnce(log)(ctx)
}

func runServe(ctx context.Context, c config, log *slog.Logger) error {
	d, err := setup(ctx, c, log)
	if err != nil {
		return err
	}
	defer d.store.Close()

	eng := &radar.Engine{Store: d.store, FX: d.fx}
	if err := eng.Refresh(ctx); err != nil {
		return err
	}
	tools := &advisor.Tools{Radar: eng, FX: d.fx}
	var adv advisor.Advisor = &advisor.Mock{Tools: tools}
	if os.Getenv("ANTHROPIC_API_KEY") != "" || os.Getenv("KURA_ADVISOR") == "claude" {
		adv = advisor.NewClaude(anthropic.NewClient(), c.model, tools, log)
	} else {
		log.Warn("ANTHROPIC_API_KEY not set: advisor runs in rule-based mock mode")
	}

	srv := &web.Server{Store: d.store, Radar: eng, FX: d.fx, Advisor: adv, Log: log,
		Crawl: d.crawlOnce(log), BaseCtx: ctx}
	if c.crawlOnStart {
		go srv.RunCrawl(ctx)
	}
	if c.every > 0 {
		go func() {
			t := time.NewTicker(c.every)
			defer t.Stop()
			for {
				select {
				case <-ctx.Done():
					return
				case <-t.C:
					srv.RunCrawl(ctx)
				}
			}
		}()
	}

	hs := &http.Server{Addr: c.addr, Handler: srv.Handler(), ReadHeaderTimeout: 10 * time.Second}
	go func() {
		<-ctx.Done()
		shutdown, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		hs.Shutdown(shutdown)
	}()
	log.Info("serving", "url", "http://"+c.addr, "advisor", adv.Mode())
	if err := hs.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
		return err
	}
	return nil
}
