from __future__ import annotations

from datetime import datetime, timezone
from decimal import Decimal

import pytest
from pydantic import ValidationError

from veritasquant.accounts.Ledger import (
    AccountingPolicyV1,
    AssetUnitV1,
    EntryDirection,
    JournalType,
    JournalV1,
    LedgerAccount,
    LedgerEntryV1,
)


def _cashUnit() -> AssetUnitV1:
    return AssetUnitV1(UnitId="CNY", AssetId="CNY", Currency="CNY")


def _entry(entryId: str, direction: EntryDirection, quantity: Decimal = Decimal("100")) -> LedgerEntryV1:
    return LedgerEntryV1(
        EntryId=entryId,
        LedgerAccount=LedgerAccount.CashAvailable,
        Direction=direction,
        Unit=_cashUnit(),
        Quantity=quantity,
        BookCurrency="CNY",
        BookAmount=quantity,
        CostAmount=Decimal("0"),
    )


def _journal(**overrides: object) -> JournalV1:
    values: dict[str, object] = {
        "JournalId": "journal-1",
        "JournalType": JournalType.OpeningBalance,
        "AccountId": "account-1",
        "Ts": datetime(2026, 8, 1, tzinfo=timezone.utc),
        "CommitSequence": 1,
        "SourceEventId": "event-1",
        "InstrumentMetadataVersion": "instrument-v1",
        "FeeScheduleVersion": "fee-v1",
        "AccountingPolicyVersion": "policy-v1",
        "Entries": (_entry("entry-1", EntryDirection.Debit), _entry("entry-2", EntryDirection.Credit)),
    }
    values.update(overrides)
    return JournalV1.model_validate(values)


def test_journal_requires_balanced_decimal_entries_per_unit() -> None:
    journal = _journal()
    assert journal.entries[0].quantity == Decimal("100")
    assert journal.model_dump(by_alias=True)["JournalId"] == "journal-1"


def test_journal_rejects_unbalanced_or_duplicate_entries() -> None:
    with pytest.raises(ValidationError, match="不平衡"):
        _journal(Entries=(_entry("entry-1", EntryDirection.Debit), _entry("entry-2", EntryDirection.Credit, Decimal("99"))))
    with pytest.raises(ValidationError, match="entryId"):
        _journal(Entries=(_entry("entry-1", EntryDirection.Debit), _entry("entry-1", EntryDirection.Credit)))
    with pytest.raises(ValidationError, match="BOOK:CNY"):
        _journal(
            Entries=(
                _entry("entry-1", EntryDirection.Debit),
                LedgerEntryV1(
                    EntryId="entry-2",
                    LedgerAccount=LedgerAccount.CashAvailable,
                    Direction=EntryDirection.Credit,
                    Unit=_cashUnit(),
                    Quantity=Decimal("100"),
                    BookCurrency="CNY",
                    BookAmount=Decimal("99"),
                    CostAmount=Decimal("0"),
                ),
            )
        )


def test_journal_requires_utc_and_valid_reversal_reference() -> None:
    with pytest.raises(ValidationError, match="UTC"):
        _journal(Ts=datetime(2026, 8, 1))
    with pytest.raises(ValidationError, match="必须引用"):
        _journal(JournalType=JournalType.Reversal)
    reversedJournal = _journal(JournalType=JournalType.Reversal, ReversalOfJournalId="journal-original")
    assert reversedJournal.reversalOfJournalId == "journal-original"


def test_ledger_rejects_float_and_invalid_measurement_unit() -> None:
    with pytest.raises(ValidationError):
        _entry("entry-1", EntryDirection.Debit, 100.0)  # type: ignore[arg-type]
    with pytest.raises(ValidationError, match="INSTRUMENT"):
        AssetUnitV1(UnitId="gold", AssetId="gold")


def test_accounting_policy_requires_fixed_rounding_and_hash() -> None:
    policy = AccountingPolicyV1(
        PolicyVersion="policy-v1",
        CostBasisMethod="MOVING_AVERAGE",
        MonetaryRounding="ROUND_HALF_EVEN",
        PolicyHash="a" * 64,
    )
    assert policy.policyVersion == "policy-v1"
