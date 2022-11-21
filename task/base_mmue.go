package task

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"io"
	"io/ioutil"
	"log"
	"net/http"
)

type EnvInfo struct {
	FrontUrl   string `json:"front_url,omitempty"`   // 前端地址
	BackendUrl string `json:"backend_url,omitempty"` // 后端地址
	Token      string `json:"token,omitempty"`       // 秘钥
	UserName   string `json:"username"`              // 用户名
	Password   string `json:"pwd"`                   // 密码
	CaptChaID  string `json:"captchaid"`             // 验证码
	AuthCode   string `json:"authcode"`              // 验证码
}

type LoginResponse struct {
	Code int `json:"code"`
	Data struct {
		Data struct {
			UserId     int         `json:"user_id"`
			UserName   string      `json:"user_name"`
			UserPower  string      `json:"user_power"`
			TenantId   string      `json:"tenant_id"`
			TenantName string      `json:"tenant_name"`
			TenantLogo string      `json:"tenant_logo"`
			IsRocuser  string      `json:"is_rocuser"`
			LibValue   interface{} `json:"lib_value"`
			AgentId    interface{} `json:"agent_id"`
		} `json:"data"`
		Token string `json:"token"`
	} `json:"data"`
	Msg    string `json:"msg"`
	Status bool   `json:"status"`
}

// mLogin 平台登录方法
func (e *EnvInfo) mLogin() {
	newBodyByte, _ := json.Marshal(e)
	req, _ := http.NewRequest("POST", e.FrontUrl+"/mmue/api/login", bytes.NewReader(newBodyByte))
	req.Header.Add("Content-Type", "application/json")

	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	res, err := client.Do(req)
	if err != nil {
	}

	if res.StatusCode == 200 {
		body, _ := ioutil.ReadAll(res.Body)
		response := LoginResponse{}
		err := json.Unmarshal(body, &response)
		if err != nil {
			return
		}
		e.Token = response.Data.Token
	}

	defer res.Body.Close()
}

// HTTPReqInfo http请求信息结构体
type HTTPReqInfo struct {
	Method  string
	Url     string
	Payload io.Reader
}

// MMUE平台发起http请求
func (e *EnvInfo) mRequest(mReq HTTPReqInfo) []byte {
	if e.Token == "" {
		e.mLogin()
	}
	req, _ := http.NewRequest(mReq.Method, mReq.Url, mReq.Payload)
	req.Header.Add("Connection", "keep-alive")
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Authorization", e.Token)
	req.Header.Add("Accept", "application/json, text/plain, */*")

	client := &http.Client{Transport: &http.Transport{TLSClientConfig: &tls.Config{InsecureSkipVerify: true}}}
	res, err := client.Do(req)
	if err != nil {
		log.Fatal(err)
	}
	if res.StatusCode != 200 {
		e.mLogin()
	}

	defer func(Body io.ReadCloser) {
		err = Body.Close()
		if err != nil {
			log.Fatal(err)
		}
	}(res.Body)

	body, _ := ioutil.ReadAll(res.Body)
	return body
}
