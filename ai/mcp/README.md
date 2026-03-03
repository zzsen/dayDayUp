# MCP (Model Context Protocol)

## 什么是 MCP

MCP（Model Context Protocol，模型上下文协议）是由 Anthropic 推出的一个**开放标准协议**，用于在 LLM 应用与外部数据源、工具之间建立安全、标准化的连接。

可以将 MCP 理解为 **AI 领域的 USB-C 接口**：正如 USB-C 为各种外设提供了统一接口，MCP 为 AI 模型连接各种外部系统提供了统一协议。

> 当前协议版本：**2025-11-25**
>
> 官方文档：https://modelcontextprotocol.io

---

## 核心架构

MCP 采用 **Client-Host-Server** 架构，基于 JSON-RPC 2.0 进行通信：

```
┌──────────────────────────────────────────────┐
│                  MCP Host                    │
│           (如 Claude Desktop、Cursor)        │
│                                              │
│  ┌──────────┐  ┌──────────┐  ┌──────────┐    │
│  │ MCP      │  │ MCP      │  │ MCP      │    │
│  │ Client A │  │ Client B │  │ Client C │    │
│  └────┬─────┘  └────┬─────┘  └────┬─────┘    │
│       │              │              │        │
└───────┼──────────────┼──────────────┼────────┘
        │              │              │
   ┌────▼─────┐  ┌────▼─────┐  ┌────▼─────┐
   │ MCP      │  │ MCP      │  │ MCP      │
   │ Server A │  │ Server B │  │ Server C │
   │ (本地)    │  │ (本地)    │  │ (远程)   │
   └──────────┘  └──────────┘  └──────────┘
```

### 三个角色

| 角色 | 说明 | 举例 |
|------|------|------|
| **Host** | 发起连接的 LLM 应用，协调管理多个 Client | Claude Desktop、Cursor、VS Code |
| **Client** | 在 Host 内部与 Server 保持 1:1 连接的组件 | Host 内置的协议连接器 |
| **Server** | 对外提供上下文和能力的程序 | 文件系统 Server、数据库 Server、API Server |

---

## 核心能力

MCP Server 向 Client 暴露三种核心原语（Primitives）：

### 1. Tools（工具）

Tools 是 **AI 模型可以主动调用的函数**，用于执行操作和产生副作用。类似于 REST API 中的 POST 端点。

**典型用途：**
- 查询数据库
- 调用外部 API
- 文件操作
- 执行计算

**协议流程：**
1. Client 通过 `tools/list` 发现可用工具
2. AI 模型决定调用某个工具
3. Client 通过 `tools/call` 执行工具调用
4. Server 返回执行结果

### 2. Resources（资源）

Resources 是**被动的、只读的数据源**，用于为 LLM 提供上下文信息。类似于 REST API 中的 GET 端点。

**典型用途：**
- 文件内容
- 数据库 Schema
- API 文档
- 系统信息

**两种形式：**
- **静态资源**：固定 URI，如 `docs://readme`
- **动态资源**：URI 模板，如 `users://{id}/profile`

### 3. Prompts（提示模板）

Prompts 是**预构建的指令模板**，引导模型更好地使用特定工具和资源。

**典型用途：**
- 代码审查模板
- SQL 查询构建器引导
- 特定领域的交互模式

### 能力对比

| 特性 | Tools | Resources | Prompts |
|------|-------|-----------|---------|
| 触发方式 | AI 模型主动调用 | Client 拉取 / 订阅 | 用户选择 |
| 副作用 | 有 | 无（只读） | 无 |
| 类比 | POST 端点 | GET 端点 | 交互模板 |
| 控制权 | 模型控制 | 应用控制 | 用户控制 |

---

## 传输层（Transport）

MCP 支持多种传输方式，核心区别在于**通信方式**和**适用场景**。

### Stdio（标准输入/输出）

```
Host (Cursor/Claude) ──── 启动子进程 ──── MCP Server
                         stdin/stdout 双向通信
```

- **工作方式**：Host 直接把 MCP Server 作为**子进程**启动，通过进程的 stdin 发请求、stdout 收响应
- **生命周期**：Host 启动时拉起进程，关闭时杀掉进程，一对一绑定
- **网络**：不需要网络，纯进程间通信
- **配置方式**：指定可执行文件路径（`command` + `args`）

