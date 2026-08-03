# FinvQuant

金融量化平台仓库。核心项目 **VeritasQuant**（严格事件驱动量化交易平台）位于 [`VeritasQuant/`](VeritasQuant/) 子目录。

## 目录结构

```
├── VeritasQuant/    # VeritasQuant 完整项目（源码/测试/文档/部署）
├── AGENTS.md        # 智能体协作指南（仓库级）
├── LICENSE          # MIT License
└── README.md        # 本文件
```

## 快速开始

进入子项目目录，按其 README 操作：

```bash
cd VeritasQuant
# 服务端（Windows 11 + Docker Desktop）
python3 scripts/DeployServer.py check
python3 scripts/DeployServer.py start
```

更多信息见 [VeritasQuant/README.md](VeritasQuant/README.md) 与 [VeritasQuant/Docs/](VeritasQuant/Docs/)。

## 许可

[MIT](LICENSE) © 2026 ACANX
