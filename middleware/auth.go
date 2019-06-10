package middleware

import (
	"encoding/json"
	"strings"

	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.pkgms.com/techops/peak-self-serve/common"
	"github.pkgms.com/techops/peak-self-serve/config"
)

var (
	sessionStore sessions.CookieStore
)

func init() {
	sessionStore = common.SessionStore
}

func AuthMiddleware(c *gin.Context) {
	session, _ := sessionStore.Get(c.Request, "okta-hosted-login-session-store")
	// Get ACL from config
	acl := config.EndpointACL.ACL
	userGrp := make(map[string]string)

	// Get request url
	urlTmp := c.Request.URL.String()
	i := strings.Index(urlTmp, "/")
	url := urlTmp[i+1:]

	if url != "" && urlTmp[i+1:5] == "auth" {
		url = "login"
	}
	// Middleware will work for only for "endpoint1", "endpoint2"
	if url == "login" || url == "logout" || url == "favicon.ico" || url == "profile" || url == "css/style.css" || url == "submitComment" || url == "" {
		c.Next()
	} else {
		if session.Values["groups"] != nil {
			allowFlag := false
			groups := []string{}
			_ = json.Unmarshal([]byte(session.Values["groups"].(string)), &userGrp)
			for k, _ := range userGrp {
				groups = append(groups, k)
				endpointArr := acl[k]
				for j := 0; j < len(endpointArr); j++ {
					if endpointArr[j] == url {
						allowFlag = true
					}
				}
			}
			if allowFlag != true {
				c.JSON(401, gin.H{"Error": "Unauthorized Request!"})
				c.Abort()
			}
		}
	}

	c.Next()
}
