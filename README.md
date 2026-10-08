# TrueOne SDK (Multi-Language Test-as-Code Contracts)

TrueOne 质量平台核心多语言契约 SDK。为各语言测试工程提供标准化用例元数据注解、步骤时序上下文以及运行期真实证据捕获能力。

## 🎯 设计目标

1. **消除多语言割裂**：无论研发使用 Go、Python 还是 Java，统一遵循完全相同的 `TrueOneEvent` 事件模型协议。
2. **拒绝盲测黑盒**：引入三段式步骤上下文（Arrange-Act-Assert），细粒度度量步骤耗时与异常。
3. **因果铁证留存**：提供原生的 `attach_evidence` API，捕获数据库变动快照、网络调用请求体与关键断言，杜绝无断言的“假通过”。

---

## 📦 多语言子模块

### 1. Go SDK (`go/`)
面向原生 `testing.T` 的轻量级契约库。
- **模块路径**：`github.com/jan-zhang986/trueone-sdk/go`
- **使用示例**：
```go
package tests

import (
    "testing"
    "github.com/vanguard/aegis-sdk-go/aegis"
)

func TestRefundStateMachine(t *testing.T) {
    c := aegis.NewCase(t, "TC-REFUND-001", "REQ-PAY-1.2", aegis.RiskP0)
    defer c.End()

    c.Step("1. 初始化订单与资金状态", func() {
        // 构造上下文
    })

    c.Step("2. 发起退款请求并捕获副作用", func() {
        c.AttachEvidence("db_diff", []byte("+ INSERT INTO trade_refund ..."))
    })

    c.Step("3. 校验退款状态机终态跃迁", func() {
        // 核心业务断言
    })
}
```

### 2. Python SDK (`python/`)
面向 `pytest` 的原生插件与装饰器。
- **安装**：`pip install -e python/`
- **使用示例**：
```python
import pytest
import aegis

@aegis.case(
    id="TC-AUTH-002",
    req="REQ-AUTH-1.0",
    title="高频请求限流拦截",
    risk="P0"
)
def test_sms_rate_limit():
    with aegis.step("1. 模拟 1 秒内连续 10 次请求"):
        aegis.attach_evidence("concurrent_count", 10)

    with aegis.step("2. 校验第 10 次请求被拦截"):
        assert 429 == 429
```

### 3. Java SDK (`java/`)
面向 `JUnit 5` 的标准扩展与注解。
- **依赖坐标**：`io.aegis:aegis-sdk-java:1.0.0`
- **使用示例**：
```java
@ExtendWith(AegisExtension.class)
public class TestOrderSettle {

    @Test
    @AegisCase(id = "TC-SETTLE-01", req = "REQ-SETTLE-3.0", risk = RiskLevel.P0)
    void testExecution() {
        Aegis.step("1. 初始化清算参数", () -> {
            assertNotNull("TC-SETTLE-01");
        });

        Aegis.step("2. 触发日终对账", () -> {
            Aegis.attachEvidence("batchId", "BATCH-20261008");
            assertTrue(true);
        });
    }
}
```

---

## 🔗 生态联动

- 由 **`trueone-cli`** 的 `trueone init` 命令自动为业务测试工程完成依赖声明注入。
- 执行器运行时产生的结构化证据由云端 **`trueone-anubis`** 收集，并在 **`trueone-web`** 活体三联屏视窗实时渲染。
