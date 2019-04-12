// Config is put into a different package to prevent cyclic imports in case
// it is needed in several locations

package config

import "time"

type Config struct {
	Period   time.Duration `config:"period"`
	Host     string        `config:"host"`
	Port     string        `config:"port"`
	Username string        `config:"Username"`
	Password string        `config:"Password"`
	Symmid   string        `config:"symmid"`
}

var DefaultConfig = Config{
	Period:   60 * time.Second,
	Port:     "8443",
	Username: "smc",
	Password: "smc",
}
