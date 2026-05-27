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

	Topics struct {
		Transactions string `env:"TOPIC_TRANSACTIONS"`
		Failed       string `env:"TOPIC_FAILED"`
	}

	Treasury struct {
		InitialBalance int64 `env:"TREASURY_INITIAL_BALANCE,default=1000000000000"`
	}

	Auth struct {
		JWTSecret         string `env:"AUTH_JWT_SECRET"`
		AccessTokenTTLMin int    `env:"AUTH_ACCESS_TTL_MIN,default=60"`
	}
}
