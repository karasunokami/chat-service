package config

import "time"

type Config struct {
	Global   GlobalConfig   `toml:"global"`
	Log      LogConfig      `toml:"log"`
	Servers  ServersConfig  `toml:"servers"`
	Sentry   SentryConfig   `toml:"sentry"`
	Clients  ClientsConfig  `toml:"clients"`
	Services ServicesConfig `toml:"services"`
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
	Debug   DebugServerConfig   `toml:"debug"`
	Client  ClientServerConfig  `toml:"client"`
	Manager ManagerServerConfig `toml:"manager"`
}

type DebugServerConfig struct {
	Addr string `toml:"addr" validate:"required,hostname_port"`
}

type ClientServerConfig struct {
	Addr           string               `toml:"addr" validate:"required,hostname_port"`
	AllowOrigins   []string             `toml:"allow_origins" validate:"required"`
	RequiredAccess RequiredAccessConfig `toml:"required_access" validate:"required"`
	SecWsProtocol  string               `toml:"sec_ws_protocol" validate:"required"`
}

type ManagerServerConfig struct {
	Addr           string               `toml:"addr" validate:"required,hostname_port"`
	AllowOrigins   []string             `toml:"allow_origins" validate:"required"`
	RequiredAccess RequiredAccessConfig `toml:"required_access" validate:"required"`
	SecWsProtocol  string               `toml:"sec_ws_protocol" validate:"required"`
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
	PSQLClient     PSQLClientConfig     `toml:"psql" validate:"required"`
}

type KeycloakClientConfig struct {
	BasePath     string `toml:"base_path" validate:"required,uri"`
	Realm        string `toml:"realm" validate:"required"`
	ClientID     string `toml:"client_id" validate:"required"`
	ClientSecret string `toml:"client_secret" validate:"required"`
	DebugMode    bool   `toml:"debug_mode"`
}

type PSQLClientConfig struct {
	Address   string `toml:"address" validate:"required,hostname_port"`
	Username  string `toml:"username" validate:"required"`
	Password  string `toml:"password" validate:"required"`
	Database  string `toml:"database" validate:"required"`
	DebugMode bool   `toml:"debug_mode"`
}

type ServicesConfig struct {
	MessageProducerService MessageProducerServiceConfig      `toml:"msg_producer" validate:"required"`
	OutboxService          OutboxServiceConfig               `toml:"outbox" validate:"required"`
	ManagerLoad            ManagerLoadServiceConfig          `toml:"manager_load" validate:"required"`
	AfcVerdictsProcessor   AfcVerdictsProcessorServiceConfig `toml:"afc_verdicts_processor" validate:"required"`
	ManagerScheduler       ManagerSchedulerConfig            `toml:"manager_scheduler" validate:"required"`
}

type MessageProducerServiceConfig struct {
	Brokers    []string `toml:"brokers" validate:"dive,hostname_port"`
	Topic      string   `toml:"topic" validate:"required"`
	BatchSize  int      `toml:"batch_size" validate:"required,gte=1,lte=100"`
	EncryptKey string   `toml:"encrypt_key"`
}

type OutboxServiceConfig struct {
	Workers    int           `toml:"workers" validate:"required,gte=1,lte=100"`
	IdleTime   time.Duration `toml:"idle_time" validate:"required"`
	ReserveFor time.Duration `toml:"reserve_for" validate:"required"`
}

type ManagerLoadServiceConfig struct {
	MaxProblemsAtSameTime int `toml:"max_problems_at_same_time" validate:"required,gte=1,lte=100"`
}

type AfcVerdictsProcessorServiceConfig struct {
	Brokers                  []string `toml:"brokers" validate:"required"`
	ConsumersCount           int      `toml:"consumers_count" validate:"required,gte=1,lte=100"`
	ConsumersGroupName       string   `toml:"consumers_group_name" validate:"required"`
	VerdictsTopicName        string   `toml:"verdicts_topic_name" validate:"required"`
	VerdictsDqlTopicName     string   `toml:"verdicts_dql_topic_name" validate:"required"`
	VerdictsSigningPublicKey string   `toml:"verdicts_signing_public_key"`
}

type ManagerSchedulerConfig struct {
	Period time.Duration `toml:"period" validate:"required"`
}
