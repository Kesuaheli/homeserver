package api

import (
	cRand "crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"homeserver/config"
	"math/rand"
	"reflect"
	"strings"

	"golang.org/x/exp/slices"
)

const (
	USERFILE  string = "config/users.json"
	LIGHTFILE string = "config/lights.json"
)

type userInfo struct {
	Username    string `json:"username"`
	Displayname string `json:"devicetype"`
}

func newUserFromDisplayname(displayname string) *userInfo {
	usernames, err := config.JSONKeys(USERFILE)
	if err != nil {
		log.Printf("ERROR: could not load usernames from file: %+v", err)
	}
	users := make([]*userInfo, len(usernames))
	for i, un := range usernames {
		users[i] = &userInfo{}
		if err = config.JSONLoad(USERFILE, un, users[i]); err != nil {
			log.Printf("ERROR: could not load user '%s' from file: %+v", un, err)
			return &userInfo{}
		}
		if displayname == users[i].Displayname {
			//user already exists
			return users[i]
		}
	}

	newUsername := getUsername()
	newUser := &userInfo{newUsername, displayname}
	if err = config.JSONSave(USERFILE, newUsername, newUser); err != nil {
		log.Printf("ERROR: could not save user to file: %+v", err)
		return newUser
	}
	log.Printf("registered new user %s (%s)", newUser.Displayname, newUser.Username)
	return newUser
}

// getUsername generates a random username
func getUsername() string {
	n := rand.Intn(16) + 5
	b := make([]byte, n)
	cRand.Read(b)
	return hex.EncodeToString(b)
}

type Light struct {
	index            string
	State            LightState `json:"state"`
	Type             LightType  `json:"type,omitempty"`
	Name             string     `json:"name"`
	ModelID          string     `json:"modelid,omitempty"`
	ManufacturerName string     `json:"manufacturername,omitempty"`
	Productname      string     `json:"productname,omitempty"`
	UniqueID         string     `json:"uniqueid,omitempty"`
	SoftwareVersion  string     `json:"swversion,omitempty"`
}

type LightType string

const (
	LightTypeOnOff            LightType = "On/off light"
	LightTypeDimmable         LightType = "Dimmable light"
	LightTypeColorTemperature LightType = "Color temperature light"
	LightTypeColor            LightType = "Color light"
	LightTypeExtendedColor    LightType = "Extended color light"
)

func (t LightType) getModelID() string {
	switch t {
	case LightTypeDimmable:
		return "LWB010"
	case LightTypeColorTemperature:
		return "LWT010"
	case LightTypeColor:
		return "LST001"
	case LightTypeExtendedColor:
		return "LCT015"
	}
	return ""
}

type LightStateColorMode string

const (
	ColorModeHSV       LightStateColorMode = "hs"
	ColorModeXY        LightStateColorMode = "xy"
	ColorModeColorTemp LightStateColorMode = "ct"
)

type LightState struct {
	On               bool                `json:"on"`
	Brightness       int                 `json:"bri,omitempty"`
	Hue              int                 `json:"hue" jsonreq:"colormode:hs,hsv,huesatval"`
	Saturation       int                 `json:"sat" jsonreq:"colormode:hs"`
	ColorTemperature int                 `json:"ct" jsonreq:"colormode:ct"`
	XY               [2]float32          `json:"xy" jsonreq:"colormode:xy"`
	ColorMode        LightStateColorMode `json:"colormode,omitempty"`
	Alert            string              `json:"alert"`
	Mode             string              `json:"mode"`
	Reachable        bool                `json:"reachable"`
}

func AllLights() (lights map[string]*Light, err error) {
	keys, err := config.JSONKeys(LIGHTFILE)
	if err != nil {
		return nil, err
	}

	lights = make(map[string]*Light, len(keys))
	for _, key := range keys {
		l, err := LightFromID(key)
		if err != nil {
			log.Printf("Error: could not load light '%s', from file: %+v", key, err)
			return nil, err
		}
		lights[key] = l
	}

	return lights, nil
}

func LightFromID(id string) (*Light, error) {
	l := &Light{}
	err := config.JSONLoad(LIGHTFILE, id, l)
	if err != nil {
		err = fmt.Errorf("Error: could not get light '%s': %+v", id, err)
	}
	if l.ModelID == "" {
		l.ModelID = l.Type.getModelID()
	}

	switch l.Type {
	case LightTypeOnOff:
		l.State.Brightness = 0
		l.State.ColorMode = ""
		l.State.ColorTemperature = 0
		l.State.Hue = 0
		l.State.Saturation = 0
	case LightTypeDimmable:
		l.State.ColorMode = ""
		l.State.ColorTemperature = 0
		l.State.Hue = 0
		l.State.Saturation = 0
	case LightTypeColorTemperature:
		l.State.ColorMode = ColorModeColorTemp
		l.State.Hue = 0
		l.State.Saturation = 0
	case LightTypeColor, LightTypeExtendedColor:
		if l.State.ColorMode != ColorModeHSV &&
			l.State.ColorMode != ColorModeXY {
			l.State.ColorMode = ColorModeHSV
		}
		l.State.ColorTemperature = 0
	}

	return l, err
}

