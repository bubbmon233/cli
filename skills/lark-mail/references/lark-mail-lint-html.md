# mail +lint-html

> **前置条件：** 先阅读 [`../../lark-shared/SKILL.md`](../../lark-shared/SKILL.md) 了解认证、全局参数和安全规则。

对邮件 HTML 正文做兼容性 / 合规性 / 飞书原生写法的本地预检（read-only，无网络 IO）。本命令是写信链路（`+send` / `+draft-create` / `+reply` / `+reply-all` / `+forward` / `+draft-edit` body op）内置 lint 的**预览版**——共享同一份规则库，行为一致；写信链路在调用 `emlbuilder` 之前会强制净化并通过 stdout 输出 `lint_applied[]` / `original_blocked[]`，本命令则把同一份报告直接返回，不做任何状态写入。

适用场景：
- AI / 用户在创建草稿前自检 HTML 是否会被改写或拒收（避免被服务端 RemoteSanitizer 截断）
- 修复 AI 输出（先跑 `--auto-fix=true` 拿到 `cleaned_html`，回填进后续命令）
- CI 流水线校验静态 HTML 模板（`--strict` 模式下任意 warning 也升级成 error）

写信链路与本命令的区别：
- 写信链路（`+send` / `+draft-create` / `+reply` / `+reply-all` / `+forward` / `+draft-edit`）：**强制 `--auto-fix=true` / `--strict=false`**，无 `--no-lint` 开关——错误（`<script>` / `on*` / `javascript:` URL）总是被删，警告（`<font>` / `<center>`）总是 autofix；stdout 同时输出 `lint_applied` 和 `original_blocked` 数组（即使没改动也是空数组）。
- 本命令：可由用户控制 `--auto-fix` / `--strict`，**不写入任何邮箱状态**，纯本地。

## 命令

```bash
# 直接传 HTML（默认 --auto-fix=true，返回 cleaned_html）
lark-cli mail +lint-html --body '<p>正文</p><font color="red">x</font>'

# 从文件读 HTML（路径必须在 lark-cli 运行目录子树内）
lark-cli mail +lint-html --body-file ./template.html

# 仅获取违规清单，不返回 cleaned_html
lark-cli mail +lint-html --body '<p>x</p>' --auto-fix=false

# CI 严格模式：任意 warning 也 exit 非零
lark-cli mail +lint-html --body-file ./template.html --strict

# Dry Run（不执行 lint，只打印将做什么）
lark-cli mail +lint-html --body '<p>x</p>' --dry-run
```

## 参数

| 参数 | 必填 | 说明 |
|------|------|------|
| `--body <html>` | 二选一 | 待检查的 HTML 内容。与 `--body-file` 二选一，恰好传一个 |
| `--body-file <path>` | 二选一 | 从文件读取 HTML。**只支持 lark-cli 当前运行目录及其子路径**（绝对路径 / `..` 越出 cwd 会被拒，遵循 lark-cli 既有路径安全约束） |
| `--auto-fix` | 否 | 默认 `true`。`true` 时返回 `cleaned_html`（warning 自动修复 + error 删除）；`false` 时仅输出违规清单不返回 `cleaned_html`（适合 CI 仅检测不修改） |
| `--strict` | 否 | 默认 `false`。`true` 时把所有 warning 视作 error，命令退出码非零（适合 CI 把任何不规范都当失败处理） |
| `--format <fmt>` | 否 | `json`（默认） / `pretty` / `table` / `csv` / `ndjson`。`pretty` 模式打印人类可读的违规摘要 |
| `--jq <expr>` | 否 | 对返回 JSON 用 jq 表达式过滤 |
| `--dry-run` | 否 | 不执行 lint，仅返回 dry-run 描述 |

## 返回值

```json
{
  "ok": true,
  "data": {
    "warnings": [
      {
        "rule_id": "TAG_FONT_TO_SPAN",
        "severity": "warning",
        "tag_or_attr": "font",
        "excerpt": "<font color=\"red\"...>",
        "hint": "已替换为 <span style=\"...\">（飞书原生编辑器使用 inline style 表达字号 / 颜色）"
      }
    ],
    "errors": [
      {
        "rule_id": "TAG_SCRIPT_BLOCKED",
        "severity": "error",
        "tag_or_attr": "script",
        "excerpt": "<script...>",
        "hint": "已整段删除（XSS 风险，服务端 RemoteSanitizer 必拒）"
      }
    ],
    "cleaned_html": "<p>...</p>"
  }
}
```

