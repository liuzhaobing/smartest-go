package task

import (
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"smartest-go/models"
	"smartest-go/pkg/logf"
	"time"
)

type QASummaryToMongo struct {
	TaskName             string      `bson:"task_name"`
	TaskType             string      `bson:"task_type"`
	JobInstanceId        string      `bson:"job_instance_id"`
	TaskConfig           string      `bson:"task_config"`
	MaxCost              interface{} `bson:"max_cost"`
	MinCost              interface{} `bson:"min_cost"`
	AvgCost              interface{} `bson:"avg_cost"`
	OverView             string      `bson:"over_view"`
	StartTime            time.Time   `bson:"start_time"`
	EndTime              time.Time   `bson:"end_time"`
	SmokeTotal           int64       `bson:"smoke_total"`
	SmokePass            int64       `bson:"smoke_pass"`
	SmokeFail            int64       `bson:"smoke_fail"`
	SmokeAccuracy        float32     `bson:"smoke_accuracy"`
	FirstVersion         float32     `bson:"first_version"`
	FirstVersionTotal    int64       `bson:"first_version_total"`
	FirstVersionPass     int64       `bson:"first_version_pass"`
	FirstVersionFail     int64       `bson:"first_version_fail"`
	FirstVersionAccuracy float32     `bson:"first_version_accuracy"`
}

func (QA *QATask) writeQAResultExcel() {
	//headers := []map[string]string{
	//	{"key": "id", "label": "用例编号"},
	//	{"key": "question", "label": "测试语句"},
	//	{"key": "exp_answer", "label": "期望回复（包含指定内容）"},
	//	{"key": "act_answer", "label": "实际回复"},
	//	{"key": "source", "label": "回复source"},
	//	{"key": "exp_group_id", "label": "QA的期望GroupId"},
	//	{"key": "act_group_id", "label": "QA的GroupId"},
	//	{"key": "is_pass", "label": "是否通过"},
	//	{"key": "is_group_id_pass", "label": "group_id是否通过"},
	//	{"key": "is_full_pass", "label": "是否完全匹配"},
	//	{"key": "is_smoke", "label": "发布必测"},
	//	{"key": "edg_cost", "label": "端测耗时(ms)"},
	//	{"key": "trace_id", "label": "TranceId"},
	//	{"key": "algo_score", "label": "算法得分score"},
	//}
	model := models.NewTaskDataModel()
	result, _ := model.GetTaskDatas(0, 100, "types='base_qa'")
	headers := make([]map[string]string, 0)
	json.Unmarshal([]byte(result[0].Headers), &headers)

	data, _ := models.ReporterDB.MongoFind(qaResultsTable, bson.M{"job_instance_id": QA.JobInstanceId})
	QA.SummaryFile = WriteResultExcel(CommonQA, QA.JobInstanceId, QA.Summary, headers, data)
}

func (QA *QATask) getResultSummary() {
	mongoSummary := &QASummaryToMongo{StartTime: QA.startTime, EndTime: QA.endTime}
	// 标题
	mongoSummary.TaskName = QA.QAConfig.TaskName
	mongoSummary.JobInstanceId = QA.JobInstanceId
	tmpC, _ := json.Marshal(QA.QAConfig)
	mongoSummary.TaskConfig = string(tmpC)
	mongoSummary.TaskType = CommonQA
	summary := fmt.Sprintf("%s %s\n", mongoSummary.TaskName, mongoSummary.JobInstanceId)

	// 耗时统计
	costInfo, _ := models.ReporterDB.MongoAggregate(qaResultsTable, []bson.M{
		{"$match": bson.M{"job_instance_id": QA.JobInstanceId}},
		{"$group": bson.M{
			"_id":      "$job_instance_id",
			"max_cost": bson.M{"$max": "$edg_cost"},
			"min_cost": bson.M{"$min": "$edg_cost"},
			"avg_cost": bson.M{"$avg": "$edg_cost"},
		}}})
	mongoSummary.MaxCost = costInfo[0].Map()["max_cost"]
	mongoSummary.MinCost = costInfo[0].Map()["min_cost"]
	mongoSummary.AvgCost = costInfo[0].Map()["avg_cost"]
	summary += fmt.Sprintf("耗时统计:最大耗时:%d, 最小耗时:%d, 平均耗时:%.2f\n", mongoSummary.MaxCost, mongoSummary.MinCost, mongoSummary.AvgCost)

	// 发布必测统计
	isSmokeTotal, _ := models.ReporterDB.MongoCount(qaResultsTable, bson.M{"job_instance_id": QA.JobInstanceId, "is_smoke": 1})
	if isSmokeTotal > 0 {
		mongoSummary.SmokeTotal = isSmokeTotal
		mongoSummary.SmokePass, _ = models.ReporterDB.MongoCount(qaResultsTable, bson.M{"job_instance_id": QA.JobInstanceId, "is_smoke": 1, "algo": "qqsim"})
		mongoSummary.SmokeFail = mongoSummary.SmokeTotal - mongoSummary.SmokePass
		mongoSummary.SmokeAccuracy = float32(mongoSummary.SmokePass) / float32(mongoSummary.SmokeTotal)
		summary += fmt.Sprintf("★★★qqsim算法检测总数:%d,未匹配数:%d,匹配率:%f\n", mongoSummary.SmokeTotal, mongoSummary.SmokeFail, mongoSummary.SmokeAccuracy)
	}

	// 最高版本统计
	mongoSummary.FirstVersionTotal, _ = models.ReporterDB.MongoCount(qaResultsTable, bson.M{"job_instance_id": QA.JobInstanceId})
	mongoSummary.FirstVersionPass, _ = models.ReporterDB.MongoCount(qaResultsTable, bson.M{"job_instance_id": QA.JobInstanceId, "is_pass": true})
	mongoSummary.FirstVersionFail = mongoSummary.FirstVersionTotal - mongoSummary.FirstVersionPass
	mongoSummary.FirstVersionAccuracy = float32(mongoSummary.FirstVersionPass) / float32(mongoSummary.FirstVersionTotal)
	summary += fmt.Sprintf("用例总数:%d,错误数:%d,正确率:%f\n", mongoSummary.FirstVersionTotal, mongoSummary.FirstVersionFail, mongoSummary.FirstVersionAccuracy)

	// 附加信息
	summary += fmt.Sprintf("请求参数:请求地址:%s,AgentId:%d,并发数:%d\n", QA.QAConfig.ConnAddr, QA.QAConfig.AgentId, QA.QAConfig.ChanNum)
	QA.Summary = summary

	// 测试总结存储到mongo
	mongoSummary.OverView = QA.Summary
	models.ReporterDB.MongoInsertOne(MongoSummaryTable, mongoSummary)
}

func (QA *QATask) sendReport() {
	if QA.QAConfig.IsReport != "yes" || QA.QAConfig.ReportString == nil {
		return
	}
	text := QA.Summary
	if QA.SummaryFile != "" {
		text += excelDownloadRouter + QA.SummaryFile
	}
	payload := &FeiShu{Payload: &FeiShuPayload{MsgType: "text", Content: &simpleText{Text: text}}}

	for _, url := range QA.QAConfig.ReportString {
		payload.Url = url.Address
		err := reportToFeiShu(payload)
		if err != nil {
			logf.Error(err)
		}
	}
}
