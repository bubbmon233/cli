# 飞书原生邮件写法与文风指引

> **前置条件：** 先阅读 [`../../lark-shared/SKILL.md`](../../lark-shared/SKILL.md) 了解通用安全规则。本文档定义写信文风、原生 HTML 写法速查与 3 套主流场景模板。所有 HTML 写法对齐 mail-editor `editor-kit` 分支默认行为。

写信路径（`+send` / `+draft-create` / `+reply` / `+reply-all` / `+forward` / `+draft-edit` body op）会自动 lint 并 autofix；本文档定义**直接写就能通过 lint** 的飞书原生写法。配合 [`HTML 兼容白名单`](./lark-mail-html-allowlist.md) 与 [`+lint-html` 用法](./lark-mail-lint-html.md) 使用。

## 一、邮件文风规范

### 风格底线

- **禁机械编号**：不要堆 "一、二、三" / "①②③" / "1) 2) 3)" 这种机械列表写法；该用列表的用 `<ul>` / `<ol>`，该用段落的用 `<p>`，让飞书原生编辑器渲染。
- **emoji 克制**：emoji 仅作状态标签（如收件箱里 ⏰ 紧急 / ✅ 已完成 / ⚠️ 风险），**不要在正文段落里堆 emoji 装饰**。
- **禁冗长 disclaimer 与客套**：删除 "希望对您有帮助" / "如有任何问题请随时联系" / "感谢您的耐心阅读" 这类填充语；信息密度优先。
- **问候 / 落款不超 1 段**：`Hi 各位 Reviewer，` / `各位同事：` 一句话即可，避免冗长寒暄。落款 `[发件人姓名] / [团队] / [日期]` 一行结束。
- **正文长度自适应**：周报 / 月报这类汇总型可上百行，决策请求可数行；不限制正文长度，但**首屏要见到关键信息**。

### 标题与决策项

- **标题 ≤ 30 字**：邮件主题行 `--subject` 控制在 30 字内；超过会在收件箱里被截断。
- **决策 / 结论前置**：第一段就给结论或决策项，让收件人扫一眼就知道是不是需要他做什么。
- **避免 "[xxx] [yyy] [zzz] 关于 ..."**：方括号前缀不超过一组（如 `[紧急]` / `[决策请求]`），多了就乱。

### 标点 / 排版

- 中文使用全角标点（`，` / `。` / `：` / `；`）；英文使用半角（`,` / `.` / `:` / `;`）。
- 中英混排时，英文 / 数字两侧加空格：`本周 PR 共 12 个`，不写成 `本周PR共12个`。
- 不要用 `&nbsp;` 制造缩进；段落前缩进用 `<p style="text-indent:2em">...</p>`。

## 二、飞书原生写法铁律（lint autofix 行为）

写信路径的 lint 会自动把 AI 写的简单 HTML 改写成飞书 mail-editor 真原生格式 —— 你**直接写最简陋的 `<p>/<ul>/<ol>/<blockquote>/<a>` 即可**，lint autofix 会补全飞书 mail-editor 内部使用的 inline style + class + data-* marker，让收件方渲染跟飞书 mail-editor 自己写的一致（无段间空行、列表无空行、链接飞书蓝、引用块左侧灰边）。

