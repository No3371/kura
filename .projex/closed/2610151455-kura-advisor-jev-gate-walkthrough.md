# Walkthrough: Kura advisor Jev gate — intent categorization + abuse identification

> **Execution Date:** 2026-10-16
> **Completed By:** Claude Opus 5.5 (executor, auditor, patch, closer — orchestrator subagents)
> **Source Plan:** 2610151455-kura-advisor-jev-gate-plan.md
> **Parent:** 2610151455-kura-advisor-jev-gate-log.md
> **Log:** 2610151455-kura-advisor-jev-gate-log.md
> **Related Projex:** 2610160313-kura-advisor-jev-gate-audit.md | 2610160323-kura-advisor-jev-gate-audit-fixes-patch.md
> **Duration:** start-execution 02:36 (7f5c79c); steps 1–6 02:40–03:02; audit 03:13; fix-forward patch 03:23–03:25; close same day
> **Result:** Partial Success — in-scope Steps 1–4 + 6 delivered and verified (5/7 success criteria ✓); Steps 5 + 7 and disclosure line deferred by scope; audit S2/S3 open for human

---

## Summary

Every shopper chat turn can now go through a TypeSafe Jev gate: one `POST /v1/systemone` with the spike-validated rich3 state + q3 questions returns a typed intent (Choice) and five abuse Nouls; a pure Go policy decides pass / short_circuit / block / fail_open; one `advisor gate` log line records labels + scores, never text. Gate = `advisor.Gated` decorator (speculative start, held first event, fail open, strikes/cooldown), so Claude/OpenAI/Mock all get it. Deterministic guards (per-IP token bucket 6/min burst 3, 1,000-rune cap) sit in `web.Server.chat` independent of Jev. Default `-judge auto` → shadow when `TYPESAFE_API_KEY` set, else off. Audit: Accept with Conditions; fix-forward patch closed S1 (overridden turn could persist in Claude/OpenAI history) + 4 minors. Enforce mode still waits on the Step 5 gold-set eval; Railway rollout (Step 7) is user-owned.

**Human prompts served:** the plan's Objective (abuse control + intent signal on the public, key-backed chat); owner decision "Kura is buyer-only" (2026-10-16) → q3 wording + one system-prompt line; close instruction "Ask the jev-gate session to close the projex and merge the branch into master."

---

## Objectives Completion

| Objective | Status | Notes |
|---|---|---|
| Log intent, intent_conf, per-class abuse, action, Jev model, latency; never raw text | Complete | `Gated.Chat` log line; `TestGatedLogsLabelsNotText` |
| Shadow never changes what the shopper sees | Complete | Content + order unchanged; first event held ≤ `Budget` (latency only) |
| Enforce: block ≥ τ_block, short-circuit off_topic ≥ τ_intent, blocked turn never in history | Complete (after patch) | History half failed for buffered replies (audit S1) → fixed in 2519587 |
| Fail open on any Jev error/timeout/missing key | Complete | 500/403/429/garbage/partial/timeout tested |
| p50 added TTFT ≤ 150 ms on Railway | Not verified | Step 7 (deferred, user-owned) |
| Gold-set precision ≥ 0.95 + confusion matrix in README | Not done | Step 5 (deferred) |
| Deterministic guards (per-IP limit + length cap) | Complete locally | Railway client-IP hop unverified (S2) |

---

## Execution Detail

> Derived from `git log master..projex/2610151455-kura-advisor-jev-gate`, the execution log, audit and patch. Deviations called out per step.

### Step 0: Spike

**Planned:** ~40 zh-TW messages through `jev-1.14.0`; exit to Option C if abuse classes don't separate.
**Actual:** Done before execution (user ran it from Taiwan; evidence committed with the plan in 38877e1 under `.projex/spike/2610151455-jev-gate/`). Passed: 48/48 intent and 13/13 abuse blocked at 0.80 with 0 clean ≥ 0.50 on rich3 + q3; p50 ~169–266 ms. Forced τ_block 0.80, τ_flag 0.50, `Budget` 800 ms.
**Deviation:** None (pre-execution).
**Files Changed:** none on this branch.

