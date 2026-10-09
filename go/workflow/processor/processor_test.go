package processor_test

import (
	"context"
	"testing"

	"github.com/vanguard-platform/aegis-sdk-go/workflow/processor"
	_ "github.com/vanguard-platform/aegis-sdk-go/workflow/processor/api"
	_ "github.com/vanguard-platform/aegis-sdk-go/workflow/processor/data"
	_ "github.com/vanguard-platform/aegis-sdk-go/workflow/processor/gate"
)

func TestProcessorRegistry_BuiltinProcessors(t *testing.T) {
	reg := processor.GetRegistry()

	// 1. 验证内置 HTTP 处理器
	httpProc, ok := reg.Get("HTTP")
	if !ok {
		t.Fatalf("内置 HTTP 处理器应已被注册")
	}
	if httpProc.GetMetadata().Category != "api" {
		t.Errorf("HTTP 处理器分类应为 api, 得到: %s", httpProc.GetMetadata().Category)
	}

	err := httpProc.ValidateConfig(map[string]interface{}{"method": "POST"})
	if err == nil {
		t.Errorf("缺少 url 字段时应该返回校验错误")
	}

	// 2. 验证内置 SQL 处理器
	sqlProc, ok := reg.Get("SQL")
	if !ok {
		t.Fatalf("内置 SQL 处理器应已被注册")
	}
	if sqlProc.GetMetadata().Category != "data" {
		t.Errorf("SQL 处理器分类应为 data, 得到: %s", sqlProc.GetMetadata().Category)
	}

	// 3. 验证内置 QUALITY_GATE 处理器
	gateProc, ok := reg.Get("QUALITY_GATE")
	if !ok {
		t.Fatalf("内置 QUALITY_GATE 处理器应已被注册")
	}
	if gateProc.GetMetadata().Category != "gate" {
		t.Errorf("QUALITY_GATE 处理器分类应为 gate, 得到: %s", gateProc.GetMetadata().Category)
	}
}

type MockCustomProcessor struct {
	processor.BaseProcessor
}

func (m *MockCustomProcessor) Execute(ctx context.Context, execCtx *processor.ExecutionContext) (*processor.ExecutionResult, error) {
	return &processor.ExecutionResult{
		Status: "SUCCESS",
		Output: map[string]interface{}{"custom_ok": true},
	}, nil
}

func TestProcessorRegistry_DynamicExtension(t *testing.T) {
	reg := processor.GetRegistry()

	reg.Register(&MockCustomProcessor{
		BaseProcessor: processor.BaseProcessor{
			Metadata: processor.ProcessorMetadata{
				Type:     "MY_CUSTOM_RPC",
				Name:     "自定义 RPC 插件",
				Category: "rpc",
			},
		},
	})

	proc, ok := reg.Get("MY_CUSTOM_RPC")
	if !ok {
		t.Fatalf("动态扩展的插件应能直接获取")
	}

	res, err := proc.Execute(context.Background(), &processor.ExecutionContext{
		NodeType: "MY_CUSTOM_RPC",
	})
	if err != nil || res.Status != "SUCCESS" {
		t.Fatalf("插件执行失败: %v", err)
	}
}
