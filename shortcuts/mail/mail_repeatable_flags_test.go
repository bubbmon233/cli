// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package mail

import (
	"reflect"
	"strings"
	"testing"

	"github.com/spf13/cobra"

	"github.com/larksuite/cli/internal/cmdutil"
	"github.com/larksuite/cli/shortcuts/common"
)

func TestNormalizeRepeatedRecipientFlagsPreservesDisplayNameCommas(t *testing.T) {
	got := normalizeRecipientFlagValues([]string{
		`Doe, Jane <jane@example.com>`,
		`"Roe, Richard" <richard@example.com>`,
		`alice@example.com,bob@example.com`,
	})

	if !strings.Contains(got, `"Doe, Jane" <jane@example.com>`) {
		t.Fatalf("normalized recipients should quote display-name comma, got %q", got)
	}
	boxes := ParseMailboxList(got)
	if len(boxes) != 4 {
		t.Fatalf("recipient count = %d, want 4; normalized=%q", len(boxes), got)
	}
	want := []Mailbox{
		{Name: "Doe, Jane", Email: "jane@example.com"},
		{Name: "Roe, Richard", Email: "richard@example.com"},
		{Email: "alice@example.com"},
		{Email: "bob@example.com"},
	}
	if !reflect.DeepEqual(boxes, want) {
		t.Fatalf("recipients = %#v, want %#v; normalized=%q", boxes, want, got)
	}
}

func TestNormalizeRepeatedCommaFlagsPreservesOrder(t *testing.T) {
	got := normalizeCommaListFlagValues([]string{
		"./a.pdf",
		" ./b.pdf,./c.pdf ",
		"",
		"./d.pdf",
	})
	want := []string{"./a.pdf", "./b.pdf", "./c.pdf", "./d.pdf"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("paths = %#v, want %#v", got, want)
	}
}

func TestNormalizeRepeatedInlineFlagsAppendsObjectAndArrayValues(t *testing.T) {
	raw, err := normalizeInlineFlagValues([]string{
		`{"cid":"hero","file_path":"./hero.png"}`,
		`[{"cid":"logo","file_path":"./logo.png"},{"cid":"qr","file_path":"./qr.png"}]`,
	})
	if err != nil {
		t.Fatalf("normalizeInlineFlagValues() error = %v", err)
	}
	specs, err := parseInlineSpecs(raw)
	if err != nil {
		t.Fatalf("parseInlineSpecs(%q) error = %v", raw, err)
	}
	want := []InlineSpec{
		{CID: "hero", FilePath: "./hero.png"},
		{CID: "logo", FilePath: "./logo.png"},
		{CID: "qr", FilePath: "./qr.png"},
	}
	if !reflect.DeepEqual(specs, want) {
		t.Fatalf("inline specs = %#v, want %#v", specs, want)
	}
}

func TestMailRepeatableFlagTypes(t *testing.T) {
	f, _, _, _ := mailShortcutTestFactory(t)
	for _, tc := range []struct {
		name         string
		shortcut     common.Shortcut
		stringArrays []string
		strings      []string
	}{
		{
			name:         "send",
			shortcut:     MailSend,
			stringArrays: []string{"to", "cc", "bcc", "attach", "inline"},
		},
		{
			name:         "draft-create",
			shortcut:     MailDraftCreate,
			stringArrays: []string{"to", "cc", "bcc", "attach", "inline"},
		},
		{
			name:         "reply",
			shortcut:     MailReply,
			stringArrays: []string{"to", "cc", "bcc", "attach", "inline"},
		},
		{
			name:         "reply-all",
			shortcut:     MailReplyAll,
			stringArrays: []string{"to", "cc", "bcc", "remove", "attach", "inline"},
		},
		{
			name:         "forward",
			shortcut:     MailForward,
			stringArrays: []string{"to", "cc", "bcc", "attach", "inline"},
		},
		{
			name:         "template-create",
			shortcut:     MailTemplateCreate,
			stringArrays: []string{"to", "cc", "bcc", "attach", "inline"},
		},
		{
			name:         "template-update",
			shortcut:     MailTemplateUpdate,
			stringArrays: []string{"attach", "inline"},
			strings:      []string{"set-to", "set-cc", "set-bcc"},
		},
	} {
		t.Run(tc.name, func(t *testing.T) {
			cmd := mountedShortcutCommand(t, f, tc.shortcut)
			for _, name := range tc.stringArrays {
				assertFlagType(t, cmd, name, "stringArray")
			}
			for _, name := range tc.strings {
				assertFlagType(t, cmd, name, "string")
			}
		})
	}
}

func mountedShortcutCommand(t *testing.T, f *cmdutil.Factory, shortcut common.Shortcut) *cobra.Command {
	t.Helper()
	parent := &cobra.Command{Use: "test"}
	shortcut.Mount(parent, f)
	for _, cmd := range parent.Commands() {
		if cmd.Name() == shortcut.Command {
			return cmd
		}
	}
	t.Fatalf("mounted command %q not found", shortcut.Command)
	return nil
}

func assertFlagType(t *testing.T, cmd *cobra.Command, name, want string) {
	t.Helper()
	flag := cmd.Flags().Lookup(name)
	if flag == nil {
		t.Fatalf("%s: flag %q not found", cmd.Name(), name)
	}
	if got := flag.Value.Type(); got != want {
		t.Fatalf("%s --%s type = %q, want %q", cmd.Name(), name, got, want)
	}
}
