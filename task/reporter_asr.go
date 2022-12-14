package task

import (
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"smartest-go/models"
	"smartest-go/pkg/logf"
	"time"
)

type ASRSummaryToMongo struct {
	TaskName                          string      `bson:"task_name"`
	TaskType                          string      `bson:"task_type"`
	JobInstanceId                     string      `bson:"job_instance_id"`
	TaskConfig                        string      `bson:"task_config"`
	MaxCostASRCloudMinds              interface{} `bson:"max_cost_asr"`
	MinCostASRCloudMinds              interface{} `bson:"min_cost_asr"`
	AvgCostASRCloudMinds              interface{} `bson:"avg_cost_asr"`
	MaxCostASRCtrl                    interface{} `bson:"max_cost_ctrl"`
	MinCostASRCtrl                    interface{} `bson:"min_cost_ctrl"`
	AvgCostASRCtrl                    interface{} `bson:"avg_cost_ctrl"`
	OverView                          string      `bson:"over_view"`
	StartTime                         time.Time   `bson:"start_time"`
	EndTime                           time.Time   `bson:"end_time"`
	ExecuteDate                       string      `bson:"execute_date"`
	SmokeTotal                        int64       `bson:"smoke_total"`
	SmokePassASRCloudMinds            int64       `bson:"smoke_pass_asr"`
	SmokeFailASRCloudMinds            int64       `bson:"smoke_fail_asr"`
	SmokeAccuracyASRCloudMinds        float32     `bson:"smoke_accuracy_asr"`
	SmokePassASRCtrl                  int64       `bson:"smoke_pass_ctrl"`
	SmokeFailASRCtrl                  int64       `bson:"smoke_fail_ctrl"`
	SmokeAccuracyASRCtrl              float32     `bson:"smoke_accuracy_ctrl"`
	FirstVersion                      float32     `bson:"first_version"`
	FirstVersionTotal                 int64       `bson:"first_version_total"`
	FirstVersionPassASRCloudMinds     int64       `bson:"first_version_pass_asr"`
	FirstVersionFailASRCloudMinds     int64       `bson:"first_version_fail_asr"`
	FirstVersionAccuracyASRCloudMinds float32     `bson:"first_version_accuracy_asr"`
	FirstVersionPassASRCtrl           int64       `bson:"first_version_pass_ctrl"`
	FirstVersionFailASRCtrl           int64       `bson:"first_version_fail_ctrl"`
	FirstVersionAccuracyASRCtrl       float32     `bson:"first_version_accuracy_ctrl"`
	ServerExtendMessage               interface{} `bson:"server_extend_message"`
}

func (ASR *ASRTask) writeASRResultExcel() {
	model := models.NewTaskDataModel()
	result, _ := model.GetTaskDatas(0, 100, "types='base_asr'")
	headers := make([]map[string]string, 0)
	json.Unmarshal([]byte(result[0].Headers), &headers)

	data, _ := models.ReporterDB.MongoFind(ASRResultsTable, bson.M{"job_instance_id": ASR.JobInstanceId})
	ASR.SummaryFile = WriteResultExcel(CommonASR, ASR.JobInstanceId, ASR.Summary, headers, data)
}