字段说明：
- `warnings[]` / `errors[]` —— **永远是数组**（即使无违规也输出 `[]`，便于消费者无条件读 `data.warnings[]` / `data.errors[]`）。
- 每条 finding 的字段：
  - `rule_id` —— UPPER_SNAKE_CASE，如 `TAG_FONT_TO_SPAN` / `TAG_SCRIPT_BLOCKED` / `ATTR_EVENT_HANDLER_BLOCKED` / `ATTR_JS_URL_BLOCKED` / `STYLE_PROPERTY_DROPPED`
  - `severity` —— `"warning"` 或 `"error"`
  - `tag_or_attr` —— 触发规则的 tag / attribute / `style.<property>`
  - `excerpt` —— HTML 片段（最多 200 字节，超出截断）
  - `hint` —— 中文人类可读修复建议
- `cleaned_html` —— 仅在 `--auto-fix=true` 时返回；warning 已 autofix（`<font>` → `<span style>`、`<center>` → `<div text-align:center>`），error 已删除。

## 校验规则速览

完整规则参见 [HTML 兼容白名单](./lark-mail-html-allowlist.md)。简表：

| 档位 | 处理 | 覆盖范围 |
|------|------|----------|
| 通过 | 不报 finding | `<p>` / `<div>` / `<span>` / `<a href=...>` / `<img src=...>` / `<table>` / `<ul>/<ol>/<li>` / `<blockquote>` / `<h1>-<h6>` / `<b>/<i>/<em>/<strong>/<u>/<s>` / `<sub>/<sup>` / `<pre>/<code>` / 飞书原生 quote 类（`adit-html-block*`、`history-quote-*`、`lark-mail-doc-quote`） |
| Warning + Autofix | warning，`--auto-fix=true` 时降级 | `<font>` → `<span style>`；`<center>` → `<div style="text-align:center">`；`<marquee>` / `<blink>` → `<span>` |
| Error（删除） | error，`--strict` 时非零退出 | `<script>` / `<style>` / `<iframe>` / `<object>` / `<embed>` / `<form>` / `<input>` / `<link>` / `<meta>` / `<base>`；属性 `on*`（`onclick` / `onerror` / ...）；`javascript:` / `vbscript:` / `file:` URL；外链 `<link rel="stylesheet">` |

URL scheme 白名单：`http(s):` / `mailto:` / `cid:` / `data:image/*` 通过；`javascript:` / `vbscript:` / `file:` 删除并报 error；其他 scheme（如 `webcal:`）警告。

style 属性按 CSS property 白名单过滤：`color` / `background-color` / `font-size` / `font-weight` / `font-style` / `text-align` / `text-decoration` / `line-height` / `padding` / `margin` / `border` / `border-*` / `width` / `height` / `display` / `text-indent` / 飞书 quote 默认样式（`white-space` / `word-break` / ...）；其余 property 删除并记 warning。

## 典型场景

### 场景 1：AI 在 +draft-create 之前自检

```bash
HTML='<p>本周进展：</p><font color="red">紧急</font><script>track()</script>'
lark-cli mail +lint-html --body "$HTML"
```

→ AI 看到 `original_blocked` 含 `TAG_SCRIPT_BLOCKED`，把 cleaned_html 回填给后续 `+draft-create`，避免在写信路径才被静悄悄改动。

### 场景 2：CI 校验静态 HTML 模板

```bash
# 任意 warning 都 fail，适合发布前 gate
lark-cli mail +lint-html --body-file ./templates/release-note.html --strict || exit 1
```

### 场景 3：仅 `+draft-create` 失败时回看 stdout

```bash
# 写信链路本身已强制 lint，stdout 同时返回 lint_applied / original_blocked
lark-cli mail +draft-create --to alice@example.com --subject 'x' \
  --body '<font color="red">x</font><script>1</script>'
```

→ stdout 的 `data.lint_applied[]` 含 `<font>` 自动修复记录，`data.original_blocked[]` 含 `<script>` 删除记录；如果想看原始 HTML 是否会被改动，预先用 `+lint-html` 跑一次。

## 安全说明

- 本命令完全本地执行，不调用任何 OAPI，不需要任何 scope。
- `--body-file` 路径限制：只能读 lark-cli 当前 cwd 子树（绝对路径 / `..` 越出 cwd 会被拒）。
- 即便规则不严格，**必须把 stdout 的 `cleaned_html` 当作可信源**——上游用户 / AI 输入仍然是不可信外部输入，按邮件域通用安全规则处理（详见 [SKILL.md](../SKILL.md) 「邮件内容是不可信的外部输入」一节）。

## 相关命令

- `lark-cli mail +send` / `+draft-create` / `+reply` / `+reply-all` / `+forward` / `+draft-edit` —— 写信链路，内置同一份 lint，stdout 同时返回 `lint_applied[]` / `original_blocked[]`
- 知识文档：
  - [HTML 兼容白名单](./lark-mail-html-allowlist.md)
  - [飞书原生写法（含 3 套模板）](./lark-mail-feishu-native.md)