| 你写的 | lint autofix 后的飞书原生格式 | 关键 marker |
|--------|------------------------------|------------|
| `<p>文字</p>` | `<div style="margin-top:4px;margin-bottom:4px;line-height:1.6"><div dir="auto" style="font-size:14px">文字</div></div>` | 段落容器双层 div 嵌套，外层定 margin/line-height、内层定 font-size |
| `<ul><li>项</li></ul>` | `<ul style="margin-top:0px;margin-bottom:0px;margin-left:0px;padding-left:0px;list-style-position:inside" data-list-bullet="true">` + `<li class="temp-li bullet1" data-li-line="true" data-list="bullet1" style="line-height:1.6;margin-top:0px;margin-bottom:0px;padding-left:0px;display:list-item;list-style-type:disc;font-family:inherit;font-size:14px;margin-left:0px;list-style-position:inside" dir="auto"><span style="font-family:inherit"><span style="color:rgb(0,0,0)">项</span></span></li>` | ul 加 `data-list-bullet="true"`；li 加 `class="temp-li bullet1"` + 全套 data-* marker + 内容双层 span 包裹 |
| `<ol><li>条</li></ol>` | 类似 ul 但 list-style-type:decimal、ol 加 `start="1"`、li 加 `class="temp-li number1"`、`data-list="number1"`、`data-ol-id="<id>"`、`data-start="N"` | 飞书 mail-editor 用 ol 内 `data-ol-id` (li 共享同一 id) + `data-start` (li 顺序号) 跟踪有序列表 |
| `<blockquote>引用</blockquote>` | `<blockquote style="padding-left:0px;color:rgb(100,106,115);border-left:2px solid rgb(187,191,196);margin:0px">` | 飞书原生引用块：左侧 2px 灰边 + 灰文字 |
| `<a href="...">链接</a>` | `<a class="not-doclink" href="..." style="cursor:pointer;text-decoration:none;color:rgb(20,86,240)">` | `not-doclink` class 防止飞书把它当内部 doc share；飞书蓝 `rgb(20,86,240)` |
| `<font color size>` / `<center>` / `<marquee>` / `<blink>` | autofix → `<span style>` / `<div style="text-align:center">` / 纯文本 / 纯文本 | HTML4 过时标签现代化 |
| 列表内多行格式（缩进/换行）| autofix 删除 ol/ul 直接子节点中的纯空白 text 节点 | 防 li 之间空白文本被渲染为可见空行 |

**关键洞察 1**：飞书 mail-editor 用 class + data-* marker 识别"native list-block"。没这些 marker 的 ul/ol 走 fallback 渲染，li 之间会产生可见空行。

**关键洞察 2**：飞书 mail-editor 不输出 `<p>` —— 段落都是 div 双层嵌套。直接写 `<p>` lint 会改成 div 嵌套；保留 `<p>` 视觉无差异但跟原生不一致。

**关键洞察 3**：所有 inline 文字内容（li / 段落里的纯文字）会被包成 `<span style="font-family:inherit"><span style="color:rgb(0,0,0)">text</span></span>` 双层 span。这是飞书 mail-editor 的 Quill-like 数据模型对所有 text leaf 的默认装饰。

可复制的飞书原生片段见 [`mail-feishu-native-snippets.md`](./mail-feishu-native-snippets.md)；模板见 [`assets/templates/`](../assets/templates/)。

## 三、原生格式速查

字号 / 颜色对齐 mail-editor editor-kit 分支：

| 元素 | 你直接写 | lint 处理 | 关键点 |
|------|----------|-----------|--------|
| 标题 | `<h1>` ~ `<h6>` | 透传 | 字号映射 title 34px / h1 26px / h2 22px / h3 20px / h4 18px / h5 16px / h6 16px，自动加粗 |
| 段落 | `<p>...</p>` | autofix → 飞书原生双层 div | 不需要手写 `<p><br></p>` 制造空行；段落之间 lint 自动处理 4px margin |
| 强调 | `<b>` / `<strong>` 加粗；`<i>` / `<em>` 斜体；`<u>` 下划线；`<s>` 删除线；`<sub>` / `<sup>` 上下标 | 透传 | 全部支持；避免 `<font weight=...>` 等过时写法 |
| 字号 | `<span style="font-size:14px">…</span>`；常用 12px / 14px / 16px / 18px / 24px | 透传 | 写信默认 14px；PC 引用区 12px / mobile 引用区 14px；不走 `<font size>` |
| 颜色 | `<span style="color:rgb(31,35,41)">…</span>` | 透传 | 主色 `rgb(31,35,41)` 黑文 / `rgb(100,106,115)` 灰副 / `rgb(245,246,247)` 浅灰底 / 飞书蓝 `rgb(20,86,240)` |
| 行间距 | `<div style="line-height:1.6">…</div>` | 透传 | 写信默认 1.6 |
| 分点（无序）| `<ul><li>…</li></ul>` | autofix → 飞书原生 inline style + `data-list-bullet` + `class="temp-li bullet1"` + 内容双层 span | li 之间无空行；不要用 "一、二、三" 中文编号当列表 |
| 分点（有序）| `<ol><li>…</li></ol>` | autofix → 飞书原生 inline style + `data-list-number` + `start="1"` + `class="temp-li number1"` + `data-ol-id` + `data-start="N"` + 内容双层 span | li 之间无空行；编号自动递增 |
| 引用 | `<blockquote>…</blockquote>` | autofix → 飞书原生左侧 2px 灰边 + 灰文字 | 用户撰写引用用 `<blockquote>`；reply / forward 引用区由飞书自动生成 |
| 链接 | `<a href="https://...">…</a>` / `<a href="mailto:...">` | autofix → `class="not-doclink"` + 飞书蓝 + 无下划线 | 拒 `javascript:` / `vbscript:`；`cid:` / `data:image/*` 放行 |
| 表格 | `<table><thead><tr><th>…</th></tr></thead><tbody><tr><td>…</td></tr></tbody></table>` | 透传 | 单元格不嵌脚本 / 表单 |
| 内嵌图片 | `<img src="./path/to/file" />` | 透传 | 路径必须是 lark-cli cwd 子树（不接受绝对路径或 `..` 跨出 cwd）；CLI 自动注册为 inline part |
| 水平分割 | `<hr>` | 透传 | 飞书原生支持 |
| 代码 | `<code>…</code>`（行内）/ `<pre>…</pre>`（块）| 透传 | 等宽字体；不渲染 markdown |
| `<style>` 块 | 透传 | 飞书 mail-editor 服务端会自动加 selector scope 前缀（class 隔离），不污染整体页面 |

