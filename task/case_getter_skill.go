package task

import (
	"context"
	"smartest-go/models"
)

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
			ExpectParamInfo: SkillBaseTest.ParamInfo,
			UseTest:         SkillBaseTest.UseTest,
			IsSmoke:         SkillBaseTest.IsSmoke,
			CaseVersion:     SkillBaseTest.CaseVersion,
			EditLogs:        SkillBaseTest.EditLogs,
		})
	}
}
