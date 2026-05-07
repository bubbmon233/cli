# 飞书原生写法片段

可直接复制使用的飞书 mail-editor 真原生 HTML 片段。这些片段是写信路径 lint autofix 的目标输出格式，**直接写就能让飞书 mail 渲染跟 mail-editor 自己写的一致**（无段间空行、列表无空行、引用块灰边、链接飞书蓝）。

> **快捷路径**：AI 写最简陋的 HTML（`<p>` / `<ul><li>` / `<ol><li>` / `<blockquote>` / `<a>`），lcpr 写信链路（`+send` / `+draft-create` / `+reply` / `+reply-all` / `+forward` / `+draft-edit` body op）会**自动 lint autofix** 成下面这些飞书原生格式，AI 不需要手抄这些复杂的 inline style + class + data marker。详见 [`lark-mail-feishu-native.md`](./lark-mail-feishu-native.md) 第二节。本文档面向**绕过 lint 直接构造原生 HTML 的高阶场景**（如 raw EML SMTP 发送 / 模板预生成）。

## 段落

```html
<div style="margin-top:4px;margin-bottom:4px;line-height:1.6"><div dir="auto" style="font-size:14px">普通段落文字</div></div>
```

加粗段落：

```html
<div style="margin-top:4px;margin-bottom:4px;line-height:1.6"><div dir="auto" style="font-size:14px"><b><span style="font-family:inherit"><span style="color:rgb(0,0,0)">加粗段落文字</span></span></b></div></div>
```

居中段落：

```html
<div style="margin-top:4px;margin-bottom:4px;line-height:1.6"><div dir="auto" style="text-align:center;font-size:14px">居中文字</div></div>
```

右对齐段落：

```html
<div style="margin-top:4px;margin-bottom:4px;line-height:1.6"><div dir="auto" style="text-align:right;font-size:14px">右对齐文字</div></div>
```

空行（段落间留白）：

```html
<div style="margin-top:4px;margin-bottom:4px;line-height:1.6"><div dir="auto" style="font-size:14px"><br></div></div>
```

## 数字列表（有序，ol）

3 项数字列表（li 间距 0，ol 内 li 共享 `data-ol-id`，每个 li 有 `data-start="N"` 顺序号）：

```html
<ol start="1" style="margin-top:0px;margin-bottom:0px;margin-left:0px;padding-left:0px;list-style-position:inside" data-list-number="true"><li class="temp-li number1" data-li-line="true" data-list="number1" data-ol-id="ol-id-here" data-start="1" style="line-height:1.6;margin-top:0px;margin-bottom:0px;padding-left:0px;display:list-item;list-style-type:decimal;font-family:inherit;font-size:14px;margin-left:0px;list-style-position:inside" dir="auto"><span style="font-family:inherit"><span style="color:rgb(0,0,0)">第一条</span></span></li><li class="temp-li number1" data-li-line="true" data-list="number1" data-ol-id="ol-id-here" data-start="2" style="line-height:1.6;margin-top:0px;margin-bottom:0px;padding-left:0px;display:list-item;list-style-type:decimal;font-family:inherit;font-size:14px;margin-left:0px;list-style-position:inside" dir="auto"><span style="font-family:inherit"><span style="color:rgb(0,0,0)">第二条</span></span></li><li class="temp-li number1" data-li-line="true" data-list="number1" data-ol-id="ol-id-here" data-start="3" style="line-height:1.6;margin-top:0px;margin-bottom:0px;padding-left:0px;display:list-item;list-style-type:decimal;font-family:inherit;font-size:14px;margin-left:0px;list-style-position:inside" dir="auto"><span style="font-family:inherit"><span style="color:rgb(0,0,0)">第三条</span></span></li></ol>
```

> **关键 marker**：`<ol>` 必带 `start="1"` + `data-list-number="true"`；每个 `<li>` 必带 `class="temp-li number1"` + `data-li-line="true"` + `data-list="number1"` + `data-ol-id="<同 ol 内 li 共享的 8-char id>"` + `data-start="<1-based 顺序号>"` + `dir="auto"`；li 文字内容必须用 `<span style="font-family:inherit"><span style="color:rgb(0,0,0)">…</span></span>` 双层 span 包裹。`<ol>` 和 `<li>` 之间不能有换行 / 缩进 / 空白（飞书会渲染为可见空行）。

