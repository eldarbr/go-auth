package main

import (
	"context"
	"log"
	"net/http"
	"time"

	"github.com/eldarbr/go-auth/internal/provider/database"
	"github.com/eldarbr/go-auth/internal/service/encrypt"
	"github.com/eldarbr/go-auth/internal/service/handler"
	"github.com/eldarbr/go-auth/internal/service/server"
	"github.com/eldarbr/go-auth/pkg/cache"
	"github.com/eldarbr/go-auth/pkg/config"
)

type programConf struct {
	DBUri             string        `yaml:"dbUri"`
	DBMigrationsPath  string        `yaml:"dbMigrations"`
	ServingURI        string        `yaml:"servingUri"`
	PrivatePemPath    string        `yaml:"privatePemPath"`
	PublicPemPath     string        `yaml:"publicPemPath"`
	SslCertfilePath   string        `yaml:"sslCertfilePath"`
	SslKeyfilePath    string        `yaml:"sslKeyfilePath"`
	PprofServingURI   string        `yaml:"pprofServingUri"`
	EnableTLSServing  bool          `yaml:"enableTlsServing"`
	AuthTokenTTL      time.Duration `yaml:"authTokenTtl"`
	RateLimitRequests int           `yaml:"rateLimitRequests"`
	RateLimitTTL      int64         `yaml:"rateLimitTtl"`
	RateLimitCapacity int           `yaml:"rateLimitCapacity"`
}

const CacheAutoEvictPeriodSeconds = 120

func (conf *programConf) setDefaults() {
	if conf == nil {
		return
	}

	conf.RateLimitRequests = 2
	conf.RateLimitTTL = 10
	conf.RateLimitCapacity = 100
}

func main() {
	var (
		programContext = context.Background()
		conf           programConf
	)

	conf.setDefaults()

	err := config.ParseConfig("secret/config.yaml", &conf)
	if err != nil {
		log.Println(err)

		return
	}

	if conf.PprofServingURI != "" {
		log.Println("Starting pprof http")

		go func() {
			log.Println(http.ListenAndServe(conf.PprofServingURI, server.NewPprofServemux()))
		}()
	}

	dbInstance, err := database.Setup(programContext, conf.DBUri, conf.DBMigrationsPath)
	if err != nil {
		log.Println(err)

		return
	}

	defer dbInstance.ClosePool()

	log.Println("Database setup ok")

	cache := cache.NewCache(conf.RateLimitTTL, conf.RateLimitCapacity)

	go cache.AutoEvict(CacheAutoEvictPeriodSeconds * time.Second)
	defer cache.StopAutoEvict()

	var serv *http.Server
	{
		jwtService, jwtErr := encrypt.NewJWTService(conf.PrivatePemPath, conf.PublicPemPath, conf.AuthTokenTTL)
		if jwtErr != nil {
			log.Println(jwtErr)

			return
		}

		authHandl := handler.NewAuthHandl(dbInstance, jwtService, cache, conf.RateLimitRequests)
		manageHandl := handler.NewManageHandl(dbInstance, jwtService, cache, conf.RateLimitRequests)
		router := server.NewRouter(handler.CommonHandl{}, authHandl, manageHandl,
			handler.NewIPRateLimitHandl(conf.RateLimitRequests, cache))
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
