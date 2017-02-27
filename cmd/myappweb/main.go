package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path"
	"runtime"

	"github.com/betacraft/yaag/middleware"
	"github.com/betacraft/yaag/yaag"
	"github.com/boltdb/bolt"
	"github.com/gorilla/context"
	"github.com/justinas/alice"
	"github.com/kardianos/osext"
	"github.com/spf13/viper"

	"myapp"
)

const isDevelopment = "true"

// App in main app
type App struct {
	router *Router
	gp     globalPresenter
	logr   appLogger
}

// globalPresenter contains the fields neccessary for presenting in all templates
type globalPresenter struct {
	SiteName    string
	Description string
	SiteURL     string
}

// SetupApp setup all condition for start project
// TODO localPresenter if we have using template
func SetupApp(r *Router, logger appLogger, templateDirectoryPath string) *App {
	gp := globalPresenter{
		SiteName:    "MyApp",
		Description: "Template web api application",
		SiteURL:     "",
	}
	return &App{
		router: r,
		gp:     gp,
		logr:   logger,
	}
}

func main() {
	pwd, err := osext.ExecutableFolder()
	if err != nil {
		log.Fatalf("cannot retrieve present working directory: %s", err)
	}

	boltdb, err := bolt.Open(path.Join(pwd, "myapp.db"), 0600, nil)
	if err != nil {
		log.Fatalf("unable to open bolt db: %s", err)
	}
	db := &myapp.DB{boltdb}
	err = db.CreateAllBuckets()
	if err != nil {
		log.Fatalf("unable to CreateAllBucketsreate all bucket: %s", err)
	}

	err = LoadConfiguration(pwd)
	if err != nil && viper.GetBool("isProduction") {
		panic(fmt.Errorf("fatal error config file: %s ", err))
	}
	//TODO config static file path and template file path

	r := NewRouter()
	logr := newLogger()
	a := SetupApp(r, logr, "")
	yaag.Init(&yaag.Config{On: true, DocTitle: "Core", DocPath: "apidoc.html", BaseUrls: map[string]string{"Production": "", "Staging": ""}})

	common := alice.New(context.ClearHandler, a.loggingHandler, a.recoverHandler, middleware.Handle)
	//auth := common.Append(a.authMiddleware(db))

	r.Post("/login", common.Then(a.Wrap(a.LoginHandler(db))))

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
	viper.SetConfigName("myapp-config")
	viper.AddConfigPath(pwd)
	devPath := pwd[:len(pwd)-3] + "/src/myapp/cmd/myappweb/"
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
