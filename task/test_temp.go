package task

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"smartest-go/pkg/app"
	"smartest-go/pkg/e"
	util "smartest-go/pkg/util/const"
	"time"
)

// Start 启动测试
func Start(context *gin.Context) {
	req := context.MustGet(util.REQUEST_KEY).(*AddTask)
	thisTaskInfo := &TaskInfo{
		Name:   req.TaskName,
		Status: 512, // 512就绪
	}

	value, ok := taskInfoMap[req.TaskName]
	// 先判断此任务是否正在执行
	if ok && value.Status == 256 { // 256执行中
		// 查到此任务正在运行
		value.Message = "任务执行中!"
		app.SuccessRespByCode(context, e.ERROR, value)
	} else {
		// 判断任务类型 决定采用什么去执行
		switch req.TaskType {
		case KnowledgeGraph:
			thisTaskInfo.JobInstanceId, thisTaskInfo.Status = KGRunner(req)
		}
		thisTaskInfo.TaskType = req.TaskType

		// 判断任务发起状态
		if thisTaskInfo.Status == 256 {
			// 开始执行后 将状态置为running 存入到正在running的taskInfoMap
			thisTaskInfo.Message = "任务发起成功!"
			thisTaskInfo.StartTime = time.Now().Format("2006-01-02 15:04:05")
			taskInfoMap[req.TaskName] = thisTaskInfo
			app.SuccessRespByCode(context, e.SUCCESS, thisTaskInfo)
		} else {
			thisTaskInfo.Message = "任务发起失败!"
			app.SuccessRespByCode(context, e.ERROR, thisTaskInfo)
		}
	}
}

// Stop 终止测试
func Stop(context *gin.Context) {
	req := context.MustGet(util.REQUEST_KEY).(*NameOfTask)
	if value, ok := taskInfoMap[req.TaskName]; ok {
		value.Cancel()
		value.Message = "任务停止成功!"
		value.Status = 128 // 128已停止
		value.EndTime = time.Now().Format("2006-01-02 15:04:05")
		app.SuccessRespByCode(context, e.SUCCESS, value)
	} else {
		// 未查到此任务 返回
		app.SuccessRespByCode(context, e.ERROR, &TaskInfo{
			Name:    req.TaskName,
			Status:  64, // 64失败
			Message: "无法停止不存在的任务!",
		})
	}
}

func KGRunner(r *AddTask) (jobInstanceId string, status int) {
	if r.TaskConfig.TestConfigKG == nil {
		return "", 64
	}

	if r.TaskDataSource.TestCaseKG == nil {
		return "", 64
	}

	var kg TaskModel = &KGTask{}
	kg = &KGTaskTest{
		KGTask: NewKGTask(r.TaskConfig.TestConfigKG,
			r.TaskDataSource.TestCaseKG,
			&KGDataSource{},
		),
	}

	var b = &BaseTask{
		T: &kg,
	}

	if r.TaskConfig.TestConfigKG.JobInstanceId == "" {
		r.TaskConfig.TestConfigKG.JobInstanceId = uuid.New().String()
	}
	go b.Run()
	return r.TaskConfig.TestConfigKG.JobInstanceId, 256
}
