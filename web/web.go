package web

import (
	"fmt"
	"log"
	"net/http"
	"strings"
	"wordcounter/cluster"
	"wordcounter/config"
	"wordcounter/distributetask"
)

// ServeWeb ...
// Server web for browser to visit at port :config.Envs["ENV_HTTP_STATIC_PORT"]
// It will redirect to Leader node if user request it from a client node's port
func ServeWeb() {

	fmt.Println("[INFO] ServeWeb")

	serveLeaderWebRedirect()
	serveHTMLStatic()
	serveHTTPEndpoint()

	port := ":" + config.Envs["ENV_HTTP_STATIC_PORT"]

	log.Printf("[INFO] Static Web server started on %v...", port)
	err := http.ListenAndServe(port, nil)
	if err != nil {
		log.Fatal(err)
	}
}

func isRequestFromLeaderIP(w http.ResponseWriter, r *http.Request) bool {
	leaderIP := cluster.GetLeaderIP()
	return strings.HasPrefix(r.Host, leaderIP)
}

func serveLeaderWebRedirect() {
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {

		handler := serveHTMLStatic()
		handler.ServeHTTP(w, r)

	})

}
func getURLScheme(r *http.Request) string {
	return r.URL.Scheme
}
func serveHTMLStatic() http.Handler {
	return http.FileServer(http.Dir("./static"))
}

func httpHandler(w http.ResponseWriter, r *http.Request) {

	if !isHTTPPOST(w, r) {
		return
	}

	if strings.HasSuffix(r.URL.Path, "/wordcount") {
		wordCountHandler(w, r)
	} else {
		nilHandler(w, r)
	}

}

func nilHandler(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "[ERROR] The Path doesn't have any handler", http.StatusInternalServerError)
}

func wordCountHandler(w http.ResponseWriter, r *http.Request) {

	// Parse our multipart form, 10 << 20 specifies a maximum
	// upload of 10 MB files.
	r.ParseMultipartForm(10 << 20)

	file, _, err := r.FormFile("usertxtfile")

	if err != nil {
		errMsg := "[ERROR] Unable to read the user upload"
		http.Error(w, errMsg, http.StatusNotAcceptable)
		fmt.Printf(errMsg, err)
		return
	}

	defer file.Close()

	statusCode, err := distributetask.HTTPProxyWordCount(file, w, r)

	if err != nil {
		http.Error(w, err.Error(), statusCode)
	}

}

func isRequestFromLeader(w http.ResponseWriter, r *http.Request) bool {
	leaderIP := cluster.GetLeaderIP()
	if !isRequestFromLeaderIP(w, r) {

		http.Error(
			w,
			fmt.Sprintf("[ERROR] Only Leader could serve your request. Leader IP: %s", leaderIP),
			http.StatusForbidden,
		)

		return false
	}
	return true
}

func isHTTPPOST(w http.ResponseWriter, r *http.Request) bool {
	if r.Method != "POST" {
		http.Error(w, "[ERROR] Invalid request method", http.StatusMethodNotAllowed)
		return false
	}
	return true
}

func serveHTTPEndpoint() {
	http.HandleFunc("/api/wordcount", httpHandler)
}
