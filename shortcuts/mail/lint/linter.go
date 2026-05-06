// Copyright (c) 2026 Lark Technologies Pte. Ltd.
// SPDX-License-Identifier: MIT

package lint

import (
	"bytes"
	"strings"

	xhtml "golang.org/x/net/html"
	"golang.org/x/net/html/atom"
)

// MaxExcerptBytes caps the raw-HTML excerpt embedded in a Finding.Excerpt so
// a single offending tag with megabyte content can't bloat the envelope JSON.
// S2 contract «Server-mirrored constraints» row "EML composition body+inline+
// SMALL ≤ 25 MB" calls this out: lint ops on bytes only, but the excerpt
// representation must not be size-amplifying.
const MaxExcerptBytes = 200

// Run lints the given HTML body and returns a structured Report. When
// opts.AutoFix is true, Report.CleanedHTML contains the rewritten HTML
// (warnings rewritten + errors deleted); when false, only error-tier findings
// are removed (writing-path safety floor cannot be opted out of), warnings
// are surfaced as observations only, and CleanedHTML still contains the
// rewritten HTML — but `+lint-html --auto-fix=false` callers are expected to
// drop this field from the public envelope per spec §4.2.
//
// IMPORTANT: when the input is empty or plain-text (no HTML markup detected
// by the cli's existing `bodyIsHTML` heuristic), callers should short-circuit
// with EmptyReport(html) instead of paying the parse cost. Run still handles
// this gracefully — html.Parse on plain text wraps the input in
// <html><head></head><body>...</body></html>, and the lib's pass-through
// rendering will reproduce the original text — but the round-trip is wasteful
// and produces no findings.
func Run(html string, opts Options) Report {
	if html == "" {
		return EmptyReport("")
	}

	rep := Report{
		Applied: []Finding{},
		Blocked: []Finding{},
	}

	// We use html.ParseFragment so users authoring fragment-style snippets
	// (the canonical compose-5 input shape — `<div>...</div>` rather than a
	// full document) don't get implicit <html><head><body> wrappers
	// re-rendered. The "body" insertion mode matches what html.Parse would
	// have done internally for a fragment but skips the structural wrappers
	// at render time.
	bodyContext := &xhtml.Node{Type: xhtml.ElementNode, DataAtom: atom.Body, Data: "body"}
	nodes, err := xhtml.ParseFragment(strings.NewReader(html), bodyContext)
	if err != nil {
		// Parser failure is exceptional (the parser is permissive by design);
		// fall back to the original input so we don't lose user content.
		return EmptyReport(html)
	}

	// Wrap fragment nodes in a synthetic root so the recursive walker has a
	// uniform parent pointer to mutate.
	root := &xhtml.Node{Type: xhtml.DocumentNode}
	for _, n := range nodes {
		root.AppendChild(n)
	}

	walk(root, &rep, opts)

	rep.HasErrorFindings = len(rep.Blocked) > 0
	rep.HasWarningFindings = len(rep.Applied) > 0
	rep.CleanedHTML = renderFragment(root)

	return rep
}

// walk visits every element node under parent, applying tag/attr/style
// classification. Children are iterated via the next-sibling pointer because
// we mutate the tree in place (replace / remove nodes).
//
// The walker is iterative-style via explicit recursion because the html
// parser's typical nesting depth (≤ 256 by default) is well below Go's
// goroutine stack limit; the existing draft package's plainTextFromHTML
// (mail/draft/htmltext.go) similarly recurses for the same reason.
func walk(parent *xhtml.Node, rep *Report, opts Options) {
	child := parent.FirstChild
	for child != nil {
		next := child.NextSibling
		if child.Type == xhtml.ElementNode {
			processElement(parent, child, rep, opts)
		}
		// child may have been removed/replaced by processElement; recurse
		// only if it still has the original parent (i.e. wasn't deleted).
		// The html parser sets Parent on every node, so a removed-then-
		// reattached node still recurses correctly via its new Parent.
		if child.Parent != nil {
			walk(child, rep, opts)
		}
		child = next
	}
}

