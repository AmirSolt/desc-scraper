package base

import (
	"log"
	"os"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Env struct {
	DOMAIN          string `validate:"url"`
	USER_SERVER_URL string `validate:"url"`
	FRONTEND_URL    string `validate:"url"`
	IS_PROD         bool   `validate:"boolean"`
	DATABASE_URL    string `validate:"url"`
	SECRET_API_KEY  string `validate:"required"`
	GLITCHTIP_DSN   string `validate:"required"`
}

func (base *Base) loadEnv() {

	if err := godotenv.Load(".env"); err != nil {

	}
	env := Env{
		DOMAIN:          os.Getenv("DOMAIN"),
		USER_SERVER_URL: os.Getenv("USER_SERVER_URL"),
		FRONTEND_URL:    os.Getenv("FRONTEND_URL"),
		IS_PROD:         strToBool(os.Getenv("IS_PROD")),
		DATABASE_URL:    os.Getenv("DATABASE_URL"),
		SECRET_API_KEY:  os.Getenv("SECRET_API_KEY"),
		GLITCHTIP_DSN:   os.Getenv("GLITCHTIP_DSN"),
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