// MarshalJSON implements the json.Marshaler interface
func (s *LightState) MarshalJSON() (buf []byte, err error) {
	var (
		refT   reflect.Type  = reflect.TypeOf(s).Elem()
		refV   reflect.Value = reflect.ValueOf(s).Elem()
		format string        = "\"%s\": %s, "
	)
	buf = append(buf, '{')
	for i := 0; i < refT.NumField(); i++ {
		f := refT.Field(i)
		v := refV.Field(i)
		jsonKey := f.Tag.Get("json")
		jsonValue := fmt.Sprint(v)
		if v.Kind() == reflect.String {
			jsonValue = "\"" + v.String() + "\""
		}
		if slices.Contains([]reflect.Kind{reflect.Array, reflect.Func, reflect.Interface, reflect.Map, reflect.Slice, reflect.Struct}, v.Kind()) {
			var vBuf []byte
			vBuf, err = json.Marshal(v.Interface())
			if err != nil {
				return buf, err
			}
			jsonValue = string(vBuf)
		}

		// skip the fields with a `json:"-"` struct tag
		if jsonKey == "-" {
			continue
		}

		if jsonKey == "" {
			jsonKey = f.Name
		}

		jsonKeyParams := strings.Split(jsonKey, ",")
		if len(jsonKeyParams) > 1 {
			jsonKey = jsonKeyParams[0]
			jsonKeyParams = jsonKeyParams[1:]
		} else {
			jsonKeyParams = []string{}
		}

		// skip the fields containing a "omitempty"
		if slices.Contains(jsonKeyParams, "omitempty") && v.IsZero() {
			continue
		}

		require, ok := f.Tag.Lookup("jsonreq")
		if !ok {
			// can already append an continue if no json reqirements are set
			buf = append(buf, []byte(fmt.Sprintf(format, jsonKey, jsonValue))...)
			continue
		}

		n := strings.Index(require, ":")
		if n <= 0 {
			continue
		}
		reqTag := require[:n]
		reqVal := strings.Split(require[n+1:], ",")

		// check the value of the `json:"reqTag"` field
		_, reqV, _ := getFieldByTag(s, "json", reqTag, 0)
		if !slices.Contains(reqVal, fmt.Sprint(reqV)) {
			continue
		}

		buf = append(buf, []byte(fmt.Sprintf(format, jsonKey, jsonValue))...)
	}

	// replace last ',' (comma) with '}'
	if i := getLastIndex(buf, ','); i == -1 {
		buf = append(buf, '}')
	} else {
		buf[i] = '}'
		buf = buf[:i+1]
	}

	return buf, nil
}

func getLastIndex[S ~[]E, E comparable](s S, v E) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == v {
			return i
		}
	}
	return -1
}

// getFieldByTag returns the first field (and its value) of i whoose struct tag 'key' starts with
// tagPrefix.
//
//	i         // The interface
//	key       // The struct key e.g. "myKey" in `myKey:myTag`
//	tagPrefix // The prefix to check for in key's tag
//	skip      // How many fields should be skipped, before start checking
//
// getFieldByTag also returns n as the n'th field from i which mathes tagPrefix. This can bee used
// as skip in another call, if multiple matches are possible.
func getFieldByTag(i any, key, tagPrefix string, skip int) (field reflect.StructField, value reflect.Value, n int) {
	var (
		refT reflect.Type  = reflect.TypeOf(i)
		refV reflect.Value = reflect.ValueOf(i)
	)
	if refT.Kind() == reflect.Pointer {
		refT = refT.Elem()
		refV = refV.Elem()
	}
	if skip >= refT.NumField() {
		return field, value, -1
	}
	for n = skip; n < refT.NumField(); n++ {
		field = refT.Field(n)
		value = refV.Field(n)
		if tag, ok := field.Tag.Lookup(key); ok &&
			strings.HasPrefix(tag, tagPrefix) {
			return field, value, n
		}
	}
	return field, value, -1
}

func (l *Light) Save() {
	config.JSONSave(LIGHTFILE, l.ID(), l)
}

func (l *Light) ID() string {
	var lights map[string]*Light
	var err error
	if lights, err = AllLights(); err != nil {
		return "ID_NOT_FOUND"
	}
	for k, v := range lights {
		if l.Name == v.Name {
			return k
		}
	}
	return "ID_NOT_FOUND"
}

func (l *Light) On() {
	l.State.On = true
	l.Save()
	log.Printf("Light '%s' turned ON", l.Name)
}
func (l *Light) Off() {
	l.State.On = false
	l.Save()
	log.Printf("Light '%s' turned OFF", l.Name)
}

func (l *Light) Brightness(v int) {
	l.State.Brightness = v
	l.Save()
	log.Printf("Light '%s' is set %d%%", l.Name, l.State.Brightness*100/255)
}

func (l *Light) ColorTemperature(v int) {
	l.State.ColorTemperature = v
	// l.State.ColorMode = ColorModeColorTemp
	l.Save()
	log.Printf("Light '%s' is set to temp: %d", l.Name, v)
}

func (l *Light) Hue(v int) {
	l.State.Hue = v
	l.Save()
	log.Printf("Light '%s' is set to hue: %d", l.Name, v)
}

func (l *Light) Saturation(v int) {
	l.State.Saturation = v
	l.Save()
	log.Printf("Light '%s' is set to saturation: %d", l.Name, v)
}
