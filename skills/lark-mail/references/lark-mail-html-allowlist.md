# 邮件 HTML 兼容白名单

> **前置条件：** 先阅读 [`../../lark-shared/SKILL.md`](../../lark-shared/SKILL.md) 了解通用安全规则。本文档定义 lark-cli mail 域写信场景下被允许的 HTML / CSS / URL 写法。

写信路径（`+send` / `+draft-create` / `+reply` / `+reply-all` / `+forward` / `+draft-edit` body op）以及 `+lint-html` 共享同一份规则库。本文档是规则的速查表——AI / 用户撰写邮件 HTML 时按本文档写就能避开所有 lint warning，确保飞书原生编辑器、收件端三方邮箱（Gmail / Outlook / Apple Mail）渲染一致。

完整命令用法见 [`+lint-html`](./lark-mail-lint-html.md)。

## 四档处理一览

| 档位 | 处理 | 覆盖范围 |
|------|------|----------|
| 通过 | 不报 warning，原样保留 | 通用块、原生强调、表格、图片、`<style>` 块（飞书 mail 服务端渲染时加 selector scope 隔离）等 |
| **飞书原生 autofix** | warning，写信路径默认补全 inline style + class + data-* marker | `<p>` → 飞书原生 div 双层嵌套；`<ul>/<ol>/<li>` → 飞书 native list-block 全套 marker；`<blockquote>` → 左侧 2px 灰边样式；`<a>` → `not-doclink` class + 飞书蓝色 |
| 警告 + 自动修复 | warning，`--auto-fix` 时降级；写信路径默认 autofix | `<font>` / `<center>` / `<marquee>` / `<blink>` HTML4 过时标签现代化 |
| 错误（删除） | error，`--strict` 非零退出；写信路径总是删 | `<script>` / `<iframe>` / `<form>` 等；`on*` 属性；`javascript:` URL |

**飞书原生 autofix 行为详细说明见 [`lark-mail-feishu-native.md`](./lark-mail-feishu-native.md) 第二节"飞书原生写法铁律"**。AI 写最简陋的 HTML，lint 自动补全成飞书 mail-editor 真原生格式。

## 标签白名单

### 通用块

| 写法 | 说明 |
|------|------|
| `<p>...</p>` | 段落，**lint autofix → 飞书原生双层 div 嵌套**：`<div style="margin:4px 0;line-height:1.6"><div dir="auto" style="font-size:14px">...</div></div>`。不需要手写 `<p><br></p>` 制造空行 |
| `<div>...</div>` | 块容器，可设 `style="line-height:1.6"` 等内联样式 |
| `<span>...</span>` | 行内容器，常用承载 `style="color:..."` / `font-size:...` |
| `<br>` | 单行换行；`<br>` 之间不需要 `</br>` |
| `<hr>` | 水平分割线 |
| `<blockquote>...</blockquote>` | **lint autofix → 飞书原生引用块**：左侧 2px 灰边 + 灰文字。reply / forward 自动生成的引用区由 CLI 处理，AI 不要重写 |
| `<pre>...</pre>` / `<code>...</code>` | 代码块（`<pre>`）/ 行内代码（`<code>`） |
| `<style>` 块 | 透传（lint 不感知）；飞书 mail 服务端渲染时自动加 selector scope class 前缀做隔离 |

### 强调

| 写法 | 说明 |
|------|------|
| `<b>` / `<strong>` | 加粗 |
| `<i>` / `<em>` | 斜体 |
| `<u>` | 下划线 |
| `<s>` / `<strike>` | 删除线 |
| `<sub>` / `<sup>` | 上下标 |

### 标题

| 写法 | 字号映射（mail-editor editor-kit 分支） |
|------|----------------------------------------|
| `<h1>` | 26px，自动加粗 |
| `<h2>` | 22px |
| `<h3>` | 20px |
| `<h4>` | 18px |
| `<h5>` / `<h6>` | 16px |

### 列表

