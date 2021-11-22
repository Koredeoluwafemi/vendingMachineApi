package database

import (
	"context"
	"github.com/go-redis/redis/v8"
	"github.com/gofiber/fiber/v2"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"html/template"
	"math"
	"mvpmatch/config"
	"strconv"
)

var DB *gorm.DB

func Start() {
	dbName := config.Database.Database
	dbUser := config.Database.Username
	dbPass := config.Database.Password
	dbHost := config.Database.Host
	dbPort := config.Database.Port
	dsn := dbUser + ":" + dbPass + "@tcp(" + dbHost + ":" + dbPort + ")/" + dbName
	dsn = dsn + "?charset=utf8&parseTime=True&loc=Local"

	var err error

	DB, err = gorm.Open(mysql.New(mysql.Config{
		DSN:                       dsn,   // data source name, refer https://github.com/go-sql-driver/mysql#dsn-data-source-name
		DefaultStringSize:         256,   // add default size for string fields, by default, will use db type `longtext` for fields without size, not a primary key, no index defined and don't have default values
		DisableDatetimePrecision:  true,  // disable datetime precision support, which not supported before MySQL 5.6
		DontSupportRenameIndex:    true,  // drop & create index when rename index, rename index not supported before MySQL 5.7, MariaDB
		DontSupportRenameColumn:   true,  // use change when rename column, rename rename not supported before MySQL 8, MariaDB
		SkipInitializeWithVersion: false, // smart configure based on used version
	}), &gorm.Config{
		//Logger: newLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
	})

	if err != nil {
		panic(err)
	}
}

func ConnectRedis() (*redis.Client, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     "localhost:6379", // host:port of the redis server
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	ctx := context.TODO()
	_, err := client.Ping(ctx).Result()
	if err != nil {
		return client, err
	}

	return client, nil
}

func Paginate(c *fiber.Ctx, pageSize int) func(db *gorm.DB) *gorm.DB {

	return func(db *gorm.DB) *gorm.DB {
		page, _ := strconv.Atoi(c.Query("page"))
		if page == 0 {
			page = 1
		}

		switch {
		case pageSize > 100:
			pageSize = 100
		case pageSize <= 0:
			pageSize = 10
		}

		offset := (page - 1) * pageSize
		return db.Offset(offset).Limit(pageSize)
	}
}

func PageLinks(c *fiber.Ctx, dbCount int, perPage int) template.HTML {

	pageNumber, _ := strconv.Atoi(c.Query("page"))

	adjacents := 2
	path := "?"
	prev := pageNumber - 1
	next := pageNumber + 1
	lastpage := int(math.Ceil(float64(dbCount) / float64(perPage)))
	lpm1 := lastpage - 1

	pagination := ""
	if lastpage > 1 {
		//pagination += "<nav>"
		pagination += "<nav><ul class='pagination'>"
		if pageNumber > 1 {
			pagination += "<li class='page-item'><a  class='page-link' href='" + path + "page=" + strconv.Itoa(prev) + "' aria-label='Previous'><span aria-hidden='true'>&laquo;</span></a></li>"
		} else {
			pagination += "<li class='disabled'><a class='page-link' href='#' aria-label='Previous'><span aria-hidden='true'>&laquo;</span></a></li>"
		}

		if lastpage < 7+(adjacents*2) {
			for counter := 1; counter <= lastpage; counter++ {
				if counter == pageNumber {
					pagination += "<li class='page-item active'><a  class='page-link'>" + strconv.Itoa(counter) + "</a></li>"
				} else {
					pagination += "<li  class='page-item'><a  class='page-link' href='" + path + "page=" + strconv.Itoa(counter) + "'>" + strconv.Itoa(counter) + "</a></li>"
				}
			}
		} else if lastpage > 5+(adjacents*2) {
			if pageNumber < 1+(adjacents*2) {
				for counter := 1; counter < 4+(adjacents*2); counter++ {
					if counter == pageNumber {
						pagination += "<li class='page-item active'><a  class='page-link'>" + strconv.Itoa(counter) + "</a></li>"
					} else {
						pagination += "<li class='page-item'><a  class='page-link' href='" + path + "page=" + strconv.Itoa(counter) + "'>" + strconv.Itoa(counter) + "</a></li>"
					}
				}

				pagination += "<li class='page-item'><span style='border: none; background: none; padding: 8px;'>...</span></li>"
				pagination += "<li class='page-item'><a class='page-link' href='" + path + "page=" + strconv.Itoa(lpm1) + "'>" + strconv.Itoa(lpm1) + "</a></li>"
				pagination += "<li class='page-item'><a class='page-link' href='" + path + "page=" + strconv.Itoa(lastpage) + "'>" + strconv.Itoa(lastpage) + "</a></li>"
			} else if (lastpage-(adjacents*2) > pageNumber) && (pageNumber > (adjacents * 2)) {
				pagination += "<li class='page-item'><a class='page-link' href='" + path + "page=1'>1</a></li>"
				pagination += "<li class='page-item'><a class='page-link' href='" + path + "page=2'>2</a></li>"
				pagination += "<li class='page-item'><span style='border: none; background: none; padding: 8px;'>...</span></li>"

				for counter := pageNumber - adjacents; counter <= pageNumber+adjacents; counter++ {
					if counter == pageNumber {
						pagination += "<li class='page-item'><a class='active page-link'>" + strconv.Itoa(counter) + "</a></li>"
					} else {
						pagination += "<li class='page-item'><a class='page-link' href='" + path + "page=" + strconv.Itoa(counter) + "'>" + strconv.Itoa(counter) + "</a></li>"
					}
				}

				pagination += "<li class='page-item'><span style='border: none; background: none; padding: 8px;'>..</span></li>"
				pagination += "<li class='page-item'><a  class='page-link' href='" + path + "page=" + strconv.Itoa(lpm1) + "'>" + strconv.Itoa(lpm1) + "</a></li>"
				pagination += "<li class='page-item'><a class='page-link' href='" + path + "page=" + strconv.Itoa(lastpage) + "'>" + strconv.Itoa(lastpage) + "</a></li>"

			} else {

				pagination += "<li class='page-item'><a class='page-link' href='" + path + "page=1'>1</a></li>"
				pagination += "<li class='page-item'><a class='page-link' href='" + path + "page=2'>2</a></li>"
				pagination += "<li class='page-item'><span style='border: none; background: none; padding: 8px;'>..</span></li>"

				for counter := lastpage - (2 + (adjacents * 2)); counter <= lastpage; counter++ {
					if counter == pageNumber {
						pagination += "<li class='page-item'><a class='active page-link'>" + strconv.Itoa(counter) + "</a></li>"
					} else {
						pagination += "<li class='page-item'><a class='page-link' href='" + path + "page=" + strconv.Itoa(counter) + "'>" + strconv.Itoa(counter) + "</a></li>"
					}
				}
			}
		}

		//if pageNumber < counter - 1 {
		pagination += "<li class='page-item'><a class='page-link' href='" + path + "page=" + strconv.Itoa(next) + "' aria-label='Next'><span aria-hidden='true'>&raquo;</span></a></li>"
		//} else {
		//	pagination += "<li class='disabled'><a href='#' aria-label='Next'><span aria-hidden='true'>&raquo;</span></a></li>"
		//}
		pagination += "</ul></nav>"
	}

	return template.HTML(pagination)
}