## 分点列表（无序，ul）

3 项分点列表：

```html
<ul style="margin-top:0px;margin-bottom:0px;margin-left:0px;padding-left:0px;list-style-position:inside" data-list-bullet="true"><li class="temp-li bullet1" data-li-line="true" data-list="bullet1" style="line-height:1.6;margin-top:0px;margin-bottom:0px;padding-left:0px;display:list-item;list-style-type:disc;font-family:inherit;font-size:14px;margin-left:0px;list-style-position:inside" dir="auto"><span style="font-family:inherit"><span style="color:rgb(0,0,0)">第一项</span></span></li><li class="temp-li bullet1" data-li-line="true" data-list="bullet1" style="line-height:1.6;margin-top:0px;margin-bottom:0px;padding-left:0px;display:list-item;list-style-type:disc;font-family:inherit;font-size:14px;margin-left:0px;list-style-position:inside" dir="auto"><span style="font-family:inherit"><span style="color:rgb(0,0,0)">第二项</span></span></li><li class="temp-li bullet1" data-li-line="true" data-list="bullet1" style="line-height:1.6;margin-top:0px;margin-bottom:0px;padding-left:0px;display:list-item;list-style-type:disc;font-family:inherit;font-size:14px;margin-left:0px;list-style-position:inside" dir="auto"><span style="font-family:inherit"><span style="color:rgb(0,0,0)">第三项</span></span></li></ul>
```

> **关键 marker**：`<ul>` 必带 `data-list-bullet="true"`；每个 `<li>` 必带 `class="temp-li bullet1"` + `data-li-line="true"` + `data-list="bullet1"` + `dir="auto"`（**ul 内 li 不需要 data-ol-id 和 data-start**，那是 ol 专用）。其他规则同 ol。

## 缩进 / 列表嵌套

飞书 mail-editor 的"缩进"通过**列表嵌套 + `margin-left:24px`** 实现，不要用 `<p style="padding-left:..">` 或 `&nbsp;` 模拟。lcpr 写信路径下两种等价 pattern 任选其一：

**Pattern A（推荐 / 简洁）：`<div style="padding-left:24px">` 包裹单层 ul/ol**

```html
<div style="margin-top:4px;margin-bottom:4px;line-height:1.6"><div dir="auto" style="font-size:14px"><b>1. 父级标题（普通 div，不是 list）</b></div></div>
<div style="padding-left:24px"><ul style="margin-top:0px;margin-bottom:4px;margin-left:0px;padding-left:0px;list-style-position:inside" data-list-bullet="true"><li class="temp-li bullet2" data-li-line="true" data-list="bullet2" style="line-height:1.6;margin-top:2px;margin-bottom:2px;padding-left:0px;display:list-item;list-style-type:circle;font-family:inherit;font-size:14px;margin-left:0px;list-style-position:inside" dir="auto"><span style="font-family:inherit"><span style="color:rgb(0,0,0)">子项一</span></span></li><li class="temp-li bullet2" data-li-line="true" data-list="bullet2" style="line-height:1.6;margin-top:2px;margin-bottom:2px;padding-left:0px;display:list-item;list-style-type:circle;font-family:inherit;font-size:14px;margin-left:0px;list-style-position:inside" dir="auto"><span style="font-family:inherit"><span style="color:rgb(0,0,0)">子项二</span></span></li></ul></div>
```

多级缩进就嵌套多个 div：`<div padding-left:24px><div padding-left:24px>...</div></div>` = 48px 缩进。

**Pattern B（mail-editor 原生）：嵌套 ul/ol + 内层 `margin:0 0 0 24px` shorthand**

```html
<ul style="margin-top:0px;margin-bottom:0px;margin-left:0px;padding-left:0px;list-style-position:inside" data-list-bullet="true"><ul style="margin:0px 0px 0px 24px;padding-left:0px;list-style-position:inside" data-list-bullet="true"><li class="temp-li bullet2" data-li-line="true" data-list="bullet2" style="line-height:1.6;margin-top:0px;margin-bottom:0px;padding-left:0px;display:list-item;list-style-type:circle;font-family:inherit;font-size:14px;margin-left:0px;list-style-position:inside" dir="auto"><span style="font-family:inherit"><span style="color:rgb(0,0,0)">缩进的子项一</span></span></li><li class="temp-li bullet2" data-li-line="true" data-list="bullet2" style="line-height:1.6;margin-top:0px;margin-bottom:0px;padding-left:0px;display:list-item;list-style-type:circle;font-family:inherit;font-size:14px;margin-left:0px;list-style-position:inside" dir="auto"><span style="font-family:inherit"><span style="color:rgb(0,0,0)">缩进的子项二</span></span></li></ul></ul>
```

