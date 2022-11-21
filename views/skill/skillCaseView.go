package skill

import (
	"encoding/json"
	"fmt"
	"github.com/gin-gonic/gin"
	"smartest-go/models"
	"smartest-go/pkg/app"
	"smartest-go/pkg/e"
	"smartest-go/pkg/logf"
	customTime "smartest-go/pkg/time"
	utilconst "smartest-go/pkg/util/const"
	"strconv"
	"strings"
	"time"
)

func ListSkill(context *gin.Context) {
	//BindAndValid
	copyReq := &models.SkillList{}
	err := app.BindAndValid(context, copyReq)
	if err != nil {
		app.ErrorResp(context, e.InvalidParams, err.Error(), nil)
		logf.Debug("Validation", err.Error())
		context.Abort()
		return
	}
	context.Set(utilconst.REQUEST_KEY, copyReq)

	//VerifyReq
	req := context.MustGet(utilconst.REQUEST_KEY).(*models.SkillList)
	model := models.NewSkillBaseTestModel()

	//extract request from param to mock query filter
	var filter map[string][]string
	jsonStr, _ := json.Marshal(req.Filter)
	err = json.Unmarshal(jsonStr, &filter)
	if err != nil {
		return
	}

	var filterString []string
	for f, v := range filter {
		if strings.Contains("id usetest case_version", f) {
			m := strings.Join(v, ",")
			filterString = append(filterString, f+" in ("+m+")")
		} else if strings.Contains("create_time", f) {
			if len(v) == 1 {
				filterString = append(filterString, f+" <= '"+v[0]+"'")
			}
			if len(v) == 2 {
				var start, end string
				l, _ := time.LoadLocation("Asia/Shanghai")
				startTime, _ := time.ParseInLocation("2006-01-02 15:05:06", v[0], l)
				endTime, _ := time.ParseInLocation("2006-01-02 15:05:06", v[1], l)
				if endTime.After(startTime) {
					start, end = v[0], v[1]
				} else {
					start, end = v[1], v[0]
				}
				filterString = append(filterString, f+" between '"+start+"' and '"+end+"'")
			}
			if len(v) >= 3 {
				m := strings.Join(v, "','")
				filterString = append(filterString, f+" in ('"+m+"')")
			}
		} else {
			m := strings.Join(v, "','")
			filterString = append(filterString, f+" in ('"+m+"')")
		}
	}

	queryFilter := strings.Join(filterString, " and ")

	//get total number of filtered data
	total, _ := model.GetSkillBaseTestTotal(queryFilter)
	pageNum := (req.PageNum - 1) * req.PageSize
	result, _ := model.GetSkillBaseTests(pageNum, req.PageSize, queryFilter)

	app.SuccessResp(context, struct {
		Count int64                   `json:"count"`
		Data  []*models.SkillBaseTest `json:"data"`
	}{
		Count: total,
		Data:  result,
	})
}

func DetailSkill(context *gin.Context) {
	id, _ := strconv.ParseInt(context.Param("id"), 10, 64)
	model := models.NewSkillBaseTestModel()
	exist, err := model.ExistSkillBaseTestByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "id not exist!", nil)
		return
	}
	result, _ := model.GetSkillBaseTest(id)

	app.SuccessResp(context, struct {
		Count int64                 `json:"count"`
		Data  *models.SkillBaseTest `json:"data"`
	}{
		Count: 1,
		Data:  result,
	})
}

func AddSkill(context *gin.Context) {
	//BindAndValid
	copyReq := &models.SkillBaseTest{}
	err := app.BindAndValid(context, copyReq)
	if err != nil {
		app.ErrorResp(context, e.InvalidParams, err.Error(), nil)
		logf.Debug("Validation", err.Error())
		context.Abort()
		return
	}
	context.Set(utilconst.REQUEST_KEY, copyReq)

	//VerifyReq
	req := context.MustGet(utilconst.REQUEST_KEY).(*models.SkillBaseTest)
	model := models.NewSkillBaseTestModel()
	id, err := model.AddSkillBaseTest(req)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}

	result, err := model.GetSkillBaseTest(id)
	app.SuccessResp(context, struct {
		Count int64                 `json:"count"`
		Data  *models.SkillBaseTest `json:"data"`
	}{
		Count: 1,
		Data:  result,
	})
}

