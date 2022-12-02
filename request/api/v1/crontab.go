package v1

import (
	"github.com/gin-gonic/gin"
	"smartest-go/pkg/app"
	"smartest-go/pkg/e"
	util "smartest-go/pkg/util/const"
	"smartest-go/task"
)

/*
定时器相关操作
*/

// ListCronPlan 从定时器中查询计划列表
func ListCronPlan(context *gin.Context) {
	cronList, err := task.CM.GetCronTaskList()
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	app.SuccessResp(context, cronList)
}

// RemoveCronPlan 从定时器中移除单个计划
func RemoveCronPlan(context *gin.Context) {
	req := context.MustGet(util.REQUEST_KEY).(*task.NameOfTask)
	_, err := task.CM.RemoveCronTaskByName(req.TaskName)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	app.SuccessResp(context, nil)
}

// AddCronPlan 新增单个计划到定时器
func AddCronPlan(context *gin.Context) {
	req := context.MustGet(util.REQUEST_KEY).(*task.AddTask)
	if req.IsCrontab == "yes" {
		job := task.InitTaskModel(req)
		_, err := task.CM.AddCronTask(req, job)
		if err != nil {
			app.ErrorResp(context, e.ERROR, err.Error(), nil)
			return
		}
	}
	app.SuccessResp(context, req)
}
