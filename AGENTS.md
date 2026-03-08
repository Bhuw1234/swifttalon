# PicoClaw 项目指南

PicoClaw 是一个超轻量级的个人 AI 助手，使用 Go 语言开发，可以在 10 美元硬件上运行，内存占用 <10MB。

## 项目概述

- **编程语言**: Go 1.25+
- **许可证**: MIT
- **官网**: https://picoclaw.io
- **GitHub**: https://github.com/sipeed/picoclaw
- **硬件支持**: x86_64, ARM64, RISC-V (Linux)

## 项目结构

```
/home/bhuwan/picoclaw/
├── cmd/picoclaw/          # CLI 入口点
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
│   │   ├── copilot_provider.go   # GitHub Copilot
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
# 构建当前平台
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
picoclaw init

# 与 Agent 对话 (单条消息)
picoclaw agent -m "你好"

# 交互式对话模式
picoclaw agent

# 启动网关 (多渠道支持)
picoclaw gateway

# 查看状态
picoclaw status
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
| `picoclaw init` | 初始化配置和工作区 |
| `picoclaw onboard` | (init 的别名) |
| `picoclaw agent -m "..."` | 与 Agent 对话 |
| `picoclaw agent` | 交互式对话模式 |
| `picoclaw gateway` | 启动网关 (多渠道) |
| `picoclaw status` | 查看状态 |
| `picoclaw auth` | 管理认证 (login/logout/status) |
| `picoclaw cron` | 管理定时任务 |
| `picoclaw migrate` | 从 OpenClaw 迁移 |
| `picoclaw skills` | 管理技能 |
| `picoclaw version` | 查看版本 |

### Agent 命令选项

```bash
# 调试模式
picoclaw agent -d
picoclaw agent --debug

# 指定消息
picoclaw agent -m "你好"
picoclaw agent --message "你好"

# 指定会话
picoclaw agent -s session_name
picoclaw agent --session session_name
```

### Gateway 命令选项

```bash
# 调试模式
picoclaw gateway -d
picoclaw gateway --debug
```

### Auth 命令

```bash
# 登录 (OAuth 或 Token)
picoclaw auth login --provider openai
picoclaw auth login --provider openai --device-code  # headless 环境
picoclaw auth login --provider anthropic
picoclaw auth login --provider github-copilot

# 登出
picoclaw auth logout --provider openai
picoclaw auth logout  # 登出所有

# 查看状态
picoclaw auth status
```

### Cron 命令

```bash
# 列出所有定时任务
picoclaw cron list

# 添加定时任务
picoclaw cron add -n "任务名称" -m "Agent 消息" -e 60          # 每 60 秒执行
picoclaw cron add -n "任务名称" -m "Agent 消息" -c "0 9 * * *"  # Cron 表达式

# 启用/禁用任务
picoclaw cron enable <job_id>
picoclaw cron disable <job_id>

# 删除任务
picoclaw cron remove <job_id>
```

### Skills 命令

```bash
# 列出已安装技能
picoclaw skills list

# 从 GitHub 安装技能
picoclaw skills install sipeed/picoclaw-skills/weather

# 安装内置技能
picoclaw skills install-builtin

# 列出内置技能
picoclaw skills list-builtin

# 删除技能
picoclaw skills remove <skill-name>

# 搜索可用技能
picoclaw skills search

# 查看技能详情
picoclaw skills show <skill-name>
```

### Migrate 命令

```bash
# 从 OpenClaw 迁移
picoclaw migrate

# 模拟运行 (不实际修改)
picoclaw migrate --dry-run

# 仅迁移配置
picoclaw migrate --config-only

# 仅迁移工作区
picoclaw migrate --workspace-only

# 强制执行
picoclaw migrate --force

# 刷新工作区文件
picoclaw migrate --refresh

# 指定 OpenClaw 目录
picoclaw migrate --openclaw-home ~/.openclaw
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
- ShengSuanYun (生算云)
- Claude CLI
- Codex CLI
- iFlow CLI
- GitHub Copilot

## Agent 工具

PicoClaw Agent 内置以下工具:

- **Filesystem**: 文件系统操作 (读取、写入、列出目录等)
- **Shell**: 执行 Shell 命令
- **Message**: 发送消息到各个渠道
- **Web**: 网页搜索 (Brave Search、DuckDuckGo)
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
    "scripts_dir": "~/.picoclaw/hooks",
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
~/.picoclaw/
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

配置文件位于 `~/.picoclaw/config.json`:

```json
{
  "agents": {
    "defaults": {
      "workspace": "~/.picoclaw/workspace",
      "model": "glm-4.7",
      "model_fallbacks": ["claude-3-haiku-20240307", "gpt-4o-mini"],
      "max_tokens": 8192,
      "max_context_tokens": 100000,
      "truncation_strategy": "remove_oldest",
      "temperature": 0.7,
      "restrict_to_workspace": true,
      "max_tool_iterations": 20
    }
  },
  "providers": {
    "openrouter": {
      "api_key": "sk-or-v1-xxx",
      "profiles": [
        {"name": "fast", "api_key": "sk-or-v1-xxx1", "weight": 10},
        {"name": "cheap", "api_key": "sk-or-v1-xxx2", "weight": 5}
      ]
    },
    "anthropic": {
      "api_key": ""
    },
    "openai": {
      "api_key": ""
    },
    "groq": {
      "api_key": "gsk_xxx"
    },
    "gemini": {
      "api_key": ""
    },
    "zhipu": {
      "api_key": ""
    },
    "vllm": {
      "api_key": "",
      "api_base": ""
    },
    "ollama": {
      "api_key": "",
      "api_base": "http://localhost:11434/v1"
    },
    "nvidia": {
      "api_key": "nvapi-xxx",
      "api_base": "",
      "proxy": "http://127.0.0.1:7890"
    },
    "moonshot": {
      "api_key": "sk-xxx"
    },
    "deepseek": {
      "api_key": "sk-xxx"
    },
    "github_copilot": {
      "api_key": "",
      "connect_mode": "stdio"
    },
    "iflow_cli": {
      "api_base": ""
    }
  },
  "channels": {
    "telegram": {
      "enabled": false,
      "token": "YOUR_BOT_TOKEN",
      "proxy": "",
      "allow_from": ["YOUR_USER_ID"]
    },
    "discord": {
      "enabled": false,
      "token": "YOUR_DISCORD_BOT_TOKEN"
    },
    "slack": {
      "enabled": false,
      "bot_token": "xoxb-xxx",
      "app_token": "xapp-xxx"
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
    "scripts_dir": "~/.picoclaw/hooks",
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
    "cache_enabled": true,
    "openai": {
      "api_key": "",
      "voice": "alloy"
    },
    "elevenlabs": {
      "api_key": "",
      "voice_id": "21m00Tcm4TlvDq8ikWAM"
    }
  }
}
```

### 主要配置项

- `agents.defaults.workspace`: 工作区目录路径
- `agents.defaults.model`: 默认使用的模型
- `agents.defaults.model_fallbacks`: 模型故障转移列表
- `agents.defaults.max_tokens`: 最大 token 数量
- `agents.defaults.max_context_tokens`: 最大上下文 token 数量
- `agents.defaults.truncation_strategy`: 上下文截断策略
- `agents.defaults.temperature`: 温度参数
- `agents.defaults.restrict_to_workspace`: 限制 Agent 在工作区内操作
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