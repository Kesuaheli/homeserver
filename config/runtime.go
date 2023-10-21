package config

import "sync"

var strConf map[string]string
var intConf map[string]int
var confMu sync.RWMutex

func init() {
	strConf = make(map[string]string)
	intConf = make(map[string]int)
}

// SetString saves a global string variable to use anywhere with GetString(k)
func SetString(k, v string) {
	confMu.Lock()
	strConf[k] = v
	confMu.Unlock()
}

// GetString returns a previously saved string variable with SetString(k, v)
func GetString(k string) (v string) {
	confMu.Lock()
	confMu.Unlock()
	return strConf[k]
}

// SetInt saves a global integer variable to use anywhere with GetInt(k)
func SetInt(k string, v int) {
	confMu.Lock()
	intConf[k] = v
	confMu.Unlock()
}

// GetString returns a previously saved integer variable with SetInt(k, v)
func GetInt(k string) (v int) {
	confMu.Lock()
	defer confMu.Unlock()
	return intConf[k]
}
