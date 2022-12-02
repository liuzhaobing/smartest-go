package task

import (
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"smartest-go/models"
	"smartest-go/pkg/logf"
)

func (KG *KGTask) writeKGResultExcel() {
	//headers := []map[string]string{
	//	{"key": "id", "label": "用例编号"},
	//	{"key": "question", "label": "测试语句"},
	//	{"key": "answer", "label": "期望答复"},
	//	{"key": "act_answer", "label": "实际答复"},
	//	{"key": "is_pass", "label": "是否通过"},
	//	{"key": "edg_cost", "label": "端测耗时(ms)"},
	//	{"key": "trace_id", "label": "TraceID"},
	//	{"key": "source", "label": "命中类型"},
	//	{"key": "resp_json", "label": "返回JSON"},
	//	{"key": "execute_time", "label": "执行时间"},
	//}
	model := models.NewTaskDataModel()
	result, _ := model.GetTaskDatas(0, 100, "types='base_kg'")
	headers := make([]map[string]string, 0)
	json.Unmarshal([]byte(result[0].Headers), &headers)
	data, _ := models.ReporterDB.MongoFind(kgResultsTable, bson.M{"job_instance_id": KG.JobInstanceId})
	KG.SummaryFile = WriteResultExcel(KnowledgeGraph, KG.JobInstanceId, KG.Summary, headers, data)
}

func (KG *KGTask) getResultSummary() {
	// 标题
	summary := fmt.Sprintf("%s %s\n", KG.KGConfig.TaskName, KG.JobInstanceId)

	// 耗时统计
	costInfo, _ := models.ReporterDB.MongoAggregate(kgResultsTable, []bson.M{
		{"$match": bson.M{"job_instance_id": KG.JobInstanceId}},
		{"$group": bson.M{
			"_id":      "$job_instance_id",
			"max_cost": bson.M{"$max": "$edg_cost"},
			"min_cost": bson.M{"$min": "$edg_cost"},
			"avg_cost": bson.M{"$avg": "$edg_cost"},
		}}})
	summary += fmt.Sprintf("耗时统计:最大耗时:%d, 最小耗时:%d, 平均耗时:%.2f\n", costInfo[0].Map()["max_cost"], costInfo[0].Map()["min_cost"], costInfo[0].Map()["avg_cost"])

	// 版本统计
	total, _ := models.ReporterDB.MongoCount(kgResultsTable, bson.M{"job_instance_id": KG.JobInstanceId})
	fail, _ := models.ReporterDB.MongoCount(kgResultsTable, bson.M{"job_instance_id": KG.JobInstanceId, "is_pass": false})
	summary += fmt.Sprintf("用例统计:%d, 错误数:%d, 正确率:%f\n", total, fail, 1-float32(fail)/float32(total))

	// 附加信息
	summary += fmt.Sprintf("请求参数:前端地址:%s,后端地址:%s,并发数:%d\n", KG.KGConfig.EnvInfo.FrontUrl, KG.KGConfig.EnvInfo.BackendUrl, KG.KGConfig.ChanNum)

	KG.Summary = summary
}

func (KG *KGTask) sendReport() {
	if KG.KGConfig.IsReport != "yes" || KG.KGConfig.ReportString == nil {
		return
	}
	text := KG.Summary
	if KG.SummaryFile != "" {
		text += excelDownloadRouter + KG.SummaryFile
	}
	payload := &FeiShu{Payload: &FeiShuPayload{MsgType: "text", Content: &simpleText{Text: text}}}

	for _, url := range KG.KGConfig.ReportString {
		payload.Url = url.Address
		err := reportToFeiShu(payload)
		if err != nil {
			logf.Error(err)
		}
	}
}
