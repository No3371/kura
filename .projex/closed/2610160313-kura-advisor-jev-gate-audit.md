# Audit: Kura advisor Jev gate (Steps 1–4, 6)

> **Status:** Complete (Accept with Conditions)
> **Author:** Claude Opus 5.5 (Agent, audit subagent)
> **Parent:** 2610151455-kura-advisor-jev-gate-log.md
> **Related Projex:** 2610151455-kura-advisor-jev-gate-plan.md | 2610151455-kura-advisor-jev-gate-log.md | 2610160323-kura-advisor-jev-gate-audit-fixes-patch.md (S1 + minors fixed) | 2610151455-kura-advisor-jev-gate-walkthrough.md (close; S2/S3 carried as open follow-ups)

> **Audit Date:** 2026-10-16 | **Auditor:** Claude Opus 5.5 (read-only subagent) | **Work Period:** 2026-10-16, one session
> **Subject:** `master..projex/2610151455-kura-advisor-jev-gate` (6 commits, 0bace83..72111ca) | **Related:** plan + execution log above; spike reference `.projex/spike/2610151455-jev-gate/main.go` (`state`, `questions`)

---

## Audit Summary

**Claim:** Steps 1–4 and 6 are implemented: the `internal/judge` client; the gate policy with the spike's rich3/q3 wording; the `advisor.Gated` decorator (speculative start, fail open, strikes and cooldown); the per-IP limiter and 1,000-rune cap; flags and wiring; the README section; and the buyer-only prompt line. Every Verification Plan item that needs no real API is said to be tested. Steps 5 and 7 and the disclosure line are deferred.

**Verdict:** Partial. Nearly everything checks out. The exception is a success criterion that the log, README and code comments all call guaranteed: "never sends a blocked turn to the advisor history". It fails reproducibly when the inner Claude turn's stream is already buffered when the verdict arrives. The test suite cannot catch this, because its Claude fake always stalls.

**Assessment:** Completeness: High | Correctness: Medium | Quality: High | Value: High

**Top Issues:**
1. A blocked or short-circuited turn can still land in Claude/OpenAI history. Cancelling the context does not stop a backend from finishing a stream it has already buffered, and the backends roll back only when `run` returns an error. Reproduced 40 of 40 times.
2. Railway's client-IP source is unverified. `-trust-proxy` keys on the last `X-Forwarded-For` hop, but Railway community threads report CDN (Fastly) IPs in that slot, and staff recommend `X-Real-IP`. If that is right, shoppers share buckets, which amounts to a near-global 6/min limit on the public demo.
3. Strikes and cooldown are keyed on a session ID the client chooses (`crypto.randomUUID()` per page load). Any reload or a fresh ID escapes the cooldown, so only the per-IP limiter actually limits a determined abuser. This weakness is in the plan's design, not the execution, but neither the plan nor the log mentions it.

---

## Claims vs Evidence

