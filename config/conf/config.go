package conf

type Config struct {
	DBHost string `json:"DB_HOST"`

	DBUser string `json:"DB_USER"`

	DBPassword string `json:"DB_PASSWORD"`

	DBPort string `json:"DB_PORT"`

	DBDatabase string `json:"DB_DATABASE"`

	GRPCPort string `json:"GRPC_PORT"`

	HTTPPort string `json:"HTTP_PORT"`

	RedisPassword string `json:"REDIS_PASSWORD"`

	RedisHost string `json:"REDIS_HOST"`

	ZipkinEndpoint string `json:"ZIPKIN_ENDPOINT"`

	ApplicationEnv string `json:"APP_ENV"`

	ApplicationName string `json:"APPLICATION_NAME"`

	KafkaBrokerAddress string `json:"KAFKA_BROKER_ADDRESS"`

	ClientAddress map[string]string `json:"CLIENT_ADDRESS"`

	LogLevel string `json:"LOG_LEVEL"`
}
