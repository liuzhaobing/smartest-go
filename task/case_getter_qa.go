package task

import (
	"context"
	"smartest-go/models"
	"strings"
)

func (QA *QATask) CaseGetterQA(c context.Context) {
	qaModel := models.NewQaBaseTestModel()
	total, err := qaModel.GetQaBaseTestTotal("1=1")
	if err != nil {
		// TODO 异常处理
		return
	}
	resultList, err := qaModel.GetQaBaseTests(0, int(total), QA.QADataSourceConfig.DBFilter)
	if err != nil {
		// TODO 异常处理
		return
	}
	if len(resultList) == 0 {
		// TODO
		return
	}
	QA.req = QA.req[0:0]
	for _, qaBaseTest := range resultList {
		QA.req = append(QA.req, &QATaskReq{
			Id:           qaBaseTest.Id,
			Query:        qaBaseTest.Question,
			ExpectAnswer: strings.Split(qaBaseTest.AnswerList, "&&"),
			ExpectGroup:  qaBaseTest.QaGroupId,
			RobotType:    qaBaseTest.RobotType,
		})
	}
}
