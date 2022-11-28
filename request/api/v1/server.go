package v1

import (
	"github.com/gin-gonic/gin"
	"smartest-go/models"
	"smartest-go/pkg/app"
	"smartest-go/pkg/e"
	util "smartest-go/pkg/util/const"
)

type ListServer struct {
	ServerTypes string `json:"types"   gorm:"column:types"  form:"types"`
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
