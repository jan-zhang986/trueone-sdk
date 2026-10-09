# -*- coding: utf-8 -*-
"""
TrueOne Test as Code SDK for Python
提供 @trueone.case, with trueone.step 等装饰器与上下文管理器，
支持 DAG 工作流原生运行与 pytest 插件深度集成。
"""
from aegis import *
from aegis import _current_case_ctx, _current_run_id_ctx, _current_step_evidence_ctx, _StepContext

try:
    from trueone_cli.src.runner.workflow_runner import WorkflowRunner
except ImportError:
    WorkflowRunner = None


def run_workflow(yaml_path: str, variables: dict = None):
    """在 Python 测试用例中直接调用 TrueOne DAG 工作流"""
    # 优先使用本地运行器
    import subprocess
    import json
    import sys
    
    # 支持轻量解析执行或调用 trueone 命令行
    cmd = [sys.executable, "-m", "trueone_cli.src.cli.main", "run", yaml_path, "--json"]
    if variables:
        for k, v in variables.items():
            cmd.extend(["-v", f"{k}={v}"])
    
    proc = subprocess.run(cmd, stdout=subprocess.PIPE, stderr=subprocess.PIPE, text=True)
    if proc.returncode != 0:
        raise RuntimeError(f"Workflow execution failed: {proc.stderr or proc.stdout}")
    
    try:
        return json.loads(proc.stdout)
    except Exception:
        return {"status": "SUCCESS" if proc.returncode == 0 else "FAILED", "raw": proc.stdout}
