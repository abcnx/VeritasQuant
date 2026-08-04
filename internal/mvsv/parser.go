// Package mvsv 解析 MVSV-1 分钟级历史行情文件格式。
//
// 格式（与 VeritasQuant/data/Mvsv.py 对齐）：
//   - 头部：# Key : Value 行，空行分隔头部与数据区；
//   - 数据区：ts|dt|o|c|l|h|v|t|cp|cr|p（11 列）。
package mvsv

import (
	"fmt"
	"strconv"
	"strings"
	"time"
	_ "time/tzdata" // 内嵌时区数据库，保证 EffectiveTimeZone 解析跨环境可用
)

// RequiredHeaders MVSV-1 必填头部键。
var RequiredHeaders = []string{
	"Format", "Field", "Count", "EffectiveTimeZone", "Code",
	"Market", "MarketCode", "CurrencyCode", "PriceAccuracy", "LotSize",
}

// FieldLayout MVSV-1 数据区固定列顺序。
const FieldLayout = "ts|dt|o|c|l|h|v|t|cp|cr|p"

// Header MVSV-1 头部信息。
type Header struct {
	Values            map[string]string
	Count             int
	EffectiveTimeZone string
}

// Row 一条分钟行情（与 finv_quote_secu_kline_min 表对齐；价格保留字符串精度）。
type Row struct {
	MarketCode int
	SecuCode   string
	Ts         int64
	Date       *int
	Time       *int
	PrevClose  *string
	Open       *string
	High       *string
	Low        *string
	Close      *string
	Paocd      *string
	Volume     *int64
	Turnover   *string
	ExtField   *string
	Remark     *string
}

// ParseResult 解析结果。
type ParseResult struct {
	Header *Header
	Rows   []Row
}

// Parse 解析 MVSV-1 内容（字节流）。
func Parse(content []byte, sourceName string) (*ParseResult, error) {
	lines := strings.Split(strings.ReplaceAll(string(content), "\r\n", "\n"), "\n")

	header, dataStart, err := parseHeader(lines)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", sourceName, err)
	}

	var rows []Row
	for i := dataStart; i < len(lines); i++ {
		line := strings.TrimRight(lines[i], "\r")
		if strings.TrimSpace(line) == "" {
			continue
		}
		row, err := parseRow(line, header, i+1)
		if err != nil {
			return nil, fmt.Errorf("%s: 第 %d 行: %w", sourceName, i+1, err)
		}
		rows = append(rows, row)
	}
	if len(rows) != header.Count {
		return nil, fmt.Errorf("%s: Count=%d，实际记录数=%d", sourceName, header.Count, len(rows))
	}
	return &ParseResult{Header: header, Rows: rows}, nil
}

func parseHeader(lines []string) (*Header, int, error) {
	values := map[string]string{}
	for index, line := range lines {
		trimmed := strings.TrimRight(line, "\r")
		if strings.TrimSpace(trimmed) == "" {
			if len(values) == 0 {
				return nil, 0, fmt.Errorf("头部为空")
			}
			header, err := buildHeader(values)
			if err != nil {
				return nil, 0, err
			}
			return header, index + 1, nil
		}
		key, value, ok := parseHeaderLine(trimmed)
		if !ok {
			return nil, 0, fmt.Errorf("第 %d 行头部格式必须为 '# Key : Value'", index+1)
		}
		if _, exists := values[key]; exists {
			return nil, 0, fmt.Errorf("第 %d 行存在重复头部键: %s", index+1, key)
		}
		values[key] = value
	}
	return nil, 0, fmt.Errorf("未找到头部与数据的空行分隔")
}

func parseHeaderLine(line string) (string, string, bool) {
	if !strings.HasPrefix(line, "# ") {
		return "", "", false
	}
	body := line[2:]
	parts := strings.SplitN(body, " : ", 2)
	if len(parts) != 2 {
		return "", "", false
	}
	key := strings.TrimSpace(parts[0])
	value := strings.Trim(strings.TrimSpace(parts[1]), `"`)
	if key == "" {
		return "", "", false
	}
	return key, value, true
}

