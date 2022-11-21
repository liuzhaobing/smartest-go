package task

import (
	"encoding/json"
	"github.com/gin-gonic/gin"
	"smartest-go/models"
	"smartest-go/pkg/app"
	"smartest-go/pkg/e"
	util "smartest-go/pkg/util/const"
	"strconv"
)

/*
开放给前端的测试计划管理结构 同时提供给内部使用
*/

type ListTask struct {
	Id        int64  `json:"id" form:"id"`
	TaskName  string `json:"task_name" form:"task_name"`
	TaskType  string `json:"task_type" form:"task_type"`
	TaskGroup string `json:"task_group" form:"task_group"`
	IsCrontab bool   `json:"is_crontab" form:"is_crontab"`

	PageNum  int `form:"page_num,default=1" json:"page_num"`
	PageSize int `form:"page_size,default=30" json:"page_size"`
}

type NameOfTask struct {
	TaskName string `json:"task_name" form:"task_name"`
}

type AddTask struct {
	Id             int64           `json:"id,omitempty" form:"id,omitempty"`
	TaskName       string          `json:"task_name" form:"task_name"`
	TaskType       string          `json:"task_type" form:"task_type"`
	TaskGroup      string          `json:"task_group" form:"task_group"`
	IsCrontab      string          `json:"is_crontab" form:"is_crontab"`
	CrontabString  string          `json:"crontab_string" form:"crontab_string"`
	TaskConfig     *TestConfig     `json:"task_config" form:"task_config"`
	TaskDataSource *TestDataSource `json:"task_data_source" form:"task_data_source"`
}

type TestConfig struct {
	TestConfigKG *KGTaskConfig `json:"config_kg,omitempty" form:"config_kg,omitempty"` // 知识图谱
}

type TestDataSource struct {
	TestCaseKG   []*KGTaskReq  `json:"cases_kg,omitempty" form:"cases_kg,omitempty"`   // 知识图谱用例数据
	KGDataSource *KGDataSource `json:"source_kg,omitempty" form:"source_kg,omitempty"` // 知识图谱用例构造
}

