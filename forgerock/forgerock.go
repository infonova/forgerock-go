package forgerock

import (
	"errors"
	"fmt"
	"github.com/anaskhan96/soup"
	"github.com/go-resty/resty/v2"
	"strings"
)

// ForgeRock client used to login to ForgeRock and any service providers using ForgeRock.
type Client struct {
	baseUrl string
	authUrl string
}

// Credentials used to login to ForgeRock.
type Credentials struct {
	Username string
	Password string
}

// Request and response data for "/json/realms/root/authenticate".
type authData struct {
	AuthID    string `json:"authId"`
	Callbacks []struct {
		Type   string `json:"type"`
		Output []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"output"`
		Input []struct {
			Name  string `json:"name"`
			Value string `json:"value"`
		} `json:"input"`
		ID int `json:"_id"`
	} `json:"callbacks"`
}

// Response data for a successful authentication.
type loginResponse struct {
	TokenID    string `json:"tokenId"`
	SuccessURL string `json:"successUrl"`
	Realm      string `json:"realm"`
}

// Response embedded in a html form used for the SAML exchange.
type ssoRedirectData struct {
	action       string
	samlResponse string
	relayState   string
}

// Creates a new ForgeRock client with the gaven base url.
func New(baseUrl string) (fr *Client, err error) {
	if baseUrl == "" {
		return nil, errors.New("missing ForgeRock base url")
	}

	fr = new(Client)
	fr.baseUrl = baseUrl
	fr.authUrl = baseUrl + "/json/realms/root/authenticate"
	return
}

// Logs in to the ForgeRock Access Management and the application.
//
// Returns resty client with the valid session, if the supplied credentials were correct.
// This can be used to issue further http requests to other systems (GitHub, Zuul, ...),
// which are using ForgeRock as their Identity Provider.
func (fr *Client) Login(appUrl string, credentials Credentials) (restyClient *resty.Client, err error) {
	if credentials.Username == "" || credentials.Password == "" {
		return nil, errors.New("missing username or password for ForgeRock login")
	}

	restyClient = resty.New()

	err = fr.loginToForgeRock(restyClient, credentials)
	if err != nil {
		return nil, err
	}

	err = fr.loginToApp(restyClient, appUrl)
	if err != nil {
		return nil, err
	}

	return
}

// Handles the login to ForgeRock AM.
func (fr *Client) loginToForgeRock(restyClient *resty.Client, credentials Credentials) (err error) {
	_, err = restyClient.R().
		Get(fr.baseUrl)
	if err != nil {
		return errors.New("failed to get initial login page \"" + fr.baseUrl + "\": " + err.Error())
	}

	headers := createForgeRockHeaders()

	authData := new(authData)
	resp, err := restyClient.R().
		SetHeaders(headers).
		SetResult(authData).
		Post(fr.authUrl)
	if err != nil {
		return errors.New("failed to get initial auth data \"" + fr.authUrl + "\": " + err.Error())
	}

	err = authData.fillCredentials(credentials)
	if err != nil {
		return err
	}

	loginResponse := new(loginResponse)
	resp, err = restyClient.R().
		SetHeaders(headers).
		SetBody(authData).
		SetResult(loginResponse).
		Post(fr.authUrl)
	if err != nil {
		return errors.New("failed to login: " + err.Error())
	}
	if resp.StatusCode() >= 400 {
		return errors.New("error response from login: " + resp.String())
	}
	if loginResponse.TokenID == "" {
		return errors.New("failed to login, tokenId is empty: " + resp.String())
	}

	return
}

// Creates headers for requests to ForgeRock.
func createForgeRockHeaders() map[string]string {
	return map[string]string{
		"Accept":             "application/json",
		"Accept-API-Version": "protocol=1.0,resource=2.1",
	}
}

// Adds the credentials to the authentication template.
func (authData *authData) fillCredentials(credentials Credentials) (err error) {
	var usernameFilled, passwordFilled bool

	for _, callback := range authData.Callbacks {
		switch callback.Type {
		case "NameCallback":
			callback.Input[0].Value = credentials.Username
			usernameFilled = true
		case "PasswordCallback":
			callback.Input[0].Value = credentials.Password
			passwordFilled = true
		default:
			return errors.New("unexpected auth data callback type: " + callback.Type)
		}
	}

	if !usernameFilled {
		errMsg := fmt.Sprintf("failed to find NameCallback to fill in username in auth data: %+v", authData)
		return errors.New(errMsg)
	}
	if !passwordFilled {
		errMsg := fmt.Sprintf("failed to find PasswordCallback to fill in password in auth data: %+v", authData)
		return errors.New(errMsg)
	}

	return
}

// Logs in to a ForgeRock login protected application.
func (frClient *Client) loginToApp(restyClient *resty.Client, appUrl string) (err error) {
	resp, err := restyClient.R().
		Get(appUrl)
	if err != nil {
		return errors.New("failed to call application url \"" + appUrl + "\": " + err.Error())
	}
	if resp.StatusCode() >= 400 {
		return errors.New("error response from application url \"" + appUrl + "\": " + resp.String())
	}

	respBody := resp.String()
	if !strings.Contains(respBody, "SAMLResponse") {
		return errors.New("expected response to contain SAMLResponse form: " + respBody)
	}

	ssoRedirectData := parseSsoRedirectData(respBody)

	resp, err = restyClient.R().
		SetFormData(ssoRedirectData.toFormData()).
		Post(ssoRedirectData.action)
	if err != nil {
		return errors.New("failed to post saml response: " + err.Error())
	}
	if resp.StatusCode() >= 400 {
		return errors.New("error response from posting saml response: " + resp.String())
	}

	return
}

// Extracts the `ssoRedirectData` out from the html form in the response body.
func parseSsoRedirectData(body string) ssoRedirectData {
	html := soup.HTMLParse(body)
	form := html.Find("form")
	action := form.Attrs()["action"]
	samlResponse := form.Find("input", "name", "SAMLResponse").Attrs()["value"]
	relayState := form.Find("input", "name", "RelayState").Attrs()["value"]

	return ssoRedirectData{
		action:       action,
		samlResponse: samlResponse,
		relayState:   relayState,
	}
}

// Creates form data parameters for the SAML exchange request to ForgeRock.
func (data ssoRedirectData) toFormData() map[string]string {
	return map[string]string{
		"SAMLResponse": data.samlResponse,
		"RelayState":   data.relayState,
	}
}
