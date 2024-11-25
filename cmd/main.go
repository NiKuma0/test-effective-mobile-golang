package main

import (
	"database/sql"
	"log"

	"github.com/gin-gonic/gin"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	"github.com/joho/godotenv"
	swaggerfiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"

	"github.com/nikuma0/test-effective-mobile-golang/config"
	_ "github.com/nikuma0/test-effective-mobile-golang/docs"
	"github.com/nikuma0/test-effective-mobile-golang/internal/http"
	"github.com/nikuma0/test-effective-mobile-golang/internal/repository/postgresql"
	"github.com/nikuma0/test-effective-mobile-golang/internal/utils"
)

//	@title			Swagger Songs API
//	@version		1.0
//	@description	This is the server, why do you read this?
//	@termsOfService	http://swagger.io/terms/

//	@contact.name	API Support
//	@contact.url	http://www.swagger.io/support
//	@contact.email	support@swagger.io

//	@license.name	Apache 2.0
//	@license.url	http://www.apache.org/licenses/LICENSE-2.0.html

//	@host		localhost:8080
//	@BasePath	/api/v1

//	@securityDefinitions.basic	BasicAuth

func main() {
	// Env Variables
	godotenv.Load()
	config, err := config.New()
	if err != nil {
		log.Fatal(err)
	}

	// Init Logging
	utils.InitLog(config)

	// connect DB
	db, err := sql.Open("postgres", config.Db)
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// migrations
	migrationDriver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		log.Fatal(err)
	}
	m, err := migrate.NewWithDatabaseInstance(
		"file://migrations",
		"postgres", migrationDriver)
	if err != nil {
		log.Fatal(err)
	}
	if err = m.Up(); err != nil && err != migrate.ErrNoChange {
		log.Fatal(err)
	}

	// gin
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(utils.LoggerMiddleware())
	handler := http.New(func() postgresql.SongsRepositoryI { return postgresql.NewSongsRepository(db) })
	v1 := r.Group("/api/v1")
	handler.Routes(v1)
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerfiles.Handler))
	r.Run(":8080")
}