### Step 1: `internal/judge` client — 0bace83

**Planned:** `Client{Key, Model, BaseURL, HTTP}`, `Ask(ctx, state, qs)`, typed non-2xx error, no client timeout.
**Actual:** As planned. `*APIError{Status, Type, Message}` from `detail.error_type`/`detail.message` (string `detail` + non-JSON bodies still yield status); `RateLimited()` (429), `Auth()` (401/403); decode errors wrapped, never `*APIError`; missing `answers` → error; 1 MiB body cap.
**Deviation:** None.

| File | Change Type | Planned? | Details |
|---|---|---|---|
| `internal/judge/judge.go` | Created | Yes | 166 lines: client, request/response types, `APIError` |
| `internal/judge/judge_test.go` | Created | Yes | 159 lines: httptest fake; Choice/Noul decode, request shape (`criteria` not `options`), error table, key never in error text, ctx deadline |

**Verification:** `go test -race -count=1 ./internal/judge` ✓.

### Step 2: Questions + state + policy (+ buyer-only prompt line) — a6ed6f4

**Planned:** rich3 state + q3 questions verbatim; pure policy; buyer-only line in system prompt.
**Actual:** `gate.go` policy half: `gateIntents`, `gateAbuse`, `gateQuestions()`, `gateStateFor(message, recent, shops)`, `gateBlockText`, `gateOffTopicText`, `GateAction`, `Thresholds` + `DefaultThresholds{0.80, 0.50, 0.70}`, pure `decide`. `prompt.go`: `shopNames()` extracted (shared by prompt + gate state), buyer-only line under 範圍與界線.
**Deviation:** `shops` order comes from `listing.Schemes()` (sorted by source id: `ALLU, BRAND OFF, GALLERY RARE`) vs spike order — accepted by orchestrator. Otherwise JSON identical to spike (executor + auditor each checked independently).

| File | Change Type | Planned? | Details |
|---|---|---|---|
| `internal/advisor/gate.go` | Created | Yes | Policy half (+164) |
| `internal/advisor/gate_test.go` | Created | Yes | 13-row policy table, partial result → error, question shape, pinned wording, prompt line |
| `internal/advisor/prompt.go` | Modified | Yes | Line 22 buyer-only rule; `shopNames()` extracted (lines ~65–91) |

**Verification:** `go test -race -count=1 ./internal/advisor` ✓; verbatim JSON compare ✓.

### Step 3: `advisor.Gated` decorator — 5c4bc7e

**Planned:** speculative start, inner emit buffered until verdict, cancel + rollback on override, fail open, strikes/cooldown, one log line.
**Actual:** `GateMode` (`GateShadow`/`GateEnforce`), `Gated{Inner, Judge, GateMode, Budget, Thresholds, MaxStrikes, Cooldown, Log}`, `NewGated`, `Mode()` = `<inner>+jev:<mode>`, `Unwrap()`, `BaseMode(Advisor)`. Inner starts in a goroutine; its **first emit is held** until the verdict. Block adds a strike in both modes; cooldown acts only in enforce. `web.Server.overview` reports `advisor.BaseMode` so the mock banner survives the wrapper.
**Deviations:**
- Held first event instead of free-running buffer — a buffer would let a fast backend commit a blocked turn before the verdict. Original rationale ("still in flight → rollback drops it") proved false for buffered replies → audit S1 → fixed by patch (see Issues).
- "Off-topic short-circuit never calls Inner" reinterpreted: Inner always starts (speculative); its output is discarded and history rolled back. Inner truly not called only during cooldown.
- Field `Gated.GateMode`, not `Mode` (Go field/method name clash).

| File | Change Type | Planned? | Details |
|---|---|---|---|
| `internal/advisor/gate.go` | Modified | Yes | Decorator half (+292) |
| `internal/advisor/gated_test.go` | Created | No (plan named `gate_test.go`) | Decorator tests split out: shadow/enforce, fail open ×6, ordering (200-event burst), 16 concurrent sessions, cooldown, recent messages, log content |
| `internal/web/chat_test.go` | Created | No | `TestOverviewReportsInnerAdvisorMode` |
| `internal/web/server.go` | Modified | No | Overview uses `advisor.BaseMode` (line 385) |

