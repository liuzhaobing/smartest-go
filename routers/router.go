package routers

import (
	"bytes"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"io/ioutil"
	"net/http"
	"os"
	"reflect"
	"smartest-go/models"
	"smartest-go/pkg/app"
	"smartest-go/pkg/e"
	pkgfile "smartest-go/pkg/file"
	"smartest-go/pkg/logf"
	util "smartest-go/pkg/util/const"
	"smartest-go/request/api/v1"
	"smartest-go/task"
	"strings"
	"sync"
	"time"
)

const (
	skill = "skill"
)

// InitRouter initialize routing information
func InitRouter() *gin.Engine {
	r := gin.New()
	//日志中间件,所有的异常捕获
	r.Use(gin.Recovery(), initLog, cors())
	user := r.Group("/user")
	user.Use()
	{
		user.POST("/login", Validation(&v1.LoginPayload{}), v1.Login)
		user.POST("/logout", v1.LogOut)
		user.GET("/info", Validation(&v1.AccessToken{}), v1.GetUserInfo)
	}
	apiV1 := r.Group("/api/v1")
	apiV1.GET("/download", func(c *gin.Context) {
		fileName := c.Query("filename")
		if fileName == "" {
			c.Redirect(http.StatusFound, "/404")
			return
		}
		//打开文件
		f, errByOpenFile := os.Open(fileName)
		defer func(f *os.File) {
			err := f.Close()
			if err != nil {

			}
		}(f)
		//非空处理
		if errByOpenFile != nil {
			c.Redirect(http.StatusFound, "/404")
			return
		}
		list := strings.Split(fileName, "/")
		c.Header("Content-Type", "application/octet-stream")
		c.Header("Content-Disposition", "attachment; filename="+list[len(list)-1])
		c.Header("Content-Transfer-Encoding", "binary")
		c.File(fileName)
		return
	})
	apiV1.POST("/upload", func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			app.ErrorResp(c, 500, err.Error(), nil)
			return
		}
		filename := time.Now().Format("20060102-15-04-05") + file.Filename
		// 上传文件至指定目录
		var once sync.Once
		once.Do(func() {
			dir, err := os.Getwd()
			if err != nil {
				logf.Error(err)
			}
			src := dir + "/upload/"
			err = pkgfile.IsNotExistMkDir(src)
			if err != nil {
				return
			}
		})
		if err := c.SaveUploadedFile(file, "./upload/"+filename); err != nil {
			fmt.Println(err)
		}
		app.SuccessResp(c, filename)
	})
	apiV1.POST("/export", models.ExportExcel)
	apiV1.GET("/files", Validation(&v1.DirPath{}), v1.GetFileList)
	apiPlan := apiV1.Group("/plan")
	apiPlan.Use()
	{
		apiPlan.GET("", Validation(&task.ListTask{}), v1.ListPlan)
		apiPlan.POST("", Validation(&task.AddTask{}), v1.AddPlan)
		apiPlan.PUT("/:id", Validation(&task.AddTask{}), v1.UpdatePlan)
		apiPlan.DELETE("/:id", v1.RemovePlan)
		apiPlan.GET("/:id", v1.GetPlanInfo)
		apiPlan.POST("/:id", v1.RunPlan)
		apiPlan.PUT("", Validation(&task.NameOfTask{}), v1.TerminatePlan)
	}
	apiV1.GET("/groups", v1.GetPlanGroups)

	apiHistory := apiV1.Group("/history")
	apiHistory.Use()
	{
		apiHistory.GET("", Validation(&task.TaskInfoSearch{}), task.LookUpStatus) // 获取测试执行历史
	}

	// 定时任务控制器
	apiCron := apiV1.Group("/crontab")
	apiCron.Use()
	{
		apiCron.GET("", v1.ListCronPlan)                                   // 获取定时任务列表
		apiCron.POST("", Validation(&task.AddTask{}), v1.AddCronPlan)      // 添加定时任务  不开放给前端
		apiCron.PUT("", Validation(&task.NameOfTask{}), v1.RemoveCronPlan) // 修改定时任务  不开放给前端
	}

	// 测试调试用的路由
	apiTask := apiV1.Group("/task")
	apiTask.Use()
	{
		apiTask.GET("", Validation(&task.NameOfTask{}), task.Status) // 获取测试执行状态  不开放给前端
		apiTask.POST("", Validation(&task.AddTask{}), task.Start)    // 发起测试任务  不开放给前端
		apiTask.PUT("", Validation(&task.NameOfTask{}), task.Stop)   // 终止测试执行  不开放给前端
	}

	// 测试飞书调试的路由
	apiFeiShu := apiV1.Group("/feishu")
	apiFeiShu.Use()
	{
		apiFeiShu.POST("", Validation(&task.FeiShu{}), task.SendReportToFeiShu)
	}

	// 测试报告数据展示
	apiReportV1 := apiV1.Group("/reports")
	apiReportV1.Use()
	{
		apiReportV1.PUT("", models.MongoUpdateFunc)
		apiReportV1.POST("", models.MongoListFuncFind)
		apiReportV1.POST("/aggregate", models.MongoListFuncAggregate)
		apiReportV1.POST("/export", models.MongoListAndExportFunc)
	}

	// 测试环境
	apiServerV1 := apiV1.Group("/server")
	apiServerV1.Use()
	{
		apiServerV1.GET("", Validation(&v1.ListServer{}), v1.ListServers)
		apiServerV1.PUT("/:id", Validation(&v1.AddServer{}), v1.UpdateServers)
		apiServerV1.DELETE("/:id", v1.RemoveServers)
		apiServerV1.POST("", Validation(&v1.AddServer{}), v1.AddServers)
	}

	// 测试数据
	apiDataV1 := apiV1.Group("/data")
	apiDataV1.Use()
	{
		apiDataV1.GET("", Validation(&v1.ListData{}), v1.ListDatas)
		apiDataV1.PUT("/:id", Validation(&v1.AddData{}), v1.UpdateDatas)
		apiDataV1.DELETE("/:id", v1.RemoveDatas)
		apiDataV1.POST("", Validation(&v1.AddData{}), v1.AddDatas)
	}

	// 测试用例管理
	apiCasesV1 := apiV1.Group("/cases")
	apiCasesV1.Use()
	{
		apiCasesV1.GET("/:type", func(context *gin.Context) {
			TaskType := context.Param("type")
			switch TaskType {
			case skill:
				v1.ListSkill(context)
			}
		})
		apiCasesV1.DELETE("/:type/:id", func(context *gin.Context) {
			TaskType := context.Param("type")
			switch TaskType {
			case skill:
				v1.RemoveSkill(context)
			}
		})
		apiCasesV1.GET("/:type/:id", func(context *gin.Context) {
			TaskType := context.Param("type")
			switch TaskType {
			case skill:
				v1.DetailSkill(context)
			}
		})
		apiCasesV1.POST("/:type", func(context *gin.Context) {
			TaskType := context.Param("type")
			switch TaskType {
			case skill:
				v1.AddSkill(context)
			}
		})
		apiCasesV1.PUT("/:type/:id", func(context *gin.Context) {
			TaskType := context.Param("type")
			switch TaskType {
			case skill:
				v1.UpdateSkill(context)
			}
		})
		apiCasesV1.POST("/:type/import/excel", func(context *gin.Context) {
			TaskType := context.Param("type")
			switch TaskType {
			case skill:
				v1.ImportSkill(context)
			}
		})
		apiCasesV1.GET("/:type/count/:column", func(context *gin.Context) {
			TaskType := context.Param("type")
			switch TaskType {
			case skill:
				v1.GetSkillCaseCountByColumn(context)
			}
		})
		apiCasesV1.GET("/:type/total/weekly", func(context *gin.Context) {
			TaskType := context.Param("type")
			switch TaskType {
			case skill:
				v1.GetSkillCaseCountByWeek(context)
			}
		})
	}

	return r
}