| Claim | Evidence | Status | Notes |
|-------|----------|--------|-------|
| Step 1 `internal/judge` client (typed `APIError`, 429/403 split, ctx-bounded, `criteria` not `options`) | `internal/judge/judge.go`, `judge_test.go` | ✓ | Matches the plan's Step 1. 1 MiB read cap. FastAPI-style array `detail` (422) degrades to status only, which is acceptable |
| Step 2 wording is the spike's rich3 + q3 verbatim; `shops` from `listing.Schemes()` | Auditor re-ran it: exported HEAD + spike to scratch and marshalled `questions("q3")` / `state("rich3")` (nil and non-nil recent) against `gateQuestions()` / `gateStateFor()` with the spike's shop list | ✓ | All three JSON trees are **IDENTICAL**. In production `shops` is sorted by source id (`ALLU, BRAND OFF, GALLERY RARE`), as the orchestrator accepted |
| Thresholds 0.80 / 0.50 / 0.70; τ_intent acts only on off_topic | `internal/advisor/gate.go:130`, `decide` gate.go:145-175 | ✓ | `TestGatePolicy` covers the edges on both sides |
| Buyer-only line in the system prompt | `internal/advisor/prompt.go:22` | ✓ | One line, as asked, plus a shared `shopNames()` refactor |
| Speculative start + "held first event"; fail open; strikes/cooldown | gate.go:205-430 | ⚠ | Fail open, ordering, cooldown and shadow are all verified. The "held turn is still in flight, so rollback drops it" rationale is **false** for buffered streams (finding S1) |
| A blocked turn leaves no inner history | `TestGatedOverrideLeavesNoClaudeHistory` | ⚠ | This test passes only because `stallAPI` hangs until the client goes away. Auditor experiment (fake Claude sends the whole `textTurn` at once, Jev blocks after 50 ms, enforce): **40/40 blocked turns stayed in Claude history** |
| Per-IP limiter (6/min, burst 3) + 1,000-rune cap | `internal/web/limit.go`, server.go:505-522, `chat_test.go` | ✓ | Works as specified locally. The trust model on Railway is unverified (S2) |
| `-judge auto\|off\|shadow\|enforce`, `-judge-model jev-1.14.0`, `-judge-budget 800ms`, `TYPESAFE_API_KEY` | `cmd/kura/main.go` `gateAdvisor`, `TestGateAdvisor` | ✓ | Default is shadow when the key is set. The key is never logged or printed |
| README "Chat gate" section | README.md:230-305 | ⚠ | Accurate except README.md:274 (the rollback claim, S1) and README.md:248 (the XFF claim, S2) |
| No real Jev calls in tests | Searched every `judge.Client` built in a test | ✓ | Every Chat-calling test uses an httptest fake. Two constructors point at `127.0.0.1:1`. `TestGateAdvisor` builds `judge.New` but never calls Chat |
| `go test ./...` and `-race` pass | Auditor re-ran `go build`, `go vet`, `go test -count=1 ./...`, `go test -race -count=3` on advisor/judge/web/cmd/kura | ✓ | All ok. The worktree stayed clean |
| Each step committed with a `Projex:` trailer | `git log master..HEAD` | ✓ | 6/6 commits carry `Projex: 2610151455-kura-advisor-jev-gate`. Step 5 is absent, as scoped |

---

## Objective Verification

### Objective 1: Classify every turn and log it without the message text
**Evidence:** gate.go:411-428 and `TestGatedLogsLabelsNotText`.
**Findings:** The log line carries intent, intent_conf, flagged, p_<class> ×5, action, enforced, model, jev_ms and in_tokens, plus blocked_by/strikes or jev_err. The text is never logged. The `jev_err_detail` value comes from `error.Error()`, which holds a URL or API message but never the key. Quality: High. **Verification:** ✓

### Objective 2: Shadow mode never changes what the shopper sees
**Evidence:** `TestGatedShadowNeverChangesTheTurn` and `TestGatedShadowNeverCoolsDown`.
**Findings:** Content and order are unchanged. Latency is not: the first event is held for up to `Budget` in shadow too. The plan specifies this, and it keeps the Step 7 measurement representative. **Verification:** ✓

### Objective 3: Enforce blocks or short-circuits, and a blocked turn never enters advisor history
**Evidence:** gate.go:396-409, agent.go:86-89, openai.go:85-88, and the auditor's scratch experiment.
**Findings:** The shopper-facing half holds: dropped events, the refusal text, and the canned off-topic reply. The history half fails when the backend's SSE decoder already holds the rest of the turn when the held emit is released. Cancelling the context does not interrupt processing of bytes already buffered, so `run` returns nil and nothing is rolled back. Swapping the order to cancel-before-release did not help (still 40/40). Treating a cancelled context as failure in `Claude.Chat` fixed it (`if err == nil { err = ctx.Err() }` → 0/40 leaks, and 0/300 with the existing release-then-cancel order). **Verification:** ⚠ Partial

### Objective 4: Fail open on any Jev failure
**Evidence:** `TestGatedFailsOpen` covers 500, 403, 429, garbage, partial answers and timeout, with a 50 ms budget and the turn finishing in under 1 s. **Verification:** ✓

