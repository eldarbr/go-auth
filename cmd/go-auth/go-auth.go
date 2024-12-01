package main

import (
	"context"
	"log"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/eldarbr/go-auth/internal/service/encrypt"
	"github.com/eldarbr/go-auth/internal/service/handler"
	"github.com/eldarbr/go-auth/internal/service/server"
	"github.com/eldarbr/go-auth/pkg/cache"
	"github.com/eldarbr/go-auth/pkg/config"
	"github.com/eldarbr/go-auth/pkg/database"
)

type programConf struct {
	DBUri               string        `yaml:"dbUri"`
	ServingURI          string        `yaml:"servingUri"`
	PrivatePemPath      string        `yaml:"privatePemPath"`
	PublicPemPath       string        `yaml:"publicPemPath"`
	SslCertfilePath     string        `yaml:"sslCertfilePath"`
	SslKeyfilePath      string        `yaml:"sslKeyfilePath"`
	PprofServingURI     string        `yaml:"pprofServingUri"`
	EnableTLSServing    bool          `yaml:"enableTlsServing"`
	AuthTokenTTL        time.Duration `yaml:"authTokenTtl"`
	RateLimitRequests   int           `yaml:"rateLimitRequests"`
	RateLimitTTL        int64         `yaml:"rateLimitTtl"`
	RateLimitCapacity   int           `yaml:"rateLimitCapacity"`
	CookieSessionDomain string        `yaml:"cookieSessionDomain"`
}

const (
	CacheAutoEvictPeriodSeconds = 120
	DBMigrationsPath            = "file://./sql" // expect the migrations to be next to the app.
)

func (conf *programConf) setDefaults() {
	if conf == nil {
		return
	}

	conf.RateLimitRequests = 2
	conf.RateLimitTTL = 10
	conf.RateLimitCapacity = 100
}

func main() {
	var conf programConf

	conf.setDefaults()

	programContext, programContextStop := signal.NotifyContext(context.Background(), syscall.SIGINT)

	defer programContextStop()

	err := config.ParseConfig("secret/config.yaml", &conf)
	if err != nil {
		log.Println(err)

		return
	}

	if conf.PprofServingURI != "" {
		log.Println("Starting pprof http")

		go func() {
			//nolint:gosec // not an exposed to prod server, it's ok.
			log.Println(http.ListenAndServe(conf.PprofServingURI, server.NewPprofServemux()))
		}()
	}

	jwtService, jwtErr := encrypt.NewJWTService(conf.PrivatePemPath, conf.PublicPemPath, conf.AuthTokenTTL)
	if jwtErr != nil {
		log.Println(jwtErr)

		return
	}

	dbInstance, err := database.Setup(programContext, conf.DBUri, DBMigrationsPath)
	if err != nil {
		log.Println(err)

		return
	}

	log.Println("Database setup ok")

	cache := cache.NewCache(conf.RateLimitTTL, conf.RateLimitCapacity)

	go cache.AutoEvict(CacheAutoEvictPeriodSeconds * time.Second)

	var serv *http.Server
	{
		authHandl := handler.NewAuthHandl(dbInstance, jwtService, cache, conf.RateLimitRequests, conf.CookieSessionDomain)
		manageHandl := handler.NewManageHandl(dbInstance, jwtService, cache, conf.RateLimitRequests)
		router := server.NewRouter(handler.CommonHandl{}, authHandl, manageHandl,
			handler.NewIPRateLimitHandl(conf.RateLimitRequests, cache))
		serv = server.NewServer(conf.ServingURI, router)
	}

	if conf.EnableTLSServing {
		go func() {
			err = serv.ListenAndServeTLS(conf.SslCertfilePath, conf.SslKeyfilePath)

			programContextStop()
		}()
	} else {
		go func() {
			err = serv.ListenAndServe()

			programContextStop()
		}()
	}

	<-programContext.Done()

	if err != nil {
		log.Println(err)
	}

	log.Println("shutting down")

	{
		shutdownContext, shutdownContextCancel := context.WithTimeout(programContext, time.Second*15)
		defer shutdownContextCancel()

		err = serv.Shutdown(shutdownContext)

		cache.StopAutoEvict()
		dbInstance.ClosePool()
	}

	if err != nil {
		log.Println(err)
	}
}
