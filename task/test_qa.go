package task

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"reflect"
	"regexp"
	"smartest-go/models"
	"smartest-go/proto/talk"
	"strconv"
	"strings"
	"sync"
	"time"
)

type QATaskConfig struct {
	TaskName      string          `json:"task_name" form:"task_name"`
	JobInstanceId string          `json:"job_instance_id" form:"job_instance_id"`
	IsReport      string          `json:"is_report" form:"is_report"`
	ReportString  []*ReportString `json:"report_string" form:"report_string"`
	ChanNum       int             `json:"chan_num"  form:"chan_num"`
	ConnAddr      string          `json:"backend_url"  form:"backend_url"`
	IsGroupId     string          `json:"is_group"  form:"is_group"`
	AgentId       int64           `json:"agent_id"  form:"agent_id"`
	RobotID       string          `json:"robot_id"  form:"robot_id"`
	TenantCode    string          `json:"tenant_code"  form:"tenant_code"`
	Version       string          `json:"version"  form:"version"`
}

type QADataSource struct {
	DBFilter string `json:"filter,omitempty"  form:"filter,omitempty"`
}

type QATaskReq struct {
	Id           int64    `json:"id,omitempty" form:"id,omitempty"`                   //用例编号
	Query        string   `json:"question,omitempty" form:"question,omitempty"`       //请求的Q列表，
	ExpectAnswer []string `json:"answer_list,omitempty" form:"answer_list,omitempty"` //期望的A列表
	ExpectGroup  int64    `json:"qa_group_id,omitempty" form:"qa_group_id,omitempty"` //期望的group_id
	RobotType    string   `json:"robot_type,omitempty" form:"robot_type,omitempty"`   //请求的机器人机型
	IsSmoke      int      `json:"is_smoke,omitempty" form:"is_smoke,omitempty"`       //请求的机器人机型
}

type QATaskRes struct {
	ActAnswer   string
	Source      string // 命中类型
	ExecuteTime int64  // 用例执行时间点
	TraceId     string // trace
	GroupId     int64
	AlgoScore   float64
}

type QATaskOnceResp struct {
	Req           *QATaskReq // 单次测试请求信息
	Res           *QATaskRes // 单次测试响应信息
	IsPass        bool       // 整体是否通过
	IsTextPass    bool       // 回复tts是否通过
	IsGroupIdPass bool       // 命中的GroupId是否一致
	IsExactMatch  bool       // 是否完全匹配
	EdgCost       jsonTime
}

// QAResults 存储qa测试结果的MongoDB表结构
type QAResults struct {
	JobInstanceId string `bson:"job_instance_id"`  //
	Id            int64  `bson:"id"`               // 用例编号
	Question      string `bson:"question"`         // 测试语句
	Answer        string `bson:"exp_answer"`       // 期望回复
	ActAnswer     string `bson:"act_answer"`       // 实际回复
	IsPass        bool   `bson:"is_pass"`          // 是否通过
	Source        string `bson:"source"`           // 回复命中类型
	GroupID       string `bson:"exp_group_id"`     // 期望GroupId
	ActGroupID    string `bson:"act_group_id"`     // 实际GroupId
	IsGroupIDPass bool   `bson:"is_group_id_pass"` // GroupId是否匹配
	IsFullPass    bool   `bson:"is_full_pass"`     // 是否完全匹配
	AlgoScore     string `bson:"algo_score"`       // 算法得分score
	EdgCost       int64  `bson:"edg_cost"`         // 端测耗时(ms)
	ExecuteTime   int64  `bson:"execute_time"`     // 此条用例运行时间点
	TaskName      string `bson:"task_name"`        // 测试计划名
	TraceId       string `bson:"trace_id"`         // TraceID
	IsSmoke       int    `bson:"is_smoke"`         //
}