### Objective 5: Deterministic guards
**Evidence:** `TestChatRateLimit`, `TestChatLimiterRefills`, `TestChatLimitBehindProxy`, `TestChatClientIP`, `TestChatRuneCap`. **Verification:** ✓ locally. ⚠ The Railway hop is unverified (S2).

### Objectives not in this execution's scope
p50 TTFT ≤ 150 ms on Railway, gold-set precision ≥ 0.95, and the confusion matrix: deferred to Steps 5 and 7, and correctly marked unverified in the log.

---

## Code/Implementation Inspection

### `internal/advisor/gate.go`: `Gated.Chat`, `heldEmit`
**Claimed:** speculative start, held first event, cancel plus rollback on override.
**Actual:** Clean, small and well-commented. Quality: High, except for the rollback gap.
**Issues Found:**
- The rollback guarantee does not hold for buffered streams (gate.go:209-215 comment; gate.go:397-401). Severity: **High**. The fix belongs in the backends (agent.go:86, openai.go:85): a turn whose context was cancelled counts as failed. Also call `cancelInner()` before `held.decide(true)` so the release never runs on a live context.
- Strikes key on `sessionID`, which comes straight from the client (server.go `chat` → `req.Session`; advisor.js:39 `crypto.randomUUID()`). Severity: Medium (plan-level).
- In shadow mode, strike marks pile up without limit inside the cooldown window: every block appends one, and `begin` resets only once the cooldown since the *last* strike has passed. The 500-session LRU and the rate limit bound this. Severity: Low.
- A blocked message stays in `recent_user_messages` for the next 3 turns. That is defensible (it helps catch split injections) but undocumented. Severity: Low/informational.

**Undocumented:** none beyond what the log's Deviations list.

### `internal/web/limit.go`, `server.go` `chat`
**Issues Found:**
- `clientIP` trusts the rightmost XFF hop (limit.go:80-88; railway.toml:7). This is not verified against Railway. Community threads report CDN IPs there, and staff point to `X-Real-IP`. Severity: **Medium**. Nothing logs the keyed client, so the problem would also be hard to notice.
- IPv6 clients each get a bucket per /128, so one host can rotate addresses. Severity: Low.
- A body over 16 KB fails `MaxBytesReader` and returns 400 "session and message are required". advisor.js:207 now shows that English operator text to the shopper, for example after a very large paste. Severity: Low.

### `cmd/kura/main.go` `gateAdvisor`
Correct. `-judge shadow` without a key is now an error too, which is a reasonable tightening. `-judge-budget` and `-chat-*` have no env fallback, while `-judge` does. Severity: Low.

### `internal/advisor/prompt.go`
The `shopNames()` extraction preserves behaviour, and the buyer-only line is as asked.

---

## Testing Validation

**Execution:** All pass (`./...`; `-race -count=3` on advisor/judge/web/cmd). No flaky tests seen.
**Coverage of the Verification Plan:** Every automated item has a named test. "Off-topic short-circuit never calls Inner" is reinterpreted as "output discarded and history rolled back", which speculative start makes unavoidable. The reinterpretation is reasonable, but its history half has the same hole as S1.
**Missing:** A blocked turn whose inner backend has *already received the full reply*. Neither the stalling fake nor the scripted advisors (which have no history) exercise this.
**Quality issues:** `TestGateStateShopsFromSchemes` registers schemes `gx`/`gy` in the global `listing` registry and never cleans them up. That leaks into later tests in the package, such as prompt shops and gate state. Low.

---

## Documentation Audit

**Completeness:** The README gate section, flag table, API row, Tests bullet and Limitations bullet are complete. The disclosure/methodology line is deferred (Open Question).
**Accuracy:** Partial. README.md:274 states the rollback guarantee (S1). README.md:248 and :337 state the XFF trust model as fact (S2). README.md:304 says "the hosted demo runs in shadow mode", but that becomes true only once the key is set on Railway (Step 7). The tuning table matches the plan's Step 2 results: 48/48, 13/13, 0/35, ~1,340 tokens ≈ $0.00006.
**Quality:** High. It is honest that the precision target is still unproven.

---

## Source Hygiene

