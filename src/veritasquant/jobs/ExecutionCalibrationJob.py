"""执行校准任务 console script 入口。"""

from __future__ import annotations

from collections.abc import Sequence

from veritasquant.application.Entrypoints import runOfflineEntrypoint


def main(argv: Sequence[str] | None = None) -> int:
    """返回执行校准任务离线参数校验的明确退出码。"""
    return runOfflineEntrypoint("vq-job-execution-calibration", "运行 VeritasQuant 执行校准任务", argv)
