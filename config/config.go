package config

import (
	"github.com/caarlos0/env/v6"
)

var Role struct {
	Seller string `env:"Seller" envDefault:"seller"`
	Buyer  string `env:"Buyer" envDefault:"buyer"`
}

func init() {
	_ = env.Parse(&App)
	_ = env.Parse(&Role)
}
