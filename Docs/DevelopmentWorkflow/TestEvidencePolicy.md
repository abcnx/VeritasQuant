# 测试与证据规范

## 稳定测试 ID

测试通过 `@pytest.mark.stable_id("<ID>")` 绑定稳定标识。ID 不得因文件移动、参数化顺序或显示名称变化而重用；追踪矩阵使用该 ID 而不是临时的 pytest node id。新行为至少提供正例、反例和边界测试。

## 每次 CI 证据

`scripts/CollectTestEvidence.py` 将 JUnit 汇总与 SHA-256、coverage SHA-256、测试选择、随机种子、UTC 收集时间、运行器 OS/架构/Python 版本以及每个构建工件的 SHA-256 写入 snake_case JSON。CI 上传 JUnit、coverage、证据 JSON 和 `dist/`，保存期为 90 天。

固定种子属性或模型测试必须显式传入 `--seed`；无随机行为的测试使用 `not_applicable`。测试跳过、JUnit 失败、缺任一输入文件或哈希不匹配均不能作为通过证据。运行、环境与证据关联的工作项 ID 必须写入 `--work-item`。

## 示例

```powershell
python3 -m coverage run -m pytest tests\unit tests\contract tests\packaging --junitxml artifacts\JUnit.xml
python3 -m coverage xml -o artifacts\Coverage.xml
python3 scripts\CollectTestEvidence.py --junit artifacts\JUnit.xml --coverage artifacts\Coverage.xml --artifact dist\veritasquant-0.1.0-py3-none-any.whl --work-item P0-009 --seed not_applicable --output artifacts\TestEvidence.json
```

该规范只定义作者和 CI 证据格式；独立 QA 验证、验收人签署与 Gate 结论仍按开发工作流执行。
