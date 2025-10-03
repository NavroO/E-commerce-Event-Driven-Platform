package config

import "os"

type Config struct {
	HTTPAddr       string
	MongoURI       string
	MongoDB        string
	RabbitURI      string
	EventsExchange string
	ServiceName    string
	LogLevel       string
}

func FromEnv() Config {
	cfg := Config{
		HTTPAddr:       getEnv("APP_HTTP_ADDR", ":8080"),
		MongoURI:       getEnv("MONGO_URI", "mongodb://localhost:27017"),
		MongoDB:        getEnv("MONGO_DB", "ecommerce"),
		RabbitURI:      getEnv("RABBIT_URI", "amqp://guest:guest@localhost:5672/"),
		EventsExchange: getEnv("EVENTS_EXCHANGE", "ecommerce.events"),
		ServiceName:    getEnv("SERVICE_NAME", "order-service"),
		LogLevel:       getEnv("LOG_LEVEL", "info"),
	}
	return cfg
}

func getEnv(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}
