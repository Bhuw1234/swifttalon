# SwiftTalon 项目指南

SwiftTalon 是一个超轻量级的个人 AI 助手，使用 Go 语言开发，可以在 10 美元硬件上运行，内存占用 <10MB。

## 项目概述

- **编程语言**: Go 1.25+
- **许可证**: MIT
- **官网**: https://swifttalon.io
- **GitHub**: https://github.com/Bhuw1234/swifttalon
- **硬件支持**: x86_64, ARM64, RISC-V (Linux)

## 项目结构

```
/home/bhuwan/swifttalon/
├── cmd/swifttalon/          # CLI 入口点
│   └── main.go           # 主程序入口
├── pkg/                   # 核心包
│   ├── agent/            # AI Agent 核心逻辑
│   │   ├── loop.go           # Agent 主循环
│   │   ├── context.go        # 上下文管理
│   │   ├── memory.go         # 记忆系统
│   │   ├── model_fallback.go # 模型故障转移
│   │   └── check_syntax.go   # 语法检查工具
│   ├── auth/             # OAuth/Token 认证
│   │   ├── oauth.go          # OAuth 认证
│   │   ├── token.go          # Token 管理
│   │   ├── store.go          # 凭证存储
│   │   ├── profiles.go       # 多密钥配置管理
│   │   └── pkce.go           # PKCE 支持
│   ├── bus/              # 消息总线
│   ├── channels/         # 通讯渠道
│   ├── config/           # 配置管理
│   ├── constants/        # 常量定义
│   ├── cron/             # 定时任务服务
│   ├── devices/          # 硬件设备支持 (USB 监控)
│   ├── health/           # 健康检查服务器
│   ├── heartbeat/        # 心跳/周期性任务
│   ├── hooks/            # 事件钩子系统
│   ├── logger/           # 日志系统
│   ├── migrate/          # 从 OpenClaw 迁移
│   ├── providers/        # LLM 提供商
│   │   ├── http_provider.go      # HTTP 提供商基类
│   │   ├── claude_provider.go    # Anthropic Claude
│   │   ├── claude_cli_provider.go # Claude CLI
│   │   ├── copilot_provider.go   # GitHub Copilot
│   │   ├── github_copilot_provider.go # GitHub Copilot v2
│   │   ├── codex_provider.go     # OpenAI Codex
│   │   ├── codex_cli_provider.go # Codex CLI / iFlow CLI
│   │   ├── profile_provider.go   # 多密钥轮换
│   │   └── ...
│   ├── session/          # 会话管理
│   ├── skills/           # 技能系统
│   ├── state/            # 持久化状态
│   ├── tools/            # Agent 工具
│   │   ├── filesystem.go   # 文件系统操作
│   │   ├── shell.go        # Shell 命令执行
│   │   ├── web.go          # 网页搜索
│   │   ├── link.go         # URL 内容提取
│   │   ├── edit.go         # 代码编辑
│   │   ├── subagent.go     # 子代理
│   │   ├── cron.go         # 定时任务管理
│   │   ├── i2c.go          # I2C 设备通信
│   │   ├── spi.go          # SPI 设备通信
│   │   ├── spawn.go        # 进程 spawn
│   │   └── message.go      # 消息发送
│   ├── utils/            # 工具函数
│   └── voice/            # 语音系统
│       ├── transcriber.go   # 语音转录 (Groq)
│       └── tts.go           # 文本转语音
├── config/               # 配置示例
├── workspace/            # 工作区模板 (嵌入到二进制)
│   ├── AGENT.md         # Agent 行为指南
│   ├── IDENTITY.md      # Agent 身份定义
│   ├── SOUL.md          # Agent 灵魂
│   ├── USER.md          # 用户偏好
│   ├── memory/          # 长期记忆
│   │   └── MEMORY.md
│   └── skills/          # 内置技能
│       ├── github/
│       ├── hardware/
│       ├── skill-creator/
│       ├── summarize/
│       ├── tmux/
│       └── weather/
└── Makefile             # 构建脚本
```

## 构建和运行

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
# 初始化配置
swifttalon init

