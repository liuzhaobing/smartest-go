package task

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"smartest-go/models"
	"smartest-go/pkg/app"
	"smartest-go/pkg/e"
	util "smartest-go/pkg/util/const"
	"sort"
	"time"
)

// TaskInfo 单个任务详情
type TaskInfo struct {
	JobInstanceId   string             `json:"job_instance_id" bson:"job_instance_id" form:"job_instance_id,omitempty"`    // 测试任务实例
	Name            string             `json:"task_name" bson:"task_name" form:"task_name,omitempty"`                      // 测试任务名称
	TaskType        string             `json:"task_type" bson:"task_type" form:"task_type,omitempty"`                      // 任务类型
	Status          int                `json:"status" bson:"status" form:"status,omitempty"`                               // 测试状态 32成功 64失败 128已停止 256执行中 512就绪
	ProgressPercent int                `json:"progress_percent" bson:"progress_percent" form:"progress_percent,omitempty"` // 测试进度 0.8
	Progress        string             `json:"progress" bson:"progress" form:"progress,omitempty"`                         // 测试进度 225/225
	Accuracy        float32            `json:"accuracy" bson:"accuracy" form:"accuracy,omitempty"`                         // 准确率
	Message         string             `json:"message" bson:"message" form:"message,omitempty"`                            // 其他消息
	StartTime       string             `json:"start_time" bson:"start_time" form:"start_time,omitempty"`                   // 开始时间
	EndTime         string             `json:"end_time" bson:"end_time" form:"end_time,omitempty"`                         // 结束时间
	Cancel          context.CancelFunc `json:"-" bson:"-" form:"-"`
	PageNum         int                `json:"-" bson:"-" form:"page_num,omitempty,default=1"`
	PageSize        int                `json:"-" bson:"-" form:"page_size,omitempty,default=30"`
}

type MyTaskList []*TaskInfo

func (t MyTaskList) Less(i, j int) bool {
	if t[i].StartTime > t[j].StartTime {
		return true
	}
	return false
}

func (t MyTaskList) Swap(i, j int) {
	t[i], t[j] = t[j], t[i]
}

func (t MyTaskList) Len() int {
	return len(t)
}

// taskInfoMap 存储当前正在运行的所有的任务
var taskInfoMap = make(map[string]*TaskInfo)

// Status 查询测试进度
func Status(context *gin.Context) {
	req := context.MustGet(util.REQUEST_KEY).(*NameOfTask)
	var infoList MyTaskList
	if req.TaskName == "" || req == nil {
		for _, value := range taskInfoMap {
			infoList = append(infoList, value)
		}
		sort.Sort(infoList)
	} else if value, ok := taskInfoMap[req.TaskName]; ok {
		infoList = append(infoList, value)
	}
	app.SuccessRespByCode(context, e.SUCCESS, infoList)
}

func LookUpStatus(context *gin.Context) {
	req := context.MustGet(util.REQUEST_KEY).(*TaskInfo)
	var infoList MyTaskList

	tempMap := make(map[string]*TaskInfo)
	statusFromDB(req, tempMap)
	statusFromLocal(req, tempMap)
	if tempMap == nil || len(tempMap) == 0 {
		app.SuccessRespByCode(context, e.SUCCESS, nil)
		return
	}
	for _, value := range tempMap {
		infoList = append(infoList, value)
	}
	sort.Sort(infoList)
	if req.PageSize*req.PageNum >= len(infoList) {
		app.SuccessRespByCode(context, e.SUCCESS, struct {
			Total int64      `json:"total"`
			Data  MyTaskList `json:"data"`
		}{
			Total: int64(len(infoList)),
			Data:  infoList[(req.PageSize * (req.PageNum - 1)):],
		})
		return
	}
	app.SuccessRespByCode(context, e.SUCCESS, struct {
		Total int64      `json:"total"`
		Data  MyTaskList `json:"data"`
	}{
		Total: int64(len(infoList)),
		Data:  infoList[req.PageSize*(req.PageNum-1) : req.PageSize*req.PageNum],
	})
}

func statusFromLocal(req *TaskInfo, tempTaskInfoMap map[string]*TaskInfo) {
	for _, value := range taskInfoMap {
		if req != nil {
			if ((req.Name != "" && value.Name == req.Name) || req.Name == "") &&
				((req.TaskType != "" && value.TaskType == req.TaskType) || req.TaskType == "") &&
				((req.JobInstanceId != "" && value.JobInstanceId == req.JobInstanceId) || req.JobInstanceId == "") &&
				((req.Status != 0 && value.Status == req.Status) || req.Status == 0) {
				tempTaskInfoMap[value.JobInstanceId] = value
			}
		} else {
			tempTaskInfoMap[value.JobInstanceId] = value
		}
	}
}

