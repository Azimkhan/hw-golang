package conf

import (
	"github.com/BurntSushi/toml"
)

// При желании конфигурацию можно вынести в internal/config.
// Организация конфига в main принуждает нас сужать API компонентов, использовать
// при их конструировании только необходимые параметры, а также уменьшает вероятность циклической зависимости.
type Config struct {
	Logger  LoggerConf
	HTTP    HTTPConf
	GRPC    GRPCConf
	Storage StorageConf
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

func NewConfig() Config {
	return Config{}
}

func (c *Config) LoadFromFile(path string) error {
	_, err := toml.DecodeFile(path, c)
	return err
}
