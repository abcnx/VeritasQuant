package mvsv

import (
	"strings"
	"testing"
)

const sampleMvsv = `# Format : "MVSV-1"
# Field : "ts|dt|o|c|l|h|v|t|cp|cr|p"
# Count : 3
# EffectiveTimeZone : "Asia/Shanghai"
# Code : "518880"
# Market : "SSE"
# MarketCode : 1
# CurrencyCode : 1
# PriceAccuracy : 3
# LotSize : 100

1785720600|20260803093000|7.001|7.002|6.999|7.003|100000|70010000000|0.001|0.000143|7.000
1785720660|20260803093100|7.002|7.001|7.000|7.003|80000|56010000000|0.000|-0.000143|7.002
1785720720|20260803093200|7.003|7.004|7.001|7.005|90000|63020000000|0.002|0.000286|7.001
`

func TestParseValid(t *testing.T) {
	result, err := Parse([]byte(sampleMvsv), "TEST.mvsv")
	if err != nil {
		t.Fatalf("解析失败: %v", err)
	}
	if result.Header.Count != 3 {
		t.Fatalf("Count=%d，期望 3", result.Header.Count)
	}
	if result.Header.Values["MarketCode"] != "1" {
		t.Fatalf("MarketCode=%s", result.Header.Values["MarketCode"])
	}
	if len(result.Rows) != 3 {
		t.Fatalf("行数=%d，期望 3", len(result.Rows))
	}
	first := result.Rows[0]
	if first.SecuCode != "518880" || first.MarketCode != 1 {
		t.Fatalf("证券标识错误: %+v", first)
	}
	if first.Ts != 1785720600 {
		t.Fatalf("ts=%d", first.Ts)
	}
	if first.Date == nil || *first.Date != 20260803 {
		t.Fatalf("date=%v", first.Date)
	}
	if first.Time == nil || *first.Time != 93000 {
		t.Fatalf("time=%v", first.Time)
	}
	if first.Close == nil || *first.Close != "7.002" {
		t.Fatalf("close=%v", first.Close)
	}
	if first.Volume == nil || *first.Volume != 100000 {
		t.Fatalf("volume=%v", first.Volume)
	}
}

func TestParseMissingMarketCode(t *testing.T) {
	content := strings.Replace(sampleMvsv, "# MarketCode : 1\n", "", 1)
	_, err := Parse([]byte(content), "TEST.mvsv")
	if err == nil || !strings.Contains(err.Error(), "MarketCode") {
		t.Fatalf("期望缺少 MarketCode 错误，实际: %v", err)
	}
}

func TestParseBadFormat(t *testing.T) {
	_, err := Parse([]byte("# Format : \"MVSV-1\"\nbroken\n"), "bad.mvsv")
	if err == nil {
		t.Fatal("期望解析失败")
	}
}

func TestParseTsTimezoneMismatch(t *testing.T) {
	content := strings.Replace(sampleMvsv,
		"1785720600|20260803093000", "1785720600|20260803103000", 1)
	_, err := Parse([]byte(content), "TEST.mvsv")
	if err == nil || !strings.Contains(err.Error(), "不一致") {
		t.Fatalf("期望 ts 与 dt 不一致错误，实际: %v", err)
	}
}

// 布局 B（12 列，pc 已过时）：ts|d|t|o|c|l|h|v|a|cp|cr|p
// d=8 位日期 + t=6 位时间，a=成交额；数据行末尾允许残留空段（旧 pc 位，如 "...|4332.1|"）
// ts=1767337200 → 2026-01-02 07:00 UTC → America/New_York（UTC-5）2026-01-02 02:00:00（20260102020000）
const sampleMvsvLayoutB = `# Format : "MVSV-1"
# Field : "ts|d|t|o|c|l|h|v|a|cp|cr|p"
# Count : 3
# EffectiveTimeZone : "America/New_York"
# Code : "GCMain"
# Market : "COMEX"
# MarketCode : 1320
# CurrencyCode : 55
# PriceAccuracy : 1
# LotSize : 100

1767337200|20260102|020000|4340|4907.5|4319.7|5626.8|5926343|0|575.4|13.282242|4332.1|
1767337260|20260102|020100|4340|4343.9|4338.5|4349.5|314|0|11.8|0.272385|4332.1|
1767337320|20260102|020200|4342.7|4344|4342|4346.3|81|0|0.1|0.002302|4343.9|
`

func TestParseLayoutB(t *testing.T) {
	result, err := Parse([]byte(sampleMvsvLayoutB), "GCmain.mvsv")
	if err != nil {
		t.Fatalf("解析布局 B 失败: %v", err)
	}
	if len(result.Rows) != 3 {
		t.Fatalf("行数=%d，期望 3", len(result.Rows))
	}
	first := result.Rows[0]
	if first.SecuCode != "GCMain" || first.MarketCode != 1320 {
		t.Fatalf("证券标识错误: %+v", first)
	}
	if first.Ts != 1767337200 {
		t.Fatalf("ts=%d", first.Ts)
	}
	// d+t 分开解析：date=20260102, time=020000
	if first.Date == nil || *first.Date != 20260102 {
		t.Fatalf("date=%v", first.Date)
	}
	if first.Time == nil || *first.Time != 20000 {
		t.Fatalf("time=%v", first.Time)
	}
	// a 列映射成交额
	if first.Turnover == nil || *first.Turnover != "0" {
		t.Fatalf("turnover=%v", first.Turnover)
	}
	if first.Close == nil || *first.Close != "4907.5" {
		t.Fatalf("close=%v", first.Close)
	}
	if first.PrevClose == nil || *first.PrevClose != "4332.1" {
		t.Fatalf("prev_close=%v", first.PrevClose)
	}
}

func TestParseUnsupportedLayout(t *testing.T) {
	content := strings.Replace(sampleMvsvLayoutB,
		"# Field : \"ts|d|t|o|c|l|h|v|a|cp|cr|p\"",
		"# Field : \"ts|d|t|o|c|l|h|v|a|cp|cr|p|x\"", 1)
	_, err := Parse([]byte(content), "bad.mvsv")
	if err == nil || !strings.Contains(err.Error(), "Field 布局不支持") {
		t.Fatalf("期望 Field 布局不支持错误，实际: %v", err)
	}
}

// 行尾空段容忍：数据行末尾允许残留旧 pc 空段（"...|4332.1|"），解析器应截断后按 12 列处理
func TestParseLayoutBTrailingEmptySegment(t *testing.T) {
	// 第三行末尾多一个空段（旧 pc 位残留），其余行正常
	content := strings.Replace(sampleMvsvLayoutB,
		"1767337320|20260102|020200|4342.7|4344|4342|4346.3|81|0|0.1|0.002302|4343.9|",
		"1767337320|20260102|020200|4342.7|4344|4342|4346.3|81|0|0.1|0.002302|4343.9||", 1)
	result, err := Parse([]byte(content), "trailing.mvsv")
	if err != nil {
		t.Fatalf("容忍行尾空段解析失败: %v", err)
	}
	if len(result.Rows) != 3 {
		t.Fatalf("行数=%d，期望 3", len(result.Rows))
	}
	last := result.Rows[2]
	if last.Close == nil || *last.Close != "4344" {
		t.Fatalf("末行 close=%v", last.Close)
	}
}
