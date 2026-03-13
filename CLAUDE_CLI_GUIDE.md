# Claude Code 完全使用指南

> Claude Code 是 Anthropic 官方出品的 AI 编程助手 CLI 工具，能直接在终端中理解代码库、编写代码、执行命令、管理文件。本文档帮助你从零开始掌握它的全部能力。

---

## 目录

1. [安装与登录](#一安装与登录)
2. [启动方式](#二启动方式)
3. [基础对话](#三基础对话)
4. [Slash 命令](#四slash-命令)
5. [键盘快捷键](#五键盘快捷键)
6. [权限控制](#六权限控制)
7. [工作模式](#七工作模式)
8. [会话管理](#八会话管理)
9. [CLAUDE.md 项目记忆](#九claudemd-项目记忆)
10. [Hooks 自动化](#十hooks-自动化)
11. [MCP 服务器扩展](#十一mcp-服务器扩展)
12. [自定义技能（Skills）](#十二自定义技能skills)
13. [子 Agent（Subagents）](#十三子-agentsubagents)
14. [IDE 集成](#十四ide-集成)
15. [CI/CD 与脚本使用](#十五cicd-与脚本使用)
16. [配置文件详解](#十六配置文件详解)
17. [环境变量参考](#十七环境变量参考)
18. [完整 CLI 参数参考](#十八完整-cli-参数参考)
19. [常见场景与最佳实践](#十九常见场景与最佳实践)

---

## 一、安装与登录

### 安装

```bash
# macOS / Linux / WSL
curl -fsSL https://claude.ai/install.sh | bash

# Windows PowerShell
irm https://claude.ai/install.ps1 | iex

# Homebrew（macOS）
brew install --cask claude-code

# WinGet（Windows）
winget install Anthropic.ClaudeCode
```

安装后验证：

```bash
claude --version
```

### 登录

```bash
claude
# 首次运行会自动引导浏览器登录
```

支持的账户类型：
- **Claude Pro / Max / Teams / Enterprise** — 订阅账户直接使用
- **Console API Key** — 用 API Key + 预付积分
- **云服务商** — AWS Bedrock、Google Vertex AI、Microsoft Foundry

切换账户 / 检查状态：

```bash
/status    # 查看当前登录账户和版本
```

---

## 二、启动方式

### 交互式会话（最常用）

```bash
claude                         # 启动空白会话
claude "帮我分析一下这个项目"   # 带初始提示启动
```

### 继续或恢复会话

```bash
claude -c                      # 继续最近的会话
claude -r                      # 打开会话选择器
claude -r "auth-refactor"      # 恢复指定名称的会话
```

### 单次执行（非交互）

```bash
claude -p "解释这个函数"        # 输出结果后退出
claude -p "找出 bug" --output-format json  # JSON 格式输出
```

### 带工作目录隔离

```bash
claude -w feature-auth         # 创建 git worktree 后进入（详见会话管理）
```

---

## 三、基础对话

### 提问和下指令

直接用自然语言描述任务即可：

```
帮我看一下 main.go 有没有内存泄漏
把 getUserById 函数改成支持批量查询
为 UserService 写单元测试
用 grep 查一下哪里调用了 deprecated API
```

### 引用文件

在消息中用 `@` 引用文件，Claude 会自动读取：

```
@src/auth/login.go 这个函数有什么问题？
帮我重构 @internal/db/query.go
```

### 粘贴代码或截图

- 直接粘贴代码片段到输入框
- `Ctrl+V` / `Cmd+V` 粘贴图片（截图、设计稿等）

### 多行输入

| 方法 | 操作 |
|------|------|
| 快速换行 | `\` + `Enter` |
| macOS | `Option+Enter` |
| Linux/Windows | `Shift+Enter` 或 `Ctrl+J` |

### 直接执行命令

在消息前加 `!` 直接运行 shell 命令：

```bash
! git log --oneline -10
! npm test
! go build ./...
```

---

## 四、Slash 命令

在输入框输入 `/` 可以看到所有命令。

### 会话控制

| 命令 | 说明 |
|------|------|
| `/clear` | 清空当前对话历史（开始新话题时用） |
| `/exit` | 退出 Claude Code |
| `/compact [说明]` | 压缩对话历史以节省上下文，可指定保留重点 |
| `/rewind` | 回退到上一个检查点 |

### 信息查看

| 命令 | 说明 |
|------|------|
| `/status` | 查看当前版本、账户、模型信息 |
| `/cost` | 查看本次会话的 token 消耗和费用 |
| `/context` | 可视化当前上下文使用量（彩色方块图） |
| `/stats` | 查看历史使用统计和连续使用记录 |
| `/usage` | 查看订阅限额和速率限制状态 |
| `/doctor` | 检查 Claude Code 安装是否正常 |

### 配置管理

| 命令 | 说明 |
|------|------|
| `/config` | 打开图形化设置界面 |
| `/permissions` | 查看或修改工具权限 |
| `/model` | 切换 AI 模型 |
| `/memory` | 编辑项目的 CLAUDE.md 记忆文件 |
| `/hooks` | 管理 Hooks 自动化配置 |
| `/mcp` | 管理 MCP 服务器连接 |
| `/theme` | 切换颜色主题 |
| `/vim` | 启用/禁用 Vim 编辑模式 |

### 会话管理

| 命令 | 说明 |
|------|------|
| `/resume [会话]` | 恢复历史会话，不带参数打开选择器 |
| `/rename <名称>` | 给当前会话起名，便于以后恢复 |
| `/export [文件名]` | 导出当前对话到文件或剪贴板 |
| `/tasks` | 列出和管理后台任务 |
| `/todos` | 查看当前任务清单 |

### 功能工具

| 命令 | 说明 |
|------|------|
| `/init` | 初始化项目 CLAUDE.md 文件 |
| `/plan` | 直接进入计划模式（只读规划） |
| `/agents` | 管理自定义子 Agent |
| `/plugins` | 浏览和安装插件 |
| `/copy` | 复制上一条回复到剪贴板 |
| `/debug` | 调试当前会话 |

---

## 五、键盘快捷键

### 会话控制

| 快捷键 | 说明 |
|--------|------|
| `Ctrl+C` | 取消当前生成（可以重新提问） |
| `Ctrl+D` | 退出 Claude Code |
| `Ctrl+L` | 清屏（不清除对话历史） |
| `Shift+Tab` | 切换权限模式（循环切换） |
| `Esc` + `Esc` | 打开回退/总结菜单 |

### 模式切换

| 快捷键 | 说明 |
|--------|------|
| `Option+P` / `Alt+P` | 切换模型 |
| `Option+T` / `Alt+T` | 开启/关闭扩展思考模式 |
| `Ctrl+O` | 切换详细输出（查看推理过程） |
| `Ctrl+T` | 显示/隐藏任务列表 |

### 后台任务

| 快捷键 | 说明 |
|--------|------|
| `Ctrl+B` | 将当前任务放到后台（tmux 用户按两次） |
| `Ctrl+F` | 关闭所有后台 Agent（按两次确认） |
| `Ctrl+G` | 在默认编辑器中打开当前输入 |

### 文本编辑

| 快捷键 | 说明 |
|--------|------|
| `Ctrl+K` | 删除到行末 |
| `Ctrl+U` | 删除整行 |
| `Ctrl+Y` | 粘贴刚删除的内容 |
| `Alt+B` | 向前跳一个单词 |
| `Alt+F` | 向后跳一个单词 |
| `Ctrl+R` | 搜索历史命令 |
| `↑` / `↓` | 浏览历史消息 |
| `Tab` | 接受自动补全建议 |
| `@` | 触发文件路径自动补全 |

---

## 六、权限控制

Claude 执行任何操作（编辑文件、运行命令等）都需要权限。你可以精细控制允许哪些操作。

### 权限模式

| 模式 | 说明 | 使用场景 |
|------|------|---------|
| `default` | 首次使用某工具时弹窗询问 | 日常开发 |
| `acceptEdits` | 自动接受文件编辑，命令仍需确认 | 频繁编辑时 |
| `plan` | 只读，不执行任何写操作 | 规划分析阶段 |
| `dontAsk` | 仅执行预先批准的操作 | 严格控制 |
| `bypassPermissions` | 跳过所有权限检查 | 仅限沙箱环境 |

切换方式：
- **交互中**：按 `Shift+Tab` 循环切换
- **启动时**：`claude --permission-mode plan`

### 精细权限规则（settings.json）

```json
{
  "permissions": {
    "allow": [
      "Bash(npm run *)",
      "Bash(git commit *)",
      "Bash(go test ./...)",
      "Read(~/.zshrc)",
      "Edit(./src/**)"
    ],
    "deny": [
      "Bash(git push *)",
      "Bash(rm -rf *)",
      "Edit(./.env)",
      "Edit(./secrets/**)"
    ]
  }
}
```

**规则语法说明：**

| 规则示例 | 说明 |
|---------|------|
| `Bash(npm run *)` | 允许所有 `npm run` 开头的命令 |
| `Bash(git commit *)` | 允许 git commit（但不允许 push） |
| `Edit(./src/**/*)` | 允许编辑 src 目录下所有文件 |
| `Read(~/.zshrc)` | 允许读取 home 目录的配置文件 |
| `WebFetch(domain:github.com)` | 只允许访问 github.com |
| `mcp__github__search_repos` | 允许特定 MCP 工具 |

### 启动时预批准工具

```bash
claude --allowedTools "Bash,Edit,Read"          # 预批准工具列表
claude --disallowedTools "WebFetch"             # 禁止指定工具
claude --dangerously-skip-permissions           # 跳过所有权限（慎用）
```

---

## 七、工作模式

### 普通对话模式

默认模式，Claude 可以读写文件、执行命令，遇到敏感操作会弹窗询问。

### 计划模式（Plan Mode）

只读模式，Claude 只能分析代码不能修改，适合先规划再执行的工作流。

```bash
claude --permission-mode plan    # 启动时进入
/plan                            # 会话中直接切换
Shift+Tab                        # 切换权限模式
```

**推荐工作流：**

```
第一步：计划模式
  → Claude 分析代码库
  → 制定修改方案
  → 你审查并确认

第二步：切换普通模式（Shift+Tab）
  → Claude 按计划执行修改
  → 运行测试验证
```

### 无头模式（-p / --print）

非交互式，执行完直接退出，适合脚本和 CI：

```bash
claude -p "检查代码有没有安全漏洞" > report.txt
cat error.log | claude -p "解释这个错误"
claude -p "生成测试用例" --output-format json
```

### 后台任务模式

长任务放后台，你可以继续做其他事：

```bash
# 在会话中，按 Ctrl+B 将任务放后台
# 用 /tasks 查看后台任务状态
```

---

## 八、会话管理

### 保存和恢复会话

Claude Code 会自动保存所有会话历史，随时可以恢复。

```bash
# 给当前会话命名
/rename auth-feature-refactor

# 继续最近的会话
claude -c

# 恢复指定名称的会话
claude -r "auth-feature-refactor"

# 打开会话选择器（用方向键导航）
claude -r
```

**会话选择器快捷键：**

| 按键 | 功能 |
|------|------|
| `↑` / `↓` | 上下导航 |
| `Enter` | 恢复选中会话 |
| `P` | 预览会话内容 |
| `R` | 重命名会话 |
| `/` | 搜索会话 |
| `B` | 按当前 git 分支过滤 |
| `A` | 切换"仅当前目录/全部项目" |
| `Esc` | 退出选择器 |

### 并行会话 —— Git Worktree

同时做多个功能，互不干扰：

```bash
# 创建隔离的 worktree 并启动会话
claude -w feature-auth      # 指定名称
claude -w                   # 自动生成名称

# 手动方式
git worktree add ../my-project-feature-a -b feature-a
cd ../my-project-feature-a
claude
```

每个 worktree 是独立的 git 分支，Claude 的修改不会影响主分支。

### 从 PR 恢复

```bash
claude --from-pr 123    # 恢复处理 PR #123 的会话
```

### 上下文管理技巧

长时间对话会消耗大量上下文（token），注意及时管理：

```bash
/context    # 查看上下文用量（彩色可视化）
/cost       # 查看 token 消耗
/compact    # 压缩历史，保留关键信息
/clear      # 彻底清空，开始新话题
```

**什么时候用 `/clear` ？**
- 换了一个完全不同的任务
- 上下文用量超过 70%
- Claude 开始忘记前面说的内容

---

## 九、CLAUDE.md 项目记忆

`CLAUDE.md` 是 Claude 的"项目说明书"，每次启动会话时自动读取。把项目规范、注意事项写进去，省去反复解释。

### 创建方式

```bash
/init       # 自动分析项目生成 CLAUDE.md
/memory     # 打开编辑器手动编辑
```

### 文件位置

| 位置 | 作用 | 是否提交 git |
|------|------|------------|
| `./CLAUDE.md` | 项目共享规范 | ✅ 提交 |
| `./CLAUDE.md.local` | 个人本地备注 | ❌ gitignore |
| `~/.claude/CLAUDE.md` | 全局个人习惯 | ❌ 个人 |

### 写什么

```markdown
# 项目技术栈
- Go 1.24，使用 Go modules
- 数据库：PostgreSQL + GORM
- 日志：zap

# 代码规范
- 错误必须向上传递，禁止 panic（测试除外）
- 所有 exported 函数必须写注释
- 数据库操作统一在 repository 层

# 常用命令
- 跑测试：`go test ./... -v`
- 格式化：`gofmt -w .`
- 构建：`go build -o bin/server ./cmd/server`

# 注意事项
- .env 文件包含本地密钥，不要提交
- migrations/ 目录的文件不要手动修改
- 数据库 migration 用：`make migrate-up`

# 文件结构
- cmd/         程序入口
- internal/    内部业务逻辑
- pkg/         可外部使用的包
```

**原则：只写 Claude 无法从代码本身推断的信息，保持在 500 行以内。**

---

## 十、Hooks 自动化

Hooks 是事件触发的自动化脚本，让 Claude 的每个操作都能触发你定制的逻辑。

### 配置文件位置

| 文件 | 作用域 |
|------|-------|
| `~/.claude/settings.json` | 用户全局 |
| `.claude/settings.json` | 项目（提交 git） |
| `.claude/settings.local.json` | 项目本地（不提交） |

### 支持的事件

| 事件 | 触发时机 |
|------|---------|
| `PreToolUse` | 工具执行前（可以阻止） |
| `PostToolUse` | 工具执行成功后 |
| `PostToolUseFailure` | 工具执行失败后 |
| `UserPromptSubmit` | 用户提交消息前 |
| `SessionStart` | 会话开始时 |
| `SessionEnd` | 会话结束时 |
| `Stop` | Claude 完成一轮回复后 |
| `Notification` | 需要通知用户时 |
| `PreCompact` | 上下文压缩前 |

### 实用 Hook 示例

**1. 编辑后自动格式化代码**

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [{
          "type": "command",
          "command": "FILE=$(echo $HOOK_INPUT | jq -r '.tool_input.file_path'); [[ \"$FILE\" == *.go ]] && gofmt -w \"$FILE\""
        }]
      }
    ]
  }
}
```

**2. 阻止修改 .env 文件**

创建脚本 `.claude/hooks/protect-env.sh`：

```bash
#!/bin/bash
INPUT=$(cat)
FILE_PATH=$(echo "$INPUT" | jq -r '.tool_input.file_path // empty')

if [[ "$FILE_PATH" == *".env"* ]]; then
  echo "拒绝：不允许修改 .env 文件" >&2
  exit 2  # exit 2 = 阻止操作
fi

exit 0  # 允许继续
```

```json
{
  "hooks": {
    "PreToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [{
          "type": "command",
          "command": "bash .claude/hooks/protect-env.sh"
        }]
      }
    ]
  }
}
```

**3. 任务完成后桌面通知（macOS）**

```json
{
  "hooks": {
    "Stop": [
      {
        "hooks": [{
          "type": "command",
          "command": "osascript -e 'display notification \"Claude 完成任务\" with title \"Claude Code\"'"
        }]
      }
    ]
  }
}
```

**4. 每次修改后自动运行测试**

```json
{
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [{
          "type": "command",
          "command": "go test ./... 2>&1 | tail -5"
        }]
      }
    ]
  }
}
```

**Hook 退出码含义：**

| 退出码 | 含义 |
|--------|------|
| `0` | 成功，允许继续 |
| `2` | 阻止当前操作（PreToolUse 有效） |
| 其他 | 报错，Claude 会看到错误信息 |

---

## 十一、MCP 服务器扩展

MCP（Model Context Protocol）让 Claude 能调用外部工具和服务，如 GitHub、数据库、Notion 等。

### 添加 MCP 服务器

```bash
# HTTP 服务器（推荐）
claude mcp add --transport http notion https://mcp.notion.com/mcp

# 本地 stdio 服务器
claude mcp add --transport stdio \
  --env GITHUB_TOKEN=your-token \
  github -- npx @modelcontextprotocol/server-github

# 带环境变量的数据库服务器
claude mcp add --transport stdio \
  --env DATABASE_URL=postgres://... \
  mydb -- node ./mcp-server.js
```

### 管理 MCP 服务器

```bash
claude mcp list            # 列出所有服务器
claude mcp get github      # 查看某个服务器详情
claude mcp remove github   # 删除服务器
```

### 配置作用域

| 作用域 | 配置文件 | 共享 |
|--------|---------|------|
| `user` | `~/.claude.json` | 否（个人） |
| `project` | `.mcp.json` | ✅（提交 git，团队共享） |
| `local` | `~/.claude.json` | 否（仅当前项目） |

团队项目推荐用 `project` 作用域，`.mcp.json` 提交到 git：

```json
{
  "mcpServers": {
    "github": {
      "command": "npx",
      "args": ["-y", "@modelcontextprotocol/server-github"],
      "env": {
        "GITHUB_TOKEN": "${GITHUB_TOKEN}"
      }
    }
  }
}
```

注意：`${GITHUB_TOKEN}` 会自动展开环境变量，密钥不会硬编码在文件里。

### 常用 MCP 服务器

| 服务器 | 安装命令 | 功能 |
|--------|---------|------|
| GitHub | `npx @modelcontextprotocol/server-github` | 操作 issues、PR、仓库 |
| 文件系统 | `npx @modelcontextprotocol/server-filesystem` | 扩展文件访问范围 |
| PostgreSQL | `npx @modelcontextprotocol/server-postgres` | 执行 SQL 查询 |
| Brave 搜索 | `npx @modelcontextprotocol/server-brave-search` | 网络搜索 |
| Notion | HTTP: `https://mcp.notion.com/mcp` | 读写 Notion 文档 |

### 在对话中使用

添加后直接告诉 Claude 使用：

```
用 GitHub MCP 查一下 issues 列表里有没有关于登录 bug 的 issue
帮我查一下数据库里 users 表有多少条记录
```

---

## 十二、自定义技能（Skills）

Skills 是可复用的工作流命令，用 `/skill-name` 调用。

### 创建技能

在 `.claude/skills/fix-issue/SKILL.md` 创建：

```markdown
---
name: fix-issue
description: 修复一个 GitHub issue
allowed-tools: Read, Edit, Bash
---

分析并修复 GitHub issue $ARGUMENTS：

1. 运行 `gh issue view $ARGUMENTS` 查看 issue 详情
2. 理解问题，找到相关代码文件
3. 实现修复
4. 写或更新单元测试
5. 运行测试确认通过
6. 生成描述清晰的 commit message 并提交
7. 创建 PR
```

### 使用技能

```bash
/fix-issue 123          # 修复 issue #123
/fix-issue 456          # 修复 issue #456
```

### 参数说明

| 变量 | 含义 |
|------|------|
| `$ARGUMENTS` | 所有参数（字符串） |
| `$ARGUMENTS[0]` | 第一个参数 |
| `$1`, `$2`... | 第 N 个参数 |

### 技能文件位置

| 位置 | 作用域 |
|------|-------|
| `.claude/skills/<name>/SKILL.md` | 项目（提交 git） |
| `~/.claude/skills/<name>/SKILL.md` | 用户全局 |

### 更多技能示例

**代码审查技能** `.claude/skills/review/SKILL.md`：

```markdown
---
name: review
description: 代码审查
allowed-tools: Read, Grep, Glob
---

对 $ARGUMENTS 进行代码审查，重点检查：
- 安全漏洞（SQL注入、XSS、权限绕过）
- 错误处理是否完善
- 并发安全问题
- 性能隐患
- 代码可读性

输出结构化的审查报告，按严重程度分级。
```

---

## 十三、子 Agent（Subagents）

Subagent 是在隔离上下文中运行的专职 AI，适合耗时的研究分析任务，不影响主会话的上下文。

### 内置 Subagent

| Agent | 特点 | 适合做什么 |
|-------|------|----------|
| `Explore` | 只读、快速 | 搜索代码、理解项目结构 |
| `Plan` | 只读、深度分析 | 规划复杂功能的实现方案 |
| `general-purpose` | 全功能 | 复杂多步骤任务 |

在对话中使用：

```
用 Explore agent 帮我找一下所有处理用户认证的文件
用 Plan agent 给我设计一个缓存系统的实现方案
```

### 创建自定义 Subagent

在 `.claude/agents/code-reviewer.md` 创建：

```markdown
---
name: code-reviewer
description: 专职代码审查 agent，检查代码质量和安全性
tools: Read, Grep, Glob
model: sonnet
permissionMode: plan
maxTurns: 20
---

你是一个严格的代码审查专家。每次审查都要检查：

1. **安全性** - SQL注入、XSS、权限控制
2. **错误处理** - 所有错误路径是否覆盖
3. **并发安全** - race condition、deadlock
4. **性能** - 不必要的内存分配、N+1 查询
5. **可维护性** - 命名、注释、复杂度

输出按严重程度分级的问题列表。
```

在对话中调用：

```
用 code-reviewer agent 审查 @src/payment/ 目录
```

### 管理 Subagent

```bash
/agents    # 打开 agent 管理界面（创建/编辑/查看）
```

---

## 十四、IDE 集成

### VS Code

1. 打开 VS Code 扩展市场
2. 搜索 `Claude Code` 并安装
3. 侧边栏会出现 Claude 图标

**功能：**
- 在编辑器内查看 Claude 建议的修改（diff 视图）
- 右键菜单直接把选中代码发给 Claude
- 任务在后台运行，不打断编码
- `@` 提及打开的文件

### JetBrains（IntelliJ、GoLand、PyCharm 等）

1. 打开 JetBrains Marketplace
2. 搜索 `Claude Code` 并安装
3. 重启 IDE 后在侧边栏找到 Claude

### 自动连接 IDE

```bash
claude --ide    # 自动检测并连接运行中的 IDE
```

---

## 十五、CI/CD 与脚本使用

### 基本用法

```bash
# 单次执行，结果写到文件
claude -p "分析 diff 并生成 PR 描述" > pr-description.md

# 管道输入
cat error.log | claude -p "总结这些错误的根本原因"

# JSON 输出（便于脚本解析）
claude -p "找出所有 TODO 注释" --output-format json | jq '.result'
```

### 在 GitHub Actions 中使用

```yaml
name: AI Code Review

on:
  pull_request:
    types: [opened, synchronize]

jobs:
  review:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Install Claude Code
        run: curl -fsSL https://claude.ai/install.sh | bash

      - name: Run AI Review
        env:
          ANTHROPIC_API_KEY: ${{ secrets.ANTHROPIC_API_KEY }}
        run: |
          git diff origin/main...HEAD > changes.diff
          claude -p "审查这个 diff，找出潜在问题" \
            --allowedTools "Read,Bash(git *)" \
            --output-format json \
            < changes.diff > review.json

      - name: Post Review Comment
        uses: actions/github-script@v6
        with:
          script: |
            const review = require('./review.json')
            github.rest.issues.createComment({
              issue_number: context.issue.number,
              owner: context.repo.owner,
              repo: context.repo.repo,
              body: review.result
            })
```

### 批量处理文件

```bash
# 批量处理多个文件
find . -name "*.go" -newer last_review | while read file; do
  echo "审查 $file..."
  claude -p "审查 $file 的代码质量" \
    --allowedTools "Read" \
    --no-session-persistence \
    >> review-report.md
done
```

### 预算控制

```bash
# 限制最大花费（API Key 用户）
claude -p "分析整个代码库" \
  --max-budget-usd 2.00    # 超过 $2 自动停止
```

---

## 十六、配置文件详解

### 文件层级（优先级从高到低）

```
系统托管设置（企业 IT 部署）
  > CLI 参数
    > .claude/settings.local.json（项目本地，不提交）
      > .claude/settings.json（项目共享，提交 git）
        > ~/.claude/settings.json（用户全局）
```

### 完整配置示例

**项目配置 `.claude/settings.json`：**

```json
{
  "permissions": {
    "defaultMode": "default",
    "allow": [
      "Bash(go test ./...)",
      "Bash(go build ./...)",
      "Bash(gofmt -w *)",
      "Bash(git add *)",
      "Bash(git commit *)",
      "Bash(git diff *)",
      "Bash(git log *)",
      "Bash(git status)"
    ],
    "deny": [
      "Bash(git push --force *)",
      "Edit(./.env)",
      "Edit(./secrets/**)"
    ]
  },
  "hooks": {
    "PostToolUse": [
      {
        "matcher": "Edit|Write",
        "hooks": [{
          "type": "command",
          "command": "FILE=$(echo $HOOK_INPUT | jq -r '.tool_input.file_path'); [[ \"$FILE\" == *.go ]] && gofmt -w \"$FILE\" 2>/dev/null; exit 0"
        }]
      }
    ]
  }
}
```

**用户全局配置 `~/.claude/settings.json`：**

```json
{
  "model": "claude-sonnet-4-6",
  "language": "zh-CN",
  "permissions": {
    "allow": [
      "Read(~/.zshrc)",
      "Read(~/.gitconfig)"
    ]
  }
}
```

---

## 十七、环境变量参考

| 变量 | 说明 | 默认值 |
|------|------|--------|
| `ANTHROPIC_API_KEY` | API Key（Console 用户） | — |
| `CLAUDE_CODE_ENABLE_TELEMETRY` | 启用遥测数据收集 | `0` |
| `MAX_THINKING_TOKENS` | 限制思考 token 数量 | 自动 |
| `ENABLE_TOOL_SEARCH` | MCP 工具搜索模式 | `auto` |
| `MAX_MCP_OUTPUT_TOKENS` | MCP 工具输出上限 | `50000` |
| `MCP_TIMEOUT` | MCP 服务器启动超时（毫秒） | `10000` |
| `CLAUDE_AUTOCOMPACT_PCT_OVERRIDE` | 自动压缩触发阈值（%） | `70` |
| `CLAUDE_ENV_FILE` | 额外的环境变量文件路径 | — |

---

## 十八、完整 CLI 参数参考

### 会话控制

```bash
claude                          # 启动交互会话
claude "初始提示"                # 带初始提示启动
claude -p "提示"                 # 非交互模式（执行后退出）
claude --print "提示"            # 同上
claude -c                        # 继续最近会话
claude --continue                # 同上
claude -r [名称]                  # 恢复会话
claude --resume [名称]            # 同上
claude -w [名称]                  # 创建 git worktree 会话
claude --worktree [名称]          # 同上
claude --from-pr 123             # 从 PR 恢复
```

### 权限与工具

```bash
--permission-mode <mode>         # 权限模式（plan/default/acceptEdits/dontAsk）
--allowedTools "Bash,Edit,Read"  # 预批准工具
--disallowedTools "WebFetch"     # 禁止工具
--dangerously-skip-permissions   # 跳过所有权限（慎用）
```

### 输出格式

```bash
--output-format text             # 纯文本（默认）
--output-format json             # JSON 结构化
--output-format stream-json      # 实时流式 JSON
```

### 模型选择

```bash
--model sonnet                   # 使用 Sonnet
--model opus                     # 使用 Opus
--model claude-opus-4-6          # 指定完整模型名
```

### 系统提示

```bash
--system-prompt "自定义提示"     # 替换默认系统提示
--append-system-prompt "补充"    # 追加到系统提示
```

### MCP 配置

```bash
--mcp-config ./mcp.json          # 加载 MCP 配置文件
--strict-mcp-config              # 只使用指定的 MCP 服务器
```

### 预算控制

```bash
--max-budget-usd 5.00            # 花费上限（仅 -p 模式）
--max-turns 20                   # 最大对话轮次
--no-session-persistence         # 不保存会话（仅 -p 模式）
```

### 调试

```bash
--debug                          # 开启调试输出
--verbose                        # 详细日志
--version                        # 显示版本号
```

---

## 十九、常见场景与最佳实践

### 场景一：修复 bug

```
帮我看一下为什么 @src/auth/login.go 的登录接口返回 401

[Claude 分析后给出建议]

按你说的改，改完跑一下测试
```

### 场景二：添加新功能（计划→实现）

```bash
# 1. 先规划
claude --permission-mode plan
> 我想给用户系统加一个"找回密码"功能，分析一下需要改哪些地方

# 2. 审查计划后，切换到普通模式
Shift+Tab

# 3. 执行
> 按刚才的方案实现，先从数据库层开始
```

### 场景三：代码审查

```
审查一下 @src/payment/ 目录，重点关注安全问题
```

### 场景四：大规模重构

```bash
# 用 worktree 隔离，不影响主分支
claude -w refactor-db-layer

> 把项目里所有直接操作数据库的代码统一迁移到 repository 层，
  先分析一下涉及哪些文件
```

### 场景五：学习陌生代码库

```
帮我理解这个项目的整体架构，从入口开始

@main.go 是怎么初始化各个模块的？

认证流程是怎么走的？从请求进来到验证完成
```

### 常见错误和解决方法

**Claude 开始"遗忘"之前说过的内容**
→ `/context` 检查用量，超过 70% 就 `/compact` 或 `/clear`

**Claude 做了不该做的操作**
→ `Ctrl+C` 立即停止；在 `settings.json` 的 `deny` 里加规则

**每次都要重复同样的说明**
→ 把规范写进 `CLAUDE.md`，一次配置永久生效

**任务太复杂，Claude 容易跑偏**
→ 先用计划模式制定方案，审查确认后再执行
→ 把大任务拆成小步骤逐步完成

**同时处理多个功能**
→ 用 `claude -w` 创建多个 worktree 并行工作

---

## 快速上手检查清单

第一次使用 Claude Code 时，按顺序做这几件事：

```
□ 1. 安装并登录（claude --version 确认正常）
□ 2. 在项目根目录运行 /init 创建 CLAUDE.md
□ 3. 在 CLAUDE.md 里写入项目规范和常用命令
□ 4. 创建 .claude/settings.json 配置常用权限
□ 5. 试用 /rename 给会话起名
□ 6. 试用 /context 查看上下文用量
□ 7. 试用计划模式分析一个功能的实现思路
```

---

*文档基于 Claude Code 最新版本整理，如有功能变化请参考 `/help` 命令获取最新信息。*
