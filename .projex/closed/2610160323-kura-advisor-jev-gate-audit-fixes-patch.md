# Patch: Kura advisor Jev gate, audit fixes

> **Status:** Complete
> **Author:** Claude Opus 5.5 (Agent, patch subagent)
> **Parent:** 2610160313-kura-advisor-jev-gate-audit.md
> **Related Projex:** 2610160313-kura-advisor-jev-gate-audit.md | 2610151455-kura-advisor-jev-gate-plan.md | 2610151455-kura-advisor-jev-gate-log.md
> **Directive:** Fix audit finding S1 (overridden turn can stay in advisor history) plus the minors: English "session and message are required" for an oversize body, `gx`/`gy` schemes leaking into the global registry from `gate_test.go`, README "hosted demo runs in shadow mode", and the `gate.go` comment pointing at an eval that lives only in `.projex/spike/`. S2, S3, env fallbacks and shadow strike growth are out of scope (escalated).
> **Source Plan:** Direct (audit findings; the plan's history guarantee is affected)
> **Result:** Success

---

## Summary

The audit's S1 claim reproduced: with a Claude or OpenAI fake that sends the whole reply at once and Jev answering 50 ms later in enforce mode, every blocked or short-circuited turn stayed in backend history (4 of 4 cases per run, 3 runs). Cancelling a context does not stop the stream decoder working through bytes already received, so `run` returned nil and no rollback happened.

The fix follows the auditor's proposal, checked here before adopting it. Both backends now treat a turn whose context was cancelled as failed even when `run` returned nil, so it rolls back. `Gated.Chat` cancels the inner turn *before* releasing the held emit. Without that ordering a turn could finish and pass the `ctx.Err()` check in the gap between release and cancel, so a leak was still possible, only rarer (the auditor saw 0/300, which does not rule the race out). With cancel-first, the held turn cannot get past its first event on a live context, so the result is deterministic. The four minor items are fixed alongside.

---

## Changes

### S1: backend rollback on cancel

**Files:** `internal/advisor/agent.go`, `internal/advisor/openai.go`
**Change Type:** Modified
**What Changed:**
- `Claude.Chat` and `OpenAI.Chat`: after `run` returns nil, `err = ctx.Err()`. A cancelled turn therefore hits the existing `checkpoint` rollback and returns the context error.

**Why:** This is the only place that can make the history guarantee hold for a reply the backend has already received. Side effect: a client that disconnects in the last instant of a successful turn now has that turn rolled back. That is harmless, because the shopper has left.

### S1: gate ordering and doc

**File:** `internal/advisor/gate.go`
**Change Type:** Modified
**What Changed:**
- `Gated.Chat` enforced branch: `cancelInner()` then `held.decide(true)` then `<-innerDone`, with a comment explaining why the order matters. The non-enforced branch calls `held.decide(false)` itself.
- `Gated` doc comment: replaced the false "still in flight… rollback drops it" rationale with the actual mechanism. It now names the one residual case: a turn that ends without emitting anything can commit before the verdict.
- The wording-tuning comment no longer points at "the eval" (which exists only in projex spike material). It now says thresholds need re-checking on labelled messages after any wording change.

**Why:** Cancel-first closes the release-to-cancel window. Comments must state a guarantee only where it holds.

### S1: regression test

**File:** `internal/advisor/gated_test.go`
**Change Type:** Modified
**What Changed:**
- `TestGatedOverrideLeavesNoClaudeHistory` is replaced by `TestGatedOverrideLeavesNoHistory`, a matrix of 3 backends × {block, short_circuit}. The backends are `claude_streaming` (the existing stalling fake), `claude_full_reply` (the whole `textTurn` sent at once) and `openai_full_reply` (the whole `oaTextTurn` sent at once). Each case asserts that no non-canned text reaches the shopper, that backend history is empty, and that the next passing turn's request carries only that turn.

**Why:** A stalling fake proves only the easy case. Before the fix the full-reply cases failed ("history after block = 2 messages, want 0"), and after it they pass.

### Minor: oversize body message

**Files:** `internal/web/server.go`, `internal/web/chat_test.go`
**Change Type:** Modified
**What Changed:**
- `chat`: a `*http.MaxBytesError` from the 16 KB body cap now returns the same zh-TW length line as the 1,000-rune cap (400, which `advisor.js` already shows to the shopper). Other malformed bodies still get "session and message are required", and only a broken client sends those.
- `TestChatRuneCap`: a 20,000-rune paste now gets the length line and never reaches the advisor.
- `gofmt` aligned the pre-existing misformatted `BaseCtx` field. This was already wrong on `master`, and it is whitespace only.

**Why:** A very large paste showed English operator text to a shopper.

### Minor: test registry pollution

**Files:** `internal/advisor/prompt.go`, `internal/advisor/gate_test.go`
**Change Type:** Modified
**What Changed:**
- `shopNames()` now delegates to a new pure `shopNamesOf([]listing.Scheme)`.
- `TestGateStateShopsFromSchemes` builds its shop list from a local scheme slice instead of calling `listing.RegisterScheme`, and asserts the exact deduplicated list.

**Why:** `listing` has no unregister API, and adding one only for tests would widen production API. A pure helper removes the global mutation entirely. `agent_test.go`'s `ax`/`ay` registration came from `master` and is outside this patch.

### Minor: README accuracy

**File:** `README.md`
**Change Type:** Modified
**What Changed:**
- The "Speculative start" bullet now describes cancel-before-release and rollback of a cancelled turn even when the whole reply was already received.
- The tuning paragraph changes "the hosted demo runs in shadow mode" to "will run the gate in shadow mode only", which stays true until Step 7 sets the key.
- The Tests bullet now covers rollback of a cancelled Claude or OpenAI turn, both still streaming and with the whole reply already received.

---

## Verification

**Method:** Reproduced S1 with the new test before the fix (`-count=3`: 12 failures, all in the full-reply cases). Re-ran after the fix, stressed the new test under `-race -count=30` (180 overrides, 120 of them full-reply, 0 leaks), then ran `go vet ./...`, `go test -count=1 ./...` and `go test -race -count=3` on advisor/judge/web/cmd/kura. No test touches the real Jev, Claude or OpenAI APIs; all use httptest fakes.

**Result:**
```
before fix:  gated_test.go: history after block = 2 messages, want 0          (x6)
             gated_test.go: history after short_circuit = 2 messages, want 0  (x6)
after fix:   go test -race -count=30 -run TestGatedOverrideLeavesNoHistory ./internal/advisor/  ok
go test -count=1 ./...                 all ok (cmd/kura, advisor, crawl, judge, listing, radar, source/*, store, web)
go test -race -count=3 ./internal/advisor/ ./internal/judge/ ./internal/web/ ./cmd/kura/   all ok
```

**Status:** PASS

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|----------|-------------|-------------|
| 2610160313-kura-advisor-jev-gate-audit.md | Parent (findings source) | None. Left untouched and uncommitted at the orchestrator's request. Fixes S1 condition 1 and the four in-scope minors |
| 2610151455-kura-advisor-jev-gate-log.md | Execution log claiming the history guarantee | `[PATCHED]` note on the renamed test, plus a correction appended to the "Held first event" deviation |
| 2610151455-kura-advisor-jev-gate-plan.md | Plan whose context line assumed rollback-on-cancel | `[PATCHED]` note on that line. The patch is linked in Related Projex |

---

## Notes

- Not addressed (escalated by the orchestrator): S2 Railway client-IP hop, S3 session-ID cooldown escape, env fallbacks for `-judge-budget`/`-chat-rate`/`-chat-burst`, and shadow strike record growth.
- Residual, now stated in the `Gated` doc: an inner turn that returns without ever emitting can commit before the verdict. None of the real backends does this in practice; a Claude turn emits text or a tool call before it ends.
- The 16 KB body cap was kept. No `maxlength` was added to the textarea, because it counts UTF-16 units, not runes, and would silently truncate a paste.