type QATask struct {
	BaseTask
	QAConfig           *QATaskConfig
	QADataSourceConfig *QADataSource
	chanNum            int
	req                []*QATaskReq
	Results            []*QATaskRes
	RespChan           chan *QATaskOnceResp
	RightCount         int
	WrongCount         int
	Summary            string
	SummaryFile        string
	JobInstanceId      string
	startTime          time.Time
	endTime            time.Time
	cost               time.Duration
	mu                 sync.Mutex
}

type QATaskTest struct {
	*QATask
}

func NewQATask(qa *QATaskConfig, req []*QATaskReq, qaDataSourceConfig *QADataSource) *QATask {
	return &QATask{
		QAConfig:           qa,
		QADataSourceConfig: qaDataSourceConfig,
		req:                req,
		Results:            make([]*QATaskRes, 0),
	}
}

func SplitQACases(s []*QATaskReq) (f []*QATaskReq) {
	for _, req := range s {
		QuestionList := strings.Split(req.Query, "&&")
		RobotTypeList := strings.Split(req.RobotType, "&&")
		for _, singleQ := range QuestionList {
			for _, singleR := range RobotTypeList {
				////修改请求信息 改成单个问题单个机型去请求
				singleRequest := Copy(req).(*QATaskReq)
				singleRequest.Query = singleQ
				singleRequest.RobotType = singleR
				f = append(f, singleRequest)
			}
		}
	}
	return f
}

func (QA *QATask) pre() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	success, value := PrepareMissionFlag(QA.QAConfig.TaskName, cancel)
	if !success {
		return
	}
	if QA.QAConfig.JobInstanceId == "" {
		QA.JobInstanceId = uuid.New().String()
	}
	value.TaskType = CommonQA
	value.JobInstanceId = QA.JobInstanceId

	// 从数据库中获取用例
	if len(QA.req) == 0 || QA.QADataSourceConfig != nil {
		QA.CaseGetterQA(ctx)
	}

	QA.req = SplitQACases(QA.req) // 先处理下带有&&的query
	QA.RespChan = make(chan *QATaskOnceResp, len(QA.req))
	QA.chanNum = QA.QAConfig.ChanNum
	QA.startTime = time.Now()
}

func (QA *QATask) run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if success, _ := RunMissionFlag(QA.QAConfig.TaskName); !success {
		return
	}

	taskInfoMap[QA.QAConfig.TaskName].Cancel = cancel
	QAChan := make(chan *QATaskReq)

	for i := 0; i < QA.chanNum; i++ {
		go func(ctx context.Context, i int, v chan *QATaskReq) {
			defer QA.chanClose()
			conn, err := grpc.Dial(QA.QAConfig.ConnAddr, grpc.WithInsecure())
			defer conn.Close()
			if err != nil {
				// TODO 失败标记
				return
			}
			for req := range v {
				select {
				case <-ctx.Done():
					close(QAChan)
					return
				default:
					res := QA.call(conn, req)
					QA.RespChan <- res
				}
			}
		}(ctx, i, QAChan)
	}

	caseTotal := len(QA.req)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for i, req := range QA.req {
			select {
			case <-QAChan:
				wg.Done()
				return
			default:
				QAChan <- req
				if value, ok := taskInfoMap[QA.QAConfig.TaskName]; ok {
					value.ProgressPercent = (i + 1) * 100 / caseTotal
					value.Progress = fmt.Sprintf(`%d/%d`, i+1, caseTotal)
					value.Accuracy = float32(QA.RightCount) / float32(QA.RightCount+QA.WrongCount)
				}
			}
		}
		wg.Done()
		close(QAChan)
	}()
	wg.Wait()

	var QAResultList []interface{}
	for resp := range QA.RespChan {
		QAResultList = append(QAResultList, &QAResults{
			JobInstanceId: QA.JobInstanceId,
			Id:            resp.Req.Id,
			Question:      resp.Req.Query,
			Answer:        strings.Join(resp.Req.ExpectAnswer, "&&"),
			ActAnswer:     resp.Res.ActAnswer,
			IsPass:        resp.IsPass,
			EdgCost:       resp.EdgCost.Milliseconds(),
			ExecuteTime:   resp.Res.ExecuteTime,
			TaskName:      QA.QAConfig.TaskName,
			Source:        resp.Res.Source,
			TraceId:       resp.Res.TraceId,
			IsSmoke:       resp.Req.IsSmoke,
		})
	}
	models.ReporterDB.MongoInsertMany(qaResultsTable, QAResultList)
}

