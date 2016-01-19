package main
import (
	"net/http"
	"github.com/gorilla/mux"
	"fmt"	
	"log"
	"encoding/json"
	_"crypto/tls"
)

func WebServer() {
	r := mux.NewRouter()
	r.HandleFunc("/", HomeHandler)
	r.HandleFunc("/status", StatusHandler)
	http.Handle("/", r)	
	host := fmt.Sprintf("%s:%d",CONFIGS.StatusHost,CONFIGS.StatusPort)
	log.Println("Web Server:",host)
	http.ListenAndServe(host, nil)
}

func jsonParser(data interface{},w http.ResponseWriter) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	if data != nil {
		json, err := json.Marshal(data)	
		if err != nil {
			w.WriteHeader(500)
			log.Println("Error generating json", err)
			fmt.Fprintln(w, "Could not generate JSON")
			return
		}
		fmt.Fprint(w, string(json))
	} else {
		w.WriteHeader(404)
		fmt.Fprint(w, "404 no data can be find.")
	}
}

func HomeHandler(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(403)
	fmt.Fprint(w, "403 Forbidden")
}

func StatusHandler(w http.ResponseWriter, r *http.Request) {
	data := PeerStatusList
	jsonParser(data,w)
}