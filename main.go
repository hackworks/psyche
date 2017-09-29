package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

var roomMapping = map[string]string{
	"garage":              "Dhruva's private room",
	"perms_dev":           "Perms Dev",
	"permissions_service": "Permissions Service",
	"triforce":            "Triforce (MTV Identity)",
}
var postURL = map[string]string{
	"garage":              "https://botnana.domain.dev.atlassian.io/message?secret=86ebf927b1754b8deb759c1c29701f4bb3ed5bb50eecac981fe9bcb26a733700745c46c4550c2ef8",
	"perms_dev":           "https://botnana.domain.dev.atlassian.io/message?secret=e74648426f6e0c67b750c7ebb7aa021f334bce305b19aeea562f72dbbf02fad59e1370f77b662332",
	"permissions_service": "https://botnana.domain.dev.atlassian.io/message?secret=126b4ff64e38d95929924a2b3527b24512ae8e366149d41f168bd0a84b8ddc8dcc07d4cf35ad4c17",
	"triforce":            "https://botnana.domain.dev.atlassian.io/message?secret=c80c4ee687de9b95784916882d005a9e69f2ce91c76604a68e77cbdda2da79690ab5987058fc1aea",
}

type msg struct {
	Text   string `json:"text"`
	Format string `json:"format"`
}

func getResponse(source string, data []byte) msg {
	str := fmt.Sprintf("message from %s: %s?", source, data)
	return msg{str, "text"}
}

func requestHandle(w http.ResponseWriter, req *http.Request) {
	if req.Method == http.MethodPost {
		w.WriteHeader(http.StatusOK)

		source := req.URL.Query().Get("source")
		target := req.URL.Query().Get("target")

		ep, ok := postURL[target]
		if !ok {
			fmt.Printf("failed to find post url for target %s", target)
			return
		}

		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			fmt.Printf("failed to read request body with error %s", err)
			return
		}

		sourceRoom, ok := roomMapping[source]
		if !ok {
			sourceRoom = "Unknown"
		}

		m := getResponse(sourceRoom, data)
		data, err = json.Marshal(&m)
		if err != nil {
			fmt.Printf("failed to marshal response body %s with error %s", m, err)
			return
		}

		body := strings.NewReader(string(data))
		http.Post(ep, "application/json", body)
	}
}

func healthcheckHandle(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("ok"))
}

func main() {
	http.HandleFunc("/", requestHandle)
	http.HandleFunc("/healthcheck", healthcheckHandle)
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		fmt.Printf("failed to start server with error %s", err)
		return
	}
}
