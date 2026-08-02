"""数据导入任务 console script 入口。

- 接收 `--job-run-id` 与 `--job-execution-key`（调度幂等契约，TechSpec 11.5）；
- 业务幂等由 command_id/inbox/outbox/checkpoint 保证；
- 离线参数校验无副作用（packaging 契约）。
"""

from __future__ import annotations

import argparse
from collections.abc import Sequence

from veritasquant.jobs.JobEntrypoint import JobEntrypoint

_PROG = "vq-job-data-ingestion"


class DataIngestionJob(JobEntrypoint):
    """数据导入任务：校验执行键后触发数据导入命令。"""

    def __init__(self) -> None:
        super().__init__(_PROG, "运行 VeritasQuant 数据导入任务")

    def buildParser(self) -> argparse.ArgumentParser:
        parser = super().buildParser()
        parser.add_argument("--source", default="", help="数据源名称")
        parser.add_argument("--instrument-id", default="", help="目标标的 ID")
        return parser

    def run(self, arguments: argparse.Namespace) -> int:
        # 此处只做参数与执行键校验；实际导入逻辑复用 application 用例
        # （P2-035 数据导入任务清单落地业务实现）。
        if not arguments.source:
            print("--source 必填", file=__import__("sys").stderr)
            return 2
        if not arguments.instrument_id:
            print("--instrument-id 必填", file=__import__("sys").stderr)
            return 2
        print(
            f"job_run_id={arguments.job_run_id} execution_key={arguments.job_execution_key} "
            f"source={arguments.source} instrument={arguments.instrument_id}"
        )
        return 0


def main(argv: Sequence[str] | None = None) -> int:
    """返回数据导入任务离线参数校验的明确退出码。"""
    return DataIngestionJob().main(argv)
