package task

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/tidwall/gjson"
	"google.golang.org/grpc"
	"io/ioutil"
	"net/http"
	"smartest-go/models"
	"smartest-go/proto/talk"
	"strconv"
	"strings"
	"sync"
	"time"
)

type SkillTaskConfig struct {
	TaskName      string          `json:"task_name" form:"task_name"`
	JobInstanceId string          `json:"job_instance_id" form:"job_instance_id"`
	IsReport      string          `json:"is_report" form:"is_report"`
	ReportString  []*ReportString `json:"report_string" form:"report_string"`
	ChanNum       int             `json:"chan_num"  form:"chan_num"`
	ConnAddr      string          `json:"backend_url"  form:"backend_url"`
	IsTestModel   string          `json:"is_test"  form:"is_test"` // 开启测试模式后 部分技能tts答复无法查看
	IsDebug       string          `json:"is_nlu"  form:"is_nlu"`   // Debug模式仅用于bug review
	NLUAddr       string          `json:"nlu_url"  form:"nlu_url"` // Debug模式必填nlu访问地址
	AgentId       int64           `json:"agent_id"  form:"agent_id"`
	RobotID       string          `json:"robot_id"  form:"robot_id"`
	TenantCode    string          `json:"tenant_code"  form:"tenant_code"`
	Version       string          `json:"version"  form:"version"`
}

type SkillDataSource struct {
	DBFilter string `json:"filter,omitempty"  form:"filter,omitempty"`
}

type SkillTaskReq struct { // 从用例中抽取的数据
	Id              int64   `json:"id,omitempty" form:"id,omitempty"` //用例编号
	Query           string  `json:"question,omitempty" form:"question,omitempty"`
	ExpectSource    string  `json:"source,omitempty" form:"source,omitempty"`
	ExpectDomain    string  `json:"domain,omitempty" form:"domain,omitempty"`
	ExpectIntent    string  `json:"intent,omitempty" form:"intent,omitempty"`
	SkillSource     string  `json:"skill_source,omitempty" form:"skill_source,omitempty"`
	SkillCn         string  `json:"skill_cn,omitempty" form:"skill_cn,omitempty"`
	RobotType       string  `json:"robot_type,omitempty" form:"robot_type,omitempty"`
	RobotID         string  `json:"robot_id,omitempty" form:"robot_id,omitempty"`
	ExpectParamInfo string  `json:"paraminfo,omitempty" form:"paraminfo,omitempty"`
	UseTest         int     `json:"usetest,omitempty" form:"usetest,omitempty"`
	IsSmoke         int     `json:"is_smoke,omitempty" form:"is_smoke,omitempty"`
	CaseVersion     float32 `json:"case_version,omitempty" form:"case_version,omitempty"`
	EditLogs        string  `json:"edit_logs,omitempty" form:"edit_logs,omitempty"`
}

type SkillTaskRes struct { // 从响应体收集到的数据
	ActSource       string
	ActDomain       string
	ActIntentHitLog string // hitlog 下的意图
	ActIntentTTS    string // tts 下的意图
	ActParamInfo    string
	ExecuteTime     int64
	Algo            string
	AnswerString    string
	MusicUrl        string
	PicUrl          string
	VideoUrl        string
	AlgoScore       float64
	ActParam        string // tts.action.params下的意图收集
	ActInputContext string
	NLUDebugInfo    string
	TraceId         string
}

type SkillTaskOnceResp struct {
	Req             *SkillTaskReq // 单次测试请求信息
	Res             *SkillTaskRes // 单次测试响应信息
	IsIntentPass    bool
	IsParamInfoPass bool
	FailReason      string
	Developer       string
	BugStatus       string
	EdgCost         jsonTime
}

