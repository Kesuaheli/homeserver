package api

import (
	"encoding/json"
	"fmt"
	"homeserver/config"
	logger "log"
	"net/http"
)

var log *logger.Logger = logger.New(logger.Writer(), "[API] ", logger.LstdFlags|logger.Lmsgprefix)

type successResponse struct {
	R any `json:"success"`
}
type errorResponse struct {
	Err apiError `json:"error"`
}
type apiError struct {
	Type        int    `json:"type"`
	Address     string `json:"address"`
	Description string `json:"description"`
}

func GetAPI(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(404)
}

func PostAPI(w http.ResponseWriter, r *http.Request) {
	buf := make([]byte, 1000)
	n, _ := r.Body.Read(buf)
	buf = buf[:n]

	reqUser := &userInfo{}
	if err := json.Unmarshal(buf, reqUser); err != nil {
		log.Printf("Error: could not parse reqested user: %+v", err)
		respondError(w, http.StatusBadRequest, errorResponse{apiError{
			Type:        2,
			Address:     "/",
			Description: "body contains invalid json",
		}})
		return
	}

	user := newUserFromDisplayname(reqUser.Displayname)
	resp := []any{
		successResponse{user},
	}
	buf, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error: could not marshal user response: %+v err: %+v", resp, err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	w.Write(buf)
}

func GetUserInfo(w http.ResponseWriter, r *http.Request, user string) {
	w.WriteHeader(http.StatusNotImplemented)
}

func GetLights(w http.ResponseWriter, r *http.Request, user string) {
	//verify user
	u := &userInfo{}
	if err := config.JSONLoad(USERFILE, user, u); err != nil {
		log.Printf("Error: could not get user: %+v", err)
		respondError(w, http.StatusBadRequest, errorResponse{apiError{
			Type:        7,
			Address:     "/username/",
			Description: fmt.Sprintf("invalid value, %s, for parameter, username", user),
		}})
		return
	}

	// respond with lights
	lights, err := AllLights()
	if err != nil {
		log.Printf("ERROR: could not get all lights: %+v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	buf, err := json.Marshal(lights)
	if err != nil {
		log.Printf("Error: could not marshal light response: %+v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(buf)
}

func GetLightInfo(w http.ResponseWriter, r *http.Request, user, light string) {
	// verify user
	u := &userInfo{}
	if err := config.JSONLoad(USERFILE, user, u); err != nil {
		log.Printf("Error: could not get user: %+v", err)
		respondError(w, http.StatusBadRequest, errorResponse{apiError{
			Type:        7,
			Address:     "/username/",
			Description: fmt.Sprintf("invalid value, %s, for parameter, username", user),
		}})
		return
	}

	// respond with light
	l, err := LightFromID(light)
	if err != nil {
		log.Printf("Error: could not get light: %+v", err)
		respondError(w, http.StatusBadRequest, errorResponse{apiError{
			Type:        7,
			Address:     "/username/lights/id",
			Description: fmt.Sprintf("invalid value, %s, for parameter, id", light),
		}})
		return
	}
	buf, err := json.Marshal(l)
	if err != nil {
		log.Printf("Error: could not marshal light response: %+v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(buf)
}

func GetLightState(w http.ResponseWriter, r *http.Request, user, light string) {

}

func PutLightState(w http.ResponseWriter, r *http.Request, user, light string) {
	// verify body
	buf := make([]byte, 1000)
	n, _ := r.Body.Read(buf)
	buf = buf[:n]

	newLightState := &LightState{}
	if err := json.Unmarshal(buf, newLightState); err != nil {
		log.Printf("ERROR: could not parse body to lightstate: %+v", err)
		respondError(w, http.StatusBadRequest, errorResponse{apiError{
			Type:        2,
			Address:     "/",
			Description: "body contains invalid json",
		}})
		return
	}

	// verify user
	u := &userInfo{}
	if err := config.JSONLoad(USERFILE, user, u); err != nil {
		log.Printf("Error: could not get user: %+v", err)
		respondError(w, http.StatusBadRequest, errorResponse{apiError{
			Type:        7,
			Address:     "/username/",
			Description: fmt.Sprintf("invalid value, %s, for parameter, username", user),
		}})
		return
	}

	// get light
	l, err := LightFromID(light)
	if err != nil {
		log.Printf("Error: could not get light: %+v", err)
		respondError(w, http.StatusBadRequest, errorResponse{apiError{
			Type:        7,
			Address:     "/username/lights/id",
			Description: fmt.Sprintf("invalid value, %s, for parameter, id", light),
		}})
		return
	}

	log.Printf("Got new light state:\n%+v", string(buf))

	var resp []any

	if newLightState.On != l.State.On {
		if newLightState.On {
			l.On()
		} else {
			l.Off()
		}
		confirm := make(map[string]any)
		confirm[fmt.Sprintf("/lights/%s/on", light)] = newLightState.On
		resp = append(resp, successResponse{confirm})
	}
	if l.Type != LightTypeOnOff &&
		newLightState.Brightness > 1 &&
		newLightState.Brightness != l.State.Brightness {
		l.Brightness(newLightState.Brightness)
		confirm := make(map[string]any)
		confirm[fmt.Sprintf("/lights/%s/bri", light)] = newLightState.Brightness
		resp = append(resp, successResponse{confirm})
	}
	if l.Type == LightTypeColorTemperature &&
		newLightState.ColorTemperature > 0 &&
		newLightState.ColorTemperature != l.State.ColorTemperature {
		l.ColorTemperature(newLightState.ColorTemperature)
		confirm := make(map[string]any)
		confirm[fmt.Sprintf("/lights/%s/ct", light)] = newLightState.ColorTemperature
		resp = append(resp, successResponse{confirm})
	}
	if l.Type == LightTypeColor || l.Type == LightTypeExtendedColor {
		if newLightState.Hue != l.State.Hue {
			l.Hue(newLightState.Hue)
			confirm := make(map[string]any)
			confirm[fmt.Sprintf("/lights/%s/hue", light)] = newLightState.Hue
			resp = append(resp, successResponse{confirm})
		}
		if newLightState.Saturation != l.State.Saturation {
			l.Saturation(newLightState.Saturation)
			confirm := make(map[string]any)
			confirm[fmt.Sprintf("/lights/%s/sat", light)] = newLightState.Saturation
			resp = append(resp, successResponse{confirm})
		}
	}

	buf, err = json.Marshal(resp)
	if err != nil {
		log.Printf("Error: could not marshal light state response: %+v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.Write(buf)
}

func respondError(w http.ResponseWriter, statusCode int, errors ...errorResponse) {
	var resp []errorResponse
	resp = append(resp, errors...)

	buf, err := json.Marshal(resp)
	if err != nil {
		log.Printf("Error: could not marshal error response: %+v", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(statusCode)
	w.Write(buf)
}
