package conf

import (
	"github.com/BurntSushi/toml"
)

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type APIConfig struct {
	Logger  LoggerConf
	HTTP    HTTPConf
	GRPC    GRPCConf
	Storage StorageConf
}

type AMQPConfig struct {
	URI          string
	Exchange     string
	ExchangeType string
	RoutingKey   string
	Queue        string
}

type StorageConf struct {
	DSN  string
	Type string // sql, inmemory
}

type GRPCConf struct {
	BindAddr string
}

type HTTPConf struct {
	BindAddr string
}

type LoggerConf struct {
	Level string
	// TODO
}

func NewConfig() APIConfig {
	return APIConfig{}
}

func (c *APIConfig) LoadFromFile(path string) error {
	_, err := toml.DecodeFile(path, c)
	return err
}
