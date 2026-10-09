package workflow

import (
	"context"
	"fmt"
	"os"

	"github.com/vanguard-platform/aegis-sdk-go/workflow/engine"
	"github.com/vanguard-platform/aegis-sdk-go/workflow/model"
	"github.com/vanguard-platform/aegis-sdk-go/workflow/parser"
)

// RunFile 从本地文件加载并执行 DAG 工作流（脚手架与测试框架最常用的极简方法）
func RunFile(ctx context.Context, filePath string, overrideVars map[string]interface{}) (*model.WorkflowExecutionResult, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to read workflow file %s: %w", filePath, err)
	}
	return RunYAML(ctx, content, overrideVars)
}

// RunYAML 从 YAML 文本解析并执行 DAG 工作流
func RunYAML(ctx context.Context, yamlContent []byte, overrideVars map[string]interface{}) (*model.WorkflowExecutionResult, error) {
	graph, err := parser.ParseWorkflowYAML(yamlContent)
	if err != nil {
		return nil, fmt.Errorf("failed to parse workflow yaml: %w", err)
	}
	return RunGraph(ctx, graph, overrideVars)
}

// RunGraph 直接执行已构造好的 WorkflowGraph 对象
func RunGraph(ctx context.Context, graph *model.WorkflowGraph, overrideVars map[string]interface{}) (*model.WorkflowExecutionResult, error) {
	dag := engine.NewDAGEngine()
	return dag.ExecuteGraph(ctx, graph, overrideVars)
}
