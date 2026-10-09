package engine

import (
	"fmt"
	"regexp"
	"strings"
)

var interpolationRegex = regexp.MustCompile(`\{\{\s*([a-zA-Z0-9_.-]+)\s*\}\}`)

// InterpolateConfig 对配置 Map 递归进行插值替换
func InterpolateConfig(config map[string]interface{}, scope *ContextScope) map[string]interface{} {
	if config == nil {
		return nil
	}
	res := make(map[string]interface{})
	for k, v := range config {
		res[k] = interpolateAny(v, scope)
	}
	return res
}

func interpolateAny(val interface{}, scope *ContextScope) interface{} {
	switch v := val.(type) {
	case string:
		return InterpolateString(v, scope)
	case map[string]interface{}:
		return InterpolateConfig(v, scope)
	case []interface{}:
		arr := make([]interface{}, len(v))
		for i, item := range v {
			arr[i] = interpolateAny(item, scope)
		}
		return arr
	default:
		return val
	}
}

// InterpolateString 解析单个字符串中的 {{ var }} 占位符
func InterpolateString(template string, scope *ContextScope) interface{} {
	matches := interpolationRegex.FindAllStringSubmatch(template, -1)
	if len(matches) == 0 {
		return template
	}

	// 如果整个字符串仅仅是一个占位符如 "{{ variables.expectedDeduction }}"，保留原始数据类型 (数字/布尔等)
	trimmed := strings.TrimSpace(template)
	if len(matches) == 1 && matches[0][0] == trimmed {
		varKey := matches[0][1]
		if val, ok := scope.Get(varKey); ok {
			return val
		}
		return template
	}

	// 否则进行字符串替换拼接
	result := template
	for _, match := range matches {
		fullPlaceholder := match[0]
		varKey := match[1]
		if val, ok := scope.Get(varKey); ok {
			result = strings.ReplaceAll(result, fullPlaceholder, fmt.Sprintf("%v", val))
		}
	}
	return result
}
