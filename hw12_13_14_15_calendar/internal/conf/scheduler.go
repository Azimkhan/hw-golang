package conf

import "github.com/BurntSushi/toml"

type SchedulerConfig struct {
	ScanInterval int
	Logger       LoggerConf
	Storage      StorageConf
	AMQP         AMQPConfig
}

func NewSchedulerConfig() SchedulerConfig {
	return SchedulerConfig{}
}

func (c *SchedulerConfig) LoadFromFile(path string) error {
	_, err := toml.DecodeFile(path, c)
	return err
}
