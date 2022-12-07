package task

import (
	"encoding/json"
	"fmt"
	"go.mongodb.org/mongo-driver/bson"
	"smartest-go/models"
	"smartest-go/pkg/logf"
	"time"
)

type SkillSummaryToMongo struct {
	TaskName                       string      `bson:"task_name"`
	TaskType                       string      `bson:"task_type"`
	JobInstanceId                  string      `bson:"job_instance_id"`
	TaskConfig                     string      `bson:"task_config"`
	MaxCost                        interface{} `bson:"max_cost"`
	MinCost                        interface{} `bson:"min_cost"`
	AvgCost                        interface{} `bson:"avg_cost"`
	OverView                       string      `bson:"over_view"`
	StartTime                      time.Time   `bson:"start_time"`
	EndTime                        time.Time   `bson:"end_time"`
	SmokeTotal                     int64       `bson:"smoke_total"`
	SmokeIntentPass                int64       `bson:"smoke_intent_pass"`
	SmokeIntentFail                int64       `bson:"smoke_intent_fail"`
	SmokeIntentAccuracy            float32     `bson:"smoke_intent_accuracy"`
	SmokeParaminfoPass             int64       `bson:"smoke_paraminfo_pass"`
	SmokeParaminfoFail             int64       `bson:"smoke_paraminfo_fail"`
	SmokeParaminfoAccuracy         float32     `bson:"smoke_paraminfo_accuracy"`
	SmokeAlgoPercent               float32     `bson:"smoke_algo_percent"`
	SmokeRegexPercent              float32     `bson:"smoke_regex_percent"`
	FirstVersion                   float32     `bson:"first_version"`
	FirstVersionTotal              int64       `bson:"first_version_total"`
	FirstVersionIntentPass         int64       `bson:"first_version_intent_pass"`
	FirstVersionIntentFail         int64       `bson:"first_version_intent_fail"`
	FirstVersionIntentAccuracy     float32     `bson:"first_version_intent_accuracy"`
	FirstVersionParaminfoPass      int64       `bson:"first_version_paraminfo_pass"`
	FirstVersionParaminfoFail      int64       `bson:"first_version_paraminfo_fail"`
	FirstVersionParaminfoAccuracy  float32     `bson:"first_version_paraminfo_accuracy"`
	FirstVersionAlgoCount          int64       `bson:"first_version_algo_count"`
	FirstVersionRegexCount         int64       `bson:"first_version_regex_count"`
	FirstVersionAlgoPercent        float32     `bson:"first_version_algo_percent"`
	FirstVersionRegexPercent       float32     `bson:"first_version_regex_percent"`
	SecondVersion                  float32     `bson:"second_version"`
	SecondVersionTotal             int64       `bson:"second_version_total"`
	SecondVersionIntentPass        int64       `bson:"second_version_intent_pass"`
	SecondVersionIntentFail        int64       `bson:"second_version_intent_fail"`
	SecondVersionIntentAccuracy    float32     `bson:"second_version_intent_accuracy"`
	SecondVersionParaminfoPass     int64       `bson:"second_version_paraminfo_pass"`
	SecondVersionParaminfoFail     int64       `bson:"second_version_paraminfo_fail"`
	SecondVersionParaminfoAccuracy float32     `bson:"second_version_paraminfo_accuracy"`
	SecondVersionAlgoCount         int64       `bson:"second_version_algo_count"`
	SecondVersionRegexCount        int64       `bson:"second_version_regex_count"`
	SecondVersionAlgoPercent       float32     `bson:"second_version_algo_percent"`
	SecondVersionRegexPercent      float32     `bson:"second_version_regex_percent"`
}

