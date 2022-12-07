package v1

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"smartest-go/models"
	"smartest-go/pkg/app"
	"smartest-go/pkg/e"
	util "smartest-go/pkg/util/const"
	"smartest-go/task"
	"strconv"
)

// ListPlan 从数据库中查询所有计划列表
func ListPlan(context *gin.Context) {
	req := context.MustGet(util.REQUEST_KEY).(*task.ListTask)
	model := models.NewTaskPlanModel()
	var query string
	if req.TaskType != "" {
		query += fmt.Sprintf(`task_type = '%s'`, req.TaskType)
	}
	if req.TaskGroup != "" {
		if query != "" {
			query += " and "
		}
		query += fmt.Sprintf(`task_group = '%s'`, req.TaskGroup)
	}
	if req.TaskName != "" {
		if query != "" {
			query += " and "
		}
		query += "task_name like '%" + req.TaskName + "%'"
	}
	total, err := model.GetTaskPlanTotal(query)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	pageNum := (req.PageNum - 1) * req.PageSize
	result, err := model.GetTaskPlans(pageNum, req.PageSize, query)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}

	data := make([]*task.AddTask, 0)
	for _, value := range result {
		s, err := task.JsonToStruct(value)
		if err != nil {
			app.ErrorResp(context, e.ERROR, err.Error(), nil)
			return
		}
		if value != nil {
			data = append(data, s)
		}
	}

	app.SuccessResp(context, struct {
		Total int64           `json:"total"`
		Data  []*task.AddTask `json:"data"`
	}{
		Total: total,
		Data:  data,
	})
}

// GetPlanInfo 从数据库中查询单个计划详情
func GetPlanInfo(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	model := models.NewTaskPlanModel()
	exist, err := model.ExistTaskPlanByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "不存在的id", nil)
		return
	}

	result, err := model.GetTaskPlan(id)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	s, err := task.JsonToStruct(result)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}

	app.SuccessResp(context, s)
}

// RemovePlan 从数据库中删除单个计划
func RemovePlan(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	model := models.NewTaskPlanModel()
	exist, err := model.ExistTaskPlanByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "不存在的id", nil)
		return
	}

	result, err := model.GetTaskPlan(id) // 先查询信息 记录下 如果后面出了问题 再给恢复回去
	// 从定时器中删除该任务
	if result.IsCrontab == "yes" {
		_, err = task.CM.RemoveCronTaskByName(result.TaskName)
		if err != nil {
			app.ErrorResp(context, e.ERROR, err.Error(), nil)
			return
		}
	}

	err = model.DeleteTaskPlan(id)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	app.SuccessResp(context, nil)
}

// AddPlan 新增单个计划到数据库
func AddPlan(context *gin.Context) {
	req := context.MustGet(util.REQUEST_KEY).(*task.AddTask)
	model := models.NewTaskPlanModel()

	q, err := task.StructToJson(req)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}

	// 先在数据库中创建任务
	id, err := model.AddTaskPlan(&models.TaskPlanBase{
		TaskName:            q.TaskName,
		TaskType:            q.TaskType,
		TaskGroup:           q.TaskGroup,
		TaskConfig:          q.TaskConfig,
		TaskDataSourceLabel: q.TaskDataSourceLabel,
		TaskDataSource:      q.TaskDataSource,
		IsCrontab:           q.IsCrontab,
		CrontabString:       q.CrontabString,
	})
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	q.Id = id

	// 检测是否需要定时任务
	if req.IsCrontab == "yes" {
		job := task.InitTaskModel(req)
		_, err := task.CM.AddCronTask(req, job)
		if err != nil {
			app.ErrorResp(context, e.ERROR, err.Error(), nil)
			// 如果定时任务创建失败 将数据库中任务一并删除
			err = model.DeleteTaskPlan(id)
			return
		}
	}
	app.SuccessResp(context, q)
}

// UpdatePlan 修改单个计划到数据库
func UpdatePlan(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	req := context.MustGet(util.REQUEST_KEY).(*task.AddTask)
	model := models.NewTaskPlanModel()
	exist, err := model.ExistTaskPlanByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "不存在的id", nil)
		return
	}

	// 先查一下修改之前的配置数据
	beforeResult, err := model.GetTaskPlan(id)
	q, err := task.StructToJson(req)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}

	newInfo := &models.TaskPlanBase{
		TaskName:            req.TaskName,
		TaskType:            req.TaskType,
		TaskGroup:           req.TaskGroup,
		IsCrontab:           req.IsCrontab,
		TaskDataSourceLabel: req.TaskDataSourceLabel,
		CrontabString:       req.CrontabString,
	}

	if req.TaskConfig != nil {
		newInfo.TaskConfig = q.TaskConfig
	}
	if req.TaskDataSource != nil {
		newInfo.TaskDataSource = q.TaskDataSource
	}
	err = model.EditTaskPlan(id, newInfo)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}

	// 查一下修改之后的配置数据
	afterResult, err := model.GetTaskPlan(id)
	s, err := task.JsonToStruct(afterResult)

	// 改完数据库 再改定时器
	if beforeResult.IsCrontab == "yes" {
		_, err := task.CM.RemoveCronTaskByName(beforeResult.TaskName)
		if err != nil {
			app.ErrorResp(context, e.ERROR, err.Error(), nil)
			return
		}
	}
	if afterResult.IsCrontab == "yes" {
		job := task.InitTaskModel(s)
		_, err := task.CM.AddCronTask(s, job)
		if err != nil {
			app.ErrorResp(context, e.ERROR, err.Error(), nil)
			return
		}
	}

	app.SuccessResp(context, s)
}

// RunPlan 运行数据库中单个计划
func RunPlan(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	model := models.NewTaskPlanModel()
	exist, err := model.ExistTaskPlanByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "不存在的id", nil)
		return
	}

	result, err := model.GetTaskPlan(id)
	if err != nil || result == nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	s, err := task.JsonToStruct(result)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}

	// 后台执行测试 发起结果直接返回给前端
	go func() {
		job := task.InitTaskModel(s)
		var b = &task.BaseTask{T: &job}
		b.Run()
	}()
	app.SuccessResp(context, s)
}

func TerminatePlan(context *gin.Context) {
	req := context.MustGet(util.REQUEST_KEY).(*task.NameOfTask)
	success, value := task.TerminateMissionFlag(req.TaskName)
	if success {
		app.SuccessRespByCode(context, e.SUCCESS, value)
		return
	}

	app.SuccessRespByCode(context, e.ERROR, &task.TaskInfo{
		Name:    req.TaskName,
		Status:  64,
		Message: "任务停止失败!",
	})
}