```json
{
  "mcpServers": {
    "demo": {
      "command": "/path/to/mcp-server",
      "args": []
    }
  }
}
```

### Streamable HTTP

```
Host (Cursor/Claude) ──── HTTP POST ──── MCP Server（监听端口）
                          网络请求通信
```

- **工作方式**：MCP Server 独立运行并监听一个 HTTP 端口，Host 通过 HTTP 请求与之通信
- **生命周期**：Server 独立启动和管理，可以多个 Client 同时连接
- **网络**：需要网络（本地 localhost 或远程公网）
- **配置方式**：指定 URL 地址
- 取代了早期的 HTTP+SSE 方案（2025-11-25 版本起）
- 支持无状态和有状态两种模式

```json
{
  "mcpServers": {
    "demo": {
      "url": "http://localhost:3000/mcp"
    }
  }
}
```

### 自定义传输

- MCP 的传输层是可插拔的
- 可以基于 WebSocket、gRPC 等自行实现

### Stdio vs HTTP 对比

| 对比项 | Stdio | Streamable HTTP |
|--------|-------|-----------------|
| 进程管理 | Host 自动管理（启停） | 需要自己启动和维护 |
| 并发连接 | 1 对 1，单客户端 | 多客户端同时连接 |
| 网络需求 | 不需要 | 需要（本地或远程） |
| 延迟 | 极低（进程内管道） | 略高（HTTP 开销） |
| 部署场景 | 本地开发、桌面工具 | 远程服务、团队共享、ChatGPT 接入 |
| 认证 | 不需要（本地进程天然安全） | 通常需要（Token / OAuth） |
| Go 代码 | `server.Run(ctx, &mcp.StdioTransport{})` | `http.ListenAndServe(addr, handler)` |

### 如何选择

- **本地自己用**（Cursor、Claude Desktop）→ 用 **Stdio**，零配置、Host 自动管理进程
- **团队共享 / 远程部署 / ChatGPT 接入** → 用 **HTTP**，独立运行、多人共用

---

## 连接生命周期

```
Client                          Server
  │                               │
  │  ── initialize ──────────►    │   (1) 发送能力和协议版本
  │  ◄─── initialize result ──   │   (2) 返回服务器能力
  │  ── initialized ─────────►   │   (3) 确认初始化完成
  │                               │
  │  ── tools/list ──────────►    │   (4) 发现可用工具
  │  ◄─── tools list result ──   │
  │                               │
  │  ── tools/call ──────────►    │   (5) 调用工具
  │  ◄─── tool result ────────   │
  │                               │
  │  ── resources/read ──────►    │   (6) 读取资源
  │  ◄─── resource contents ──   │
  │                               │
  │         ...                   │
  │                               │
  │  ── shutdown ────────────►    │   (n) 关闭连接
  │                               │
```

---

## 安全设计原则

MCP 规范强调以下安全要点：

1. **用户同意**：所有操作需经用户明确授权
2. **数据隐私**：Server 应最小化数据暴露
3. **工具安全**：工具调用应有清晰的权限边界
4. **最小权限**：Server 只应请求必要的访问权限

---

## 生态与工具

### 主流 MCP Host

