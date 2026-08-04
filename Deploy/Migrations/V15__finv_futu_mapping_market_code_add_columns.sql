-- =====================================================================
-- FinvQuant PostgreSQL V15：finv_futu_mapping_market_code 增加列（market_name / exchange）
--
-- 决策（ACANX 2026-08-05）：
--   - V8 已发布（PR #295 合并），按迁移规范不修改已发布迁移，新增 V15 增量迁移补充字段；
--   - market_name：市场名称（如 港股主板 / 美股指数 / 上交所 A 股）；
--   - exchange：对应交易所（富途 exchange 代码，如 SEHK / US / SSE；无交易所用 N/A）；
--   - TEXT 可空（存量行由 V100007 种子填充）；ADD COLUMN IF NOT EXISTS 保证幂等；
--   - 初始数据见 V100007__finv_futu_mapping_market_code_seed.sql（数据种子段）。
-- 迁移策略：与既有迁移一致，单事务、失败回滚。
-- =====================================================================

BEGIN;

ALTER TABLE finv_futu_mapping_market_code
    ADD COLUMN IF NOT EXISTS market_name TEXT;

ALTER TABLE finv_futu_mapping_market_code
    ADD COLUMN IF NOT EXISTS exchange TEXT;

COMMIT;