type Interface interface {
	DeepCopy() interface{}
}

func Copy(src interface{}) interface{} {
	if src == nil {
		return nil
	}
	original := reflect.ValueOf(src)
	cpy := reflect.New(original.Type()).Elem()
	copyRecursive(original, cpy)

	return cpy.Interface()
}

func copyRecursive(src, dst reflect.Value) {
	if src.CanInterface() {
		if copier, ok := src.Interface().(Interface); ok {
			dst.Set(reflect.ValueOf(copier.DeepCopy()))
			return
		}
	}

	switch src.Kind() {
	case reflect.Ptr:
		originalValue := src.Elem()

		if !originalValue.IsValid() {
			return
		}
		dst.Set(reflect.New(originalValue.Type()))
		copyRecursive(originalValue, dst.Elem())

	case reflect.Interface:
		if src.IsNil() {
			return
		}
		originalValue := src.Elem()
		copyValue := reflect.New(originalValue.Type()).Elem()
		copyRecursive(originalValue, copyValue)
		dst.Set(copyValue)

	case reflect.Struct:
		t, ok := src.Interface().(time.Time)
		if ok {
			dst.Set(reflect.ValueOf(t))
			return
		}
		for i := 0; i < src.NumField(); i++ {
			if src.Type().Field(i).PkgPath != "" {
				continue
			}
			copyRecursive(src.Field(i), dst.Field(i))
		}

	case reflect.Slice:
		if src.IsNil() {
			return
		}
		dst.Set(reflect.MakeSlice(src.Type(), src.Len(), src.Cap()))
		for i := 0; i < src.Len(); i++ {
			copyRecursive(src.Index(i), dst.Index(i))
		}

	case reflect.Map:
		if src.IsNil() {
			return
		}
		dst.Set(reflect.MakeMap(src.Type()))
		for _, key := range src.MapKeys() {
			originalValue := src.MapIndex(key)
			copyValue := reflect.New(originalValue.Type()).Elem()
			copyRecursive(originalValue, copyValue)
			copyKey := Copy(key.Interface())
			dst.SetMapIndex(reflect.ValueOf(copyKey), copyValue)
		}

	default:
		dst.Set(src)
	}
}

