package cfg

import (
	"errors"
	"github.com/spf13/viper"
)

type Env struct {
	DBUser      string `mapstructure:"POSTGRES_USER"`
	DBPass      string `mapstructure:"POSTGRES_PASSWORD"`
	DBName      string `mapstructure:"POSTGRES_DB"`
	DBHost      string `mapstructure:"DB_HOST"`
	DBPort      string `mapstructure:"DB_PORT"`
	URLDomain   string `mapstructure:"URL_DOMAIN"`
	ServerPort  string `mapstructure:"SERVER_PORT"`
	ShortLen    string `mapstructure:"SHORT_LEN"`
	FillingChar string `mapstructure:"FILLING_CHAR"`
	CheckURLs   string `mapstructure:"CHECK_URLS"`
}

func ReadCFG() (*Env, error) {
	env := Env{}
	viper.SetConfigFile(".env")
	err := viper.ReadInConfig()
	if err != nil {
		return nil, errors.New("No .env file provided in root")
	}
	err = viper.Unmarshal(&env)
	if err != nil {
		return nil, errors.New("Error while .env file read")
	}
	return &env, nil
}