### 禁用 / 降级清单

| 写法 | 处理 | 说明 |
|------|------|------|
| `<font color size>` | autofix → `<span style="color:..;font-size:..">` | HTML4 过时标签，autofix 现代化 |
| `<center>` | autofix → `<div style="text-align:center">` | 同上 |
| `<marquee>` / `<blink>` | autofix → `<span>` 纯文本 | HTML5 已废，动画属性丢弃 |
| `<style>` 块 | 透传（lint 不感知）| 飞书 mail 服务端自动加 selector scope 前缀，class 隔离不污染页面；用 SMTP 直发 / 接收侧渲染保留 |
| `<link rel="stylesheet">` | 删除 | 不允许外链 CSS |
| `<script>` / `<iframe>` / `<form>` / `<input>` | 整段删除 | XSS / 表单不允许在邮件正文中 |

完整规则见 [`HTML 兼容白名单`](./lark-mail-html-allowlist.md)。

## 四、3 套主流场景模板（完整可复用 HTML）

每套模板含问候开场 + 正文骨架 + 落款，AI 套用时直接替换 `[变量]` 即可。所有模板均通过 lint（无 warning，无 error）。

### 模板 1 — 通知（制度变更 / 系统告警 / 例行播报）

```html
<div style="font-size:14px;line-height:1.6;color:rgb(31,35,41)">
  <p>各位同事：</p>
  <p><br></p>
  <h3 style="color:rgb(31,35,41)">[一句话主题，如「XX 系统将于本周日 22:00 维护升级」]</h3>
  <p>本次升级将影响 <span style="color:rgb(225,77,42);font-weight:bold">[影响范围，例：所有内网用户访问 XX 系统]</span>，预计停机时间 <span style="font-weight:bold">[X] 分钟</span>。请相关同事提前安排工作。</p>
  <p><br></p>
  <h4 style="color:rgb(31,35,41)">关键变更</h4>
  <ul>
    <li>变更点 1：[简短描述]</li>
    <li>变更点 2：[简短描述]</li>
    <li>变更点 3：[简短描述]</li>
  </ul>
  <p><br></p>
  <h4 style="color:rgb(31,35,41)">字段对照</h4>
  <table style="border-collapse:collapse;width:100%">
    <thead>
      <tr style="background-color:rgb(245,246,247)">
        <th style="border:1px solid rgb(220,220,220);padding:8px;text-align:left;font-weight:bold;color:rgb(31,35,41)">字段</th>
        <th style="border:1px solid rgb(220,220,220);padding:8px;text-align:left;font-weight:bold;color:rgb(31,35,41)">变更前</th>
        <th style="border:1px solid rgb(220,220,220);padding:8px;text-align:left;font-weight:bold;color:rgb(31,35,41)">变更后</th>
      </tr>
    </thead>
    <tbody>
      <tr>
        <td style="border:1px solid rgb(220,220,220);padding:8px">[字段名]</td>
        <td style="border:1px solid rgb(220,220,220);padding:8px;color:rgb(100,106,115)">[旧值]</td>
        <td style="border:1px solid rgb(220,220,220);padding:8px;color:rgb(225,77,42);font-weight:bold">[新值]</td>
      </tr>
      <tr>
        <td style="border:1px solid rgb(220,220,220);padding:8px">[字段名]</td>
        <td style="border:1px solid rgb(220,220,220);padding:8px;color:rgb(100,106,115)">[旧值]</td>
        <td style="border:1px solid rgb(220,220,220);padding:8px;color:rgb(225,77,42);font-weight:bold">[新值]</td>
      </tr>
    </tbody>
  </table>
  <p><br></p>
  <h4 style="color:rgb(31,35,41)">生效时间与联系人</h4>
  <p>生效时间：<span style="font-weight:bold">[YYYY-MM-DD HH:mm]</span></p>
  <p>问题反馈：<a href="mailto:[owner@example.com]">[Owner 姓名]</a>（[团队]）</p>
  <p><br></p>
  <p style="color:rgb(100,106,115);font-size:12px">如对本次升级有疑问，请在 [日期] 前邮件反馈。逾期视为已知悉。</p>
  <p><br></p>
  <p>—</p>
  <p>[发件人姓名] / [团队] / [日期]</p>
</div>
```

