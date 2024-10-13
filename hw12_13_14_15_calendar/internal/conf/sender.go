package conf

import "github.com/BurntSushi/toml"

type SenderConfig struct {
	Logger  LoggerConf
	Storage StorageConf
	AMQP    AMQPConfig
}

func NewSenderConfig() *SenderConfig {
	return &SenderConfig{}
}

func (c *SenderConfig) LoadFromFile(path string) error {
	_, err := toml.DecodeFile(path, c)
	return err
}