func (ASR *ASRTask) getResultSummary() {
	nowDate := time.Now().Format("2006-01-02")
	mongoSummary := &ASRSummaryToMongo{StartTime: ASR.startTime, EndTime: ASR.endTime, ExecuteDate: nowDate}
	// 标题
	mongoSummary.TaskName = ASR.ASRConfig.TaskName
	mongoSummary.JobInstanceId = ASR.JobInstanceId
	tmpC, _ := json.Marshal(ASR.ASRConfig)
	mongoSummary.TaskConfig = string(tmpC)
	mongoSummary.TaskType = CommonASR
	summary := fmt.Sprintf("%s %s\n", mongoSummary.TaskName, mongoSummary.JobInstanceId)

	// 服务器版本信息
	data, _ := models.ReporterDB.MongoFind(ASRResultsTable, bson.M{"job_instance_id": ASR.JobInstanceId})
	mongoSummary.ServerExtendMessage = data[0].Map()["server_extend_message"]

	// 耗时统计
	if ASR.ASRConfig.IsASRCloudMinds == "yes" {
		costInfoASR, _ := models.ReporterDB.MongoAggregate(ASRResultsTable, []bson.M{
			{"$match": bson.M{"job_instance_id": ASR.JobInstanceId}},
			{"$group": bson.M{
				"_id":      "$job_instance_id",
				"max_cost": bson.M{"$max": "$edg_cost_asr"},
				"min_cost": bson.M{"$min": "$edg_cost_asr"},
				"avg_cost": bson.M{"$avg": "$edg_cost_asr"},
			}}})
		mongoSummary.MaxCostASRCloudMinds = costInfoASR[0].Map()["max_cost"]
		mongoSummary.MinCostASRCloudMinds = costInfoASR[0].Map()["min_cost"]
		mongoSummary.AvgCostASRCloudMinds = costInfoASR[0].Map()["avg_cost"]
		summary += fmt.Sprintf("ASR CloudMinds 最大耗时:%d, 最小耗时:%d, 平均耗时:%.2f, ", mongoSummary.MaxCostASRCloudMinds, mongoSummary.MinCostASRCloudMinds, mongoSummary.AvgCostASRCloudMinds)
	}
	if ASR.ASRConfig.IsASRCtrl == "yes" {
		costInfoASRCtrl, _ := models.ReporterDB.MongoAggregate(ASRResultsTable, []bson.M{
			{"$match": bson.M{"job_instance_id": ASR.JobInstanceId}},
			{"$group": bson.M{
				"_id":      "$job_instance_id",
				"max_cost": bson.M{"$max": "$edg_cost_asr_ctrl"},
				"min_cost": bson.M{"$min": "$edg_cost_asr_ctrl"},
				"avg_cost": bson.M{"$avg": "$edg_cost_asr_ctrl"},
			}}})
		mongoSummary.MaxCostASRCtrl = costInfoASRCtrl[0].Map()["max_cost"]
		mongoSummary.MinCostASRCtrl = costInfoASRCtrl[0].Map()["min_cost"]
		mongoSummary.AvgCostASRCtrl = costInfoASRCtrl[0].Map()["avg_cost"]
		summary += fmt.Sprintf("ASR Control 最大耗时:%d, 最小耗时:%d, 平均耗时:%.2f", mongoSummary.MaxCostASRCtrl, mongoSummary.MinCostASRCtrl, mongoSummary.AvgCostASRCtrl)
	}
	summary += "\n"

	// 发布必测统计
	isSmokeTotal, _ := models.ReporterDB.MongoCount(ASRResultsTable, bson.M{"job_instance_id": ASR.JobInstanceId, "is_smoke": 1})
	if isSmokeTotal > 0 {
		smokePassASRCloudMinds, _ := models.ReporterDB.MongoCount(ASRResultsTable, bson.M{"job_instance_id": ASR.JobInstanceId, "is_smoke": 1, "is_asr_pass": true})
		smokePassASRCtrl, _ := models.ReporterDB.MongoCount(ASRResultsTable, bson.M{"job_instance_id": ASR.JobInstanceId, "is_smoke": 1, "is_asr_ctrl_pass": true})

		mongoSummary.SmokeTotal = isSmokeTotal
		summary += fmt.Sprintf("★★★服务检测用例总数:%d, ", mongoSummary.SmokeTotal)

		if ASR.ASRConfig.IsASRCloudMinds == "yes" {
			mongoSummary.SmokePassASRCloudMinds = smokePassASRCloudMinds
			mongoSummary.SmokeFailASRCloudMinds = isSmokeTotal - smokePassASRCloudMinds
			mongoSummary.SmokeAccuracyASRCloudMinds = float32(mongoSummary.SmokePassASRCloudMinds) / float32(mongoSummary.SmokeTotal)
			summary += fmt.Sprintf("ASR CloudMinds 错误数:%d, 正确率:%f, ", mongoSummary.SmokeFailASRCloudMinds, mongoSummary.SmokeAccuracyASRCloudMinds)
		}
		if ASR.ASRConfig.IsASRCtrl == "yes" {
			mongoSummary.SmokePassASRCtrl = smokePassASRCtrl
			mongoSummary.SmokeFailASRCtrl = isSmokeTotal - smokePassASRCtrl
			mongoSummary.SmokeAccuracyASRCtrl = float32(mongoSummary.SmokePassASRCtrl) / float32(mongoSummary.SmokeTotal)
			summary += fmt.Sprintf("ASR Control 错误数:%d, 正确率:%f", mongoSummary.SmokeFailASRCtrl, mongoSummary.SmokeAccuracyASRCtrl)
		}
		summary += "\n"
	}
	// 最高版本统计
	mongoSummary.FirstVersionTotal, _ = models.ReporterDB.MongoCount(ASRResultsTable, bson.M{"job_instance_id": ASR.JobInstanceId})
	summary += fmt.Sprintf("用例总数:%d, ", mongoSummary.FirstVersionTotal)

	if ASR.ASRConfig.IsASRCloudMinds == "yes" {
		mongoSummary.FirstVersionPassASRCloudMinds, _ = models.ReporterDB.MongoCount(ASRResultsTable, bson.M{"job_instance_id": ASR.JobInstanceId, "is_asr_pass": true})
		mongoSummary.FirstVersionFailASRCloudMinds = mongoSummary.FirstVersionTotal - mongoSummary.FirstVersionPassASRCloudMinds
		mongoSummary.FirstVersionAccuracyASRCloudMinds = float32(mongoSummary.FirstVersionPassASRCloudMinds) / float32(mongoSummary.FirstVersionTotal)
		summary += fmt.Sprintf("ASR CloudMinds 错误数:%d, 正确率:%f, ", mongoSummary.FirstVersionFailASRCloudMinds, mongoSummary.FirstVersionAccuracyASRCloudMinds)
	}

	if ASR.ASRConfig.IsASRCtrl == "yes" {
		mongoSummary.FirstVersionPassASRCtrl, _ = models.ReporterDB.MongoCount(ASRResultsTable, bson.M{"job_instance_id": ASR.JobInstanceId, "is_asr_ctrl_pass": true})
		mongoSummary.FirstVersionFailASRCtrl = mongoSummary.FirstVersionTotal - mongoSummary.FirstVersionPassASRCtrl
		mongoSummary.FirstVersionAccuracyASRCtrl = float32(mongoSummary.FirstVersionPassASRCtrl) / float32(mongoSummary.FirstVersionTotal)
		summary += fmt.Sprintf("ASR Control 错误数:%d, 正确率:%f", mongoSummary.FirstVersionFailASRCtrl, mongoSummary.FirstVersionAccuracyASRCtrl)
	}
	summary += "\n"

	// 附加信息

	summary += fmt.Sprintf("请求参数:ASR CloudMinds 请求地址:%s, ASR Control 请求地址:%s, AgentId:%d, 并发数:%d\n服务版本信息:%s\n",
		ASR.ASRConfig.ASRAddr, ASR.ASRConfig.ASRCtrlAddr, ASR.ASRConfig.AgentId, ASR.ASRConfig.ChanNum, mongoSummary.ServerExtendMessage)
	ASR.Summary = summary

	// 测试总结存储到mongo
	mongoSummary.OverView = ASR.Summary
	models.ReporterDB.MongoInsertOne(MongoSummaryTable, mongoSummary)
}

func (ASR *ASRTask) sendReport() {
	if ASR.ASRConfig.IsReport != "yes" || ASR.ASRConfig.ReportString == nil {
		return
	}
	text := ASR.Summary
	if ASR.SummaryFile != "" {
		text += excelDownloadRouter + ASR.SummaryFile
	}
	payload := &FeiShu{Payload: &FeiShuPayload{MsgType: "text", Content: &simpleText{Text: text}}}

	for _, url := range ASR.ASRConfig.ReportString {
		payload.Url = url.Address
		err := reportToFeiShu(payload)
		if err != nil {
			logf.Error(err)
		}
	}
}