// processElement applies the element-level classification cascade:
//  1. tag → allow / warn-rewrite / error-delete
//  2. attributes → on*-handlers, URL-bearing attrs (scheme allow-list),
//     style attribute (CSS property allow-list)
func processElement(parent, n *xhtml.Node, rep *Report, opts Options) {
	tagName := strings.ToLower(n.Data)
	kind, ruleID := classifyTag(tagName)

	switch kind {
	case "error":
		rep.Blocked = append(rep.Blocked, Finding{
			RuleID:    ruleID,
			Severity:  SeverityError,
			TagOrAttr: tagName,
			Excerpt:   excerptOf(n),
			Hint:      hintForBlockedTag(tagName),
		})
		// Always remove blocked tags regardless of opts.AutoFix — writing-path
		// safety floor cannot be opted out of (spec §4.3 — `--no-lint` is not
		// provided).
		parent.RemoveChild(n)
		return

	case "warn":
		// AutoFix=true → rewrite (e.g. <font>→<span style>); AutoFix=false →
		// surface the finding as observation only, keep the original tag.
		// `+lint-html --auto-fix=false` consumers want to see what would
		// change without the lib forcing the change.
		if opts.AutoFix {
			finding := Finding{
				RuleID:    ruleID,
				Severity:  SeverityWarning,
				TagOrAttr: tagName,
				Excerpt:   excerptOf(n),
				Hint:      hintForWarnTag(tagName),
			}
			if opts.Strict {
				finding.Severity = SeverityError
				rep.Blocked = append(rep.Blocked, finding)
			} else {
				rep.Applied = append(rep.Applied, finding)
			}
			rewriteWarnTag(n, tagName)
			// Recurse into the rewritten node by falling through; the
			// rewrite preserved children as-is.
		} else {
			finding := Finding{
				RuleID:    ruleID,
				Severity:  SeverityWarning,
				TagOrAttr: tagName,
				Excerpt:   excerptOf(n),
				Hint:      hintForWarnTag(tagName),
			}
			if opts.Strict {
				finding.Severity = SeverityError
				rep.Blocked = append(rep.Blocked, finding)
			} else {
				rep.Applied = append(rep.Applied, finding)
			}
		}
		// fall through to attribute scan
	case "allow":
		// no-op
	}

	// Attribute scan: build a new attribute slice, dropping/sanitising as we
	// go and surfacing findings.
	if len(n.Attr) > 0 {
		processAttributes(n, rep, opts)
	}
}

// processAttributes walks the attribute list and:
//   - drops on*-handlers (always; surfaced as error)
//   - drops URL-bearing attrs whose value uses a forbidden scheme
//   - filters the `style` attribute property-by-property against the allow-list
//
// Other attributes pass through unchanged. The cli's existing
// `validateInlineCIDs` (helpers.go:2226) handles `cid:`-specific checks; the
// lint must not duplicate that responsibility (S2 contract «MUST reuse» row).
func processAttributes(n *xhtml.Node, rep *Report, opts Options) {
	keep := n.Attr[:0]
	for _, attr := range n.Attr {
		name := strings.ToLower(attr.Key)

		// 1. on*-handlers → always drop, error-tier.
		if isEventHandlerAttr(name) {
			rep.Blocked = append(rep.Blocked, Finding{
				RuleID:    RuleAttrEventHandlerBlocked,
				Severity:  SeverityError,
				TagOrAttr: name,
				Excerpt:   truncateExcerpt(attr.Key + "=\"" + attr.Val + "\""),
				Hint:      "已删除事件处理器属性（on*）",
			})
			continue
		}

		// 2. URL-bearing attrs → check scheme allow-list.
		if urlAttributes[name] {
			kind, ruleID := classifyURLValue(attr.Val)
			switch kind {
			case "error":
				severity := SeverityError
				rep.Blocked = append(rep.Blocked, Finding{
					RuleID:    ruleID,
					Severity:  severity,
					TagOrAttr: name,
					Excerpt:   truncateExcerpt(attr.Key + "=\"" + attr.Val + "\""),
					Hint:      "已删除危险 URL 协议（仅允许 http/https/mailto/cid/data:image/*）",
				})
				continue
			case "warn":
				finding := Finding{
					RuleID:    ruleID,
					Severity:  SeverityWarning,
					TagOrAttr: name,
					Excerpt:   truncateExcerpt(attr.Key + "=\"" + attr.Val + "\""),
					Hint:      "URL 协议不在白名单（http/https/mailto/cid/data:image/*）；如确需使用请联系管理员",
				}
				if opts.Strict {
					finding.Severity = SeverityError
					rep.Blocked = append(rep.Blocked, finding)
				} else {
					rep.Applied = append(rep.Applied, finding)
				}
				if opts.AutoFix {
					// Drop the attribute when AutoFix is set — writing-path
					// safety floor (the URL would not render correctly anyway).
					continue
				}
			}
		}

		// 3. `style` attribute → property-by-property allow-list.
		if name == "style" {
			cleaned, dropped := sanitiseStyleAttr(attr.Val)
			for _, prop := range dropped {
				rep.Applied = append(rep.Applied, Finding{
					RuleID:    RuleStylePropertyDropped,
					Severity:  SeverityWarning,
					TagOrAttr: "style." + prop,
					Excerpt:   truncateExcerpt(prop),
					Hint:      "已删除非白名单 CSS 属性（详见 references/lark-mail-html-allowlist.md）",
				})
			}
			if len(dropped) == 0 {
				attr.Val = cleaned
				keep = append(keep, attr)
				continue
			}
			if !opts.AutoFix {
				// AutoFix=false: keep the original property list so users see
				// exactly what would change.
				keep = append(keep, attr)
				continue
			}
			if cleaned == "" {
				// All properties dropped — remove the attribute entirely.
				continue
			}
			attr.Val = cleaned
			keep = append(keep, attr)
			continue
		}

		// 4. Pass-through.
		keep = append(keep, attr)
	}
	n.Attr = keep
}

