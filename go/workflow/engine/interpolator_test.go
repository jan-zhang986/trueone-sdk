package engine_test

import (
	"testing"

	"github.com/vanguard-platform/aegis-sdk-go/workflow/engine"
)

func TestInterpolateValue(t *testing.T) {
	scope := engine.NewContextScope(map[string]interface{}{
		"variables.userId": "user_007",
		"variables.amount": 99.9,
	})

	// 测试字符串内嵌插值
	res1 := engine.InterpolateString("Hello {{ variables.userId }}!", scope)
	if res1 != "Hello user_007!" {
		t.Errorf("期望 Hello user_007!, 实际: %v", res1)
	}

	// 测试纯占位符保真类型
	res2 := engine.InterpolateString("{{ variables.amount }}", scope)
	if val, ok := res2.(float64); !ok || val != 99.9 {
		t.Errorf("期望保持 float64(99.9), 实际: %v (%T)", res2, res2)
	}
}