> 内层用 `margin:0 0 0 24px` 必须是 **shorthand**（4 值短写法）。lint 的 `hasInlineStyleProp` 会识别 shorthand 已声明 margin-left，不再追加 `margin-left:0` 覆盖你的 24px。如果误写成长写法 `margin-left:24px;margin-top:0;...`，lint 也能识别保留。

**Pattern C：ol(1) → 嵌入 ul circle → ol(2) 接续编号**

```html
<ol start="1" style="margin-top:0px;margin-bottom:0px;margin-left:0px;padding-left:0px;list-style-position:inside" data-list-number="true"><li class="temp-li number1" data-li-line="true" data-list="number1" data-ol-id="ol-id-here" data-start="1" style="line-height:1.6;margin-top:4px;margin-bottom:4px;padding-left:0px;display:list-item;list-style-type:decimal;font-family:inherit;font-size:14px;margin-left:0px;list-style-position:inside" dir="auto"><span style="font-family:inherit"><span style="color:rgb(0,0,0)">第一项</span></span></li></ol><ul style="margin-top:0px;margin-bottom:0px;margin-left:0px;padding-left:0px;list-style-position:inside" data-list-bullet="true"><ul style="margin:0px 0px 0px 24px;padding-left:0px;list-style-position:inside" data-list-bullet="true"><li class="temp-li bullet2" data-li-line="true" data-list="bullet2" data-ol-id="ol-id-here" style="line-height:1.6;margin-top:4px;margin-bottom:4px;padding-left:0px;display:list-item;list-style-type:circle;font-family:inherit;font-size:14px;margin-left:0px;list-style-position:inside" dir="auto"><span style="font-family:inherit"><span style="color:rgb(0,0,0)">第一项的子项</span></span></li></ul></ul><ol start="2" style="margin-top:0px;margin-bottom:0px;margin-left:0px;padding-left:0px;list-style-position:inside" data-list-number="true"><li class="temp-li number1" data-li-line="true" data-list="number1" data-ol-id="ol-id-here" data-start="2" style="line-height:1.6;margin-top:4px;margin-bottom:4px;padding-left:0px;display:list-item;list-style-type:decimal;font-family:inherit;font-size:14px;margin-left:0px;list-style-position:inside" dir="auto"><span style="font-family:inherit"><span style="color:rgb(0,0,0)">第二项</span></span></li></ol>
```

> 显式写 `start="2"` + `data-start="2"` + 共享 `data-ol-id`；lint 的 `ensureAttr` 不覆盖已存在的属性，所以这套接续语义保留。`<ol>` 的 start 属性写在 attribute 而不是 style 里，更不受 ensureFeishuListStyle 影响。

**Pattern D：ol decimal → 嵌入 ol lower-alpha（数字 1/2/3 → 字母 a/b/c）**

```html
<ol style="margin-top:0px;margin-bottom:0px;margin-left:0px;padding-left:0px;list-style-position:inside" data-list-number="true"><ol start="1" style="margin:0px 0px 0px 24px;padding-left:0px;list-style-position:inside" data-list-number="true"><li class="temp-li number2" data-li-line="true" data-list="number2" data-ol-id="ol-id-here" data-start="a" style="line-height:1.6;margin-top:4px;margin-bottom:4px;padding-left:0px;display:list-item;list-style-type:lower-alpha;font-family:inherit;font-size:14px;margin-left:0px;list-style-position:inside" dir="auto"><span style="font-family:inherit"><span style="color:rgb(0,0,0)">字母子项 a</span></span></li><li class="temp-li number2" data-li-line="true" data-list="number2" data-ol-id="ol-id-here" data-start="b" style="line-height:1.6;margin-top:4px;margin-bottom:4px;padding-left:0px;display:list-item;list-style-type:lower-alpha;font-family:inherit;font-size:14px;margin-left:0px;list-style-position:inside" dir="auto"><span style="font-family:inherit"><span style="color:rgb(0,0,0)">字母子项 b</span></span></li></ol></ol>
```

