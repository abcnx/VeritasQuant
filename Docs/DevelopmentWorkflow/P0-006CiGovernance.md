# P0-006 CI 与合并治理记录

## 已实现配置

`.github/workflows/Ci.yml` 在 `ubuntu-latest` 与 `windows-latest` 的 Python 3.13 运行以下不可跳过步骤：依赖锁与前置检查、Ruff、Mypy、JUnit 与 coverage、wheel/sdist 构建、仓库外 wheel 安装、秘密扫描、`pip-audit` 和许可证校验。质量和构建产物以 CI artifact 保存 90 天。

## 必须由代码托管管理员完成的配置

以下状态不能由仓库文件替代，完成前 P0-006 不具备验收证据：

1. 恢复或指定有效远程 Git 仓库与默认主分支。
2. 将 `Quality (ubuntu-latest)`、`Quality (windows-latest)` 与 `Security baseline` 设为必需检查，禁止管理员普通绕过。
3. 要求至少一名非作者评审；事件、排序、账本、订单、风控、配置、API 和安全改动另加对应 CodeOwner。
4. 在两个平台从空缓存运行一次，归档 Build ID、commit、解释器版本、JUnit、coverage、哈希和产物保留链接。

当前目录已可被 Git 识别为仓库，且 `origin` 可访问 `main` 与 `dev` 分支；本地 Ubuntu WSL 的 Linux 构建、测试和仓库外 wheel 验证见 [M0LinuxValidationEvidence.md](M0LinuxValidationEvidence.md)。但当前未取得 GitHub 分支保护管理凭据，也没有实际 Actions 运行记录，故上述远程治理和 CI 证据仍无法验证。详见 `RSK-P0-002` 与 `RSK-P0-003`。

## 当前执行顺序

`CHG-P0-001` 处于 `APPROVAL_PENDING`。当前优先级是先使用已归档的 Windows 本地构建、测试和仓库外 wheel 验证证据支持开发；Linux CI 的远程实施与空缓存运行作为 `ACT-P0-003` 排入后续。

受保护分支和必需检查配置作为 `ACT-P0-004` 排入后续。该行动完成前，任何本地测试结果都不能证明合并会被 CI 失败阻断。

该调整不修改 P0-006 的双平台验收标准，也不修改技术方案。Linux 大小写敏感环境的完整运行证据仍必须在 M0 正式评审前提供；在此之前 M0 只能是 `INSUFFICIENT_EVIDENCE`。