| 写法 | 说明 |
|------|------|
| `<ul><li>...</li></ul>` | 无序列表，**lint autofix → 飞书 native list-block**（加 `data-list-bullet="true"` + `class="temp-li bullet1"` + 全套 inline style + 内容双层 span 包裹）。可多级嵌套 |
| `<ol><li>...</li></ol>` | 有序列表，**lint autofix → 飞书 native list-block**（加 `data-list-number="true"` + `start="1"` + `class="temp-li number1"` + `data-ol-id` + `data-start="N"` + 全套 inline style + 内容双层 span 包裹）。可多级嵌套 |

**不要**用 `<p>一、...</p><p>二、...</p>` 这种「中文编号 + 段落」的列表样式——飞书 mail-editor 不识别且会吞编号字符。

**lint 还会删除 ol/ul 直接子节点中的纯空白文本节点**（多行格式 HTML 中 `<li>` 之间的换行/缩进会被飞书渲染为可见空行，lint 自动 strip 掉）。

### 表格

```html
<table>
  <thead>
    <tr><th>字段</th><th>值</th></tr>
  </thead>
  <tbody>
    <tr><td>x</td><td>y</td></tr>
  </tbody>
</table>
```

允许：`<table>` / `<thead>` / `<tbody>` / `<tfoot>` / `<tr>` / `<td>` / `<th>` / `<caption>` / `<colgroup>` / `<col>`。

**不要**在 `<td>` 内嵌 `<form>` / `<input>` / `<script>`——会被 lint 删除。

### 链接

```html
<a href="https://example.com">外链</a>
<a href="mailto:user@example.com">邮件链接</a>
```

允许的 URL scheme：`http://` / `https://` / `mailto:` / `cid:`（内嵌图片，由 CLI 自动注入）/ `data:image/*`（base64 内嵌图片）。

**禁止**：`javascript:` / `vbscript:` / `file:` —— lint 直接删属性并报 error。

**lint autofix 自动加飞书原生 link 样式**：`class="not-doclink"`（防止飞书把它误识为内部 doc share）+ `style="cursor:pointer;text-decoration:none;color:rgb(20,86,240)"`（飞书蓝、无下划线）。AI 直接写 `<a href>` 即可，lint 会补全。

### 内嵌图片

```html
<img src="./logo.png" alt="logo">
```

- **路径必须是相对路径**（cwd 子树内）；CLI 自动上传到 Drive 并改写为 `cid:` 引用。
- **不接受**绝对路径或 `..` 跨出 cwd 的相对路径。
- 也支持手动 `<img src="cid:abc">` 配合 `--inline '[{"cid":"abc","file_path":"./logo.png"}]'` 精确控制 CID。

## CSS inline style 白名单

inline `style="..."` 属性会按 property 白名单做过滤；非白名单 property 静默删除并报 `STYLE_PROPERTY_DROPPED` 警告。`<style>` 标签透传（lint 不感知，飞书 mail 服务端按渲染场景处理；用 SMTP 直发或外部签发流程时 `<style>` 完整保留）；不允许外链 `<link rel="stylesheet">`。

允许的 CSS property：

| 类别 | 允许的 property |
|------|----------------|
| 颜色 | `color` / `background-color` |
| 字号 / 字重 | `font-size` / `font-weight` / `font-style` / `font-family` |
| 排版 | `text-align` / `text-decoration` / `text-indent` / `text-transform` / `line-height` / `letter-spacing` / `vertical-align` |
| 盒模型 | `padding` / `padding-*` / `margin` / `margin-*` / `width` / `height` / `min-width` / `max-width` / `min-height` / `max-height` / `display` / `box-sizing` |
| 边框 | `border` / `border-*`（含 `border-top` / `border-radius` / `border-color` 等所有 `border-` 前缀） |
| 列表 | `list-style` / `list-style-type` / `list-style-position` |
| 文本流 | `white-space` / `word-break` / `word-wrap` / `overflow` / `overflow-wrap` / `hyphens` |
| 动画 | `transition`（飞书 link 自带 `transition:color 0.3s` 需放行） |
| 杂项 | `cursor` / `opacity` |

**不在白名单的 property（如 `position` / `z-index` / `transform` / `animation` / `filter`）会被删除并报 STYLE_PROPERTY_DROPPED warning。**

## 标准写法示例

### 字号

