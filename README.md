# MiniCode

一个基于 Fantasy 库的简单 AI Agent 项目，使用智谱 GLM API。

## 🎯 项目简介

MiniCode 是一个学习项目，展示如何使用 [Fantasy](https://charm.land/fantasy) 库创建 AI Agent，并通过 OpenAI 兼容接口连接到智谱 GLM 模型。

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
go run . "你好，介绍一下你自己"
```

或编译后运行：

```bash
go build -o minicode .
./minicode "你的问题"
```

## 📖 使用说明

### 基本用法

```bash
# 直接运行
go run . "你的问题"

# 示例
go run . "用 Go 语言写一个 Hello World"
go run . "解释一下什么是 goroutine"
```

### VSCode 调试

本项目包含完整的 VSCode 配置，详见 [.vscode/README.md](.vscode/README.md)

1. 打开项目
2. 按 `F5` 开始调试
3. 选择调试配置并输入问题

## 🏗️ 项目结构

```
MiniCode/
├── main.go              # 主程序
├── go.mod              # Go 模块文件
├── go.sum              # 依赖锁定文件
├── docs/               # 文档目录
│   ├── 01.md
│   └── phase01-hello-agent.md
└── .vscode/            # VSCode 配置
    ├── launch.json     # 调试配置
    ├── tasks.json      # 任务配置
    ├── settings.json   # 工作区设置
    ├── extensions.json # 推荐扩展
    └── README.md       # VSCode 配置说明
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
- [智谱 GLM 文档](https://bigmodel.cn/dev/api)
- [Phase 01: Hello Agent](docs/phase01-hello-agent.md)

## 🤝 贡献

欢迎提交 Issue 和 Pull Request！

## 📄 许可证

MIT License

## 🙏 致谢

- [Charmbracelet](https://charm.sh/) - Fantasy 库的开发者
- [智谱 AI](https://bigmodel.cn/) - 提供 GLM 模型
