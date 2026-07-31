"""API 服务 console script 入口。"""

from __future__ import annotations

from collections.abc import Sequence

from veritasquant.application.Entrypoints import runOfflineEntrypoint


def main(argv: Sequence[str] | None = None) -> int:
    """返回 API 服务离线参数校验的明确退出码。"""
    return runOfflineEntrypoint("vq-api-server", "启动 VeritasQuant API 服务", argv)
