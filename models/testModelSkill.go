package models

import (
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/jinzhu/gorm"
	"smartest-go/pkg/util"
	"strconv"
	"time"
)

const SkillCaseTableName = "skill_base_test"

type SkillBaseTest struct {
	Id          int64         `form:"id,omitempty"         json:"id"     gorm:"primary_key"    gorm:"column:id"`
	Question    string        `form:"question,omitempty"   json:"question"    gorm:"column:question"`
	Source      string        `form:"source,omitempty"     json:"source"      gorm:"column:source"`
	Domain      string        `form:"domain,omitempty"     json:"domain"      gorm:"column:domain"`
	Intent      string        `form:"intent,omitempty"     json:"intent"      gorm:"column:intent"`
	SkillSource string        `form:"skill_source,omitempty"   json:"skill_source"   gorm:"column:skill_source"`
	SkillCn     string        `form:"skill_cn,omitempty"       json:"skill_cn"       gorm:"column:skill_cn"`
	RobotType   string        `form:"robot_type,omitempty"     json:"robot_type"     gorm:"column:robot_type"`
	ActionName  string        `form:"action_name,omitempty"    json:"action_name"    gorm:"column:action_name"`
	Params      string        `form:"params,omitempty"         json:"params"         gorm:"column:params"`
	RobotId     string        `form:"robot_id,omitempty"       json:"robot_id"       gorm:"column:robot_id"`
	UseTest     int           `form:"usetest,omitempty"        json:"usetest"        gorm:"column:usetest"`
	IsSmoke     int           `form:"is_smoke,omitempty"        json:"is_smoke"        gorm:"column:is_smoke"`
	CreateTime  util.JSONTime `form:"create_time,omitempty"    json:"create_time"    gorm:"column:create_time"`
	UpdateTime  util.JSONTime `form:"update_time,omitempty"    json:"update_time"    gorm:"column:update_time"`
	ParamInfo   string        `form:"paraminfo,omitempty"      json:"paraminfo"      gorm:"column:paraminfo"`
	CaseVersion float32       `form:"case_version,omitempty"   json:"case_version"   gorm:"column:case_version"`
	EditLogs    string        `form:"edit_logs,omitempty"      json:"edit_logs"      gorm:"column:edit_logs"`

	Session *Session `json:"-" gorm:"-"`
}

type SkillList struct {
	PageNum  int `form:"pagenum,default=1" json:"pagenum"`
	PageSize int `form:"pagesize,default=30" json:"pagesize"`
	Filter   struct {
		Id          []string `form:"id,omitempty"         json:"id,omitempty"     gorm:"primary_key"    gorm:"column:id"`
		Question    []string `form:"question,omitempty"   json:"question,omitempty"    gorm:"column:question"`
		Source      []string `form:"source,omitempty"     json:"source,omitempty"      gorm:"column:source"`
		Domain      []string `form:"domain,omitempty"     json:"domain,omitempty"      gorm:"column:domain"`
		Intent      []string `form:"intent,omitempty"     json:"intent,omitempty"      gorm:"column:intent"`
		RobotType   []string `form:"robot_type,omitempty"     json:"robot_type,omitempty"     gorm:"column:robot_type"`
		ActionName  []string `form:"action_name,omitempty"    json:"action_name,omitempty"    gorm:"column:action_name"`
		Params      []string `form:"params,omitempty"         json:"params,omitempty"         gorm:"column:params"`
		RobotId     []string `form:"robot_id,omitempty"       json:"robot_id,omitempty"       gorm:"column:robot_id"`
		UseTest     []string `form:"usetest,omitempty"        json:"usetest,omitempty"        gorm:"column:usetest"`
		CreateTime  []string `form:"create_time,omitempty"    json:"create_time,omitempty"    gorm:"create_time,omitempty"`
		ParamInfo   []string `form:"paraminfo,omitempty"      json:"paraminfo,omitempty"      gorm:"column:paraminfo"`
		CaseVersion []string `form:"case_version,omitempty"   json:"case_version,omitempty"   gorm:"column:case_version"`
	} `json:"filter,omitempty"`
}

type Import struct {
	FileName  string `form:"file_name,omitempty" json:"file_name,omitempty"`
	SheetName string `form:"sheet_name,omitempty" json:"sheet_name,omitempty"`
}

type SkillGroupResult struct {
	SkillCn string
	Count   int
}

func (SkillBaseTest) TableName() string {
	return SkillCaseTableName
}

func NewSkillBaseTestModel() *SkillBaseTest {
	return &SkillBaseTest{Session: NewSession()}
}

// ExistSkillBaseTestByID checks if an SkillBaseTest exists based on ID
func (a *SkillBaseTest) ExistSkillBaseTestByID(id int64) (bool, error) {
	s := &SkillBaseTest{}
	err := a.Session.db.Select("id").Where("id = ? ", id).First(s).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}

	if s.Id > 0 {
		return true, nil
	}

	return false, nil
}

