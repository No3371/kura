# Patch: Kura Gallery Collector Audit Fixes (F1a, F2, F3)

> **Status:** Complete
> **Author:** Claude Opus 5.5 (patch, orchestrator subagent, depth 2)
> **Parent:** 2610140203-kura-gallery-collector-audit.md
> **Directive:** `/patch-projex.md` fix audit findings F1(a), F2, F3 of `2610140203-kura-gallery-collector-audit.md`. Out of scope: F1(b)/F4 backfill (pending human decision), F5 (radar comment; `internal/radar` + `internal/web` stay untouched), F6 (informational). No requests to any shop host; fixtures + tests only.
> **Source Plan:** Direct (audit findings; audited plan 2610140136-kura-gallery-collector-plan.md stays Complete). Rides the plan's close: committed on open ephemeral branch `projex/2610140136-kura-gallery-collector` (worktree `.projexwt/2610140136-kura-gallery-collector`), squashed with it.
> **Related Projex:** 2610140203-kura-gallery-collector-audit.md (findings source) | 2610140136-kura-gallery-collector-plan.md (audited plan; Out of Scope backfill note understated per F1) | 2610140136-kura-gallery-collector-log.md (audited execution) | 2610140046-kura-sold-observation-audit-fixes-patch.md (precedent)
> **Result:** Success
> **Rationale promoted:** F1 re-read gap → README § fetcher `-details` bullet (shipped doc) | F3 gallery-replaces-list-photo → `listing_images` comment in `schema.sql`

---

## Summary

Three fix-forward items from the audit, one commit `b16a1a1`:
- **F1(a)** README sentence implied old listings converge to a gallery via re-checks. Re-checks (`store.RecheckDue`) only reach active listings in incomplete targets that the crawl's list pages no longer show → reworded to state that gap.
- **F2** GALLERY RARE `TestParseDetail` asserted len + first + last; middle swap/duplicate mutants survived → exact 15-URL slice assertion.
- **F3** `listing_images` comment said "photos beyond listings.image_url" (false: gallery leads with main photo and replaces list photo on detail page); edited line ran 124 cols → reworded + rewrapped to ≤78.

Backfill itself (F1(b), F4) not implemented — human decision.

---

## Changes

### README re-read gap (F1a)

**File:** `README.md`
**Change Type:** Modified
**What Changed:**
- L88–90 (§ Choosing a source: respect the "no", fetcher `-details` bullet): "keep the list photo until a re-check reads it again." → "keep the list photo: Kura reads a page again only for a re-check, which skips listings the crawl's list pages still show and brands whose sweep completed, so most such listings never gain a gallery."

**Why:** Matches § Knowing when something sold (re-check = incomplete-sweep brand + listing no longer on list pages) and `RecheckDue` `last_seen < crawlStart`. Frames the pending backfill decision on the real mechanism.

---

### GALLERY RARE gallery assertion (F2)

**File:** `internal/source/grare/parse_test.go`
**Change Type:** Modified
**What Changed:**
- L4: import `fmt`.
- L174–180 (`TestParseDetail`): builds `wantImages` `-01`…`-15` (`?size=l&w=NjAw&t=20261010201553`, shared by all 15 fixture slides — checked by grep of `detail_bag.html`) and asserts `slices.Equal(d.Images, wantImages)`, same shape as BRAND OFF's test.

**Why:** Kills audit mutants M6b (swap 2↔3) and M6c (duplicate 2 over 3), which pass a count + endpoints check.

---

### `listing_images` comment (F3)

**File:** `internal/store/schema.sql`
**Change Type:** Modified
**What Changed:**
- L51–56: "Product photos beyond listings.image_url (the list-page photo), in the shop's order." → "A listing's product-page photo gallery, in the shop's order; when present the detail page shows it in place of listings.image_url (the list-page photo)." Remainder (writers, read-only service, AUTOINCREMENT/Version rationale) kept; block rewrapped to 74–78 cols.

**Why:** Source Hygiene rule 3 (describes what is). Claim checked against `detail.js` `photoHTML` (`it.images?.length ? it.images : … [it.image_url]`). Comment-only; schema `CREATE` statements unchanged, so embedded `schema.sql` (`store.go` `//go:embed`) behaviour unchanged.

---

## Verification

**Method:** `go vet ./...` + `go test -count=1 ./...` in worktree. `gofmt -l` on CR-stripped `parse_test.go` → empty. `git diff --stat` → only the 3 intended files (no CRLF churn). Diff line widths checked.

**Result:**
```
go vet ./...  → (no output)
ok  	github.com/No3371/kura/cmd/kura	1.109s
ok  	github.com/No3371/kura/internal/advisor	2.157s
ok  	github.com/No3371/kura/internal/crawl	1.840s
?   	github.com/No3371/kura/internal/fx	[no test files]
ok  	github.com/No3371/kura/internal/listing	0.347s
ok  	github.com/No3371/kura/internal/radar	0.665s
ok  	github.com/No3371/kura/internal/source/allu	1.375s
ok  	github.com/No3371/kura/internal/source/brandoff	1.086s
ok  	github.com/No3371/kura/internal/source/grare	0.729s
ok  	github.com/No3371/kura/internal/store	1.555s
ok  	github.com/No3371/kura/internal/web	2.040s

 README.md                           |  4 +++-
 internal/source/grare/parse_test.go | 11 +++++++----
 internal/store/schema.sql           | 10 ++++++----
```

**Status:** PASS

---

## Impact on Related Projex

| Document | Relationship | Update Made |
|----------|-------------|-------------|
| 2610140203-kura-gallery-collector-audit.md | Parent (F1a, F2, F3) | none — point-in-time record; this patch answers it. Condition "F1 (a): README sentence corrected" met. |
| 2610140136-kura-gallery-collector-plan.md | Audited plan; Out of Scope backfill note omits the not-shown condition (F1) | none — closed record; correction recorded here and in README |
| 2610140136-kura-gallery-collector-log.md | Audited execution | none |

---

## Notes

- Still open for orchestrator/human: F1(b) backfill question (scope: without it, galleries appear essentially only on listings first read after deploy) | F4 (folds into F1(b)) | F5 radar comment (blocked by radar-untouched requirement) | landing commit at close needs conventional subject + `Projex: 2610140136-kura-gallery-collector` trailer.
- If a backfill is later approved, the README sentence edited here ("most such listings never gain a gallery") must change with it.
- `new-projex.sh --projex-dir .projex/closed` created `.projex/closed/closed/…`; file moved up one level before commit. Pass `--projex-dir .projex` (or check script semantics) next time.
- Edit wrote `parse_test.go` with LF in a `core.autocrlf=true` checkout; git warns it will be CRLF on next touch — index content LF either way, diff unaffected.
