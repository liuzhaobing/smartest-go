package models

import (
	"github.com/360EntSecGroup-Skylar/excelize"
	"github.com/jinzhu/gorm"
	"smartest-go/pkg/logf"
	"strconv"
)

type QaBaseTest struct {
	Id         int64    `json:"id"  gorm:"primary_key"    gorm:"column:id"` //
	Question   string   `json:"question"   gorm:"column:question"`          // 问题
	AnswerList string   `json:"answer_list"   gorm:"column:answer_list"`    // 答案列表
	QaType     string   `json:"qa_type"   gorm:"column:qa_type"`            // QA行业分类
	QaSource   string   `json:"qa_source"   gorm:"column:qa_source"`        // qa的来源
	QaGroupId  int64    `json:"qa_group_id"   gorm:"column:qa_group_id"`    // QA的group_id
	CreateTime JSONTime `json:"create_time"   gorm:"column:create_time"`    // 创建时间
	UpdateTime JSONTime `json:"update_time"   gorm:"column:update_time"`    // 更新时间
	RobotType  string   `json:"robot_type"    gorm:"column:robot_type"`     // 机型
	RobotId    string   `json:"robot_id"    gorm:"column:robot_id"`         // 机型

	Session *Session `json:"-" gorm:"-"`
}

func (QaBaseTest) TableName() string {
	return "qa_base_test"
}

func NewQaBaseTestModel() *QaBaseTest {
	return &QaBaseTest{Session: NewSession()}
}

// ExistQaBaseTestByID checks if an QaBaseTest exists based on ID
func (a *QaBaseTest) ExistQaBaseTestByID(id int64) (bool, error) {
	qaBaseTest := &QaBaseTest{}
	err := a.Session.db.Select("id").Where("id = ? ", id).First(qaBaseTest).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return false, err
	}

	if qaBaseTest.Id > 0 {
		return true, nil
	}

	return false, nil
}

// GetQaBaseTestTotal gets the total number of qa_base_tests based on the constraints
func (a *QaBaseTest) GetQaBaseTestTotal(query interface{}, args ...interface{}) (int64, error) {
	var count int64
	err := a.Session.db.Model(&QaBaseTest{}).Where(query, args...).Count(&count).Error
	return count, err
}

// GetQaBaseTests gets a list of qa_base_tests based on paging constraints
func (a *QaBaseTest) GetQaBaseTests(pageNum int, pageSize int, maps interface{}) ([]*QaBaseTest, error) {
	var qaBaseTests []*QaBaseTest
	err := a.Session.db.Where(maps).Offset(pageNum).Limit(pageSize).Find(&qaBaseTests).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	}

	return qaBaseTests, nil
}

// GetQaBaseTestById Get a single qa_base_test based on ID
func (a *QaBaseTest) GetQaBaseTestById(id int64) (*QaBaseTest, error) {
	qaBaseTest := &QaBaseTest{}
	err := a.Session.db.Where("id = ?", id).First(qaBaseTest).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	} else if err != nil && err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return qaBaseTest, nil
}

// GetQaBaseTest Get a single qa_base_test based on ID
func (a *QaBaseTest) GetQaBaseTest(query interface{}, args ...interface{}) (*QaBaseTest, error) {
	qaBaseTest := &QaBaseTest{}
	err := a.Session.db.Where(query, args...).First(qaBaseTest).Error
	if err != nil && err != gorm.ErrRecordNotFound {
		return nil, err
	} else if err != nil && err == gorm.ErrRecordNotFound {
		return nil, nil
	}

	return qaBaseTest, nil
}

// EditQaBaseTest modify a single qa_base_test
func (a *QaBaseTest) EditQaBaseTest(id int64, data map[string]interface{}) error {
	tx := GetSessionTx(a.Session)
	return tx.Model(&QaBaseTest{}).Where("id = ?", id).Updates(data).Error
}

// UpdateQaBaseTest modify a single qa_base_test
func (a *QaBaseTest) UpdateQaBaseTest(id int64, QABaseTest interface{}) error {
	tx := GetSessionTx(a.Session)
	return tx.Save(&QABaseTest).Error
}

// AddQaBaseTest add a single qa_base_test
func (a *QaBaseTest) AddQaBaseTest(qaBaseTest *QaBaseTest) (int64, error) {
	tx := GetSessionTx(a.Session)
	err := tx.Create(qaBaseTest).Error
	if err != nil {
		return 0, err
	}
	return qaBaseTest.Id, nil
}

// DeleteQaBaseTest delete a single qa_base_test
func (a *QaBaseTest) DeleteQaBaseTest(id int64) error {
	tx := GetSessionTx(a.Session)
	return tx.Where("id = ?", id).Delete(QaBaseTest{}).Error
}

type QAGroupResult struct {
	QAType string
	Count  int
}

// GetGroupQABaseTest group by的结果
func (a *QaBaseTest) GetGroupQABaseTest(sql string, values ...interface{}) ([]*QAGroupResult, error) {
	tx := GetSessionTx(a.Session)
	var res []*QAGroupResult
	result, err := tx.Raw(sql, values...).Rows()
	if err != nil {
		return nil, err
	}
	for result.Next() {
		result.Scan()
		var col = &QAGroupResult{}
		result.Scan(&col.QAType, &col.Count)
		res = append(res, col)
	}
	return res, nil
}

func (a *QaBaseTest) ExcelToDB(filename, sheet string) {
	//open file
	f, err := excelize.OpenFile(filename)
	if err != nil {
		logf.Debug(err)
		return
	}
	rows := f.GetRows(sheet)
	tx := GetSessionTx(a.Session)
	//read excel
	for _, row := range rows {
		tmpQReq := QaBaseTest{
			Id:         0,
			Question:   "",
			AnswerList: "",
			QaType:     "",
			QaSource:   "",
			QaGroupId:  0,
			RobotType:  "",
		}
		for i, colCell := range row {
			if i == 0 {
				id, _ := strconv.ParseInt(colCell, 10, 64)
				tmpQReq.Id = id
			}
			if i == 1 {
				tmpQReq.Question = colCell
			}
			if i == 2 {
				tmpQReq.AnswerList = colCell
			}
			if i == 3 {
				tmpQReq.QaType = colCell
			}
			if i == 4 {
				tmpQReq.QaSource = colCell
			}
			if i == 5 {
				QaGroupId, _ := strconv.ParseInt(colCell, 10, 64)
				tmpQReq.QaGroupId = QaGroupId
			}
			if i == 6 {
				tmpQReq.RobotType = colCell
			}
			if i == 7 {
				tmpQReq.RobotId = colCell
			}
		}
		//write to db
		tx.Create(tmpQReq)
	}

}