**Verification:** `go test -race -count=10 -run Gate ./internal/advisor` ✓; `-race ./internal/web` ✓.

### Step 4: Deterministic guards — 63574b1

**Planned:** per-IP token bucket 6/min burst 3 → 429; 1,000-rune cap → 400; IP from last trusted XFF hop on Railway.
**Actual:** `ChatLimiter` (map + mutex, no new dep; idle sweep once ≥ 4096 clients), `clientIP(r, trustProxy)`. `Server.ChatLimit` (nil = off) + `Server.TrustProxy`; refusals through new `chatError` (JSON content type). `advisor.js` shows the server's `error` text on 429/400.
**Deviation:** Small UI change in `advisor.js` (otherwise zh-TW refusal text never reaches the shopper) — within "existing error text".

| File | Change Type | Planned? | Details |
|---|---|---|---|
| `internal/web/limit.go` | Created | No (plan put it in `server.go`) | 94 lines: `ChatLimiter`, `clientIP` |
| `internal/web/server.go` | Modified | Yes | `Server` fields (lines 47–52); `chatError` (492–503); limiter + rune cap in `chat` (506–528) |
| `internal/web/static/js/advisor.js` | Modified | No | Lines 181, 203–209: show server error text |
| `internal/web/chat_test.go` | Modified | Yes | Limiter 429, refill, sweep, `clientIP` table, behind-proxy buckets, rune cap |

**Verification:** `go test -race -count=1 ./internal/web` ✓.

### Step 5: Gold set + `kura judge-eval`

**Planned:** ~150-row `testdata/judge_gold.jsonl`, `judge-eval` subcommand, final thresholds, README confusion matrix.
**Actual:** Not executed — deferred by orchestrator scope.

### Step 6: Wiring + docs — d19ac9c

**Planned:** `-judge auto|off|shadow|enforce`, `-judge-model`, `-judge-budget`; README "Chat gate"; methodology disclosure line.
**Actual:** Flags `-judge` (env `KURA_JUDGE`), `-judge-model` (env `KURA_JUDGE_MODEL`), `-judge-budget`, `-chat-rate`, `-chat-burst`, `-trust-proxy`. `gateAdvisor` wraps the advisor; explicit shadow/enforce without key → startup error; key never logged. README "Chat gate" section (mermaid diagram, policy, rationale, tuning table from the 48-message spike set labelled as not the gold-set eval). `railway.toml` start command gains `-trust-proxy`.
**Deviations:** Added `-chat-burst`, `-trust-proxy`, env fallbacks for `-judge`/`-judge-model`; `-judge shadow` without key also errors. Disclosure line not done (Open Question wording).

| File | Change Type | Planned? | Details |
|---|---|---|---|
| `cmd/kura/main.go` | Modified | Yes | config fields (48–51), flags (81–86), limiter wiring in `runServe` (234–244), `gateAdvisor` (306–338) |
| `cmd/kura/main_test.go` | Modified | No | `TestGateAdvisor` (9 cases; no Chat call) |
| `README.md` | Modified | Yes | Diagram (51–53), Chat gate section (230–306), run example, flag table, API row, Tests, Limitations |
| `railway.toml` | Modified | No | Comment + `-trust-proxy` in `startCommand` |

**Verification:** `go test -race -count=1 ./cmd/kura` ✓; smoke `serve -judge enforce` without key → `-judge enforce needs TYPESAFE_API_KEY`.

### Step 7: Rollout

**Planned:** shadow on Railway → ≥ 50 turns → TTFT on/off → enforce.
**Actual:** Not executed — deferred, user-owned (needs `TYPESAFE_API_KEY` service variable).

### Post-execution — 72111ca

