package parser_test

import (
	"testing"

	"github.com/vanguard-platform/aegis-sdk-go/workflow/parser"
)

func TestParseWorkflowYAML(t *testing.T) {
	yamlContent := []byte(`
id: order_flow
name: "订单核销"
variables:
  userId: "u123"
  amount: 99.5
nodes:
  - id: step_1
    name: "创建订单"
    type: HTTP
    config:
      url: "/orders"
`)
	graph, err := parser.ParseWorkflowYAML(yamlContent)
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if graph.ID != "order_flow" {
		t.Errorf("期望 id order_flow, 实际: %s", graph.ID)
	}
	if graph.Variables["userId"] != "u123" {
		t.Errorf("期望 userId=u123, 实际: %v", graph.Variables["userId"])
	}
}
