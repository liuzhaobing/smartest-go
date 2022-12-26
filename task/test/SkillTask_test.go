package test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"google.golang.org/grpc"
	"smartest-go/proto/talk"
	"testing"
)

func Test_svTalkOnce(t *testing.T) {
	traceID := uuid.New().String()
	request := &talk.TalkRequest{
		IsFull:     true,
		AgentID:    666,
		SessionID:  traceID,
		QuestionID: traceID,
		EventType:  talk.Text,
		EnvInfo:    make(map[string]string),
		RobotID:    traceID,
		TenantCode: "cloudminds",
		Position:   "104.061;30.5444",
		Version:    "v3",
		IsHa:       false,
		TestMode:   true,
		Asr:        talk.Asr{Lang: "ZH"},
	}
	addr := "172.16.32.2:32247" //m8地址
	//addr := "172.16.23.85:30811"  //fit地址
	request.RobotID = "5C1AEC03573747D"
	request.EnvInfo["devicetype"] = "ginger"
	request.InputContext = ""
	request.Asr.Text = "现在几点了"

	conn, _ := grpc.Dial(addr, grpc.WithInsecure())
	Resp, err := svTalkOnce(conn, request)
	if err != nil {
		fmt.Println(err.Error())
		panic(err)
	}

	fmt.Println(fmt.Sprintf(`{"Cost":%dms,"Source":"%s","Domain":"%s","Intent":"%s"}`,
		Resp.Cost,
		Resp.Source,
		Resp.HitLog.Fields["domain"].GetStringValue(),
		Resp.HitLog.Fields["intent"].GetStringValue(),
	))
	for _, info := range Resp.Tts {
		fmt.Println(fmt.Sprintf(`{"Answer":"%s","PicUrl":"%s","VideoUrl":"%s","AudioUrl":"%s"}`,
			info.Text,
			info.Action.Param.PicUrl,
			info.Action.Param.VideoUrl,
			info.Action.Param.Url,
		))
	}
	if Resp.HitLog.Fields["paraminfo"] != nil {
		ActualParamInfoListOriginal := Resp.HitLog.Fields["paraminfo"].GetListValue().Values
		// 将多个槽位map存入list
		var ActualParamInfoList []interface{}
		for _, ActualParamInfoListOne := range ActualParamInfoListOriginal {
			ActualParamInfoList = append(ActualParamInfoList, map[string]string{
				"BeforeValue": ActualParamInfoListOne.GetStructValue().GetFields()["BeforeValue"].GetStringValue(),
				"EntityType":  ActualParamInfoListOne.GetStructValue().GetFields()["EntityType"].GetStringValue(),
				"Name":        ActualParamInfoListOne.GetStructValue().GetFields()["Name"].GetStringValue(),
				"Value":       ActualParamInfoListOne.GetStructValue().GetFields()["Value"].GetStringValue(),
			})
		}
		ActualParamInfoStrByte, _ := json.Marshal(ActualParamInfoList)
		fmt.Println(string(ActualParamInfoStrByte))
	}
}

func svTalkOnce(conn *grpc.ClientConn, r *talk.TalkRequest) (resp *talk.TalkResponse, err error) {
	c := talk.NewTalkClient(conn)
	resp, err = c.Talk(context.Background(), r)
	return resp, err
}