# 与 Agent 对话 (单条消息)
swifttalon agent -m "你好"

# 交互式对话模式
swifttalon agent

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

| 命令 | 描述 |
|------|------|
| `swifttalon init` | 初始化配置和工作区 |
| `swifttalon onboard` | (init 的别名) |
| `swifttalon agent -m "..."` | 与 Agent 对话 |
| `swifttalon agent` | 交互式对话模式 |
| `swifttalon gateway` | 启动网关 (多渠道) |
| `swifttalon status` | 查看状态 |
| `swifttalon auth` | 管理认证 (login/logout/status) |
| `swifttalon cron` | 管理定时任务 |
| `swifttalon migrate` | 从 OpenClaw 迁移 |
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

### Gateway 命令选项

```bash
# 调试模式
swifttalon gateway -d
swifttalon gateway --debug
```

### Auth 命令

```bash
# 登录 (OAuth 或 Token)
swifttalon auth login --provider openai
swifttalon auth login --provider openai --device-code  # headless 环境
swifttalon auth login --provider anthropic
swifttalon auth login --provider github-copilot

# 登出
swifttalon auth logout --provider openai
swifttalon auth logout  # 登出所有

# 查看状态
swifttalon auth status
```

### Cron 命令

```bash
# 列出所有定时任务
swifttalon cron list

# 添加定时任务
swifttalon cron add -n "任务名称" -m "Agent 消息" -e 60          # 每 60 秒执行
swifttalon cron add -n "任务名称" -m "Agent 消息" -c "0 9 * * *"  # Cron 表达式

# 启用/禁用任务
swifttalon cron enable <job_id>
swifttalon cron disable <job_id>

# 删除任务
swifttalon cron remove <job_id>
```

### Skills 命令

```bash
# 列出已安装技能
swifttalon skills list

# 从 GitHub 安装技能
swifttalon skills install sipeed/swifttalon-skills/weather

# 安装内置技能
swifttalon skills install-builtin

# 列出内置技能
swifttalon skills list-builtin

# 删除技能
swifttalon skills remove <skill-name>

# 搜索可用技能
swifttalon skills search

# 查看技能详情
swifttalon skills show <skill-name>
```

### Migrate 命令

```bash
# 从 OpenClaw 迁移
swifttalon migrate

# 模拟运行 (不实际修改)
swifttalon migrate --dry-run

# 仅迁移配置
swifttalon migrate --config-only

# 仅迁移工作区
swifttalon migrate --workspace-only

# 强制执行
swifttalon migrate --force

# 刷新工作区文件
swifttalon migrate --refresh

# 指定 OpenClaw 目录
swifttalon migrate --openclaw-home ~/.openclaw
```

## 支持的通讯渠道

- Telegram
- Discord
- QQ
- DingTalk (钉钉)
- LINE
- Slack
- WhatsApp
- Feishu (飞书)
- MaixCAM
- OneBot

## 支持的 LLM 提供商

- OpenRouter (推荐)
- Anthropic (Claude)
- OpenAI (GPT)
- Google Gemini
- Zhipu (智谱)
- Groq
- vLLM (本地)
- Moonshot (月之暗面)
- Ollama (本地)
- NVIDIA NIM
- DeepSeek
- Claude CLI
- Codex CLI
- iFlow CLI
- GitHub Copilot

## Agent 工具

SwiftTalon Agent 内置以下工具:

- **Filesystem**: 文件系统操作 (读取、写入、列出目录等)
- **Shell**: 执行 Shell 命令
- **Message**: 发送消息到各个渠道
- **Web**: 网页搜索 (Brave Search)
- **Link**: URL 内容提取和 AI 摘要
- **Edit**: 代码编辑 (基于 ollama/editor)
- **Subagent**: 子代理 (嵌套 Agent)
- **Cron**: 定时任务管理
- **I2C**: I2C 设备通信 (Linux)
- **SPI**: SPI 设备通信 (Linux)
- **Spawn**: 进程 spawn

## 核心特性

### 模型故障转移 (Model Fallback)

当主模型失败时，自动切换到备用模型。在配置中设置:

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

