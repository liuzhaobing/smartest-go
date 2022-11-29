package task

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"smartest-go/models"
	"smartest-go/pkg/logf"
)

func (Skill *SkillTask) getResultSummary() string {
	total, _ := models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"job_instance_id": Skill.SkillConfig.JobInstanceId})
	fail, _ := models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"job_instance_id": Skill.SkillConfig.JobInstanceId, "is_pass": false})
	costInfo, _ := models.ReporterDB.MongoAggregate(SkillResultsTable, []bson.M{
		{"$match": bson.M{"job_instance_id": Skill.SkillConfig.JobInstanceId}},
		{"$group": bson.M{
			"_id":     "$job_instance_id",
			"maxCost": bson.M{"$max": "$edg_cost"},
			"minCost": bson.M{"$min": "$edg_cost"},
			"avgCost": bson.M{"$avg": "$edg_cost"},
		}}})
	summary := fmt.Sprintf("%s %s\n用例统计:%d, 错误数:%d, 正确率:%f, 用例并发数:%d\n最大耗时:%d, 最小耗时:%d, 平均耗时:%f\n请求地址:%s,AgentId:%d",
		Skill.SkillConfig.TaskName,
		Skill.SkillConfig.JobInstanceId,
		total, fail, 1-float32(fail)/float32(total), Skill.SkillConfig.ChanNum,
		costInfo[0].Map()["maxCost"],
		costInfo[0].Map()["minCost"],
		costInfo[0].Map()["avgCost"],
		Skill.SkillConfig.ConnAddr,
		Skill.SkillConfig.AgentId)
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