func (Skill *SkillTask) writeSkillResultExcel() {
	//headers := []map[string]string{
	//	{"key": "id", "label": "用例编号"},
	//	{"key": "question", "label": "测试语句"},
	//	{"key": "source", "label": "期望source"},
	//	{"key": "act_source", "label": "实际source"},
	//	{"key": "domain", "label": "期望domain"},
	//	{"key": "act_domain", "label": "实际domain"},
	//	{"key": "intent", "label": "期望intent"},
	//	{"key": "act_intent", "label": "实际intent(hitlog.intent)"},
	//	{"key": "is_pass", "label": "意图是否通过"},
	//	{"key": "act_intent_tts", "label": "实际intent(tts...intent)"},
	//	{"key": "is_smoke", "label": "发布必测"},
	//	{"key": "parameters", "label": "params实际值"},
	//	{"key": "edg_cost", "label": "端测耗时(ms)"},
	//	{"key": "paraminfo", "label": "槽位ParamInfo期望值"},
	//	{"key": "act_param_info", "label": "槽位ParamInfo实际值"},
	//	{"key": "param_info_is_pass", "label": "槽位是否通过"},
	//	{"key": "answer_string", "label": "回答内容"},
	//	{"key": "answer_url", "label": "回答Url"},
	//	{"key": "case_version", "label": "用例版本"},
	//	{"key": "algo", "label": "Algo"},
	//	{"key": "algo_score", "label": "算法得分"},
	//	{"key": "act_input_context", "label": "多轮input context"},
	//	{"key": "robot_id", "label": "多轮组RobotId"},
	//	{"key": "trace_id", "label": "TranceId"},
	//	{"key": "act_robot_type", "label": "机器人类型"},
	//	{"key": "nlu_debug_info", "label": "NLUDebugInfo"},
	//	{"key": "entity_trie", "label": "EntityTrie"},
	//	{"key": "ner_trie", "label": "NERTrie"},
	//	{"key": "fail_reason", "label": "失败原因"},
	//	{"key": "filter_developer", "label": "BUG初筛责任人"},
	//	{"key": "assign_reason", "label": "BUG流转说明"},
	//	{"key": "fix_developer", "label": "BUG修复责任人"},
	//	{"key": "bug_status", "label": "BUG解决状态"},
	//}
	model := models.NewTaskDataModel()
	result, _ := model.GetTaskDatas(0, 100, "types='base_skill'")
	headers := make([]map[string]string, 0)
	json.Unmarshal([]byte(result[0].Headers), &headers)
	data, _ := models.ReporterDB.MongoFind(SkillResultsTable, bson.M{"job_instance_id": Skill.JobInstanceId})
	Skill.SummaryFile = WriteResultExcel(SystemSkill, Skill.JobInstanceId, Skill.Summary, headers, data)
}

