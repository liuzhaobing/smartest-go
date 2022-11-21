package main

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"net/http"
	"smartest-go/models"
	"smartest-go/pkg/logf"
	"smartest-go/pkg/setting"
	"smartest-go/routers"
	"smartest-go/task"
)

func init() {
	setting.Setup()
	logf.Setup()
	models.MongoSetup()
	models.Setup()
	initPlansFromDataBase()
}

func main() {
	gin.SetMode(setting.ServerSetting.RunMode)
	r := routers.InitRouter()
	endPoint := fmt.Sprintf(":%d", setting.ServerSetting.HttpPort)
	maxHeaderBytes := 1 << 20

	server := &http.Server{
		Addr:           endPoint,
		Handler:        r,
		MaxHeaderBytes: maxHeaderBytes,
	}
	log.Printf("[info] start httpweb server listening %s", endPoint)
	server.ListenAndServe()
}

// 从数据库中初始化所有需要定时任务的计划
func initPlansFromDataBase() {
	model := models.NewTaskPlanModel()
	total, _ := model.GetTaskPlanTotal("1=1")
	result, _ := model.GetTaskPlans(0, int(total), "is_crontab='yes'")
	for _, t := range result {
		a, err := task.JsonToStruct(t)
		if err != nil {
			continue
		}
		job := task.InitTaskModel(a)
		_, err = task.CM.AddCronTask(a, job)
		if err != nil {
			continue
		}
	}
}
