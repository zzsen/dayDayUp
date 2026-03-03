package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"
	"net/http"

	"github.com/modelcontextprotocol/go-sdk/mcp"
)

var httpAddr = flag.String("http", "", "HTTP 监听地址（如 :3000），不指定则使用 Stdio 模式")

func main() {
	flag.Parse()

	server := mcp.NewServer(
		&mcp.Implementation{
			Name:    "Weather Demo Server (Official SDK)",
			Version: "1.0.0",
		},
		&mcp.ServerOptions{
			Instructions: "天气查询和计算工具演示服务器，使用官方 Go SDK 构建",
		},
	)

	registerTools(server)
	registerResources(server)
	registerPrompts(server)

	if *httpAddr != "" {
		mcpHandler := mcp.NewStreamableHTTPHandler(func(r *http.Request) *mcp.Server {
			return server
		}, nil)
		mux := http.NewServeMux()
		mux.Handle("/mcp", mcpHandler)
		mux.Handle("/", mcpHandler)
		log.Printf("MCP HTTP Server listening at http://localhost%s/mcp", *httpAddr)
		log.Fatal(http.ListenAndServe(*httpAddr, mux))
	} else {
		log.Println("MCP Server running in Stdio mode")
		if err := server.Run(context.Background(), &mcp.StdioTransport{}); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	}
}

// --- Tool 参数与返回值结构体 ---

// WeatherArgs get_weather 工具的输入参数
type WeatherArgs struct {
	City string `json:"city" jsonschema:"城市名称"`
}

// CalcArgs calculate 工具的输入参数
type CalcArgs struct {
	Operation string  `json:"operation" jsonschema:"运算类型"`
	X         float64 `json:"x" jsonschema:"第一个操作数"`
	Y         float64 `json:"y" jsonschema:"第二个操作数"`
}

// registerTools 注册所有工具
//
// 包含以下工具：
// 1. get_weather - 查询指定城市的天气信息
// 2. calculate - 简单算术计算器
func registerTools(server *mcp.Server) {
	mcp.AddTool(server, &mcp.Tool{
		Name:        "get_weather",
		Description: "查询指定城市的天气信息",
	}, handleGetWeather)

	mcp.AddTool(server, &mcp.Tool{
		Name:        "calculate",
		Description: "执行基本算术运算（加减乘除）",
		InputSchema: json.RawMessage(`{
			"type": "object",
			"properties": {
				"operation": {
					"type": "string",
					"description": "运算类型",
					"enum": ["add", "subtract", "multiply", "divide"]
				},
				"x": {
					"type": "number",
					"description": "第一个操作数"
				},
				"y": {
					"type": "number",
					"description": "第二个操作数"
				}
			},
			"required": ["operation", "x", "y"],
			"additionalProperties": false
		}`),
	}, handleCalculate)
}

// registerResources 注册所有资源
//
// 包含以下资源：
// 1. config://app - 应用配置信息（静态资源）
func registerResources(server *mcp.Server) {
	server.AddResource(&mcp.Resource{
		Name:     "应用配置",
		URI:      "config://app",
		MIMEType: "application/json",
	}, handleReadAppConfig)
}

// registerPrompts 注册所有提示模板
//
// 包含以下模板：
// 1. summarize - 文本摘要提示模板
func registerPrompts(server *mcp.Server) {
	server.AddPrompt(&mcp.Prompt{
		Name:        "summarize",
		Description: "对给定文本生成摘要",
		Arguments: []*mcp.PromptArgument{
			{
				Name:        "text",
				Description: "需要进行摘要的文本内容",
				Required:    true,
			},
			{
				Name:        "style",
				Description: "摘要风格：brief（简短）或 detailed（详细），默认 brief",
			},
		},
	}, handleSummarize)
}

// --- Tool Handlers ---

// handleGetWeather 处理天气查询请求
//
// 模拟返回指定城市的天气数据（温度、湿度、天气状况）。
// 实际场景中应替换为真实天气 API 调用。
func handleGetWeather(ctx context.Context, req *mcp.CallToolRequest, args WeatherArgs) (*mcp.CallToolResult, any, error) {
	weatherData := map[string]any{
		"city":        args.City,
		"temperature": 15 + rand.Intn(20),
		"humidity":    40 + rand.Intn(40),
		"condition":   randomCondition(),
		"unit":        "℃",
	}

	data, _ := json.MarshalIndent(weatherData, "", "  ")
	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: string(data)},
		},
	}, nil, nil
}

// handleCalculate 处理算术计算请求
func handleCalculate(ctx context.Context, req *mcp.CallToolRequest, args CalcArgs) (*mcp.CallToolResult, any, error) {
	var result float64
	switch args.Operation {
	case "add":
		result = args.X + args.Y
	case "subtract":
		result = args.X - args.Y
	case "multiply":
		result = args.X * args.Y
	case "divide":
		if args.Y == 0 {
			return &mcp.CallToolResult{
				Content: []mcp.Content{
					&mcp.TextContent{Text: "错误：除数不能为零"},
				},
				IsError: true,
			}, nil, nil
		}
		result = args.X / args.Y
	default:
		return &mcp.CallToolResult{
			Content: []mcp.Content{
				&mcp.TextContent{Text: fmt.Sprintf("不支持的运算类型：%s", args.Operation)},
			},
			IsError: true,
		}, nil, nil
	}

	return &mcp.CallToolResult{
		Content: []mcp.Content{
			&mcp.TextContent{Text: fmt.Sprintf("%.4f", result)},
		},
	}, nil, nil
}

// --- Resource Handlers ---

// handleReadAppConfig 返回应用配置信息
func handleReadAppConfig(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
	config := map[string]any{
		"server_name": "Weather Demo Server (Official SDK)",
		"version":     "1.0.0",
		"sdk":         "github.com/modelcontextprotocol/go-sdk",
		"protocol":    "MCP 2025-11-25",
		"features":    []string{"tools", "resources", "prompts"},
	}

	data, _ := json.MarshalIndent(config, "", "  ")
	return &mcp.ReadResourceResult{
		Contents: []*mcp.ResourceContents{
			{
				URI:      "config://app",
				MIMEType: "application/json",
				Text:     string(data),
			},
		},
	}, nil
}

// --- Prompt Handlers ---

// handleSummarize 生成文本摘要提示
func handleSummarize(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	text := req.Params.Arguments["text"]
	if text == "" {
		return nil, fmt.Errorf("参数 text 不能为空")
	}

	style := req.Params.Arguments["style"]
	if style == "" {
		style = "brief"
	}

	var instruction string
	switch style {
	case "detailed":
		instruction = "请对以下文本进行详细摘要，包含主要观点、关键细节和结论："
	default:
		instruction = "请用 1-2 句话简要概括以下文本的核心内容："
	}

	return &mcp.GetPromptResult{
		Description: "文本摘要",
		Messages: []*mcp.PromptMessage{
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: instruction},
			},
			{
				Role:    "user",
				Content: &mcp.TextContent{Text: text},
			},
		},
	}, nil
}

// --- Helpers ---

func randomCondition() string {
	conditions := []string{"晴", "多云", "阴", "小雨", "大雨", "雷阵雨", "雪", "雾"}
	return conditions[rand.Intn(len(conditions))]
}
