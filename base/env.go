package base

import (
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Env struct {
	DATABASE_URL   string `validate:"required"`
	MEMCACHED_URL  string `validate:"required"`
	SECRET_API_KEY string `validate:"required"`
}

func (base *Base) loadEnv() {

	if err := godotenv.Load(".env"); err != nil {

	}
	env := Env{
		DATABASE_URL:   os.Getenv("DATABASE_URL"),
		MEMCACHED_URL:  os.Getenv("MEMCACHED_URL"),
		SECRET_API_KEY: os.Getenv("SECRET_API_KEY"),
	}

	validate := validator.New(validator.WithRequiredStructEnabled())
	err := validate.Struct(env)
	if err != nil {
		log.Fatal("Error .env:", err)
	}

	base.Env = &env
}

func strToBool(s string) bool {
	return s == "true"
}
