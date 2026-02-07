package config

import (
	"fmt"
	"log"
	"path/filepath"
	"runtime"
	"testing"

	env "github.com/Netflix/go-env"
	"github.com/joho/godotenv"
)

func init() {
	_, thisFile, _, _ := runtime.Caller(0)
	basepath := filepath.Dir(thisFile)

	if testing.Testing() {
		Load(fmt.Sprintf("%v/../.env.testing", basepath))
	} else {
		Load(fmt.Sprintf("%v/../.env", basepath))
	}
}

func Load(envFilePath string) {
	if err := godotenv.Load(envFilePath); err != nil {
		log.Printf("Error on loading .env file from %v: %+v\n", envFilePath, err)
	}

	if _, err := env.UnmarshalFromEnviron(&Env); err != nil {
		log.Fatalf("Error on unmarshaling .env file: %+v\n", err)
	}
}

var Env struct {
	App struct {
		Name        string `env:"APP_NAME"`
		Environment string `env:"APP_ENV"`
		Debug       bool   `env:"APP_DEBUG"`
		Port        string `env:"APP_PORT,default=3000"`
	}

	Doc struct {
		Auth struct {
			Username string `env:"DOC_AUTH_USERNAME"`
			Password string `env:"DOC_AUTH_PASSWORD"`
		}
	}
}