// SkillResults 存储Skill测试结果的MongoDB表结构
type SkillResults struct {
	JobName         string  `json:"task_name,omitempty"      bson:"task_name,omitempty"`             // JobName+JobId 唯一标识一个测试任务
	JobInstanceId   string  `json:"job_instance_id,omitempty"      bson:"job_instance_id,omitempty"` // JobInstanceId+ExecuteTime 唯一标识一个测试执行历史
	ExecuteTime     int64   `json:"execute_time,omitempty"      bson:"execute_time,omitempty"`       // JobInstanceId+ExecuteTime 唯一标识一个测试执行历史
	CaseNumber      int64   `json:"id,omitempty"      bson:"id,omitempty"`
	Question        string  `json:"question,omitempty"      bson:"question,omitempty"`
	Source          string  `json:"source,omitempty"      bson:"source,omitempty"`
	ActSource       string  `json:"act_source,omitempty"      bson:"act_source,omitempty"`
	Domain          string  `json:"domain,omitempty"      bson:"domain,omitempty"`
	ActDomain       string  `json:"act_domain,omitempty"      bson:"act_domain,omitempty"`
	Intent          string  `json:"intent,omitempty"      bson:"intent,omitempty"`
	ActIntent       string  `json:"act_intent,omitempty"      bson:"act_intent,omitempty"`
	IsPass          bool    `json:"is_pass,omitempty"      bson:"is_pass,omitempty"`
	ActIntentTTS    string  `json:"act_intent_tts,omitempty"      bson:"act_intent_tts,omitempty"`
	IsSmoke         int     `json:"is_smoke,omitempty"      bson:"is_smoke,omitempty"`
	Parameters      string  `json:"parameters,omitempty"      bson:"parameters,omitempty"`
	Cost            int64   `json:"edg_cost,omitempty"      bson:"edg_cost,omitempty"`
	ParamInfo       string  `json:"paraminfo,omitempty"      bson:"paraminfo,omitempty"`
	ActParamInfo    string  `json:"act_param_info,omitempty"      bson:"act_param_info,omitempty"`
	ParamInfoIsPass bool    `json:"param_info_is_pass,omitempty"      bson:"param_info_is_pass,omitempty"`
	AnswerString    string  `json:"answer_string,omitempty"      bson:"answer_string,omitempty"`
	AnswerUrl       string  `json:"answer_url,omitempty"      bson:"answer_url,omitempty"`
	CaseVersion     float32 `json:"case_version,omitempty"      bson:"case_version,omitempty"`
	Algo            string  `json:"algo,omitempty"      bson:"algo,omitempty"`
	AlgoScore       float64 `json:"algo_score,omitempty"      bson:"algo_score,omitempty"`
	ActInputContext string  `json:"act_input_context,omitempty"      bson:"act_input_context,omitempty"`
	RobotId         string  `json:"robot_id,omitempty"      bson:"robot_id,omitempty"`
	TraceId         string  `json:"trace_id,omitempty"      bson:"trace_id,omitempty"`
	ActRobotType    string  `json:"act_robot_type,omitempty"      bson:"act_robot_type,omitempty"`
	NLUDebugInfo    string  `json:"nlu_debug_info,omitempty"      bson:"nlu_debug_info,omitempty"`
	FailReason      string  `json:"fail_reason,omitempty"      bson:"fail_reason,omitempty"`
	FilterDeveloper string  `json:"filter_developer,omitempty"      bson:"filter_developer,omitempty"`
	AssignReason    string  `json:"assign_reason,omitempty"      bson:"assign_reason,omitempty"`
	FixDeveloper    string  `json:"fix_developer,omitempty"      bson:"fix_developer,omitempty"`
	BugStatus       string  `json:"bug_status,omitempty"      bson:"bug_status,omitempty"`
}

type SkillTask struct {
	BaseTask
	SkillConfig           *SkillTaskConfig
	SkillDataSourceConfig *SkillDataSource
	chanNum               int
	req                   []*SkillTaskReq
	Results               []*SkillTaskRes
	RespChan              chan *SkillTaskOnceResp
	RightCount            int
	WrongCount            int
	startTime             time.Time
	endTime               time.Time
	cost                  time.Duration
	mu                    sync.Mutex
}

type SkillTaskTest struct {
	*SkillTask
}

