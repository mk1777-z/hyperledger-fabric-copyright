package middle

import (
	"bytes"
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/jung-kurt/gofpdf"
	"github.com/xuri/excelize/v2"
)

// 导出数据到Excel文件
func ExportDataToExcel(
	timeRange string,
	categoryData []map[string]interface{},
	priceData []map[string]interface{},
	trendData map[string]interface{},
	activityData map[string]interface{},
) ([]byte, error) {
	// 创建新的Excel文件
	f := excelize.NewFile()
	defer func() {
		if err := f.Close(); err != nil {
			fmt.Println("关闭Excel文件失败:", err)
		}
	}()

	// 设置报表标题
	title := fmt.Sprintf("版权统计数据报表 (%s)", timeRange)
	f.SetCellValue("Sheet1", "A1", title)

	// 设置标题样式
	titleStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   14,
			Family: "宋体",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})
	if err == nil {
		f.SetCellStyle("Sheet1", "A1", "A1", titleStyle)
	}
	f.MergeCell("Sheet1", "A1", "G1")

	// 设置副标题
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	f.SetCellValue("Sheet1", "A2", "生成时间: "+currentTime)
	f.MergeCell("Sheet1", "A2", "G2")

	// 设置表头样式
	headerStyle, err := f.NewStyle(&excelize.Style{
		Font: &excelize.Font{
			Bold:   true,
			Size:   11,
			Family: "宋体",
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
		Fill: excelize.Fill{
			Type:    "pattern",
			Color:   []string{"#D9EAD3"},
			Pattern: 1,
		},
		Border: []excelize.Border{
			{Type: "top", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
	})

	// 设置单元格样式
	cellStyle, err := f.NewStyle(&excelize.Style{
		Border: []excelize.Border{
			{Type: "top", Color: "000000", Style: 1},
			{Type: "left", Color: "000000", Style: 1},
			{Type: "right", Color: "000000", Style: 1},
			{Type: "bottom", Color: "000000", Style: 1},
		},
		Alignment: &excelize.Alignment{
			Horizontal: "center",
		},
	})

	// 添加版权分类数据
	f.SetCellValue("Sheet1", "A4", "版权分类统计")
	f.MergeCell("Sheet1", "A4", "B4")
	f.SetCellValue("Sheet1", "A5", "分类名称")
	f.SetCellValue("Sheet1", "B5", "数量")
	if err == nil {
		f.SetCellStyle("Sheet1", "A5", "B5", headerStyle)
	}

	rowNum := 6
	for _, item := range categoryData {
		name := item["name"].(string)
		value := item["value"].(int)
		f.SetCellValue("Sheet1", "A"+strconv.Itoa(rowNum), name)
		f.SetCellValue("Sheet1", "B"+strconv.Itoa(rowNum), value)
		if err == nil {
			f.SetCellStyle("Sheet1", "A"+strconv.Itoa(rowNum), "B"+strconv.Itoa(rowNum), cellStyle)
		}
		rowNum++
	}

	// 添加价格分布数据
	rowNum += 2
	f.SetCellValue("Sheet1", "A"+strconv.Itoa(rowNum), "价格分布")
	f.MergeCell("Sheet1", "A"+strconv.Itoa(rowNum), "B"+strconv.Itoa(rowNum))
	rowNum++
	f.SetCellValue("Sheet1", "A"+strconv.Itoa(rowNum), "价格区间")
	f.SetCellValue("Sheet1", "B"+strconv.Itoa(rowNum), "数量")
	if err == nil {
		f.SetCellStyle("Sheet1", "A"+strconv.Itoa(rowNum), "B"+strconv.Itoa(rowNum), headerStyle)
	}

	rowNum++
	for _, item := range priceData {
		rangeStr := item["range"].(string)
		count := item["count"].(int)
		f.SetCellValue("Sheet1", "A"+strconv.Itoa(rowNum), rangeStr)
		f.SetCellValue("Sheet1", "B"+strconv.Itoa(rowNum), count)
		if err == nil {
			f.SetCellStyle("Sheet1", "A"+strconv.Itoa(rowNum), "B"+strconv.Itoa(rowNum), cellStyle)
		}
		rowNum++
	}

	// 添加交易趋势数据
	rowNum += 2
	f.SetCellValue("Sheet1", "A"+strconv.Itoa(rowNum), "交易趋势")
	f.MergeCell("Sheet1", "A"+strconv.Itoa(rowNum), "C"+strconv.Itoa(rowNum))
	rowNum++
	f.SetCellValue("Sheet1", "A"+strconv.Itoa(rowNum), "月份")
	f.SetCellValue("Sheet1", "B"+strconv.Itoa(rowNum), "交易数量")
	f.SetCellValue("Sheet1", "C"+strconv.Itoa(rowNum), "交易金额")
	if err == nil {
		f.SetCellStyle("Sheet1", "A"+strconv.Itoa(rowNum), "C"+strconv.Itoa(rowNum), headerStyle)
	}

	rowNum++
	months := trendData["months"].([]string)
	counts := trendData["counts"].([]int)
	amounts := trendData["amounts"].([]float64)
	for i := 0; i < len(months); i++ {
		f.SetCellValue("Sheet1", "A"+strconv.Itoa(rowNum), months[i])
		f.SetCellValue("Sheet1", "B"+strconv.Itoa(rowNum), counts[i])
		f.SetCellValue("Sheet1", "C"+strconv.Itoa(rowNum), amounts[i])
		if err == nil {
			f.SetCellStyle("Sheet1", "A"+strconv.Itoa(rowNum), "C"+strconv.Itoa(rowNum), cellStyle)
		}
		rowNum++
	}

	// 添加用户活跃度数据
	rowNum += 2
	f.SetCellValue("Sheet1", "A"+strconv.Itoa(rowNum), "用户活跃度")
	f.MergeCell("Sheet1", "A"+strconv.Itoa(rowNum), "D"+strconv.Itoa(rowNum))
	rowNum++
	f.SetCellValue("Sheet1", "A"+strconv.Itoa(rowNum), "日期")
	f.SetCellValue("Sheet1", "B"+strconv.Itoa(rowNum), "新增用户")
	f.SetCellValue("Sheet1", "C"+strconv.Itoa(rowNum), "活跃用户")
	f.SetCellValue("Sheet1", "D"+strconv.Itoa(rowNum), "交易用户")
	if err == nil {
		f.SetCellStyle("Sheet1", "A"+strconv.Itoa(rowNum), "D"+strconv.Itoa(rowNum), headerStyle)
	}

	rowNum++
	dates := activityData["dates"].([]string)
	newUsers := activityData["newUsers"].([]int)
	activeUsers := activityData["activeUsers"].([]int)
	tradingUsers := activityData["tradingUsers"].([]int)

	for i := 0; i < len(dates); i++ {
		f.SetCellValue("Sheet1", "A"+strconv.Itoa(rowNum), dates[i])
		f.SetCellValue("Sheet1", "B"+strconv.Itoa(rowNum), newUsers[i])
		f.SetCellValue("Sheet1", "C"+strconv.Itoa(rowNum), activeUsers[i])
		f.SetCellValue("Sheet1", "D"+strconv.Itoa(rowNum), tradingUsers[i])
		if err == nil {
			f.SetCellStyle("Sheet1", "A"+strconv.Itoa(rowNum), "D"+strconv.Itoa(rowNum), cellStyle)
		}
		rowNum++
	}

	// 设置列宽
	f.SetColWidth("Sheet1", "A", "A", 15)
	f.SetColWidth("Sheet1", "B", "D", 12)

	// 保存到内存
	buffer, err := f.WriteToBuffer()
	if err != nil {
		return nil, fmt.Errorf("写入Excel到内存失败: %v", err)
	}

	return buffer.Bytes(), nil
}

// 导出数据到PDF文件
func ExportDataToPDF(
	timeRange string,
	categoryData []map[string]interface{},
	priceData []map[string]interface{},
	trendData map[string]interface{},
	activityData map[string]interface{},
) ([]byte, error) {
	// 创建PDF文件 - 使用不需要额外字体的配置
	pdf := gofpdf.New("P", "mm", "A4", "")
	// 使用默认字体，不使用中文字体
	pdf.AddPage()

	// 报告标题 - 使用英文
	title := fmt.Sprintf("Copyright Statistics Report (%s)", timeRange)
	pdf.SetFont("Arial", "B", 14)
	pdf.CellFormat(190, 10, title, "0", 1, "C", false, 0, "")

	// 生成日期 - 使用英文
	pdf.SetFont("Arial", "", 10)
	currentTime := time.Now().Format("2006-01-02 15:04:05")
	pdf.CellFormat(190, 7, "Generated: "+currentTime, "0", 1, "R", false, 0, "")
	pdf.Ln(5)

	log.Printf("开始生成PDF报告，标题: %s, 生成时间: %s", title, currentTime)

	// 版权分类数据 - 使用英文
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(190, 8, "1. Category Distribution", "0", 1, "L", false, 0, "")
	pdf.Ln(2)

	// 创建表格
	pdf.SetFillColor(217, 234, 211)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(60, 7, "Category", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 7, "Count", "1", 1, "C", true, 0, "")

	// 转换分类名称
	categoryMap := map[string]string{
		"音乐": "Music",
		"文学": "Literature",
		"艺术": "Art",
		"摄影": "Photography",
		"影视": "Film",
		"其他": "Other",
	}

	pdf.SetFont("Arial", "", 10)
	for _, item := range categoryData {
		var name string
		var value int

		if nameStr, ok := item["name"].(string); ok {
			// 转换中文分类名称为英文
			if engName, exists := categoryMap[nameStr]; exists {
				name = engName
			} else {
				name = nameStr // 如果没有对应的英文名，保留原名
			}
		} else {
			name = "Unknown"
		}

		if valueInt, ok := item["value"].(int); ok {
			value = valueInt
		} else {
			value = 0
		}

		pdf.CellFormat(60, 7, name, "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 7, strconv.Itoa(value), "1", 1, "C", false, 0, "")
	}
	pdf.Ln(5)

	// 价格分布数据 - 使用英文
	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(190, 8, "2. Price Distribution", "0", 1, "L", false, 0, "")
	pdf.Ln(2)

	pdf.SetFillColor(217, 234, 211)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(60, 7, "Price Range", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 7, "Count", "1", 1, "C", true, 0, "")

	// 转换价格区间名称
	priceRangeMap := map[string]string{
		"0-100":     "0-100",
		"100-500":   "100-500",
		"500-1000":  "500-1000",
		"1000-5000": "1000-5000",
		"5000以上":    "Above 5000",
	}

	pdf.SetFont("Arial", "", 10)
	for _, item := range priceData {
		var rangeStr string
		var count int

		if r, ok := item["range"].(string); ok {
			// 转换中文价格区间为英文
			if engRange, exists := priceRangeMap[r]; exists {
				rangeStr = engRange
			} else {
				rangeStr = r
			}
		} else {
			rangeStr = "Unknown"
		}

		if c, ok := item["count"].(int); ok {
			count = c
		} else {
			count = 0
		}

		pdf.CellFormat(60, 7, rangeStr, "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 7, strconv.Itoa(count), "1", 1, "C", false, 0, "")
	}
	pdf.Ln(5)

	// 交易趋势数据 - 检查是否需要添加第二页
	if pdf.GetY() > 180 {
		pdf.AddPage()
	}

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(190, 8, "3. Transaction Trends", "0", 1, "L", false, 0, "")
	pdf.Ln(2)

	pdf.SetFillColor(217, 234, 211)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(50, 7, "Month", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 7, "Transactions", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 7, "Amount", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 10)
	months := trendData["months"].([]string)
	counts := trendData["counts"].([]int)
	amounts := trendData["amounts"].([]float64)

	for i := 0; i < len(months); i++ {
		pdf.CellFormat(50, 7, months[i], "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 7, strconv.Itoa(counts[i]), "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 7, fmt.Sprintf("%.2f", amounts[i]), "1", 1, "C", false, 0, "")
	}
	pdf.Ln(5)

	// 用户活跃度数据 - 检查是否需要添加第二页
	if pdf.GetY() > 200 {
		pdf.AddPage()
	}

	pdf.SetFont("Arial", "B", 12)
	pdf.CellFormat(190, 8, "4. User Activity", "0", 1, "L", false, 0, "")
	pdf.Ln(2)

	pdf.SetFillColor(217, 234, 211)
	pdf.SetFont("Arial", "B", 10)
	pdf.CellFormat(30, 7, "Date", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 7, "New Users", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 7, "Active Users", "1", 0, "C", true, 0, "")
	pdf.CellFormat(40, 7, "Trading Users", "1", 1, "C", true, 0, "")

	pdf.SetFont("Arial", "", 10)
	dates := activityData["dates"].([]string)
	newUsers := activityData["newUsers"].([]int)
	activeUsers := activityData["activeUsers"].([]int)
	tradingUsers := activityData["tradingUsers"].([]int)

	for i := 0; i < len(dates); i++ {
		pdf.CellFormat(30, 7, dates[i], "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 7, strconv.Itoa(newUsers[i]), "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 7, strconv.Itoa(activeUsers[i]), "1", 0, "C", false, 0, "")
		pdf.CellFormat(40, 7, strconv.Itoa(tradingUsers[i]), "1", 1, "C", false, 0, "")
	}

	// 添加页码
	pdf.SetY(280)
	pdf.SetFont("Arial", "", 8)
	pdf.CellFormat(190, 10, fmt.Sprintf("Page %d", pdf.PageNo()), "0", 0, "C", false, 0, "")

	// 生成PDF流 - 添加错误处理和日志
	var buf bytes.Buffer
	err := pdf.Output(&buf)
	if err != nil {
		log.Printf("PDF生成失败: %v", err)
		return nil, fmt.Errorf("生成PDF失败: %v", err)
	}

	log.Printf("PDF生成成功，大小: %d 字节", buf.Len())
	return buf.Bytes(), nil
}

// 简化 PDF 导出时的类型安全转换辅助函数
func safeStringValue(m map[string]interface{}, key string) string {
	if val, ok := m[key].(string); ok {
		return val
	}
	return ""
}

func safeIntValue(m map[string]interface{}, key string) int {
	if val, ok := m[key].(int); ok {
		return val
	}
	return 0
}
