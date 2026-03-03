# MCP Server 示例 — 官方 SDK（go-sdk）

基于 [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) 构建的 MCP Server 示例，支持 **Stdio** 和 **Streamable HTTP** 双模式。

## 功能列表

### Tools（工具）

| 工具名 | 说明 | 参数 |
|--------|------|------|
| `get_weather` | 查询指定城市的天气信息（模拟数据） | `city`（string，必填） |
| `calculate` | 四则运算计算器 | `operation`（add/subtract/multiply/divide，必填）、`x`（number，必填）、`y`（number，必填） |

### Resources（资源）

| URI | 说明 | 类型 |
|-----|------|------|
| `config://app` | 返回应用配置信息 | `application/json` |

### Prompts（提示模板）

| 名称 | 说明 | 参数 |
|------|------|------|
| `summarize` | 对给定文本生成摘要 | `text`（必填）、`style`（可选，`brief` 或 `detailed`） |

## 启动方式

```bash
# 安装依赖
go mod tidy

# Stdio 模式（默认，推荐用于 Cursor / Claude Desktop）
go run main.go

# Streamable HTTP 模式（监听 3000 端口）
go run main.go -http :3000
```

## Cursor 配置

### 方式一：Stdio 模式（推荐）

Stdio 模式最简单稳定，Cursor 自动管理进程生命周期，无需手动启动 Server。

编辑 `~/.cursor/mcp.json`（全局）或项目下 `.cursor/mcp.json`：

```json
{
  "mcpServers": {
    "weather-demo": {
      "command": "go",
      "args": ["run", "D:/workSpace/open_source_project/dayDayUp/ai/mcp/example-official/main.go"]
    }
  }
}
```

也可以先编译再配置（启动更快）：

```bash
go build -o mcp-server.exe .
```

```json
{
  "mcpServers": {
    "weather-demo": {
      "command": "D:/workSpace/open_source_project/dayDayUp/ai/mcp/example-official/mcp-server.exe"
    }
  }
}
```

配置后**重启 Cursor** 生效。

### 方式二：Streamable HTTP 模式

> ⚠️ **兼容性说明**：Streamable HTTP 是 MCP 2025-11-25 规范的新传输方式，部分客户端（如 Cursor）可能尚未完全兼容官方 Go SDK 的实现。如遇连接问题，建议使用 Stdio 模式，或使用社区 SDK（`example/`）的 SSE 模式。

**第 1 步：启动 Server**

```bash
go run main.go -http :3000
```

**第 2 步：配置 Cursor**

```json
{
  "mcpServers": {
    "weather-demo": {
      "url": "http://localhost:3000/mcp"
    }
  }
}
```

**第 3 步：验证**

可用命令行验证 Server 是否正常工作：

```powershell
Invoke-WebRequest -Uri "http://localhost:3000/mcp" -Method POST `
  -ContentType "application/json" `
  -Headers @{Accept="application/json, text/event-stream"} `
  -Body '{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-11-25","capabilities":{},"clientInfo":{"name":"test","version":"1.0.0"}}}' `
  -UseBasicParsing
```

返回 200 且包含 `serverInfo` 即表示 Server 正常。

### Claude Desktop 配置

编辑 `claude_desktop_config.json`（路径见主 README）：

```json
{
  "mcpServers": {
    "weather-demo": {
      "command": "D:/workSpace/open_source_project/dayDayUp/ai/mcp/example-official/mcp-server.exe",
      "args": []
    }
  }
}
```

### MCP Inspector（调试）

```bash
npx @modelcontextprotocol/inspector go run main.go
```

## 如何新增功能

以新增一个 `get_time` 工具（查询当前时间）为例：

### 第 1 步：定义参数结构体

官方 SDK 通过 Go 结构体 + `jsonschema` tag 自动生成参数 Schema：

```go
type TimeArgs struct {
    Timezone string `json:"timezone" jsonschema:"时区，如 Asia/Shanghai"`
}
```

### 第 2 步：编写 Handler

Handler 签名为泛型函数，第三个参数是自动反序列化后的强类型结构体：

```go
func handleGetTime(ctx context.Context, req *mcp.CallToolRequest, args TimeArgs) (*mcp.CallToolResult, any, error) {
    tz := args.Timezone
    if tz == "" {
        tz = "Local"
    }

    loc, err := time.LoadLocation(tz)
    if err != nil {
        return &mcp.CallToolResult{
            Content: []mcp.Content{
                &mcp.TextContent{Text: fmt.Sprintf("无效时区：%s", tz)},
            },
            IsError: true,
        }, nil, nil
    }

    now := time.Now().In(loc).Format("2006-01-02 15:04:05 MST")
    return &mcp.CallToolResult{
        Content: []mcp.Content{
            &mcp.TextContent{Text: now},
        },
    }, nil, nil
}
```

### 第 3 步：注册工具

在 `registerTools` 函数中调用 `mcp.AddTool`：

```go
func registerTools(server *mcp.Server) {
    // ... 已有的工具 ...

    mcp.AddTool(server, &mcp.Tool{
        Name:        "get_time",
        Description: "查询当前服务器时间",
    }, handleGetTime)
}
```

### 第 4 步：重启 Server

重新运行即可，Handler 代码对 Stdio / HTTP 两种模式完全通用。

### 新增 Resource / Prompt 同理

**Resource 调用链**：定义 `&mcp.Resource{...}` → 编写 Handler `func(ctx, *ReadResourceRequest) (*ReadResourceResult, error)` → `server.AddResource()` 注册

**Prompt 调用链**：定义 `&mcp.Prompt{...}` → 编写 Handler `func(ctx, *GetPromptRequest) (*GetPromptResult, error)` → `server.AddPrompt()` 注册

### 与社区 SDK（mcp-go）的关键区别

| 步骤 | 社区 SDK（mcp-go） | 官方 SDK（go-sdk） |
|------|---------------------|---------------------|
| 定义参数 | 链式 API：`mcp.WithString("city", mcp.Required())` | Go 结构体 + `jsonschema` tag |
| 提取参数 | 运行时：`request.RequireString("city")` | 编译期泛型：直接用 `args.City` |
| Handler 返回值 | 2 个：`(*CallToolResult, error)` | 3 个：`(*CallToolResult, any, error)` |
| 注册方式 | `s.AddTool(tool, handler)` | `mcp.AddTool(server, tool, handler)` |
| HTTP 传输 | SSE（`/sse` + `/message`） | Streamable HTTP（单路径 POST） |

## 项目结构

```
example-official/
├── main.go     # Server 初始化、注册、Handler（Stdio + HTTP 双模式）
├── go.mod
├── go.sum
└── README.md
```

## 依赖

- Go 1.23+
- [modelcontextprotocol/go-sdk](https://github.com/modelcontextprotocol/go-sdk) v1.4.0