// GetSkillBaseTestTotal gets the total number of skill_base_tests based on the constraints
func (a *SkillBaseTest) GetSkillBaseTestTotal(query interface{}, args ...interface{}) (int64, error) {
	var count int64
	err := a.Session.db.Model(&SkillBaseTest{}).Where(query, args...).Count(&count).Error
	return count, err
}

// GetSkillBaseTests gets a list of skill_base_tests based on paging constraints
func (a *SkillBaseTest) GetSkillBaseTests(pageNum int, pageSize int, maps interface{}, args ...interface{}) ([]*SkillBaseTest, error) {
	var s []*SkillBaseTest
	err := a.Session.db.Where(maps, args...).Offset(pageNum).Limit(pageSize).Find(&s).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return s, nil
}

// GetSkillBaseTest Get a single skill_base_test based on ID
func (a *SkillBaseTest) GetSkillBaseTest(id int64) (*SkillBaseTest, error) {
	s := &SkillBaseTest{}
	err := a.Session.db.Where("id = ?", id).First(s).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	} else if err != nil && err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return s, nil
}

// EditSkillBaseTest modify a single skill_base_test
func (a *SkillBaseTest) EditSkillBaseTest(id int64, s *SkillBaseTest) error {
	tx := GetSessionTx(a.Session)
	return tx.Model(&SkillBaseTest{}).Where("id = ?", id).Updates(s).Error
}

// AddSkillBaseTest add a single skill_base_test
func (a *SkillBaseTest) AddSkillBaseTest(s *SkillBaseTest) (int64, error) {
	tx := GetSessionTx(a.Session)
	err := tx.Create(s).Error
	if err != nil {
		return 0, err
	}
	return s.Id, nil
}

// DeleteSkillBaseTest delete a single skill_base_test
func (a *SkillBaseTest) DeleteSkillBaseTest(id int64) error {
	tx := GetSessionTx(a.Session)
	return tx.Where("id = ?", id).Delete(SkillBaseTest{}).Error
}

// GetGroupSkillBaseTest group by result based on constrains
func (a *SkillBaseTest) GetGroupSkillBaseTest(sql string, values ...interface{}) ([]*SkillGroupResult, error) {
	tx := GetSessionTx(a.Session)
	var res []*SkillGroupResult
	result, err := tx.Raw(sql, values...).Rows()
	if err != nil {
		return nil, err
	}
	for result.Next() {
		result.Scan()
		var col = &SkillGroupResult{}
		result.Scan(&col.SkillCn, &col.Count)
		res = append(res, col)
	}
	return res, nil
}

// ExcelToDB import excel data to mysql database
func (a *SkillBaseTest) ExcelToDB(filename, sheet string) {
	f, err := excelize.OpenFile(filename)
	if err != nil {
		return
	}
	tx := GetSessionTx(a.Session)
	nowTime := util.JSONTime{Time: time.Now()}
	rows := f.GetRows(sheet)

	tableHeader := make(map[int]string)
	for index, row := range rows {
		if index == 0 {
			// 记录表头
			for i, cellValue := range row {
				tableHeader[i] = cellValue
			}
			continue
		}
		tmpReq := &SkillBaseTest{}
		for i, cellValue := range row {
			// 记录表数据
			if tableHeader[i] == "id" {
				num, _ := strconv.Atoi(cellValue)
				tmpReq.Id = int64(num)
			}
			if tableHeader[i] == "question" {
				tmpReq.Question = cellValue
			}
			if tableHeader[i] == "source" {
				tmpReq.Source = cellValue
			}

			if tableHeader[i] == "domain" {
				tmpReq.Domain = cellValue
			}

			if tableHeader[i] == "intent" {
				tmpReq.Intent = cellValue
			}

			if tableHeader[i] == "skill_source" {
				tmpReq.SkillSource = cellValue
			}

			if tableHeader[i] == "skill_cn" {
				tmpReq.SkillCn = cellValue
			}

			if tableHeader[i] == "robot_id" {
				tmpReq.RobotId = cellValue
			}

			if tableHeader[i] == "usetest" {
				num, _ := strconv.Atoi(cellValue)
				tmpReq.UseTest = num
			}

			if tableHeader[i] == "is_smoke" {
				num, _ := strconv.Atoi(cellValue)
				tmpReq.IsSmoke = num
			}

			if tableHeader[i] == "paraminfo" {
				tmpReq.ParamInfo = cellValue
			}

			if tableHeader[i] == "case_version" {
				num64, _ := strconv.ParseFloat(cellValue, 32)
				tmpReq.CaseVersion = float32(num64)
			}

			if tableHeader[i] == "robot_type" {
				tmpReq.RobotType = cellValue
			}
			tmpReq.CreateTime = nowTime
			tmpReq.UpdateTime = nowTime
		}
		tx.Create(tmpReq)
	}
}
