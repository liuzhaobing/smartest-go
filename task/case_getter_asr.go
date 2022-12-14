package task

import (
	"context"
	"smartest-go/models"
)

/*
用于ASR的用例采集
*/

func (ASR *ASRTask) CaseGetterASR(c context.Context) {
	asrModel := models.NewASRBaseTestModel()
	total, err := asrModel.GetASRBaseTestTotal("1=1")
	if err != nil {
		return
	}
	resultList, err := asrModel.GetASRBaseTests(0, int(total), ASR.ASRDataSourceConfig.DBFilter)
	if err != nil {
		return
	}
	if len(resultList) == 0 {
		return
	}
	ASR.req = ASR.req[0:0]
	for _, asrBaseTest := range resultList {
		ASR.req = append(ASR.req, &ASRTaskReq{
			Id:         asrBaseTest.Id,
			ExpMessage: asrBaseTest.Message,
			Tags:       asrBaseTest.Tags,
			WavFile:    asrBaseTest.WavFile,
			IsSmoke:    asrBaseTest.IsSmoke,
		})
	}
}
