"""API 服务 console script 入口。

- 默认离线参数校验（无副作用，供 packaging 契约测试）；
- `--serve` 时组装依赖并启动 uvicorn；
- 导入模块不连接数据库、不启动线程。
"""

from __future__ import annotations

import argparse
from collections.abc import Sequence

from veritasquant.application.Entrypoints import configureStandardStreams


def main(argv: Sequence[str] | None = None) -> int:
    """解析 API 服务参数；--serve 时启动 uvicorn 服务。"""
    configureStandardStreams()
    parser = argparse.ArgumentParser(prog="vq-api-server", description="启动 VeritasQuant API 服务")
    parser.add_argument("--config", help="显式绝对配置文件路径")
    parser.add_argument("--resource-dir", help="显式绝对部署资源目录")
    parser.add_argument("--runtime-dir", help="显式绝对运行产物目录")
    parser.add_argument("--host", default="127.0.0.1", help="监听地址（默认 127.0.0.1）")
    parser.add_argument("--port", type=int, default=8000, help="监听端口（默认 8000）")
    parser.add_argument(
        "--serve",
        action="store_true",
        help="实际启动服务；缺省时仅做离线参数校验",
    )
    try:
        arguments = parser.parse_args(argv)
    except SystemExit as error:
        return error.code if isinstance(error.code, int) else 1
    if not arguments.serve:
        return 0
    return _serve(arguments)


def _serve(arguments: argparse.Namespace) -> int:
    """组装依赖并启动 uvicorn；运行错误返回非零退出码。"""
    import uvicorn

    from veritasquant.apps.server.ApiApp import buildApiDependencies, createApp
    from veritasquant.apps.server.ApiRuntime import (
        ErrorCatalogProbe,
        PackagedApiVersionProvider,
    )
    from veritasquant.application.ApiErrors import ApiErrorCatalog

    catalog = ApiErrorCatalog.loadPackaged()
    provider = PackagedApiVersionProvider()
    deps = buildApiDependencies(
        errorCatalog=catalog,
        versionProvider=provider,
        readinessProbes=(ErrorCatalogProbe(catalog),),
    )
    app = createApp(deps)
    try:
        uvicorn.run(app, host=arguments.host, port=arguments.port, log_level="info")
    except Exception:  # noqa: BLE001 - 启动失败统一记录并返回非零
        return 1
    return 0
