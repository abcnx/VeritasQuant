"""行情导入命令入口。"""

from __future__ import annotations

from collections.abc import Sequence

from veritasquant.application.Entrypoints import runOfflineEntrypoint


def main(argv: Sequence[str] | None = None) -> int:
    """返回行情导入命令离线参数校验的明确退出码。"""
    return runOfflineEntrypoint("vq-import-market-data", "导入 VeritasQuant 行情数据", argv)
