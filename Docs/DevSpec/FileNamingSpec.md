# 文件命名规范（FileNamingSpec）

> 所属：FinvQuant 开发规范 · 存放：`Docs/DevSpec/`
> 适用范围：全部文件、目录与 Python 标识符命名。

## 1. 文件与目录命名

- 项目中的程序文件、资源文件、文档文件和非 Python 包的组件目录优先使用**大驼峰（PascalCase）**命名，例如 `EventLoop.py`、`Apps/GuiClient/`。

## 2. Python 标识符

- 项目自定义 Python 类名使用 **PascalCase**。
- 方法、函数、参数、局部变量和模型字段优先使用**小驼峰（lowerCamelCase）**，例如 `createOrder()`、`sourceSequence`、`riskPolicyVersion`。
- 既定核心字段 `ts` 保持不变。
- 该规则只约束 Python 标识符，**不改写**版本化事件、API、数据库或文件协议的 wire 字段；风格不一致时使用唯一显式 alias。
- Python 包目录仍遵循导入约定，不因标识符风格调整包路径。

## 3. 标识符语言

- 代码标识符、事件名和配置键使用清晰的**英文**。