func NewSkillTask(Skill *SkillTaskConfig, req []*SkillTaskReq, SkillDataSourceConfig *SkillDataSource) *SkillTask {
	return &SkillTask{
		SkillConfig:           Skill,
		SkillDataSourceConfig: SkillDataSourceConfig,
		req:                   req,
		Results:               make([]*SkillTaskRes, 0),
	}
}

func SplitSkillCases(s []*SkillTaskReq) (f []*SkillTaskReq) {
	for _, req := range s {
		QuestionList := strings.Split(req.Query, "&&")
		RobotTypeList := strings.Split(req.RobotType, "&&")
		for _, singleQ := range QuestionList {
			for _, singleR := range RobotTypeList {
				////修改请求信息 改成单个问题单个机型去请求
				singleRequest := Copy(req).(*SkillTaskReq)
				singleRequest.Query = singleQ
				singleRequest.RobotType = singleR
				f = append(f, singleRequest)
			}
		}
	}
	return f
}

func (Skill *SkillTask) pre() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	success, value := PrepareMissionFlag(Skill.SkillConfig.TaskName, cancel)
	if !success {
		return
	}
	if Skill.SkillConfig.JobInstanceId == "" {
		Skill.SkillConfig.JobInstanceId = uuid.New().String()
	}
	value.TaskType = SystemSkill
	value.JobInstanceId = Skill.SkillConfig.JobInstanceId

	// 从数据库中获取用例
	if len(Skill.req) == 0 || Skill.SkillDataSourceConfig != nil {
		Skill.CaseGetterSkill(ctx)
	}

	Skill.req = SplitSkillCases(Skill.req) // 先处理下带有&&的query
	Skill.RespChan = make(chan *SkillTaskOnceResp, len(Skill.req))
	Skill.chanNum = Skill.SkillConfig.ChanNum
	Skill.startTime = time.Now()
}

func (Skill *SkillTask) run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if success, _ := RunMissionFlag(Skill.SkillConfig.TaskName); !success {
		return
	}

	taskInfoMap[Skill.SkillConfig.TaskName].Cancel = cancel
	SkillChan := make(chan *SkillTaskReq)

	for i := 0; i < Skill.chanNum; i++ {
		go func(ctx context.Context, i int, v chan *SkillTaskReq) {
			defer Skill.chanClose()
			conn, err := grpc.Dial(Skill.SkillConfig.ConnAddr, grpc.WithInsecure())
			defer conn.Close()
			if err != nil {
				// TODO 失败标记
				return
			}
			for req := range v {
				select {
				case <-ctx.Done():
					close(SkillChan)
					return
				default:
					res := Skill.call(conn, req)
					Skill.RespChan <- res
				}
			}
		}(ctx, i, SkillChan)
	}

	caseTotal := len(Skill.req)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for i, req := range Skill.req {
			select {
			case <-SkillChan:
				wg.Done()
				return
			default:
				SkillChan <- req
				if value, ok := taskInfoMap[Skill.SkillConfig.TaskName]; ok {
					value.ProgressPercent = (i + 1) * 100 / caseTotal
					value.Progress = fmt.Sprintf(`%d/%d`, i+1, caseTotal)
					value.Accuracy = float32(Skill.RightCount) / float32(Skill.RightCount+Skill.WrongCount)
				}
			}
		}
		wg.Done()
		close(SkillChan)
	}()
	wg.Wait()

	var SkillResultList []interface{}
	for resp := range Skill.RespChan {
		SkillResultList = append(SkillResultList, &SkillResults{
			JobInstanceId:   Skill.SkillConfig.JobInstanceId,
			JobName:         Skill.SkillConfig.TaskName,
			ExecuteTime:     resp.Res.ExecuteTime,
			CaseNumber:      resp.Req.Id,
			Question:        resp.Req.Query,
			Source:          resp.Req.ExpectSource,
			ActSource:       resp.Res.ActSource,
			Domain:          resp.Req.ExpectDomain,
			ActDomain:       resp.Res.ActDomain,
			Intent:          resp.Req.ExpectIntent,
			ActIntent:       resp.Res.ActIntentHitLog,
			IsPass:          resp.IsIntentPass,
			ActIntentTTS:    resp.Res.ActIntentTTS,
			IsSmoke:         resp.Req.IsSmoke,
			Cost:            resp.EdgCost.Milliseconds(),
			ParamInfo:       resp.Req.ExpectParamInfo,
			ActParamInfo:    resp.Res.ActParamInfo,
			ParamInfoIsPass: resp.IsParamInfoPass,
			AnswerString:    resp.Res.AnswerString,
			AnswerUrl:       resp.Res.MusicUrl + " " + resp.Res.PicUrl + " " + resp.Res.VideoUrl,
			CaseVersion:     resp.Req.CaseVersion,
			Algo:            resp.Res.Algo,
			AlgoScore:       resp.Res.AlgoScore,
			ActInputContext: resp.Res.ActInputContext,
			RobotId:         resp.Req.RobotID,
			TraceId:         resp.Res.TraceId,
			ActRobotType:    resp.Req.RobotType,
			NLUDebugInfo:    resp.Res.NLUDebugInfo,
			FailReason:      resp.FailReason,
			FilterDeveloper: resp.Developer,
			AssignReason:    "",
			FixDeveloper:    "",
			BugStatus:       resp.BugStatus,
		})
	}
	models.ReporterDB.MongoInsertMany(SkillResultsTable, SkillResultList)
}

