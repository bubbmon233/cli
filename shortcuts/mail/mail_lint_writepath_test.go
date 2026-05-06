// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mail

import (
	"encoding/json"
	"strings"
	"testing"

	"github.com/larksuite/cli/internal/httpmock"
	"github.com/larksuite/cli/shortcuts/mail/lint"
)

// jsonDecoderUnmarshal is a thin alias used by helpers in this file to keep
// the import set explicit even when the helper would otherwise be one-line.
func jsonDecoderUnmarshal(b []byte, v interface{}) error { return json.Unmarshal(b, v) }

// =====================================================================
// Writing-path lint integration tests — compose 5 + +draft-edit emit
// `lint_applied[]` and `original_blocked[]` arrays in the stdout envelope
// always (spec §4.3 contract).
// =====================================================================

// TestRunWritePathLint_PlainTextReturnsEmptyReport verifies the helper
// short-circuits on plain-text input.
func TestRunWritePathLint_PlainTextReturnsEmptyReport(t *testing.T) {
	cleaned, rep := runWritePathLint("")
	if cleaned != "" {
		t.Errorf("cleaned = %q, want empty", cleaned)
	}
	if rep.Applied == nil || rep.Blocked == nil {
		t.Error("Applied/Blocked must be non-nil")
	}
	if len(rep.Applied) != 0 || len(rep.Blocked) != 0 {
		t.Errorf("expected empty report, got applied=%d blocked=%d",
			len(rep.Applied), len(rep.Blocked))
	}
}

// TestRunWritePathLint_HTMLAlwaysAutofixesNeverStrict verifies the writing
// path uses {AutoFix: true, Strict: false} — strict warnings would block
// users on legitimate <font> tags, which spec §4.3 forbids.
func TestRunWritePathLint_HTMLAlwaysAutofixesNeverStrict(t *testing.T) {
	cleaned, rep := runWritePathLint(`<p><font color="red">x</font></p>`)
	if !strings.Contains(cleaned, "<span") {
		t.Errorf("expected autofix to rewrite <font>, cleaned=%q", cleaned)
	}
	if len(rep.Applied) != 1 {
		t.Errorf("expected 1 warning surfaced, got %d", len(rep.Applied))
	}
	// In strict mode the warning would be in Blocked instead. Confirm the
	// writing-path path does NOT promote.
	if len(rep.Blocked) != 0 {
		t.Errorf("writing-path must NOT use strict; expected 0 blocked, got %d", len(rep.Blocked))
	}
}

// TestApplyLintToEnvelope_AlwaysSetsContractKeys verifies the helper writes
// non-nil empty arrays even when the report has no findings.
func TestApplyLintToEnvelope_AlwaysSetsContractKeys(t *testing.T) {
	data := map[string]interface{}{"existing": "value"}
	rep := lint.EmptyReport(`<p>x</p>`)
	applyLintToEnvelope(data, rep)

	if data["existing"] != "value" {
		t.Error("existing key was clobbered")
	}
	la, ok := data["lint_applied"].([]lint.Finding)
	if !ok {
		t.Fatalf("lint_applied wrong type: %T", data["lint_applied"])
	}
	if la == nil {
		t.Error("lint_applied is nil — must be empty slice")
	}
	ob, ok := data["original_blocked"].([]lint.Finding)
	if !ok {
		t.Fatalf("original_blocked wrong type: %T", data["original_blocked"])
	}
	if ob == nil {
		t.Error("original_blocked is nil — must be empty slice")
	}
}

// =====================================================================
// End-to-end: +draft-create writing path emits envelope with lint fields.
// =====================================================================

