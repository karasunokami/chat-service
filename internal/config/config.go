package config

type Config struct {
	Global  GlobalConfig  `toml:"global"`
	Log     LogConfig     `toml:"log"`
	Servers ServersConfig `toml:"servers"`
}

type GlobalConfig struct {
	Env string `toml:"env" validation:"required,oneof=dev stage prod"`
}

type LogConfig struct {
	Level string `toml:"level" validation:"required,oneof=debug info warn error"`
}

type ServersConfig struct {
	Debug DebugServerConfig `toml:"debug"`
}

type DebugServerConfig struct {
	Addr string `toml:"addr" validate:"required,hostname_port"`
}