支持的故障场景:
- 速率限制 (429)
- 认证失败 (401/403)
- 账单/配额问题
- 超时和网络错误
- 上下文长度超限

### 事件钩子系统 (Hooks)

事件驱动的自动化系统，支持以下事件:

- `pre_tool`: 工具执行前
- `post_tool`: 工具执行后
- `pre_llm`: LLM 调用前
- `post_llm`: LLM 调用后
- `on_error`: 发生错误时
- `on_message`: 收到消息时
- `on_message_sent`: 发送消息后

配置示例:

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

### 文本转语音 (TTS)

支持 OpenAI 和 ElevenLabs TTS 服务:

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

OpenAI 可用声音: alloy, echo, fable, onyx, nova, shimmer, ash, ballad, coral, sage

### 多密钥管理 (Auth Profiles)

支持为同一提供商配置多个 API 密钥，自动轮换和故障转移:

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

支持的截断策略:
- `remove_oldest`: 移除最早的消息
- `summarize`: 摘要旧消息 (如果支持)

## 工作区结构

```
~/.swifttalon/
├── config.json           # 配置文件
├── workspace/           # 工作区
│   ├── sessions/       # 对话会话历史
│   ├── memory/         # 长期记忆 (MEMORY.md)
│   ├── state/          # 持久化状态
│   ├── cron/           # 定时任务数据库
│   ├── skills/        # 自定义技能
│   ├── AGENT.md       # Agent 行为指南
│   ├── HEARTBEAT.md   # 周期性任务
│   ├── IDENTITY.md    # Agent 身份
│   ├── SOUL.md        # Agent 灵魂
│   └── USER.md        # 用户偏好
├── skills/            # 全局技能目录
├── hooks/             # 钩子脚本目录
└── cache/             # 缓存目录
    └── tts/           # TTS 音频缓存
```

## 配置说明

