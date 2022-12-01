package v1

import (
	"github.com/gin-gonic/gin"
	"smartest-go/models"
	"smartest-go/pkg/app"
	"smartest-go/pkg/e"
	util "smartest-go/pkg/util/const"
	"strconv"
)

type ListServer struct {
	ServerTypes string `json:"types"   gorm:"column:types"  form:"types"`
}

type AddServer struct {
	ServerName    string `json:"name"   form:"column:name"`
	ServerTypes   string `json:"types"   form:"column:types"`
	ServerAddress string `json:"address"   form:"column:address"`
}

func ListServers(context *gin.Context) {
	req := context.MustGet(util.REQUEST_KEY).(*ListServer)
	model := models.NewTaskServerModel()
	query := "1=1"
	if req.ServerTypes != "" {
		query = "types like '%" + req.ServerTypes + "%'"
	}
	total, err := model.GetTaskServerTotal(query)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	result, err := model.GetTaskServers(0, 100, query)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	app.SuccessResp(context, struct {
		Total int64                    `json:"total"`
		Data  []*models.TaskServerBase `json:"data"`
	}{
		Total: total,
		Data:  result,
	})
}

func RemoveServers(context *gin.Context) {
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	model := models.NewTaskServerModel()
	exist, err := model.ExistTaskServerByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "不存在的id", nil)
		return
	}
	err = model.DeleteTaskServer(id)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	app.SuccessResp(context, nil)
}

func AddServers(context *gin.Context) {
	req := context.MustGet(util.REQUEST_KEY).(*AddServer)
	model := models.NewTaskServerModel()
	id, err := model.AddTaskServer(&models.TaskServerBase{
		ServerName:    req.ServerName,
		ServerTypes:   req.ServerTypes,
		ServerAddress: req.ServerAddress,
	})
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	app.SuccessResp(context, id)
}

func UpdateServers(context *gin.Context) {
	req := context.MustGet(util.REQUEST_KEY).(*AddServer)
	id, err := strconv.ParseInt(context.Param("id"), 10, 64)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	model := models.NewTaskServerModel()
	exist, err := model.ExistTaskServerByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "不存在的id", nil)
		return
	}

	err = model.EditTaskServer(id, &models.TaskServerBase{
		ServerName:    req.ServerName,
		ServerTypes:   req.ServerTypes,
		ServerAddress: req.ServerAddress,
	})
	if err != nil {
		app.ErrorResp(context, e.ERROR, "修改失败！", nil)
		return
	}
	app.SuccessResp(context, id)
}
