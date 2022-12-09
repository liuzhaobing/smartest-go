package task

import (
	"context"
	"github.com/360EntSecGroup-Skylar/excelize"
	"smartest-go/models"
	"strconv"
	"strings"
)

/*
用于Skill的用例采集
*/

func (Skill *SkillTask) CaseGetterSkill(c context.Context) {
	SkillModel := models.NewSkillBaseTestModel()
	total, err := SkillModel.GetSkillBaseTestTotal("1=1")
	if err != nil {
		// TODO 异常处理
		return
	}
	resultList, err := SkillModel.GetSkillBaseTests(0, int(total), Skill.SkillDataSourceConfig.DBFilter)
	if err != nil {
		// TODO 异常处理
		return
	}
	if len(resultList) == 0 {
		// TODO
		return
	}
	Skill.req = Skill.req[0:0]
	for _, SkillBaseTest := range resultList {
		Skill.req = append(Skill.req, &SkillTaskReq{
			Id:              SkillBaseTest.Id,
			Query:           SkillBaseTest.Question,
			ExpectSource:    SkillBaseTest.Source,
			ExpectDomain:    SkillBaseTest.Domain,
			ExpectIntent:    SkillBaseTest.Intent,
			SkillSource:     SkillBaseTest.SkillSource,
			SkillCn:         SkillBaseTest.SkillCn,
			RobotType:       SkillBaseTest.RobotType,
			RobotID:         SkillBaseTest.RobotId,
			ExpectParams:    SkillBaseTest.Params,
			ExpectParamInfo: SkillBaseTest.ParamInfo,
			UseTest:         SkillBaseTest.UseTest,
			IsSmoke:         SkillBaseTest.IsSmoke,
			CaseVersion:     SkillBaseTest.CaseVersion,
			EditLogs:        SkillBaseTest.EditLogs,
		})
	}
}

func ExcelSkillReader(fileName, sheetName string) (req []*SkillTaskReq) {
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
		tmpReq := &SkillTaskReq{}
		for i, cellValue := range row {
			// 记录表数据
			if tableHeader[i] == "id" {
				num, _ := strconv.Atoi(cellValue)
				tmpReq.Id = int64(num)
			}
			if tableHeader[i] == "question" {
				tmpReq.Query = cellValue
			}
			if tableHeader[i] == "source" {
				tmpReq.ExpectSource = cellValue
			}

			if tableHeader[i] == "domain" {
				tmpReq.ExpectDomain = cellValue
			}

			if tableHeader[i] == "intent" {
				tmpReq.ExpectIntent = cellValue
			}

			if tableHeader[i] == "skill_source" {
				tmpReq.SkillSource = cellValue
			}

			if tableHeader[i] == "skill_cn" {
				tmpReq.SkillCn = cellValue
			}

			if tableHeader[i] == "robot_id" {
				tmpReq.RobotID = cellValue
			}

			if tableHeader[i] == "usetest" {
				num, _ := strconv.Atoi(cellValue)
				tmpReq.UseTest = num
			}

			if tableHeader[i] == "is_smoke" {
				num, _ := strconv.Atoi(cellValue)
				tmpReq.IsSmoke = num
			}

			if tableHeader[i] == "paraminfo" {
				tmpReq.ExpectParamInfo = cellValue
			}

			if tableHeader[i] == "case_version" {
				num64, _ := strconv.ParseFloat(cellValue, 32)
				tmpReq.CaseVersion = float32(num64)
			}

			if tableHeader[i] == "robot_type" {
				tmpReq.RobotType = cellValue
			}
		}
		req = append(req, tmpReq)
	}
	return
}
