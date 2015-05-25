package config

import (
	"os"
	"testing"

	"github.com/ptrost/mosaic2go/test"
)

func TestConfigGet(t *testing.T) {
	config := New("../test_fixtures/config.json")
	defer os.Remove("../test_fixtures/config.json")
	expected := "value1"
	result := config.Get("key1")

	test.Assert("Config.Get", expected, result, t)
}

func TestConfigSet(t *testing.T) {
	config := New("../test_fixtures/config.json")
	defer os.Remove("../test_fixtures/config.json")
	config.Set("newkey", "newval")
	expected := "newval"
	result := config.Get("newkey")

	test.Assert("Config.Set", expected, result, t)
}

func TestConfigGetUnknownKey(t *testing.T) {
	config := New("../test_fixtures/config.json")
	defer func() {
		os.Remove("../test_fixtures/config.json")
		if r := recover(); r == nil {
			t.Error("Config.Get: Didn't panic for unknown key.")
		}
	}()
	_ = config.Get("unknownkey")
}

func TestCreateConfigFile(t *testing.T) {
	file := "../test_fixtures/config.json"
	if _, err := os.Stat(file); !os.IsNotExist(err) {
		t.Fatal("File test_fixtures/config.json must not exist before testing createConfigFile().")
	}
	defer os.Remove("../test_fixtures/config.json")
	createConfigFile(file)

	if _, err := os.Stat(file); os.IsNotExist(err) {
		t.Fatalf("Config.createConfigFile: Didn't create file %s", file)
	}
}
