package task

import (
	"fmt"
	"net/http"
)

func checkMusicUrl(url string) string {
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return fmt.Sprintf(`{"error":%s}`, err.Error())
	}
	res, err := http.DefaultClient.Do(req)
	if err != nil {
		return fmt.Sprintf(`{"error":%s}`, err.Error())
	}
	res.Body.Close()
	return fmt.Sprintf(`{"status_code":%d}`, res.StatusCode)
}