### 模板 2 — 周报（周度进展 / 月度汇总）

```html
<div style="font-size:14px;line-height:1.6;color:rgb(31,35,41)">
  <p>各位同事：</p>
  <p><br></p>
  <p>以下为 <span style="font-weight:bold">[团队] [起止日期] 周报</span>，关键指标与下周计划如下。</p>
  <p><br></p>
  <h3 style="color:rgb(31,35,41)">本周进展</h3>
  <ul>
    <li><span style="font-weight:bold">[模块 A]</span> 完成 [具体成果]，<span style="color:rgb(100,106,115)">[数据 / 链接]</span></li>
    <li><span style="font-weight:bold">[模块 B]</span> 完成 [具体成果]</li>
    <li><span style="font-weight:bold">[模块 C]</span> 完成 [具体成果]</li>
  </ul>
  <p><br></p>
  <h3 style="color:rgb(31,35,41)">下周计划</h3>
  <ol>
    <li>[计划项 1：明确产出物 + Owner + 截止日]</li>
    <li>[计划项 2]</li>
    <li>[计划项 3]</li>
  </ol>
  <p><br></p>
  <h3 style="color:rgb(31,35,41)">关键指标</h3>
  <table style="border-collapse:collapse;width:100%">
    <thead>
      <tr style="background-color:rgb(245,246,247)">
        <th style="border:1px solid rgb(220,220,220);padding:8px;text-align:left;font-weight:bold;color:rgb(31,35,41)">指标</th>
        <th style="border:1px solid rgb(220,220,220);padding:8px;text-align:right;font-weight:bold;color:rgb(31,35,41)">本周</th>
        <th style="border:1px solid rgb(220,220,220);padding:8px;text-align:right;font-weight:bold;color:rgb(31,35,41)">上周</th>
        <th style="border:1px solid rgb(220,220,220);padding:8px;text-align:right;font-weight:bold;color:rgb(31,35,41)">环比</th>
      </tr>
    </thead>
    <tbody>
      <tr>
        <td style="border:1px solid rgb(220,220,220);padding:8px">[DAU]</td>
        <td style="border:1px solid rgb(220,220,220);padding:8px;text-align:right">[本周值]</td>
        <td style="border:1px solid rgb(220,220,220);padding:8px;text-align:right;color:rgb(100,106,115)">[上周值]</td>
        <td style="border:1px solid rgb(220,220,220);padding:8px;text-align:right;color:rgb(225,77,42);font-weight:bold">[+X%]</td>
      </tr>
      <tr>
        <td style="border:1px solid rgb(220,220,220);padding:8px">[转化率]</td>
        <td style="border:1px solid rgb(220,220,220);padding:8px;text-align:right">[本周值]</td>
        <td style="border:1px solid rgb(220,220,220);padding:8px;text-align:right;color:rgb(100,106,115)">[上周值]</td>
        <td style="border:1px solid rgb(220,220,220);padding:8px;text-align:right;color:rgb(225,77,42);font-weight:bold">[+X%]</td>
      </tr>
      <tr>
        <td style="border:1px solid rgb(220,220,220);padding:8px">[P0 故障数]</td>
        <td style="border:1px solid rgb(220,220,220);padding:8px;text-align:right">[本周值]</td>
        <td style="border:1px solid rgb(220,220,220);padding:8px;text-align:right;color:rgb(100,106,115)">[上周值]</td>
        <td style="border:1px solid rgb(220,220,220);padding:8px;text-align:right">[持平]</td>
      </tr>
    </tbody>
  </table>
  <p><br></p>
  <h3 style="color:rgb(31,35,41)">风险与依赖</h3>
  <table style="border-collapse:collapse;width:100%">
    <thead>
      <tr style="background-color:rgb(245,246,247)">
        <th style="border:1px solid rgb(220,220,220);padding:8px;text-align:left;font-weight:bold;color:rgb(31,35,41)">项</th>
        <th style="border:1px solid rgb(220,220,220);padding:8px;text-align:left;font-weight:bold;color:rgb(31,35,41)">影响</th>
        <th style="border:1px solid rgb(220,220,220);padding:8px;text-align:left;font-weight:bold;color:rgb(31,35,41)">应对 / Owner</th>
      </tr>
    </thead>
    <tbody>
      <tr>
        <td style="border:1px solid rgb(220,220,220);padding:8px">[风险项 1]</td>
        <td style="border:1px solid rgb(220,220,220);padding:8px"><span style="color:rgb(225,77,42);font-weight:bold">高</span> / [影响范围]</td>
        <td style="border:1px solid rgb(220,220,220);padding:8px">[应对方案] / @[Owner]</td>
      </tr>
      <tr>
        <td style="border:1px solid rgb(220,220,220);padding:8px">[依赖项 2]</td>
        <td style="border:1px solid rgb(220,220,220);padding:8px"><span style="color:rgb(100,106,115)">中</span> / [影响范围]</td>
        <td style="border:1px solid rgb(220,220,220);padding:8px">[等待 X 团队回复] / @[Owner]</td>
      </tr>
    </tbody>
  </table>
  <p><br></p>
  <p style="color:rgb(100,106,115);font-size:12px">备注：[补充信息，如周报会议时间、详细数据看板链接等]</p>
  <p><br></p>
  <p>—</p>
  <p>[发件人姓名] / [团队] / [日期]</p>
</div>
```