| Host | 说明 |
|------|------|
| [Claude Desktop](https://claude.ai) | Anthropic 官方桌面客户端 |
| [Cursor](https://cursor.com) | AI 代码编辑器 |
| [VS Code](https://code.visualstudio.com) | 通过 Copilot Chat 支持 MCP |
| [Windsurf](https://codeium.com/windsurf) | AI 代码编辑器 |

### 主流 SDK

| 语言 | SDK | 地址 |
|------|-----|------|
| TypeScript | 官方 SDK | https://github.com/modelcontextprotocol/typescript-sdk |
| Python | 官方 SDK | https://github.com/modelcontextprotocol/python-sdk |
| Go | 官方 SDK | https://github.com/modelcontextprotocol/go-sdk |
| Go | mcp-go（社区） | https://github.com/mark3labs/mcp-go |
| Java | 官方 SDK | https://github.com/modelcontextprotocol/java-sdk |

### 调试工具

- **MCP Inspector**：官方提供的可视化调试工具，可测试 Server 的 Tools / Resources / Prompts

---

## 客户端接入配置

以下以本项目的 Go 示例为例，说明各客户端如何接入自定义 MCP Server。

> **前提**：先编译出可执行文件
> ```bash
> cd ai/mcp/example
> go build -o mcp-server.exe .
> ```

### Claude Desktop

**配置文件路径：**

| 系统 | 路径 |
|------|------|
| Windows | `%APPDATA%\Claude\claude_desktop_config.json` |
| macOS | `~/Library/Application Support/Claude/claude_desktop_config.json` |
| Linux | `~/.config/Claude/claude_desktop_config.json` |

也可以通过 Claude Desktop 菜单进入：**Settings → Developer → Edit Config**

**配置示例：**

```json
{
  "mcpServers": {
    "weather-demo": {
      "command": "D:/path/to/ai/mcp/example/mcp-server.exe",
      "args": [],
      "env": {}
    }
  }
}
```

**生效步骤：**
1. 保存配置文件
2. **完全退出** Claude Desktop（不是最小化，是退出进程）
3. 重新启动 Claude Desktop
4. 在聊天输入框看到 🔨 图标即表示 MCP Server 加载成功

---

### Cursor

Cursor 支持两种配置方式：

#### 方式一：UI 配置

1. 打开 Cursor Settings（`Ctrl + ,`）
2. 导航到 **Tools & MCP**
3. 点击 **Add new MCP server**
4. 填写 Name、Type（stdio）、Command 等信息

#### 方式二：JSON 文件配置

在项目根目录创建 `.cursor/mcp.json`（项目级）或 `~/.cursor/mcp.json`（全局）：

```json
{
  "mcpServers": {
    "weather-demo": {
      "command": "D:/path/to/ai/mcp/example/mcp-server.exe",
      "args": []
    }
  }
}
```

**远程 HTTP Server 配置：**

```json
{
  "mcpServers": {
    "remote-server": {
      "url": "http://localhost:3000/mcp",
      "headers": {
        "Authorization": "Bearer your-token"
      }
    }
  }
}
```

> 修改配置后需要**重启 Cursor** 才能生效。

---

### VS Code（GitHub Copilot）

在项目中创建 `.vscode/mcp.json`：

```json
{
  "servers": {
    "weather-demo": {
      "type": "stdio",
      "command": "D:/path/to/ai/mcp/example/mcp-server.exe",
      "args": []
    }
  }
}
```

**远程 HTTP Server：**

```json
{
  "servers": {
    "remote-server": {
      "type": "http",
      "url": "http://localhost:3000/mcp"
    }
  }
}
```

> 需要 VS Code 1.99+，且安装了 GitHub Copilot Chat 扩展。

---

### ChatGPT（OpenAI）

ChatGPT 的 MCP 接入方式与上述本地客户端不同，它要求 **Server 以 HTTPS 公网端点暴露**：

**接入步骤：**

1. **部署 MCP Server** 到公网（或使用 ngrok 暴露本地服务）
2. 将 Server 传输层改为 HTTP（非 Stdio），示例代码需调整为 `server.ServeHTTP` 或 `server.ServeStreamableHTTP`
3. 进入 ChatGPT → **Settings → Apps → Create**
4. 填写 App 名称、描述和公网 HTTPS 端点 URL
5. 配置认证方式（支持 OAuth / OpenID Connect）

**本地开发调试（ngrok）：**

```bash
# 假设你的 HTTP MCP Server 监听 3000 端口
ngrok http 3000
```

将 ngrok 生成的 HTTPS URL 填入 ChatGPT 的 App 配置中即可。

> ChatGPT 的 MCP 支持已面向所有套餐开放（Free / Plus / Business / Enterprise）。

---

### MCP Inspector（调试工具）

MCP Inspector 是官方提供的可视化调试工具，无需安装，直接通过 npx 启动：

```bash
# 启动 Inspector 并连接到你的 Server
npx @modelcontextprotocol/inspector D:/path/to/ai/mcp/example/mcp-server.exe

# 也可以连接 go run 启动的 Server
npx @modelcontextprotocol/inspector go run D:/path/to/ai/mcp/example/main.go
```

启动后在浏览器打开 `http://localhost:6274`，可以：
- 查看并测试所有 **Tools**（输入参数、查看返回结果）
- 浏览 **Resources**（读取资源内容）
- 预览 **Prompts**（测试不同参数组合）
- 监控 Server 日志和通知

**自定义端口：**

```bash
CLIENT_PORT=8080 SERVER_PORT=9000 npx @modelcontextprotocol/inspector ./mcp-server.exe
```

> 需要 Node.js 22.7+ 环境。

---

### 配置速查表

| 客户端 | 配置文件 | 传输方式 | 备注 |
|--------|----------|----------|------|
| Claude Desktop | `claude_desktop_config.json` | Stdio | 修改后需完全退出重启 |
| Cursor | `.cursor/mcp.json` | Stdio / HTTP | 支持项目级和全局配置 |
| VS Code | `.vscode/mcp.json` | Stdio / HTTP / SSE | 需 Copilot Chat 扩展 |
| ChatGPT | Web UI 配置 | HTTPS（公网） | 需公网端点 + OAuth |
| MCP Inspector | 命令行参数 | Stdio | 调试专用，`npx` 启动 |

---

## Go 语言 MCP Server 示例

本目录提供了两个功能相同的 Go MCP Server 示例，分别使用不同的 SDK 实现：

| 目录 | SDK | 说明 |
|------|-----|------|
| [example/](./example/) | `mark3labs/mcp-go`（社区） | API 更简洁，链式调用风格 |
| [example-official/](./example-official/) | `modelcontextprotocol/go-sdk`（官方） | 泛型驱动，结构体自动生成 Schema |

两个示例均实现以下功能：
- **Tool**：`get_weather` — 查询指定城市的天气信息
- **Tool**：`calculate` — 基本算术计算器
- **Resource**：`config://app` — 返回应用配置信息
- **Prompt**：`summarize` — 文本摘要提示模板

### 运行方式

```bash
# 社区 SDK 版本（仅支持 Stdio）
cd example
go mod tidy
go run main.go

# 官方 SDK 版本 — Stdio 模式（默认）
cd example-official
go mod tidy
go run main.go

# 官方 SDK 版本 — HTTP 模式（监听 3000 端口）
cd example-official
go run main.go -http :3000
```

**Stdio 模式**：Server 通过 stdin/stdout 与 MCP Client 通信，适合 Claude Desktop、Cursor 等本地客户端。

**HTTP 模式**：Server 监听指定端口，通过 Streamable HTTP 协议通信，适合远程部署和多客户端场景。Cursor 中配置 URL 为 `http://localhost:3000/mcp` 即可连接。

### 两个 SDK 的核心差异对比

| 对比项 | mcp-go（社区） | go-sdk（官方） |
|--------|----------------|----------------|
| **Tool 注册** | `mcp.NewTool()` 链式构建 + `s.AddTool()` | `mcp.AddTool()` 泛型函数，参数结构体自动推导 Schema |
| **参数定义** | 链式 API：`mcp.WithString("city", mcp.Required())` | Go 结构体 + `jsonschema` tag |
| **Handler 签名** | `func(ctx, CallToolRequest) (*CallToolResult, error)` | `func(ctx, *CallToolRequest, TypedArgs) (*CallToolResult, any, error)` |
| **Resource** | `s.AddResource(resource, handler)` | `server.AddResource(&mcp.Resource{...}, handler)` |
| **Prompt** | `s.AddPrompt(prompt, handler)` | `server.AddPrompt(&mcp.Prompt{...}, handler)` |
| **启动方式** | `server.ServeStdio(s)` | `server.Run(ctx, &mcp.StdioTransport{})` |
| **类型安全** | 运行时通过 `request.RequireString()` 提取 | 编译期泛型保证类型安全 |
| **第三个返回值** | 无 | 支持结构化输出（Structured Output） |

---

## 参考资料

- [MCP 官方规范](https://modelcontextprotocol.io/specification/2025-11-25)
- [MCP 架构概览](https://modelcontextprotocol.io/docs/learn/architecture)
- [MCP Server 概念](https://modelcontextprotocol.io/docs/learn/server-concepts)
- [mcp-go 官方文档](https://mcp-go.dev/)
- [MCP GitHub 组织](https://github.com/modelcontextprotocol)