// TestMailDraftCreate_WritePathLintEnvelope verifies +draft-create's stdout
// envelope always carries `lint_applied[]` / `original_blocked[]` arrays.
func TestMailDraftCreate_WritePathLintEnvelope(t *testing.T) {
	f, stdout, _, reg := mailShortcutTestFactory(t)
	chdirTemp(t)
	registerMailboxProfileMock(reg)
	registerDraftCreateOK(reg)

	err := runMountedMailShortcut(t, MailDraftCreate, []string{
		"+draft-create",
		"--to", "alice@example.com",
		"--subject", "Test",
		"--body", `<p>safe</p><script>alert(1)</script><font color="red">red</font>`,
	}, f, stdout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data := decodeShortcutEnvelopeData(t, stdout)

	// Both arrays must be present.
	la, ok := data["lint_applied"].([]interface{})
	if !ok {
		t.Fatalf("lint_applied missing or wrong type: %T", data["lint_applied"])
	}
	ob, ok := data["original_blocked"].([]interface{})
	if !ok {
		t.Fatalf("original_blocked missing or wrong type: %T", data["original_blocked"])
	}

	// We expect: 1 warning (<font>) + 1 error (<script>).
	if len(la) < 1 {
		t.Errorf("expected ≥1 lint_applied (warning for <font>), got %d", len(la))
	}
	if len(ob) < 1 {
		t.Errorf("expected ≥1 original_blocked (error for <script>), got %d", len(ob))
	}
}

// TestMailDraftCreate_PlainTextWritePathEnvelopesEmptyArrays verifies the
// envelope still carries empty arrays on the plain-text path.
func TestMailDraftCreate_PlainTextWritePathEnvelopesEmptyArrays(t *testing.T) {
	f, stdout, _, reg := mailShortcutTestFactory(t)
	chdirTemp(t)
	registerMailboxProfileMock(reg)
	registerDraftCreateOK(reg)

	err := runMountedMailShortcut(t, MailDraftCreate, []string{
		"+draft-create",
		"--to", "alice@example.com",
		"--subject", "Test",
		"--body", "plain text only",
		"--plain-text",
	}, f, stdout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	data := decodeShortcutEnvelopeData(t, stdout)
	la, ok := data["lint_applied"]
	if !ok {
		t.Fatal("lint_applied missing on plain-text path")
	}
	ob, ok := data["original_blocked"]
	if !ok {
		t.Fatal("original_blocked missing on plain-text path")
	}
	// Empty arrays are decoded as []interface{}{} (or could be nil; both
	// are valid JSON []).
	if las, _ := la.([]interface{}); len(las) != 0 {
		t.Errorf("lint_applied should be empty, got %d", len(las))
	}
	if obs, _ := ob.([]interface{}); len(obs) != 0 {
		t.Errorf("original_blocked should be empty, got %d", len(obs))
	}
}

// TestMailDraftCreate_AutofixApplied verifies that the writing path actually
// rewrites the body before sending it to drafts.create — the user's <font>
// tag must NOT reach the network as <font>.
func TestMailDraftCreate_AutofixApplied(t *testing.T) {
	f, stdout, _, reg := mailShortcutTestFactory(t)
	chdirTemp(t)
	registerMailboxProfileMock(reg)
	stub := &httpmock.Stub{
		Method: "POST",
		URL:    "/user_mailboxes/me/drafts",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{"draft_id": "d_test"},
		},
	}
	reg.Register(stub)

	err := runMountedMailShortcut(t, MailDraftCreate, []string{
		"+draft-create",
		"--to", "alice@example.com",
		"--subject", "Test",
		"--body", `<font color="red">x</font>`,
	}, f, stdout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Decode the raw EML and confirm <font> was rewritten before reaching
	// emlbuilder. The base64url payload contains the HTML body in raw form.
	captured := mustDecodeRawEMLFromStub(t, stub)
	if strings.Contains(captured, "<font") {
		t.Errorf("write-path should have rewritten <font>, EML still contains it: %q", captured)
	}
	if !strings.Contains(captured, "<span") {
		t.Errorf("expected <span> wrapper in EML, got %q", captured)
	}
}

// TestMailDraftCreate_ScriptStrippedBeforeSend verifies <script> is removed
// from the EML before drafts.create is invoked (writing-path safety floor).
func TestMailDraftCreate_ScriptStrippedBeforeSend(t *testing.T) {
	f, stdout, _, reg := mailShortcutTestFactory(t)
	chdirTemp(t)
	registerMailboxProfileMock(reg)
	stub := &httpmock.Stub{
		Method: "POST",
		URL:    "/user_mailboxes/me/drafts",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{"draft_id": "d_test"},
		},
	}
	reg.Register(stub)

	err := runMountedMailShortcut(t, MailDraftCreate, []string{
		"+draft-create",
		"--to", "alice@example.com",
		"--subject", "Test",
		"--body", `<p>before</p><script>alert(1)</script><p>after</p>`,
	}, f, stdout)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	eml := mustDecodeRawEMLFromStub(t, stub)
	if strings.Contains(eml, "<script") {
		t.Errorf("script should be stripped before EML send, got %q", eml)
	}
	if strings.Contains(eml, "alert(1)") {
		t.Errorf("script content should be removed, got %q", eml)
	}
	if !strings.Contains(eml, "before") || !strings.Contains(eml, "after") {
		t.Errorf("surrounding paragraphs should survive, got %q", eml)
	}
}

// =====================================================================
// Helpers — mail_shortcut_test.go ships the factory; these are local
// httpmock registrations specific to the lint integration tests.
// =====================================================================

// registerMailboxProfileMock registers a stock GET .../profile response so
// resolveComposeSenderEmail finds an address.
func registerMailboxProfileMock(reg *httpmock.Registry) {
	reg.Register(&httpmock.Stub{
		Method: "GET",
		URL:    "/user_mailboxes/me/profile",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"primary_email_address": "sender@example.com",
				"send_as":               []interface{}{},
			},
		},
	})
}

// registerDraftCreateOK registers a successful drafts.create response.
func registerDraftCreateOK(reg *httpmock.Registry) {
	reg.Register(&httpmock.Stub{
		Method: "POST",
		URL:    "/user_mailboxes/me/drafts",
		Body: map[string]interface{}{
			"code": 0,
			"data": map[string]interface{}{
				"draft_id": "d_test123",
			},
		},
	})
}

// mustDecodeRawEMLFromStub extracts the `raw` field from a captured body and
// base64url-decodes it. The stub.CapturedBody is populated by the httpmock
// after a match (registry.go:42 — the stub records every captured request).
func mustDecodeRawEMLFromStub(t *testing.T, stub *httpmock.Stub) string {
	t.Helper()
	if len(stub.CapturedBody) == 0 {
		t.Fatal("stub did not capture any request body")
	}
	var captured map[string]interface{}
	if err := jsonUnmarshal(stub.CapturedBody, &captured); err != nil {
		t.Fatalf("decode captured body: %v", err)
	}
	raw, ok := captured["raw"].(string)
	if !ok {
		t.Fatalf("captured body has no `raw` string field: %#v", captured)
	}
	return decodeBase64URL(raw)
}

func jsonUnmarshal(b []byte, v interface{}) error {
	return jsonDecoderUnmarshal(b, v)
}
