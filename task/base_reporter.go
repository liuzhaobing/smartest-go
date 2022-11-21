package task

import (
	"bytes"
	"encoding/json"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net/http"
	"smartest-go/pkg/app"
	"smartest-go/pkg/e"
	"smartest-go/pkg/logf"
	util "smartest-go/pkg/util/const"
)

type FeiShuPayload struct {
	MsgType string      `json:"msg_type"`
	Content interface{} `json:"content"`
}

// WebHookToFeiShu 向飞书机器人发送消息
func WebHookToFeiShu(url string, payload *FeiShuPayload) {
	method := "POST"
	p, _ := json.Marshal(&payload)
	client := &http.Client{}
	req, err := http.NewRequest(method, url, bytes.NewReader(p))
	if err != nil {
		logf.Error("err is ", err)
		return
	}
	req.Header.Add("Content-Type", "application/json")
	res, err := client.Do(req)
	defer res.Body.Close()
	body, err := ioutil.ReadAll(res.Body)
	if err != nil {
		logf.Error("http err ,err is ", err)
	}
	logf.Info(string(body))
}

type simpleText struct {
	Text string `json:"text"`
}

type richText struct {
	Chinese *richTextInfo `json:"zh_cn,omitempty"`
	English *richTextInfo `json:"en_us,omitempty"`
}

type richTextInfo struct {
	Title   string                         `json:"title"`
	Content [][]*richTextInfoContentDetail `json:"content"`
}

type richTextInfoContentDetail struct {
	Tag       string `json:"tag"`
	Text      string `json:"text,omitempty"`
	Href      string `json:"href,omitempty"`
	UserName  string `json:"user_name,omitempty"`
	UserId    string `json:"user_id,omitempty"`
	ImageKey  string `json:"image_key,omitempty"`
	FileKey   string `json:"file_key,omitempty"`
	EmojiType string `json:"emoji_type,omitempty"`
}

type FeiShu struct {
	Url     string         `json:"url" form:"url"`
	Payload *FeiShuPayload `json:"payload" form:"payload"`
}

func reportToFeiShu(f *FeiShu) error {
	c, err := json.Marshal(f.Payload.Content)
	if err != nil {
		return err
	}
	switch f.Payload.MsgType {
	case "text":
		var p *simpleText
		err = json.Unmarshal(c, &p)
		if err != nil {
			return err
		}
		f.Payload.Content = p
	case "post":
		var p *richText
		err = json.Unmarshal(c, &p)
		if err != nil {
			return err
		}
		f.Payload.Content = p
	}
	WebHookToFeiShu(f.Url, f.Payload)
	return nil
}

func SendReportToFeiShu(context *gin.Context) {
	req := context.MustGet(util.REQUEST_KEY).(*FeiShu)
	err := reportToFeiShu(req)
	if err != nil {
		app.ErrorResp(context, e.ERROR, err.Error(), nil)
		return
	}
	app.SuccessRespByCode(context, e.SUCCESS, nil)
}