func JsonToStruct(j *models.TaskPlanBase) (*AddTask, error) {
	// 根据类型 将数据库string类型转换为前端可识别struct
	s := &AddTask{
		Id:            j.Id,
		TaskName:      j.TaskName,
		TaskType:      j.TaskType,
		TaskGroup:     j.TaskGroup,
		IsCrontab:     j.IsCrontab,
		CrontabString: j.CrontabString,
	}

	switch j.TaskType {
	case KnowledgeGraph:
		s.TaskConfig = &TestConfig{TestConfigKG: &KGTaskConfig{}}
		err := json.Unmarshal([]byte(j.TaskConfig), s.TaskConfig)
		if err != nil {
			return nil, err
		}

		s.TaskDataSource = &TestDataSource{TestCaseKG: make([]*KGTaskReq, 0)}
		err = json.Unmarshal([]byte(j.TaskDataSource), &s.TaskDataSource)
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

func StructToJson(s *AddTask) (*models.TaskPlanBase, error) {
	// 根据类型 将前端可识别struct类型转换为数据库string
	j := &models.TaskPlanBase{
		TaskName:      s.TaskName,
		TaskType:      s.TaskType,
		TaskGroup:     s.TaskGroup,
		IsCrontab:     s.IsCrontab,
		CrontabString: s.CrontabString,
	}

	switch s.TaskType {
	case KnowledgeGraph:
		kgConfig, err := json.Marshal(s.TaskConfig)
		if err != nil {
			return nil, err
		}
		j.TaskConfig = string(kgConfig)
		kgData, err := json.Marshal(s.TaskDataSource)
		if err != nil {
			return nil, err
		}
		j.TaskDataSource = string(kgData)
	}
	return j, nil
}

// ListPlan 从数据库中查询所有计划列表
func ListPlan(context *gin.Context) {
	req := context.MustGet(util.REQUEST_KEY).(*ListTask)
	model := models.NewTaskPlanModel()
	total, err := model.GetTaskPlanTotal("1=1")
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}

	pageNum := (req.PageNum - 1) * req.PageSize
	result, err := model.GetTaskPlans(pageNum, req.PageSize, "1=1")
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}

	data := make([]*AddTask, 0)
	for _, value := range result {
		s, err := JsonToStruct(value)
		if err != nil {
			app.ErrorResp(context, e.ERROR, err.Error(), nil)
			return
		}
		if value != nil {
			data = append(data, s)
		}
	}

	app.SuccessResp(context, struct {
		Total int64      `json:"total"`
		Data  []*AddTask `json:"data"`
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
	s, err := JsonToStruct(result)
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
		_, err = CM.RemoveCronTaskByName(result.TaskName)
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
	req := context.MustGet(util.REQUEST_KEY).(*AddTask)
	model := models.NewTaskPlanModel()

	q, err := StructToJson(req)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}

	// 先在数据库中创建任务
	id, err := model.AddTaskPlan(&models.TaskPlanBase{
		TaskName:       q.TaskName,
		TaskType:       q.TaskType,
		TaskGroup:      q.TaskGroup,
		TaskConfig:     q.TaskConfig,
		TaskDataSource: q.TaskDataSource,
		IsCrontab:      q.IsCrontab,
		CrontabString:  q.CrontabString,
	})
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	q.Id = id

	// 检测是否需要定时任务
	if req.IsCrontab == "yes" {
		job := InitTaskModel(req)
		_, err := CM.AddCronTask(req, job)
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
	req := context.MustGet(util.REQUEST_KEY).(*AddTask)
	model := models.NewTaskPlanModel()
	exist, err := model.ExistTaskPlanByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "不存在的id", nil)
		return
	}

	// 先查一下修改之前的配置数据
	beforeResult, err := model.GetTaskPlan(id)
	q, err := StructToJson(req)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}

	newInfo := &models.TaskPlanBase{
		TaskName:      req.TaskName,
		TaskType:      req.TaskType,
		TaskGroup:     req.TaskGroup,
		IsCrontab:     req.IsCrontab,
		CrontabString: req.CrontabString,
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
	s, err := JsonToStruct(afterResult)

	// 改完数据库 再改定时器
	if beforeResult.IsCrontab == "yes" {
		_, err := CM.RemoveCronTaskByName(beforeResult.TaskName)
		if err != nil {
			app.ErrorResp(context, e.ERROR, err.Error(), nil)
			return
		}
	}
	if afterResult.IsCrontab == "yes" {
		job := InitTaskModel(s)
		_, err := CM.AddCronTask(s, job)
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
	s, err := JsonToStruct(result)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}

	// 后台执行测试 发起结果直接返回给前端
	go func() {
		job := InitTaskModel(s)
		var b = &BaseTask{T: &job}
		b.Run()
	}()
	app.SuccessResp(context, s)
}

func TerminatePlan(context *gin.Context) {
	req := context.MustGet(util.REQUEST_KEY).(*NameOfTask)
	success, value := TerminateMissionFlag(req.TaskName)
	if success {
		app.SuccessRespByCode(context, e.SUCCESS, value)
		return
	}

	app.SuccessRespByCode(context, e.ERROR, &TaskInfo{
		Name:    req.TaskName,
		Status:  64,
		Message: "任务停止失败!",
	})
}

func InitTaskModel(config *AddTask) TaskModel {
	switch config.TaskType {
	case KnowledgeGraph:
		var kg TaskModel = &KGTask{}
		kgConfig := config.TaskConfig.TestConfigKG
		kgReqs := config.TaskDataSource.TestCaseKG
		kgDataSource := config.TaskDataSource.KGDataSource
		if len(kgReqs) != 0 {
			kg = &KGTaskTest{KGTask: NewKGTask(kgConfig, kgReqs, nil)}
		} else if kgDataSource != nil {
			kg = &KGTaskTest{KGTask: NewKGTask(kgConfig, make([]*KGTaskReq, 0), kgDataSource)}
		}

		return kg
	}
	return nil
}

var (
	KnowledgeGraph = "kg"
)
