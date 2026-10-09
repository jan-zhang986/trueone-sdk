package engine

import (
	"context"
	"fmt"
	"sync"
	"time"

	wfModel "github.com/vanguard-platform/aegis-sdk-go/workflow/model"
	"github.com/vanguard-platform/aegis-sdk-go/workflow/processor"
	_ "github.com/vanguard-platform/aegis-sdk-go/workflow/processor/api"
	_ "github.com/vanguard-platform/aegis-sdk-go/workflow/processor/data"
	_ "github.com/vanguard-platform/aegis-sdk-go/workflow/processor/gate"
)

// ContextScope 线程安全的执行上下文共享池
type ContextScope struct {
	mu   sync.RWMutex
	data map[string]interface{}
}

func NewContextScope(initData map[string]interface{}) *ContextScope {
	m := make(map[string]interface{})
	if initData != nil {
		for k, v := range initData {
			m[k] = v
		}
	}
	return &ContextScope{data: m}
}

func (s *ContextScope) Get(key string) (interface{}, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	val, ok := s.data[key]
	return val, ok
}

func (s *ContextScope) Set(key string, val interface{}) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.data[key] = val
}

func (s *ContextScope) Snapshot() map[string]interface{} {
	s.mu.RLock()
	defer s.mu.RUnlock()
	snap := make(map[string]interface{})
	for k, v := range s.data {
		snap[k] = v
	}
	return snap
}

// DAGEngine 插件化、高性能有向无环图调度引擎（微内核架构）
type DAGEngine struct {
	registry *processor.ProcessorRegistry
}

// NewDAGEngine 创建并初始化 DAG 引擎，绑定默认插件注册中心
func NewDAGEngine() *DAGEngine {
	return &DAGEngine{
		registry: processor.GetRegistry(),
	}
}

// NewDAGEngineWithRegistry 允许注入自定义注册中心
func NewDAGEngineWithRegistry(reg *processor.ProcessorRegistry) *DAGEngine {
	return &DAGEngine{
		registry: reg,
	}
}

// RegisterProcessor 注册新节点处理器插件
func (e *DAGEngine) RegisterProcessor(p processor.ProcessorInterface) {
	e.registry.Register(p)
}

