# FinvQuant 智能体协作指南

## 仓库结构

本仓库的核心项目为 **VeritasQuant**，完整代码位于 [`VeritasQuant/`](VeritasQuant/) 子目录：

```
.
├── VeritasQuant/          # VeritasQuant 量化交易平台（完整项目）
├── AGENTS.md              # 本文件（仓库级协作指南）
├── LICENSE                # MIT License（Copyright 2026 ACANX）
└── README.md              # 仓库概览与快速入口
```

## 工作约定

- **所有开发、测试、构建、运行命令必须在 `VeritasQuant/` 目录内执行**（该目录是 `pyproject.toml` 与源码的所在位置）。
- **进入 `VeritasQuant/` 后，以该目录内的 [`VeritasQuant/AGENTS.md`](VeritasQuant/AGENTS.md) 为唯一权威协作指南**——它包含项目的"方案优先"原则、不可违反的事件语义、风控/控制约束、Git 工作流与提交规范。
- 仓库级本文件只负责说明结构；项目级规则一律以 `VeritasQuant/AGENTS.md` 为准。
- 根目录文件（`.gitignore`、`AGENTS.md`、`LICENSE`、`README.md`）描述仓库外壳，不描述项目内部行为。
