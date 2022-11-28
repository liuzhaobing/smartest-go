package models

import (
	"github.com/jinzhu/gorm"
)

type TaskServerBase struct {
	Id            int64  `json:"id"  gorm:"primary_key"    gorm:"column:id"`
	ServerName    string `json:"name"   gorm:"column:name"`
	ServerTypes   string `json:"types"   gorm:"column:types"`
	ServerAddress string `json:"address"   gorm:"column:address"`

	Session *Session `json:"-" gorm:"-"`
}

func (TaskServerBase) TableName() string {
	return "servers"
}

func NewTaskServerModel() *TaskServerBase {
	return &TaskServerBase{Session: NewSession()}
}

func (a *TaskServerBase) ExistTaskServerByID(id int64) (bool, error) {
	servers := &TaskServerBase{}
	err := a.Session.db.Select("id").Where("id = ? ", id).First(servers).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}
	if servers.Id > 0 {
		return true, nil
	}
	return false, nil
}

func (a *TaskServerBase) GetTaskServerTotal(query interface{}, args ...interface{}) (int64, error) {
	var count int64
	err := a.Session.db.Model(&TaskServerBase{}).Where(query, args...).Count(&count).Error
	return count, err
}

func (a *TaskServerBase) GetTaskServers(pageNum int, pageSize int, maps interface{}, args ...interface{}) ([]*TaskServerBase, error) {
	var servers []*TaskServerBase
	err := a.Session.db.Where(maps, args...).Offset(pageNum).Limit(pageSize).Find(&servers).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}
	return servers, nil
}

func (a *TaskServerBase) GetTaskServer(id int64) (*TaskServerBase, error) {
	server := &TaskServerBase{}
	err := a.Session.db.Where("id = ?", id).First(server).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	} else if err != nil && err == gorm.ErrRecordNotFound {
		return nil, nil
	}
	return server, nil
}

func (a *TaskServerBase) EditTaskServer(id int64, data *TaskServerBase) error {
	tx := GetSessionTx(a.Session)
	return tx.Model(&TaskServerBase{}).Where("id = ?", id).Updates(data).Error
}

func (a *TaskServerBase) AddTaskServer(server *TaskServerBase) (int64, error) {
	tx := GetSessionTx(a.Session)
	err := tx.Create(server).Error
	if err != nil {
		return 0, err
	}
	return server.Id, nil
}

func (a *TaskServerBase) DeleteTaskServer(id int64) error {
	tx := GetSessionTx(a.Session)
	return tx.Where("id = ?", id).Delete(TaskServerBase{}).Error
}
