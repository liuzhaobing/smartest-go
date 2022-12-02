package task

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"smartest-go/models"
	"smartest-go/pkg/mongo"
	"strconv"
	"strings"
	"sync"
	"time"
)

type jsonTime struct {
	time.Duration
}

type Space struct {
	SpaceName string `json:"space_name"`
}

type KGPayload struct {
	Spaces   []*Space `json:"spaces"`
	Question string   `json:"question"`
}

// 知识图谱会话接口
func (e *EnvInfo) mChat(kgPayload *KGPayload) *chatResponse {
	payload, _ := json.Marshal(kgPayload)
	r := HTTPReqInfo{
		Method:  "POST",
		Payload: bytes.NewReader(payload),
	}

	if e.BackendUrl != "" {
		r.Url = e.BackendUrl + "/kgqa/v1/chat"
	} else {
		r.Url = e.FrontUrl + "/graph/kgqa/v1/chat"
	}
	var c chatResponse
	err := json.Unmarshal(e.mRequest(r), &c)
	if err != nil {
		return nil
	}
	return &c
}

// 知识图谱会话响应结构体
type chatResponse struct {
	Code int `json:"code"`
	Data struct {
		Type       string `json:"@type"`
		EntityName string `json:"entity_name"`
		Disambi    string `json:"disambi"`
		Answer     string `json:"answer"`
		Attr       struct {
			Describ string `json:"describ"`
		} `json:"attr"`
		Source  string `json:"source"`
		TraceId string `json:"trace_id"`
	} `json:"data"`
}

type KGTaskConfig struct {
	TaskName      string          `json:"task_name" form:"task_name"`
	JobInstanceId string          `json:"job_instance_id" form:"job_instance_id"`
	ChanNum       int             `json:"chan_num" form:"chan_num"`
	IsReport      string          `json:"is_report" form:"is_report"`
	ReportString  []*ReportString `json:"report_string" form:"report_string"`
	EnvInfo       *EnvInfo        `json:"env_info" form:"env_info"`
	Spaces        []*Space        `json:"spaces" form:"spaces"` // [{"space_name":"common_kg_v4"}]
}

type KGDataSource struct {
	CaseNum      int64            `json:"case_num,omitempty" form:"case_num,omitempty"`           // 用例总数
	CType        int64            `json:"c_type,omitempty" form:"c_type,omitempty"`               // 1单跳 2两跳
	IsContinue   string           `json:"is_continue,omitempty" form:"is_continue,omitempty"`     // 是否持续测试
	IsRandom     string           `json:"is_random,omitempty" form:"is_random,omitempty"`         // 是否随机测试
	KGDataBase   *mongo.MongoInfo `json:"kg_data_base,omitempty" form:"kg_data_base,omitempty"`   // 用例构造使用的数据源
	TemplateJson []*KGTemplate    `json:"template_json,omitempty" form:"template_json,omitempty"` // 用例构造使用的模板
}

type KGTaskReq struct {
	Id           int64  `json:"id"`            // 用例编号
	Query        string `json:"query"`         // 测试语句
	ExpectAnswer string `json:"expect_answer"` // 期望答案
}

type KGTaskRes struct {
	ActAnswer   string // 实际答案
	ActJson     string // 返回json
	Source      string // 命中类型
	ExecuteTime int64  // 用例执行时间点
	TraceId     string // trace
}

type KGTaskOnceResp struct {
	Req     *KGTaskReq // 单次测试请求信息
	Res     *KGTaskRes // 单次测试响应信息
	IsPass  bool       // 单次测试测试结果
	EdgCost jsonTime
}

// KGResults 存储kg测试结果的MongoDB表结构
type KGResults struct {
	JobInstanceId string `bson:"job_instance_id"`
	Id            int64  `bson:"id"`
	Question      string `bson:"question"`
	Answer        string `bson:"answer"`
	ActAnswer     string `bson:"act_answer"`
	IsPass        bool   `bson:"is_pass"`
	RespJson      string `bson:"resp_json"`
	EdgCost       int64  `bson:"edg_cost"`
	ExecuteTime   int64  `bson:"execute_time"`
	TaskName      string `bson:"task_name"`
	Source        string `bson:"source"`
	TraceId       string `bson:"trace_id"`
}

type KGTask struct {
	BaseTask
	KGConfig           *KGTaskConfig        // 环境配置信息
	KGDataSourceConfig *KGDataSource        // 数据源配置信息
	chanNum            int                  // 实际并发数
	req                []*KGTaskReq         // 用例集
	Results            []*KGTaskRes         // 结果集
	RespChan           chan *KGTaskOnceResp //结果管道
	RightCount         int
	WrongCount         int
	JobInstanceId      string
	Summary            string
	SummaryFile        string
	KGCaseGetterMongo  *mongo.MongoInfo
	startTime          time.Time
	endTime            time.Time
	cost               time.Duration
	mu                 sync.Mutex
}

type KGTaskTest struct {
	*KGTask
}

