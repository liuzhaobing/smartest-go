package models

import (
	"github.com/jinzhu/gorm"
)

/*
用于存储用户自定义的测试数据等
*/

type TaskDataBase struct {
	Id      int64  `json:"id"  gorm:"primary_key"    gorm:"column:id"`
	Name    string `json:"name"   gorm:"column:name"`
	Types   string `json:"types"   gorm:"column:types"`
	Headers string `json:"headers"   gorm:"column:headers"`
	Data    string `json:"data"   gorm:"column:data"`

	Session *Session `json:"-" gorm:"-"`
}

func (TaskDataBase) TableName() string {
	return "custom_data"
}

func NewTaskDataModel() *TaskDataBase {
	return &TaskDataBase{Session: NewSession()}
}

func (a *TaskDataBase) ExistTaskDataByID(id int64) (bool, error) {
	Datas := &TaskDataBase{}
	err := a.Session.db.Select("id").Where("id = ? ", id).First(Datas).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}
	if Datas.Id > 0 {
		return true, nil
	}
	return false, nil
}

func (a *TaskDataBase) GetTaskDataTotal(query interface{}, args ...interface{}) (int64, error) {
	var count int64
	err := a.Session.db.Model(&TaskDataBase{}).Where(query, args...).Count(&count).Error
	return count, err
}

func (a *TaskDataBase) GetTaskDatas(pageNum int, pageSize int, maps interface{}, args ...interface{}) ([]*TaskDataBase, error) {
	var Datas []*TaskDataBase
	err := a.Session.db.Where(maps, args...).Offset(pageNum).Limit(pageSize).Find(&Datas).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return Datas, nil
}

func (a *TaskDataBase) GetTaskData(id int64) (*TaskDataBase, error) {
	Data := &TaskDataBase{}
	err := a.Session.db.Where("id = ?", id).First(Data).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	} else if err != nil && err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return Data, nil
}

func (a *TaskDataBase) EditTaskData(id int64, data *TaskDataBase) error {
	tx := GetSessionTx(a.Session)
	return tx.Model(&TaskDataBase{}).Where("id = ?", id).Updates(data).Error
}

func (a *TaskDataBase) AddTaskData(Data *TaskDataBase) (int64, error) {
	tx := GetSessionTx(a.Session)
	err := tx.Create(Data).Error
	if err != nil {
		return 0, err
	}
	return Data.Id, nil
}

func (a *TaskDataBase) DeleteTaskData(id int64) error {
	tx := GetSessionTx(a.Session)
	return tx.Where("id = ?", id).Delete(TaskDataBase{}).Error
}
