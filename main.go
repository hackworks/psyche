package main

import (
	"bytes"
	"database/sql"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"sync"
)

type roomInfo struct {
	name string
	url  string
}

type registerMsg struct {
	Key  string `json:"key"`
	Name string `json:"name"`
	URL  string `json:"url"`
}

var roomMapping sync.Map

type sendMsg struct {
	Text   string `json:"text"`
	Format string `json:"format"`
}

type recvMsg struct {
	Message string `json:"message"`
	Sender  struct {
		ID string `json:"id"`
	} `json:"sender"`
}

type sqlDB struct {
	*sql.DB
}

func init() {
	roomMapping.Store("garage", &roomInfo{"Dhruva's private room", "https://botnana.domain.dev.atlassian.io/message?secret=86ebf927b1754b8deb759c1c29701f4bb3ed5bb50eecac981fe9bcb26a733700745c46c4550c2ef8"})
	roomMapping.Store("perms_dev", &roomInfo{"Perms Dev", "https://botnana.domain.dev.atlassian.io/message?secret=e74648426f6e0c67b750c7ebb7aa021f334bce305b19aeea562f72dbbf02fad59e1370f77b662332"})
	roomMapping.Store("permissions_service", &roomInfo{"Permissions Service", "https://botnana.domain.dev.atlassian.io/message?secret=126b4ff64e38d95929924a2b3527b24512ae8e366149d41f168bd0a84b8ddc8dcc07d4cf35ad4c17"})
	roomMapping.Store("triforce", &roomInfo{"Triforce (MTV Identity)", "https://botnana.domain.dev.atlassian.io/message?secret=c80c4ee687de9b95784916882d005a9e69f2ce91c76604a68e77cbdda2da79690ab5987058fc1aea"})
}

func getResponse(source string, data recvMsg) sendMsg {
	str := fmt.Sprintf("Message from room %s: %s?", source, data.Message)
	return sendMsg{str, "text"}
}

func relayHandle(w http.ResponseWriter, req *http.Request) {
	stmt := fmt.Sprintf("received request via HTTP %s from host %s\n", req.Method, req.Host)
	httperr := http.StatusMethodNotAllowed

	defer func() {
		if httperr != 200 {
			fmt.Println(stmt)

			w.WriteHeader(httperr)
			w.Write([]byte(stmt))
			w.Write([]byte("\r\n"))
		}
	}()

	if req.Method != http.MethodPost {
		return
	}

	source := req.URL.Query().Get("source")
	target := req.URL.Query().Get("target")

	val, ok := roomMapping.Load(target)
	if !ok {
		httperr = http.StatusBadRequest
		stmt = fmt.Sprintf("failed to find post url for target %s", target)
		return
	}
	room, _ := val.(*roomInfo)

	var r recvMsg
	err := json.NewDecoder(req.Body).Decode(&r)
	if err != nil {
		httperr = http.StatusBadRequest
		stmt = fmt.Sprintf("failed to read request body with error %s", err)
		return
	}

	sourceRoom := "Unknown"
	if val, ok := roomMapping.Load(source); ok {
		s, _ := val.(*roomInfo)
		sourceRoom = s.name
	}

	m := getResponse(sourceRoom, r)
	body := new(bytes.Buffer)
	err = json.NewEncoder(body).Encode(&m)
	if err != nil {
		httperr = http.StatusInternalServerError
		stmt = fmt.Sprintf("failed to encode response body with error %s", err)
		return
	}

	resp, err := http.Post(room.url, "application/json", body)
	if err != nil {
		httperr = http.StatusInternalServerError
		stmt = fmt.Sprintf("http post to %s failed with error %s", room.url, err)
		return
	}

	defer resp.Body.Close()

	data, err := ioutil.ReadAll(resp.Body)
	httperr = resp.StatusCode
	stmt = fmt.Sprintf("http post received response %s", data)
}

func (d *sqlDB) registerHandle(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("Not yet implemented"))
}

func registerInMemoryHandle(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		return
	}

	var r registerMsg
	err := json.NewDecoder(req.Body).Decode(&r)
	if err != nil {
		return
	}

	roomMapping.Store(r.Key, &roomInfo{r.Name, r.URL})
}

func healthcheckHandle(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("ok"))
}

func main() {
	// Use DB is available for storing registrations
	if pgurl, ok := os.LookupEnv("PG_PSYCHE_URL"); ok {
		if dbh, err := sql.Open("postgres", pgurl); err == nil {
			dbh.SetMaxOpenConns(5)
			http.HandleFunc("/register", (&sqlDB{dbh}).registerHandle)
		}
	} else {
		http.HandleFunc("/register", registerInMemoryHandle)
	}

	http.HandleFunc("/relay", relayHandle)
	http.HandleFunc("/healthcheck", healthcheckHandle)
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("failed to start server with error %s\n", err)
		return
	}
}
