# ccenv 设计文档

> 跨平台（Windows + macOS）的 Claude Code 启动器。用具名 profile 管理不同的
> endpoint / 认证 / 模型组合，注入环境变量后启动 `claude` 并透传参数。
> Go + cobra 编写，零交互直达，输入最少字符。

## 1. 背景与目标

现有的 `localclaude.ps1`（PowerShell）只能在 Windows 上把 Claude Code 接到本地
llama-swap。ccenv 把它升级为跨平台、多场景的通用启动器，支持 4 种使用场景：

1. 官方账号 plan（订阅登录态）
2. 官方 API（API key）
3. 本地自定义 endpoint（如 llama-swap，等价于 localclaude）
4. DeepSeek 官方 API（Anthropic 兼容端点）

核心诉求：**配置存 `~/.ccenv`，日常启动输入最少字符。**

## 2. 命令集

```
ccenv <name> [claude args...]   # 启动：注入 name profile 的环境变量后启动 claude，剩余参数原样透传
ccenv add <name>                # 增：交互式创建 profile（含模型选择流程）
ccenv edit <name>               # 改：两段式交互编辑 profile
ccenv ls                        # 查：列出所有 profile
ccenv rm <name>                 # 删：删除 profile
ccenv yolo                      # 一键：增量写入 ~/.claude/settings.json 开启 bypass + 危险模式
ccenv                           # 裸跑：打印用法 + 等价 ls
```

- 保留字：`add` / `edit` / `ls` / `rm` / `yolo`，路由优先级高于 profile 名。
- profile 名若与保留字冲突，文档提示规避（不做特殊转义）。
- 路由逻辑：第一个参数命中保留字 → 走对应子命令；否则 → 当作 profile 名走启动逻辑。

## 3. Profile 数据模型与存储

**存储**：`~/.ccenv/config.toml`（用 `os.UserHomeDir()` 解析家目录，跨平台）。
格式 TOML（`pelletier/go-toml/v2`）。

**字段**：

| 字段 | 含义 | 留空语义 |
|------|------|---------|
| `base_url` | claude 用的完整 endpoint URL | 留空 = 不注入，用官方默认 |
| `auth_token` | 认证 token / API key（明文存储） | 留空 = 不注入 |
| `models_url` | 拉取模型列表的完整 URL（OpenAI `/v1/models` 格式） | 留空 = 无列表接口 |
| `model` | 默认模型 id | 留空 = 不注入模型相关变量 |
| `auto_compact_window` | 压缩窗口绝对 token 数 | 留空 = 不注入（不处理） |

**示例 config.toml**：

```toml
[profiles.local]
base_url   = "http://127.0.0.1:11434"
auth_token = "local"
models_url = "http://127.0.0.1:11434/v1/models"
model      = "unsloth/gemma-4-12b-it-qat+mtp4+ngl99+np1+c128k+nothink"
auto_compact_window = 104857

[profiles.ds]
base_url   = "https://api.deepseek.com/anthropic"
auth_token = "sk-xxx"
models_url = "https://api.deepseek.com/v1/models"
model      = "deepseek-chat"

[profiles.api]
auth_token = "sk-ant-xxx"

[profiles.plan]
# 全空 = 官方账号 plan，启动时不注入任何环境变量
```

## 4. 启动逻辑（核心）

`ccenv <name> [args...]`：

1. 从 config.toml 读取 name profile，不存在 → 报错并列出可用 profile。
2. 在**当前进程环境基础上追加/覆盖**以下变量（不清空已有 env）：

| profile 字段非空 | 注入的环境变量 |
|------|------|
| `base_url` | `ANTHROPIC_BASE_URL=<base_url>` |
| `auth_token` | `ANTHROPIC_AUTH_TOKEN=<auth_token>` |
| `model` | `ANTHROPIC_MODEL` + `ANTHROPIC_DEFAULT_HAIKU_MODEL` + `ANTHROPIC_DEFAULT_SONNET_MODEL` + `ANTHROPIC_DEFAULT_OPUS_MODEL`（全设为该模型，复刻 localclaude） |
| `auto_compact_window` | `CLAUDE_CODE_AUTO_COMPACT_WINDOW=<值>` |
| 全空（plan profile） | 不注入任何变量 |