> **关键 marker（嵌套 / 缩进）**：
> - **Pattern A** 的外层 div 用 `padding-left:24px`（lint 不动 div padding，最稳）；适合 div 编号 + 子项 list 的层级
> - **Pattern B/C/D** 的内层 list 用 `margin:0 0 0 24px` shorthand；lint 修复后 shorthand 内的 margin-left:24px 不会被覆盖（`hasInlineStyleProp` 识别 shorthand）
> - **子层 li class 必须降一级**：`bullet1` → `bullet2`，`number1` → `number2`；对应 list-style-type：`disc` → `circle`，`decimal` → `lower-alpha`
> - **ol 编号接续**：同一 list 集合（哪怕被嵌套 ul 打断）共享 `data-ol-id`；ol 父项的 `data-start="N"` 和 `<ol start="N">` 跨嵌套继续递增；ul 子层若属于某个 ol 集合也带同一 `data-ol-id`
> - **每层 li 之间不能有空白文本节点**（同顶层规则；lint autofix 会 strip 掉）

## 引用块（blockquote）

```html
<blockquote style="padding-left:0px;color:rgb(100,106,115);border-left:2px solid rgb(187,191,196);margin:0px"><div dir="auto" style="font-size:14px;padding-left:12px"><span style="font-family:inherit"><span style="color:rgb(0,0,0)">引用文字内容</span></span></div></blockquote>
```

## 链接（飞书蓝 + 无下划线 + not-doclink class）

```html
<a class="not-doclink" href="https://example.com" style="cursor:pointer;text-decoration:none;color:rgb(20,86,240)" rel="nofollow noopener noreferrer">链接文字</a>
```

> **关键 marker**：`class="not-doclink"` 防止飞书把它误识为内部 doc share；`color:rgb(20,86,240)` 是飞书品牌蓝；`text-decoration:none` 去掉默认下划线；`rel` 三件套是 noopener noreferrer 安全属性。

@用户 mention（药丸样式）：

```html
<a href="mailto:user@bytedance.com" style="color:rgb(20,86,240);padding:2px;text-decoration:none;border-radius:999em;margin:0px 2px">@姓名</a>
```

> **关键 marker**：`border-radius:999em` 让链接显示成药丸形（飞书 mention chip 标准样式）；`padding:2px` 给 chip 加左右内边距；`margin:0 2px` 跟前后文字留间隙。

## 文字强调

加粗 + 颜色：

```html
<b><span style="font-family:inherit"><span style="color:rgb(245,74,69)">红色加粗文字</span></span></b>
```

斜体：

```html
<i><span style="font-family:inherit"><span style="color:rgb(0,0,0)">斜体文字</span></span></i>
```

下划线：

```html
<u><span style="font-family:inherit"><span style="color:rgb(0,0,0)">下划线文字</span></span></u>
```

中划线（删除线）：

```html
<s><span style="font-family:inherit"><span style="color:rgb(0,0,0)">中划线文字</span></span></s>
```

## 颜色调色盘

| 用途 | 飞书原生 RGB |
|------|------|
| 主文本（黑） | `rgb(0,0,0)` 或 `rgb(31,35,41)` |
| 副文本（灰） | `rgb(100,106,115)` |
| 三级文本（浅灰）| `rgb(143,149,158)` |
| 飞书蓝（链接 / mention）| `rgb(20,86,240)` |
| 飞书深蓝（重点标题）| `rgb(36,91,219)` |
| 飞书红（警示 / 错误）| `rgb(245,74,69)` 或 `rgb(216,57,49)` |
| 飞书橙（截止 / 紧急）| `rgb(225,77,42)` |
| 浅红底（警示行 / 表头）| `rgb(254,241,241)` 或 `rgb(254,240,240)` |
| 浅灰底（普通表头）| `rgb(242,243,245)` |
| 表格描边 | `rgb(222,224,227)` |

## 表格

飞书原生表格（合并单元格 + 浅红底表头 + 居中红字）：

