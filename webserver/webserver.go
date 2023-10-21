package webserver

import (
	"fmt"
	"homeserver/config"
	logger "log"
	"net"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

var log logger.Logger = *logger.New(logger.Writer(), "[WEB] ", logger.LstdFlags|logger.Lmsgprefix)

func init() {
	findAndSaveLocalIP()
}

// Run starts the webserver.
//
// If it fails it the first 3 seconds, Run returns an error
func Run() error {
	var http_err error
	go func() {
		if p := config.GetInt("port"); p > 0 {
			http_err = http.ListenAndServe(fmt.Sprintf(":%d", p), router())
			return
		}
		http_err = fmt.Errorf("port variable is not defined")
	}()
	// wait for an error
	time.Sleep(time.Second * 3)
	return http_err
}

func router() http.Handler {
	r := mux.NewRouter()

	r.NotFoundHandler = http.HandlerFunc(handle_404)

	r.HandleFunc("/", handle_index)
	r.HandleFunc("/description.xml", handle_discorver_xml)
	r.HandleFunc("/api", handleApi)
	r.HandleFunc("/api/{user}", handleUserInfo)
	r.HandleFunc("/api/{user}/lights", handleLights)
	r.HandleFunc("/api/{user}/lights/new", handleNewLights)
	r.HandleFunc("/api/{user}/lights/{light}", handleLightInfo)
	r.HandleFunc("/api/{user}/lights/{light}/state", handleLightState)

	return logRequest(r)
}

// logging middleware
func logRequest(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Printf("%s %s %s\n", r.RemoteAddr, r.Method, r.URL)
		handler.ServeHTTP(w, r)
	})
}

func findAndSaveLocalIP() {
	con, error := net.Dial("udp", "8.8.8.8:80")
	if error != nil {
		log.Fatal(error)
	}
	config.SetString("ip", con.LocalAddr().(*net.UDPAddr).IP.String())
	con.Close()
}
