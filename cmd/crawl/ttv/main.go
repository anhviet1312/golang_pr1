package main

import "github.com/joho/godotenv"

func init() {
	godotenv.Load("../../.env") // for develop
	godotenv.Load("./.env")     // for production
}

func main() {

}
