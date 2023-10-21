package webserver

import (
	"homeserver/config"
	"homeserver/home"
	"homeserver/webserver/api"
	"html/template"
	"net/http"

	"github.com/gorilla/mux"
)

func handle_404(w http.ResponseWriter, r *http.Request) {
	log.Printf("404 handler was called - req: %+v", r)
	w.Write([]byte("nope.\n404 Not Found"))
}

func handle_index(w http.ResponseWriter, r *http.Request) {
	log.Println("Index site was called")

	w.Write([]byte("Hello World!"))
}

func handle_discorver_xml(w http.ResponseWriter, r *http.Request) {
	tmpl, err := template.ParseFiles("webserver/templates/smart_home/discover.xml")
	if err != nil {
		log.Printf("err on loading template: %+v", err)
		w.WriteHeader(500)
		return
	}

	data := struct {
		IP   string
		Port int
		Mac  string
	}{
		IP:   config.GetString("ip"),
		Port: config.GetInt("port"),
		Mac:  config.GetString("macAddr"),
	}

	w.Header().Set("Content-Type", "text/xml")
	err = tmpl.Execute(w, data)
	if err != nil {
		log.Printf("err on executing template: %+v", err)
		w.WriteHeader(500)
		return
	}
}

func handleApi(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		api.GetAPI(w, r)
	case http.MethodPost:
		api.PostAPI(w, r)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func handleUserInfo(w http.ResponseWriter, r *http.Request) {
	urlVars := mux.Vars(r)
	switch r.Method {
	case http.MethodGet:
		api.GetUserInfo(w, r, urlVars["user"])
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
}

func handleLights(w http.ResponseWriter, r *http.Request) {
	urlVars := mux.Vars(r)
	switch r.Method {
	case http.MethodGet:
		api.GetLights(w, r, urlVars["user"])
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

}

func handleNewLights(w http.ResponseWriter, r *http.Request) {
	switch r.Method {
	case http.MethodGet:
		w.WriteHeader(http.StatusNotImplemented)
		return
	case http.MethodPost:
		if err := home.CloseSmartDeviceAdvertiser(); err != nil {
			log.Printf("ERROR: could not close Smart Device Advertiser: %+v", err)
		}
		home.AdvertiseSmartDevices()
	}
}

func handleLightInfo(w http.ResponseWriter, r *http.Request) {
	urlVars := mux.Vars(r)
	switch r.Method {
	case http.MethodGet:
		api.GetLightInfo(w, r, urlVars["user"], urlVars["light"])
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

}

func handleLightState(w http.ResponseWriter, r *http.Request) {
	urlVars := mux.Vars(r)
	switch r.Method {
	case http.MethodGet:
		api.GetLightState(w, r, urlVars["user"], urlVars["light"])
	case http.MethodPut:
		api.PutLightState(w, r, urlVars["user"], urlVars["light"])
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		return
	}

}
