package config

// TODO: 임시
type Config struct {
	Symbols  []string
	DBDriver string
	DBSource string
}

// TODO: 임시
func LoadConfig() (*Config, error) {
	return &Config{
		Symbols:  []string{"BTCUSDT", "ETHUSDT"},
		DBDriver: "mysql",
		DBSource: "root:123456@tcp(localhost:3306)/binance?allowAllFiles=true",
	}, nil
}
