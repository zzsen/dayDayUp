package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math/rand"

	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

var httpAddr = flag.String("http", "", "HTTP 监听地址（如 :3000），不指定则使用 Stdio 模式")

func main() {
	flag.Parse()

	s := server.NewMCPServer(
		"Weather Demo Server",
		"1.0.0",
		server.WithToolCapabilities(false),
		server.WithResourceCapabilities(true, false),
		server.WithPromptCapabilities(false),
		server.WithRecovery(),
	)

	registerTools(s)
	registerResources(s)
	registerPrompts(s)

	if *httpAddr != "" {
		sseServer := server.NewSSEServer(s, server.WithBaseURL("http://localhost"+*httpAddr))
		log.Printf("MCP SSE Server listening at http://localhost%s", *httpAddr)
		log.Printf("Cursor 配置 URL: http://localhost%s/sse", *httpAddr)
		if err := sseServer.Start(*httpAddr); err != nil {
			log.Fatalf("Server error: %v", err)
		}
	} else {
		if err := server.ServeStdio(s); err != nil {
			fmt.Printf("Server error: %v\n", err)
		}
	}
}

// registerTools 注册所有工具
//
// 包含以下工具：
// 1. get_weather - 查询指定城市的天气信息
// 2. calculate - 简单算术计算器
func registerTools(s *server.MCPServer) {
	weatherTool := mcp.NewTool("get_weather",
		mcp.WithDescription("查询指定城市的天气信息"),
		mcp.WithString("city",
			mcp.Required(),
			mcp.Description("城市名称，如：北京、上海、广州"),
		),
	)
	s.AddTool(weatherTool, handleGetWeather)

	calcTool := mcp.NewTool("calculate",
		mcp.WithDescription("执行基本算术运算"),
		mcp.WithString("operation",
			mcp.Required(),
			mcp.Description("运算类型"),
			mcp.Enum("add", "subtract", "multiply", "divide"),
		),
		mcp.WithNumber("x",
			mcp.Required(),
			mcp.Description("第一个操作数"),
		),
		mcp.WithNumber("y",
			mcp.Required(),
			mcp.Description("第二个操作数"),
		),
	)
	s.AddTool(calcTool, handleCalculate)
}

// registerResources 注册所有资源
//
// 包含以下资源：
// 1. config://app - 应用配置信息（静态资源）
func registerResources(s *server.MCPServer) {
	appConfig := mcp.NewResource(
		"config://app",
		"应用配置",
		mcp.WithResourceDescription("返回当前应用的配置信息"),
		mcp.WithMIMEType("application/json"),
	)
	s.AddResource(appConfig, handleReadAppConfig)
}

// registerPrompts 注册所有提示模板
//
// 包含以下模板：
// 1. summarize - 文本摘要提示模板
func registerPrompts(s *server.MCPServer) {
	summarizePrompt := mcp.NewPrompt("summarize",
		mcp.WithPromptDescription("对给定文本生成摘要"),
		mcp.WithArgument("text",
			mcp.ArgumentDescription("需要进行摘要的文本内容"),
			mcp.RequiredArgument(),
		),
		mcp.WithArgument("style",
			mcp.ArgumentDescription("摘要风格：brief（简短）或 detailed（详细），默认 brief"),
		),
	)
	s.AddPrompt(summarizePrompt, handleSummarize)
}

// --- Tool Handlers ---

// handleGetWeather 处理天气查询请求
//
// 模拟返回指定城市的天气数据（温度、湿度、天气状况）。
// 实际场景中应替换为真实天气 API 调用。
func handleGetWeather(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	city, err := request.RequireString("city")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	weatherData := map[string]any{
		"city":        city,
		"temperature": 15 + rand.Intn(20),
		"humidity":    40 + rand.Intn(40),
		"condition":   randomCondition(),
		"unit":        "℃",
	}

	data, _ := json.MarshalIndent(weatherData, "", "  ")
	return mcp.NewToolResultText(string(data)), nil
}

// handleCalculate 处理算术计算请求
func handleCalculate(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	op, err := request.RequireString("operation")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	x, err := request.RequireFloat("x")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	y, err := request.RequireFloat("y")
	if err != nil {
		return mcp.NewToolResultError(err.Error()), nil
	}

	var result float64
	switch op {
	case "add":
		result = x + y
	case "subtract":
		result = x - y
	case "multiply":
		result = x * y
	case "divide":
		if y == 0 {
			return mcp.NewToolResultError("除数不能为零"), nil
		}
		result = x / y
	}

	return mcp.NewToolResultText(fmt.Sprintf("%.4f", result)), nil
}

// --- Resource Handlers ---

// handleReadAppConfig 返回应用配置信息
func handleReadAppConfig(ctx context.Context, request mcp.ReadResourceRequest) ([]mcp.ResourceContents, error) {
	config := map[string]any{
		"server_name": "Weather Demo Server",
		"version":     "1.0.0",
		"protocol":    "MCP 2025-11-25",
		"features":    []string{"tools", "resources", "prompts"},
	}

	data, _ := json.MarshalIndent(config, "", "  ")
	return []mcp.ResourceContents{
		mcp.TextResourceContents{
			URI:      "config://app",
			MIMEType: "application/json",
			Text:     string(data),
		},
	}, nil
}

// --- Prompt Handlers ---

// handleSummarize 生成文本摘要提示
func handleSummarize(ctx context.Context, request mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
	text := request.Params.Arguments["text"]
	if text == "" {
		return nil, fmt.Errorf("参数 text 不能为空")
	}

	style := request.Params.Arguments["style"]
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

	return mcp.NewGetPromptResult(
		"文本摘要",
		[]mcp.PromptMessage{
			mcp.NewPromptMessage(
				mcp.RoleUser,
				mcp.NewTextContent(instruction),
			),
			mcp.NewPromptMessage(
				mcp.RoleUser,
				mcp.NewTextContent(text),
			),
		},
	), nil
}

// --- Helpers ---

func randomCondition() string {
	conditions := []string{"晴", "多云", "阴", "小雨", "大雨", "雷阵雨", "雪", "雾"}
	return conditions[rand.Intn(len(conditions))]
}
