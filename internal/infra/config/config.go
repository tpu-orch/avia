package config

import (
	"fmt"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	HTTP     HTTPConfig     `yaml:"http"`
	Postgres PostgresConfig `yaml:"postgres"`
	// Kafka    KafkaConfig    `yaml:"kafka"`
}

type HTTPConfig struct {
	Address      string        `yaml:"address"`
	ReadTimeout  time.Duration `yaml:"read_timeout"`
	WriteTimeout time.Duration `yaml:"write_timeout"`
}

type PostgresConfig struct {
	DSN             string        `yaml:"dsn"`
	MaxConns        int32         `yaml:"max_conns"`
	MinConns        int32         `yaml:"min_conns"`
	MaxConnLifetime time.Duration `yaml:"max_conn_lifetime"`
	ConnectTimeout  time.Duration `yaml:"connect_timeout"`
}

// type KafkaConfig struct {
// 	Brokers []string `yaml:"brokers"`
// 	GroupID string   `yaml:"group_id"`
// 	Topic   string   `yaml:"topic"`
// }

func Load(path string) (*Config, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return nil, fmt.Errorf("read config %s: %w", path, err)
	}

	var cfg Config
	if err := yaml.Unmarshal(content, &cfg); err != nil {
		return nil, fmt.Errorf("decode config %s: %w", path, err)
	}

	// Gin HTTP-сервер
	if cfg.HTTP.Address == "" {
		cfg.HTTP.Address = ":8080"
	}
	if cfg.HTTP.ReadTimeout <= 0 {
		cfg.HTTP.ReadTimeout = 5 * time.Second
	}
	if cfg.HTTP.WriteTimeout <= 0 {
		cfg.HTTP.WriteTimeout = 10 * time.Second
	}

	// PostgreSQL
	if cfg.Postgres.MaxConns <= 0 {
		cfg.Postgres.MaxConns = 10
	}
	if cfg.Postgres.MinConns < 0 {
		cfg.Postgres.MinConns = 0
	}
	if cfg.Postgres.MinConns > cfg.Postgres.MaxConns {
		return nil, fmt.Errorf("postgres min_conns cannot exceed max_conns")
	}
	if cfg.Postgres.MaxConnLifetime <= 0 {
		cfg.Postgres.MaxConnLifetime = time.Hour
	}
	if cfg.Postgres.ConnectTimeout <= 0 {
		cfg.Postgres.ConnectTimeout = 5 * time.Second
	}

	// Kafka
	// if cfg.Kafka.GroupID == "" {
	// 	cfg.Kafka.GroupID = "analyzer"
	// }
	// if len(cfg.Kafka.Brokers) == 0 {
	// 	return nil, fmt.Errorf("kafka brokers must be configured")
	// }

	return &cfg, nil
}