Hardened two tests that build a `judge.Client` without calling it to an unroutable `BaseURL` (`http://127.0.0.1:1`). Full `go test ./...` + `-race ./...` ✓. Source-hygiene scan of added lines: 0 hits.

### Audit fix-forward — 2519587 (patch 2610160323-kura-advisor-jev-gate-audit-fixes-patch.md)

| File | Change Type | Details |
|---|---|---|
| `internal/advisor/agent.go` | Modified | Lines 87–92: `if err == nil { err = ctx.Err() }` → cancelled turn rolls back |
| `internal/advisor/openai.go` | Modified | Lines 86–90: same |
| `internal/advisor/gate.go` | Modified | Enforced branch: `cancelInner()` before `held.decide(true)`; `Gated` doc states the real mechanism + residual; tuning comment no longer points at spike eval |
| `internal/advisor/gated_test.go` | Modified | `TestGatedOverrideLeavesNoHistory`: 3 backends (stalling Claude, full-reply Claude, full-reply OpenAI) × {block, short_circuit} |
| `internal/advisor/prompt.go`, `gate_test.go` | Modified | Pure `shopNamesOf`; test no longer mutates the global scheme registry |
| `internal/web/server.go`, `chat_test.go` | Modified | `*http.MaxBytesError` → zh-TW length line; gofmt of pre-existing `BaseCtx` field |
| `README.md` | Modified | Speculative-start bullet, "will run in shadow mode", Tests bullet |

---

## Complete Change Log

> **Derived from:** `git diff --stat master..projex/2610151455-kura-advisor-jev-gate` — 20 files, +2614 / −26 (incl. 4 projex docs).

### Files Created
| File | Purpose | Lines | In Plan? |
|---|---|---|---|
| `internal/judge/judge.go` | Jev HTTP client | 166 | Yes |
| `internal/judge/judge_test.go` | Client tests (httptest) | 159 | Yes |
| `internal/advisor/gate.go` | Questions, state, policy, `Gated` decorator | 463 | Yes |
| `internal/advisor/gate_test.go` | Policy + wording tests | 169 | Yes |
| `internal/advisor/gated_test.go` | Decorator tests | 604 | No (split from `gate_test.go`) |
| `internal/web/limit.go` | `ChatLimiter`, `clientIP` | 94 | No (plan: inside `server.go`) |
| `internal/web/chat_test.go` | Chat guard + overview tests | 224 | No (plan named web tests generally) |

### Files Modified
| File | Changes | Lines Affected | In Plan? |
|---|---|---|---|
| `README.md` | Chat gate section, flags, API, tests, limitations | +96/−3 (51–53, 230–306, 316, 334–338, 353, 371–375, 398–401) | Yes |
| `cmd/kura/main.go` | Gate + limiter flags, `gateAdvisor` | +53/−1 | Yes |
| `cmd/kura/main_test.go` | `TestGateAdvisor` | +41 | No |
| `internal/advisor/agent.go` | Cancelled turn rolls back | +6 (87–92) | No (patch) |
| `internal/advisor/openai.go` | Same | +5 (86–90) | No (patch) |
| `internal/advisor/prompt.go` | Buyer-only line; `shopNames`/`shopNamesOf` | +20/−10 | Yes |
| `internal/web/server.go` | Limiter, rune cap, `chatError`, BaseMode overview, MaxBytesError message | +40/−5 | Yes |
| `internal/web/static/js/advisor.js` | Show server refusal text | 181, 203–209 | No |
| `railway.toml` | `-trust-proxy` + comment | 1–9 | No |

### Files Deleted
None.

### Planned But Not Changed
| File | Planned Change | Why Not Done |
|---|---|---|
| `testdata/judge_gold.jsonl` | ~150 labelled zh-TW rows | Step 5 deferred |
| `cmd/kura` `judge-eval` subcommand | Eval report | Step 5 deferred |
| Methodology/footer copy | Third-party classifier disclosure line | Wording = Open Question |

---

## Success Criteria Verification

