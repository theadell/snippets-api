package config

import (
	"fmt"
	"log/slog"
	"os"
	"strconv"
	"time"
)

type Config struct {
	Server ServerConfig
	DB     DBConfig
	Enc    EncryptionConfig
	Redis  RedisConfig
}

type ServerConfig struct {
	Host string
	Port int
}
type EncryptionConfig struct {
	SystemKey string
}

type DBConfig struct {
	PrimaryDSN      string
	ReplicaDSNs     []string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime time.Duration
}

type RedisConfig struct {
	Enabled  bool
	Addr     string
	Password string
	DB       int
	TTL      time.Duration
	Logger   *slog.Logger
}

func Load() (*Config, error) {
	serverCfg, err := loadServerConfig()
	if err != nil {
		return nil, fmt.Errorf("server config: %w", err)
	}

	dbCfg, err := loadDBConfig()
	if err != nil {
		return nil, fmt.Errorf("database config: %w", err)
	}
	encCfg, err := LoadEncryptionConfig()
	if err != nil {
		return nil, fmt.Errorf("encryption config: %w", err)
	}
	redisCfg, err := loadRedisConfig()
	if err != nil {
		return nil, fmt.Errorf("redis config: %w", err)
	}

	return &Config{
		Server: serverCfg,
		DB:     dbCfg,
		Enc:    encCfg,
		Redis:  redisCfg,
	}, nil
}

func loadServerConfig() (ServerConfig, error) {
	host := os.Getenv("SERVER_HOST")
	if host == "" {
		host = "0.0.0.0"
	}

	portStr := os.Getenv("SERVER_PORT")
	if portStr == "" {
		portStr = "8080"
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return ServerConfig{}, fmt.Errorf("invalid port: %w", err)
	}

	return ServerConfig{
		Host: host,
		Port: port,
	}, nil
}

func LoadEncryptionConfig() (EncryptionConfig, error) {
	key := os.Getenv("ENCRYPTION_KEY")
	if key == "" {
		return EncryptionConfig{}, fmt.Errorf("ENCRYPTION_KEY is required")
	}
	return EncryptionConfig{
		SystemKey: key,
	}, nil
}

func loadDBConfig() (DBConfig, error) {
	primaryDSN := os.Getenv("DB_PRIMARY_DSN")
	if primaryDSN == "" {
		return DBConfig{}, fmt.Errorf("DB_PRIMARY_DSN is required")
	}

	config := DBConfig{
		PrimaryDSN:      primaryDSN,
		ReplicaDSNs:     []string{},
		MaxOpenConns:    25,
		MaxIdleConns:    10,
		ConnMaxLifetime: 5 * time.Minute,
	}

	for i := 0; ; i++ {
		key := fmt.Sprintf("DB_REPLICA_DSN_%d", i)
		if dsn := os.Getenv(key); dsn != "" {
			config.ReplicaDSNs = append(config.ReplicaDSNs, dsn)
		} else if i > 0 {
			break
		} else if i == 0 && os.Getenv("DB_REPLICA_DSN") != "" {
			config.ReplicaDSNs = append(config.ReplicaDSNs, os.Getenv("DB_REPLICA_DSN"))
			break
		} else {
			break
		}
	}

	if val, err := strconv.Atoi(os.Getenv("DB_MAX_OPEN_CONNS")); err == nil {
		config.MaxOpenConns = val
	}
	if val, err := strconv.Atoi(os.Getenv("DB_MAX_IDLE_CONNS")); err == nil {
		config.MaxIdleConns = val
	}
	if val := os.Getenv("DB_CONN_MAX_LIFETIME"); val != "" {
		if duration, err := time.ParseDuration(val); err == nil {
			config.ConnMaxLifetime = duration
		}
	}

	return config, nil
}

func loadRedisConfig() (RedisConfig, error) {
	config := RedisConfig{
		Enabled:  true,
		Addr:     "localhost:6379",
		Password: "",
		DB:       0,
		TTL:      2 * time.Hour,
	}

	if addr := os.Getenv("REDIS_ADDR"); addr != "" {
		config.Addr = addr
	}

	if disabled := os.Getenv("REDIS_DISABLED"); disabled == "true" || disabled == "1" {
		config.Enabled = false
	}

	if password := os.Getenv("REDIS_PASSWORD"); password != "" {
		config.Password = password
	}

	if dbStr := os.Getenv("REDIS_DB"); dbStr != "" {
		if db, err := strconv.Atoi(dbStr); err == nil {
			config.DB = db
		}
	}

	if ttlStr := os.Getenv("REDIS_TTL"); ttlStr != "" {
		if ttl, err := time.ParseDuration(ttlStr); err == nil {
			config.TTL = ttl
		}
	}

	return config, nil
}
