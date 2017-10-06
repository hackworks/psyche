package main

import (
	"database/sql"
	"fmt"
	"log"
	"net/http"
	"net/url"
	"os"

	"bitbucket.org/psyche/plugins"
	"bitbucket.org/psyche/types"
	_ "github.com/lib/pq"
)

var psyches = make(plugins.Psyches)

func healthcheckHandle(w http.ResponseWriter, req *http.Request) {
	w.Write([]byte("ok\r\n"))
}

func httpHandler(endpoint string) func(w http.ResponseWriter, req *http.Request) {
	return func(w http.ResponseWriter, req *http.Request) {
		msg, err := types.NewRecvMsg(req.Body)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte(err.Error()))
			w.Write([]byte("\r\n"))
			return
		}

		p, ok := psyches[endpoint]
		if !ok {
			return
		}

		_, err = p.Handle(req.URL, msg)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(err.Error()))
			w.Write([]byte("\r\n"))

			if p, ok := psyches["relay"]; ok {
				u := url.URL{}
				u.Query().Add("source", msg.Context)
				u.Query().Add("target", "error:error")
				msg.Message = fmt.Sprintf("psyche request error: endpoint=%s, error=%s", endpoint, err)

				p.Handle(&u, msg)
			}
		}

		return
	}
}

func main() {
	var err error
	var dbh *sql.DB
	http.HandleFunc("/healthcheck", healthcheckHandle)

	// To run locally, run postgres and set the following env
	// PG_PSYCHE_URL="postgres://postgres@localhost:5432/postgres?sslmode=disable"
	if pgurl, ok := os.LookupEnv("PG_PSYCHE_URL"); ok {
		dbh, err = sql.Open("postgres", pgurl)
		if err != nil {
			log.Fatalf("failed to initialize DB connection with error %s", err)
		}
		dbh.SetMaxOpenConns(50)
	}

	// Plugins that require persistence
	if dbh != nil {
		psyches["register"] = plugins.NewRegisterPlugin(dbh, psyches)
		http.HandleFunc("/register", httpHandler("register"))

		psyches["indexer"] = plugins.NewIndexerPlugin(dbh, psyches)
		http.HandleFunc("/indexer", httpHandler("indexer"))

		psyches["search"] = plugins.NewSearchPlugin(dbh, psyches)
		http.HandleFunc("/search", httpHandler("search"))
	}

	psyches["relay"] = plugins.NewRelayPlugin(dbh, psyches)
	http.HandleFunc("/relay", httpHandler("relay"))

	// Start the server
	if err := http.ListenAndServe(":8080", nil); err != nil {
		fmt.Printf("failed to start server with error %s\n", err)
		return
	}
}