// Validation 绑定签证的中间件
func Validation(req interface{}) gin.HandlerFunc {
	return func(c *gin.Context) {
		//深拷贝
		copyReq := Copy(req)
		err := app.BindAndValid(c, copyReq)
		if err != nil {
			app.ErrorResp(c, e.InvalidParams, err.Error(), nil)
			logf.Debug("Validation", err.Error())
			c.Abort()
			return
		}
		c.Set(util.REQUEST_KEY, copyReq)
	}
}

//跨域中间件
func cors() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		if origin != "" {
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
			c.Header("Access-Control-Allow-Headers", "Origin, X-Requested-With, Content-Type, Accept, Authorization, "+util.HeaderXToken)
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers, Cache-Control, Content-Language, Content-Type, "+util.HeaderXToken)
			c.Header("Access-Control-Allow-Credentials", "false")
			c.Set("content-type", "application/json")
		}
		if method == "OPTIONS" {
			c.AbortWithStatus(http.StatusNoContent)
		}
		c.Next()
	}
}

func initLog(c *gin.Context) {
	// 开始时间
	startTime := time.Now()
	// 请求路由
	path := c.Request.RequestURI

	// 排除文件上传的请求体打印
	isFormData := strings.Contains(c.Request.Header.Get("Content-Type"), "multipart/form-data")
	// requestBody
	var requestBody []byte
	if !isFormData {
		requestBody, _ = c.GetRawData()
		c.Request.Body = ioutil.NopCloser(bytes.NewBuffer(requestBody))
		c.Set("requestBody", string(requestBody))
	}

	//处理请求
	c.Next()
	// 处理结果
	result, exists := c.Get(util.LogResponse)
	if exists {
		result = result.(*app.Response)
	}

	// 执行时间
	latencyTime := time.Since(startTime)
	// 请求方式
	reqMethod := c.Request.Method
	// http状态码
	statusCode := c.Writer.Status()
	// 请求IP
	clientIP := c.ClientIP()
	//token := c.GetHeader(tool.HeaderToken)
	// 日志格式
	logf.InfoWithFields(logrus.Fields{
		"req_body":     string(requestBody),
		"http_code":    statusCode,
		"latency_time": fmt.Sprintf("%13v", latencyTime),
		"ip":           clientIP,
		"method":       reqMethod,
		"path":         path,
		"result":       result,
		"msg":          reqMethod,
	})
}