3. `exec.LookPath("claude")` 找二进制，找不到 → 明确报错。
4. `exec.Command` 启动 claude，**继承 stdin/stdout/stderr**（claude 是交互式 TUI，
   终端完全交给它），第一个 profile 名之后的所有参数原样透传。

## 5. add / edit 交互流程

交互库：`AlecAivazis/survey`（输入框、单选、多选）。

### 5.1 `ccenv add <name>`

逐项交互输入：

1. **base_url** — "Base URL (留空=官方默认 api.anthropic.com): "
2. **auth_token** — "Auth token / API key (留空=不注入，用于官方 plan): "
3. **是否有模型列表接口** — "该 endpoint 是否支持 OpenAI 兼容的 /v1/models 接口? (y/N)
   注意: 仅支持 OpenAI 格式，官方 Anthropic 不支持"
   - **y**：
     1. 输入 **models_url**（完整 URL）
     2. 现场 `GET <models_url>`，带 `Authorization: Bearer <auth_token>`（若 auth_token 非空）
     3. 成功 → 列出模型（序号 + id + context_length）→ survey 单选选默认模型
        - 选定后问压缩比例（默认 80%）
        - 该模型有 `context_length` → `auto_compact_window = round(context_length × 比例)`
        - 无 `context_length` 字段 → `auto_compact_window` 留空
     4. 失败（连不上 / 格式错）→ 提示错误 → 降级为手填模型名（同 N 分支），window 留空
   - **N**：手填模型名（"模型名 (留空=不注入模型): "），`models_url` 与 `auto_compact_window` 留空
4. 写入 config.toml。

### 5.2 `ccenv edit <name>`（两段式）

1. **多选菜单**：survey MultiSelect 列出可改字段（base_url / auth_token / models_url / model / 压缩窗口），**空格选中，回车提交**。
2. 仅对选中的字段，走 5.1 对应的交互输入流程（预填现有值作为默认）。
3. 若选中了"模型"且 profile 有 models_url，重新拉 /models 让用户选。
4. 写回 config.toml（保留未改字段）。

## 6. 其他命令

### 6.1 `ccenv ls`
读 config.toml，美化输出每个 profile：名字、base_url、model、是否有 models_url、是否设了压缩窗口。auth_token 脱敏显示（如 `sk-***`）。

### 6.2 `ccenv rm <name>`
删除指定 profile，不存在 → 报错。写回 config.toml。

### 6.3 `ccenv yolo`
读 `~/.claude/settings.json`（不存在则建 `{}`）→ 解析 JSON → **增量**设置：
- `skipDangerousModePermissionPrompt = true`
- `permissions.defaultMode = "bypassPermissions"`

保留其他所有键，写回。JSON 损坏 → 报错不静默覆盖。

## 7. 错误处理

- profile 不存在（启动 / edit / rm）→ 报错 + 列出可用 profile。
- claude 二进制未找到 → 明确报错提示安装 / 加 PATH。
- /models 拉取失败 → 降级手填（见 5.1）。
- config.toml 损坏 → 报错，**不静默覆盖**。
- `~/.claude/settings.json` JSON 损坏 → 报错，不覆盖。

## 8. 项目结构

```
~/Documents/ccenv/
├── go.mod                  # module ccenv
├── main.go                 # cobra root + 入口路由
├── internal/
│   ├── config/             # TOML 读写、profile CRUD
│   ├── profile/            # profile 数据模型
│   ├── launcher/           # 环境变量注入 + 启动 claude
│   ├── models/             # /v1/models 拉取解析
│   └── claudecfg/          # yolo: ~/.claude/settings.json 增量写
└── cmd/                    # cobra 各子命令 (add/edit/ls/rm/yolo + 默认启动)
```

**依赖**：
- `github.com/spf13/cobra` — CLI 框架
- `github.com/AlecAivazis/survey/v2` — 交互式输入
- `github.com/pelletier/go-toml/v2` — TOML 读写

**平台**：Windows + macOS（Go 原生跨平台，无平台特定代码）。

## 9. 设计边界（YAGNI）

明确**不做**：
- 不支持非 OpenAI 格式的模型列表协议（仅 `/v1/models`）。
- 不做 default profile（裸 `ccenv` 不启动，只打印用法 + ls）。
- 不做 `--pick` 临时选模型（模型选择收敛进 add/edit）。
- 不做 `ccenv models` 单独命令。
- 不加密 auth_token（明文存储，用户已确认接受）。
- yolo 不做 per-profile 开关（全局一键）。

