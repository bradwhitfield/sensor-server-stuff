package main

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/julienschmidt/httprouter"
)

type write struct {
	Measurement string            `json:"measurement"`
	Tags        map[string]string `json:"tags"`
	Key         string            `json:"key"`
	Value       string            `json:"value"`
}

// TODO: use zap
func writeHandler(idb *InfluxDB) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Body == nil {
			w.WriteHeader(http.StatusBadRequest)
			w.Write([]byte("Missing JSON body.")) // TODO: better error mesasage
			return
		}

		b, err := ioutil.ReadAll(r.Body)
		defer func() {
			if err := r.Body.Close(); err != nil {
				fmt.Printf("Failed to close body. Error was %s\n", err.Error())
			}
		}()
		if err != nil {
			fmt.Println("Failed to read body.")
			http.Error(w, "oops, sorry", http.StatusInternalServerError)
			return
		}

		var dbw write
		err = json.Unmarshal(b, &dbw)
		if err != nil {
			fmt.Println(err)
			http.Error(w, "The request could not be deserialized.", http.StatusInternalServerError)
			return
		}

		go func() {
			err := idb.Write(dbw.Measurement, dbw.Tags, dbw.Key, dbw.Value)
			if err != nil {
				fmt.Printf("Shit's broke: %s\n", err.Error())
			}
		}()

		w.WriteHeader(http.StatusAccepted)
		return
	})
}

func main() {
	idb := InfluxDB{
		AuthToken: os.Getenv("AUTH_TOKEN"),
		Org:       os.Getenv("ORG"),
		Bucket:    os.Getenv("BUCKET"),
		Host:      os.Getenv("INFLUX_HOST"),
		Client: &http.Client{
			Timeout: time.Second * 5,
		},
	}

	router := httprouter.New()
	router.Handler(http.MethodPost, "/write", writeHandler(&idb))
	http.ListenAndServe(":8080", router)
}
