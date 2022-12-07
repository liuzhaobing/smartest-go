package v1

import (
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"smartest-go/pkg/app"
	util "smartest-go/pkg/util/const"
	"sort"
)

type DirPath struct {
	Path string `json:"path" form:"path"`
}

func GetFileList(context *gin.Context) {
	req := context.MustGet(util.REQUEST_KEY).(*DirPath)
	files := GetFiles(req.Path)
	sort.Slice(files, func(i, j int) bool {
		return files[i] > files[j]
	})

	app.SuccessResp(context, files)
}

func GetFiles(folder string) (Files []string) {
	files, _ := ioutil.ReadDir(folder)
	for _, file := range files {
		if !file.IsDir() {
			Files = append(Files, folder+"/"+file.Name())
		}
	}
	return
}
