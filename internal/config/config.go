package config

type Config struct {
	Global  GlobalConfig  `toml:"global"`
	Log     LogConfig     `toml:"log"`
	Servers ServersConfig `toml:"servers"`
	Sentry  SentryConfig  `toml:"sentry"`
	Clients ClientsConfig `toml:"clients"`
}

const GlobalEnvProd = "prod"

type GlobalConfig struct {
	Env string `toml:"env" validation:"required,oneof=dev stage prod"`
}

func (c *GlobalConfig) IsInProdEnv() bool {
	return c.Env == GlobalEnvProd
}

type LogConfig struct {
	Level string `toml:"level" validation:"required,oneof=debug info warn error"`
}

type ServersConfig struct {
	Debug  DebugServerConfig  `toml:"debug"`
	Client ClientServerConfig `toml:"client"`
}

type DebugServerConfig struct {
	Addr string `toml:"addr" validate:"required,hostname_port"`
}

type ClientServerConfig struct {
	Addr           string               `toml:"addr" validate:"required,hostname_port"`
	AllowOrigins   []string             `toml:"allow_origins" validate:"required"`
	RequiredAccess RequiredAccessConfig `toml:"required_access" validate:"required"`
}

type RequiredAccessConfig struct {
	Resource string `toml:"resource" validate:"required"`
	Role     string `toml:"role" validate:"required"`
}

type SentryConfig struct {
	Dsn string `toml:"dsn" validate:"required,unix_addr"`
}

type ClientsConfig struct {
	KeycloakClient KeycloakClientConfig `toml:"keycloak" validate:"required"`
}

type KeycloakClientConfig struct {
	BasePath     string `toml:"base_path" validate:"required,uri"`
	Realm        string `toml:"realm" validate:"required"`
	ClientID     string `toml:"client_id" validate:"required"`
	ClientSecret string `toml:"client_secret" validate:"required"`
	DebugMode    bool   `toml:"debug_mode"`
}
