package tests

import (
	"bytes"
	"encoding/json"
	"github.com/joho/godotenv"
	"path/filepath"
	"runtime"
)

func GetLocalPath(file string) string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), file)
}

func ToJSON(target interface{}) string {
	str := new(bytes.Buffer)
	encoder := json.NewEncoder(str)
	encoder.SetEscapeHTML(false)
	encoder.SetIndent("", "    ")
	err := encoder.Encode(target)
	if err != nil {
		return err.Error()
	}

	return str.String()
}

func LoadTestEnv() error {
	return godotenv.Load(GetLocalPath("../.env"))
}