```html
<span style="font-size:14px">写信默认 14px</span>
<span style="font-size:12px">PC 引用区默认 12px</span>
```

不要使用 `<font size="3">` —— lint 会重写为等价的 `<span style="font-size:16px">`。

### 颜色

```html
<span style="color:rgb(31,35,41)">主色（黑文）</span>
<span style="color:rgb(100,106,115)">副色（灰文）</span>
<span style="color:rgb(225,77,42)">强调色（橙红）</span>
```

飞书原生主色板：
- 主色 `rgb(31,35,41)`（黑文）
- 副色 `rgb(100,106,115)`（灰副）
- 强调 `rgb(225,77,42)`（橙红，用于关键字段强调）
- 浅灰底 `rgb(245,246,247)`（用于表头加粗灰底）

### 行间距

```html
<div style="line-height:1.6">写信默认 1.6 行间距</div>
```

### 段间留白

```html
<p>段一</p>
<p><br></p>
<p>段二</p>
```

不要堆叠 `margin-top` / `margin-bottom`；用一个 `<p><br></p>` 即可。

### 引用

```html
<blockquote>用户撰写的引用文字</blockquote>
```

reply / forward 的引用区由 CLI 自动生成（基于 `adit-html-block` / `history-quote-*` 类），AI 不要手动构造这部分。

## 错误示例（lint 会报错）

```html
<!-- ❌ XSS：on* 属性 -->
<img src="cid:logo" onerror="alert(1)">
→ ATTR_EVENT_HANDLER_BLOCKED, onerror 已删除

<!-- ❌ javascript: URL -->
<a href="javascript:void(0)">click</a>
→ ATTR_JS_URL_BLOCKED

<!-- ❌ <script> 标签 -->
<p>正文</p><script>track()</script>
→ TAG_SCRIPT_BLOCKED, <script> 整段删除

<!-- ❌ 外链 CSS -->
<link rel="stylesheet" href="https://cdn/style.css">
→ TAG_LINK_BLOCKED

<!-- ❌ <iframe> 嵌入 -->
<iframe src="https://x.com"></iframe>
→ TAG_IFRAME_BLOCKED

<!-- ❌ <form> 表单 -->
<form action="/submit"><input name="x"></form>
→ TAG_FORM_BLOCKED, TAG_INPUT_BLOCKED

<!-- ⚠️ <font> 已过时（写信路径自动修复为 <span style>） -->
<font color="red" size="3">x</font>
→ TAG_FONT_TO_SPAN, autofix 后 <span style="color:red;font-size:16px">

<!-- ⚠️ <center> 已过时（autofix 为 <div style="text-align:center">） -->
<center>居中</center>
→ TAG_CENTER_TO_DIV

<!-- ⚠️ 不在白名单的 CSS property -->
<p style="position:absolute; color:red">x</p>
→ STYLE_PROPERTY_DROPPED (position), color 保留
```

## 飞书原生 quote / 编辑器 class

reply / forward 的 quote 区域使用一组飞书原生 class，**lint 全部放行**——AI 不要试图改动这些 class，会破坏原有引用结构：

- `adit-html-block`、`adit-html-block--collapsed`、`adit-html-block--header`
- `history-quote-meta-wrapper`、`history-quote-gap-tag`、`history-quote-forward-title`、`history-quote-meta-after-forward-title`
- `lark-mail-doc-quote`
- `quote-head-meta-mailto`、`lme-line-signal`

`+reply` / `+reply-all` / `+forward` 由 CLI 自动构造这些 quote 元素；AI 编辑写信内容时只关心**用户撰写的正文**部分（引用区由 CLI 自动追加在末尾或前置）。

## 相关文档

- [飞书原生写法（含 3 套模板）](./lark-mail-feishu-native.md)
- [`+lint-html` 用法](./lark-mail-lint-html.md)
- 写信 shortcut: [`+send`](./lark-mail-send.md) / [`+draft-create`](./lark-mail-draft-create.md) / [`+reply`](./lark-mail-reply.md) / [`+reply-all`](./lark-mail-reply-all.md) / [`+forward`](./lark-mail-forward.md) / [`+draft-edit`](./lark-mail-draft-edit.md)
