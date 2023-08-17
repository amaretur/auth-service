package config

import (
	"fmt"
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

// Конфигурация mongodb
type MongoDB struct {
	Protocol	string
	Path		string
	Params		string
	Username	string
	Password	string
	OpenTimeout	time.Duration
	Database	string
}

func (m MongoDB) ConnectURL() string {

	url := fmt.Sprintf(
		"%s://%s:%s@%s/",
		m.Protocol,
		m.Username,
		m.Password,
		m.Path,
	)

	if m.Params != "" {
		url += fmt.Sprintf("?%s", m.Params)
	}

	return url
}

type Config struct {
	Http	Http
	Jwt		Jwt
	MongoDB	MongoDB
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

		MongoDB: MongoDB{
			Protocol: viper.GetString("mongodb.protocol"),
			Path: viper.GetString("mongodb.path"),
			Params: viper.GetString("mongodb.params"),
			Username: viper.GetString("mongodb.username"),
			Password: viper.GetString("mongodb.password"),
			OpenTimeout: viper.GetDuration("mongodb.open_timeout"),
			Database: viper.GetString("mongodb.database"),
		},
	}

	return c, nil
}