func (QA *QATask) end() {
	if success, _ := EndMissionFlag(QA.QAConfig.TaskName); !success {
		return
	}
	QA.endTime = time.Now()
	QA.getResultSummary()
	QA.writeQAResultExcel()
	QA.sendReport()
}

func (QA *QATask) stop() {
	QA.endTime = time.Now()
}

func (QA *QATask) chanClose() {
	QA.mu.Lock()
	defer QA.mu.Unlock()
	if QA.chanNum <= 1 {
		close(QA.RespChan)
	} else {
		QA.chanNum--
	}
}

func (QA *QATask) call(conn *grpc.ClientConn, req *QATaskReq) *QATaskOnceResp {
	executeTime, _ := strconv.ParseInt(time.Now().Format("20060102150405"), 10, 64)
	Res := &QATaskOnceResp{
		Req: req,
		Res: &QATaskRes{
			ActAnswer:   "",
			ExecuteTime: executeTime,
			TraceId:     uuid.New().String() + "@cloudminds-test.com",
		},
	}
	// do test
	c := talk.NewTalkClient(conn)
	r := &talk.TalkRequest{
		IsFull:     true,
		AgentID:    QA.QAConfig.AgentId,
		SessionID:  "QATest",
		QuestionID: Res.Res.TraceId,
		EventType:  talk.Text,
		EnvInfo:    make(map[string]string),
		RobotID:    QA.QAConfig.RobotID,
		TestMode:   true,
		TenantCode: QA.QAConfig.TenantCode,
		Version:    QA.QAConfig.Version, //speech.Header.Version,
		Asr: talk.Asr{
			Lang: "ZH",
			Text: req.Query,
		},
	}

	startReq := time.Now()
	resp, err := c.Talk(context.Background(), r)
	Res.EdgCost.Duration = time.Now().Sub(startReq)
	if err != nil {
		// TODO 失败处理
		return Res
	}

	//获取GroupId，没办法这个太深了，只能这样去取，要是有更好办法可以提供
	if resp.HitLog.Fields["qaresult"] != nil {
		answer := resp.HitLog.Fields["qaresult"].GetStructValue().Fields["answer"]
		if answer != nil {
			groupId := answer.GetStructValue().Fields["qgroupid"]
			if groupId != nil {
				id := groupId.GetStringValue()
				gid, err := strconv.ParseInt(id, 10, 64)
				if err == nil {
					Res.Res.GroupId = gid
					if Res.Res.GroupId == req.ExpectGroup {
						Res.IsGroupIdPass = true
					}
				}
			}
		}
	}
	//获取算法得分score，需求来源：http://jira.cloudminds.com/browse/SV-5722
	if resp.HitLog.Fields["qaresult"] != nil {
		qaResult := resp.HitLog.Fields["qaresult"].GetStructValue().Fields
		if qaResult["answer"] != nil {
			answer := qaResult["answer"].GetStructValue().Fields
			if answer["score"] != nil {
				Res.Res.AlgoScore = answer["score"].GetNumberValue() //以浮点类型形式传递值
			}
		}
	}
	if len(resp.Tts) == 0 {
		Res.Res.ActAnswer = ""
		Res.IsPass = false
		Res.IsGroupIdPass = false
		return Res
	}

	Res.Res.ActAnswer = resp.Tts[0].Text

	fmt.Println(Res.Res.ActAnswer)
	Res.Res.Source = resp.Source
	if resp.HitLog.Fields["domain"] != nil && resp.HitLog.Fields["domain"].GetStringValue() != "" {
		Res.Res.Source += "/"
		Res.Res.Source += resp.HitLog.Fields["domain"].GetStringValue()
	}
	if resp.HitLog.Fields["intent"] != nil && resp.HitLog.Fields["intent"].GetStringValue() != "" {
		Res.Res.Source += "/"
		Res.Res.Source += resp.HitLog.Fields["intent"].GetStringValue()
	}
	//断言答案：看看切片里面有没有这个回答内容
	for _, expA := range req.ExpectAnswer {
		// 看看每个期望答案是不是实际答案的子集
		if strings.Contains(RemoveSpecialSign(Res.Res.ActAnswer), RemoveSpecialSign(expA)) { //2022-10-19 由于Excel导入时特殊字符影响断言 所以先将字符去除后来断言
			Res.IsTextPass = true
			break
		}
	}

	//是否要通过group_id来判断对错
	if QA.QAConfig.IsGroupId == "yes" {
		if Res.IsTextPass && Res.IsGroupIdPass {
			Res.IsPass = true
		} else {
			Res.IsPass = false
		}
	} else {
		if Res.IsTextPass {
			Res.IsPass = true
		} else {
			Res.IsPass = false
		}
	}

	if resp.HitLog.Fields["algo"].GetStringValue() == "exact" {
		Res.IsExactMatch = true
	}

	if Res.IsPass {
		QA.RightCount++
	} else {
		QA.WrongCount++
	}
	return Res
}

