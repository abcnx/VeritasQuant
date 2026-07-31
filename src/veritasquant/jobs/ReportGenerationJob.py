"""报告生成任务 console script 入口。"""

from __future__ import annotations

from collections.abc import Sequence

from veritasquant.application.Entrypoints import runOfflineEntrypoint


def main(argv: Sequence[str] | None = None) -> int:
    """返回报告生成任务离线参数校验的明确退出码。"""
    return runOfflineEntrypoint("vq-job-report-generation", "运行 VeritasQuant 报告生成任务", argv)
