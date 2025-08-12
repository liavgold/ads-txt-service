package config

import (
	"fmt"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Port          int           `json:"port"`            
	CacheBackend  string        `json:"cache_backend"`  
	CacheTTL      time.Duration `json:"cache_ttl"`     
	LimiterMaxReq  int           `json:"limiter_max_req"`
	LimmiterTTL  int        	`json:"limiter_ttl"`
	LogLevel      string        `json:"log_level"`      
	HttpClientTO  time.Duration `json:"http_client_to"` 
	RedisAddr     string        `json:"redis_addr"`     
	RedisPassword string        `json:"redis_password"` 
}

var DefaultConfig = Config{
	Port:          8080,
	CacheBackend:  "redis",
	CacheTTL:      300 * time.Second,
	LimmiterTTL:  5,
	LimiterMaxReq: 5,
	LogLevel:      "info",
	HttpClientTO:  10 * time.Second,
	RedisAddr:     "localhost:6379",
	RedisPassword: "",
}


func LoadFromEnv() (*Config, error) {
	cfg := DefaultConfig

	if err := loadEnvVars(&cfg); err != nil {
		return nil, fmt.Errorf("failed to load environment variables: %w", err)
	}

	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}

func loadEnvVars(cfg *Config) error {
	var errs []error

	addError := func(err error) {
		if err != nil {
			errs = append(errs, err)
		}
	}

	if portStr := os.Getenv("PORT"); portStr != "" {
		port, err := strconv.Atoi(portStr)
		addError(err)
		cfg.Port = port
	}

	if cacheBackend := os.Getenv("CACHE_BACKEND"); cacheBackend != "" {
		cfg.CacheBackend = cacheBackend
	}

	if ttlStr := os.Getenv("CACHE_TTL_SECONDS"); ttlStr != "" {
		ttl, err := strconv.Atoi(ttlStr)
		addError(err)
		cfg.CacheTTL = time.Duration(ttl) * time.Second
	}

	if maxReqLimiterStr := os.Getenv("LIMITER_MAX_REQ"); maxReqLimiterStr != "" {
		maxReq, err := strconv.Atoi(maxReqLimiterStr)
		addError(err)
		cfg.LimiterMaxReq = maxReq
	}

	if maxReqLimiterTtlStr := os.Getenv("LIMITER_TTL"); maxReqLimiterTtlStr != "" {
		maxReq, err := strconv.Atoi(maxReqLimiterTtlStr)
		addError(err)
		cfg.LimmiterTTL = maxReq
	}

	if logLevel := os.Getenv("LOG_LEVEL"); logLevel != "" {
		cfg.LogLevel = logLevel
	}

	if timeoutStr := os.Getenv("HTTP_CLIENT_TIMEOUT_SECONDS"); timeoutStr != "" {
		timeout, err := strconv.Atoi(timeoutStr)
		addError(err)
		cfg.HttpClientTO = time.Duration(timeout) * time.Second
	}

	if redisAddr := os.Getenv("REDIS_ADDR"); redisAddr != "" {
		cfg.RedisAddr = redisAddr
	}

	cfg.RedisPassword = os.Getenv("REDIS_PASSWORD")

	if len(errs) > 0 {
		return fmt.Errorf("errors loading environment variables: %v", errs)
	}

	return nil
}

func (c *Config) Validate() error {
	var errs []error

	if c.Port <= 0 || c.Port > 65535 {
		errs = append(errs, fmt.Errorf("port %d is invalid, must be between 1 and 65535", c.Port))
	}

	if c.CacheBackend != "redis" && c.CacheBackend != "memory" {
		errs = append(errs, fmt.Errorf("cache backend %q is unsupported, must be 'redis' or 'memory'", c.CacheBackend))
	}

	if c.CacheTTL <= 0 {
		errs = append(errs, fmt.Errorf("cache TTL %v is invalid, must be positive", c.CacheTTL))
	}

	if c.LimiterMaxReq <= 0 {
		errs = append(errs, fmt.Errorf("max requests per second %d is invalid, must be positive", c.LimiterMaxReq))
	}

	if c.LimmiterTTL <= 0 {
		errs = append(errs, fmt.Errorf("max requests per second %d is invalid, must be positive", c.LimmiterTTL))
	}

	if c.LogLevel != "debug" && c.LogLevel != "info" && c.LogLevel != "warn" && c.LogLevel != "error" {
		errs = append(errs, fmt.Errorf("log level %q is invalid, must be 'debug', 'info', 'warn', or 'error'", c.LogLevel))
	}

	if c.HttpClientTO <= 0 {
		errs = append(errs, fmt.Errorf("HTTP client timeout %v is invalid, must be positive", c.HttpClientTO))
	}

	if c.CacheBackend == "redis" && c.RedisAddr == "" {
		errs = append(errs, fmt.Errorf("redis address is empty but required for redis cache backend"))
	}

	if len(errs) > 0 {
		return fmt.Errorf("validation errors: %v", errs)
	}

	return nil
}