// RemoveSpecialSign 删除特殊字符，仅保留中文、英文、阿拉伯数字
func RemoveSpecialSign(origin string) (final string) {
	result := regexp.MustCompile(fmt.Sprintf(`[一-龥\w]+`)).FindAll([]byte(origin), -1)
	for _, r := range result {
		final += string(r)
	}
	return final
}

var (
	_              TaskModel = &QATask{}
	qaResultsTable           = "qa_results"
)

type Interface interface {
	DeepCopy() interface{}
}

func Copy(src interface{}) interface{} {
	if src == nil {
		return nil
	}
	original := reflect.ValueOf(src)
	cpy := reflect.New(original.Type()).Elem()
	copyRecursive(original, cpy)

	return cpy.Interface()
}

func copyRecursive(src, dst reflect.Value) {
	if src.CanInterface() {
		if copier, ok := src.Interface().(Interface); ok {
			dst.Set(reflect.ValueOf(copier.DeepCopy()))
			return
		}
	}

	switch src.Kind() {
	case reflect.Ptr:
		originalValue := src.Elem()

		if !originalValue.IsValid() {
			return
		}
		dst.Set(reflect.New(originalValue.Type()))
		copyRecursive(originalValue, dst.Elem())

	case reflect.Interface:
		if src.IsNil() {
			return
		}
		originalValue := src.Elem()
		copyValue := reflect.New(originalValue.Type()).Elem()
		copyRecursive(originalValue, copyValue)
		dst.Set(copyValue)

	case reflect.Struct:
		t, ok := src.Interface().(time.Time)
		if ok {
			dst.Set(reflect.ValueOf(t))
			return
		}
		for i := 0; i < src.NumField(); i++ {
			if src.Type().Field(i).PkgPath != "" {
				continue
			}
			copyRecursive(src.Field(i), dst.Field(i))
		}

	case reflect.Slice:
		if src.IsNil() {
			return
		}
		dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
		for i := 0; i < src.Len(); i++ {
			copyRecursive(src.Index(i), dst.Index(i))
		}

	case reflect.Map:
		if src.IsNil() {
			return
		}
		dst.Set(reflect.MakeMap(src.Type()))
		for _, key := range src.MapKeys() {
			originalValue := src.MapIndex(key)
			copyValue := reflect.New(originalValue.Type()).Elem()
			copyRecursive(originalValue, copyValue)
			copyKey := Copy(key.Interface())
			dst.SetMapIndex(reflect.ValueOf(copyKey), copyValue)
		}

	default:
		dst.Set(src)
	}
}