func (Skill *SkillTask) getResultSummary() {
	mongoSummary := &SkillSummaryToMongo{StartTime: Skill.startTime, EndTime: Skill.endTime}
	// 标题
	mongoSummary.TaskName = Skill.SkillConfig.TaskName
	mongoSummary.JobInstanceId = Skill.JobInstanceId
	tmpC, _ := json.Marshal(Skill.SkillConfig)
	mongoSummary.TaskConfig = string(tmpC)
	mongoSummary.TaskType = SystemSkill
	summary := fmt.Sprintf("%s %s\n", mongoSummary.TaskName, mongoSummary.JobInstanceId)

	// 耗时统计
	costInfo, _ := models.ReporterDB.MongoAggregate(SkillResultsTable, []bson.M{
		{"$match": bson.M{"job_instance_id": Skill.JobInstanceId}},
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
	isSmokeTotal, _ := models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"job_instance_id": Skill.JobInstanceId, "is_smoke": 1})
	if isSmokeTotal > 0 {
		smokeIntentPass, _ := models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"job_instance_id": Skill.JobInstanceId, "is_smoke": 1, "is_pass": true})
		smokeParamPass, _ := models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"job_instance_id": Skill.JobInstanceId, "is_smoke": 1, "param_info_is_pass": true})

		mongoSummary.SmokeTotal = isSmokeTotal
		mongoSummary.SmokeIntentPass = smokeIntentPass
		mongoSummary.SmokeIntentFail = isSmokeTotal - smokeIntentPass
		mongoSummary.SmokeIntentAccuracy = float32(mongoSummary.SmokeIntentPass) / float32(mongoSummary.SmokeTotal)
		mongoSummary.SmokeParaminfoPass = smokeParamPass
		mongoSummary.SmokeParaminfoFail = isSmokeTotal - smokeParamPass
		mongoSummary.SmokeParaminfoAccuracy = float32(mongoSummary.SmokeParaminfoPass) / float32(mongoSummary.SmokeTotal)
		summary += fmt.Sprintf("★★★发布必测用例总数:%d,错误数:%d,意图正确率:%f,槽位正确率:%f,错误数:%d\n",
			mongoSummary.SmokeTotal,
			mongoSummary.SmokeIntentFail,
			mongoSummary.SmokeIntentAccuracy,
			mongoSummary.SmokeParaminfoAccuracy,
			mongoSummary.SmokeParaminfoFail)
	}

	// 查询版本信息
	caseVersionInfo, _ := models.ReporterDB.MongoAggregate(SkillResultsTable, []bson.M{
		{"$match": bson.M{"job_instance_id": Skill.JobInstanceId}},
		{"$group": bson.M{"_id": bson.M{"case_version": "$case_version"}}},
		{"$sort": bson.M{"_id.case_version": -1}},
		{"$project": bson.M{"_id": 0, "case_version": "$_id.case_version"}},
	})

	if len(caseVersionInfo) > 1 {
		if caseVersionInfo[0].Map()["case_version"] != nil {
			mongoSummary.FirstVersion = float32((caseVersionInfo[0].Map()["case_version"]).(float64))
		}
		if caseVersionInfo[1].Map()["case_version"] != nil {
			mongoSummary.SecondVersion = float32((caseVersionInfo[1].Map()["case_version"]).(float64))
		}
	}

	// 最高版本统计
	mongoSummary.FirstVersionTotal, _ = models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"job_instance_id": Skill.JobInstanceId})
	mongoSummary.FirstVersionIntentPass, _ = models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"job_instance_id": Skill.JobInstanceId, "is_pass": true})
	mongoSummary.FirstVersionIntentFail = mongoSummary.FirstVersionTotal - mongoSummary.FirstVersionIntentPass
	mongoSummary.FirstVersionIntentAccuracy = float32(mongoSummary.FirstVersionIntentPass) / float32(mongoSummary.FirstVersionTotal)
	mongoSummary.FirstVersionParaminfoPass, _ = models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"job_instance_id": Skill.JobInstanceId, "param_info_is_pass": true})
	mongoSummary.FirstVersionParaminfoFail = mongoSummary.FirstVersionTotal - mongoSummary.FirstVersionParaminfoPass
	mongoSummary.FirstVersionParaminfoAccuracy = float32(mongoSummary.FirstVersionParaminfoPass) / float32(mongoSummary.FirstVersionTotal)

	firstVersionSystemCount, _ := models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"job_instance_id": Skill.JobInstanceId, "act_source": "system_service"})
	mongoSummary.FirstVersionRegexCount, _ = models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"job_instance_id": Skill.JobInstanceId, "algo": "regex"})
	mongoSummary.FirstVersionAlgoCount = firstVersionSystemCount - mongoSummary.FirstVersionRegexCount
	mongoSummary.FirstVersionRegexPercent = float32(mongoSummary.FirstVersionRegexCount) / float32(firstVersionSystemCount)
	mongoSummary.FirstVersionAlgoPercent = float32(mongoSummary.FirstVersionAlgoCount) / float32(firstVersionSystemCount)

	summary += fmt.Sprintf("用例版本:%.2f,用例总数:%d,错误数:%d,意图正确率:%f,槽位正确率:%f,意图支撑中算法占比%f,工程模板占比%f\n",
		mongoSummary.FirstVersion,
		mongoSummary.FirstVersionTotal,
		mongoSummary.FirstVersionIntentFail,
		mongoSummary.FirstVersionIntentAccuracy,
		mongoSummary.FirstVersionParaminfoAccuracy,
		mongoSummary.FirstVersionAlgoPercent,
		mongoSummary.FirstVersionRegexPercent)

	// 第二版本统计
	if mongoSummary.SecondVersion > 0 {
		mongoSummary.SecondVersionTotal, _ = models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"case_version": bson.M{"$lte": mongoSummary.SecondVersion}, "job_instance_id": mongoSummary.JobInstanceId})
		mongoSummary.SecondVersionIntentPass, _ = models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"case_version": bson.M{"$lte": mongoSummary.SecondVersion}, "job_instance_id": mongoSummary.JobInstanceId, "is_pass": true})
		mongoSummary.SecondVersionIntentFail = mongoSummary.SecondVersionTotal - mongoSummary.SecondVersionIntentPass
		mongoSummary.SecondVersionIntentAccuracy = float32(mongoSummary.SecondVersionIntentPass) / float32(mongoSummary.SecondVersionTotal)
		mongoSummary.SecondVersionParaminfoPass, _ = models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"case_version": bson.M{"$lte": mongoSummary.SecondVersion}, "job_instance_id": mongoSummary.JobInstanceId, "param_info_is_pass": true})
		mongoSummary.SecondVersionParaminfoFail = mongoSummary.SecondVersionTotal - mongoSummary.SecondVersionIntentPass
		mongoSummary.SecondVersionParaminfoAccuracy = float32(mongoSummary.SecondVersionParaminfoPass) / float32(mongoSummary.SecondVersionTotal)

		secondVersionSystemCount, _ := models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"case_version": bson.M{"$lte": mongoSummary.SecondVersion}, "job_instance_id": mongoSummary.JobInstanceId, "act_source": "system_service"})
		mongoSummary.SecondVersionRegexCount, _ = models.ReporterDB.MongoCount(SkillResultsTable, bson.M{"case_version": bson.M{"$lte": mongoSummary.SecondVersion}, "job_instance_id": mongoSummary.JobInstanceId, "algo": "regex"})
		mongoSummary.SecondVersionAlgoCount = secondVersionSystemCount - mongoSummary.SecondVersionRegexCount
		mongoSummary.SecondVersionRegexPercent = float32(mongoSummary.SecondVersionRegexCount) / float32(secondVersionSystemCount)
		mongoSummary.SecondVersionAlgoPercent = float32(mongoSummary.SecondVersionAlgoCount) / float32(secondVersionSystemCount)

		summary += fmt.Sprintf("用例版本:%.2f,用例总数:%d,错误数:%d,意图正确率:%f,槽位正确率:%f,意图支撑中算法占比%f,工程模板占比%f\n",
			mongoSummary.SecondVersion,
			mongoSummary.SecondVersionTotal,
			mongoSummary.SecondVersionIntentFail,
			mongoSummary.SecondVersionIntentAccuracy,
			mongoSummary.SecondVersionParaminfoAccuracy,
			mongoSummary.SecondVersionAlgoPercent,
			mongoSummary.SecondVersionRegexPercent)
	}

	// 附加信息
	summary += fmt.Sprintf("请求参数:请求地址:%s,AgentId:%d,并发数:%d\n", Skill.SkillConfig.ConnAddr, Skill.SkillConfig.AgentId, Skill.SkillConfig.ChanNum)
	Skill.Summary = summary

	// 测试总结存储到mongo
	mongoSummary.OverView = Skill.Summary
	models.ReporterDB.MongoInsertOne(MongoSummaryTable, mongoSummary)
}

func (Skill *SkillTask) sendReport() {
	if Skill.SkillConfig.IsReport != "yes" || Skill.SkillConfig.ReportString == nil {
		return
	}
	text := Skill.Summary
	if Skill.SummaryFile != "" {
		text += excelDownloadRouter + Skill.SummaryFile
	}
	payload := &FeiShu{Payload: &FeiShuPayload{MsgType: "text", Content: &simpleText{Text: text}}}

	for _, url := range Skill.SkillConfig.ReportString {
		payload.Url = url.Address
		err := reportToFeiShu(payload)
		if err != nil {
			logf.Error(err)
		}
	}
}

var (
	MongoSummaryTable = "summary"
)
