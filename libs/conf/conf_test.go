package conf

import (
	"io/ioutil"
	"os"
	"testing"
)

func writeTempToml(content string) (string, error) {
	tmpfile, err := ioutil.TempFile("", "testconf-*.toml")
	if err != nil {
		return "", err
	}
	if _, err := tmpfile.Write([]byte(content)); err != nil {
		tmpfile.Close()
		return "", err
	}
	_ = tmpfile.Close()
	return tmpfile.Name(), nil
}

func TestInitAndGet(t *testing.T) {
	devContent := "key1 = 'dev_value'"
	prodContent := "key1 = 'prod_value'"
	devPath, err := writeTempToml(devContent)
	if err != nil {
		t.Fatalf("Failed to create dev toml: %v", err)
	}
	defer os.Remove(devPath)
	prodPath, err := writeTempToml(prodContent)
	if err != nil {
		t.Fatalf("Failed to create prod toml: %v", err)
	}
	defer os.Remove(prodPath)

	os.Setenv("ISDEBUG", "1")
	os.Setenv("devConf", devPath)
	os.Unsetenv("prodConf")
	config = nil
	Init()
	if v := Get("key1"); string(v) != "dev_value" {
		t.Errorf("Expected dev_value, got %v", string(v))
	}

	os.Unsetenv("ISDEBUG")
	os.Setenv("prodConf", prodPath)
	os.Unsetenv("devConf")
	config = nil
	Init()
	if v := Get("key1"); string(v) != "prod_value" {
		t.Errorf("Expected prod_value, got %v", string(v))
	}
}
