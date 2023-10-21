package config

import (
	"os"
	"testing"
)

type testData struct {
	Key1  string `json:"key1"`
	Key2  int    `json:"key2"`
	Key69 bool   `json:"key69"`
}

var testMap map[string]*testData

func TestMain(m *testing.M) {
	setup()
	code := m.Run()
	// shutdown()
	os.Exit(code)
}

func setup() {
	testMap = make(map[string]*testData)
	testMap["first"] = &testData{"save test", -1, true}
	testMap["second"] = &testData{"another one", 555, true}
	testMap["some"] = &testData{"the last one", 961, false}
}

func TestSave(t *testing.T) {
	type args struct {
		file string
		key  string
		data testData
	}
	type testCase struct {
		name string
		args args
	}

	var tests []testCase
	for k, v := range testMap {
		tests = append(tests,
			testCase{
				name: "Write to test.json",
				args: args{"test.json", k, *v},
			},
			testCase{
				name: "Write to test/test.json",
				args: args{"test/test.json", k, *v},
			},
			testCase{
				name: "Write to test/other/file/save/foo.bar",
				args: args{"test/other/file/save/foo.bar", k, *v},
			},
		)

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := JSONSave(tt.args.file, tt.args.key, tt.args.data); err != nil {
				t.Errorf("Save() error = %v", err)
			}
		})
	}
}

func TestLoad(t *testing.T) {
	type args struct {
		file string
		key  string
		data testData
	}
	type testCase struct {
		name string
		args args
	}
	tests := []testCase{}
	for k, v := range testMap {
		tests = append(tests,
			testCase{
				name: "Read from test.json",
				args: args{"test.json", k, *v},
			},
			testCase{
				name: "Read from test/test.json",
				args: args{"test/test.json", k, *v},
			},
			testCase{
				name: "Read from test/other/file/read/foo.bar",
				args: args{"test/other/file/read/foo.bar", k, *v},
			},
		)

	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := JSONSave(tt.args.file, tt.args.key, tt.args.data); err != nil {
				t.Fatalf("Load() error on Save() = %v", err)
			}
			data := testData{}
			if err := JSONLoad(tt.args.file, tt.args.key, &data); err != nil {
				t.Fatalf("Load() error = %v", err)
			}
			if tt.args.data != data {
				t.Errorf("Load() want: %v; got %v", tt.args.data, data)
			}
		})
	}
}
