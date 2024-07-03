package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

func wipeReports() {

	conn := loadDB()
	defer conn.Close(context.Background())

	content := "DELETE FROM report_events; DELETE FROM reports;"

	response, err := conn.Exec(context.Background(), content)
	if err != nil {
		log.Fatal("Error db:", err)
	}

	fmt.Println("------")
	fmt.Println(response)
	fmt.Println("------")

}

func pushToDB() {

	conn := loadDB()
	defer conn.Close(context.Background())

	schemaFilePath := "models/sql/schema.sql"
	contentSchema, errSchema := os.ReadFile(schemaFilePath)
	if errSchema != nil {
		log.Fatal(errSchema)
	}
	response, err := conn.Exec(context.Background(), string(contentSchema))
	if err != nil {
		log.Fatal("Error db:", err)
	}

	fmt.Println("------")
	fmt.Println(response)
	fmt.Println("------")

}

func resetDb() {

	ctx := context.Background()

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error db:", err)
	}

	config, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		log.Fatalf("Unable to parse configuration: %v", err)
	}

	// Open connection pool
	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		log.Fatalf("Unable to connect to database: %v", err)
	}
	defer pool.Close()

	// execResponseDrop, err := pool.Exec(ctx, "DROP DATABASE IF EXISTS postgres")
	execResponseDrop, err := pool.Exec(ctx, "DROP SCHEMA public CASCADE;")
	if err != nil {
		log.Fatalf("error dropping database: %v", err)
	}

	fmt.Println("------")
	fmt.Println(execResponseDrop)
	fmt.Println("------")

	// Create the database
	// execResponseCreate, err := pool.Exec(ctx, "CREATE DATABASE postgres")
	execResponseCreate, err := pool.Exec(ctx, "CREATE SCHEMA public;")
	if err != nil {
		log.Fatalf("error creating database: %v", err)
	}

	fmt.Println("------")
	fmt.Println(execResponseCreate)
	fmt.Println("------")

}

func loadDB() *pgx.Conn {

	if err := godotenv.Load(".env"); err != nil {
		log.Fatal("Error db:", err)
	}

	conn, dbErr := pgx.Connect(context.Background(), os.Getenv("DATABASE_URL"))
	if dbErr != nil {
		log.Fatalln("Error db:", dbErr)
	}

	// defer conn.Close(context.Background())

	// conn.Exec(context.Background(), "")

	return conn
}
