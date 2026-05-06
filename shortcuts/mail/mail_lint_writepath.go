// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mail

import (
	"github.com/larksuite/cli/shortcuts/mail/lint"
)

// runWritePathLint is the single entrypoint compose 5 + +draft-edit body ops
// use to invoke the lint lib before writing to emlbuilder / draftpkg.Apply.
//
// The writing-path safety contract (spec §4.3) is:
//   - AutoFix is ALWAYS true (no `--no-lint` opt-out); errors are dropped
//     and warnings are auto-fixed in place.
//   - Strict is ALWAYS false; warnings never bump the exit code on the
//     write path (compare with `+lint-html --strict` which is a CI tool).
//   - The returned report is appended to the writing-path stdout envelope
//     under the contract keys `lint_applied` (warnings) and
//     `original_blocked` (errors); both arrays are always present (possibly
//     empty) so consumers can rely on `data.lint_applied[]` and
//     `data.original_blocked[]` unconditionally.
//   - When the body is plain-text, the lib short-circuits and returns an
//     EmptyReport; the cleaned HTML equals the input verbatim. Compose 5
//     callers are expected to gate the call on their existing useHTML
//     branch (S2 contract «N-way isomorphism» — diff template) so the
//     plain-text path doesn't pay the parse cost.
//
// Returns the cleaned HTML + the report. Callers MUST use the returned
// `cleaned` value as the body that goes to bld.HTMLBody / draftpkg.Apply
// (writing the original `body` would defeat the safety contract).
func runWritePathLint(body string) (cleaned string, rep lint.Report) {
	if body == "" {
		return "", lint.EmptyReport("")
	}
	rep = lint.Run(body, lint.Options{AutoFix: true, Strict: false})
	return rep.CleanedHTML, rep
}

// applyLintToEnvelope mutates the OutFormat data map by adding the contract
// keys `lint_applied` / `original_blocked`. The caller passes the existing
// envelope data map; the helper merges in the lint findings without
// disturbing other keys. Both keys are ALWAYS present in the merged map
// (spec §4.3 — even when no change was applied, the arrays must be
// rendered as `[]`).
func applyLintToEnvelope(data map[string]interface{}, rep lint.Report) {
	// rep.Applied / rep.Blocked are guaranteed non-nil by the lib; defensive
	// renormalisation here is cheap and ensures the envelope encoder writes
	// `[]` even if the report came from a partial mock.
	applied := rep.Applied
	if applied == nil {
		applied = []lint.Finding{}
	}
	blocked := rep.Blocked
	if blocked == nil {
		blocked = []lint.Finding{}
	}
	data["lint_applied"] = applied
	data["original_blocked"] = blocked
}

// emptyLintEnvelopeFields returns the writing-path stdout-envelope fields
// representing "no lint pass occurred" (e.g. plain-text body branch). Used by
// compose 5's plain-text path so the public envelope still carries the
// contract keys as empty arrays.
func emptyLintEnvelopeFields() (lintApplied, originalBlocked []lint.Finding) {
	return []lint.Finding{}, []lint.Finding{}
}

// lintFinding aliases the lint package's Finding type for callers that don't
// want to import shortcuts/mail/lint directly (e.g. function signatures in
// existing mail_*.go files that want to keep their import set minimal). It is
// purely a syntactic convenience — both names refer to the same struct.
type lintFinding = lint.Finding

// emptyLintFindings returns two non-nil empty Finding slices, used by helpers
// that initialise their outputs before knowing whether the body is HTML.
// Equivalent to emptyLintEnvelopeFields but named to reflect "findings" rather
// than "envelope fields" so call-sites read consistently with their context.
func emptyLintFindings() (applied, blocked []lint.Finding) {
	return []lint.Finding{}, []lint.Finding{}
}
