# CLAUDE.md

本文件为 Claude Code (claude.ai/code) 在此仓库中工作时提供指导。

## 项目概述

JVMS (JDK Version Manager) 是一个基于 Windows 的 JDK 版本管理工具，使用 Go 语言编写。它允许用户在 Windows 系统上安装、切换和管理多个 JDK 版本，通过符号链接实现高效的版本切换。

## 构建和开发命令

### 构建应用程序
```bash
# 为 Windows AMD64 构建
GOOS=windows GOARCH=amd64 go build -ldflags "-s -w" -o jvms.exe

# 带版本信息构建（用于 CI）
go build -ldflags "-s -w -X main.version=<version>" -o jvms.exe

# 整理依赖
go mod tidy
```

### 测试
```bash
# 运行所有测试
go test ./...

# 运行详细输出的测试
go test -v ./...

# 运行特定包的测试
go test ./internal/cmdCli
```

### 开发工作流
```bash
# 本地运行应用程序
go run main.go <command>

# 示例：列出已安装的 JDK
go run main.go list

# 示例：显示可用的 JDK 版本
go run main.go rls
```

## 架构概述

### 核心组件

**主入口点 (`main.go`)**
- 使用 `github.com/codegangsta/cli` 初始化 CLI 应用程序
- 通过 JSON 存储管理配置加载/保存
- 设置代理配置和目录结构

**命令系统 (`internal/cmdCli/`)**
- 每个命令在独立文件中实现（init.go, install.go, list.go 等）
- 通过 `cmds.go` 注册命令，聚合所有可用命令
- 所有命令接收共享的 `*entity.Config` 进行状态管理

**配置管理 (`internal/entity/`)**
- `Config` 结构体管理 JAVA_HOME、当前 JDK 版本、下载源和代理设置
- 配置以 JSON 格式持久化在应用程序目录中

**工具模块 (`utils/`)**
- `file/`: 文件系统操作和路径管理
- `jdk/`: JDK 检测、验证和供应商特定处理（包括 Azul 支持）
- `web/`: 支持代理的 HTTP 客户端，用于下载 JDK 包

### 关键设计模式

**基于符号链接的版本切换**
- 在 `jvms init` 期间在 JAVA_HOME 中创建单个符号链接
- 版本切换更新符号链接目标而不是修改 PATH
- 符号链接创建/修改需要管理员权限

**下载索引系统**
- 使用远程 JSON 索引 (`jdkdlindex.json`) 跟踪可用的 JDK 版本
- 支持企业环境的自定义下载服务器
- 默认索引托管在 GitHub releases

**存储目录结构**
- `store/`: 包含所有已安装的 JDK 版本（按版本命名）
- `download/`: 下载的 JDK 包的临时位置
- 配置存储为应用程序目录中的 `jvms.json`

## 重要实现细节

### Windows 特定注意事项
- 使用 `cmd /C setx` 系统范围设置环境变量
- 环境变量修改需要管理员权限
- Windows 上的符号链接操作需要提升权限

### 版本管理
- JDK 版本存储在 `store/` 目录中，文件夹名为版本号
- 当前版本在配置文件中跟踪
- 支持 Oracle JDK 和替代发行版（Amazon Corretto、Azul、IBM）

### 命令结构
每个命令遵循以下模式：
- 函数返回带有 Name、Usage、Description、Flags 和 Action 的 `*cli.Command`
- Action 函数接收 CLI 上下文并执行实际操作
- 配置通过依赖注入在所有命令间共享

## 开发注意事项

- 应用程序设计为无外部依赖的单一可执行文件
- 使用 Go 内置的网络和归档库进行下载和解压
- 配置更改在命令执行后自动持久化
- 使用 `gopkg.in/cheggaaa/pb.v1` 实现下载操作的进度条