# MCP Server 示例 — 社区 SDK（mcp-go）

基于 [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) 构建的 MCP Server 示例，支持 **Stdio** 和 **SSE** 双模式。

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

# SSE 模式（HTTP，监听 3000 端口）
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
      "args": ["run", "D:/workSpace/open_source_project/dayDayUp/ai/mcp/example/main.go"]
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
      "command": "D:/workSpace/open_source_project/dayDayUp/ai/mcp/example/mcp-server.exe"
    }
  }
}
```

配置后**重启 Cursor** 生效。

### 方式二：SSE 模式

SSE 模式需要**先手动启动 Server**，再配置 Cursor。

**第 1 步：启动 Server**

```bash
go run main.go -http :3000
```

看到以下输出表示启动成功：

```
MCP SSE Server listening at http://localhost:3000
Cursor 配置 URL: http://localhost:3000/sse
```

**第 2 步：配置 Cursor**

```json
{
  "mcpServers": {
    "weather-demo": {
      "url": "http://localhost:3000/sse"
    }
  }
}
```

**第 3 步：Cursor 刷新连接**

在 Cursor **Settings → Tools & MCP** 中，点击该 Server 旁边的刷新按钮。

> Cursor 会先尝试 Streamable HTTP（会失败并显示一条 warning），然后**自动回退到 SSE** 并成功连接。看到绿色状态即表示连接成功，warning 可忽略。

### Claude Desktop 配置

编辑 `claude_desktop_config.json`（路径见主 README）：

```json
{
  "mcpServers": {
    "weather-demo": {
      "command": "D:/workSpace/open_source_project/dayDayUp/ai/mcp/example/mcp-server.exe",
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

### 第 1 步：定义工具和参数

在 `registerTools` 函数中添加工具定义：

```go
func registerTools(s *server.MCPServer) {
    // ... 已有的工具 ...

    timeTool := mcp.NewTool("get_time",
        mcp.WithDescription("查询当前服务器时间"),
        mcp.WithString("timezone",
            mcp.Description("时区，如 Asia/Shanghai，可选"),
        ),
    )
    s.AddTool(timeTool, handleGetTime)
}
```

### 第 2 步：编写 Handler

```go
func handleGetTime(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
    tz := request.GetString("timezone", "Local")

    loc, err := time.LoadLocation(tz)
    if err != nil {
        return mcp.NewToolResultError(fmt.Sprintf("无效时区：%s", tz)), nil
    }

    now := time.Now().In(loc).Format("2006-01-02 15:04:05 MST")
    return mcp.NewToolResultText(now), nil
}
```

### 第 3 步：重启 Server

重新运行即可，Stdio 和 SSE 模式的 Handler 代码完全通用。

### 新增 Resource / Prompt 同理

**Resource 调用链**：`mcp.NewResource()` → 编写 Handler → `s.AddResource()` 注册

**Prompt 调用链**：`mcp.NewPrompt()` → 编写 Handler → `s.AddPrompt()` 注册

## 项目结构

```
example/
├── main.go     # Server 初始化、注册、Handler（Stdio + SSE 双模式）
├── go.mod
├── go.sum
└── README.md
```

## 依赖

- Go 1.23+
- [mark3labs/mcp-go](https://github.com/mark3labs/mcp-go) v0.32.0
