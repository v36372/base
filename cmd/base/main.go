package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"

	"github.com/boltdb/bolt"
	"github.com/gorilla/context"
	"github.com/justinas/alice"
	"github.com/kardianos/osext"
	"github.com/spf13/viper"

	"base"
)

type baseConfig struct {
	Port string
}

// App in main app
type App struct {
	router *Router
	logr   appLogger
	config baseConfig
}

// SetupApp setup all condition for start project
func SetupApp(r *Router, logger appLogger) *App {
	var config baseConfig
	if viper.GetBool("isDevelopment") {
		config = baseConfig{
			Port: viper.GetString("port"),
		}
	} else {
		config = baseConfig{
			Port: os.Getenv("PORT"),
		}
	}

	return &App{
		router: r,
		logr:   logger,
		config: config,
	}
}

func main() {
	pwd, err := osext.ExecutableFolder()
	if err != nil {
		log.Fatalf("cannot retrieve present working directory: %s", err)
	}

	boltdb, err := bolt.Open(path.Join(pwd, "base.db"), 0600, nil)
	if err != nil {
		log.Fatalf("unable to open bolt db: %s", err)
	}
	db := &base.DB{boltdb}
	err = db.CreateAllBuckets()
	if err != nil {
		log.Fatalf("unable to CreateAllBucketsreate all bucket: %s", err)
	}

	err = LoadConfiguration(pwd)
	if err != nil && viper.GetBool("isProduction") {
		panic(fmt.Errorf("fatal error config file: %s ", err))
	}

	r := NewRouter()
	logr := newLogger()
	a := SetupApp(r, logr)

	common := alice.New(context.ClearHandler, a.loggingHandler, a.recoverHandler)

	r.Post("/", common.Then(a.Wrap(a.IndexHandler(db))))

	err = http.ListenAndServe(":3000", r)
	if err != nil {
		log.Fatalf("error on serve server %s", err)
	}
	defer func() {
		err = db.Close()
		if err != nil {
			log.Fatalf("error on closing db %s", err)
		}
	}()
}

// LoadConfiguration load file config in directory
func LoadConfiguration(pwd string) error {
	viper.SetConfigName("base-config")
	viper.AddConfigPath(pwd)
	devPath := pwd[:len(pwd)-3] + "/src/base/cmd/base/"
	_, file, _, _ := runtime.Caller(1)
	configPath := path.Dir(file)
	viper.AddConfigPath(devPath)
	viper.AddConfigPath(configPath)

	// setup for config path of product deployment
	if os.Getenv("isDevelopment") == "false" {
		productionConfigPath := ""
		viper.AddConfigPath(productionConfigPath)
	}
	viper.SetDefault("path", devPath)
	return viper.ReadInConfig() // Find and read the config file
}
