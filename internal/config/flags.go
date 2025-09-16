package config

import (
	"flag"
	"os"
)

type Config struct {
	ServerHost  string
	ServerPort  string
	LogLevel    string
	DatabaseDSN string
	SecretKey   string
}

var AppConfig Config

func ParseConfig() {
	flagServerHost := flag.String("h", "localhost", "host")
	flagServerPort := flag.String("p", "8080", "port")
	flagLogLevel := flag.String("l", "info", "log level")
	flagDatabaseDSN := flag.String("d", "",
		"postgres data source name in format `postgres://username:password@host:port/database_name?sslmode=disable`")
	flagSecretKey := flag.String("s", "", "authorization secret key")
	flag.Parse()

	host, hostExists := os.LookupEnv("SERVER_HOST")
	if hostExists {
		AppConfig.ServerHost = host
	} else {
		AppConfig.ServerHost = *flagServerHost
	}
	port, portExists := os.LookupEnv("SERVER_PORT")
	if portExists {
		AppConfig.ServerPort = port
	} else {
		AppConfig.ServerPort = *flagServerPort
	}
	logLevel, logLevelExists := os.LookupEnv("LOG_LEVEL")
	if logLevelExists {
		AppConfig.LogLevel = logLevel
	} else {
		AppConfig.LogLevel = *flagLogLevel
	}
	databaseDSN, databaseDSNExists := os.LookupEnv("DATABASE_DSN")
	if databaseDSNExists {
		AppConfig.DatabaseDSN = databaseDSN
	} else {
		AppConfig.DatabaseDSN = *flagDatabaseDSN
	}
	secretKey, secretKeyExists := os.LookupEnv("SECRET_KEY")
	if secretKeyExists {
		AppConfig.SecretKey = secretKey
	} else if *flagSecretKey != "" {
		AppConfig.SecretKey = *flagSecretKey
	}
}