func statusFromDB(req *TaskInfo, tempTaskInfoMap map[string]*TaskInfo) {
	f := bson.M{}
	if req != nil {
		if req.Name != "" {
			f["task_name"] = req.Name
		}
		if req.TaskType != "" {
			f["task_type"] = req.TaskType
		}
		if req.JobInstanceId != "" {
			f["job_instance_id"] = req.JobInstanceId
		}
		if req.Status != 0 {
			f["status"] = req.Status
		}
	}
	results, err := models.ReporterDB.MongoFind(TasksTable, f)
	if err != nil {
		return
	}
	for _, r := range results {
		var dbTaskInfo *TaskInfo
		b, _ := json.Marshal(r.Map())
		json.Unmarshal(b, &dbTaskInfo)
		tempTaskInfoMap[dbTaskInfo.JobInstanceId] = dbTaskInfo
	}
}

/*
状态变更的相关操作
*/

// PrepareMissionFlag 任务准备时 保存记录  //512准备中 -> 256执行中 -> 128已停止 -> 64失败 -> 32成功
func PrepareMissionFlag(taskName string, c context.CancelFunc) (bool, *TaskInfo) {
	thisTaskInfo := &TaskInfo{
		Name:      taskName,
		Status:    512,
		Message:   "任务准备中!",
		StartTime: time.Now().Format("2006-01-02 15:04:05"),
		Cancel:    c}

	if value, ok := taskInfoMap[taskName]; ok && value.Status == 256 { // 执行中的任务不允许再次发起 pre()
		return false, value
	}

	taskInfoMap[taskName] = thisTaskInfo
	finalResult, _ := json.Marshal(thisTaskInfo)
	fmt.Println(string(finalResult))
	return true, thisTaskInfo
}

// RunMissionFlag 任务发起时 保存记录  //512准备中 -> 256执行中 -> 128已停止 -> 64失败 -> 32成功
func RunMissionFlag(taskName string) (bool, *TaskInfo) {
	if value, ok := taskInfoMap[taskName]; ok && value.Status == 512 { // 只有准备中的任务才能执行 run()
		value.Message = "任务执行中!"
		value.Status = 256
		finalResult, _ := json.Marshal(value)
		fmt.Println(string(finalResult))
		return true, value
	}
	return false, nil
}

// EndMissionFlag 任务完成时 保存记录  //512准备中 -> 256执行中 -> 128已停止 -> 64失败 -> 32成功
func EndMissionFlag(taskName string) (success bool, info *TaskInfo) {
	value, ok := taskInfoMap[taskName]
	if ok && value.Status == 256 { // 只有执行中的任务才能结束 end()
		value.Status = 32
		value.Message = "任务执行结束!"
		value.EndTime = time.Now().Format("2006-01-02 15:04:05")
		models.ReporterDB.MongoInsertOne(TasksTable, value)
		delete(taskInfoMap, taskName)
		finalResult, _ := json.Marshal(value)
		fmt.Println(string(finalResult))
		return true, value
	}
	if ok && value.Status == 128 { // 已经手动停止的任务也默认已结束 end()
		return true, value
	}
	return false, nil
}

// TerminateMissionFlag 任务手动停止时 更新记录  //512准备中 -> 256执行中 -> 128已停止 -> 64失败 -> 32成功
func TerminateMissionFlag(taskName string) (success bool, info *TaskInfo) {
	value, ok := taskInfoMap[taskName]
	if ok && (value.Status == 256 || value.Status == 512) {
		value.Cancel()
		value.Message = "任务停止成功!"
		value.Status = 128
		value.EndTime = time.Now().Format("2006-01-02 15:04:05")
		models.ReporterDB.MongoInsertOne(TasksTable, value)
		delete(taskInfoMap, taskName)
		finalResult, _ := json.Marshal(value)
		fmt.Println(string(finalResult))
		return true, value
	}
	if ok && (value.Status == 32 || value.Status == 64) {
		return true, value
	}
	return false, nil
}

var (
	TasksTable = "tasks" // 任务结束时记录本次任务job_instance_id到mongo数据库
)
