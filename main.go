package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

var roomMapping = map[string]string{
	"garage":              "Dhruva private room",
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
	str := fmt.Sprintf("Message from %s: %s?", source, data)
	return msg{str, "text"}
}

func requestHandle(w http.ResponseWriter, req *http.Request) {
	var stmt string
	var httperr int

	defer func() {
		if httperr != 200 {
			fmt.Println(stmt)

			w.WriteHeader(httperr)
			w.Write([]byte(stmt))
			w.Write([]byte("\r\n"))
		}
	}()

	fmt.Printf("micros_psyche: received messge from host %s\n", req.Host)

	if req.Method == http.MethodPost {
		source := req.URL.Query().Get("source")
		target := req.URL.Query().Get("target")

		ep, ok := postURL[target]
		if !ok {
			httperr = http.StatusBadRequest
			stmt = fmt.Sprintf("micros_psyche: failed to find post url for target %s", target)
			return
		}

		data, err := ioutil.ReadAll(req.Body)
		if err != nil {
			httperr = http.StatusBadRequest
			stmt = fmt.Sprintf("micros_psyche: failed to read request body with error %s", err)
			return
		}

		sourceRoom, ok := roomMapping[source]
		if !ok {
			sourceRoom = "Unknown"
		}

		m := getResponse(sourceRoom, data)
		body := new(bytes.Buffer)
		err = json.NewEncoder(body).Encode(&m)
		if err != nil {
			httperr = http.StatusInternalServerError
			stmt = fmt.Sprintf("micros_psyche: failed to encode response body with error %s", err)
			return
		}

		resp, err := http.Post(ep, "application/json", body)
		if err != nil {
			httperr = http.StatusInternalServerError
			stmt = fmt.Sprintf("micros_psyche: http post to %s failed with error %s", ep, err)
			return
		}

		data, err = ioutil.ReadAll(resp.Body)
		resp.Body.Close()

		httperr = resp.StatusCode
		stmt = fmt.Sprintf("micros_psyche: http post received response %s", data)
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
		fmt.Printf("failed to start server with error %s\n", err)
		return
	}
}