func (Skill *SkillTask) end() {
	if success, _ := EndMissionFlag(Skill.SkillConfig.TaskName); !success {
		return
	}
	Skill.endTime = time.Now()
	Skill.sendReport()
}

func (Skill *SkillTask) stop() {
	Skill.endTime = time.Now()
}

func (Skill *SkillTask) chanClose() {
	Skill.mu.Lock()
	defer Skill.mu.Unlock()
	if Skill.chanNum <= 1 {
		close(Skill.RespChan)
	} else {
		Skill.chanNum--
	}
}

func (Skill *SkillTask) call(conn *grpc.ClientConn, req *SkillTaskReq) *SkillTaskOnceResp {
	executeTime, _ := strconv.ParseInt(time.Now().Format("20060102150405"), 10, 64)
	Res := &SkillTaskOnceResp{
		Req: req,
		Res: &SkillTaskRes{
			ExecuteTime: executeTime,
			TraceId:     uuid.New().String() + "@cloudminds-test.com",
		},
		IsIntentPass:    false,
		IsParamInfoPass: false,
	}
	// do test
	c := talk.NewTalkClient(conn)
	r := &talk.TalkRequest{
		IsFull:     true,
		AgentID:    Skill.SkillConfig.AgentId,
		SessionID:  Res.Res.TraceId,
		QuestionID: Res.Res.TraceId,
		EventType:  talk.Text,
		EnvInfo:    make(map[string]string),
		RobotID:    Skill.SkillConfig.RobotID,
		TestMode:   true,
		TenantCode: Skill.SkillConfig.TenantCode,
		Version:    Skill.SkillConfig.Version,
		Asr: talk.Asr{
			Lang: "ZH",
			Text: req.Query,
		},
	}

	// 设置实际机型
	r.EnvInfo["devicetype"] = req.RobotType

	// 设置多轮
	if req.RobotID != "" {
		r.SessionID = req.RobotID
	}
	if 1 == 1 {
		if req.RobotID != "" {
			r.RobotID = req.RobotID
			r.SessionID = req.RobotID
		} else {
			r.RobotID = Res.Res.TraceId
		}
	}

	// 设置测试模式
	if Skill.SkillConfig.IsTestModel == "no" {
		r.TestMode = false
	}

	startReq := time.Now()
	resp, err := c.Talk(context.Background(), r)
	Res.EdgCost.Duration = time.Now().Sub(startReq)

	if err != nil {
		// TODO 失败处理
		return Res
	}

	Skill.getSkillResponseData(resp, Res.Res) // 从响应体收集信息
	Skill.assertSkillIntent(Res)              // 断言意图
	Skill.assertSkillParamInfo(Res)           // 断言槽位
	if !Res.IsParamInfoPass {
		Res.FailReason = "槽位命中错误"
	}
	Skill.debugModelCheck(Res) // Debug模式检测算法命中
	Skill.tagToDeveloper(Res)  // BUG分配

	if Res.IsIntentPass {
		Skill.RightCount++
	} else {
		Skill.WrongCount++
	}
	return Res
}

