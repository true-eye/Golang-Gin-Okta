package controller

import (
	"fmt"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	_ "github.pkgms.com/techops/peak-self-serve/middleware"
)

// IndexController is the default controller

type TestController struct{}

// Submit Comment Handler
func (ctrl *TestController) SubmitCommentHandler(c *gin.Context) {
	value := c.PostForm("value1")
	comment := c.PostForm("comment")
	fmt.Println(value, comment)
	type customData struct {
		IsAuthenticated bool
		Comment           string
		Token        string
		SelectedMenu    string
	}

	data := customData{
		IsAuthenticated: true,
		Comment:           comment,
		Token:        value,
		SelectedMenu:    "group1",
	}
	tpl.ExecuteTemplate(c.Writer, "comment.gohtml", data)
}

// GetIndex home page
func (ctrl *TestController) Endpoint1(c *gin.Context) {
	if IsAuthenticated(c.Request) {

		session := sessions.Default(c)
		var count int
		v := session.Get("count")
		e := session.Get("endpoints")
		var endpoint string
		if v == nil {
			count = 0
			endpoint = "endpoint1"
		} else {
			count = v.(int)
			count++
			endpoint = e.(string) + ".endpoint1"
		}
		session.Set("count", count)
		session.Set("endpoints", endpoint)
		session.Save()
		type customData struct {
			IsAuthenticated bool
			Count           int
			Endpoint        string
			SelectedMenu    string
		}

		data := customData{
			IsAuthenticated: true,
			Count:           count,
			Endpoint:        endpoint,
			SelectedMenu:    "group1",
		}
		tpl.ExecuteTemplate(c.Writer, "group1.gohtml", data)
	} else {
		c.JSON(401, gin.H{"Error": "Unauthorized Request!"})
	}
}

// GetVersion version json
func (ctrl *TestController) Endpoint2(c *gin.Context) {
	if IsAuthenticated(c.Request) {

		session := sessions.Default(c)
		var count int
		v := session.Get("count")
		e := session.Get("endpoints")
		var endpoint string
		if v == nil {
			count = 0
			endpoint = "endpoint2"
		} else {
			count = v.(int)
			count++
			endpoint = e.(string) + ".endpoint2"
		}
		session.Set("count", count)
		session.Set("endpoints", endpoint)
		session.Save()
		type customData struct {
			IsAuthenticated bool
			Count           int
			Endpoint        string
			SelectedMenu    string
		}

		data := customData{
			IsAuthenticated: true,
			Count:           count,
			Endpoint:        endpoint,
			SelectedMenu:    "group2",
		}
		tpl.ExecuteTemplate(c.Writer, "group2.gohtml", data)
	} else {
		c.JSON(401, gin.H{"Error": "Unauthorized Request!"})
	}
}
