//go:build darwin
// +build darwin

package config

var App struct {
	Name   string `env:"Name" envDefault:"Pali"`
	Mode   string `env:"Mode" envDefault:"live"`
	Port   string `env:"Port" envDefault:"3000"`
	ENV    string `env:"ENV" envDefault:"local"`
	JWTKey string `env:"JWTKey" envDefault:"kiuru72h2ywn"`
	Url    string `env:"Url" envDefault:"http://127.0.0.1:3000/"`
}

var Database struct {
	Connection string `env:"Connection" envDefault:"mysql"`
	Port       string `env:"Port" envDefault:"3306"`
	Host       string `env:"Host" envDefault:"127.0.0.1"`
	Database   string `env:"Database" envDefault:"mvpmatch"`
	Username   string `env:"Username" envDefault:"root"`
	Password   string `env:"Password" envDefault:"root"`
}
