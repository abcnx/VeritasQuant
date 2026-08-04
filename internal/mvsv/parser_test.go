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
