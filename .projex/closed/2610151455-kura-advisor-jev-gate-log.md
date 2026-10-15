# Execution Log: Kura advisor Jev gate

> **Status:** Complete (Steps 1–4, 6 executed; Steps 5, 7 deferred to follow-ups)
> **Started:** 20261016 (session start)
> **Repo Root:** S:/Repos/portfolio/kura
> **Plan File:** .projex/2610151455-kura-advisor-jev-gate-plan.md
> **Parent:** 2610151455-kura-advisor-jev-gate-plan.md
> **Base Branch:** master
> **Worktree Path:** kura/.projexwt/2610151455-kura-advisor-jev-gate
> **Walkthrough:** 2610151455-kura-advisor-jev-gate-walkthrough.md

Scope (orchestrator): Steps 1, 2, 3, 4, 6. Step 0 done before execution. Step 5 (gold set + `judge-eval`) and Step 7 (Railway rollout) out of scope → follow-ups. Methodology/footer disclosure copy (Step 6 bullet 3) out of scope (wording = Open Question).

## Pre-Check Results

```
REPO_ROOT=S:/Repos/portfolio/kura
BRANCH=master
PLAN_REL=.projex/2610151455-kura-advisor-jev-gate-plan.md

PASS  Plan is committed (38877e1 projex: Jev gate plan (ready) + spike evidence)
WARN  Working tree has 2 uncommitted change(s)
PRE-CHECK PASSED
```

WARN = two untracked PNGs at base root, not ours; untouched (worktree mode keeps base checkout as-is).

## Steps

### 20261016 - Init
**Action:** Plan status Ready → In Progress, committed on master (7f5c79c). `projex-worktree.sh` → branch `projex/2610151455-kura-advisor-jev-gate`, worktree `.projexwt/2610151455-kura-advisor-jev-gate`. Bootstrap: `go mod download` ✓. Baseline `go build ./... && go vet ./... && go test ./...` ✓ (all packages ok). `go test -race` available (CGO_ENABLED=1, msys64 gcc).
**Result:** Ready to execute.
**Status:** Success

### 20261016 - Step 1: `internal/judge` client
**Action:** New `internal/judge/judge.go`: `Client{Key, Model, BaseURL, HTTP}`, `New(key, model)`, `Ask(ctx, state, map[string]Question) (Result, error)` → `POST {BaseURL}/v1/systemone`, bearer key, no client-side timeout (ctx bounds it). `Question{Type, Instructions, Criteria any}` (Choice options + Noul true/false both under `criteria`). `Answer{Type, Choice, Noul *float64, Confidence *float64, Probabilities}`; `Result{Model, Answers, Usage{InputTokens, OutputTokens}}`. Non-2xx → `*APIError{Status, Type, Message}` decoded from `detail.error_type`/`detail.message` (string `detail` and non-JSON bodies still yield status), with `RateLimited()` (429) and `Auth()` (403/401). Decode errors wrapped, never `*APIError`; response without `answers` is an error. Body read capped at 1 MiB. New `internal/judge/judge_test.go` (httptest fake).
**Result:** `go vet` ✓; `go test -race -count=1 ./internal/judge` ✓. Tests: Choice/Noul decode (Noul has no confidence) + request shape (path, bearer, pinned model, `criteria` not `options`, state passthrough) | error table 403 auth / 429 object / 429 string detail / 502 HTML / 500 JSON | error text never holds the key | decode errors wrapped | ctx deadline honoured (50 ms vs 2 s fake).
**Status:** Success

