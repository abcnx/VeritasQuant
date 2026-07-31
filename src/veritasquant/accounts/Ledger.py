"""阶段 1 不可变复式账本 journal 与分录契约。"""

from __future__ import annotations

from datetime import datetime, timezone
from decimal import Decimal
from enum import StrEnum
from typing import TypeVar

from pydantic import field_validator, model_validator

from veritasquant.core.Models import PascalAlias, StrictModel


class LedgerContractError(ValueError):
    """账本事实不满足逐单位平衡或来源追溯契约。"""


class JournalType(StrEnum):
    """技术方案定义的不可变账本事实类别。"""

    OpeningBalance = "OPENING_BALANCE"
    OrderReservation = "ORDER_RESERVATION"
    OrderRelease = "ORDER_RELEASE"
    Trade = "TRADE"
    Fee = "FEE"
    Tax = "TAX"
    CorporateAction = "CORPORATE_ACTION"
    MarkToMarket = "MARK_TO_MARKET"
    Margin = "MARGIN"
    Settlement = "SETTLEMENT"
    Delivery = "DELIVERY"
    FxConversion = "FX_CONVERSION"
    BrokerCorrection = "BROKER_CORRECTION"
    Reversal = "REVERSAL"
    ManualAdjustment = "MANUAL_ADJUSTMENT"


class LedgerAccount(StrEnum):
    """阶段 1 显式科目，禁止用无约束字符串猜测科目。"""

    CashAvailable = "CASH_AVAILABLE"
    CashFrozen = "CASH_FROZEN"
    CashReceivable = "CASH_RECEIVABLE"
    CashPayable = "CASH_PAYABLE"
    SecuritiesAvailable = "SECURITIES_AVAILABLE"
    SecuritiesFrozen = "SECURITIES_FROZEN"
    SecuritiesReceivable = "SECURITIES_RECEIVABLE"
    MarginAvailable = "MARGIN_AVAILABLE"
    MarginFrozen = "MARGIN_FROZEN"
    TradingClearing = "TRADING_CLEARING"
    FeeExpense = "FEE_EXPENSE"
    TaxExpense = "TAX_EXPENSE"
    RoundingResidual = "ROUNDING_RESIDUAL"
    ExternalCapital = "EXTERNAL_CAPITAL"
    RealizedProfitLoss = "REALIZED_PROFIT_LOSS"
    UnrealizedProfitLoss = "UNREALIZED_PROFIT_LOSS"


class EntryDirection(StrEnum):
    """复式记账方向。"""

    Debit = "DEBIT"
    Credit = "CREDIT"


class CostBasisMethod(StrEnum):
    """运行开始时固定的成本法。"""

    MovingAverage = "MOVING_AVERAGE"
    Fifo = "FIFO"


class AccountingPolicyV1(StrictModel):
    """运行级会计策略，不允许在同一运行内切换。"""

    policyVersion: str = PascalAlias("PolicyVersion", min_length=1)
    costBasisMethod: CostBasisMethod = PascalAlias("CostBasisMethod")
    monetaryRounding: str = PascalAlias("MonetaryRounding", pattern="^ROUND_HALF_EVEN$")
    policyHash: str = PascalAlias("PolicyHash", pattern=r"^[0-9a-f]{64}$")

    @field_validator("costBasisMethod", mode="before")
    @classmethod
    def parseCostBasisMethod(cls, value: object) -> CostBasisMethod:
        return _parseEnum(value, CostBasisMethod, "成本法")


class AssetUnitV1(StrictModel):
    """用于独立借贷平衡的资产或币种计量单位。"""

    unitId: str = PascalAlias("UnitId", min_length=1)
    assetId: str = PascalAlias("AssetId", min_length=1)
    currency: str | None = PascalAlias("Currency", default=None, pattern=r"^[A-Z]{3}$")

    @model_validator(mode="after")
    def validateUnit(self) -> "AssetUnitV1":
        if self.currency is None and not self.unitId.startswith("INSTRUMENT:"):
            raise LedgerContractError("非币种计量单位必须使用 INSTRUMENT: 前缀")
        if self.currency is not None and self.unitId != self.currency:
            raise LedgerContractError("币种计量单位 ID 必须等于币种代码")
        return self