配置文件位于 `~/.swifttalon/config.json`:

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
        {"name": "profile1", "api_key": "sk-ant-xxx1", "weight": 10},
        {"name": "profile2", "api_key": "sk-ant-xxx2", "weight": 5}
      ]
    },
    "openai": {
      "api_key": "",
      "api_base": "",
      "profiles": [
        {"name": "primary", "api_key": "sk-xxx1", "weight": 10},
        {"name": "backup", "api_key": "sk-xxx2", "weight": 5, "disabled": false}
      ]
    },
    "openrouter": {
      "api_key": "sk-or-v1-xxx",
      "api_base": "",
      "profiles": [
        {"name": "fast", "api_key": "sk-or-v1-xxx1", "weight": 10},
        {"name": "cheap", "api_key": "sk-or-v1-xxx2", "weight": 5}
      ]
    },
    "groq": {
      "api_key": "gsk_xxx",
      "api_base": ""
    },
    "zhipu": {
      "api_key": "YOUR_ZHIPU_API_KEY",
      "api_base": ""
    },
    "gemini": {
      "api_key": "",
      "api_base": ""
    },
    "vllm": {
      "api_key": "",
      "api_base": ""
    },
    "nvidia": {
      "api_key": "nvapi-xxx",
      "api_base": "",
      "proxy": "http://127.0.0.1:7890"
    },
    "moonshot": {
      "api_key": "sk-xxx",
      "api_base": ""
    },
    "ollama": {
      "api_key": "",
      "api_base": "http://localhost:11434/v1"
    },
    "iflow_cli": {
      "api_base": ""
    }
  },
  "channels": {
    "telegram": {
      "enabled": false,
      "token": "YOUR_TELEGRAM_BOT_TOKEN",
      "proxy": "",
      "allow_from": ["YOUR_USER_ID"]
    },
    "discord": {
      "enabled": false,
      "token": "YOUR_DISCORD_BOT_TOKEN",
      "allow_from": []
    },
    "maixcam": {
      "enabled": false,
      "host": "0.0.0.0",
      "port": 18790,
      "allow_from": []
    },
    "whatsapp": {
      "enabled": false,
      "bridge_url": "ws://localhost:3001",
      "allow_from": []
    },
    "feishu": {
      "enabled": false,
      "app_id": "",
      "app_secret": "",
      "encrypt_key": "",
      "verification_token": "",
      "allow_from": []
    },
    "dingtalk": {
      "enabled": false,
      "client_id": "YOUR_CLIENT_ID",
      "client_secret": "YOUR_CLIENT_SECRET",
      "allow_from": []
    },
    "slack": {
      "enabled": false,
      "bot_token": "xoxb-YOUR-BOT-TOKEN",
      "app_token": "xapp-YOUR-APP-TOKEN",
      "allow_from": []
    },
    "line": {
      "enabled": false,
      "channel_secret": "YOUR_LINE_CHANNEL_SECRET",
      "channel_access_token": "YOUR_LINE_CHANNEL_ACCESS_TOKEN",
      "webhook_host": "0.0.0.0",
      "webhook_port": 18791,
      "webhook_path": "/webhook/line",
      "allow_from": []
    },
    "onebot": {
      "enabled": false,
      "ws_url": "ws://127.0.0.1:3001",
      "access_token": "",
      "reconnect_interval": 5,
      "group_trigger_prefix": [],
      "allow_from": []
    }
  },
  "tools": {
    "web": {
      "search": {
        "api_key": "YOUR_BRAVE_API_KEY",
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
      "pre_llm": {"enabled": false, "script": "pre_llm.sh"},
      "post_llm": {"enabled": false, "script": "post_llm.sh"},
      "on_error": {"enabled": false, "script": "error_handler.sh"},
      "on_message": {"enabled": false, "script": "message_received.sh"},
      "on_message_sent": {"enabled": false, "script": "message_sent.sh"}
    }
  },
  "voice": {
    "enabled": false,
    "provider": "openai",
    "voice": "alloy",
    "model": "tts-1",
    "speed": 1.0,
    "cache_enabled": true,
    "cache_dir": "~/.swifttalon/cache/tts",
    "openai": {
      "api_key": "",
      "api_base": "https://api.openai.com/v1",
      "model": "tts-1",
      "voice": "alloy",
      "speed": "1.0",
      "response": "mp3"
    },
    "elevenlabs": {
      "api_key": "",
      "base_url": "https://api.elevenlabs.io/v1",
      "voice_id": "21m00Tcm4TlvDq8ikWAM",
      "model_id": "eleven_multilingual_v2",
      "language_code": "en",
      "seed": 0
    }
  }
}
```

### 主要配置项

- `agents.defaults.workspace`: 工作区目录路径
- `agents.defaults.restrict_to_workspace`: 限制 Agent 在工作区内操作
- `agents.defaults.model`: 默认使用的模型
- `agents.defaults.model_fallbacks`: 模型故障转移列表
- `agents.defaults.max_tokens`: 最大 token 数量
- `agents.defaults.max_context_tokens`: 最大上下文 token 数量
- `agents.defaults.truncation_strategy`: 上下文截断策略
- `agents.defaults.temperature`: 温度参数
- `agents.defaults.max_tool_iterations`: 最大工具迭代次数
- `providers.*.profiles`: 多密钥配置列表

## 开发约定

- 使用 Go 标准项目结构
- 遵循 Go 语言惯例
- 单元测试使用 `make test` 或 `go test ./...`
- 代码格式化: `go fmt`
- 静态分析: `go vet`

## 开发命令

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

# 完整检查 (fmt + vet + test)
make check

# 构建并运行
make run ARGS="agent -m 'hello'"
```

## 测试

```bash
# 运行所有测试
make test

# 或
go test ./...
```

## 安全沙箱

Agent 默认在沙箱环境中运行，只能访问配置的工作区目录。可通过 `restrict_to_workspace` 配置项控制。

## 构建目标平台

| 平台 | 架构 | 构建命令 |
|------|------|----------|
| Linux | amd64 | `GOOS=linux GOARCH=amd64 go build` |
| Linux | arm64 | `GOOS=linux GOARCH=arm64 go build` |
| Linux | riscv64 | `GOOS=linux GOARCH=riscv64 go build` |
| macOS | arm64 | `GOOS=darwin GOARCH=arm64 go build` |
| Windows | amd64 | `GOOS=windows GOARCH=amd64 go build` |
