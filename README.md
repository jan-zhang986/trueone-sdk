# TrueOne SDK (Multi-Language Test-as-Code Contracts)

TrueOne Test-as-Code 多语言契约 SDK 核心仓库。

## 目录结构

- `python/`: 面向 `pytest` 的 Python SDK (`@aegis.case`, `with aegis.step`, `attach_evidence`)
- `go/`: 面向原生 `testing.T` 的 Go SDK (`aegis.NewCase`, `c.Step`, `c.AttachEvidence`)
- `java/`: 面向 `JUnit 5` 的 Java SDK (`@AegisCase`, `Aegis.step`, `@ExtendWith(AegisExtension.class)`)

## 核心职责

1. **元数据声明 (Spec)**：将需求 ID (`req`)、风险等级 (`risk=P0/P1`) 与物理用例绑定。
2. **步骤时序追踪 (Steps)**：划分测试阶段（AAA 三段式结构），输出毫秒级步骤事件。
3. **证据链采集 (Evidence)**：挂载数据库变更快照 (DB diff)、接口请求响应体与业务断言快照。