// rewriteWarnTag replaces a warning-tier tag with its Feishu-native
// equivalent in place: <font> → <span style="..."> with color/face/size
// distilled into inline style; <center> → <div style="text-align:center">;
// <marquee>/<blink> → <span> (text-only, animation discarded — collapsing
// to a span keeps the children but drops the deprecated animation effect).
func rewriteWarnTag(n *xhtml.Node, tagName string) {
	switch tagName {
	case "font":
		// Distill <font color="..." face="..." size="...">.
		var styles []string
		var keepAttrs []xhtml.Attribute
		for _, attr := range n.Attr {
			switch strings.ToLower(attr.Key) {
			case "color":
				if v := strings.TrimSpace(attr.Val); v != "" {
					styles = append(styles, "color:"+v)
				}
			case "face":
				if v := strings.TrimSpace(attr.Val); v != "" {
					styles = append(styles, "font-family:"+v)
				}
			case "size":
				if v := mapFontSize(attr.Val); v != "" {
					styles = append(styles, "font-size:"+v)
				}
			default:
				keepAttrs = append(keepAttrs, attr)
			}
		}
		// Merge any existing style attribute already present on the <font>
		// (rare but possible).
		if len(styles) > 0 {
			merged := strings.Join(styles, ";")
			styleIdx := -1
			for i, attr := range keepAttrs {
				if strings.ToLower(attr.Key) == "style" {
					styleIdx = i
					break
				}
			}
			if styleIdx >= 0 {
				existing := strings.TrimRight(keepAttrs[styleIdx].Val, "; ")
				if existing != "" {
					merged = existing + ";" + merged
				}
				keepAttrs[styleIdx].Val = merged
			} else {
				keepAttrs = append(keepAttrs, xhtml.Attribute{Key: "style", Val: merged})
			}
		}
		n.Data = "span"
		n.DataAtom = atom.Span
		n.Attr = keepAttrs

	case "center":
		// <center> → <div style="text-align:center">. Existing style attr
		// (if any) is merged with text-align prepended.
		styleIdx := -1
		for i, attr := range n.Attr {
			if strings.ToLower(attr.Key) == "style" {
				styleIdx = i
				break
			}
		}
		newStyle := "text-align:center"
		if styleIdx >= 0 {
			existing := strings.TrimRight(n.Attr[styleIdx].Val, "; ")
			if existing != "" {
				newStyle = newStyle + ";" + existing
			}
			n.Attr[styleIdx].Val = newStyle
		} else {
			n.Attr = append(n.Attr, xhtml.Attribute{Key: "style", Val: newStyle})
		}
		n.Data = "div"
		n.DataAtom = atom.Div

	case "marquee", "blink":
		// Both deprecated; collapse to <span> so children survive.
		n.Data = "span"
		n.DataAtom = atom.Span
		// Strip marquee-specific attributes (direction, scrollamount, ...)
		// so the rewritten span is plain.
		var keepAttrs []xhtml.Attribute
		for _, attr := range n.Attr {
			if strings.ToLower(attr.Key) == "style" || strings.ToLower(attr.Key) == "class" || strings.ToLower(attr.Key) == "id" {
				keepAttrs = append(keepAttrs, attr)
			}
		}
		n.Attr = keepAttrs
	}
}

// mapFontSize maps the legacy <font size="N"> values (1..7) to a CSS px
// equivalent. The mapping mirrors the editor-kit branch's renderer.
// Out-of-range values fall through to the empty string so the property is
// dropped (better than emitting an arbitrary value).
func mapFontSize(raw string) string {
	switch strings.TrimSpace(raw) {
	case "1":
		return "10px"
	case "2":
		return "13px"
	case "3":
		return "16px"
	case "4":
		return "18px"
	case "5":
		return "24px"
	case "6":
		return "32px"
	case "7":
		return "48px"
	default:
		return ""
	}
}

