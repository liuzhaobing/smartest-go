package models

import (
	"github.com/jinzhu/gorm"
)

type TaskPlanBase struct {
	Id                  int64  `json:"id"  gorm:"primary_key"    gorm:"column:id"`
	TaskName            string `json:"task_name"   gorm:"column:task_name"`
	TaskType            string `json:"task_type"   gorm:"column:task_type"`
	TaskGroup           string `json:"task_group"   gorm:"column:task_group"`
	TaskConfig          string `json:"task_config"     gorm:"column:task_config"`
	TaskDataSource      string `json:"task_data_source" gorm:"column:task_data_source"`
	TaskDataSourceLabel string `json:"task_data_source_label" gorm:"column:task_data_source_label"`
	IsCrontab           string `json:"is_crontab"   gorm:"column:is_crontab"` // yes / no
	CrontabString       string `json:"crontab_string"  gorm:"column:crontab_string"`

	Session *Session `json:"-" gorm:"-"`
}

func (TaskPlanBase) TableName() string {
	return "plans"
}

func NewTaskPlanModel() *TaskPlanBase {
	return &TaskPlanBase{Session: NewSession()}
}

func (a *TaskPlanBase) ExistTaskPlanByID(id int64) (bool, error) {
	plans := &TaskPlanBase{}
	err := a.Session.db.Select("id").Where("id = ? ", id).First(plans).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}
	if plans.Id > 0 {
		return true, nil
	}
	return false, nil
}

func (a *TaskPlanBase) GetTaskPlanTotal(query interface{}, args ...interface{}) (int64, error) {
	var count int64
	err := a.Session.db.Model(&TaskPlanBase{}).Where(query, args...).Count(&count).Error
	return count, err
}

func (a *TaskPlanBase) GetTaskPlans(pageNum int, pageSize int, maps interface{}, args ...interface{}) ([]*TaskPlanBase, error) {
	var plans []*TaskPlanBase
	err := a.Session.db.Where(maps, args...).Offset(pageNum).Limit(pageSize).Find(&plans).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return plans, nil
}

func (a *TaskPlanBase) GetTaskPlan(id int64) (*TaskPlanBase, error) {
	plan := &TaskPlanBase{}
	err := a.Session.db.Where("id = ?", id).First(plan).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	} else if err != nil && err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return plan, nil
}

func (a *TaskPlanBase) EditTaskPlan(id int64, data *TaskPlanBase) error {
	tx := GetSessionTx(a.Session)
	return tx.Model(&TaskPlanBase{}).Where("id = ?", id).Updates(data).Error
}

func (a *TaskPlanBase) AddTaskPlan(plan *TaskPlanBase) (int64, error) {
	tx := GetSessionTx(a.Session)
	err := tx.Create(plan).Error
	if err != nil {
		return 0, err
	}
	return plan.Id, nil
}

func (a *TaskPlanBase) DeleteTaskPlan(id int64) error {
	tx := GetSessionTx(a.Session)
	return tx.Where("id = ?", id).Delete(TaskPlanBase{}).Error
}

// GetGroupTaskPlan group by result based on constrains
func (a *TaskPlanBase) GetGroupTaskPlan(sql string, values ...interface{}) ([]*SkillGroupResult, error) {
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