### 1. Log fields, no raw text
**Method:** `TestGatedLogsLabelsNotText`; audit read of the log line.
**Evidence:** log carries mode, action, enforced, model, jev_ms, in_tokens, intent, intent_conf, flagged, p_<class> ×5, blocked_by/strikes or jev_err kind/detail; message text never present.
**Result:** PASS

### 2. Shadow never changes what the shopper sees
**Method:** `TestGatedShadowNeverChangesTheTurn`, `TestGatedShadowNeverCoolsDown`.
**Evidence:** 3 verdicts in shadow → identical events; strikes logged, no cooldown.
**Result:** PASS (latency: first event held ≤ `Budget`, by design)

### 3. Enforce block / short-circuit; blocked turn never in advisor history
**Method:** `TestGatedEnforceBlocks`, `TestGatedEnforceShortCircuitsOffTopic`, `TestGatedOverrideLeavesNoHistory` (3 backends × 2 overrides).
**Evidence:** before patch, full-reply cases: `history after block = 2 messages, want 0` (×6), `history after short_circuit = 2 messages, want 0` (×6). After: `go test -race -count=30 -run TestGatedOverrideLeavesNoHistory ./internal/advisor/ ok` (180 overrides, 0 leaks).
**Result:** PASS (after fix-forward). Residual stated in `Gated` doc: an inner turn that returns without emitting anything can commit before the verdict; no real backend does this.

### 4. Fail open
**Method:** `TestGatedFailsOpen`.
**Evidence:** 500, 403, 429, garbage, partial answers, timeout (50 ms budget, turn < 1 s) → turn proceeds, `action=fail_open`.
**Result:** PASS

### 5. p50 added TTFT ≤ 150 ms on Railway
**Method:** Railway measurement (Step 7).
**Result:** NOT VERIFIED — deferred.

### 6. Gold-set precision ≥ 0.95 + confusion matrix
**Method:** `kura judge-eval` (Step 5).
**Result:** NOT DONE — deferred. README tuning table is the 48-message spike set, explicitly marked not the gold-set eval.

### 7. Deterministic guards
**Method:** `TestChatRateLimit`, `TestChatLimiterRefills`, `TestChatLimitBehindProxy`, `TestChatClientIP`, `TestChatRuneCap`.
**Evidence:** 429 after burst 3 per IP, other IP unaffected; refill at 6/min; 1001 runes → 400; 20,000-rune paste → zh-TW length line; advisor not called on refusal.
**Result:** PASS locally; Railway XFF hop unverified (S2).

### Acceptance Criteria Summary

| Criterion | Method | Result | Evidence |
|---|---|---|---|
| Log fields, no text | Unit test + audit | Pass | `TestGatedLogsLabelsNotText` |
| Shadow unchanged | Unit tests | Pass | `TestGatedShadowNeverChangesTheTurn` |
| Enforce + no history | Unit tests | Pass (post-patch) | `TestGatedOverrideLeavesNoHistory` |
| Fail open | Unit tests | Pass | `TestGatedFailsOpen` |
| TTFT ≤ 150 ms (Railway) | Railway run | Not verified | Step 7 |
| Gold-set precision + matrix | `judge-eval` | Not done | Step 5 |
| Deterministic guards | Unit tests | Pass (local) | `TestChat*` |

**Overall:** 5/7 pass; 2 deferred by scope (Steps 5, 7).

**Close-time re-run (at 4f3cf79, worktree):** `go vet ./...` ✓; `go test -count=1 ./...` → all 11 packages with tests ok (fx has none). Orchestrator + parent session also ran `go test -race` on advisor/judge/web/cmd/kura ✓.

---

## Deviations from Plan