### 20261016 - Step 2: Questions + state + policy (+ buyer-only prompt line)
**Action:** New `internal/advisor/gate.go` (policy half): `gateIntents` (q3 options) | `gateAbuse` (q2/q3 Noul wording) | `gateQuestions()` → `map[string]judge.Question` (intent Choice, options under `criteria`; Nouls with "In \`message\`: " + true + "?", `criteria{true,false}`) | `gateStateFor(message, recent, shops)` (rich3 state) | `gateBlockText` (= Claude refusal text) | `gateOffTopicText` (canned zh-TW: what the advisor does, owned bags out of scope, 2 example questions) | `GateAction` (pass/short_circuit/block/fail_open) | `Thresholds` + `DefaultThresholds{0.80, 0.50, 0.70}` | pure `decide(judge.Result, Thresholds) (verdict, error)`: block if any noul ≥ Block (BlockedBy = highest), else off_topic ≥ Intent → short_circuit, else pass; Flagged = any noul in [Flag, Block); a missing/nil abuse answer → error (caller fails open). `internal/advisor/prompt.go`: extracted `shopNames()` (dedup of `listing.Schemes()` shops, now shared by prompt + gate state); added one buyer-only line under 範圍與界線. New `internal/advisor/gate_test.go`.
**Result:** Verbatim check vs spike: copied spike `main.go` into scratch, dumped `questions("q3")` and `state("rich3", …)` (nil and non-nil recent), dumped `gateQuestions()`/`gateStateFor` from a temporary test (deleted), normalized `shops` to one placeholder → JSON **IDENTICAL** for questions and both states. `go vet` ✓; `go test -race -count=1 ./internal/advisor` ✓. Tests: 13-row policy table (both line edges for block/flag/intent, multi-class block naming, block beats off_topic, flagged alongside block/short_circuit, authenticity + low-confidence pass) | partial result → error, missing intent → pass | question shape (6 names, 7 options, no fraud) | pinned wording | shops from schemes deduped | prompt buyer-only line.
**Status:** Success

### 20261016 - Step 3: `advisor.Gated` decorator (speculative start, fail open, strikes/cooldown)
**Action:** `internal/advisor/gate.go` (decorator half): `GateMode` (`GateShadow`/`GateEnforce`, `String()`), `Gated{Inner, Judge, GateMode, Budget, Thresholds, MaxStrikes, Cooldown, Log; now; sessions sessionStore[gateMark]}`, `NewGated(inner, judge, mode, budget, log)`, `Mode()` = `<inner>+jev:<mode>`, `Unwrap()`, `BaseMode(Advisor)`. `Chat`: (1) `begin` reads strikes / cooldown (enforce only) under the session mutex, snapshots ≤3 earlier messages, records this one; cooldown → block without Jev or Inner; (2) Inner.Chat starts in a goroutine on a cancellable child ctx with a held emit; (3) Jev `Ask` under `context.WithTimeout(ctx, Budget)` (default 800 ms) with rich3 state (`shopNames()`) + q3 questions; (4) `decide` → verdict, any Jev/decode/partial error → `fail_open`; block adds a strike (both modes); enforce block/short_circuit → drop held events, cancel Inner, wait for it to return (rollback), then return `shopperErr(gateBlockText, cause)` or emit the canned off-topic text; otherwise release events and return Inner's error unchanged; (5) one `advisor gate` log line: mode, action, enforced, model, jev_ms, in_tokens, intent, intent_conf, flagged, p_<class> ×5, blocked_by + strikes (on block), or jev_err kind (timeout/canceled/auth/rate_limited/http_N/error) + detail on fail_open; never message text. `internal/web/server.go` overview reports `advisor.BaseMode(s.Advisor)` so the UI's `advisor === "mock"` demo banner survives the wrapper. Tests: `internal/advisor/gated_test.go` (httptest fake Jev, scripted/burst inner advisors, stalling Claude fake API), `internal/web/chat_test.go` (overview mode).
**Result:** `go vet` ✓. `go test -race -count=10 -run Gate ./internal/advisor` ✓; `go test -race -count=3 ./internal/advisor` ✓; `go test -race ./internal/web` ✓. Coverage vs Verification Plan: shadow vs enforce (3 verdicts in shadow unchanged; enforce block / short_circuit / low-conf off_topic passes) | fail open on 500, 403, 429, garbage, partial answers, timeout (50 ms budget, turn < 1 s) | blocked and short-circuited turns leave no Claude history (stalling Claude fake + `checkpoint`; next request carries only the next turn) | events never interleaved + in order (200-event burst, both modes; recorder fails on concurrent emit or emit after return) | 16 concurrent sessions mixed good/bad | cooldown after 3 strikes (4th turn: no Jev call, no Inner call; other session unaffected; clean after 10 min; one new strike does not re-trigger) | shadow never cools down but logs strikes | recent_user_messages (null first, capped at 3, per session) | log has labels/scores, not text | Mode/BaseMode | err kinds. First 10× run exposed `shops: null` when no scheme is registered (only in an isolated test run) → `gateStateFor` now sends `[]`.
**Status:** Success

