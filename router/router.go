package router

import (
	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-gonic/gin"
	ginprometheus "github.com/mcuadros/go-gin-prometheus"
	"github.pkgms.com/techops/peak-self-serve/controller"
	api "gopkg.in/appleboy/gin-status-api.v1"
)

// Route makes the routing
func Route(app *gin.Engine) {
	indexController := new(controller.IndexController)
	testController := new(controller.TestController)
	authController := new(controller.AuthController)
	p := ginprometheus.NewPrometheus("gin")
	p.Use(app)
	store := cookie.NewStore([]byte("secret"))
	app.Use(sessions.Sessions("mysession", store))
	app.GET(
		"/", authController.HomeHandler,
	)
	// Test
	app.GET(
		"/endpoint1", testController.Endpoint1,
	)
	app.GET(
		"/endpoint2", testController.Endpoint2,
	)
	// Auth
	app.GET("/login", authController.LoginHandler)
	app.GET("/authorization-code/callback", authController.AuthCallbackHandler)
	app.GET("/profile", authController.ProfileHandler)
	app.POST("/logout", authController.Logout)
	app.GET("/logout", authController.Logout)
	app.POST("/submitComment", testController.SubmitCommentHandler)

	apiGroup := app.Group("/api")
	{
		apiGroup.GET("/version", indexController.GetVersion)
		apiGroup.GET("/status", api.StatusHandler)
	}

}