1. **Held first event, not a buffer** — Reason: buffer lets a fast backend commit a blocked turn. Impact: same TTFT; tool execution may wait ≤ 800 ms. Rationale initially wrong for buffered replies → patched (backend ctx check + cancel-before-release). Recommendation: plan's Step 3 should describe the held-emit + backend-ctx-check design.
2. **Short-circuit "never calls Inner" → "output discarded + history rolled back"** — unavoidable with speculative start.
3. **`Gated.GateMode` field name** — Go name clash.
4. **`shops` order from `listing.Schemes()`** — accepted.
5. **`/api/overview` reports `BaseMode`** — keeps the mock demo banner.
6. **`advisor.js` shows server refusal text** — needed for zh-TW 429/400 text to reach the shopper.
7. **Extra flags** `-chat-burst`, `-trust-proxy`, env `KURA_JUDGE`/`KURA_JUDGE_MODEL`; explicit `-judge shadow` without key errors.
8. **Strikes in shadow** counted + logged, cooldown enforce-only; strikes reset after cooldown expires.
9. **Test/file split** — `gated_test.go`, `limit.go`, `chat_test.go` as separate files.

---

## Issues Encountered

### S1: overridden turn could persist in Claude/OpenAI history
- **Description:** cancelling ctx does not stop the stream decoder finishing an already-received reply → `run` returned nil → no rollback. Auditor: 40/40 leaks with a full-reply fake; existing test used a stalling fake only.
- **Severity:** High (enforce mode only; shadow never overrides).
- **Resolution:** 2519587 — backends treat a cancelled ctx as failure; gate cancels before releasing the held emit; 3-backend regression test.
- **Prevention:** "no history after cancel" tests need a fake that already delivered the full reply.

### Minor (fixed in patch)
- >16 KB body showed English "session and message are required" → zh-TW length line.
- `TestGateStateShopsFromSchemes` mutated global scheme registry → pure `shopNamesOf`.
- README present-tense "hosted demo runs in shadow mode" → "will run".
- `gate.go` comment pointed at an eval living only in `.projex/spike/` → reworded.

### Execution-time (from log)
- `round2` already declared in `tools.go` → gate helper `score`.
- Isolated test run sent `shops: null` → `gateStateFor` sends `[]`.
- Bash heredocs with apostrophes mangled by the shell wrapper → used Edit / script files.
- `gofmt -l` flags pre-existing CRLF checkouts only (`core.autocrlf`); new files clean.

### Close-time
- `close-precheck.ps1` / `.sh` returned `RESULT=ERROR` both ways: from the worktree plan path it resolved the log's `Worktree Path` (`kura/.projexwt/…`) against the repo root as `S:\Repos\portfolio\kura\kura/.projexwt/…`; from the base-checkout plan path it looked for `…-plan-log.md`. Advisory only; gates checked manually (base: no tracked changes, only the two unrelated untracked PNGs; worktree: fully clean).

---

## Key Insights

**Rationale promoted:** `README.md` "Chat gate" section (thresholds + recall/precision split, speculative start + cancel-before-release rollback, buyer-only scope, tuning table); `Gated` doc comment states the rollback mechanism and residual.

### Lessons Learned

1. **Cancel ≠ stop for buffered streams**
   - Context: audit S1.
   - Insight: a ctx cancel only interrupts blocking reads; a decoder holding the whole reply finishes and reports success.
   - Application: any "cancel → rollback" invariant must check `ctx.Err()` after the work returns, not rely on the work erroring.
2. **Stalling fakes prove only the easy case**
   - Insight: concurrency invariants need fakes at both extremes (stall + instant full reply).
3. **Spike wording as verbatim reference pays off**
   - Executor and auditor both diffed marshalled JSON against the spike — cheap, decisive check that tuned thresholds still apply.

### Pattern Discoveries

1. **Held-first-event decorator** — `Gated` holds the inner backend at its first emit until a side verdict arrives; one mutex serializes emit. Reuse: any pre-flight classifier/moderation wrapped around a streaming agent.
2. **`BaseMode` unwrap** — decorators that change `Mode()` need an unwrap path for UI checks keyed on the inner mode.

### Gotchas / Pitfalls

1. **Trusting `X-Forwarded-For`** — without a proxy any client picks its own bucket; behind Railway the right hop is unverified (S2). Avoidance: explicit `-trust-proxy`, verify keyed IP from logs before relying on it.
2. **Client-chosen session IDs** — strikes/cooldown keyed on `sessionID` are escapable by reload (S3); the per-IP limiter is the real control.

