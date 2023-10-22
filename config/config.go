package config

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
)

var jsonMu sync.RWMutex

// JSONSave saves a struct in a json file.
//
//	file string the file name/path of the json file
//	key  string the json key for indexing the data. It will override existing data if any
//	data any    the data to save
func JSONSave(file, key string, data any) error {
	fileData := make(map[string]any)

	jsonMu.Lock()
	defer jsonMu.Unlock()
	if _, err := os.Stat(file); os.IsNotExist(err) {
		// create dir of file
		path := strings.Split(file, string(os.PathSeparator))
		if dir := strings.Join(path[:len(path)-1], string(os.PathSeparator)); dir != "" {
			if err = os.MkdirAll(dir, 0777); err != nil {
				return err
			}
		}
	} else {
		// read current files
		buf, err := os.ReadFile(file)
		if err != nil {
			return err
		}
		err = json.Unmarshal(buf, &fileData)
		if err != nil {
			return err
		}
	}

	// add data
	fileData[key] = data
	buf, err := json.MarshalIndent(fileData, "", "	")
	if err != nil {
		return err
	}
	return os.WriteFile(file, buf, 0644)
}

// JSONLoad loads a struct from a json file and stores it in data.
//
//		file string the file name/path of the json file
//	 key  string the json key for indexing the data
//	 data any    a pointer to the data to store in
func JSONLoad(file, key string, data any) error {
	jsonMu.RLock()
	defer jsonMu.RUnlock()
	buf, err := os.ReadFile(file)
	if err != nil {
		return err
	}

	fileData := make(map[string]any)
	if err = json.Unmarshal(buf, &fileData); err != nil {
		return err
	}

	fileDataKey, ok := fileData[key]
	if !ok {
		return fmt.Errorf("key '%s' not exists", key)
	}

	buf, _ = json.Marshal(fileDataKey)
	return json.Unmarshal(buf, data)
}

// JSONKeys returns all keys previously saved in the given file. If file is empty or does not exist
// JSONKeys returns an empty slice and err = nil. For all other cases either the saved keys, or an
// error is returned, but not both.
func JSONKeys(file string) (keys []string, err error) {
	jsonMu.RLock()
	defer jsonMu.RUnlock()
	buf, err := os.ReadFile(file)
	if os.IsNotExist(err) {
		return []string{}, nil
	} else if err != nil {
		return []string{}, err
	}

	fileData := make(map[string]any)
	if err = json.Unmarshal(buf, &fileData); err != nil {
		return []string{}, err
	}

	keys = make([]string, 0, len(fileData))
	for k := range fileData {
		keys = append(keys, k)
	}
	return keys, nil
}