```html
<table data-ace-table-col-widths="100;100;100;" style="border-collapse:collapse;word-break:break-word;width:auto;text-align:start"><colgroup><col width="100"><col width="100"><col width="100"></colgroup><tbody><tr style="height:20px"><td colspan="3" style="background-color:rgb(254,240,240);min-width:300px;width:300px;min-height:20px;box-sizing:border-box;padding:9px 8px;vertical-align:top;border-width:1px;border-style:solid;border-color:rgb(222,224,227)"><div style="line-height:1.6"><div dir="auto" style="text-align:center;font-size:14px"><i><b><span style="font-family:inherit"><span style="color:rgb(245,74,69)">表格头</span></span></b></i></div></div></td></tr><tr style="height:20px"><td style="min-width:100px;width:100px;min-height:20px;box-sizing:border-box;padding:9px 8px;vertical-align:top;border-width:1px;border-style:solid;border-color:rgb(222,224,227)"><br></td><td style="min-width:100px;width:100px;min-height:20px;box-sizing:border-box;padding:9px 8px;vertical-align:top;border-width:1px;border-style:solid;border-color:rgb(222,224,227)"><br></td><td style="min-width:100px;width:100px;min-height:20px;box-sizing:border-box;padding:9px 8px;vertical-align:top;border-width:1px;border-style:solid;border-color:rgb(222,224,227)"><br></td></tr></tbody></table>
```

## 大标题（>= h1 视觉，但用 div 不用 h1-6）

`<h1>` 在飞书 mail-editor 渲染会带特殊间距，**首选 div 双层嵌套 + 大字号 b 加粗**：

```html
<div style="margin-top:4px;margin-bottom:4px;line-height:1.6"><div dir="auto" style="font-size:14px"><b><span style="font-size:26px"><span style="color:rgb(36,91,219)">飞书深蓝大标题</span></span></b></div></div>
```

## 字体栈

飞书 mail-editor 默认字体栈（标题 / 文字）：

```css
font-family: LarkHackSafariFont, LarkEmojiFont, LarkChineseQuote, -apple-system, "Helvetica Neue", Tahoma, "PingFang SC", "Microsoft Yahei", Arial, "Hiragino Sans GB", sans-serif, "Apple Color Emoji", "Segoe UI Emoji", "Segoe UI Symbol";
```

简化版（够用）：

```css
font-family: LarkHackSafariFont, LarkEmojiFont, LarkChineseQuote, -apple-system, "PingFang SC", Arial, sans-serif;
```

## 飞书原生写法 7 条铁律（lint autofix 实现）

写信路径 lint 会自动把 AI 简单 HTML 改写成上述飞书原生格式。7 条铁律：

1. **段落用双层 div 嵌套，不用 `<p>`**：外层 `margin:4px 0;line-height:1.6`，内层 `dir="auto" style="font-size:14px"`。
2. **ul/ol 必带 data-list-bullet/data-list-number marker**，重置 margin/padding 为 0，加 `list-style-position:inside`；ol 还要 `start="1"`。
3. **li 必带 class + data marker + 内容双层 span 包裹**：`class="temp-li bullet1/number1"` + `data-li-line="true"` + `data-list="bullet1/number1"` + `dir="auto"`；ol 内 li 还要 `data-ol-id="<同 ol 共享 id>"` + `data-start="<位置>"`；内容必须用 `<span style="font-family:inherit"><span style="color:rgb(0,0,0)">…</span></span>` 双层包裹。
4. **ol/ul 内 `<li>` 之间不能有空白文本节点**（换行/缩进会被渲染为空行）；lint autofix 自动 strip 掉。
5. **blockquote 用左侧 2px 灰边样式**：`padding-left:0px;color:rgb(100,106,115);border-left:2px solid rgb(187,191,196);margin:0px`，内层 div 用 `padding-left:12px` 留边距。
6. **链接加 `class="not-doclink"`**：默认飞书蓝 `rgb(20,86,240)` + 无下划线。
7. **HTML4 过时标签自动现代化**：`<font>` → `<span style>`、`<center>` → `<div style="text-align:center">`、`<marquee>/<blink>` → `<span>` 纯文本。

详见 [`lark-mail-feishu-native.md`](./lark-mail-feishu-native.md) 第二节"飞书原生写法铁律（lint autofix 行为）"。

## 相关文档

- [飞书原生写法与文风指引](./lark-mail-feishu-native.md)
- [HTML 兼容白名单](./lark-mail-html-allowlist.md)
- [`+lint-html` 用法](./lark-mail-lint-html.md)
- [模板目录](./mail-template-catalog.md)
- 模板库：[`assets/templates/`](../assets/templates/)
