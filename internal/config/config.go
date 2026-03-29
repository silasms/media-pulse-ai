package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port           string
	WorkerPoolSize int
	QueueCapacity  int
	AIProvider     string
	AIApiKey       string
	RedisAddr      string
	RedisPassword  string
	RedisDB        int
	RedisEnabled   bool
	CacheTTL       time.Duration
	RateLimitRPS   int
	LogLevel       string
}

func Load() *Config {
	return &Config{
		Port:           getEnv("PORT", "8080"),
		WorkerPoolSize: getEnvInt("WORKER_POOL_SIZE", 5),
		QueueCapacity:  getEnvInt("QUEUE_CAPACITY", 100),
		AIProvider:     getEnv("AI_PROVIDER", "mock"),
		AIApiKey:       getEnv("AI_API_KEY", ""),
		RedisAddr:      getEnv("REDIS_ADDR", "localhost:6379"),
		RedisPassword:  getEnv("REDIS_PASSWORD", ""),
		RedisDB:        getEnvInt("REDIS_DB", 0),
		RedisEnabled:   getEnvBool("REDIS_ENABLED", false),
		CacheTTL:       time.Duration(getEnvInt("CACHE_TTL_MINUTES", 15)) * time.Minute,
		RateLimitRPS:   getEnvInt("RATE_LIMIT_RPS", 50),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
	}
}

func getEnv(key, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val := os.Getenv(key); val != "" {
		if i, err := strconv.Atoi(val); err == nil {
			return i
		}
	}
	return defaultVal
}

func getEnvBool(key string, defaultVal bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return defaultVal
}
