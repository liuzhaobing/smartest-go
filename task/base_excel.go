package task

import (
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"go.mongodb.org/mongo-driver/bson"
	"smartest-go/models"
	"smartest-go/pkg/file"
	"smartest-go/pkg/logf"
	"strconv"
	"time"
)

// WriteResultExcel 将数据写入到本地excel文件
func WriteResultExcel(taskType, JobInstanceId, summary string, headers []map[string]string, data []*bson.D) string {
	// 初始化excel对象
	f := excelize.NewFile()
	sheetName1 := "Sheet1"
	f.SetColWidth(sheetName1, models.ExcelCell[1], models.ExcelCell[len(models.ExcelCell)-1], 20)

	// 首行summary
	f.SetCellValue(sheetName1, "A1", summary)
	f.SetRowHeight(sheetName1, 1, 60)
	f.MergeCell(sheetName1, "A1", "F1")

	// 表头
	for index, header := range headers {
		f.SetCellValue(sheetName1, models.ExcelCell[index]+"2", header["label"])
	}

	// 表数据
	count := 3
	for _, mp := range data {
		mmp := mp.Map()
		axis := "A" + strconv.Itoa(count)
		count++
		var oneRowValue []interface{}
		for _, header := range headers {
			if mmp[header["key"]] != nil {
				oneRowValue = append(oneRowValue, mmp[header["key"]])
			} else {
				oneRowValue = append(oneRowValue, "")
			}
		}
		f.SetSheetRow(sheetName1, axis, &oneRowValue)
	}

	// 写磁盘
	savePath := fmt.Sprintf(`./runtime/%s/`, taskType)                                                // 路径
	saveName := taskType + "-" + time.Now().Format("20060102-150405") + "-" + JobInstanceId + ".xlsx" // 文件名
	err := file.IsNotExistMkDir(savePath)
	if err != nil {
		return ""
	}
	filename := savePath + saveName // 文件全路径
	if err := f.SaveAs(filename); err != nil {
		logf.Error("filename err :", err)
	}
	return filename
}

var (
	excelDownloadRouter = "http://172.16.23.33:27997/api/v1/download?filename="
)