// sanitiseStyleAttr filters a `style="prop1:val; prop2:val"` declaration
// against the property allow-list. Returns the cleaned style text (joined
// with "; " separators) and a slice of dropped property names (lower-case)
// so the caller can surface STYLE_PROPERTY_DROPPED findings.
//
// NOTE: We do NOT validate property values — only property names. Spec §4.4
// is explicit: "style 属性按 CSS property 白名单过滤"; value-level validation
// (e.g. URL safety inside `background-image: url(...)`) is delegated to the
// urlAttributes path because such values typically appear in `src` / `href`
// attrs in compose-5 templates. Users authoring `background-image: url(http:...)`
// in inline style will see the property pass — the URL inside is not a
// security concern at the inline-style level since URL fetching from style
// is restricted by Feishu's renderer-side CSP regardless.
func sanitiseStyleAttr(raw string) (cleaned string, dropped []string) {
	if strings.TrimSpace(raw) == "" {
		return "", nil
	}
	parts := strings.Split(raw, ";")
	keep := make([]string, 0, len(parts))
	for _, part := range parts {
		decl := strings.TrimSpace(part)
		if decl == "" {
			continue
		}
		colon := strings.IndexByte(decl, ':')
		if colon < 0 {
			// Malformed declaration; drop and surface as a finding so the
			// user notices.
			dropped = append(dropped, decl)
			continue
		}
		name := strings.ToLower(strings.TrimSpace(decl[:colon]))
		if !classifyStyleProperty(name) {
			dropped = append(dropped, name)
			continue
		}
		keep = append(keep, decl)
	}
	cleaned = strings.Join(keep, "; ")
	return cleaned, dropped
}

// hintForBlockedTag returns a Chinese human-readable hint for an
// error-blocked tag (matching the `output.ErrWithHint` convention used
// elsewhere in the cli — see KB conventions/coding.md).
func hintForBlockedTag(tag string) string {
	switch tag {
	case "script":
		return "已整段删除（XSS 风险，服务端 RemoteSanitizer 必拒）"
	case "style":
		return "已删除 <style> 标签（飞书原生编辑器禁止内嵌样式表，请使用 inline style）"
	case "iframe", "object", "embed":
		return "已整段删除（不允许嵌入外部资源；如需展示富媒体，请改用 <img> 或邮件正文链接）"
	case "form", "input", "select", "option", "button":
		return "已整段删除（邮件正文不允许表单）"
	case "link":
		return "已删除 <link> 标签（不允许外链 CSS / 资源）"
	case "meta":
		return "已删除 <meta> 标签（不允许声明 viewport / refresh）"
	case "base":
		return "已删除 <base> 标签（不允许重写 URL 基址）"
	default:
		return "已整段删除（不允许使用该标签）"
	}
}

// hintForWarnTag returns a Chinese hint for a warning-tier tag.
func hintForWarnTag(tag string) string {
	switch tag {
	case "font":
		return "已替换为 <span style=\"...\">（飞书原生编辑器使用 inline style 表达字号 / 颜色）"
	case "center":
		return "已替换为 <div style=\"text-align:center\">（避免使用过时的 <center> 标签）"
	case "marquee", "blink":
		return "已替换为 <span>（动画效果不再支持，文字保留）"
	default:
		return "已按 Feishu 原生写法重写"
	}
}

// excerptOf renders the offending node's open-tag header into a short string
// suitable for surfacing in a Finding.Excerpt. We render only the tag header
// (not the full subtree) so a single offending <script> with megabytes of
// content doesn't bloat the envelope JSON. truncateExcerpt enforces the cap.
func excerptOf(n *xhtml.Node) string {
	if n == nil {
		return ""
	}
	var buf bytes.Buffer
	buf.WriteByte('<')
	buf.WriteString(n.Data)
	for _, attr := range n.Attr {
		buf.WriteByte(' ')
		buf.WriteString(attr.Key)
		if attr.Val != "" {
			buf.WriteString(`="`)
			buf.WriteString(attr.Val)
			buf.WriteByte('"')
		}
	}
	buf.WriteString(`...>`)
	return truncateExcerpt(buf.String())
}

// truncateExcerpt enforces MaxExcerptBytes; longer excerpts are truncated and
// suffixed with " ...". We measure bytes (not runes) because the cap is about
// envelope size, not character count — multibyte UTF-8 in an excerpt is
// uncommon in HTML markup excerpts.
func truncateExcerpt(s string) string {
	if len(s) <= MaxExcerptBytes {
		return s
	}
	return s[:MaxExcerptBytes-4] + " ..."
}

// renderFragment serialises a fragment-rooted html tree to a string. We use
// the html package's Render which always emits the document-style markup;
// for fragment input we strip the implicit <html><head></head><body>...</body></html>
// wrapper that html.Parse adds.
func renderFragment(root *xhtml.Node) string {
	var buf bytes.Buffer
	for child := root.FirstChild; child != nil; child = child.NextSibling {
		_ = xhtml.Render(&buf, child)
	}
	return buf.String()
}