func buildHeader(values map[string]string) (*Header, error) {
	for _, key := range RequiredHeaders {
		if _, ok := values[key]; !ok {
			return nil, fmt.Errorf("缺少必填头部: %s", key)
		}
	}
	if values["Format"] != "MVSV-1" {
		return nil, fmt.Errorf("Format 必须严格为 MVSV-1，实际: %s", values["Format"])
	}
	if values["Field"] != FieldLayout {
		return nil, fmt.Errorf("Field 必须严格为 %s，实际: %s", FieldLayout, values["Field"])
	}
	count, err := strconv.Atoi(values["Count"])
	if err != nil || count < 0 {
		return nil, fmt.Errorf("Count 必须为非负整数")
	}
	if _, err := time.LoadLocation(values["EffectiveTimeZone"]); err != nil {
		return nil, fmt.Errorf("EffectiveTimeZone 非法: %s", values["EffectiveTimeZone"])
	}
	return &Header{
		Values:            values,
		Count:             count,
		EffectiveTimeZone: values["EffectiveTimeZone"],
	}, nil
}

func parseRow(line string, header *Header, lineNumber int) (Row, error) {
	fields := strings.Split(line, "|")
	if len(fields) != 11 {
		return Row{}, fmt.Errorf("列数=%d，期望 11（ts|dt|o|c|l|h|v|t|cp|cr|p）", len(fields))
	}

	row := Row{}
	row.Ts, _ = strconv.ParseInt(strings.TrimSpace(fields[0]), 10, 64)
	row.SecuCode = header.Values["Code"]
	row.MarketCode, _ = strconv.Atoi(header.Values["MarketCode"])

	// dt（本地时间）→ date/time
	if dt := strings.TrimSpace(fields[1]); len(dt) >= 14 {
		date, err1 := strconv.Atoi(dt[:8])
		clock, err2 := strconv.Atoi(dt[8:14])
		if err1 == nil && err2 == nil {
			row.Date, row.Time = &date, &clock
		}
	}

	// ts 与 dt/EffectiveTimeZone 一致性校验
	if err := validateTsConsistency(row.Ts, fields[1], header.EffectiveTimeZone); err != nil {
		return Row{}, fmt.Errorf("第 %d 行: %w", lineNumber, err)
	}

	// 价格/数量：保留字符串精度（表列为 NUMERIC）
	row.Open = decimalOrNil(fields[2])
	row.Close = decimalOrNil(fields[3])
	row.Low = decimalOrNil(fields[4])
	row.High = decimalOrNil(fields[5])
	row.Volume = int64OrNil(fields[6])
	row.Turnover = decimalOrNil(fields[7])
	// fields[8]=cp 涨跌值、fields[9]=cr 涨跌幅(%) 不落表
	row.PrevClose = decimalOrNil(fields[10])

	return row, nil
}

func validateTsConsistency(ts int64, dtField string, tzName string) error {
	if ts <= 0 || len(dtField) < 14 {
		return nil // 数据缺失时不强校验
	}
	location, err := time.LoadLocation(tzName)
	if err != nil {
		return nil
	}
	local := time.Unix(ts, 0).In(location)
	expected := fmt.Sprintf("%04d%02d%02d%02d%02d%02d",
		local.Year(), local.Month(), local.Day(), local.Hour(), local.Minute(), local.Second())
	if expected != dtField {
		return fmt.Errorf("ts 与 dt/EffectiveTimeZone 不一致（期望 %s，实际 %s）", expected, dtField)
	}
	return nil
}

func decimalOrNil(raw string) *string {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil
	}
	return &value
}

func int64OrNil(raw string) *int64 {
	value := strings.TrimSpace(raw)
	if value == "" {
		return nil
	}
	parsed, err := strconv.ParseInt(value, 10, 64)
	if err != nil {
		return nil
	}
	return &parsed
}