**Scope:** Diff only (`master..projex/2610151455-kura-advisor-jev-gate`, `.projex/**` excluded)

| Rule | Location (symbol) | Quote | Severity |
|------|-------------------|-------|----------|
| 6 (false rationale) | `advisor` → `Gated` doc comment | "Held at its first event the turn is still in flight when the gate cancels it, and the backend's own rollback drops it." | High (tied to S1: the comment states a guarantee that does not hold) |

No rule 1–5 violations: no projex IDs or paths, no `file:ln`, no changelog narration, no step banners. "deliberately no fraud option" names the rejected alternative and its consequence. `ChatLimiter` names the rejected `x/time/rate` dependency.

**Rule 6 gaps:**
- `Gated.Chat`: the order "release held emit, then cancel inner" has no comment, and it is the trap behind S1.
- `clientIP`: the "rightmost hop" choice names its alternative (trusting all hops) but not the Railway-specific assumption it depends on.

**Rule 6 vacuity:** none found.

**Rationale promotion:** Partial. The README carries the threshold rationale and the tuning table. The comment at gate.go:18-21 says wording changes need "the eval re-run", but the only eval artefacts (`messages.jsonl`, spike `main.go`) live in `.projex/spike/`, and `judge-eval` (Step 5) does not exist yet. Unpromoted: where the tuning set lives and how to re-run it. Low. Step 5 should ship it under `testdata/`.

**Commit composition** (informational; fix forward): 6 landed branch commits checked. Step subjects follow `projex: step N - …` 5/5. The completion commit uses `projex: complete …` and touches tests, with a trailer. `Projex:` trailer 6/6.
**Violations:** None.

**Trailer survival on `master`:** 2/30 recent commits carry a `Projex:` trailer (842cc0e, 8950a55). The rate is not zero, so the channel is not dead. Most recent code commits on master (`feat(advisor)`, `feat(web)` …) carry no trailer and appear to come from outside projex flows. Informational.

---

## Gap Analysis

### Promised But Not Delivered
| Promise | Status | Impact |
|---------|--------|--------|
| A blocked turn never reaches advisor history | Partial (buffered-stream leak) | Medium |
| Step 5 gold set + `judge-eval` | Deferred (scoped) | Medium (blocks enforce) |
| Step 7 Railway rollout + TTFT | Deferred (scoped, user-owned) | Medium |
| Methodology/footer disclosure line | Deferred (Open Question) | Low–Medium (a 4th vendor receives shopper text once the key is set) |

### Undocumented Issues
| Issue | Severity | Affects |
|-------|----------|---------|
| Buffered-stream history leak (S1) | High | Claude/OpenAI backends in enforce |
| Railway XFF hop unverified (S2) | Medium | Hosted demo rate limit |
| Session-keyed strikes are client-controlled (S3) | Medium | Abuse cooldown |
| >16 KB body shows English operator error to shopper | Low | Chat panel |
| Global scheme registration in test | Low | advisor tests |

### Unhandled Edge Cases
- An inner backend that commits history without emitting any event before the verdict. The log names this residual, and the S1 fix covers it too.
- An IPv6 client rotating addresses within its /64.

---

## Quality Assessment

**Completeness: High.** Every in-scope step was delivered, including thoughtful extras: `BaseMode` keeps the mock banner, `-trust-proxy` is opt-in, and the env fallback supports Step 7. **Gaps:** S1 and S2.
**Correctness: Medium.** Policy, fail open, ordering, cooldown and limiter are correct. The history invariant is broken (S1).
**Code Quality: High.** Small, readable, dependency-free; the race-clean design uses one mutex; the policy is pure. **Tech debt:** Low.
**Value Delivered: High.** Shadow-mode classification and logging plus deterministic cost guards are deployable today. Enforce mode needs S1 fixed and the Step 5 eval first.

---

## Open Findings

- **Downstream:** The `/api/overview` `advisor` field now reports the inner mode. The only consumer is advisor.js:62 (`=== "mock"`), so this is safe.
- **Risk:** If XFF's last hop is a Fastly edge on Railway, reviewers clicking through the demo will hit 429s together. Likelihood: Medium.
- **Could be better:** Log a truncated or hashed client key on 429, so the Railway trust model can be checked from logs during Step 7.

