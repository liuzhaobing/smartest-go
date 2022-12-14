package task

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"io"
	"io/ioutil"
	"os"
	"smartest-go/models"
	asrcloudminds "smartest-go/proto/asr"
	asrcontrol "smartest-go/proto/asrctl"
	"smartest-go/proto/common"
	"strconv"
	"strings"
	"sync"
	"time"
)

type ASRTaskConfig struct {
	TaskName        string          `json:"task_name" form:"task_name"`
	JobInstanceId   string          `json:"job_instance_id" form:"job_instance_id"`
	IsReport        string          `json:"is_report" form:"is_report"`
	ReportString    []*ReportString `json:"report_string" form:"report_string"`
	ChanNum         int             `json:"chan_num"  form:"chan_num"`
	AgentId         int64           `json:"agent_id" form:"agent_id"`
	RobotId         string          `json:"robot_id" form:"robot_id"`
	IsASRCloudMinds string          `json:"is_asr_cloud_minds" form:"is_asr_cloud_minds"`
	IsASRCtrl       string          `json:"is_asr_ctrl" form:"is_asr_ctrl"`
	ASRAddr         string          `json:"asr_addr" form:"asr_addr"`
	ASRCtrlAddr     string          `json:"asr_ctrl_addr" form:"asr_ctrl_addr"`
	IsAddHot        string          `json:"is_add_hot" form:"is_add_hot"` // 是否需要在前端添加热词
	EnvInfo         *EnvInfo        `json:"env_info" form:"env_info"`     // 前端信息
}

type ASRDataSource struct {
	DBFilter string `json:"filter,omitempty"  form:"filter,omitempty"`
}

type ASRTaskReq struct {
	Id         int64  `json:"id,omitempty" form:"id,omitempty"` // 用例编号
	ExpMessage string `json:"exp_message,omitempty"`            // 期望识别结果
	Tags       string `json:"tags,omitempty"`                   // 用例标签
	WavFile    string `json:"wav_file,omitempty"`               // wav音频文件路径
	WavData    []byte `json:"wav_data,omitempty"`               // wav音频文件转字节
	IsSmoke    int64  `json:"is_smoke,omitempty"`
}

type ASRTaskRes struct { // 从响应体收集到的数据
	ASRCloudMindsMessage string
	ASRCtrlMessage       string
	ASRCloudMindsTraceId string
	ASRCtrlTraceId       string
	ServerExtendMessage  string //版本信息
	ExecuteTime          int64
}

type ASRTaskOnceResp struct {
	Req           *ASRTaskReq
	Res           *ASRTaskRes
	IsASRPass     bool
	IsASRCtrlPass bool
	EdgCostASR    jsonTime
	EdgCostCtrl   jsonTime
}

type ASRResults struct {
	JobName                 string `json:"task_name,omitempty"      bson:"task_name,omitempty"`             // JobName+JobId 唯一标识一个测试任务
	JobInstanceId           string `json:"job_instance_id,omitempty"      bson:"job_instance_id,omitempty"` // JobInstanceId+ExecuteTime 唯一标识一个测试执行历史
	ExecuteTime             int64  `json:"execute_time,omitempty"      bson:"execute_time,omitempty"`       // JobInstanceId+ExecuteTime 唯一标识一个测试执行历史
	CaseNumber              int64  `json:"id,omitempty"      bson:"id,omitempty"`
	CostASR                 int64  `json:"edg_cost_asr,omitempty"      bson:"edg_cost_asr,omitempty"`
	CostASRCtrl             int64  `json:"edg_cost_asr_ctrl,omitempty"      bson:"edg_cost_asr_ctrl,omitempty"`
	ExpectMessage           string `json:"expect_message,omitempty"  bson:"expect_message,omitempty"`
	FilePath                string `json:"file_path,omitempty"  bson:"file_path,omitempty"`
	ActMessageASRCloudMinds string `json:"act_message_asr_cloud_minds,omitempty"  bson:"act_message_asr_cloud_minds,omitempty"`
	ActMessageASRCtrl       string `json:"act_message_asr_ctrl,omitempty"  bson:"act_message_asr_ctrl,omitempty"`
	IsASRPass               bool   `json:"is_asr_pass,omitempty" bson:"is_asr_pass,omitempty"`
	IsASRCtrlPass           bool   `json:"is_asr_ctrl_pass,omitempty" bson:"is_asr_ctrl_pass,omitempty"`
	ASRCloudMindsTraceId    string `json:"asr_cloud_minds_trace_id,omitempty"  bson:"asr_cloud_minds_trace_id,omitempty"`
	ASRCtrlTraceId          string `json:"asr_ctrl_trace_id,omitempty"  bson:"asr_ctrl_trace_id,omitempty"`
	ServerExtendMessage     string `json:"server_extend_message,omitempty"  bson:"server_extend_message,omitempty"` //版本信息
	Tags                    string `json:"tags,omitempty"  bson:"tags,omitempty"`
	IsSmoke                 int64  `bson:"is_smoke,omitempty"  bson:"is_smoke,omitempty"`
}

