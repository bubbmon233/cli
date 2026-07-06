// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package service

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	larkcore "github.com/larksuite/oapi-sdk-go/v3/core"

	"github.com/larksuite/cli/internal/client"
	"github.com/larksuite/cli/internal/vfs/localfileio"
)

func serviceTransformTestResp(body string) *larkcore.ApiResp {
	return &larkcore.ApiResp{
		StatusCode: http.StatusOK,
		Header: http.Header{
			"Content-Type": []string{"application/json"},
		},
		RawBody: []byte(body),
	}
}

func decodeServiceTransformOutput(t *testing.T, out *bytes.Buffer) map[string]interface{} {
	t.Helper()
	var got map[string]interface{}
	if err := json.Unmarshal(out.Bytes(), &got); err != nil {
		t.Fatalf("invalid JSON output: %v\n%s", err, out.String())
	}
	return got
}

func TestServiceResponseTransformInjectsMailAttachmentDownloadURLHint(t *testing.T) {
	body := `{"code":0,"msg":"ok","data":{"download_urls":[{"attachment_id":"att_1","download_url":"https://example.com/a"}]}}`
	resp := serviceTransformTestResp(body)

	var out bytes.Buffer
	var errOut bytes.Buffer
	err := client.HandleResponse(resp, client.ResponseOptions{
		Out:       &out,
		ErrOut:    &errOut,
		FileIO:    &localfileio.LocalFileIO{},
		Transform: serviceResponseTransform(mailAttachmentDownloadURLSchemaPath),
	})
	if err != nil {
		t.Fatalf("HandleResponse failed: %v", err)
	}

	got := decodeServiceTransformOutput(t, &out)
	data, ok := got["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data = %T, want object; output: %s", got["data"], out.String())
	}
	if data["download_url_usage_hint"] != mailAttachmentDownloadURLUsageHint {
		t.Fatalf("download_url_usage_hint = %v, want fixed hint", data["download_url_usage_hint"])
	}
}

func TestServiceResponseTransformSkipsNonTargetSchemaPath(t *testing.T) {
	body := `{"code":0,"msg":"ok","data":{"download_urls":[{"attachment_id":"att_1","download_url":"https://example.com/a"}]}}`
	resp := serviceTransformTestResp(body)

	var out bytes.Buffer
	var errOut bytes.Buffer
	err := client.HandleResponse(resp, client.ResponseOptions{
		Out:       &out,
		ErrOut:    &errOut,
		FileIO:    &localfileio.LocalFileIO{},
		Transform: serviceResponseTransform("mail.user_mailbox.template.attachments.download_url"),
	})
	if err != nil {
		t.Fatalf("HandleResponse failed: %v", err)
	}

	got := decodeServiceTransformOutput(t, &out)
	data, ok := got["data"].(map[string]interface{})
	if !ok {
		t.Fatalf("data = %T, want object; output: %s", got["data"], out.String())
	}
	if _, exists := data["download_url_usage_hint"]; exists {
		t.Fatalf("non-target response got hint: %s", out.String())
	}
}

func TestServiceResponseTransformHintVisibleToJQ(t *testing.T) {
	body := `{"code":0,"msg":"ok","data":{"download_urls":[]}}`
	resp := serviceTransformTestResp(body)

	var out bytes.Buffer
	var errOut bytes.Buffer
	err := client.HandleResponse(resp, client.ResponseOptions{
		JqExpr:    ".data.download_url_usage_hint",
		Out:       &out,
		ErrOut:    &errOut,
		FileIO:    &localfileio.LocalFileIO{},
		Transform: serviceResponseTransform(mailAttachmentDownloadURLSchemaPath),
	})
	if err != nil {
		t.Fatalf("HandleResponse failed: %v", err)
	}
	if strings.TrimSpace(out.String()) != mailAttachmentDownloadURLUsageHint {
		t.Fatalf("jq output = %q, want hint", out.String())
	}
}

func TestServiceResponseTransformDoesNotInjectBusinessError(t *testing.T) {
	body := `{"code":99991400,"msg":"invalid token","data":{"download_urls":[]}}`
	resp := serviceTransformTestResp(body)

	var out bytes.Buffer
	var errOut bytes.Buffer
	err := client.HandleResponse(resp, client.ResponseOptions{
		Out:       &out,
		ErrOut:    &errOut,
		FileIO:    &localfileio.LocalFileIO{},
		Transform: serviceResponseTransform(mailAttachmentDownloadURLSchemaPath),
	})
	if err == nil {
		t.Fatal("expected business error")
	}
	if strings.Contains(out.String(), "download_url_usage_hint") {
		t.Fatalf("business error response emitted hint: %s", out.String())
	}
}

func TestInjectMailAttachmentDownloadURLHintPreservesServerValue(t *testing.T) {
	result := map[string]interface{}{
		"data": map[string]interface{}{
			"download_url_usage_hint": "server-provided",
		},
	}

	got := injectMailAttachmentDownloadURLHint(result)
	data := got.(map[string]interface{})["data"].(map[string]interface{})
	if data["download_url_usage_hint"] != "server-provided" {
		t.Fatalf("server-provided hint was overwritten: %v", data["download_url_usage_hint"])
	}
}
