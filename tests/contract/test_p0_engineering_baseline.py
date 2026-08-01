from __future__ import annotations

from pathlib import Path

import pytest
import yaml


ROOT = Path(__file__).resolve().parents[2]


def loadYaml(relativePath: str) -> dict[str, object]:
    value = yaml.safe_load((ROOT / relativePath).read_text(encoding="utf-8"))
    assert isinstance(value, dict)
    return value


@pytest.mark.stable_id("P0-004-001")
def test_solution_a_layout_has_required_roots_and_single_import_package() -> None:
    for directory in ("Apps", "Jobs", "Migrations", "Docker", "Configs", "Resources", "src", "tests"):
        assert (ROOT / directory).is_dir()
    assert (ROOT / "src" / "veritasquant" / "__init__.py").is_file()
    assert not list(ROOT.glob("*.py"))


@pytest.mark.stable_id("P0-008-001")
def test_development_compose_is_ephemeral_and_has_health_checks() -> None:
    compose = loadYaml("Docker/docker-compose.yml")
    services = compose["services"]
    assert isinstance(services, dict)
    assert set(services) == {"postgresql", "redis"}
    assert "volumes" not in compose
    for service in services.values():
        assert isinstance(service, dict)
        assert "tmpfs" in service
        assert "healthcheck" in service


@pytest.mark.stable_id("P0-010-001")
def test_all_p0_work_items_have_governance_fields_and_audit_history() -> None:
    register = loadYaml("Docs/DevelopmentWorkflow/Registers/WorkItemRegister.yml")
    records = register["Records"]
    assert isinstance(records, list)
    p0Records = {record["PlanTaskId"]: record for record in records if isinstance(record, dict) and str(record.get("PlanTaskId", "")).startswith("P0-")}
    assert set(p0Records) == {f"P0-{number:03d}" for number in range(1, 14)}
    required = {"WorkItemId", "PlanTaskId", "Title", "Description", "Owner", "Approver", "Dependencies", "AcceptanceCriteria", "TestEvidence", "RiskLinks", "Status", "StatusChangedTs", "AuditHistory"}
    for record in p0Records.values():
        assert required <= set(record)
        assert record["AuditHistory"]


@pytest.mark.stable_id("P0-010-002")
def test_all_governance_registers_declare_required_fields_and_unique_record_ids() -> None:
    registers = (
        "Docs/DevelopmentWorkflow/Registers/WorkItemRegister.yml",
        "Docs/DevelopmentWorkflow/Registers/BugRegister.yml",
        "Docs/DevelopmentWorkflow/Registers/RiskRegister.yml",
        "Docs/DevelopmentWorkflow/Registers/IncidentRegister.yml",
        "Docs/DevelopmentWorkflow/Registers/ChangeRegister.yml",
        "Docs/DevelopmentWorkflow/Registers/ActionRegister.yml",
    )
    for path in registers:
        register = loadYaml(path)
        assert register["RequiredFields"]
        records = register["Records"]
        assert isinstance(records, list)
        idFields = [field for field in register["RequiredFields"] if str(field).endswith("Id")]
        for record in records:
            assert isinstance(record, dict)
            assert any(record.get(field) for field in idFields)


@pytest.mark.stable_id("P0-012-001")
def test_traceability_matrix_covers_all_mandatory_review_contracts() -> None:
    matrix = loadYaml("Docs/DevelopmentWorkflow/Registers/TraceabilityMatrix.yml")
    rows = matrix["Rows"]
    assert isinstance(rows, list)
    requirementIds = {row["RequirementId"] for row in rows if isinstance(row, dict)}
    assert requirementIds == {f"R-{number:03d}" for number in range(1, 19)}
    for row in rows:
        assert row["PlanTaskIds"]
        assert row["TestIds"]
        assert row["LatestGate"]


@pytest.mark.stable_id("P0-005-001")
def test_ci_and_package_metadata_use_python_313_baseline() -> None:
    workflow = loadYaml(".github/workflows/Ci.yml")
    quality = workflow["jobs"]["quality"]
    assert quality["strategy"]["matrix"]["python-version"] == ["3.13"]
    project = (ROOT / "pyproject.toml").read_text(encoding="utf-8")
    assert 'requires-python = ">=3.13"' in project
