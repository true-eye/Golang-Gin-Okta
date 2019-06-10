package controller

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"html/template"
	"io/ioutil"
	"log"
	"net/http"

	"github.com/gin-contrib/location"
	"github.com/gin-gonic/contrib/sessions"
	"github.com/gin-gonic/gin"
	verifier "github.com/okta/okta-jwt-verifier-golang"
	"github.pkgms.com/techops/peak-self-serve/common"
	"github.pkgms.com/techops/peak-self-serve/config"
	"github.pkgms.com/techops/peak-self-serve/utils"
)

var (
	tpl          *template.Template
	nonce        string
	sessionStore sessions.CookieStore
)

func init() {
	tpl = template.Must(template.ParseGlob("templates/*"))
	nonce = common.Nonce
	sessionStore = common.SessionStore
}

type AuthController struct {
}

func (ctrl *AuthController) HomeHandler(c *gin.Context) {
	type customData struct {
		Profile         map[string]string
		IsAuthenticated bool
		SelectedMenu    string
	}

	data := customData{
		Profile:         getProfileData(c.Request),
		IsAuthenticated: IsAuthenticated(c.Request),
		SelectedMenu:    "home",
	}
	tpl.ExecuteTemplate(c.Writer, "home.gohtml", data)
}

func (ctrl *AuthController) LoginHandler(c *gin.Context) {
	nonce, _ = utils.GenerateNonce()
	var redirectPath string
	url := location.Get(c)
	q := c.Request.URL.Query()
	q.Add("client_id", config.Okta.ClientId)
	q.Add("response_type", "code")
	q.Add("response_mode", "query")
	q.Add("scope", "openid profile email")
	q.Add("redirect_uri", url.Scheme+"://"+url.Host+"/authorization-code/callback")
	q.Add("state", config.Okta.State)
	q.Add("nonce", nonce)

	redirectPath = config.Okta.Issuer + "/v1/authorize?" + q.Encode()

	c.Redirect(http.StatusMovedPermanently, redirectPath)
}

func (ctrl *AuthController) AuthCallbackHandler(c *gin.Context) {
	if c.Request.URL.Query().Get("state") != config.Okta.State {
		log.Println("The state was not as expected")
		return
	}
	// Make sure the code was provided
	if c.Request.URL.Query().Get("code") == "" {
		log.Println("The code was not returned or is not accessible")
		return
	}

	exchange := exchangeCode(c.Request.URL.Query().Get("code"), c.Request)

	session, err := sessionStore.Get(c.Request, "okta-hosted-login-session-store")
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}
	_, verificationError := verifyToken(exchange.IdToken)

	if verificationError != nil {
		fmt.Println(verificationError)
	}

	if verificationError == nil {
		session.Values["id_token"] = exchange.IdToken
		session.Values["access_token"] = exchange.AccessToken
		session.Save(c.Request, c.Writer)
	}

	profileMap := getProfileData(c.Request)
	userID := profileMap["sub"]
	result := populateGroups(c, userID)
	if result == "fail" {
		log.Println("Error while fetching user's group information")
	}

	http.Redirect(c.Writer, c.Request, "/", http.StatusMovedPermanently)

}

func (ctrl *AuthController) ProfileHandler(c *gin.Context) {
	if IsAuthenticated(c.Request) {
		type customData struct {
			Profile         map[string]string
			IsAuthenticated bool
			SelectedMenu    string
		}

		data := customData{
			Profile:         getProfileData(c.Request),
			IsAuthenticated: true,
			SelectedMenu:    "profile",
		}
		tpl.ExecuteTemplate(c.Writer, "profile.gohtml", data)
	} else {
		c.JSON(401, gin.H{"Error": "Unauthorized Request!"})
	}
}

func (ctrl *AuthController) Logout(c *gin.Context) {
	session, err := sessionStore.Get(c.Request, "okta-hosted-login-session-store")
	if err != nil {
		http.Error(c.Writer, err.Error(), http.StatusInternalServerError)
	}

	// var redirectPath string
	// url := location.Get(c)

	// if err != nil || session.Values["id_token"] == nil {
	// 	c.Redirect(http.StatusMovedPermanently, "/")
	// }

	// q := c.Request.URL.Query()
	// q.Add("id_token_hink", session.Values["id_token"].(string))
	// q.Add("post_logout_redirect_uri", url.Scheme+"://"+url.Host+"/")

	// redirectPath = config.Okta.Issuer + "/logout?" + q.Encode()

	delete(session.Values, "id_token")
	delete(session.Values, "access_token")
	delete(session.Values, "groups")

	session.Save(c.Request, c.Writer)

	c.Redirect(http.StatusMovedPermanently, "/")

}