// SkillParamInfo 槽位信息
type SkillParamInfo struct {
	BeforeValue string `json:"BeforeValue"`
	EntityType  string `json:"EntityType"`
	Name        string `json:"Name"`
	Value       string `json:"Value"`
}

func (Skill *SkillTask) getSkillResponseData(resp *talk.TalkResponse, Res *SkillTaskRes) { // 从响应体中抽取所需要的所有内容
	Res.ActSource = resp.Source
	Res.ActDomain = resp.HitLog.Fields["domain"].GetStringValue()
	Res.ActIntentHitLog = resp.HitLog.Fields["intent"].GetStringValue()
	if resp.Tts != nil {
		Res.ActIntentTTS = resp.Tts[0].Action.Param.Intent
		if resp.Tts[0].Action.Param.Params != nil {
			data, _ := json.Marshal(resp.Tts[0].Action.Param.Params)
			Res.ActParam = string(data)
		}
	}
	if resp.HitLog.Fields["paraminfo"] != nil {
		paramList := resp.HitLog.Fields["paraminfo"].GetListValue().Values
		actParamList := make([]*SkillParamInfo, 0)
		for _, p := range paramList {
			actParamList = append(actParamList, &SkillParamInfo{
				BeforeValue: p.GetStructValue().GetFields()["BeforeValue"].GetStringValue(),
				EntityType:  p.GetStructValue().GetFields()["EntityType"].GetStringValue(),
				Name:        p.GetStructValue().GetFields()["Name"].GetStringValue(),
				Value:       p.GetStructValue().GetFields()["Value"].GetStringValue(),
			})
		}
		p, _ := json.Marshal(actParamList)
		Res.ActParamInfo = string(p)
	}
	Res.Algo = resp.HitLog.Fields["algo"].GetStringValue()
	for _, tt := range resp.Tts {
		if tt.Text != "" {
			Res.AnswerString = tt.Text
		}
		if tt.Action.Param.Url != "" {
			Res.MusicUrl = tt.Action.Param.Url
		}
		if tt.Action.Param.PicUrl != "" {
			Res.PicUrl = tt.Action.Param.PicUrl
		}
		if tt.Action.Param.VideoUrl != "" {
			Res.VideoUrl = tt.Action.Param.VideoUrl
		}
	}
	if resp.HitLog.Fields["qaresult"] != nil {
		qaResult := resp.HitLog.Fields["qaresult"].GetStructValue().Fields
		if qaResult["answer"] != nil {
			if qaResult["answer"].GetStructValue().Fields["score"] != nil {
				Res.AlgoScore = qaResult["answer"].GetStructValue().Fields["score"].GetNumberValue() // 以浮点类型形式传递值
			}
		}
	}
	if resp.HitLog.Fields["beforeContext"] != nil {
		if resp.HitLog.Fields["beforeContext"].GetStructValue().Fields["SystemContext"] != nil {
			Res.ActInputContext = resp.HitLog.Fields["beforeContext"].GetStructValue().Fields["SystemContext"].GetStringValue()
		}
	}
}

