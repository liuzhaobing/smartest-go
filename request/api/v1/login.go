package v1

import (
	"github.com/gin-gonic/gin"
	"smartest-go/pkg/app"
	util "smartest-go/pkg/util/const"
)

type LoginPayload struct {
	Username string `json:"username" form:"username"`
	Password string `json:"password" form:"password"`
}

type AccessToken struct {
	Token string `json:"token" from:"token"`
}

func Login(context *gin.Context) {
	req := context.MustGet(util.REQUEST_KEY).(*LoginPayload)
	app.SuccessResp(context, &AccessToken{Token: req.Username + "-token"})
}

func LogOut(context *gin.Context) {
	app.SuccessResp(context, "success")
}

type UserInfo struct {
	Roles        []string `json:"roles"`
	Introduction string   `json:"introduction"`
	Avatar       string   `json:"avatar"`
	Name         string   `json:"name"`
}

func GetUserInfo(context *gin.Context) {
	req := context.MustGet(util.REQUEST_KEY).(*AccessToken)
	app.SuccessResp(context, &UserInfo{
		Roles:        []string{"admin"},
		Introduction: req.Token,
		Avatar:       "https://wpimg.wallstcn.com/f778738c-e4f8-4870-b634-56703b4acafe.gif",
		Name:         "Super Admin",
	})
}
