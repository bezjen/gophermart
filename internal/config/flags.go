package config

import (
	"flag"
	"os"
)

type Config struct {
	RunAddr     string
	LogLevel    string
	DatabaseDSN string
	SecretKey   string
	AccrualAddr string
}

var AppConfig Config

func ParseConfig() {
	flagRunAddr := flag.String("a", "localhost:8080", "address and port to run server")
	flagLogLevel := flag.String("l", "info", "log level")
	flagDatabaseDSN := flag.String("d", "",
		"postgres data source name in format `postgres://username:password@host:port/database_name?sslmode=disable`")
	flagSecretKey := flag.String("s", "", "authorization secret key")
	flagAccrualAddr := flag.String("aa", "", "accrual server address and port")
	flag.Parse()

	addr, addrExists := os.LookupEnv("RUN_ADDRESS")
	if addrExists {
		AppConfig.RunAddr = addr
	} else {
		AppConfig.RunAddr = *flagRunAddr
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
	accrualAddr, accrualAddrExists := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS")
	if accrualAddrExists {
		AppConfig.AccrualAddr = accrualAddr
	} else {
		AppConfig.AccrualAddr = *flagAccrualAddr
	}
}