func (Skill *SkillTask) assertSkillIntent(Res *SkillTaskOnceResp) {
	// 期望命中非系统技能&&用户技能时，只要命中的source与预期的一致即可断言通过
	if Res.Req.ExpectSource != "system_service" &&
		Res.Req.ExpectSource != "user_service" &&
		Res.Req.ExpectSource == Res.Res.ActSource {
		Res.IsIntentPass = true
		return
	}

	// 意图全对
	if Res.Req.ExpectSource == Res.Res.ActSource &&
		Res.Req.ExpectDomain == Res.Res.ActDomain &&
		Res.Req.ExpectIntent == Res.Res.ActIntentHitLog {
		Res.IsIntentPass = true
		return
	}

	// 意图中命中CanSing和SongRandomly
	if strings.Contains("CanSing&&SongRandomly&&SingAnthorSong", Res.Req.ExpectIntent) &&
		strings.Contains("CanSing&&SongRandomly&&SingAnthorSong", Res.Res.ActIntentHitLog) &&
		Res.Res.ActIntentHitLog != "" {
		Res.IsIntentPass = true
		return
	}

	// 记录错误原因
	if Res.Req.ExpectSource != Res.Res.ActSource {
		Res.FailReason = "domain未命中"
		Res.IsIntentPass = false
		return
	}
	if Res.Req.ExpectIntent != Res.Res.ActIntentHitLog {
		Res.FailReason = "intent未命中"
		Res.IsIntentPass = false
		return
	}
}

func (Skill *SkillTask) assertSkillParamInfo(Res *SkillTaskOnceResp) {
	if Res.Req.ExpectParamInfo == "" && Res.Res.ActParamInfo == "" {
		Res.IsParamInfoPass = true // 期望槽位与实际槽位同时为空
		return
	}
	if Res.Res.ActParamInfo == "" || Res.Req.ExpectParamInfo == "" {
		Res.IsParamInfoPass = false // 期望槽位与实际槽位至少有一个为空
		return
	}

	expParam := make([]*SkillParamInfo, 0)
	err := json.Unmarshal([]byte(Res.Req.ExpectParamInfo), &expParam)
	if err != nil {
		Res.IsParamInfoPass = false // 期望槽位转struct失败 期望本身问题
		return
	}
	actParam := make([]*SkillParamInfo, 0)
	err = json.Unmarshal([]byte(Res.Res.ActParamInfo), &actParam)
	if err != nil {
		Res.IsParamInfoPass = false
		return
	}

	// around 技能特殊处理
	if Res.Req.ExpectDomain == "around" {
		if len(expParam) != len(actParam) {
			Res.IsParamInfoPass = false // 期望槽位与实际槽位的数量不一致
			return
		}
		for _, act := range actParam {
			if act.Value == "" {
				Res.IsParamInfoPass = false // 实际槽位后值为空
				return
			}
		}
		Res.IsParamInfoPass = true
		return
	}

	// 判断所有预期结果是否都命中
	for _, exp := range expParam {
		if !strings.Contains(Res.Res.ActParamInfo, exp.Name) ||
			!strings.Contains(Res.Res.ActParamInfo, exp.BeforeValue) {
			Res.IsParamInfoPass = false // BeforeValue 和 Name 均需要一致
			return
		}
	}
	// 判断所有实际结果是否都命中
	for _, act := range actParam {
		if act.Value == "" ||
			!strings.Contains(Res.Req.ExpectParamInfo, act.Name) ||
			!strings.Contains(Res.Req.ExpectParamInfo, act.BeforeValue) {
			Res.IsParamInfoPass = false // BeforeValue 和 Name 均需要一致
			return
		}
	}
	Res.IsParamInfoPass = true
	return
}

var (
	JacksonZhang = "@Jackson Zhang 张发展"
	MikeLuo      = "@Mike Luo 罗镇权"
	KevinRen     = "@Kevin Ren 任珂"
	AaronYang    = "@Aaron Yang 杨武"
	YoungZhao    = "@Young Zhao 赵杨"
	DavidLi      = "@David Li 李超凡"
	ZipperZhao   = "@Zipper Zhao 赵鹏"
	SevelLiu     = "@Sevel Liu 刘兆兵"
	LeonZhou     = "@Leon Zhou 周磊"

	JacksonEntity = "activity businessbrand person_name company robot_name crosstalk_performer crosstalk_title date drama_performer drama_title drama_type fcurrency food holiday joke_type location news orientation timescope person_virtual poem_content poem_title poem_type poem_writer product singer song_title solarterm song_type story_title story_type storytelling_performer storytelling_title vehicle year year_number tcurrency height weight iq_number age"
)

