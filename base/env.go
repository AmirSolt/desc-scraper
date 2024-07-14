package base

import (
	"log"
	"os"
	"strconv"

	"github.com/go-playground/validator/v10"
	"github.com/joho/godotenv"
)

type Env struct {
	DATABASE_URL        string `validate:"required"`
	NUMBER_OF_INSTANCES int    `validate:"required"`
}

func (base *Base) loadEnv() {

	if err := godotenv.Load(".env"); err != nil {
		log.Println("Warrning .env does not exist")
	}

	numOfInstances, errConv := strconv.Atoi(os.Getenv("NUMBER_OF_INSTANCES"))
	if errConv != nil {
		log.Fatal("Error env converting string to int:", errConv)
	}

	env := Env{
		DATABASE_URL:        os.Getenv("DATABASE_URL"),
		NUMBER_OF_INSTANCES: numOfInstances,
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