### 模板 3 — 决策请求（方案审批 / 预算申请 / 拍板）

```html
<div style="font-size:14px;line-height:1.6;color:rgb(31,35,41)">
  <p>Hi 各位 Reviewer，</p>
  <p><br></p>
  <p>下方是 <span style="font-weight:bold">[方案 / 项目名称]</span> 的审批请求，请协助拍板。</p>
  <p><br></p>
  <h3 style="color:rgb(31,35,41)">请求事项</h3>
  <p style="color:rgb(225,77,42);font-weight:bold">请于 [YYYY-MM-DD] 前选定方案 [A/B/C]，并 reply 此邮件确认。</p>
  <p><br></p>
  <h3 style="color:rgb(31,35,41)">背景</h3>
  <p>[2-3 句上下文：项目处于哪个阶段、为何需要决策、不决策的影响。例：「Q3 用户增长目标 30%，目前两套获客方案在评审阶段，需在本周内决定走 A 还是 B 以确保按时启动」。]</p>
  <p><br></p>
  <h3 style="color:rgb(31,35,41)">选项与建议</h3>
  <table style="border-collapse:collapse;width:100%">
    <thead>
      <tr style="background-color:rgb(245,246,247)">
        <th style="border:1px solid rgb(220,220,220);padding:8px;text-align:left;font-weight:bold;color:rgb(31,35,41)">方案</th>
        <th style="border:1px solid rgb(220,220,220);padding:8px;text-align:left;font-weight:bold;color:rgb(31,35,41)">优势</th>
        <th style="border:1px solid rgb(220,220,220);padding:8px;text-align:left;font-weight:bold;color:rgb(31,35,41)">劣势</th>
        <th style="border:1px solid rgb(220,220,220);padding:8px;text-align:left;font-weight:bold;color:rgb(31,35,41)">推荐</th>
      </tr>
    </thead>
    <tbody>
      <tr>
        <td style="border:1px solid rgb(220,220,220);padding:8px"><span style="font-weight:bold">A：[方案 A 名称]</span></td>
        <td style="border:1px solid rgb(220,220,220);padding:8px">[2-3 项关键优势]</td>
        <td style="border:1px solid rgb(220,220,220);padding:8px;color:rgb(100,106,115)">[2-3 项关键劣势]</td>
        <td style="border:1px solid rgb(220,220,220);padding:8px;color:rgb(225,77,42);font-weight:bold;text-align:center">推荐</td>
      </tr>
      <tr>
        <td style="border:1px solid rgb(220,220,220);padding:8px"><span style="font-weight:bold">B：[方案 B 名称]</span></td>
        <td style="border:1px solid rgb(220,220,220);padding:8px">[关键优势]</td>
        <td style="border:1px solid rgb(220,220,220);padding:8px;color:rgb(100,106,115)">[关键劣势]</td>
        <td style="border:1px solid rgb(220,220,220);padding:8px;text-align:center">备选</td>
      </tr>
      <tr>
        <td style="border:1px solid rgb(220,220,220);padding:8px"><span style="font-weight:bold">C：[维持现状 / 不做]</span></td>
        <td style="border:1px solid rgb(220,220,220);padding:8px">[优势：成本低 / 风险小]</td>
        <td style="border:1px solid rgb(220,220,220);padding:8px;color:rgb(100,106,115)">[劣势：错过窗口期等]</td>
        <td style="border:1px solid rgb(220,220,220);padding:8px;text-align:center">不推荐</td>
      </tr>
    </tbody>
  </table>
  <p><br></p>
  <p>建议方案：<span style="color:rgb(225,77,42);font-weight:bold">A：[方案 A 名称]</span>。原因：[1 句话核心理由，例：「与 Q3 KPI 节奏匹配且 ROI 最高」]。</p>
  <p><br></p>
  <h3 style="color:rgb(31,35,41)">需要的决策与时间</h3>
  <ul>
    <li>请 [Approver 1]、[Approver 2] 在 <span style="font-weight:bold">[YYYY-MM-DD HH:mm]</span> 前 reply 此邮件确认方案选择</li>
    <li>抄送：[Stakeholder 1]、[Stakeholder 2]，知悉即可，无需 reply</li>
    <li>详细方案文档：<a href="https://[doc-url]">[方案详细文档链接]</a>（可选预读）</li>
  </ul>
  <p><br></p>
  <p style="color:rgb(100,106,115);font-size:12px">如对方案有疑问，可直接邮件回复或联系 [Owner 姓名] (<a href="mailto:[owner@example.com]">[owner@example.com]</a>)。</p>
  <p><br></p>
  <p>—</p>
  <p>[发件人姓名] / [团队] / [日期]</p>
  <p style="color:rgb(100,106,115);font-size:12px">联系电话：[可选]</p>
</div>
```