var developer = map[string]string{
	/*
		Step 1：[赵杨]排查工程模板问题、多轮问题        (algo==regex || robot_id!=nil)
		Step 2：[杨武]排查机型问题                   (robot_type!=nil)
		Step 3：[任珂]排查compute、system_dialet问题 (domain==compute || domain==system_dialet)
		Step 4：[张发展]排查domain未命中问题           (source!=system_service && source!=user_service)
					domainname==other :domain模型识别错误
					domainname!=other
						intentname==other :intent未命中
							is_param_pass==true           :intent模型未识别导致domain未识别
							is_param_pass==false          :槽位未识别导致intent模型未识别导致domain未识别
							domain==around and param==空  :around槽位未识别导致domain未识别
		Step 5: 排查intent未命中问题                 (is_pass==false)
					intentname!=intent
						is_param_pass==true              :intent模型未命中
							[张发展] domain in (music poetry crosstalk story storytelling)
						is_param_pass==false             :intent槽位未命中
							[任珂] domain in (weather_new stock compute dance_action joke system)   :实体树问题
							[张发展] domain in (times meta currency )                               :NER实体问题
							[罗镇权] domain in (robot_character)                                    :期望无实体但槽位识别错误
		Step 6: [罗镇权]排查intent未命中问题          (intent in "system_urgent system robot_character meta currency weather_new stock")
		Step 7: [赵鹏/张发展]排查paraminfo未命中问题   (actual param is @sys or @ner)
		Step 8: [TimeOut]排查超时问题                (Cost > 600 ms)
		Step 9：[All]排查其他domain
				[付霞]   (music)
				[罗镇权] (dance around joke)
				[任珂]   (times poetry)
				[杨武]   (twister search)
				[赵杨]   (indoornavigation multimedia)
				[李超凡] (system_urgent system robot_character meta storytelling currency ha feedback)
				[赵鹏]   (weather_new stock crosstalk translate drama constellation repeat)
	*/
	MikeLuo:    "dance around joke",
	KevinRen:   "times poetry",
	AaronYang:  "twister search",
	YoungZhao:  "indoornavigation multimedia",
	DavidLi:    "system_urgent system robot_character meta storytelling currency ha feedback",
	ZipperZhao: "weather_new stock crosstalk translate drama constellation repeat",
}

