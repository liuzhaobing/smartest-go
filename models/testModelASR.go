package models

import (
	"github.com/jinzhu/gorm"
)

type ASRBaseTest struct {
	Id      int64  `json:"id"  gorm:"primary_key"    gorm:"column:id"`
	WavFile string `json:"wav_file"   gorm:"column:wav_file"`
	Message string `json:"message"   gorm:"column:message"`
	Tags    string `json:"tags"   gorm:"column:tags"`
	IsSmoke int64  `json:"is_smoke"  gorm:"is_smoke"`

	Session *Session `json:"-" gorm:"-"`
}

func (ASRBaseTest) TableName() string {
	return "asr_base_test"
}

func NewASRBaseTestModel() *ASRBaseTest {
	return &ASRBaseTest{Session: NewSession()}
}

func (a *ASRBaseTest) ExistASRBaseTestByID(id int64) (bool, error) {
	asr_base_test := &ASRBaseTest{}
	err := a.Session.db.Select("id").Where("id = ? ", id).First(asr_base_test).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}

	if asr_base_test.Id > 0 {
		return true, nil
	}

	return false, nil
}

func (a *ASRBaseTest) GetASRBaseTestTotal(query interface{}, args ...interface{}) (int64, error) {
	var count int64
	err := a.Session.db.Model(&ASRBaseTest{}).Where(query, args...).Count(&count).Error
	return count, err
}

func (a *ASRBaseTest) GetASRBaseTests(pageNum int, pageSize int, maps interface{}, args ...interface{}) ([]*ASRBaseTest, error) {
	var asr_base_tests []*ASRBaseTest
	err := a.Session.db.Where(maps, args...).Offset(pageNum).Limit(pageSize).Find(&asr_base_tests).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return asr_base_tests, nil
}

func (a *ASRBaseTest) GetASRBaseTest(id int64) (*ASRBaseTest, error) {
	asr_base_test := &ASRBaseTest{}
	err := a.Session.db.Where("id = ?", id).First(asr_base_test).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	} else if err != nil && err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return asr_base_test, nil
}

func (a *ASRBaseTest) EditASRBaseTest(id int64, data map[string]interface{}) error {
	tx := GetSessionTx(a.Session)
	return tx.Model(&ASRBaseTest{}).Where("id = ?", id).Updates(data).Error
}

func (a *ASRBaseTest) AddASRBaseTest(asr_base_test *ASRBaseTest) (int64, error) {
	tx := GetSessionTx(a.Session)
	err := tx.Create(asr_base_test).Error
	if err != nil {
		return 0, err
	}
	return asr_base_test.Id, nil
}

func (a *ASRBaseTest) DeleteASRBaseTest(id int64) error {
	tx := GetSessionTx(a.Session)
	return tx.Where("id = ?", id).Delete(ASRBaseTest{}).Error
}

type ASRGroupResult struct {
	ASRCn string
	Count int
}

func (a *ASRBaseTest) GetGroupASRBaseTest(sql string, values ...interface{}) ([]*ASRGroupResult, error) {
	tx := GetSessionTx(a.Session)
	var res []*ASRGroupResult
	result, err := tx.Raw(sql, values...).Rows()
	if err != nil {
		return nil, err
	}
	for result.Next() {
		result.Scan()
		var col = &ASRGroupResult{}
		result.Scan(&col.ASRCn, &col.Count)
		res = append(res, col)
	}
	return res, nil
}
