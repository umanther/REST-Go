package restgo

import (
	"net/url"
	"fmt"
	"net/http"
	"errors"
	"log"
	"bytes"
	"encoding/base64"
)

// htmlMethod enum ------
type htmlMethod string

const (
	MethodGet    htmlMethod = "GET"
	MethodPut    htmlMethod = "PUT"
	MethodPost   htmlMethod = "POST"
	MethodPatch  htmlMethod = "PATCH"
	MethodDelete htmlMethod = "DELETE"
)

//------------------------


// htmlResponses enum -----
type htmlResponse int

const (
	responseSuccessBody      htmlResponse = 200
	responseCreated          htmlResponse = 201
	responseSuccessNoBody    htmlResponse = 204
	responseBadRequest       htmlResponse = 400
	responseUnauthorized     htmlResponse = 401
	responseForbidden        htmlResponse = 403
	responseNotFound         htmlResponse = 404
	responseMethodNotAllowed htmlResponse = 405
)

//-------------------------

// Used to hold query parameters for requests
type QueryParameter struct {
	Key   string
	Value string
}

// Storage for a ServiceNow api connection
type apiConnection struct {
	baseURL     url.URL
	connected   bool
	sessionKey  string
	credentials string
}

func NewAPIConnection(url string) *apiConnection {
	con := new(apiConnection)
	con.SetBaseURL(url)
	con.connected = false

	return con
}

func restRequest(con *apiConnection, method htmlMethod, resource string, value string, params []QueryParameter) (resp *http.Response, err error) {

	if resource == "" || resource == "/" {
		return nil, errors.New("restRequest: Resource name required")
	}

	for resource[0] == byte('/') {
		resource = resource[1:]
	}

	for resource[len(resource)-1] == byte('/') {
		resource = resource[0:len(resource)-2]
	}
	resource = "/" + resource

	if value != "" {
		for value[0] == byte('/') {
			value = value[1:]
		}
		value = "/" + value

		for value[len(value)-1] == byte('/') {
			value = value[0:len(value)-2]
		}
	}

	requestString := con.GetFullPath() + resource + value

	if len(params) > 0 {
		requestString += "?"
		spacer := ""
		for _, item := range params {
			requestString += spacer + url.PathEscape(item.Key) + "=" + url.PathEscape(item.Value)
			spacer = "&"
		}
	}

	switch method {
	case MethodGet:
		fmt.Println(requestString)
		resp, err = http.Get(requestString)
	default:
		return nil, errors.New("restRequest: Unsupported method requested")
	}

	return
}

func (con *apiConnection) Connect(username string, password string) error {
	if con.baseURL == *new(url.URL) {
		return errors.New("Missing URL: Set a URL value before calling Connect()")
	}

	encodedCreds := new(bytes.Buffer)
	base64.NewEncoder(base64.StdEncoding, encodedCreds).Write([]byte(username + ":" + password))
	con.credentials = encodedCreds.String()

	resp, err := restRequest(con, MethodGet, "/garbage", "", nil)

	if err != nil {
		return err
	}

	if resp.Header.Get("Server") == "ServiceNow" {
		con.connected = true
		return nil
	} else {
		return err
	}
}

func (con *apiConnection) SetBaseURL(baseURL string) {
	u, err := url.ParseRequestURI(baseURL)
	if err != nil {
		log.Fatal(err)
	}
	con.baseURL = *u
	con.baseURL.Path = "/api/now"
}

func (con *apiConnection) GetBaseURL() string {
	return fmt.Sprintf("%s://%s", con.baseURL.Scheme, con.baseURL.Host)
}

func (con *apiConnection) GetFullPath() string {
	return fmt.Sprintf("%s://%s%s", con.baseURL.Scheme, con.baseURL.Host, con.baseURL.Path)
}

func (con *apiConnection) IsConnected() bool {
	return con.connected
}