# FinvQuant 智能体协作指南

## 适用范围

本文件适用于整个仓库（FinvQuant 量化策略交易平台）。

## 仓库结构

本仓库为"外壳 + 子项目"结构：

```
.
├── VeritasQuant/          # 子项目：VeritasQuant 量化交易平台（完整代码）
├── cmd/ internal/         # FinvQuant Go 服务端（量化平台服务端）
├── Web/                   # 前端（Vue3 + Vite8 + Vuetify4）
├── Deploy/                # Docker Compose 部署编排
├── .github/workflows/     # CI/CD（构建 + 推送 GHCR）
├── AGENTS.md              # 本文件（仓库级协作指南）
├── LICENSE                # MIT License（Copyright 2026 ACANX）
├── README.md              # 仓库概览与快速入口
└── Prompt.md              # 结构化需求文档（持续更新）
```

## 项目目标

FinvQuant 是一个面向多资产的**严格事件驱动量化交易平台**，支持多种基金智能定投方案和用户自定义定投规则的历史回测。在进入模拟盘、券商仿真和受控实盘前，必须先产出可复现的研究和回测结果。

## 归档规则

`VeritasQuant/Archive/` 保存已合并到技术方案中的历史源文档。除非任务明确要求维护归档，否则**不得修改、移动或删除**其中的文件。设计决策变更时，必须同步更新 `VeritasQuant/Docs/VeritasQuantTechSpec.md`。

## 工作约定

- **进入 `VeritasQuant/` 后，以该目录内的 [`VeritasQuant/AGENTS.md`](VeritasQuant/AGENTS.md) 为项目级唯一权威协作指南**——它包含"方案优先"原则、不可违反的事件语义、风控/控制约束、Git 工作流与提交规范。
- **所有 Python 相关的开发、测试、构建、运行和部署命令统一使用 `python3` 执行**；当前默认开发、测试、运行和部署环境为 **Windows 11 上的 Python 3.13**。
- Go 服务端（`cmd/`、`internal/`）与前端（`Web/`）的开发约定见 `Prompt.md` 与各自代码注释。
- 仓库级本文件只负责结构说明；项目级规则一律以 `VeritasQuant/AGENTS.md` 为准。
- 根目录文件（`.gitignore`、`AGENTS.md`、`LICENSE`、`README.md`）描述仓库外壳，不描述项目内部行为。
