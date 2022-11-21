package models

import (
	"encoding/json"
	"fmt"
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/gin-gonic/gin"
	"smartest-go/pkg/app"
	"smartest-go/pkg/e"
	"smartest-go/pkg/file"
	"smartest-go/pkg/logf"
	utilconst "smartest-go/pkg/util/const"
	"strconv"
	"time"
)

var (
	ExcelCell = []string{"A", "B", "C", "D", "E", "F", "G", "H", "I", "J", "K", "L", "M", "N", "O", "P", "Q", "R", "S", "T", "U", "V", "W", "X", "Y", "Z",
		"AA", "AB", "AC", "AD", "AE", "AF", "AG", "AH", "AI", "AJ", "AK", "AL", "AM", "AN", "AO", "AP", "AQ", "AR", "AS", "AT", "AU", "AV", "AW", "AX", "AY", "AZ"}
)

// ExportExcelByJson convert json to local Excel file and return filename
func ExportExcelByJson(taskName, jsonData string) (filename string) {
	var listMapInstance []map[string]interface{}

	err := json.Unmarshal([]byte(jsonData), &listMapInstance)
	if err != nil {
		fmt.Println(err)
	}

	filename = ExportExcelByMap(taskName, listMapInstance)
	return
}

// ExportExcelByMap convert []map[string]interface{} to local Excel file and return filename
func ExportExcelByMap(taskName string, listMapInstance []map[string]interface{}) (filename string) {
	// get Excel bar
	var tableHeader []string
	for key := range listMapInstance[0] {
		tableHeader = append(tableHeader, key)
	}

	// set Excel bar
	f := excelize.NewFile()
	sheetName1 := "Sheet1"
	f.SetColWidth(sheetName1, ExcelCell[1], ExcelCell[len(ExcelCell)-1], 20)

	for index, header := range tableHeader {
		f.SetCellValue(sheetName1, ExcelCell[index]+"1", header)
	}

	// set Excel data
	count := 2
	for _, mmp := range listMapInstance {
		axis := "A" + strconv.Itoa(count)
		count++
		var oneRowValue []interface{}
		for _, key := range tableHeader {
			if mmp[key] != nil {
				oneRowValue = append(oneRowValue, mmp[key])
			} else {
				oneRowValue = append(oneRowValue, "")
			}
		}
		f.SetSheetRow(sheetName1, axis, &oneRowValue)
	}

	// save Excel file on host
	err := file.IsNotExistMkDir("./runtime/export/")
	if err != nil {
		return
	}
	filename = "./runtime/export/" + taskName + "_" + time.Now().Format("2006-01-02-15-04-05") + ".xlsx"
	if err := f.SaveAs(filename); err != nil {
		logf.Error("filename err :", err)
	}
	return
}

type ExcelExport struct {
	Name string      `json:"name" form:"name"`
	Data interface{} `json:"data" form:"data"`
}

func ExportExcel(context *gin.Context) {
	/*
		jsonFromWeb = {
			"name": "filename prefix",
			"data": [{},{},{}]
		}
	*/
	//BindAndValid
	copyReq := &ExcelExport{}
	err := app.BindAndValid(context, copyReq)
	if err != nil {
		app.ErrorResp(context, e.InvalidParams, err.Error(), nil)
		logf.Debug("Validation", err.Error())
		context.Abort()
		return
	}
	context.Set(utilconst.REQUEST_KEY, copyReq)

	//VerifyReq
	req := context.MustGet(utilconst.REQUEST_KEY).(*ExcelExport)
	jsonData, _ := json.Marshal(req.Data)
	filename := ExportExcelByJson(req.Name, string(jsonData))

	app.SuccessResp(context, struct {
		Count int64  `json:"count"`
		Data  string `json:"data"`
	}{
		Count: 1,
		Data:  filename,
	})
}
