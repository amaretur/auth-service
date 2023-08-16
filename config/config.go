package config

import (
	"time"
	"strings"
	"path/filepath"

	"github.com/spf13/viper"
)

// Настройки http сервера
type Http struct {
	Port			int
	MaxHeaderBytes	int
	ReadTimeout		time.Duration
	WriteTimeout	time.Duration
}

// Конфигурация jwt токена
type Jwt struct {
	AccessExpire	time.Duration
	RefreshExpire	time.Duration
	Secret			string
}

type Config struct {
	Http	Http
	Jwt		Jwt
}

func Init(path string) (*Config, error) {

	dir, filename := filepath.Split(path)
	ext := strings.TrimPrefix(filepath.Ext(filename), ".")
	filenameWithountExt := strings.TrimSuffix(filename, "." + ext)

	viper.SetConfigName(filenameWithountExt)
	viper.SetConfigType(ext)
	viper.AddConfigPath(dir)

	if err := viper.ReadInConfig(); err != nil {
		return nil, err
	}

	c := &Config{
		Http: Http{
			Port: viper.GetInt("server.port"),
			MaxHeaderBytes: viper.GetInt("server.max_header_bytes"),
			ReadTimeout: viper.GetDuration("server.read_timeout"),
			WriteTimeout: viper.GetDuration("server.write_timeout"),
		},

		Jwt: Jwt{
			AccessExpire: viper.GetDuration("jwt.access_expire"),
			RefreshExpire: viper.GetDuration("jwt.refresh_expire"),
			Secret: viper.GetString("jwt.secret"),
		},
	}

	return c, nil
}