func (Skill *SkillTask) tagToDeveloper(Res *SkillTaskOnceResp) {
	if Res.IsParamInfoPass && Res.IsIntentPass {
		return
	}

	Res.BugStatus = "New"

	if Res.Res.Algo == "regex" || Res.Req.RobotID != "" {
		Res.Developer = YoungZhao // 排查模板匹配和多轮问题
		return
	}
	if Res.Req.RobotType != "" {
		Res.Developer = AaronYang // 排查机型问题
		return
	}
	if Res.Req.ExpectDomain == "compute" ||
		Res.Req.ExpectDomain == "system_dialet" {
		Res.Developer = KevinRen // 排查compute、system_dialet问题
		return
	}
	if Res.FailReason == "domain未命中" &&
		Res.Res.NLUDebugInfo != "" {
		if Res.EdgCost.Milliseconds() > 600 {
			Res.FailReason = "TimeOut导致Domain未命中"
			Res.Developer = SevelLiu
			return
		}

		if gjson.Get(Res.Res.NLUDebugInfo, "domainname").String() == "other" {
			Res.FailReason = "domain模型识别错误"
			Res.Developer = LeonZhou // 排查domain未命中
			return
		}
		if gjson.Get(Res.Res.NLUDebugInfo, "intentname").String() == "other" {
			if Res.IsParamInfoPass {
				Res.FailReason = "intent模型未识别导致domain未识别"
			} else {
				Res.FailReason = "槽位未识别导致intent模型未识别导致domain未识别"
			}
			Res.Developer = MikeLuo
			return
		}
	}
	if Res.FailReason == "intent未命中" &&
		Res.Res.NLUDebugInfo != "" {
		if gjson.Get(Res.Res.NLUDebugInfo, "intentname").String() != Res.Req.ExpectIntent {
			if Res.IsParamInfoPass {
				Res.FailReason = "intent模型未命中"
				if strings.Contains("music poetry crosstalk story storytelling", Res.Req.ExpectDomain) {
					Res.Developer = JacksonZhang
				} else {
					Res.Developer = MikeLuo
				}
				return
			}
			Res.FailReason = "intent槽位未命中"
			if strings.Contains("weather_new stock compute dance_action joke system", Res.Req.ExpectDomain) {
				Res.FailReason = "实体树问题"
				Res.Developer = KevinRen
				return
			}
			if strings.Contains("times meta currency", Res.Req.ExpectDomain) {
				Res.FailReason = "NER实体问题"
				Res.Developer = JacksonZhang
				return
			}
			if strings.Contains("robot_character", Res.Req.ExpectDomain) {
				Res.FailReason = "期望无实体但槽位识别错误"
				Res.Developer = MikeLuo
				return
			}
		}
		if strings.Contains("system_urgent system robot_character meta currency weather_new stock", Res.Req.ExpectDomain) {
			Res.Developer = MikeLuo // 排查部分技能intent未命中问题
			return
		}
	}
	if Res.FailReason == "槽位命中错误" && Res.Res.NLUDebugInfo != "" {
		if Res.Req.ExpectParamInfo != "" {
			expParam := make([]*SkillParamInfo, 0)
			err := json.Unmarshal([]byte(Res.Req.ExpectParamInfo), &expParam)
			if err != nil {
				Res.FailReason = "用例槽位期望值错误"
				Res.Developer = SevelLiu
				return
			}
			for _, exp := range expParam {
				tp := strings.Split(exp.EntityType, ".")
				if strings.Contains(JacksonEntity, tp[len(tp)-1]) {
					Res.Developer = JacksonZhang
					return
				}
			}
		}
		if Res.Res.ActParamInfo != "" {
			Res.Developer = ZipperZhao
			return
		}
	}
	for developerName, developerDomains := range developer {
		if strings.Contains(developerDomains, Res.Req.ExpectDomain) {
			Res.Developer = developerName
			return
		}
	}
	Res.Developer = SevelLiu
	return
}

func (Skill *SkillTask) debugModelCheck(Res *SkillTaskOnceResp) {
	if Skill.SkillConfig.IsDebug == "yes" && Skill.SkillConfig.NLUAddr != "" {
		Res.Res.NLUDebugInfo = NluRequest(
			Skill.SkillConfig.NLUAddr,
			Res.Req.Query,
			Res.Res.TraceId+"_nlu",
			Res.Res.ActInputContext)
	}
}

// NluRequest NLU请求
func NluRequest(httpAddr, query, traceId, contextInfo string) string {
	url := fmt.Sprintf(`http://%s/nlp-sdk/nlu/intent-recognize`, httpAddr)
	queryInfo := fmt.Sprintf(`{
	"traceid": "%s",
	"agentid": "1",
	"query": "%s",
	"context": "%s",
	"robot_name": ""
}`, traceId, query, contextInfo)
	payload := strings.NewReader(queryInfo)
	req, _ := http.NewRequest("POST", url, payload)
	req.Header.Add("Content-Type", "application/json")
	res, _ := http.DefaultClient.Do(req)
	if res != nil {
		body, _ := ioutil.ReadAll(res.Body)
		bodyJson := string(body)
		result := gjson.Get(bodyJson, "data")
		debugInfoString := result.String()
		res.Body.Close()
		return debugInfoString
	}
	return ""
}

var (
	_                 TaskModel = &SkillTask{}
	SkillResultsTable           = "skill_results"
)
