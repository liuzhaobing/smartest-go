package task

import (
	"encoding/json"
	"smartest-go/models"
)

/*
开放给前端的测试计划管理结构 同时提供给内部使用
*/

type ListTask struct {
	Id        int64  `json:"id" form:"id"`
	TaskName  string `json:"task_name" form:"task_name"`
	TaskType  string `json:"task_type" form:"task_type"`
	TaskGroup string `json:"task_group" form:"task_group"`
	IsCrontab bool   `json:"is_crontab" form:"is_crontab"`

	PageNum  int `form:"page_num,default=1" json:"page_num"`
	PageSize int `form:"page_size,default=30" json:"page_size"`
}

type NameOfTask struct {
	TaskName string `json:"task_name" form:"task_name"`
}

type AddTask struct {
	Id                  int64           `json:"id,omitempty" form:"id,omitempty"`
	TaskName            string          `json:"task_name" form:"task_name"`
	TaskType            string          `json:"task_type" form:"task_type"`
	TaskGroup           string          `json:"task_group" form:"task_group"`
	IsCrontab           string          `json:"is_crontab" form:"is_crontab"`
	CrontabString       string          `json:"crontab_string" form:"crontab_string"`
	TaskDataSourceLabel string          `json:"task_data_source_label" form:"task_data_source_label"`
	TaskConfig          *TestConfig     `json:"task_config" form:"task_config"`
	TaskDataSource      *TestDataSource `json:"task_data_source" form:"task_data_source"`
}

type Excel struct {
	FileName  string `json:"file_name" form:"file_name"`
	SheetName string `json:"sheet_name" form:"sheet_name"`
}

type TestConfig struct {
	TestConfigKG    *KGTaskConfig    `json:"config_kg,omitempty" form:"config_kg,omitempty"` // 知识图谱
	TestConfigQA    *QATaskConfig    `json:"config_qa,omitempty" form:"config_qa,omitempty"`
	TestConfigSkill *SkillTaskConfig `json:"config_skill,omitempty" form:"config_skill,omitempty"`
	TestConfigASR   *ASRTaskConfig   `json:"config_asr,omitempty" form:"config_asr,omitempty"`
}

type ReportString struct {
	Address string `json:"address,omitempty" form:"address,omitempty"`
}

type TestDataSource struct {
	TestCaseKG   []*KGTaskReq  `json:"cases_kg,omitempty" form:"cases_kg,omitempty"`   // 知识图谱用例数据
	KGDataSource *KGDataSource `json:"source_kg,omitempty" form:"source_kg,omitempty"` // 知识图谱用例构造
	KGExcel      *Excel        `json:"excel_kg,omitempty" form:"excel_kg,omitempty"`

	TestCaseQA   []*QATaskReq  `json:"cases_qa,omitempty" form:"cases_qa,omitempty"`   // QA 用例数据
	QADataSource *QADataSource `json:"source_qa,omitempty" form:"source_qa,omitempty"` // QA 用例构造
	QAExcel      *Excel        `json:"excel_qa,omitempty" form:"excel_qa,omitempty"`   // QA Excel用例

	TestCaseSkill   []*SkillTaskReq  `json:"cases_skill,omitempty" form:"cases_skill,omitempty"`
	SkillDataSource *SkillDataSource `json:"source_skill,omitempty" form:"source_skill,omitempty"`
	SkillExcel      *Excel           `json:"excel_skill,omitempty" form:"excel_skill,omitempty"`

	ASRDataSource *ASRDataSource `json:"source_asr,omitempty" form:"source_asr,omitempty"`
}

func JsonToStruct(j *models.TaskPlanBase) (*AddTask, error) {
	// 根据类型 将数据库string类型转换为前端可识别struct
	s := &AddTask{
		Id:                  j.Id,
		TaskName:            j.TaskName,
		TaskType:            j.TaskType,
		TaskGroup:           j.TaskGroup,
		IsCrontab:           j.IsCrontab,
		CrontabString:       j.CrontabString,
		TaskDataSourceLabel: j.TaskDataSourceLabel,
	}

	switch j.TaskType {
	case KnowledgeGraph:
		s.TaskConfig = &TestConfig{TestConfigKG: &KGTaskConfig{}}
		err := json.Unmarshal([]byte(j.TaskConfig), s.TaskConfig)
		if err != nil {
			return nil, err
		}

		s.TaskDataSource = &TestDataSource{
			TestCaseKG:   make([]*KGTaskReq, 0),
			KGDataSource: &KGDataSource{},
			KGExcel:      &Excel{},
		}
		err = json.Unmarshal([]byte(j.TaskDataSource), &s.TaskDataSource)
		if err != nil {
			return nil, err
		}
	case CommonQA:
		s.TaskConfig = &TestConfig{TestConfigQA: &QATaskConfig{}}
		err := json.Unmarshal([]byte(j.TaskConfig), s.TaskConfig)
		if err != nil {
			return nil, err
		}

		s.TaskDataSource = &TestDataSource{
			TestCaseQA:   make([]*QATaskReq, 0),
			QADataSource: &QADataSource{},
			QAExcel:      &Excel{},
		}
		err = json.Unmarshal([]byte(j.TaskDataSource), &s.TaskDataSource)
		if err != nil {
			return nil, err
		}
	case SystemSkill:
		s.TaskConfig = &TestConfig{TestConfigSkill: &SkillTaskConfig{}}
		err := json.Unmarshal([]byte(j.TaskConfig), s.TaskConfig)
		if err != nil {
			return nil, err
		}

		s.TaskDataSource = &TestDataSource{
			TestCaseSkill:   make([]*SkillTaskReq, 0),
			SkillDataSource: &SkillDataSource{},
			SkillExcel:      &Excel{},
		}
		err = json.Unmarshal([]byte(j.TaskDataSource), &s.TaskDataSource)
		if err != nil {
			return nil, err
		}
	case CommonASR:
		s.TaskConfig = &TestConfig{TestConfigASR: &ASRTaskConfig{}}
		err := json.Unmarshal([]byte(j.TaskConfig), s.TaskConfig)
		if err != nil {
			return nil, err
		}
		s.TaskDataSource = &TestDataSource{
			ASRDataSource: &ASRDataSource{},
		}
		err = json.Unmarshal([]byte(j.TaskDataSource), &s.TaskDataSource)
		if err != nil {
			return nil, err
		}
	}
	return s, nil
}