> **注**：上面 3 套是教学示例。**实际可用模板见 [`assets/templates/`](../assets/templates/)** —— 那里有 8 个按飞书原生格式（lint autofix 后的输出）写好的真模板，直接 cat 给 lcpr `--body` 即可使用。模板检索用 [`scripts/mail_template_tool.py`](../scripts/mail_template_tool.py) 的 `search` / `summarize` / `extract`。

## 五、套用模板的实务

1. **替换 `[变量]`**：所有方括号占位符是 placeholder，AI 套用时全部替换。
2. **删除不适用的段**：模板提供完整结构，但具体场景下某些段可能不需要（例：通知模板的"字段对照表"不是每次都有）；删除整段而不是留空表格。
3. **正文 `--body` 直接传 HTML**：写信链路（`+send` / `+draft-create` / ...）的 `--body` 参数原样接收上述 HTML 即可：

   ```bash
   lark-cli mail +draft-create --as user \
     --to alice@example.com \
     --subject 'Q3 获客方案审批' \
     --body "$(cat ./template-decision.html)"
   ```

4. **lint 兜底**：写信链路会自动 lint，warning 自动修复，error 整段删除；stdout 同时返回 `lint_applied[]` / `original_blocked[]`。如果想先看会被怎么改，先跑 [`+lint-html`](./lark-mail-lint-html.md)。

5. **预览草稿链接**：所有写信 shortcut 在创建草稿后会返回 `reference` 字段（草稿打开链接），AI 应展示给用户去飞书邮箱 UI 里复核。

## 相关文档

- [飞书原生写法片段库](./mail-feishu-native-snippets.md)
- [模板目录](./mail-template-catalog.md) / [模板索引 JSON](./mail-template-index.json)
- [HTML 兼容白名单](./lark-mail-html-allowlist.md)
- [`+lint-html` 用法](./lark-mail-lint-html.md)
- 写信 shortcut: [`+send`](./lark-mail-send.md) / [`+draft-create`](./lark-mail-draft-create.md) / [`+reply`](./lark-mail-reply.md) / [`+reply-all`](./lark-mail-reply-all.md) / [`+forward`](./lark-mail-forward.md) / [`+draft-edit`](./lark-mail-draft-edit.md)