### 20261016 - Step 4: Deterministic guards in `web.Server.chat`
**Action:** New `internal/web/limit.go`: `ChatLimiter` (map + mutex token bucket per client, no new dependency; `NewChatLimiter(perMinute, burst)`, `Allow(client)`; idle full buckets swept at most once a minute once ≥ 4096 clients) and `clientIP(r, trustProxy)` (last `X-Forwarded-For` hop when trusted, else `RemoteAddr` host; header ignored when not trusted). `internal/web/server.go`: `Server.ChatLimit *ChatLimiter` (nil = off) + `Server.TrustProxy bool`; `chat` checks the limiter first → 429 `{"error":"目前詢問的人較多，請稍後再試"}`; after decode, `maxChatRunes = 1000` → 400 `{"error":"訊息太長了，請精簡到 1000 字以內再試一次"}`; refusals go through new `chatError` (JSON + content type; the old 400 body was the same JSON shape sent as text/plain). `internal/web/static/js/advisor.js`: on a 429/400 the panel shows the server's `error` text instead of the generic network line (8 lines; `node --check` ✓). Tests in `internal/web/chat_test.go`.
**Result:** `go vet` ✓; `go test -race -count=1 ./internal/web` ✓. Tests: limiter 429 after burst 3 for one IP (new port, same bucket), advisor not called, other IP unaffected | refill at 6/min (refused at +9 s, allowed at +11 s, burst cap after idle) | sweep of idle clients | `clientIP` table (no-proxy ignores XFF, last hop wins, multiple headers, IPv6, blank header) | behind-proxy buckets per shopper, spoofed first hop stays in its bucket | rune cap: 1000 ✓, 1001 → 400 + advisor not called | no limiter when `ChatLimit` nil.
**Status:** Success

### 20261016 - Step 6: Wiring + docs (README + flags; disclosure copy out of scope)
**Action:** `cmd/kura/main.go`: serve flags `-judge` (`auto` default, env `KURA_JUDGE`), `-judge-model` (`jev-1.14.0`, env `KURA_JUDGE_MODEL`), `-judge-budget` (800ms), `-chat-rate` (6/min, 0 = off), `-chat-burst` (3), `-trust-proxy` (off). New `gateAdvisor(adv, c, key, log)`: auto → shadow when `TYPESAFE_API_KEY` set, else off; explicit shadow/enforce without key → startup error (`-judge enforce needs TYPESAFE_API_KEY`); non-positive budget / unknown mode → error; logs `chat gate` mode/model/budget (never the key); wraps with `advisor.NewGated(adv, judge.New(key, model), …)`. `runServe` builds `web.NewChatLimiter(rate, burst)` when rate > 0 and passes `ChatLimit` + `TrustProxy`. `cmd/kura/main_test.go`: `TestGateAdvisor` (9 cases; error text never holds the key; no Chat call, so no network). `README.md`: new "Chat gate" section (mermaid diagram, guards, Jev questions, policy lines, recall/precision rationale, buyer-only, speculative start, shadow + fail open, logs/analytics filter, data-handling line, tuning table from the 48-message spike set clearly marked as not the gold-set eval), package table, run example, flag table, API row, Tests bullet, Limitations bullet (gold set + `judge-eval` + spend cap pending). `railway.toml`: `-trust-proxy` appended to the start command + comment.
**Result:** `go vet ./cmd/kura` ✓; `go test -race -count=1 ./cmd/kura` ✓. Smoke: built binary, `serve -h` lists the 6 flags; `env -u TYPESAFE_API_KEY … serve -judge enforce` → `serve failed err="-judge enforce needs TYPESAFE_API_KEY"` (key env unset for the child, value never read). Binary + smoke db were in scratchpad, deleted.
**Status:** Success