func exchangeCode(code string, r *http.Request) Exchange {
	authHeader := base64.StdEncoding.EncodeToString(
		[]byte(config.Okta.ClientId + ":" + config.Okta.ClientSecret))

	q := r.URL.Query()
	q.Add("grant_type", "authorization_code")
	q.Add("code", code)
	q.Add("redirect_uri", "http://localhost:8080/authorization-code/callback")

	url := config.Okta.Issuer + "/v1/token?" + q.Encode()

	req, _ := http.NewRequest("POST", url, bytes.NewReader([]byte("")))
	h := req.Header
	h.Add("Authorization", "Basic "+authHeader)
	h.Add("Accept", "application/json")
	h.Add("Content-Type", "application/x-www-form-urlencoded")
	h.Add("Connection", "close")
	h.Add("Content-Length", "0")

	client := &http.Client{}
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	var exchange Exchange
	json.Unmarshal(body, &exchange)

	return exchange

}

func IsAuthenticated(r *http.Request) bool {
	session, err := sessionStore.Get(r, "okta-hosted-login-session-store")
	if err != nil || session.Values["id_token"] == nil || session.Values["id_token"] == "" {
		return false
	}

	return true
}

func getProfileData(r *http.Request) map[string]string {
	m := make(map[string]string)

	session, err := sessionStore.Get(r, "okta-hosted-login-session-store")

	if err != nil || session.Values["access_token"] == nil || session.Values["access_token"] == "" {
		return m
	}
	reqUrl := config.Okta.Issuer + "/v1/userinfo"

	req, _ := http.NewRequest("GET", reqUrl, bytes.NewReader([]byte("")))
	h := req.Header
	h.Add("Authorization", "Bearer "+session.Values["access_token"].(string))
	h.Add("Accept", "application/json")

	client := &http.Client{}
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	json.Unmarshal(body, &m)

	return m
}

func verifyToken(t string) (*verifier.Jwt, error) {
	tv := map[string]string{}
	tv["nonce"] = nonce
	tv["aud"] = config.Okta.ClientId
	jv := verifier.JwtVerifier{
		Issuer:           config.Okta.Issuer,
		ClaimsToValidate: tv,
	}

	result, err := jv.New().VerifyIdToken(t)

	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}

	if result != nil {
		return result, nil
	}

	return nil, fmt.Errorf("token could not be verified: %s", "")
}

func populateGroups(c *gin.Context, userID string) string {
	session, err := sessionStore.Get(c.Request, "okta-hosted-login-session-store")

	if err != nil || session.Values["access_token"] == nil || session.Values["access_token"] == "" {
		return "fail"
	}
	groupURL := fmt.Sprintf("/users/%v/groups", userID)
	reqUrl := config.Okta.APIURL + groupURL

	req, _ := http.NewRequest("GET", reqUrl, bytes.NewReader([]byte("")))
	h := req.Header
	h.Add("Authorization", "SSWS "+config.Okta.APIToken)
	h.Add("Accept", "application/json")
	h.Add("Content-Type", "application/json")

	client := &http.Client{}
	resp, _ := client.Do(req)
	body, _ := ioutil.ReadAll(resp.Body)
	defer resp.Body.Close()
	groups := []common.Group{}
	json.Unmarshal(body, &groups)

	// Save user's group
	userGrp := make(map[string]string)
	for i := 0; i < len(groups); i++ {
		userGrp[groups[i].Profile.Name] = "yes"
	}
	userGrpData, err := json.Marshal(userGrp)
	if err != nil {
		log.Println("Encoding Group info failed")
	}
	session.Values["groups"] = string(userGrpData)
	session.Save(c.Request, c.Writer)

	return "ok"
}

type Exchange struct {
	Error            string `json:"error,omitempty"`
	ErrorDescription string `json:"error_description,omitempty"`
	AccessToken      string `json:"access_token,omitempty"`
	TokenType        string `json:"token_type,omitempty"`
	ExpiresIn        int    `json:"expires_in,omitempty"`
	Scope            string `json:"scope,omitempty"`
	IdToken          string `json:"id_token,omitempty"`
}