type ASRTask struct {
	BaseTask
	ASRConfig           *ASRTaskConfig
	ASRDataSourceConfig *ASRDataSource
	chanNum             int
	req                 []*ASRTaskReq
	Results             []*ASRTaskRes
	RespChan            chan *ASRTaskOnceResp
	RightCount          int
	WrongCount          int
	Summary             string
	SummaryFile         string
	JobInstanceId       string
	startTime           time.Time
	endTime             time.Time
	cost                time.Duration
	mu                  sync.Mutex
}

type ASRTaskTest struct {
	*ASRTask
}

func NewASRTask(asr *ASRTaskConfig, req []*ASRTaskReq, asrDataSourceConfig *ASRDataSource) *ASRTask {
	return &ASRTask{
		ASRConfig:           asr,
		ASRDataSourceConfig: asrDataSourceConfig,
		req:                 req,
		Results:             make([]*ASRTaskRes, 0),
	}
}

func (ASR *ASRTask) pre() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	success, value := PrepareMissionFlag(ASR.ASRConfig.TaskName, cancel)
	if !success {
		return
	}
	if ASR.ASRConfig.JobInstanceId == "" {
		ASR.JobInstanceId = uuid.New().String()
	}
	value.TaskType = CommonASR
	value.JobInstanceId = ASR.JobInstanceId

	if len(ASR.req) == 0 || ASR.ASRDataSourceConfig != nil {
		ASR.CaseGetterASR(ctx) // 从数据库中获取用例
	}

	ASR.RespChan = make(chan *ASRTaskOnceResp, len(ASR.req))
	ASR.chanNum = ASR.ASRConfig.ChanNum
	ASR.startTime = time.Now()
	ASR.AddHWByCases() // 用例测试前先添加热词
}

func (ASR *ASRTask) AddHWByCases() {

}

