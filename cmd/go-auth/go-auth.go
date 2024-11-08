package main

import (
	"context"
	"log"
	"net/http"

	"github.com/eldarbr/go-auth/internal/config"
	"github.com/eldarbr/go-auth/internal/provider/database"
	"github.com/eldarbr/go-auth/internal/service/encrypt"
	"github.com/eldarbr/go-auth/internal/service/handler"
	"github.com/eldarbr/go-auth/internal/service/server"
)

var conf struct {
	DBUri            string `yaml:"dbUri"`
	DBMigrationsPath string `yaml:"dbMigrations"`
	ServingURI       string `yaml:"servingUri"`
	PrivatePemPath   string `yaml:"privatePemPath"`
	PublicPemPath    string `yaml:"publicPemPath"`
	SslCertfilePath  string `yaml:"sslCertfilePath"`
	SslKeyfilePath   string `yaml:"sslKeyfilePath"`
	EnableTLSServing bool   `yaml:"enableTlsServing"`
}

func main() {
	programContext := context.Background()

	err := config.ParseConfig("secret/config.yaml", &conf)
	if err != nil {
		log.Println(err)

		return
	}

	dbInstance, err := database.Setup(programContext, conf.DBUri, conf.DBMigrationsPath)
	if err != nil {
		log.Println(err)

		return
	}

	defer dbInstance.ClosePool()

	var serv *http.Server
	{
		jwtService, jwtErr := encrypt.NewJWTService(conf.PrivatePemPath, conf.PublicPemPath)
		if jwtErr != nil {
			log.Println(jwtErr)

			return
		}

		authHandl := handler.NewAuthHandl(dbInstance, jwtService)
		manageHandl := handler.NewManageHandl(dbInstance, jwtService)
		router := server.NewRouter(handler.CommonHandl{}, authHandl, manageHandl)
		serv = server.NewServer(conf.ServingURI, router)
	}

	if conf.EnableTLSServing {
		err = serv.ListenAndServeTLS(conf.SslCertfilePath, conf.SslKeyfilePath)
	} else {
		err = serv.ListenAndServe()
	}

	if err != nil {
		log.Println(err)
	}
}
