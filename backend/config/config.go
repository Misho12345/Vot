package config

import "os"

var (
	DbHost     = os.Getenv("DB_HOST")
	DbUser     = os.Getenv("DB_USER")
	DbPassword = os.Getenv("DB_PASSWORD")
	DbName     = os.Getenv("DB_NAME")
	JwtSecret  = []byte(os.Getenv("JWT_SECRET"))
)