func (ASR *ASRTask) run() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if success, _ := RunMissionFlag(ASR.ASRConfig.TaskName); !success {
		return
	}

	taskInfoMap[ASR.ASRConfig.TaskName].Cancel = cancel
	ASRChan := make(chan *ASRTaskReq)

	for i := 0; i < ASR.chanNum; i++ {
		go func(ctx context.Context, i int, v chan *ASRTaskReq) {
			defer ASR.chanClose()
			connCloudMinds, err := grpc.Dial(ASR.ASRConfig.ASRAddr, grpc.WithInsecure())
			defer connCloudMinds.Close()
			if err != nil && ASR.ASRConfig.IsASRCloudMinds == "yes" {
				value := taskInfoMap[ASR.ASRConfig.TaskName]
				value.Status = 64
				value.Message = ASR.ASRConfig.ASRAddr + "地址请求失败(asr)！" + err.Error()
				value.Cancel()
			}
			connCtrl, err := grpc.Dial(ASR.ASRConfig.ASRCtrlAddr, grpc.WithInsecure())
			defer connCtrl.Close()
			if err != nil && ASR.ASRConfig.IsASRCtrl == "yes" {
				value := taskInfoMap[ASR.ASRConfig.TaskName]
				value.Status = 64
				value.Message = ASR.ASRConfig.ASRCtrlAddr + "地址请求失败(asr control)！" + err.Error()
				value.Cancel()
			}

			for req := range v {
				select {
				case <-ctx.Done():
					close(ASRChan)
					return
				default:
					if !strings.Contains(req.WavFile, "./upload/wav/") {
						req.WavFile = "./upload/wav/" + req.WavFile
					}
					res := ASR.call(connCloudMinds, connCtrl, req)
					ASR.RespChan <- res
				}
			}
		}(ctx, i, ASRChan)
	}

	caseTotal := len(ASR.req)

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		for i, req := range ASR.req {
			select {
			case <-ASRChan:
				wg.Done()
				return
			default:
				ASRChan <- req
				if value, ok := taskInfoMap[ASR.ASRConfig.TaskName]; ok {
					value.ProgressPercent = (i + 1) * 100 / caseTotal
					value.Progress = fmt.Sprintf(`%d/%d`, i+1, caseTotal)
					value.Accuracy = float32(ASR.RightCount) / float32(ASR.RightCount+ASR.WrongCount)
				}
			}
		}
		wg.Done()
		close(ASRChan)
	}()
	wg.Wait()

	var ASRResultList []interface{}
	for resp := range ASR.RespChan {
		ASRResultList = append(ASRResultList, &ASRResults{
			JobName:                 ASR.ASRConfig.TaskName,
			JobInstanceId:           ASR.JobInstanceId,
			ExecuteTime:             resp.Res.ExecuteTime,
			CaseNumber:              resp.Req.Id,
			CostASR:                 resp.EdgCostASR.Milliseconds(),
			CostASRCtrl:             resp.EdgCostCtrl.Milliseconds(),
			ExpectMessage:           resp.Req.ExpMessage,
			FilePath:                resp.Req.WavFile,
			ActMessageASRCloudMinds: resp.Res.ASRCloudMindsMessage,
			IsASRPass:               resp.IsASRPass,
			IsASRCtrlPass:           resp.IsASRCtrlPass,
			ActMessageASRCtrl:       resp.Res.ASRCtrlMessage,
			ASRCloudMindsTraceId:    resp.Res.ASRCloudMindsTraceId,
			ASRCtrlTraceId:          resp.Res.ASRCtrlTraceId,
			ServerExtendMessage:     resp.Res.ServerExtendMessage,
			Tags:                    resp.Req.Tags,
			IsSmoke:                 resp.Req.IsSmoke,
		})
	}
	models.ReporterDB.MongoInsertMany(ASRResultsTable, ASRResultList)
}

func (ASR *ASRTask) end() {
	ASR.endTime = time.Now()
	ASR.getResultSummary()
	ASR.writeASRResultExcel()
	ASR.sendReport()
	if success, _ := EndMissionFlag(ASR.ASRConfig.TaskName, ASR.Summary, ASR.SummaryFile); !success {
		return
	}
}

func (ASR *ASRTask) stop() {
	ASR.endTime = time.Now()
}

func (ASR *ASRTask) chanClose() {
	ASR.mu.Lock()
	defer ASR.mu.Unlock()
	if ASR.chanNum <= 1 {
		close(ASR.RespChan)
	} else {
		ASR.chanNum--
	}
}

func (ASR *ASRTask) call(connCloudMinds, connCtrl *grpc.ClientConn, req *ASRTaskReq) *ASRTaskOnceResp {
	executeTime, _ := strconv.ParseInt(time.Now().Format("20060102150405"), 10, 64)
	Res := &ASRTaskOnceResp{
		Req: req,
		Res: &ASRTaskRes{
			ExecuteTime: executeTime,
		},
		IsASRPass:     false,
		IsASRCtrlPass: false,
	}

	// do test
	if ASR.ASRConfig.IsASRCloudMinds == "yes" && connCloudMinds != nil {
		ASR.asrCloudMindsCall(connCloudMinds, req, Res)
	}
	if ASR.ASRConfig.IsASRCtrl == "yes" && connCtrl != nil {
		ASR.asrCtrlCall(connCtrl, req, Res)
	}

	if ASR.ASRConfig.IsASRCloudMinds == "yes" {
		if Res.IsASRPass {
			ASR.RightCount++
		} else {
			ASR.WrongCount++
		}
	} else {
		if Res.IsASRCtrlPass {
			ASR.RightCount++
		} else {
			ASR.WrongCount++
		}
	}

	return Res
}