### Technical Insights

- Jev Noul answers carry no `confidence`; the policy works on `noul` alone, intent uses Choice `confidence`.
- `listing.Schemes()` is the single source for shop names in both prompt and gate state.

---

## Recommendations

### Immediate Follow-ups
- [ ] **Deploy note:** merging changes Railway behaviour — `railway.toml` now passes `-trust-proxy`, so the per-IP chat limiter (6/min, burst 3) is active on the hosted demo from the next deploy. Gate stays off until `TYPESAFE_API_KEY` is set.
- [ ] **S2 (open, human):** verify Railway client-IP source (last XFF hop vs `X-Real-IP`) by logging the keyed client on a test request; switch or make configurable if it is a Fastly edge IP — else shoppers share a near-global 6/min bucket.
- [ ] **S3 (open, human):** strikes/cooldown escapable by rotating session ID — key on client IP or record as accepted risk in plan Risks.
- [ ] Step 5: gold set + `kura judge-eval`; promote spike tuning set to `testdata/`; set final thresholds; README confusion matrix.
- [ ] Step 7: shadow on Railway (set `TYPESAFE_API_KEY`), ≥ 50 turns, TTFT gate on/off, then `KURA_JUDGE=enforce`.
- [ ] Methodology/footer disclosure line (third-party classifier; ZDR question).

### Future Considerations
- Env fallbacks for `-judge-budget`, `-chat-rate`, `-chat-burst` (Railway tuning without editing `railway.toml`).
- Cap shadow strike records at `MaxStrikes` (grow inside the cooldown window today; bounded by 500-session LRU + rate limit).
- IPv6 /128 buckets let one host rotate addresses; consider /64.
- Token-based spend cap (separate plan, per Out of Scope).
- Plan Open Questions still open: enforce default, canned vs Claude redirect, raw-text logging, intent UI, ZDR, rate-limit numbers.

### Plan Improvements
- Step 3 should specify held-first-event + backend `ctx.Err()` rollback, and a full-reply fake in Verification.
- Risks table should include client-controlled session IDs and the proxy trust model.

---

## Related Projex Updates

### Documents Updated
| Document | Update |
|---|---|
| 2610151455-kura-advisor-jev-gate-plan.md | `Completed` + `Walkthrough` fields; status already `Complete (…)`; → `closed/` |
| 2610151455-kura-advisor-jev-gate-log.md | `Walkthrough` field; status already terminal; → `closed/` |
| 2610160313-kura-advisor-jev-gate-audit.md | Related Projex links patch + walkthrough; status `Complete (Accept with Conditions)` → `closed/`. Open conditions S2/S3 carried here as follow-ups |
| 2610160323-kura-advisor-jev-gate-audit-fixes-patch.md | Born closed — no change |

Not touched: 2610140405-jev-embeddinggemma2-matching-proposal.md (U6 pre-rank still deferred, unrelated), 2610130449-kura-mock-def.md (spend-cap canvas gap not addressed here), 2610130513-kura-def-nav-vs-design-canvas-redteam.md, 2610130815-kura-advisor-panel-plan.md — none reference this plan. Plan names no Nav.

### New Projex Suggested
| Type | Description |
|---|---|
| Plan | Jev gate Step 5: gold set + `kura judge-eval` + threshold pick |
| Plan | Jev gate Step 7: Railway shadow rollout, S2 client-IP verification, TTFT measurement, enforce switch |
| Plan | Chat token spend cap (Out of Scope here) |

---

## Appendix

### References
- Branch commits: 0bace83 (step 1) | a6ed6f4 (step 2) | 5c4bc7e (step 3) | 63574b1 (step 4) | d19ac9c (step 6) | 72111ca (complete) | 2519587 (audit fix) | 07d8e5a (patch doc) | 4f3cf79 (audit doc)
- Master base: 7f5c79c (start execution), 38877e1 (plan + spike evidence)
- Spike evidence: `.projex/spike/2610151455-jev-gate/`
