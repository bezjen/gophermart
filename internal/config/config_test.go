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
				RunAddr:     "localhost:8080",
				LogLevel:    "info",
				DatabaseDSN: "",
				SecretKey:   "",
				AccrualAddr: "",
			},
		},
		{
			name: "Flags only",
			args: []string{"gophermart.exe", "-a=localhost1:8081"},
			env:  map[string]string{},
			expectedConfig: Config{
				RunAddr:     "localhost1:8081",
				LogLevel:    "info",
				DatabaseDSN: "",
				SecretKey:   "",
				AccrualAddr: "",
			},
		},
		{
			name: "Environment variables only",
			args: []string{"gophermart.exe"},
			env: map[string]string{
				"RUN_ADDRESS": "localhost1:8081",
			},
			expectedConfig: Config{
				RunAddr:     "localhost1:8081",
				LogLevel:    "info",
				DatabaseDSN: "",
				SecretKey:   "",
				AccrualAddr: "",
			},
		},
		{
			name: "Env for address, flag for log level",
			args: []string{"gophermart.exe", "-l=warn"},
			env: map[string]string{
				"RUN_ADDRESS": "localhost1:8081",
			},
			expectedConfig: Config{
				RunAddr:     "localhost1:8081",
				LogLevel:    "warn",
				DatabaseDSN: "",
				SecretKey:   "",
				AccrualAddr: "",
			},
		},
		{
			name: "Both environment and flags (use env)",
			args: []string{"gophermart.exe", "-a=localhost1:8081"},
			env: map[string]string{
				"RUN_ADDRESS": "localhost1:8081",
			},
			expectedConfig: Config{
				RunAddr:     "localhost1:8081",
				LogLevel:    "info",
				DatabaseDSN: "",
				SecretKey:   "",
				AccrualAddr: "",
			},
		},
		{
			name: "Env for log level",
			args: []string{"gophermart.exe"},
			env: map[string]string{
				"LOG_LEVEL": "fatal",
			},
			expectedConfig: Config{
				RunAddr:     "localhost:8080",
				LogLevel:    "fatal",
				DatabaseDSN: "",
				SecretKey:   "",
				AccrualAddr: "",
			},
		},
		{
			name: "Flag for log level",
			args: []string{"gophermart.exe", "-l=fatal"},
			env:  map[string]string{},
			expectedConfig: Config{
				RunAddr:     "localhost:8080",
				LogLevel:    "fatal",
				DatabaseDSN: "",
				SecretKey:   "",
				AccrualAddr: "",
			},
		},
		{
			name: "Env for data source name",
			args: []string{"gophermart.exe"},
			env: map[string]string{
				"DATABASE_URI": "ds",
			},
			expectedConfig: Config{
				RunAddr:     "localhost:8080",
				LogLevel:    "info",
				DatabaseDSN: "ds",
				SecretKey:   "",
				AccrualAddr: "",
			},
		},
		{
			name: "Flag for data source name",
			args: []string{"gophermart.exe", "-d=ds"},
			env:  map[string]string{},
			expectedConfig: Config{
				RunAddr:     "localhost:8080",
				LogLevel:    "info",
				DatabaseDSN: "ds",
				SecretKey:   "",
				AccrualAddr: "",
			},
		},
		{
			name: "Env for secret key",
			args: []string{"gophermart.exe"},
			env: map[string]string{
				"SECRET_KEY": "secret_key",
			},
			expectedConfig: Config{
				RunAddr:     "localhost:8080",
				LogLevel:    "info",
				DatabaseDSN: "",
				SecretKey:   "secret_key",
				AccrualAddr: "",
			},
		},
		{
			name: "Flag for secret key",
			args: []string{"gophermart.exe", "-s=secret_key1"},
			env:  map[string]string{},
			expectedConfig: Config{
				RunAddr:     "localhost:8080",
				LogLevel:    "info",
				DatabaseDSN: "",
				SecretKey:   "secret_key1",
				AccrualAddr: "",
			},
		},
		{
			name: "Env for accrual system address",
			args: []string{"gophermart.exe"},
			env: map[string]string{
				"ACCRUAL_SYSTEM_ADDRESS": "localhost:8081",
			},
			expectedConfig: Config{
				RunAddr:     "localhost:8080",
				LogLevel:    "info",
				DatabaseDSN: "",
				SecretKey:   "",
				AccrualAddr: "localhost:8081",
			},
		},
		{
			name: "Flag for accrual system address",
			args: []string{"gophermart.exe", "-aa=localhost:8082"},
			env:  map[string]string{},
			expectedConfig: Config{
				RunAddr:     "localhost:8080",
				LogLevel:    "info",
				DatabaseDSN: "",
				SecretKey:   "",
				AccrualAddr: "localhost:8082",
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