### 20261016 - Post-execution: full verification
**Action:** Hardened the two tests that build a `judge.Client` without calling it (`TestGatedMode`, `TestOverviewReportsInnerAdvisorMode`) to an unroutable `BaseURL` (`http://127.0.0.1:1`), so a future edit cannot reach the real API by accident. `go build ./... && go vet ./...`; `go test -count=1 ./...`; `go test -race -count=1 ./...`. Source hygiene scan of added lines (`projex|spike|step N|plan|rich3|q3|2610`) in `.go/.js/.toml`.
**Result:** build ✓ vet ✓; `go test ./...` ✓ (11 packages ok, fx has no tests); `go test -race ./...` ✓ (all packages, gate included); earlier `-race -count=10 -run Gate ./internal/advisor` ✓. No real Jev call in any test: every Chat-calling test targets an httptest fake; `judge.New` (real base URL) is only built in `TestGateAdvisor`, which never calls Chat. Hygiene scan: 0 hits.
**Status:** Success

**Verification Plan coverage:**
- client decodes Choice/Noul, maps 429/5xx, respects ctx timeout → `internal/judge/judge_test.go` ✓
- policy table (every Step 2 row + line edges) → `TestGatePolicy`, `TestGatePolicyRejectsPartialResult` ✓
- Gated shadow vs enforce → `TestGatedShadowNeverChangesTheTurn`, `TestGatedEnforceBlocks`, `TestGatedEnforceShortCircuitsOffTopic` ✓
- fail open on error/timeout → `TestGatedFailsOpen` (500/403/429/garbage/partial/timeout) ✓
- blocked turn leaves no inner history (Claude fake + `checkpoint`) → `TestGatedOverrideLeavesNoClaudeHistory` (block + short_circuit) ✓
  - [PATCHED] Audit S1 found this held only for a stalling fake; renamed `TestGatedOverrideLeavesNoHistory` and extended to Claude/OpenAI full-reply fakes via 2610160323-kura-advisor-jev-gate-audit-fixes-patch.md
- events never interleaved, flushed in order → `TestGatedKeepsEventOrder`, `TestGatedConcurrentSessions`, recorder checks ✓
- cooldown after 3 strikes → `TestGatedCooldownAfterStrikes`, `TestGatedShadowNeverCoolsDown` ✓
- off-topic short-circuit never calls Inner → see Deviations (output discarded + history rolled back, tested) ✓ (reinterpreted)
- limiter 429 → `TestChatRateLimit`, `TestChatLimiterRefills`, `TestChatLimitBehindProxy`, `TestChatClientIP` ✓
- rune cap → `TestChatRuneCap` ✓
- `go test -race` on the gate ✓
- manual preview with real key + Railway p50/p95 → **not run** (user-owned real-API verification; Step 7)

**Success Criteria status:** log fields (intent, intent_conf, per-class abuse, action, model, latency; no raw text) ✓ | shadow never changes the turn ✓ | enforce block / short_circuit / no history ✓ | fail open ✓ | p50 added TTFT ≤ 150 ms on Railway → unverified (Step 7) | gold-set precision ≥ 0.95 + confusion matrix → not done (Step 5) | deterministic guards ✓.

## Deviations

