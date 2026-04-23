package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AllureBaseURL   string
	AllureToken     string
	RequestTimeout  time.Duration
}

func Load() *Config {
	timeoutStr := os.Getenv("REQUEST_TIMEOUT")
	if timeoutStr == "" {
		timeoutStr = "30"
	}

	timeoutSec, err := strconv.Atoi(timeoutStr)
	if err != nil || timeoutSec <= 0 {
		timeoutSec = 30
	}

	return &Config{
		AllureBaseURL:  os.Getenv("ALLURE_BASE_URL"),
		AllureToken:    os.Getenv("ALLURE_TOKEN"),
		RequestTimeout: time.Duration(timeoutSec) * time.Second,
	}
}
