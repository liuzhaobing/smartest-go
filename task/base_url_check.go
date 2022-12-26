package task

import (
	"fmt"
	"net/http"
)

func checkMusicUrl(url string) string {
	req, _ := http.NewRequest("GET", url, nil)
	res, _ := http.DefaultClient.Do(req)
	defer res.Body.Close()
	return fmt.Sprintf(`{"status_code":%d}`, res.StatusCode)
}
