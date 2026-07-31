"""调度服务 console script 入口。"""

from __future__ import annotations

from collections.abc import Sequence

from veritasquant.application.Entrypoints import runOfflineEntrypoint


def main(argv: Sequence[str] | None = None) -> int:
    """返回调度服务离线参数校验的明确退出码。"""
    return runOfflineEntrypoint("vq-scheduler-service", "启动 VeritasQuant 调度服务", argv)
