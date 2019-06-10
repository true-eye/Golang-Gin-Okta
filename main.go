package main

import (
	"flag"
	"path/filepath"

	"github.com/gin-contrib/location"
	"github.com/gin-gonic/gin"
	_ "github.com/golang/glog"
	"github.com/k0kubun/pp"
	"github.pkgms.com/techops/peak-self-serve/config"
	"github.pkgms.com/techops/peak-self-serve/middleware"
	"github.pkgms.com/techops/peak-self-serve/router"
)

func main() {

	addr := flag.String("addr", config.Server.Addr, "Address to listen and serve")
	flag.Parse()
	app := gin.New()
	pp.Println(config.Okta)

	app.Use(gin.Logger())
	app.Use(gin.Recovery())
	app.Use(location.Default())
	app.Use(middleware.AuthMiddleware)

	app.Static("/images", filepath.Join(config.Server.StaticDir, "img"))
	app.StaticFile("/favicon.ico", filepath.Join(config.Server.StaticDir, "img/favicon.ico"))
	app.LoadHTMLGlob(config.Server.ViewDir + "/*")
	app.MaxMultipartMemory = config.Server.MaxMultipartMemory << 20

	router.Route(app)
	// Listen and Serve
	app.Run(*addr)
}