- **Held first event instead of a free-running buffer (Step 3).** Plan: inner emit writes into a buffer until the verdict. Implemented: inner turn still starts speculatively, but its first `emit` blocks until the verdict, then events pass through in order (or are dropped). Why: with a buffer, a fast backend could finish and commit the turn to its history before a block verdict, breaking "never sends a blocked turn to the advisor history"; holding at the first event keeps the turn in flight so cancel → rollback. Latency effect identical for first text (both wait for max(Jev, TTFT)); tool execution can wait up to the verdict (≤ 800 ms). Residual: a backend that commits history without emitting any event before Jev answers is not covered (none of the three does in practice). Plan's Option A fallback not needed; `-race` clean. **Correction (audit S1):** "keeps the turn in flight so cancel → rollback" was false for a backend already holding the whole reply: a cancel does not stop it decoding buffered bytes, so `run` returned nil and the turn stayed in history. Fixed in 2610160323-kura-advisor-jev-gate-audit-fixes-patch.md (backends treat a cancelled turn as failed; gate cancels before releasing the held emit).
- **"Off-topic short-circuit never calls Inner" reinterpreted.** With speculative start (plan Step 3) Inner.Chat always starts; the test asserts its output never reaches the shopper and its history is rolled back. Inner is truly never called only during a cooldown (tested).
- **Field name `Gated.GateMode`, not `Mode`.** Go forbids a field and a method both named `Mode`; `Mode()` is the `Advisor` interface method.
- **`shops` order.** `listing.Schemes()` sorts by source id → `["ALLU","BRAND OFF","GALLERY RARE"]` vs spike `["BRAND OFF","GALLERY RARE","ALLU"]`. Accepted per orchestrator ("the one change is that shops comes from listing.Schemes()"). Otherwise state + questions JSON is identical to spike rich3/q3 (checked, Step 2). Empty shop list is sent as `[]` (spike never had one).
- **Overview reports `advisor.BaseMode`.** `Gated.Mode()` = `<inner>+jev:<mode>` per plan, but the UI shows its demo banner for `advisor === "mock"`, so `/api/overview` now reports the innermost mode. The `chat` log line keeps the full mode.
- **Small UI change.** `advisor.js` shows the server `error` text for 429/400 refusals (otherwise the plan's zh-TW 429/400 messages never reach the shopper; the panel would show the generic network line). Within "existing error text"; no new UI element.
- **Added flags.** `-chat-burst` and `-trust-proxy` (plan names `-chat-rate 0` only in Rollback; the XFF rule needs an explicit trust switch, since trusting XFF without a proxy lets any client pick its own bucket). `railway.toml` start command gains `-trust-proxy`; without it every Railway client would share the edge proxy's bucket (a global 6/min limit). `KURA_JUDGE` / `KURA_JUDGE_MODEL` env fallbacks added for the Step 7 "switch via a Railway variable".
- **Shadow mode and strikes.** Strikes are counted and logged in shadow, but cooldown acts only in enforce (shadow must not change what the shopper sees). After a cooldown expires the session strike count resets.
- **Explicit `-judge shadow` without key also errors** (plan specifies this only for enforce); `auto` without key → off.
- **README tuning table** uses the 48-message spike set (rich3 + q3), labelled as tuning, not the Step 5 gold-set eval; the README states the precision target is still unproven.

## Issues Encountered

- `go vet`: `round2` already declared in `tools.go` → gate helper renamed `score`. Field/method clash `Mode` → `GateMode`.
- 10× repeated run: `TestGatedSendsRecentMessages` failed when run in isolation (no schemes registered → `shops: null`) → `gateStateFor` sends `[]`.
- Bash heredocs containing apostrophes were mangled by the shell wrapper twice → switched to Edit / script files; no effect on output.
- `gofmt -l` lists pre-existing advisor files only because the worktree checks them out with CRLF (`core.autocrlf`); new files are gofmt-clean.

## Data Gathered

- Verbatim check: gate `gateQuestions()` and `gateStateFor()` JSON (nil + non-nil recent) == spike `questions("q3")` / `state("rich3", …)` with `shops` normalized.

## Follow-ups (out of this execution's scope)

- Step 5: `testdata/judge_gold.jsonl` (~150 rows) + `kura judge-eval`; pick final thresholds; publish confusion matrix in README.
- Step 7: deploy in shadow on Railway (needs `TYPESAFE_API_KEY` service variable, set by the user), ≥ 50 real turns, measure TTFT gate on/off, then `KURA_JUDGE=enforce`.
- Step 6 bullet 3: methodology/footer disclosure line (wording is an Open Question).
- Plan Open Questions remain open (enforce default, canned vs Claude redirect, raw-text logging, intent UI, ZDR, rate-limit numbers). Implemented defaults: canned off-topic reply, never log text, no intent SSE event, 6/min burst 3, 3 strikes → 10 min.

## User Interventions

None (orchestrator-dispatched; no mid-execution direction).