func UpdateSkill(context *gin.Context) {
	//BindAndValid
	copyReq := &models.SkillBaseTest{}
	err := app.BindAndValid(context, copyReq)
	if err != nil {
		app.ErrorResp(context, e.InvalidParams, err.Error(), nil)
		logf.Debug("Validation", err.Error())
		context.Abort()
		return
	}
	context.Set(utilconst.REQUEST_KEY, copyReq)

	//VerifyReq
	id, _ := strconv.ParseInt(context.Param("id"), 10, 64)

	req := context.MustGet(utilconst.REQUEST_KEY).(*models.SkillBaseTest)
	model := models.NewSkillBaseTestModel()
	exist, err := model.ExistSkillBaseTestByID(id)
	if err != nil || !exist {
		app.ErrorResp(context, e.ERROR, "id not exist!", nil)
		return
	}
	err = model.EditSkillBaseTest(id, req)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}

	result, err := model.GetSkillBaseTest(id)
	app.SuccessResp(context, struct {
		Count int64                 `json:"count"`
		Data  *models.SkillBaseTest `json:"data"`
	}{
		Count: 1,
		Data:  result,
	})
}

func RemoveSkill(context *gin.Context) {
	id, _ := strconv.ParseInt(context.Param("id"), 10, 64)
	model := models.NewSkillBaseTestModel()
	result, err := model.GetSkillBaseTest(id)
	if result == nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	//soft delete
	result.UseTest = 2
	err = model.EditSkillBaseTest(id, result)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}

	app.SuccessResp(context, struct {
		Count int64                 `json:"count"`
		Data  *models.SkillBaseTest `json:"data"`
	}{
		Count: 1,
		Data:  result,
	})
}

func ImportSkill(context *gin.Context) {
	//BindAndValid
	copyReq := &models.Import{}
	err := app.BindAndValid(context, copyReq)
	if err != nil {
		app.ErrorResp(context, e.InvalidParams, err.Error(), nil)
		logf.Debug("Validation", err.Error())
		context.Abort()
		return
	}
	context.Set(utilconst.REQUEST_KEY, copyReq)

	//VerifyReq
	req := context.MustGet(utilconst.REQUEST_KEY).(*models.Import)
	model := models.NewSkillBaseTestModel()
	model.ExcelToDB("./upload/"+req.FileName, req.SheetName)
	app.SuccessResp(context, nil)
}

func GetSkillCaseCountByColumn(context *gin.Context) {
	column := context.Param("column")
	sql := fmt.Sprintf(`SELECT %s,count(1) count FROM %s GROUP BY %s ORDER BY count DESC`, column, models.SkillCaseTableName, column)

	skill := models.NewSkillBaseTestModel()
	res, _ := skill.GetGroupSkillBaseTest(sql)

	var groupCount []map[string]interface{}
	for _, r := range res {
		groupCount = append(groupCount, map[string]interface{}{
			"name":  r.SkillCn,
			"value": r.Count,
		})
	}

	app.SuccessResp(context, struct {
		Count int                      `json:"count"`
		Data  []map[string]interface{} `json:"data"`
	}{
		Count: len(groupCount),
		Data:  groupCount,
	})
}

func GetSkillCaseCountByWeek(context *gin.Context) {
	monthStr := context.Query("month")
	month, _ := strconv.Atoi(monthStr)
	weeklyData := customTime.WeeklyTime(month)

	// reverse slice
	for i, j := 0, len(weeklyData)-1; i < j; i, j = i+1, j-1 {
		weeklyData[i], weeklyData[j] = weeklyData[j], weeklyData[i]
	}

	skill := models.NewSkillBaseTestModel()
	var weekCount []map[string]interface{}
	for _, week := range weeklyData {
		newSkillTotal, _ := skill.GetSkillBaseTestTotal("create_time between ? and ?", week.StartTime, week.EndTime)
		lastSkillTotal, _ := skill.GetSkillBaseTestTotal("create_time <= ?", week.StartTime)
		weekCount = append(weekCount, map[string]interface{}{
			"week_time":  week.WeekTh,
			"old_count":  lastSkillTotal,
			"new_count":  newSkillTotal,
			"start_time": week.StartTime,
			"end_time":   week.EndTime,
		})
	}
	app.SuccessResp(context, struct {
		Count int                      `json:"count"`
		Data  []map[string]interface{} `json:"data"`
	}{
		Count: len(weekCount),
		Data:  weekCount,
	})
}