func (ASR *ASRTask) asrCloudMindsCall(connCloudMinds *grpc.ClientConn, req *ASRTaskReq, Res *ASRTaskOnceResp) {
	wavData, err := ioutil.ReadFile(req.WavFile)
	if err != nil && req.WavData == nil {
		Res.IsASRCtrlPass = false
		return
	}

	if wavData != nil && strings.Contains(req.WavFile, ".wav") {
		req.WavData = wavData[44:]
	}

	c := asrcloudminds.NewAsrServiceClient(connCloudMinds)
	startTime := time.Now()
	Res.Res.ASRCloudMindsTraceId = uuid.New().String() + "@cloudminds-test.com"
	stream, err := c.AsrDecoder(context.Background())
	if err != nil {
		fmt.Printf("%v AsrDecoder Client AsrDecoder error: %v\n", Res.Res.ASRCloudMindsTraceId, err)
	}

	go func() {
		wavSize := len(req.WavData)
		blockSize := 1024
		t := wavSize / blockSize
		m := wavSize % blockSize

		for i := 0; i < t; i++ {
			data := req.WavData[i*blockSize : (i+1)*blockSize]

			err = stream.Send(&asrcloudminds.AsrDecoderRequest{
				WavData: data,
				WavSize: int32(blockSize),
				WavEnd:  int32(0),
				Guid:    Res.Res.ASRCloudMindsTraceId,
				AgentId: strconv.FormatInt(ASR.ASRConfig.AgentId, 10),
				RobotId: Res.Res.ASRCloudMindsTraceId,
			})

			if err != nil {
				fmt.Printf("%v AsrDecoder stream request err: %v\n", Res.Res.ASRCloudMindsTraceId, err)
				return
			}
		}

		var data []byte
		if m > 0 {
			data = req.WavData[t*blockSize : t*blockSize+m+1]
		}

		err = stream.Send(&asrcloudminds.AsrDecoderRequest{
			WavData: data,
			WavSize: int32(len(data)),
			WavEnd:  int32(0),
			Guid:    Res.Res.ASRCloudMindsTraceId,
			AgentId: strconv.FormatInt(ASR.ASRConfig.AgentId, 10),
			RobotId: Res.Res.ASRCloudMindsTraceId,
		})

		if err != nil {
			fmt.Printf("%v AsrDecoder stream request err: %v\n", Res.Res.ASRCloudMindsTraceId, err)
			return
		}

		var emptyData []byte
		err = stream.Send(&asrcloudminds.AsrDecoderRequest{
			WavData: emptyData,
			WavSize: int32(0),
			WavEnd:  int32(1),
			Guid:    Res.Res.ASRCloudMindsTraceId,
			AgentId: strconv.FormatInt(ASR.ASRConfig.AgentId, 10),
			RobotId: Res.Res.ASRCloudMindsTraceId,
		})

		if err != nil {
			fmt.Printf("%v AsrDecoder stream request err: %v\n", Res.Res.ASRCloudMindsTraceId, err)
			return
		}
		startTime = time.Now()
	}()

	for {
		res, err := stream.Recv()

		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Printf("%v AsrDecoder get stream err: %v\n", Res.Res.ASRCloudMindsTraceId, err)
			break
		}
		// 保存服务版本信息
		if res.Extend != "" {
			Res.Res.ServerExtendMessage = res.Extend
		}
		// 打印返回值
		if res.Message != "" {
			Res.Res.ASRCloudMindsMessage = res.Message
		}
	}

	//最后关闭流
	err = stream.CloseSend()

	if err != nil {
		fmt.Printf("%v AsrDecoder close stream err: %v\n", Res.Res.ASRCloudMindsTraceId, err)
	}
	Res.EdgCostASR.Duration = time.Now().Sub(startTime)

	Res.IsASRPass = resultAssertion(Res.Res.ASRCloudMindsMessage, req.ExpMessage)
}

