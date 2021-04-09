// config.go
package config

import (
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"

	"github.com/joho/godotenv"
)

type PingerCnf struct {
	WorkerCount uint32
	DelayGlobal uint32
	Separator   string
}

type Config struct {
	PingerCnf
}

// String делает правильный вывод стркутуры (прячет пароль)
func (c *Config) String() string {
	var out string
	v := reflect.ValueOf(*c)
	typeOfS := v.Type()
	for i := 0; i < v.NumField(); i++ {

		switch typeOfS.Field(i).Name {
		case "LdapConfig":
			// заменить 4-ое поле пароля на ***
			var vo string
			//lds := make([]string, 4)
			tt := fmt.Sprintf("%v", v.Field(i).Interface())
			lds := make([]string, len(tt))
			lds = strings.Split(tt, " ")
			lds[3] = "***"
			vo = strings.Join(lds, " ")
			out += fmt.Sprintf("Field: %s\tValue: %v\n", typeOfS.Field(i).Name, vo)
		default:
			out += fmt.Sprintf("Field: %s\tValue: %v\n", typeOfS.Field(i).Name, v.Field(i).Interface())
		}
	}

	return fmt.Sprint(out)
}

// ConfigLoad загружает .env
func ConfigLoad() (*Config, error) {

	if err := godotenv.Load(); err != nil {
		return nil, fmt.Errorf("%s: %s", "load .env file", err.Error())
	}
	if err := godotenv.Load(); err != nil {
		return nil, err
	}
	return &Config{
		PingerCnf{
			WorkerCount: uint32(getEnvAsInt("WORKER_COUNT", 10)),
			DelayGlobal: uint32(getEnvAsInt("DELAY_GLOBAL", 0)),
			Separator:   getEnv("SEPARATOR", "++"),
		},
	}, nil
}

// getEnv возвращает значение если существует, либо устанавливает по умолчанию
func getEnv(env string, defaultVal string) string {
	if val, ok := os.LookupEnv(env); ok {
		return val
	}
	return defaultVal
}

func getEnvAsBool(env string, defaultVal bool) bool {
	valStr := getEnv(env, "")
	if val, err := strconv.ParseBool(valStr); err == nil {
		return val
	}
	return defaultVal
}

func getEnvAsInt(env string, defaultVal int) int {
	valStr := getEnv(env, "")
	if val, err := strconv.Atoi(valStr); err == nil {
		return val
	}
	return defaultVal
}

func getEnvAsSlice(env string, defaultVal []string) []string {
	valStr := getEnv(env, "")
	if valStr != "" {
		return strings.Split(valStr, " ")
	}
	return defaultVal
}
