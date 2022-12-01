package v1

import (
	"github.com/gin-gonic/gin"
	"smartest-go/models"
	"smartest-go/pkg/app"
	"smartest-go/pkg/e"
	util "smartest-go/pkg/util/const"
	"strconv"
)

type ListData struct {
	Types string `json:"types"   gorm:"column:types"  form:"types"`
}

type AddData struct {
	Name    string `json:"name"   form:"column:name"`
	Types   string `json:"types"   form:"column:types"`
	Headers string `json:"headers"   form:"column:headers"`
	Data    string `json:"data"   form:"column:data"`
}

func ListDatas(context *gin.Context) {
	req := context.MustGet(util.REQUEST_KEY).(*ListData)
	model := models.NewTaskDataModel()
	query := "1=1"
	if req.Types != "" {
		query = "types like '%" + req.Types + "%'"
	}
	total, err := model.GetTaskDataTotal(query)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	result, err := model.GetTaskDatas(0, 100, query)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	app.SuccessResp(context, struct {
		Total int64                  `json:"total"`
		Data  []*models.TaskDataBase `json:"data"`
	}{
		Total: total,
		Data:  result,
	})
}

func RemoveDatas(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	model := models.NewTaskDataModel()
	exist, err := model.ExistTaskDataByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "不存在的id", nil)
		return
	}
	err = model.DeleteTaskData(id)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	app.SuccessResp(context, nil)
}

func AddDatas(context *gin.Context) {
	req := context.MustGet(util.REQUEST_KEY).(*AddData)
	model := models.NewTaskDataModel()
	id, err := model.AddTaskData(&models.TaskDataBase{
		Name:    req.Name,
		Types:   req.Types,
		Headers: req.Headers,
		Data:    req.Data,
	})
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	app.SuccessResp(context, id)
}

func UpdateDatas(context *gin.Context) {
	req := context.MustGet(util.REQUEST_KEY).(*AddData)
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	model := models.NewTaskDataModel()
	exist, err := model.ExistTaskDataByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "不存在的id", nil)
		return
	}

	err = model.EditTaskData(id, &models.TaskDataBase{
		Name:    req.Name,
		Types:   req.Types,
		Headers: req.Headers,
		Data:    req.Data,
	})
	if err != nil {
		app.ErrorResp(context, e.ERROR, "修改失败！", nil)
		return
	}
	app.SuccessResp(context, id)
}
