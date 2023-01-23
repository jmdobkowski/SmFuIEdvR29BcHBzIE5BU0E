package config

import (
	"log"
	"os"
	"strconv"
)

type Config struct {
	ApiKey             string
	ConcurrentRequests int
	Port               int
}

// LoadFromEnv loads the configuration from environment variables
// If the variable is not set default value is used.
// If the variable is set the value must be valid.
func LoadFromEnv() Config {
	cfg := Config{
		ApiKey:             "DEMO_KEY",
		ConcurrentRequests: 5,
		Port:               8080,
	}

	if v, ok := os.LookupEnv("API_KEY"); ok {
		cfg.ApiKey = v
		log.Printf("read API_KEY = %s", v)
	}

	if v, ok := os.LookupEnv("CONCURRENT_REQUESTS"); ok {
		i, err := strconv.Atoi(v)
		if err != nil {
			panic("could not parse CONCURRENT_REQUESTS = '" + v + "'")
		}
		log.Printf("read CONCURRENT_REQUESTS = %d", i)
		cfg.ConcurrentRequests = i
	}

	if v, ok := os.LookupEnv("PORT"); ok {
		i, err := strconv.Atoi(v)
		if err != nil {
			panic("could not parse PORT = '" + v + "'")
		}
		if i < 0 || i > 65536 {
			panic("PORT = '" + v + "' out of range (0, 65536)")
		}
		log.Printf("read PORT = %d", i)
		cfg.Port = i
	}

	return cfg
}
