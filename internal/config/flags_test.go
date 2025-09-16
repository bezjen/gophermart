package config

import (
	"flag"
	"os"
	"testing"
)

func TestParseFlags(t *testing.T) {
	tests := []struct {
		name           string
		args           []string
		env            map[string]string
		expectedConfig Config
	}{
		{
			name: "Default values",
			args: []string{"gophermart.exe"},
			env:  map[string]string{},
			expectedConfig: Config{
				ServerHost:  "localhost",
				ServerPort:  "8080",
				LogLevel:    "info",
				DatabaseDSN: "",
				SecretKey:   "",
			},
		},
		{
			name: "Flags only",
			args: []string{"gophermart.exe", "-h=localhost1", "-p=8081"},
			env:  map[string]string{},
			expectedConfig: Config{
				ServerHost:  "localhost1",
				ServerPort:  "8081",
				LogLevel:    "info",
				DatabaseDSN: "",
				SecretKey:   "",
			},
		},
		{
			name: "Environment variables only",
			args: []string{"gophermart.exe"},
			env: map[string]string{
				"SERVER_HOST": "gophermart",
				"SERVER_PORT": "8081",
			},
			expectedConfig: Config{
				ServerHost:  "gophermart",
				ServerPort:  "8081",
				LogLevel:    "info",
				DatabaseDSN: "",
				SecretKey:   "",
			},
		},
		{
			name: "Env for host, flag for port",
			args: []string{"gophermart.exe", "-p=8081"},
			env: map[string]string{
				"SERVER_HOST": "gophermart",
			},
			expectedConfig: Config{
				ServerHost:  "gophermart",
				ServerPort:  "8081",
				LogLevel:    "info",
				DatabaseDSN: "",
				SecretKey:   "",
			},
		},
		{
			name: "Both environment and flags (use env)",
			args: []string{"gophermart.exe", "-h=localhost2", "-p=8082"},
			env: map[string]string{
				"SERVER_HOST": "gophermart",
				"SERVER_PORT": "8081",
			},
			expectedConfig: Config{
				ServerHost:  "gophermart",
				ServerPort:  "8081",
				LogLevel:    "info",
				DatabaseDSN: "",
				SecretKey:   "",
			},
		},
		{
			name: "Env for log level",
			args: []string{"gophermart.exe"},
			env: map[string]string{
				"LOG_LEVEL": "fatal",
			},
			expectedConfig: Config{
				ServerHost:  "localhost",
				ServerPort:  "8080",
				LogLevel:    "fatal",
				DatabaseDSN: "",
				SecretKey:   "",
			},
		},
		{
			name: "Flag for log level",
			args: []string{"gophermart.exe", "-l=fatal"},
			env:  map[string]string{},
			expectedConfig: Config{
				ServerHost:  "localhost",
				ServerPort:  "8080",
				LogLevel:    "fatal",
				DatabaseDSN: "",
				SecretKey:   "",
			},
		},
		{
			name: "Env for data source name",
			args: []string{"gophermart.exe"},
			env: map[string]string{
				"DATABASE_DSN": "ds",
			},
			expectedConfig: Config{
				ServerHost:  "localhost",
				ServerPort:  "8080",
				LogLevel:    "info",
				DatabaseDSN: "ds",
				SecretKey:   "",
			},
		},
		{
			name: "Flag for data source name",
			args: []string{"gophermart.exe", "-d=ds"},
			env:  map[string]string{},
			expectedConfig: Config{
				ServerHost:  "localhost",
				ServerPort:  "8080",
				LogLevel:    "info",
				DatabaseDSN: "ds",
				SecretKey:   "",
			},
		},
		{
			name: "Env for secret key",
			args: []string{"gophermart.exe"},
			env: map[string]string{
				"SECRET_KEY": "secret_key",
			},
			expectedConfig: Config{
				ServerHost:  "localhost",
				ServerPort:  "8080",
				LogLevel:    "info",
				DatabaseDSN: "",
				SecretKey:   "secret_key",
			},
		},
		{
			name: "Flag for secret key",
			args: []string{"gophermart.exe", "-s=secret_key1"},
			env:  map[string]string{},
			expectedConfig: Config{
				ServerHost:  "localhost",
				ServerPort:  "8080",
				LogLevel:    "info",
				DatabaseDSN: "",
				SecretKey:   "secret_key1",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			os.Args = tt.args
			for key, val := range tt.env {
				err := os.Setenv(key, val)
				if err != nil {
					t.Errorf("Failed to set env, error: %v", err)
					return
				}
			}

			flag.CommandLine = flag.NewFlagSet(os.Args[0], flag.ExitOnError)
			AppConfig = Config{}

			ParseConfig()

			if AppConfig != tt.expectedConfig {
				t.Errorf("Expected %+v, got %+v", tt.expectedConfig, AppConfig)
			}

			for key := range tt.env {
				err := os.Unsetenv(key)
				if err != nil {
					t.Errorf("Failed to unset env, error: %v", err)
					return
				}
			}
		})
	}
}