---

## Findings

### Critical (Must Address)
- None.

### Significant (Should Address)
- **S1: Overridden turns can persist in advisor history.** Locations: gate.go:396-401 (`Gated.Chat` enforced branch), agent.go:86-89 (`Claude.Chat`), openai.go:85-88 (`OpenAI.Chat`), with the claim repeated in gate.go:209-215 and README.md:274. Experiment: the fake Claude sends the whole reply at once, Jev blocks, enforce → 40/40 leaks. → In both backends, roll back when `ctx.Err() != nil` even if `run` returned nil. Call `cancelInner()` before `held.decide(true)`. Add a test with a non-stalling Claude fake. The auditor confirmed the backend change gives 0/40 and 0/300.
- **S2: The `-trust-proxy` hop choice is unverified on Railway.** Locations: limit.go:80-88 (`clientIP`), railway.toml:7, README.md:248, :337. → Before or during Step 7, log the derived key on a test request. Prefer `X-Real-IP` (Railway's documented overwrite) or make the header configurable.
- **S3: Strikes and cooldown can be escaped by rotating session ID.** Locations: gate.go:366 (`g.sessions.get(sessionID)`), advisor.js:39. → Key strikes on client IP (pass it through ctx), or document that the cooldown is only a courtesy and the limiter is the real control. Either way, note it in the plan's Risks.

### Minor (Nice to Fix)
- **Over-16 KB body shows "session and message are required" to the shopper** (server.go:516 together with advisor.js:207). → Map a `MaxBytesReader` failure to the zh-TW length message, or set `maxlength` on the textarea.
- **Test pollutes the global scheme registry** (gate_test.go:137-138). → Unregister in `t.Cleanup`, or use a unique fixture.
- **Shadow strike marks grow inside the window** (gate.go:323-334). → Cap the stored strikes at `MaxStrikes`.
- **README.md:304 present-tense "hosted demo runs in shadow mode"** → Change to "will run" until Step 7 sets the key.
- **No env fallback for `-judge-budget` and `-chat-rate`/`-chat-burst`** → Optional, for Railway tuning without a redeploy of `railway.toml`.
- **Tuning set lives only in `.projex/spike/`** → Promote it to `testdata/` with Step 5.

### Positive
- The verbatim wording requirement was met exactly, as the auditor's independent JSON comparison confirmed.
- Fail open is thorough (6 failure classes) and fast (the 50 ms budget is honoured).
- The `heldEmit` single-mutex design is race-clean under 200-event bursts and 16 concurrent sessions.
- The executor caught the `/api/overview` mock-banner regression and the "XFF without a proxy" spoofing trap unprompted.
- The README is honest about what the 48-row spike does and does not prove.

---

## Recommendations

**Immediate:** Fix S1 (backend ctx check + cancel ordering + non-stalling test) before anyone runs `-judge enforce`. Correct the README.md:274 and gate.go comment wording to match.
**Future:** Use Step 7 to verify the Railway client IP (S2). Decide S3 when the Open Question "Rate-limit numbers" is answered.
**Process:** Tests of "no history after cancel" should include a fake that has already delivered the full reply. A fake that stalls only proves the easy case.

---

## Final Verdict

**Status:** Accept with Conditions

**Overall Assessment:** Completeness: High | Correctness: Medium | Quality: High | Value: High

**Conditions:**
- [ ] S1 fixed in `Claude.Chat` and `OpenAI.Chat`, with a regression test that uses a non-stalling backend fake. The gate.go comment and README.md:274 must match the behaviour.
- [ ] S2 verified on Railway (or switched to `X-Real-IP`) before the rate limit is relied on publicly.
- [ ] S3 either addressed or recorded as an accepted risk in the plan.

**Sign-off:** Yes, for shadow-mode deployment as-is: S1 matters only in enforce mode, and shadow never overrides a turn. No for `-judge enforce` until S1 is fixed.