func StructToJson(s *AddTask) (*models.TaskPlanBase, error) {
	// 将前端可识别struct类型转换为数据库string
	j := &models.TaskPlanBase{
		TaskName:            s.TaskName,
		TaskType:            s.TaskType,
		TaskGroup:           s.TaskGroup,
		IsCrontab:           s.IsCrontab,
		CrontabString:       s.CrontabString,
		TaskDataSourceLabel: s.TaskDataSourceLabel,
	}

	Config, err := json.Marshal(s.TaskConfig)
	if err != nil {
		return nil, err
	}
	j.TaskConfig = string(Config)
	Data, err := json.Marshal(s.TaskDataSource)
	if err != nil {
		return nil, err
	}
	j.TaskDataSource = string(Data)
	return j, nil
}

func InitTaskModel(config *AddTask) TaskModel {
	switch config.TaskType {
	case KnowledgeGraph:
		var kg TaskModel = &KGTask{}
		kgConfig := config.TaskConfig.TestConfigKG

		switch config.TaskDataSourceLabel {
		case KnowledgeGraphSource:
			kg = &KGTaskTest{KGTask: NewKGTask(kgConfig, make([]*KGTaskReq, 0), config.TaskDataSource.KGDataSource)}
		case KnowledgeGraphCases:
			kg = &KGTaskTest{KGTask: NewKGTask(kgConfig, config.TaskDataSource.TestCaseKG, nil)}
		case KnowledgeGraphExcel:
			kg = &KGTaskTest{KGTask: NewKGTask(kgConfig, ExcelKGReader(config.TaskDataSource.KGExcel.FileName, config.TaskDataSource.KGExcel.SheetName), nil)}
		}
		return kg
	case CommonQA:
		var qa TaskModel = &QATask{}
		qaConfig := config.TaskConfig.TestConfigQA

		switch config.TaskDataSourceLabel {
		case CommonQASource:
			qa = &QATaskTest{QATask: NewQATask(qaConfig, make([]*QATaskReq, 0), config.TaskDataSource.QADataSource)}
		case CommonQACases:
			qa = &QATaskTest{QATask: NewQATask(qaConfig, config.TaskDataSource.TestCaseQA, nil)}
		case CommonQAExcel:
			qa = &QATaskTest{QATask: NewQATask(qaConfig, ExcelQAReader(config.TaskDataSource.QAExcel.FileName, config.TaskDataSource.QAExcel.SheetName), nil)}
		}
		return qa
	case SystemSkill:
		var Skill TaskModel = &SkillTask{}
		SkillConfig := config.TaskConfig.TestConfigSkill

		switch config.TaskDataSourceLabel {
		case SystemSkillSource:
			Skill = &SkillTaskTest{SkillTask: NewSkillTask(SkillConfig, make([]*SkillTaskReq, 0), config.TaskDataSource.SkillDataSource)}
		case SystemSkillCases:
			Skill = &SkillTaskTest{SkillTask: NewSkillTask(SkillConfig, config.TaskDataSource.TestCaseSkill, nil)}
		case SystemSkillExcel:
			Skill = &SkillTaskTest{SkillTask: NewSkillTask(SkillConfig, ExcelSkillReader(config.TaskDataSource.SkillExcel.FileName, config.TaskDataSource.SkillExcel.SheetName), nil)}
		}
		return Skill
	case CommonASR:
		var ASR TaskModel = &ASRTask{}
		ASRConfig := config.TaskConfig.TestConfigASR

		switch config.TaskDataSourceLabel {
		case CommonASRSource:
			ASR = &ASRTaskTest{ASRTask: NewASRTask(ASRConfig, make([]*ASRTaskReq, 0), config.TaskDataSource.ASRDataSource)}
		}
		return ASR
	}
	return nil
}

var (
	KnowledgeGraph       = "kg"
	KnowledgeGraphSource = "source_kg"
	KnowledgeGraphCases  = "cases_kg"
	KnowledgeGraphExcel  = "excel_kg"
	CommonQA             = "qa"
	CommonQASource       = "source_qa"
	CommonQACases        = "cases_qa"
	CommonQAExcel        = "excel_qa"
	SystemSkill          = "skill"
	SystemSkillSource    = "source_skill"
	SystemSkillCases     = "cases_skill"
	SystemSkillExcel     = "excel_skill"
	CommonASR            = "asr"
	CommonASRSource      = "source_asr"
)