func (ASR *ASRTask) asrCtrlCall(connCtrl *grpc.ClientConn, req *ASRTaskReq, Res *ASRTaskOnceResp) {
	c := asrcontrol.NewSpeechClient(connCtrl)
	startTime := time.Now()
	Res.Res.ASRCtrlTraceId = uuid.New().String() + "@cloudminds-test.com"
	vendor := "CloudMinds"
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()
	stream, err := c.StreamingRecognize(ctx)
	if err != nil {
		fmt.Printf("创建数据流失败: [%+v]\n", err)
		return
	}
	// 构造asr control请求体
	aqa := asrcontrol.RecognitionRequest{
		CommonReqInfo: &common.CommonReqInfo{
			Guid:        Res.Res.ASRCtrlTraceId,
			Timestamp:   time.Now().Unix(),
			Version:     "1.0",
			TenantId:    Res.Res.ASRCtrlTraceId,
			UserId:      Res.Res.ASRCtrlTraceId,
			RobotId:     Res.Res.ASRCtrlTraceId,
			ServiceCode: "ginger",
		},
		Body: &asrcontrol.Body{
			Option: map[string]string{
				"returnDetail":  "true",
				"recognizeOnly": "true",
				"tstAgentId":    strconv.FormatInt(ASR.ASRConfig.AgentId, 10),
			},
			Data: &asrcontrol.Body_Data{
				Rate:     16000,
				Format:   "pcm",
				Language: "CH",
				Dialect:  "zh",
				Vendor:   vendor,
			},
			NeedWakeup: false,
			Type:       1,
		},
		Extra: nil,
	}

	if vendor == "Baidu" {
		aqa.Body.Data.Dialect = "1537"
	} else if vendor == "IFlyTek" || vendor == "IFlyTekRealTime" {
		aqa.Body.Data.Dialect = "zh_cn"
	}

	audioFile, err := os.Open(req.WavFile)
	if err != nil {
		fmt.Println(err)
	}

	status := 0
	var buffer = make([]byte, 1280)
	skip := false
	if strings.Contains(req.WavFile, ".wav") {
		skip = true
	}
	cnt := 0

	for {
		lens, err := audioFile.Read(buffer)
		if err != nil || lens == 0 {
			status = 2
		}
		if skip {
			aqa.Body.Data.Speech = buffer[44:lens]
			skip = false
		} else {
			aqa.Body.Data.Speech = buffer[:lens]
		}

		cnt++
		if cnt == 40 {
			aqa.Body.StreamFlag = 2
			time.Sleep(time.Duration(200) * time.Millisecond)
		}

		// 向服务端发送 指令
		if err := stream.Send(&aqa); err != nil {
			return
		}
		if status == 2 {
			break
		}
		status = 1
	}

	aqa.Body.Data.Speech = nil
	aqa.Extra = &common.Extra{ExtraType: "audioExtra", ExtraBody: "val"}
	// 向服务端发送 指令
	if err := stream.Send(&aqa); err != nil {
		return
	}

	audioFile.Close()
	stream.CloseSend()

	// 接收 从服务端返回的数据流
	for {
		res, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			break
		}
		if res.DetailMessage != nil {
			Res.Res.ASRCtrlMessage = res.DetailMessage.Fields["recognizedText"].GetStringValue()
		}
	}
	Res.EdgCostCtrl.Duration = time.Now().Sub(startTime)
	Res.IsASRCtrlPass = resultAssertion(Res.Res.ASRCtrlMessage, req.ExpMessage)
}

//语音识别结果,将字符去掉,方便断言
func trimCharBeforeAssertion(s string) string {
	return TrimChar(s, []string{",", "。", "`", ".", " ", "(", ")"})
}

// TrimChar 将字符串中特定字符去除
func TrimChar(inputString string, s []string) string {
	for _, s2 := range s {
		inputString = strings.Replace(inputString, s2, "", -1)
	}
	return inputString
}

//断言识别的文本结果是否符合预期
func resultAssertion(str1, str2 string) bool {
	if trimCharBeforeAssertion(str1) == trimCharBeforeAssertion(str2) {
		return true
	} else {
		return false
	}
}

var (
	_               TaskModel = &ASRTask{}
	ASRResultsTable           = "asr_results"
)
