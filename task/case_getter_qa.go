package task

import (
	"context"
	"github.com/360EntSecGroup-Skylar/excelize"
	"smartest-go/models"
	"strconv"
	"strings"
)

/*
用于QA的用例采集
*/

func (QA *QATask) CaseGetterQA(c context.Context) {
	qaModel := models.NewQaBaseTestModel()
	total, err := qaModel.GetQaBaseTestTotal("1=1")
	if err != nil {
		// TODO 异常处理
		return
	}
	resultList, err := qaModel.GetQaBaseTests(0, int(total), QA.QADataSourceConfig.DBFilter)
	if err != nil {
		// TODO 异常处理
		return
	}
	if len(resultList) == 0 {
		// TODO
		return
	}
	QA.req = QA.req[0:0]
	for _, qaBaseTest := range resultList {
		QA.req = append(QA.req, &QATaskReq{
			Id:           qaBaseTest.Id,
			Query:        qaBaseTest.Question,
			ExpectAnswer: strings.Split(qaBaseTest.AnswerList, "&&"),
			ExpectGroup:  qaBaseTest.QaGroupId,
			RobotType:    qaBaseTest.RobotType,
			IsSmoke:      qaBaseTest.IsSmoke,
		})
	}
}

func ExcelQAReader(fileName, sheetName string) (req []*QATaskReq) {
	if !strings.Contains(fileName, "./upload/") {
		fileName = "./upload/" + fileName
	}
	f, err := excelize.OpenFile(fileName)
	if err != nil {
		return
	}
	rows := f.GetRows(sheetName)

	tableHeader := make(map[int]string)
	for index, row := range rows {
		if index == 0 {
			// 记录表头
			for i, cellValue := range row {
				tableHeader[i] = cellValue
			}
			continue
		}
		tmpReq := &QATaskReq{
			Id:           0,
			Query:        "",
			ExpectAnswer: []string{},
			ExpectGroup:  0,
			RobotType:    "",
		}
		for i, cellValue := range row {
			// 记录表数据
			if tableHeader[i] == "id" {
				num, _ := strconv.Atoi(cellValue)
				tmpReq.Id = int64(num)
			}
			if tableHeader[i] == "question" {
				tmpReq.Query = cellValue // 这里先不去处理&& 在pre阶段去统一处理
			}
			if tableHeader[i] == "answer_list" {
				tmpReq.ExpectAnswer = strings.Split(cellValue, "&&")
			}
			if tableHeader[i] == "qa_group_id" {
				num, _ := strconv.Atoi(cellValue)
				tmpReq.ExpectGroup = int64(num)
			}
			if tableHeader[i] == "robot_type" {
				tmpReq.RobotType = cellValue
			}
			if tableHeader[i] == "is_smoke" {
				num, _ := strconv.Atoi(cellValue)
				tmpReq.IsSmoke = num
			}
		}
		req = append(req, tmpReq)
	}
	return
}