class LedgerEntryV1(StrictModel):
    """journal 内不可修改的单条借方或贷方分录。"""

    entryId: str = PascalAlias("EntryId", min_length=1)
    ledgerAccount: LedgerAccount = PascalAlias("LedgerAccount")
    direction: EntryDirection = PascalAlias("Direction")
    unit: AssetUnitV1 = PascalAlias("Unit")
    quantity: Decimal = PascalAlias("Quantity", gt=Decimal("0"))
    bookCurrency: str = PascalAlias("BookCurrency", pattern=r"^[A-Z]{3}$")
    bookAmount: Decimal = PascalAlias("BookAmount", ge=Decimal("0"))
    costAmount: Decimal = PascalAlias("CostAmount", ge=Decimal("0"))

    @field_validator("ledgerAccount", mode="before")
    @classmethod
    def parseLedgerAccount(cls, value: object) -> LedgerAccount:
        return _parseEnum(value, LedgerAccount, "账本科目")

    @field_validator("direction", mode="before")
    @classmethod
    def parseDirection(cls, value: object) -> EntryDirection:
        return _parseEnum(value, EntryDirection, "分录方向")


class JournalV1(StrictModel):
    """账户范围内按计量单位严格平衡的不可变复式 journal。"""

    journalId: str = PascalAlias("JournalId", min_length=1)
    journalType: JournalType = PascalAlias("JournalType")
    accountId: str = PascalAlias("AccountId", min_length=1)
    subaccountId: str | None = PascalAlias("SubaccountId", default=None, min_length=1)
    ts: datetime = PascalAlias("Ts")
    commitSequence: int = PascalAlias("CommitSequence", ge=1)
    sourceEventId: str = PascalAlias("SourceEventId", min_length=1)
    reversalOfJournalId: str | None = PascalAlias("ReversalOfJournalId", default=None, min_length=1)
    instrumentMetadataVersion: str = PascalAlias("InstrumentMetadataVersion", min_length=1)
    feeScheduleVersion: str = PascalAlias("FeeScheduleVersion", min_length=1)
    accountingPolicyVersion: str = PascalAlias("AccountingPolicyVersion", min_length=1)
    entries: tuple[LedgerEntryV1, ...] = PascalAlias("Entries", min_length=2)

    @field_validator("journalType", mode="before")
    @classmethod
    def parseJournalType(cls, value: object) -> JournalType:
        return _parseEnum(value, JournalType, "journal 类型")

    @field_validator("ts")
    @classmethod
    def validateUtc(cls, value: datetime) -> datetime:
        if value.tzinfo is None or value.utcoffset() != timezone.utc.utcoffset(value):
            raise LedgerContractError("账本 ts 必须为 UTC 时区时间")
        return value

    @model_validator(mode="after")
    def validateJournal(self) -> "JournalV1":
        if self.journalType is JournalType.Reversal and self.reversalOfJournalId is None:
            raise LedgerContractError("REVERSAL 必须引用原 journal")
        if self.journalType is not JournalType.Reversal and self.reversalOfJournalId is not None:
            raise LedgerContractError("仅 REVERSAL 可以声明原 journal 引用")
        if self.reversalOfJournalId == self.journalId:
            raise LedgerContractError("journal 不得引用自身为冲正对象")
        entryIds = [entry.entryId for entry in self.entries]
        if len(entryIds) != len(set(entryIds)):
            raise LedgerContractError("同一 journal 的 entryId 不得重复")
        balances: dict[str, Decimal] = {}
        bookBalances: dict[str, Decimal] = {}
        for entry in self.entries:
            signedQuantity = entry.quantity if entry.direction is EntryDirection.Debit else -entry.quantity
            balances[entry.unit.unitId] = balances.get(entry.unit.unitId, Decimal("0")) + signedQuantity
            signedBookAmount = entry.bookAmount if entry.direction is EntryDirection.Debit else -entry.bookAmount
            bookBalances[entry.bookCurrency] = bookBalances.get(entry.bookCurrency, Decimal("0")) + signedBookAmount
        unbalancedUnits = [unitId for unitId, balance in balances.items() if balance != 0]
        unbalancedUnits.extend(
            f"BOOK:{currency}" for currency, balance in bookBalances.items() if balance != 0
        )
        if unbalancedUnits:
            raise LedgerContractError(f"journal 按计量单位不平衡: {', '.join(sorted(unbalancedUnits))}")
        return self


EnumType = TypeVar("EnumType", bound=StrEnum)


def _parseEnum(value: object, enumType: type[EnumType], label: str) -> EnumType:
    """仅接受模型枚举或协议中完全匹配的字符串。"""
    if isinstance(value, enumType):
        return value
    if not isinstance(value, str):
        raise LedgerContractError(f"{label}必须是受控字符串")
    try:
        return enumType(value)
    except ValueError as error:
        raise LedgerContractError(f"未知{label}: {value}") from error
