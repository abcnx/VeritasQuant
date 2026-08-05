// Package mvsv 解析 MVSV-1 分钟级历史行情文件格式。
//
// 格式（与 VeritasQuant/data/Mvsv.py 对齐）：
//   - 头部：# Key : Value 行，空行分隔头部与数据区；
//   - 数据区：列布局由头部 # Field 声明，当前支持两种：
//     ① ts|dt|o|c|l|h|v|t|cp|cr|p（11 列，dt=14 位日期时间，t=成交额）
//     ② ts|d|t|o|c|l|h|v|a|cp|cr|p|pc（13 列，d=8 位日期 + t=6 位时间，a=成交额）
//   其余列（cp 涨跌值 / cr 涨跌幅 / pc）解析但不落表。
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

// columnKind 数据区列语义。
type columnKind int

const (
	colTS columnKind = iota // ts：UTC 时间戳（秒）
	colDate                 // d：8 位交易日期 yyyymmdd
	colDateTime             // dt：14 位日期时间 yyyymmddHHMMSS
	colTime                 // t：6 位时间 HHMMSS
	colOpen                 // o：开盘价
	colClose                // c：收盘价
	colLow                  // l：最低价
	colHigh                 // h：最高价
	colVolume               // v：成交量
	colTurnover             // t（布局 A）/ a（布局 B）：成交额
	colPrevClose            // p：前一收盘价
	colIgnore               // cp / cr / pc：解析但不落表
)

// fieldLayout 一种数据区列布局（列名序列 → 语义序列）。
type fieldLayout struct {
	Field string       // 头部 # Field 声明值
	Kinds []columnKind // 与 Field 列一一对应的语义
}

// supportedLayouts 当前支持的 MVSV 数据区列布局。
var supportedLayouts = map[string]fieldLayout{
	"ts|dt|o|c|l|h|v|t|cp|cr|p": {
		Field: "ts|dt|o|c|l|h|v|t|cp|cr|p",
		Kinds: []columnKind{
			colTS, colDateTime, colOpen, colClose, colLow, colHigh,
			colVolume, colTurnover, colIgnore, colIgnore, colPrevClose,
		},
	},
	"ts|d|t|o|c|l|h|v|a|cp|cr|p|pc": {
		Field: "ts|d|t|o|c|l|h|v|a|cp|cr|p|pc",
		Kinds: []columnKind{
			colTS, colDate, colTime, colOpen, colClose, colLow, colHigh,
			colVolume, colTurnover, colIgnore, colIgnore, colPrevClose, colIgnore,
		},
	},
}

// Header MVSV-1 头部信息。
type Header struct {
	Values            map[string]string
	Count             int
	EffectiveTimeZone string
	Layout            fieldLayout // 数据区列布局（由 # Field 决定）
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
	layout, ok := supportedLayouts[values["Field"]]
	if !ok {
		return nil, fmt.Errorf("Field 布局不支持: %s（支持: %s / %s）",
			values["Field"], "ts|dt|o|c|l|h|v|t|cp|cr|p", "ts|d|t|o|c|l|h|v|a|cp|cr|p|pc")
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
		Layout:            layout,
	}, nil
}

func parseRow(line string, header *Header, lineNumber int) (Row, error) {
	fields := strings.Split(line, "|")
	if len(fields) != len(header.Layout.Kinds) {
		return Row{}, fmt.Errorf("列数=%d，期望 %d（Field: %s）",
			len(fields), len(header.Layout.Kinds), header.Layout.Field)
	}

	row := Row{}
	row.SecuCode = header.Values["Code"]
	row.MarketCode, _ = strconv.Atoi(header.Values["MarketCode"])

	// 本地时间字段（dt 14 位 / d+t 拼接），用于 ts 一致性校验
	localDT := ""

	for index, kind := range header.Layout.Kinds {
		raw := strings.TrimSpace(fields[index])
		switch kind {
		case colTS:
			row.Ts, _ = strconv.ParseInt(raw, 10, 64)
		case colDate:
			// d：8 位日期 yyyymmdd
			if len(raw) >= 8 {
				date, err := strconv.Atoi(raw[:8])
				if err == nil {
					row.Date = &date
				}
			}
			localDT += raw
		case colDateTime:
			// dt：14 位日期时间 yyyymmddHHMMSS
			if len(raw) >= 14 {
				date, err1 := strconv.Atoi(raw[:8])
				clock, err2 := strconv.Atoi(raw[8:14])
				if err1 == nil && err2 == nil {
					row.Date, row.Time = &date, &clock
				}
			}
			localDT += raw
		case colTime:
			// t（布局 B）：6 位时间 HHMMSS
			if len(raw) >= 6 {
				clock, err := strconv.Atoi(raw[:6])
				if err == nil {
					row.Time = &clock
				}
			}
			localDT += raw
		case colOpen:
			row.Open = decimalOrNil(raw)
		case colClose:
			row.Close = decimalOrNil(raw)
		case colLow:
			row.Low = decimalOrNil(raw)
		case colHigh:
			row.High = decimalOrNil(raw)
		case colVolume:
			row.Volume = int64OrNil(raw)
		case colTurnover:
			row.Turnover = decimalOrNil(raw)
		case colPrevClose:
			row.PrevClose = decimalOrNil(raw)
		case colIgnore:
			// cp / cr / pc：不落表
		}
	}

	// ts 与本地时间/EffectiveTimeZone 一致性校验
	if err := validateTsConsistency(row.Ts, localDT, header.EffectiveTimeZone); err != nil {
		return Row{}, fmt.Errorf("第 %d 行: %w", lineNumber, err)
	}
	return row, nil
}

func validateTsConsistency(ts int64, localDT string, tzName string) error {
	if ts <= 0 || len(localDT) < 14 {
		return nil // 数据缺失时不强校验
	}
	location, err := time.LoadLocation(tzName)
	if err != nil {
		return nil
	}
	local := time.Unix(ts, 0).In(location)
	expected := fmt.Sprintf("%04d%02d%02d%02d%02d%02d",
		local.Year(), local.Month(), local.Day(), local.Hour(), local.Minute(), local.Second())
	if expected != localDT {
		return fmt.Errorf("ts 与本地时间/EffectiveTimeZone 不一致（期望 %s，实际 %s）", expected, localDT)
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
