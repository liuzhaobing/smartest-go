package task

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"smartest-go/models"
	"smartest-go/pkg/logf"
)

func (QA *QATask) getResultSummary() string {
	// 标题
	summary := fmt.Sprintf("%s %s\n", QA.QAConfig.TaskName, QA.JobInstanceId)

	// 耗时统计
	costInfo, _ := models.ReporterDB.MongoAggregate(qaResultsTable, []bson.M{
		{"$match": bson.M{"job_instance_id": QA.JobInstanceId}},
		{"$group": bson.M{
			"_id":      "$job_instance_id",
			"max_cost": bson.M{"$max": "$edg_cost"},
			"min_cost": bson.M{"$min": "$edg_cost"},
			"avg_cost": bson.M{"$avg": "$edg_cost"},
		}}})
	summary += fmt.Sprintf("耗时统计:最大耗时:%d, 最小耗时:%d, 平均耗时:%.2f\n", costInfo[0].Map()["max_cost"], costInfo[0].Map()["min_cost"], costInfo[0].Map()["avg_cost"])

	// 发布必测统计
	isSmokeTotal, _ := models.ReporterDB.MongoCount(qaResultsTable, bson.M{"job_instance_id": QA.JobInstanceId, "is_smoke": 1})
	if isSmokeTotal != 0 {
		smokePass, _ := models.ReporterDB.MongoCount(qaResultsTable, bson.M{"job_instance_id": QA.JobInstanceId, "is_smoke": 1, "is_pass": true})
		summary += fmt.Sprintf("★★★发布必测用例总数:%d,错误数:%d,正确率:%f\n", isSmokeTotal, isSmokeTotal-smokePass, float32(smokePass)/float32(isSmokeTotal))
	}

	// 最高版本统计
	firstVersionTotal, _ := models.ReporterDB.MongoCount(qaResultsTable, bson.M{"job_instance_id": QA.JobInstanceId})
	firstVersionIntentPass, _ := models.ReporterDB.MongoCount(qaResultsTable, bson.M{"job_instance_id": QA.JobInstanceId, "is_pass": true})
	summary += fmt.Sprintf("用例总数:%d,错误数:%d,正确率:%f\n",
		firstVersionTotal,
		firstVersionTotal-firstVersionIntentPass,
		float32(firstVersionIntentPass)/float32(firstVersionTotal))

	// 附加信息
	summary += fmt.Sprintf("请求参数:请求地址:%s,AgentId:%d,并发数:%d\n", QA.QAConfig.ConnAddr, QA.QAConfig.AgentId, QA.QAConfig.ChanNum)
	return summary
}

func (QA *QATask) sendReport() {
	if QA.QAConfig.IsReport != "yes" || QA.QAConfig.ReportString == nil {
		return
	}
	text := QA.getResultSummary()
	payload := &FeiShu{Payload: &FeiShuPayload{MsgType: "text", Content: &simpleText{Text: text}}}

	for _, url := range QA.QAConfig.ReportString {
		payload.Url = url.Address
		err := reportToFeiShu(payload)
		if err != nil {
			logf.Error(err)
		}
	}
}
