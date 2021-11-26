//go:build darwin
// +build darwin

package config

var App struct {
	Name   string `env:"Name" envDefault:"Mvp"`
	Mode   string `env:"Mode" envDefault:"live"`
	Port   string `env:"Port" envDefault:"3000"`
	ENV    string `env:"ENV" envDefault:"local"`
	JWTKey string `env:"JWTKey" envDefault:"kiuru72h2ywn"`
	Url    string `env:"Url" envDefault:"http://127.0.0.1:3000/"`
}