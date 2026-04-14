# SwiftTalon 项目指南

SwiftTalon 是一个超轻量级的个人 AI 助手，使用 Go 语言开发，可以在 10 美元硬件上运行，内存占用 <20MB。

> **⚠️ 安全声明**: SwiftTalon 没有官方代币/加密货币。所有在 `pump.fun` 或其他交易平台上的声明都是**诈骗**。唯一官方网站是 [swifttalon.io](https://swifttalon.io)。

## 项目概述

- **编程语言**: Go 1.25.5
- **许可证**: MIT
- **官网**: https://swifttalon.io
- **GitHub**: https://github.com/Bhuw1234/swifttalon
- **硬件支持**: x86_64, ARM64, RISC-V (Linux)
- **项目统计**: 111 个 Go 源文件, 27 个测试文件
- **主程序**: `cmd/swifttalon/main.go` (~1465 行)
- **Logo**: 🐙 (章鱼)

## 核心特性

### 🪶 超轻量级
- **内存占用**: <20MB (比 OpenClaw 减少 99%)
- **启动速度**: <1 秒 (在 0.6GHz 单核上)
- **硬件成本**: 最低 $10 (LicheeRV-Nano 等)

### 🤖 AI 自举开发
- 95% 核心代码由 AI Agent 生成
- 人类在环优化和精炼
- 完全原生 Go 实现

### 🌍 真正可移植
- 单二进制文件，无依赖
- 支持 RISC-V、ARM、x86 架构
- 一键构建，随处运行

## 最新特性

### TUI 终端界面 (v2.0)
- **Dracula 暗色主题**: 现代化配色方案
- **Glamour Markdown 渲染**: 支持代码语法高亮、表格、列表
- **Bubble Tea 框架**: 流畅的交互体验
- **实时打字指示器**: 10 帧动画循环

### 模型支持与故障转移
- **多模型提供商**: 支持 15+ 种 LLM 提供商
- **动态模型列表**: `/list models` 实时显示可用模型
- **智能故障转移**: 主模型失败时自动切换备用模型
- **多密钥管理**: 支持同一提供商配置多个 API Key 自动轮换

### GitHub Copilot 集成
- **双模式支持**: stdio (推荐) 和 gRPC 连接
- **OAuth 认证**: 完整的授权流程
- **权限管理**: 细粒度权限控制

### 事件钩子系统
- **生命周期事件**: pre_tool, post_tool, pre_llm, post_llm
- **消息事件**: on_message, on_message_sent
- **错误处理**: on_error 统一错误处理
- **自定义脚本**: 支持自定义 Shell 脚本扩展

## 项目结构

```
/home/bhuwan/swifttalon/
├── cmd/swifttalon/              # CLI 入口点
│   ├── main.go                  # 主程序入口 (~1465 行)
│   └── workspace/               # 嵌入的工作区模板
│       ├── AGENT.md             # Agent 行为指南
│       ├── IDENTITY.md          # Agent 身份定义
│       ├── SOUL.md              # Agent 灵魂
│       ├── USER.md              # 用户偏好
│       ├── memory/              # 长期记忆
│       └── skills/              # 内置技能
├── pkg/                         # 核心包
│   ├── agent/                   # AI Agent 核心逻辑
│   │   ├── loop.go              # Agent 主循环
│   │   ├── context.go           # 上下文管理
│   │   ├── memory.go            # 记忆系统
│   │   ├── model_fallback.go    # 模型故障转移
│   │   └── check_syntax.go      # 语法检查工具
│   ├── auth/                    # OAuth/Token 认证
│   │   ├── oauth.go             # OAuth 认证流程
│   │   ├── token.go             # Token 管理
│   │   ├── store.go             # 凭证存储
│   │   ├── profiles.go          # 多密钥配置管理
│   │   └── pkce.go              # PKCE 安全扩展
│   ├── bus/                     # 消息总线
│   │   ├── bus.go               # 发布-订阅消息系统
│   │   └── types.go             # 消息类型定义
│   ├── channels/                # 通讯渠道
│   │   ├── telegram.go          # Telegram 机器人
│   │   ├── discord.go           # Discord 机器人
│   │   ├── qq.go                # QQ 频道
│   │   ├── dingtalk.go          # 钉钉机器人
│   │   ├── feishu_*.go          # 飞书机器人 (32/64位)
│   │   ├── slack.go             # Slack 机器人
│   │   ├── line.go              # LINE 机器人
│   │   ├── whatsapp.go          # WhatsApp Bridge
│   │   ├── maixcam.go           # MaixCAM 嵌入式
│   │   └── onebot.go            # OneBot 协议
│   ├── config/                  # 配置管理
│   │   ├── config.go            # 配置解析与验证
│   │   └── config_test.go       # 配置测试
│   ├── constants/               # 常量定义
│   ├── cron/                    # 定时任务服务
│   ├── devices/                 # 硬件设备支持
│   │   ├── service.go           # 设备服务管理
│   │   ├── source.go            # 设备数据源接口
│   │   └── sources/             # 设备源实现
│   │       ├── usb_linux.go     # Linux USB 监控
│   │       └── usb_stub.go      # 其他平台桩
│   ├── health/                  # 健康检查服务器
│   ├── heartbeat/               # 心跳/周期性任务
│   ├── hooks/                   # 事件钩子系统
│   ├── logger/                  # 结构化日志系统
│   ├── migrate/                 # 从 OpenClaw/PicoClaw 迁移
│   ├── providers/               # LLM 提供商
│   │   ├── http_provider.go     # HTTP API 基类
│   │   ├── claude_provider.go   # Anthropic Claude API
│   │   ├── claude_cli_provider.go # Claude CLI 集成
│   │   ├── codex_provider.go    # OpenAI Codex API
│   │   ├── codex_cli_provider.go # Codex CLI / iFlow CLI
│   │   ├── github_copilot_provider.go # GitHub Copilot SDK
│   │   ├── copilot_provider.go  # GitHub Copilot HTTP
│   │   ├── profile_provider.go  # 多密钥轮换
│   │   ├── tool_call_extract.go # 工具调用提取
│   │   └── types.go             # 提供商类型定义
│   ├── session/                 # 会话管理
│   ├── skills/                  # 技能系统
│   ├── state/                   # 持久化状态
│   ├── tools/                   # Agent 工具集
│   │   ├── filesystem.go        # 文件操作 (读/写/列目录/追加)
│   │   ├── shell.go             # Shell 命令执行
│   │   ├── web.go               # 网页搜索 (Brave/DuckDuckGo)
│   │   ├── link.go              # URL 内容提取
│   │   ├── edit.go              # 智能代码编辑
│   │   ├── subagent.go          # 子代理 (嵌套 Agent)
│   │   ├── cron.go              # 定时任务管理
│   │   ├── i2c.go               # I2C 设备通信
│   │   ├── spi.go               # SPI 设备通信
│   │   ├── spawn.go             # 进程 spawn
│   │   ├── message.go           # 消息发送
│   │   └── toolloop.go          # 工具执行循环
│   ├── tui/                     # 终端用户界面
│   │   ├── app.go               # TUI 主应用 (Bubble Tea)
│   │   ├── styles.go            # Dracula 主题样式
│   │   └── types.go             # TUI 类型定义
│   ├── utils/                   # 工具函数
│   └── voice/                   # 语音系统
│       ├── transcriber.go       # 语音转录 (Groq Whisper)
│       └── tts.go               # 文本转语音 (OpenAI/ElevenLabs)
├── workspace/                   # 工作区模板
│   ├── skills/                  # 内置技能
│   │   ├── github/              # GitHub 操作
│   │   ├── hardware/            # 硬件交互
│   │   ├── skill-creator/       # 技能创建
│   │   ├── summarize/           # 内容摘要
│   │   ├── tmux/                # Tmux 管理
│   │   └── weather/             # 天气查询
│   └── memory/                  # 记忆模板
└── Makefile                     # 构建脚本
```

## 快速开始

### 安装依赖

```bash
make deps
```

### 构建

```bash
# 生成嵌入文件并构建当前平台
make build

# 构建所有平台
make build-all
```

### 安装

```bash
# 安装到 ~/.local/bin
make install

# 完全卸载 (删除所有数据)
make uninstall-all
```

### 运行

```bash
# 初始化配置和工作区
swifttalon init

# 启动 TUI 终端界面 (推荐，默认模式)
swifttalon agent

# 使用 readline 交互模式
swifttalon agent --no-tui

# 单条消息对话
swifttalon agent -m "你好，SwiftTalon"

# 启动网关 (多渠道支持)
swifttalon gateway

# 查看状态
swifttalon status
```

### Docker 运行

```bash
# 复制配置模板
cp config/config.example.json config/config.json

# 编辑配置
vim config/config.json

# 启动网关
docker compose --profile gateway up -d
```

## CLI 命令参考

### 主命令

| 命令 | 描述 |
|------|------|
| `swifttalon init` | 初始化配置和工作区 |
| `swifttalon onboard` | init 的别名 |
| `swifttalon agent` | 交互式 TUI 模式 (默认) |
| `swifttalon agent --no-tui` | readline 交互模式 |
| `swifttalon agent -m "..."` | 单条消息对话 |
| `swifttalon gateway` | 启动网关 (多渠道) |
| `swifttalon status` | 查看系统状态 |
| `swifttalon auth` | 管理认证 (login/logout/status) |
| `swifttalon cron` | 管理定时任务 |
| `swifttalon migrate` | 从 OpenClaw/PicoClaw 迁移 |
| `swifttalon skills` | 管理技能 |
| `swifttalon version` | 查看版本 |

### Agent 命令选项

```bash
# 调试模式
swifttalon agent -d
swifttalon agent --debug

# 指定消息
swifttalon agent -m "你好"
swifttalon agent --message "你好"

# 指定会话
swifttalon agent -s session_name
swifttalon agent --session session_name
```

### TUI 快捷键

| 快捷键 | 功能 |
|--------|------|
| `Enter` | 发送消息 |
| `Ctrl+N` | 新建会话 |
| `Ctrl+M` | 切换模型 |
| `Ctrl+L` | 清屏 |
| `Ctrl+C` | 退出 |
| `Esc` | 关闭弹窗 |
| `↑/k` | 向上滚动 |
| `↓/j` | 向下滚动 |
| `?` | 显示帮助 |

### Auth 认证命令

```bash
# OAuth 登录
swifttalon auth login --provider openai
swifttalon auth login --provider anthropic
swifttalon auth login --provider github-copilot

# Headless 环境 (设备码)
swifttalon auth login --provider openai --device-code

# 登出
swifttalon auth logout --provider openai
swifttalon auth logout          # 登出所有

# 查看状态
swifttalon auth status
```

### Cron 定时任务

```bash
# 列出所有任务
swifttalon cron list

# 添加任务
swifttalon cron add -n "每日报告" -m "生成日报" -e 60          # 每 60 秒
swifttalon cron add -n "早安" -m "早上好" -c "0 9 * * *"      # Cron 表达式

# 启用/禁用
swifttalon cron enable <job_id>
swifttalon cron disable <job_id>

# 删除任务
swifttalon cron remove <job_id>
```

### Skills 技能管理

```bash
# 列出已安装技能
swifttalon skills list

# 安装内置技能
swifttalon skills install-builtin

# 从 GitHub 安装
swifttalon skills install owner/repo/skill-name

# 列出内置技能
swifttalon skills list-builtin

# 删除技能
swifttalon skills remove <skill-name>

# 查看详情
swifttalon skills show <skill-name>
```

### Migrate 迁移命令

```bash
# 从 OpenClaw 迁移
swifttalon migrate

# 模拟运行
swifttalon migrate --dry-run

# 仅迁移配置
swifttalon migrate --config-only

# 仅迁移工作区
swifttalon migrate --workspace-only

# 强制执行
swifttalon migrate --force

# 刷新工作区文件
swifttalon migrate --refresh
```

## 支持的通讯渠道

| 渠道 | 难度 | 配置项 | 说明 |
|------|------|--------|------|
| Telegram | ⭐ | Token | 最流行的 Bot 平台 |
| Discord | ⭐ | Token + Intents | 开发者社区 |
| QQ | ⭐ | AppID + Secret | 国内主流 |
| MaixCAM | ⭐ | 内置 | 嵌入式设备 |
| DingTalk | ⭐⭐ | ClientID + Secret | 企业办公 |
| Feishu | ⭐⭐ | AppID + Secret + EncryptKey | 字节生态 |
| Slack | ⭐⭐ | Bot Token + App Token | 团队协作 |
| LINE | ⭐⭐ | Secret + Token + Webhook | 日本/台湾 |
| WhatsApp | ⭐⭐ | Bridge URL | 需要 Bridge |
| OneBot | ⭐⭐ | WebSocket URL | 通用协议 |

## 支持的 LLM 提供商

### HTTP API 提供商

| 提供商 | 推荐模型 | 特点 |
|--------|----------|------|
| **OpenRouter** | anthropic/claude-opus-4-5 | 一站式多模型访问 |
| **Anthropic** | claude-3-opus | Claude 官方 API |
| **OpenAI** | gpt-4o | GPT 官方 API |
| **Google** | gemini-2.0-flash | Gemini 官方 API |
| **智谱 (Zhipu)** | glm-4.7 | 国产模型推荐 |
| **DeepSeek** | deepseek-chat | 推理模型 |
| **Moonshot** | moonshot/kimi-k2.5 | 长上下文 200K |
| **Groq** | groq/llama-3.3-70b | 极速推理 |
| **NVIDIA** | nvidia/llama-3.1-nemotron | 企业级 NIM |
| **Ollama** | ollama/qwen2.5:14b | 本地部署 |
| **vLLM** | - | 本地服务化 |
| **ShengSuanYun** | - | 国产算力云 |

### CLI 提供商

| 提供商 | 说明 |
|--------|------|
| Claude CLI | 本地 Claude Code CLI |
| Codex CLI | 本地 OpenAI Codex CLI |
| iFlow CLI | 本地 iFlow CLI |

### 特殊提供商

| 提供商 | 说明 |
|--------|------|
| GitHub Copilot | SDK 集成，支持 stdio/grpc |

## Agent 工具详解

### 文件系统工具
- `read_file`: 读取文件内容
- `write_file`: 写入文件
- `list_dir`: 列出目录
- `edit_file`: 智能代码编辑
- `append_file`: 追加内容

### 执行工具
- `exec`: 执行 Shell 命令 (带安全沙箱)

### 网络工具
- `web_search`: 网页搜索 (Brave/DuckDuckGo)
- `link`: URL 内容提取和摘要

### 通信工具
- `send_message`: 发送消息到各渠道

### 定时工具
- `cron_add`: 添加定时任务
- `cron_list`: 列出任务
- `cron_remove`: 删除任务

### 硬件工具
- `i2c_read`/`i2c_write`: I2C 设备通信
- `spi_read`/`spi_write`: SPI 设备通信

### 高级工具
- `subagent`: 创建子代理处理复杂任务
- `spawn`: 异步进程管理

## 核心特性详解

### 模型故障转移

```json
{
  "agents": {
    "defaults": {
      "model": "glm-4.7",
      "model_fallbacks": ["claude-3-haiku-20240307", "gpt-4o-mini"]
    }
  }
}
```

支持场景:
- 速率限制 (429)
- 认证失败 (401/403)
- 账单/配额问题
- 超时和网络错误
- 上下文长度超限

### 多密钥管理 (Auth Profiles)

```json
{
  "providers": {
    "openai": {
      "profiles": [
        {"name": "primary", "api_key": "sk-xxx1", "weight": 10},
        {"name": "backup", "api_key": "sk-xxx2", "weight": 5}
      ]
    }
  }
}
```

特性:
- 基于 weight 的负载均衡
- 自动故障检测和切换
- 冷却时间管理
- 认证错误自动隔离

### 上下文管理

```json
{
  "agents": {
    "defaults": {
      "max_context_tokens": 100000,
      "truncation_strategy": "remove_oldest"
    }
  }
}
```

策略:
- `remove_oldest`: 移除最早的消息
- `summarize`: 摘要旧消息

### 网页搜索

```json
{
  "tools": {
    "web": {
      "brave": {
        "enabled": false,
        "api_key": "YOUR_BRAVE_API_KEY",
        "max_results": 5
      },
      "duckduckgo": {
        "enabled": true,
        "max_results": 5
      }
    }
  }
}
```

- **Brave**: 需要 API Key (2000 次/月免费)
- **DuckDuckGo**: 无需 API Key，自动回退

### 心跳任务 (Heartbeat)

创建工作区 `HEARTBEAT.md`:

```markdown
# Periodic Tasks

- Check my email for important messages
- Review my calendar for upcoming events
- Check the weather forecast
```

配置:
```json
{
  "heartbeat": {
    "enabled": true,
    "interval": 30
  }
}
```

### 设备监控

```json
{
  "devices": {
    "enabled": false,
    "monitor_usb": true
  }
}
```

自动监控 USB 设备插入/拔出事件。

### 事件钩子系统

```json
{
  "hooks": {
    "enabled": true,
    "scripts_dir": "~/.swifttalon/hooks",
    "events": {
      "pre_tool": {"enabled": true, "script": "pre_tool.sh"},
      "post_tool": {"enabled": true, "script": "post_tool.sh"},
      "on_error": {"enabled": true, "script": "error_handler.sh"}
    }
  }
}
```

支持事件:
- `pre_tool`: 工具执行前
- `post_tool`: 工具执行后
- `pre_llm`: LLM 调用前
- `post_llm`: LLM 调用后
- `on_error`: 发生错误时
- `on_message`: 收到消息时
- `on_message_sent`: 发送消息后

### 文本转语音 (TTS)

```json
{
  "voice": {
    "enabled": true,
    "provider": "openai",
    "voice": "alloy",
    "model": "tts-1",
    "speed": 1.0,
    "cache_enabled": true,
    "openai": {
      "api_key": "sk-xxx",
      "voice": "alloy"
    },
    "elevenlabs": {
      "api_key": "xxx",
      "voice_id": "21m00Tcm4TlvDq8ikWAM"
    }
  }
}
```

OpenAI 声音选项: alloy, echo, fable, onyx, nova, shimmer, ash, ballad, coral, sage

## 工作区结构

```
~/.swifttalon/
├── config.json              # 主配置文件
├── workspace/               # 工作区目录
│   ├── sessions/           # 对话会话历史
│   ├── memory/             # 长期记忆
│   │   └── MEMORY.md       # 记忆文件
│   ├── state/              # 持久化状态
│   ├── cron/               # 定时任务数据库
│   ├── skills/             # 自定义技能
│   ├── AGENT.md            # Agent 行为指南
│   ├── HEARTBEAT.md        # 周期性任务定义
│   ├── IDENTITY.md         # Agent 身份定义
│   ├── SOUL.md             # Agent 灵魂/个性
│   └── USER.md             # 用户偏好配置
├── skills/                 # 全局技能目录
├── hooks/                  # 钩子脚本目录
└── cache/                  # 缓存目录
    └── tts/                # TTS 音频缓存
```

## 完整配置示例

```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.swifttalon/workspace",
      "restrict_to_workspace": true,
      "model": "glm-4.7",
      "model_fallbacks": ["claude-3-haiku-20240307", "gpt-4o-mini"],
      "max_tokens": 8192,
      "max_context_tokens": 100000,
      "truncation_strategy": "remove_oldest",
      "temperature": 0.7,
      "max_tool_iterations": 20
    }
  },
  "providers": {
    "anthropic": {
      "api_key": "",
      "api_base": "",
      "profiles": [
        {"name": "profile1", "api_key": "sk-ant-xxx1", "weight": 10}
      ]
    },
    "openai": {
      "api_key": "",
      "api_base": "",
      "profiles": [
        {"name": "primary", "api_key": "sk-xxx1", "weight": 10},
        {"name": "backup", "api_key": "sk-xxx2", "weight": 5}
      ]
    },
    "openrouter": {
      "api_key": "sk-or-v1-xxx",
      "profiles": [
        {"name": "fast", "api_key": "sk-or-v1-xxx1", "weight": 10}
      ]
    },
    "groq": {"api_key": "gsk_xxx"},
    "zhipu": {"api_key": "YOUR_ZHIPU_API_KEY"},
    "gemini": {"api_key": ""},
    "ollama": {"api_base": "http://localhost:11434/v1"},
    "github_copilot": {
      "api_key": "",
      "connect_mode": "stdio"
    }
  },
  "channels": {
    "telegram": {
      "enabled": false,
      "token": "YOUR_TELEGRAM_BOT_TOKEN",
      "allow_from": ["YOUR_USER_ID"]
    },
    "discord": {
      "enabled": false,
      "token": "YOUR_DISCORD_BOT_TOKEN"
    },
    "maixcam": {
      "enabled": false,
      "host": "0.0.0.0",
      "port": 18790
    }
  },
  "tools": {
    "web": {
      "brave": {
        "enabled": false,
        "api_key": "YOUR_BRAVE_API_KEY",
        "max_results": 5
      },
      "duckduckgo": {
        "enabled": true,
        "max_results": 5
      }
    }
  },
  "heartbeat": {
    "enabled": true,
    "interval": 30
  },
  "devices": {
    "enabled": false,
    "monitor_usb": true
  },
  "gateway": {
    "host": "0.0.0.0",
    "port": 18790
  },
  "hooks": {
    "enabled": false,
    "scripts_dir": "~/.swifttalon/hooks",
    "events": {
      "pre_tool": {"enabled": false, "script": "pre_tool.sh"},
      "post_tool": {"enabled": false, "script": "post_tool.sh"},
      "on_error": {"enabled": false, "script": "error_handler.sh"}
    }
  },
  "voice": {
    "enabled": false,
    "provider": "openai",
    "voice": "alloy",
    "model": "tts-1",
    "speed": 1.0,
    "cache_enabled": true
  }
}
```

## 开发指南

### 开发命令

```bash
# 下载依赖
make deps

# 更新依赖
make update-deps

# 代码格式化
make fmt

# 静态分析
make vet

# 运行测试
make test

# 完整检查
make check

# 构建并运行
make run ARGS="agent -m 'hello'"
```

### 测试

```bash
# 运行所有测试
make test

# 或
go test ./...

# 带覆盖率
go test -cover ./...
```

### 构建目标平台

| 平台 | 架构 | 命令 |
|------|------|------|
| Linux | amd64 | `GOOS=linux GOARCH=amd64 go build` |
| Linux | arm64 | `GOOS=linux GOARCH=arm64 go build` |
| Linux | riscv64 | `GOOS=linux GOARCH=riscv64 go build` |
| macOS | arm64 | `GOOS=darwin GOARCH=arm64 go build` |
| Windows | amd64 | `GOOS=windows GOARCH=amd64 go build` |

## 安全沙箱

Agent 默认在沙箱环境中运行，只能访问配置的工作区目录。

### 受保护的命令

以下命令即使禁用沙箱也会被阻止:
- `rm -rf`, `del /f`, `rmdir /s` — 批量删除
- `format`, `mkfs`, `diskpart` — 磁盘格式化
- `dd if=` — 磁盘镜像
- 写入 `/dev/sd[a-z]` — 直接磁盘写入
- `shutdown`, `reboot`, `poweroff` — 系统关机
- Fork 炸弹 `:(){ :|:& };:`

## 环境变量

所有配置项都支持环境变量覆盖，格式为 `PICOCLAW_<SECTION>_<KEY>`:

```bash
# 禁用工作区限制
export PICOCLAW_AGENTS_DEFAULTS_RESTRICT_TO_WORKSPACE=false

# 设置模型
export PICOCLAW_AGENTS_DEFAULTS_MODEL=gpt-4o

# 禁用心跳
export PICOCLAW_HEARTBEAT_ENABLED=false

# 设置心跳间隔
export PICOCLAW_HEARTBEAT_INTERVAL=60
```

> **注意**: 环境变量前缀 `PICOCLAW_` 是为了保持与旧版本兼容性。

## 依赖说明

主要依赖:

| 包 | 版本 | 用途 |
|----|------|------|
| `github.com/anthropics/anthropic-sdk-go` | v1.22.1 | Anthropic SDK |
| `github.com/openai/openai-go/v3` | v3.22.0 | OpenAI SDK |
| `github.com/github/copilot-sdk/go` | v0.1.23 | GitHub Copilot SDK |
| `github.com/charmbracelet/bubbletea` | v1.3.10 | TUI 框架 |
| `github.com/charmbracelet/glamour` | v1.0.0 | Markdown 渲染 |
| `github.com/charmbracelet/lipgloss` | v1.1.0 | TUI 样式 |
| `github.com/bwmarrin/discordgo` | v0.29.0 | Discord |
| `github.com/mymmrac/telego` | v1.6.0 | Telegram |
| `github.com/slack-go/slack` | v0.17.3 | Slack |
| `golang.org/x/oauth2` | v0.35.0 | OAuth2 |

## 硬件部署示例

| 设备 | 价格 | 用途 |
|------|------|------|
| LicheeRV-Nano | $9.9 | 最小化家庭助手 |
| NanoKVM | $30-50 | 服务器自动维护 |
| NanoKVM-Pro | $100 | 专业级远程管理 |
| MaixCAM | $50 | 智能监控 |
| MaixCAM2 | $100 | 4K AI 摄像头 |

## 社区与贡献

- **GitHub**: https://github.com/Bhuw1234/swifttalon
- **Issues**: 提交 Bug 报告和功能请求
- **Discussions**: 讨论和想法交流
- **Website**: https://swifttalon.io

## 许可证

MIT License - 详见 LICENSE 文件