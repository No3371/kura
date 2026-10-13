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
	"github.com/No3371/kura/internal/source/allu"
	"github.com/No3371/kura/internal/source/brandoff"
	"github.com/No3371/kura/internal/source/grare"
	"github.com/No3371/kura/internal/store"
	"github.com/No3371/kura/internal/web"
)

const defaultBrands = "HERMES,CHANEL,LOUIS VUITTON,CELINE,LOEWE,GOYARD"

type config struct {
	db, brands, addr, model    string
	sources                    string
	advisor, openaiModel       string
	openaiEffort               string
	delay, every, recheckAfter time.Duration
	maxPages, details          int
	fxFallback                 float64
	crawlOnStart, verbose      bool
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
	fs.StringVar(&c.sources, "sources", "brandoff,grare,allu", "comma-separated sources to crawl (brandoff, grare, allu)")
	fs.DurationVar(&c.delay, "delay", 3*time.Second, "minimum delay between requests to the source (robots.txt Crawl-delay wins if longer)")
	fs.IntVar(&c.maxPages, "max-pages", 0, "max list pages per brand per sweep (0 = all; capped sweeps never count unseen items toward sold)")
	fs.IntVar(&c.details, "details", 150, "max product pages to fetch per source per crawl for colour/size/accessories")
	fs.DurationVar(&c.recheckAfter, "recheck-after", 24*time.Hour, "least time between product-page re-checks of one listing, for brands whose sweep did not complete (0 = never; re-checks use what -details leaves)")
	fs.Float64Var(&c.fxFallback, "fx-fallback", 0.20, "JPY→TWD rate used until a live rate is fetched")
	fs.BoolVar(&c.verbose, "v", false, "debug logging")
	if cmd == "serve" {
		fs.StringVar(&c.addr, "addr", "127.0.0.1:8080", "listen address")
		fs.DurationVar(&c.every, "crawl-every", 0, "crawl on this interval while serving (e.g. 6h; 0 = never)")
		fs.BoolVar(&c.crawlOnStart, "crawl-on-start", false, "crawl once at startup")
		fs.StringVar(&c.model, "model", envOr("KURA_MODEL", "claude-opus-5-5"), "Claude model for the advisor")
		fs.StringVar(&c.advisor, "advisor", envOr("KURA_ADVISOR", "auto"), "advisor backend: auto | claude | openai | mock (auto: Claude if ANTHROPIC_API_KEY is set, else OpenAI if OPENAI_API_KEY is set, else mock)")
		fs.StringVar(&c.openaiModel, "openai-model", envOr("KURA_OPENAI_MODEL", "gpt-5.4"), "OpenAI model for the advisor (OPENAI_BASE_URL targets a compatible endpoint)")
		fs.StringVar(&c.openaiEffort, "openai-effort", envOr("KURA_OPENAI_EFFORT", ""), "OpenAI reasoning_effort (e.g. medium); empty = not sent, for models without reasoning")
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
	if n, err := st.BackfillGrades(ctx, listing.Schemes()...); err != nil {
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
	available := map[string]crawl.Source{
		brandoff.SourceName: &brandoff.Source{Fetcher: crawl.NewFetcher(c.delay), Brands: brands},
		grare.SourceName:    &grare.Source{Fetcher: crawl.NewFetcher(c.delay), Brands: brands},
		allu.SourceName:     &allu.Source{Fetcher: crawl.NewFetcher(c.delay), Brands: brands},
	}
	var sources []crawl.Source
	for n := range strings.SplitSeq(c.sources, ",") {
		src, ok := available[strings.TrimSpace(n)]
		if !ok {
			return nil, fmt.Errorf("unknown source %q", strings.TrimSpace(n))
		}
		sources = append(sources, src)
	}
	if err := checkSources(sources); err != nil {
		return nil, err
	}
	cr := &crawl.Crawler{Store: st, Sources: sources, Log: log, MaxPages: c.maxPages, DetailBudget: c.details,
		RecheckAfter: c.recheckAfter}
	return &deps{store: st, fx: conv, crawler: cr}, nil
}

// checkSources refuses a source list the crawler and radar can't keep apart.
func checkSources(sources []crawl.Source) error {
	names := map[string]bool{}
	for _, src := range sources {
		// Each shop's grades are shown in its own scale with its own
		// definitions; a source without a registered scheme would be
		// relabelled into another shop's or shown without one.
		if _, ok := listing.SchemeFor(src.Name()); !ok {
			return fmt.Errorf("source %q has no registered condition scheme", src.Name())
		}
		// Listing ids are "<source>:<id>": two adapters with one name would
		// overwrite each other's rows.
		if names[src.Name()] {
			return fmt.Errorf("duplicate source %q", src.Name())
		}
		names[src.Name()] = true
	}
	return nil
}

// crawlOnce refreshes FX then sweeps every source's brands.
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
	adv, err := newAdvisor(c, tools, log)
	if err != nil {
		return err
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

// newAdvisor picks the advisor backend. auto prefers Claude, then OpenAI,
// then the rule-based mock, by which API key is set.
func newAdvisor(c config, tools *advisor.Tools, log *slog.Logger) (advisor.Advisor, error) {
	mode := c.advisor
	if mode == "auto" {
		switch {
		case os.Getenv("ANTHROPIC_API_KEY") != "":
			mode = "claude"
		case os.Getenv("OPENAI_API_KEY") != "":
			mode = "openai"
		default:
			log.Warn("no ANTHROPIC_API_KEY or OPENAI_API_KEY: advisor runs in rule-based mock mode")
			mode = "mock"
		}
	}
	switch mode {
	case "claude":
		return advisor.NewClaude(anthropic.NewClient(), c.model, tools, log), nil
	case "openai":
		key := os.Getenv("OPENAI_API_KEY")
		if key == "" {
			return nil, errors.New("advisor openai needs OPENAI_API_KEY")
		}
		return advisor.NewOpenAI(key, os.Getenv("OPENAI_BASE_URL"), c.openaiModel, c.openaiEffort, tools, log), nil
	case "mock":
		return &advisor.Mock{Tools: tools}, nil
	}
	return nil, fmt.Errorf("unknown -advisor %q (want auto, claude, openai or mock)", mode)
}
