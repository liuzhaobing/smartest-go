package task

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"smartest-go/models"
	"smartest-go/pkg/logf"
)

func (QA *QATask) getResultSummary() string {
	total, _ := models.ReporterDB.MongoCount(qaResultsTable, bson.M{"job_instance_id": QA.QAConfig.JobInstanceId})
	fail, _ := models.ReporterDB.MongoCount(qaResultsTable, bson.M{"job_instance_id": QA.QAConfig.JobInstanceId, "is_pass": false})
	costInfo, _ := models.ReporterDB.MongoAggregate(qaResultsTable, []bson.M{
		{"$match": bson.M{"job_instance_id": QA.QAConfig.JobInstanceId}},
		{"$group": bson.M{
			"_id":     "$job_instance_id",
			"maxCost": bson.M{"$max": "$edg_cost"},
			"minCost": bson.M{"$min": "$edg_cost"},
			"avgCost": bson.M{"$avg": "$edg_cost"},
		}}})
	summary := fmt.Sprintf("%s %s\n用例统计:%d, 错误数:%d, 正确率:%f, 用例并发数:%d\n最大耗时:%d, 最小耗时:%d, 平均耗时:%f\n请求地址:%s,AgentId:%d",
		QA.QAConfig.TaskName,
		QA.QAConfig.JobInstanceId,
		total, fail, 1-float32(fail)/float32(total), QA.QAConfig.ChanNum,
		costInfo[0].Map()["maxCost"],
		costInfo[0].Map()["minCost"],
		costInfo[0].Map()["avgCost"],
		QA.QAConfig.ConnAddr,
		QA.QAConfig.AgentId)
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
