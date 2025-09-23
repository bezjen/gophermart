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

	if *flagRunAddr != "localhost:8080" {
		AppConfig.RunAddr = *flagRunAddr
	} else if addr, addrExists := os.LookupEnv("RUN_ADDRESS"); addrExists {
		AppConfig.RunAddr = addr
	} else {
		AppConfig.RunAddr = *flagRunAddr
	}

	if *flagLogLevel != "info" {
		AppConfig.LogLevel = *flagLogLevel
	} else if logLevel, logLevelExists := os.LookupEnv("LOG_LEVEL"); logLevelExists {
		AppConfig.LogLevel = logLevel
	} else {
		AppConfig.LogLevel = *flagLogLevel
	}

	if *flagDatabaseDSN != "" {
		AppConfig.DatabaseDSN = *flagDatabaseDSN
	} else if databaseDSN, databaseDSNExists := os.LookupEnv("DATABASE_URI"); databaseDSNExists {
		AppConfig.DatabaseDSN = databaseDSN
	} else {
		AppConfig.DatabaseDSN = *flagDatabaseDSN
	}

	if *flagSecretKey != "" {
		AppConfig.SecretKey = *flagSecretKey
	} else if secretKey, secretKeyExists := os.LookupEnv("SECRET_KEY"); secretKeyExists {
		AppConfig.SecretKey = secretKey
	} else {
		AppConfig.SecretKey = *flagSecretKey
	}

	if *flagAccrualAddr != "" {
		AppConfig.AccrualAddr = *flagAccrualAddr
	} else if accrualAddr, accrualAddrExists := os.LookupEnv("ACCRUAL_SYSTEM_ADDRESS"); accrualAddrExists {
		AppConfig.AccrualAddr = accrualAddr
	} else {
		AppConfig.AccrualAddr = *flagAccrualAddr
	}
}
