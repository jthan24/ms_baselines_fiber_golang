package config

import (
	"log"
	"os"
	"testing"

	"bou.ke/monkey"
	"github.com/stretchr/testify/assert"
)

func cleanEnv() {
	os.Clearenv()
}

func setRequiredEnvs() {
	os.Setenv("SERVICE_NAME", "ms-baselines-golang")
	os.Setenv(
		"DB_CONNECTION_STRING",
		"myConnectionString",
	)
}

func testRequired(t *testing.T) {
	fakeLogFatal := func(msg ...interface{}) {
		panic("log.Fatal called")
	}
	patch := monkey.Patch(log.Fatal, fakeLogFatal)
	defer patch.Unpatch()
	setRequiredEnvs()
	GetConfig()
}

func testRequiredFail(t *testing.T) {
	c = nil
	fakeLogFatal := func(msg ...interface{}) {
		panic("log.Fatal called")
	}
	patch := monkey.Patch(log.Fatal, fakeLogFatal)
	defer patch.Unpatch()
	assert.Panics(t, func() { GetConfig() }, "Not panic")
}

func TestController(t *testing.T) {
	fs := map[string]func(*testing.T){
		"testRequired":     testRequired,
		"testRequiredFail": testRequiredFail,
	}
	for name, f := range fs {
		cleanEnv()
		t.Run(name, f)
	}
}
