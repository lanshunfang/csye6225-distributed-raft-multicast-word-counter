package web

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"wordcounter/config"
	"wordcounter/distributetask"
)

func ServeWeb() {

	serveHTMLStatic()
	serveHTTPEndpoint()

	port := ":" + config.Envs["ENV_HTTP_STATIC_PORT"]

	log.Printf("[INFO] Static Web server started on %v...", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func serveHTMLStatic() {
	fs := http.FileServer(http.Dir("./static"))
	http.Handle("/", fs)

}

func wordCountHandler(w http.ResponseWriter, r *http.Request) {
	if !isHTTPPOST(w, r) {
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		http.Error(w, "Error reading request body",
			http.StatusInternalServerError)
	}

	fmt.Fprint(w, "POST done")

	statusCode, err := distributetask.HTTPProxyWordCount(string(body), w, r)

	if err != nil {
		http.Error(w, err.Error(), statusCode)
	}

}

func isHTTPPOST(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != "POST" {
		http.Error(w, "Invalid request method", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

func serveHTTPEndpoint() {
	http.HandleFunc("/api/wordcount", wordCountHandler)
}
