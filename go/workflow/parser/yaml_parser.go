package parser

import (
	"fmt"

	"github.com/vanguard-platform/aegis-sdk-go/workflow/model"
	"gopkg.in/yaml.v3"
)

// ParseWorkflowYAML 解析 YAML 文本为工作流图结构，并规范化 variables 变量池
func ParseWorkflowYAML(content []byte) (*model.WorkflowGraph, error) {
	// 动态解析为原始 Map 以归一化 variables / vars / params 关键字
	var rawMap map[string]interface{}
	if err := yaml.Unmarshal(content, &rawMap); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workflow yaml map: %w", err)
	}

	mergedVariables := make(map[string]interface{})

	// 优先级 1: variables
	if vars, ok := rawMap["variables"].(map[string]interface{}); ok {
		for k, v := range vars {
			mergedVariables[k] = v
		}
	}
	// 优先级 2: vars (简写)
	if vars, ok := rawMap["vars"].(map[string]interface{}); ok {
		for k, v := range vars {
			if _, exists := mergedVariables[k]; !exists {
				mergedVariables[k] = v
			}
		}
	}
	// 优先级 3: params (向下兼容旧格式)
	if params, ok := rawMap["params"].(map[string]interface{}); ok {
		for k, v := range params {
			if _, exists := mergedVariables[k]; !exists {
				mergedVariables[k] = v
			}
		}
	}

	var graph model.WorkflowGraph
	if err := yaml.Unmarshal(content, &graph); err != nil {
		return nil, fmt.Errorf("failed to unmarshal workflow struct: %w", err)
	}

	graph.Variables = mergedVariables

	// 基础结构校验
	if graph.ID == "" {
		return nil, fmt.Errorf("workflow missing required field: id")
	}
	if len(graph.Nodes) == 0 {
		return nil, fmt.Errorf("workflow must contain at least one node")
	}

	return &graph, nil
}
