package task

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"smartest-go/models"
	"smartest-go/pkg/logf"
)

func (Skill *SkillTask) getResultSummary() string {
	// 标题
	summary := fmt.Sprintf("%s %s\n", Skill.SkillConfig.TaskName, Skill.JobInstanceId)

	// 耗时统计
	costInfo, _ := models.ReporterDB.MongoAggregate(SkillResultsTable, []bson.M{
		{"$match": bson.M{"job_instance_id": Skill.JobInstanceId}},
		{"$group": bson.M{
			"_id":      "$job_instance_id",
			"max_cost": bson.M{"$max": "$edg_cost"},
			"min_cost": bson.M{"$min": "$edg_cost"},
			"avg_cost": bson.M{"$avg": "$edg_cost"},
		}}})
	summary += fmt.Sprintf("耗时统计:最大耗时:%d, 最小耗时:%d, 平均耗时:%.2f\n", costInfo[0].Map()["max_cost"], costInfo[0].Map()["min_cost"], costInfo[0].Map()["avg_cost"])

	// 发布必测统计
	isSmokeTotal, _ := models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"job_instance_id": Skill.JobInstanceId, "is_smoke": 1})
	if isSmokeTotal != 0 {
		smokeIntentPass, _ := models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"job_instance_id": Skill.JobInstanceId, "is_smoke": 1, "is_pass": true})
		smokeParamPass, _ := models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"job_instance_id": Skill.JobInstanceId, "is_smoke": 1, "param_info_is_pass": true})
		summary += fmt.Sprintf("★★★发布必测用例总数:%d,错误数:%d,意图正确率:%f,槽位正确率:%f,错误数:%d\n",
			isSmokeTotal,
			isSmokeTotal-smokeIntentPass,
			float32(smokeIntentPass)/float32(isSmokeTotal),
			float32(smokeParamPass)/float32(isSmokeTotal),
			isSmokeTotal-smokeParamPass)
	}

	// 查询版本信息
	caseVersionInfo, _ := models.ReporterDB.MongoAggregate(SkillResultsTable, []bson.M{
		{"$match": bson.M{"job_instance_id": Skill.JobInstanceId}},
		{"$group": bson.M{"_id": bson.M{"case_version": "$case_version"}}},
		{"$sort": bson.M{"_id.case_version": -1}},
		{"$project": bson.M{"_id": 0, "case_version": "$_id.case_version"}},
	})
	FirstVersion := 0.0
	SecondVersion := 0.0
	if len(caseVersionInfo) > 1 {
		if caseVersionInfo[0].Map()["case_version"] != nil {
			FirstVersion, _ = (caseVersionInfo[0].Map()["case_version"]).(float64)
		}
		if caseVersionInfo[1].Map()["case_version"] != nil {
			SecondVersion, _ = (caseVersionInfo[1].Map()["case_version"]).(float64)
		}
	}

	// 最高版本统计
	firstVersionTotal, _ := models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"job_instance_id": Skill.JobInstanceId})
	firstVersionIntentPass, _ := models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"job_instance_id": Skill.JobInstanceId, "is_pass": true})
	firstVersionParamPass, _ := models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"job_instance_id": Skill.JobInstanceId, "param_info_is_pass": true})
	firstVersionRegexCount, _ := models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"job_instance_id": Skill.JobInstanceId, "algo": "regex"})
	firstVersionSystemCount, _ := models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"job_instance_id": Skill.JobInstanceId, "act_source": "system_service"})
	summary += fmt.Sprintf("用例版本:%.2f,用例总数:%d,错误数:%d,意图正确率:%f,槽位正确率:%f,意图支撑中算法占比%f,工程模板占比%f\n",
		FirstVersion,
		firstVersionTotal,
		firstVersionTotal-firstVersionIntentPass,
		float32(firstVersionIntentPass)/float32(firstVersionTotal),
		float32(firstVersionParamPass)/float32(firstVersionTotal),
		1-float32(firstVersionRegexCount)/float32(firstVersionSystemCount),
		float32(firstVersionRegexCount)/float32(firstVersionSystemCount))

	// 第二版本统计
	if SecondVersion != 0.0 {
		secondVersionTotal, _ := models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"case_version": bson.M{"$lte": SecondVersion}, "job_instance_id": Skill.JobInstanceId})
		secondVersionIntentPass, _ := models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"case_version": bson.M{"$lte": SecondVersion}, "job_instance_id": Skill.JobInstanceId, "is_pass": true})
		secondVersionParamPass, _ := models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"case_version": bson.M{"$lte": SecondVersion}, "job_instance_id": Skill.JobInstanceId, "param_info_is_pass": true})
		secondVersionRegexCount, _ := models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"case_version": bson.M{"$lte": SecondVersion}, "job_instance_id": Skill.JobInstanceId, "algo": "regex"})
		secondVersionSystemCount, _ := models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"case_version": bson.M{"$lte": SecondVersion}, "job_instance_id": Skill.JobInstanceId, "act_source": "system_service"})
		summary += fmt.Sprintf("用例版本:%.2f,用例总数:%d,错误数:%d,意图正确率:%f,槽位正确率:%f,意图支撑中算法占比%f,工程模板占比%f\n",
			SecondVersion,
			secondVersionTotal,
			secondVersionTotal-secondVersionIntentPass,
			float32(secondVersionIntentPass)/float32(secondVersionTotal),
			float32(secondVersionParamPass)/float32(secondVersionTotal),
			1-float32(secondVersionRegexCount)/float32(secondVersionSystemCount),
			float32(secondVersionRegexCount)/float32(secondVersionSystemCount))
	}

	// 附加信息
	summary += fmt.Sprintf("请求参数:请求地址:%s,AgentId:%d,并发数:%d\n", Skill.SkillConfig.ConnAddr, Skill.SkillConfig.AgentId, Skill.SkillConfig.ChanNum)
	return summary
}

func (Skill *SkillTask) sendReport() {
	if Skill.SkillConfig.IsReport != "yes" || Skill.SkillConfig.ReportString == nil {
		return
	}
	text := Skill.getResultSummary()
	payload := &FeiShu{Payload: &FeiShuPayload{MsgType: "text", Content: &simpleText{Text: text}}}

	for _, url := range Skill.SkillConfig.ReportString {
		payload.Url = url.Address
		err := reportToFeiShu(payload)
		if err != nil {
			logf.Error(err)
		}
	}
}
