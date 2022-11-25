package task

import (
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"smartest-go/models"
	"smartest-go/pkg/logf"
)

func (KG *KGTask) getResultSummary() string {
	total, _ := models.ReporterDB.MongoCount(kgResultsTable, bson.M{"job_instance_id": KG.KGConfig.JobInstanceId})
	fail, _ := models.ReporterDB.MongoCount(kgResultsTable, bson.M{"job_instance_id": KG.KGConfig.JobInstanceId, "is_pass": false})
	costInfo, _ := models.ReporterDB.MongoAggregate(kgResultsTable, []bson.M{
		{"$match": bson.M{"job_instance_id": KG.KGConfig.JobInstanceId}},
		{"$group": bson.M{
			"_id":     "$job_instance_id",
			"maxCost": bson.M{"$max": "$edg_cost"},
			"minCost": bson.M{"$min": "$edg_cost"},
			"avgCost": bson.M{"$avg": "$edg_cost"},
		}}})
	summary := fmt.Sprintf("%s %s\n用例统计:%d, 错误数:%d, 正确率:%f, 用例并发数:%d\n最大耗时:%d, 最小耗时:%d, 平均耗时:%f",
		KG.KGConfig.TaskName,
		KG.KGConfig.JobInstanceId,
		total, fail, 1-float32(fail)/float32(total), KG.KGConfig.ChanNum,
		costInfo[0].Map()["maxCost"],
		costInfo[0].Map()["minCost"],
		costInfo[0].Map()["avgCost"])
	return summary
}

func (KG *KGTask) sendReport() {
	if KG.KGConfig.IsReport != "yes" || KG.KGConfig.ReportString == nil {
		return
	}
	text := KG.getResultSummary()
	payload := &FeiShu{Payload: &FeiShuPayload{MsgType: "text", Content: &simpleText{Text: text}}}

	for _, url := range KG.KGConfig.ReportString {
		payload.Url = url.Address
		err := reportToFeiShu(payload)
		if err != nil {
			logf.Error(err)
		}
	}
}