// ExecuteGraph 调度执行 DAG
func (e *DAGEngine) ExecuteGraph(ctx context.Context, graph *wfModel.WorkflowGraph, initParams map[string]interface{}) (*wfModel.WorkflowExecutionResult, error) {
	startTime := time.Now()
	executionID := fmt.Sprintf("wf-exec-%d", startTime.UnixMilli())

	// 统一全局变量池
	allVars := make(map[string]interface{})
	if graph.Variables != nil {
		for k, v := range graph.Variables {
			allVars[k] = v
		}
	}
	if graph.Params != nil {
		for k, v := range graph.Params {
			if _, exists := allVars[k]; !exists {
				allVars[k] = v
			}
		}
	}
	if initParams != nil {
		for k, v := range initParams {
			allVars[k] = v
		}
	}

	scope := NewContextScope(map[string]interface{}{
		"variables": allVars,
		"vars":      allVars,
		"params":    allVars,
	})
	for k, v := range allVars {
		scope.Set("variables."+k, v)
		scope.Set("vars."+k, v)
		scope.Set("params."+k, v)
		scope.Set(k, v)
	}

	nodeMap := make(map[string]*wfModel.WorkflowNode)
	inDegree := make(map[string]int)
	adjList := make(map[string][]string)

	for i := range graph.Nodes {
		node := &graph.Nodes[i]
		nodeMap[node.ID] = node
		inDegree[node.ID] = len(node.DependsOn)
		for _, dep := range node.DependsOn {
			adjList[dep] = append(adjList[dep], node.ID)
		}
	}

	resultMap := make(map[string]wfModel.NodeExecutionResult)
	var mu sync.Mutex
	statusChan := make(chan string)

	var readyQueue []string
	for id, deg := range inDegree {
		if deg == 0 {
			readyQueue = append(readyQueue, id)
		}
	}

	totalNodes := len(graph.Nodes)
	completedNodes := 0
	hasFailure := false

	var wg sync.WaitGroup

	executeNodeAsync := func(nodeID string) {
		defer wg.Done()
		node := nodeMap[nodeID]

		nodeStart := time.Now()
		var res *wfModel.NodeExecutionResult

		// 变量插值
		resolvedConfig := InterpolateConfig(node.Config, scope)

		// 动态获取处理器
		proc, exists := e.registry.Get(string(node.Type))
		if !exists {
			res = &wfModel.NodeExecutionResult{
				NodeID:         node.ID,
				NodeName:       node.Name,
				Status:         wfModel.NodeStatusFailed,
				DurationMs:     time.Since(nodeStart).Milliseconds(),
				ResolvedConfig: resolvedConfig,
				Error:          fmt.Sprintf("no processor registered for type: %s", node.Type),
			}
		} else {
			if valErr := proc.ValidateConfig(resolvedConfig); valErr != nil {
				res = &wfModel.NodeExecutionResult{
					NodeID:         node.ID,
					NodeName:       node.Name,
					Status:         wfModel.NodeStatusFailed,
					DurationMs:     time.Since(nodeStart).Milliseconds(),
					ResolvedConfig: resolvedConfig,
					Error:          valErr.Error(),
				}
			} else {
				execCtx := &processor.ExecutionContext{
					NodeID:             node.ID,
					NodeName:           node.Name,
					NodeType:           string(node.Type),
					Config:             resolvedConfig,
					PredecessorResults: make(map[string]interface{}),
					ScopeGetter:        scope.Get,
					ScopeSetter:        scope.Set,
				}
				for _, depID := range node.DependsOn {
					if depRes, ok := resultMap[depID]; ok {
						execCtx.PredecessorResults[depID] = depRes.Output
					}
				}

				procRes, err := proc.Execute(ctx, execCtx)
				if err != nil && procRes == nil {
					res = &wfModel.NodeExecutionResult{
						NodeID:         node.ID,
						NodeName:       node.Name,
						Status:         wfModel.NodeStatusFailed,
						DurationMs:     time.Since(nodeStart).Milliseconds(),
						ResolvedConfig: resolvedConfig,
						Error:          err.Error(),
					}
				} else if procRes != nil {
					nodeStatus := wfModel.NodeStatusSuccess
					if procRes.Status == "FAILED" {
						nodeStatus = wfModel.NodeStatusFailed
					}
					res = &wfModel.NodeExecutionResult{
						NodeID:         node.ID,
						NodeName:       node.Name,
						Status:         nodeStatus,
						DurationMs:     procRes.DurationMs,
						ResolvedConfig: resolvedConfig,
						Output:         procRes.Output,
						Evidence:       procRes.Evidence,
						Error:          procRes.Error,
					}
				}
			}
		}

		// 产出注入变量池
		if res != nil && res.Output != nil {
			scope.Set(nodeID, map[string]interface{}{"output": res.Output})
			scope.Set(nodeID+".output", res.Output)
			scope.Set("nodes."+nodeID+".output", res.Output)
			for outK, outV := range res.Output {
				scope.Set(fmt.Sprintf("%s.output.%s", nodeID, outK), outV)
			}
		}

		mu.Lock()
		resultMap[nodeID] = *res
		if res.Status == wfModel.NodeStatusFailed {
			hasFailure = true
		}
		mu.Unlock()

		statusChan <- nodeID
	}

	for _, id := range readyQueue {
		wg.Add(1)
		go executeNodeAsync(id)
	}

	for completedNodes < totalNodes {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case finishedNodeID := <-statusChan:
			completedNodes++
			mu.Lock()
			nodeRes := resultMap[finishedNodeID]
			if nodeRes.Status == wfModel.NodeStatusSuccess {
				for _, nextID := range adjList[finishedNodeID] {
					inDegree[nextID]--
					if inDegree[nextID] == 0 {
						wg.Add(1)
						go executeNodeAsync(nextID)
					}
				}
			} else {
				for _, nextID := range adjList[finishedNodeID] {
					if _, exists := resultMap[nextID]; !exists {
						resultMap[nextID] = wfModel.NodeExecutionResult{
							NodeID:   nextID,
							NodeName: nodeMap[nextID].Name,
							Status:   wfModel.NodeStatusSkipped,
							Error:    fmt.Sprintf("dependency node %s failed", finishedNodeID),
						}
						totalNodes--
					}
				}
			}
			mu.Unlock()
		}
	}

	wg.Wait()

	finalStatus := wfModel.WorkflowStatusSuccess
	if hasFailure {
		finalStatus = wfModel.WorkflowStatusFailed
	}

	return &wfModel.WorkflowExecutionResult{
		WorkflowID:  graph.ID,
		ExecutionID: executionID,
		Status:      finalStatus,
		TotalNodes:  len(graph.Nodes),
		DurationMs:  time.Since(startTime).Milliseconds(),
		NodeResults: resultMap,
		Context:     scope.Snapshot(),
	}, nil
}
