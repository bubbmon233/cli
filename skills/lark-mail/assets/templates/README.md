# Mail HTML Templates

按飞书原生格式（mail-editor lint autofix 后的输出）预制的 HTML 模板，AI 直接 cat 给写信路径 `--body` 即可使用。

> **本目录待添加 8 套 MVP 场景模板**：notify (system-upgrade / policy-change) / weekly (team-report) / decision (proposal-review) / incident (postmortem) / kickoff (project-launch) / welcome (new-hire) / newsletter (daily-brief)

## 命名约定

`{scene}--{subscene}.html`，其中：
- `scene`: 场景大类（notify / weekly / decision / incident / kickoff / welcome / newsletter / sync / followup / release / escalation / review）
- `subscene`: 子场景（如 system-upgrade / team-report / proposal-review）

## 检索

通过 [`scripts/mail_template_tool.py`](../../scripts/mail_template_tool.py) 的 `search` / `summarize` / `extract` 命令检索；元数据在 [`references/mail-template-catalog.md`](../../references/mail-template-catalog.md) + [`references/mail-template-index.json`](../../references/mail-template-index.json)。

## 编写规范

每个模板应：
1. 严格遵循飞书原生写法（详见 [`references/lark-mail-feishu-native.md`](../../references/lark-mail-feishu-native.md) 第二节）；可复制片段见 [`references/mail-feishu-native-snippets.md`](../../references/mail-feishu-native-snippets.md)
2. 通过 `lcpr mail +lint-html --strict --body "$(cat path/to/template.html)"` 验证 0 error / 0 warning
3. 通过 `lcpr mail +draft-create` 创建草稿验证视觉
4. `[变量]` 占位符用方括号包裹，AI 套用时全部替换
