package engine_test

import (
	"context"
	"testing"
	"time"

	"github.com/vanguard-platform/aegis-sdk-go/workflow/engine"
	wfModel "github.com/vanguard-platform/aegis-sdk-go/workflow/model"
)

func TestDAGEngine_ParallelAndDependencyExecution(t *testing.T) {
	dag := engine.NewDAGEngine()

	graph := &wfModel.WorkflowGraph{
		ID:   "wf-sdk-test",
		Name: "SDK DAG 并发核对测试",
		Nodes: []wfModel.WorkflowNode{
			{
				ID:        "node-1",
				Name:      "HTTP 下单",
				Type:      wfModel.NodeTypeHTTP,
				Config:    map[string]interface{}{"url": "http://api.mock/order/create"},
				DependsOn: []string{},
			},
			{
				ID:        "node-2",
				Name:      "HTTP 发券",
				Type:      wfModel.NodeTypeHTTP,
				Config:    map[string]interface{}{"url": "http://api.mock/coupon/issue"},
				DependsOn: []string{},
			},
			{
				ID:        "node-3",
				Name:      "SQL 核算对账",
				Type:      wfModel.NodeTypeSQL,
				Config:    map[string]interface{}{"sql": "SELECT 1"},
				DependsOn: []string{"node-1", "node-2"},
			},
			{
				ID:        "node-4",
				Name:      "门禁放行",
				Type:      wfModel.NodeTypeQualityGate,
				Config:    map[string]interface{}{"rule": "zero_loss"},
				DependsOn: []string{"node-3"},
			},
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	res, err := dag.ExecuteGraph(ctx, graph, map[string]interface{}{"biz_id": "BIZ-12345"})
	if err != nil {
		t.Fatalf("DAG 执行失败: %v", err)
	}

	if res.Status != wfModel.WorkflowStatusSuccess {
		t.Fatalf("期望状态 SUCCESS, 实际: %s", res.Status)
	}
	if len(res.NodeResults) != 4 {
		t.Fatalf("期望 4 个节点全部执行完成, 实际: %d", len(res.NodeResults))
	}
}
