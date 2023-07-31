package database

import (
	"context"
	"github.com/go-redis/redis/v8"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

const (
	Port     = "3306"
	Host     = "127.0.0.1"
	Database = "mvpmatch"
	Username = "root"
	Password = "koreroot"
)

var DB *gorm.DB

func Start() {
	dbName := Database
	dbUser := Username
	dbPass := Password
	dbHost := Host
	dbPort := Port
	dsn := dbUser + ":" + dbPass + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName
	dsn = dsn + "?charset=utf8&parseTime=True&loc=Local"

	var err error

	DB, err = gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,   // data source name,
		DefaultStringSize:         256,   // add default size for string fields,
		DisableDatetimePrecision:  true,  // disable datetime precision support,
		DontSupportRenameIndex:    true,  // drop & create index when rename index,
		DontSupportRenameColumn:   true,  // use change when rename column,
		SkipInitializeWithVersion: false, // smart configure based on used version
	}), &gorm.Config{
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		panic(err)
	}
}

func ConnectRedis() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // host:port of the redis server
		Password: "",               // no password set
		DB:       0,                // use default DB
	})

	ctx := context.TODO()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return client, err
	}

	return client, nil
}
