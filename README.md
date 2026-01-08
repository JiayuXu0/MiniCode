# MiniCode

**30 天打造 Mini 版 Claude Code**

一个从零开始构建 AI 编程助手的实战教程项目。通过 30 天的渐进式学习，你将掌握如何使用 Go 语言构建一个类似 Claude Code 的终端 AI Agent。

## 🎯 项目目标

通过这个教程，你将学会：
- 使用 Fantasy SDK 构建 AI Agent
- 实现工具调用（文件搜索、代码编辑、命令执行等）
- 构建交互式 TUI 界面
- 实现流式输出、多轮对话、上下文管理
- 最终打造一个可用的 AI 编程助手

## 📅 教程进度

| 阶段 | 天数 | 主题 | 状态 |
|------|------|------|------|
| 基础 | Day 1 | [创建第一个 AI Agent](docs/01.md) | ✅ |
| 基础 | Day 2 | [给 Agent 添加工具](docs/02.md) | ✅ |
| 基础 | Day 3 | [构建交互式 TUI 界面](docs/03.md) | ✅ |
| 基础 | Day 4 | 实现流式输出 | 🚧 |
| 进阶 | Day 5-10 | 核心工具集（Read/Write/Bash/Grep） | 📋 |
| 进阶 | Day 11-20 | Agent 循环与上下文管理 | 📋 |
| 高级 | Day 21-30 | 高级特性与优化 | 📋 |

## 🚀 快速开始

### 1. 克隆项目

```bash
git clone <your-repo-url>
cd MiniCode
```

### 2. 安装依赖

```bash
go mod tidy
```

### 3. 设置 API Key

获取智谱 GLM API Key：https://bigmodel.cn/

```bash
export OPENAI_API_KEY=你的GLM_API_KEY
```

### 4. 运行

```bash
go run .
```

或编译后运行：

```bash
go build -o minicode .
./minicode
```

## 📖 使用说明

### 交互式界面

启动后进入交互式聊天界面：

- **Enter**: 发送消息
- **Ctrl+C / Esc**: 退出程序

### 功能特性

- 交互式终端聊天界面
- 支持多轮对话上下文
- 内置 glob 文件搜索工具

### VSCode 调试

本项目包含完整的 VSCode 配置，详见 [.vscode/README.md](.vscode/README.md)

1. 打开项目
2. 按 `F5` 开始调试
3. 选择调试配置并输入问题

## 🏗️ 项目结构

```
MiniCode/
├── main.go              # 主程序入口
├── tui/
│   └── tui.go           # TUI 界面实现
├── tools/
│   └── glob.go          # glob 文件搜索工具
├── go.mod               # Go 模块文件
├── go.sum               # 依赖锁定文件
├── docs/                # 文档目录
└── .vscode/             # VSCode 配置
```

## 🔧 技术栈

- **语言**: Go 1.25.5
- **AI SDK**: [Fantasy](https://charm.land/fantasy) v0.5.5
- **LLM**: 智谱 GLM-4-Flash
- **接口**: OpenAI 兼容 API

## 📚 核心概念

### Fantasy 库

Fantasy 是一个统一的 AI Agent SDK，提供：
- 统一的 LLM 调用接口
- 支持多个 Provider（Anthropic、OpenAI、Google 等）
- 工具（Tool）系统
- 流式输出

### OpenAI 兼容接口

通过 `openaicompat` Provider，可以连接任何支持 OpenAI API 格式的服务：
- 智谱 GLM
- 阿里通义千问
- 百度文心一言
- 本地部署的模型（如 Ollama）

## 🎓 学习资源

- [Fantasy 官方文档](https://charm.land/fantasy)
- [Bubble Tea 教程](https://github.com/charmbracelet/bubbletea)
- [智谱 GLM 文档](https://bigmodel.cn/dev/api)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License