func NewKGTask(kg *KGTaskConfig, req []*KGTaskReq, kgDataSourceConfig *KGDataSource) *KGTask {
	return &KGTask{
		KGConfig:           kg,
		req:                req,
		KGDataSourceConfig: kgDataSourceConfig,
		Results:            make([]*KGTaskRes, 0),
	}
}

func (KG *KGTask) pre() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	success, value := PrepareMissionFlag(KG.KGConfig.TaskName, cancel)
	if !success {
		return
	}
	if KG.KGConfig.JobInstanceId == "" {
		KG.JobInstanceId = uuid.New().String()
	}
	value.TaskType = KnowledgeGraph
	value.JobInstanceId = KG.JobInstanceId

	// 需要先把用例放在 KG.req 如果非前端传递用例 则调用方法收集用例
	if len(KG.req) == 0 || KG.KGDataSourceConfig != nil {
		KG.CaseGetterKG(ctx)
	}
	KG.RespChan = make(chan *KGTaskOnceResp, len(KG.req))
	KG.chanNum = KG.KGConfig.ChanNum
	KG.starTime = time.Now()
}

func (KG *KGTask) run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if success, _ := RunMissionFlag(KG.KGConfig.TaskName); !success {
		return
	}

	taskInfoMap[KG.KGConfig.TaskName].Cancel = cancel
	KGChan := make(chan *KGTaskReq)

	for i := 0; i < KG.chanNum; i++ {
		go func(ctx context.Context, i int, v chan *KGTaskReq) {
			defer KG.chanClose()

			for req := range v {
				select {
				case <-ctx.Done():
					close(KGChan)
					return
				default:
					res := KG.call(req)
					KG.RespChan <- res
				}
			}
		}(ctx, i, KGChan)
	}

	caseTotal := len(KG.req)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for i, req := range KG.req {
			select {
			case <-KGChan:
				wg.Done()
				return
			default:
				KGChan <- req
				if value, ok := taskInfoMap[KG.KGConfig.TaskName]; ok {
					value.ProgressPercent = (i + 1) * 100 / caseTotal
					value.Progress = fmt.Sprintf(`%d/%d`, i+1, caseTotal)
					value.Accuracy = float32(KG.RightCount) / float32(KG.RightCount+KG.WrongCount)
				}
			}
		}
		wg.Done()
		close(KGChan)
	}()
	wg.Wait()

	var KGResultList []interface{}
	for resp := range KG.RespChan {
		KGResultList = append(KGResultList, &KGResults{
			JobInstanceId: KG.JobInstanceId,
			Id:            resp.Req.Id,
			Question:      resp.Req.Query,
			Answer:        resp.Req.ExpectAnswer,
			ActAnswer:     resp.Res.ActAnswer,
			IsPass:        resp.IsPass,
			RespJson:      resp.Res.ActJson,
			EdgCost:       resp.EdgCost.Milliseconds(),
			ExecuteTime:   resp.Res.ExecuteTime,
			TaskName:      KG.KGConfig.TaskName,
			Source:        resp.Res.Source,
			TraceId:       resp.Res.TraceId,
		})
	}
	models.ReporterDB.MongoInsertMany(kgResultsTable, KGResultList)
}

func (KG *KGTask) end() {
	KG.endTime = time.Now()
	KG.getResultSummary()
	KG.writeKGResultExcel()
	KG.sendReport()
	if success, _ := EndMissionFlag(KG.KGConfig.TaskName, KG.Summary, KG.SummaryFile); !success {
		return
	}
}

func (KG *KGTask) stop() {
	KG.endTime = time.Now()
}

func (KG *KGTask) chanClose() {
	KG.mu.Lock()
	defer KG.mu.Unlock()
	if KG.chanNum <= 1 {
		close(KG.RespChan)
	} else {
		KG.chanNum--
	}
}

// 单条用例执行详情
func (KG *KGTask) call(req *KGTaskReq) *KGTaskOnceResp {
	executeTime, _ := strconv.ParseInt(time.Now().Format("20060102150405"), 10, 64)
	Res := &KGTaskOnceResp{
		Req: req,
		Res: &KGTaskRes{
			ActAnswer:   "",
			ActJson:     "",
			ExecuteTime: executeTime,
		},
	}
	// do test
	startReq := time.Now()

	res := KG.KGConfig.EnvInfo.mChat(&KGPayload{Spaces: KG.KGConfig.Spaces, Question: req.Query})
	Res.EdgCost.Duration = time.Now().Sub(startReq)

	jsonByte, _ := json.Marshal(res)
	Res.Res.ActJson = string(jsonByte)

	if res != nil {
		Res.Res.ActAnswer = res.Data.Answer
		Res.Res.Source = res.Data.Source
		Res.Res.TraceId = res.Data.TraceId

		// do assertion
		if strings.Contains(Res.Res.ActAnswer, Res.Req.ExpectAnswer) {
			Res.IsPass = true
			KG.RightCount++
		} else {
			Res.IsPass = false
			KG.WrongCount++
		}
	}
	fmt.Println(Res.Res.ActAnswer)
	return Res
}

var (
	_              TaskModel = &KGTask{}
	kgResultsTable           = "kg_results"
)
