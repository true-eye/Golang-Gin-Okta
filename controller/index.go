package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.pkgms.com/techops/peak-self-serve/config"
)

// IndexController is the default controller
type IndexController struct{}

// GetIndex home page
func (ctrl *IndexController) GetIndex(c *gin.Context) {
	c.HTML(http.StatusOK, "home.gohtml", gin.H{
		"title":   "Gin Skeleton",
		"content": "This is a skeleton based on gin framework",
	})
}

// GetVersion version json
func (ctrl *IndexController) GetVersion(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"version": config.Server.Version,
	})
}
