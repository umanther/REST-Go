package restgo

import (
	"net/url"
	"fmt"
	"net/http"
	"log"
	"bytes"
	"encoding/base64"
	"os"
	"strings"
	"time"
	"errors"
)

var debugLogger *log.Logger

func init() {
	debugLogger = log.New(os.Stderr, "restgo) ", log.LstdFlags)
}

//-------------------------
// ----- E N U M S -----

// HTMLMethod enum ------

// Allowed HTML Methods used for requests
type HTMLMethod string

const (
	MethodGet    HTMLMethod = "GET"
	MethodPut    HTMLMethod = "PUT"
	MethodPost   HTMLMethod = "POST"
	MethodPatch  HTMLMethod = "PATCH"
	MethodDelete HTMLMethod = "DELETE"
)

// HTMlResponse enum -----

// List of possible HTML responses
type HTMLResponse int

const (
	ResponseSuccessBody      HTMLResponse = 200
	ResponseCreated          HTMLResponse = 201
	ResponseSuccessNoBody    HTMLResponse = 204
	ResponseBadRequest       HTMLResponse = 400
	ResponseUnauthorized     HTMLResponse = 401
	ResponseForbidden        HTMLResponse = 403
	ResponseNotFound         HTMLResponse = 404
	ResponseMethodNotAllowed HTMLResponse = 405
)

//-------------------------
// ----- G L O B A L S -----

var defaultHeader = map[string]string{"X-Accept": "All"}

//-------------------------
// ----- T Y P E S -----

// Used to hold query parameters for requests
type QueryParameter struct {
	Key   string
	Value string
}

// Storage for a ServiceNow api connection
type apiConnection struct {
	baseURL           url.URL
	connected         bool
	sessionKey        string
	credentials       string
	additionalHeaders map[string]string
}

//-------------------------
// ----- F U N C T I O N S -----

// ---------- New Object Initializer ----------

// Creates and inializes a new apiConnection
func NewAPIConnection(baseURL string) (*apiConnection, error) {

	if u, err := url.Parse(baseURL); err != nil {
		return nil, err
	} else {
		return &apiConnection{baseURL: *u, additionalHeaders: make(map[string]string)}, nil
	}
}

// ---------- Action Functions ----------

// Takes an apiConnection, HTML Method, resoucre name, optional value, parameters and returns an http.Responce and error status
func RestRequest(con apiConnection, method HTMLMethod, resource string, value string, params []QueryParameter) (resp *http.Response, err error) {

	// Verify Connect has been executed first
	if ! con.IsConnected() {
		return nil, errors.New(`RestRequest: Connection not established.  Execute "Connect" first.`)
	}

	var requestString string

	// If no resource is used, default to full path
	if resource == `` || resource == `/` {
		requestString = con.GetFullPath()

	} else {
		// Pre-process 'resource'
		// Strip any leading /'s
		for resource[0] == byte('/') {
			resource = resource[1:]
		}
		// Add leading '/'
		resource = "/" + resource

		// Strip any trailing /'s
		for resource[len(resource)-1] == byte('/') {
			resource = resource[:len(resource)-1]
		}
		// Pre-process 'value'
		if value != "" {
			// Strip any leading /'s
			for value[0] == byte('/') {
				value = value[1:]
			}
			// Add leading '/'
			value = "/" + value

			// Strip any trailing /'s
			for value[len(value)-1] == byte('/') {
				value = value[:len(value)-1]
			}
		}

		// Combine final requestString
		requestString = con.GetFullPath() + resource + value
	}

	// Process any parameters provided and add them to requestString
	if len(params) > 0 {
		requestString += "?"
		spacer := ""
		for _, item := range params {
			requestString += spacer + url.PathEscape(item.Key) + "=" + url.PathEscape(item.Value)
			spacer = "&"
		}
	}

	// Create new "client" to process the request
	requestClient := &http.Client{Timeout: time.Second * 30}

	// Create new request from requestString
	req, err := http.NewRequest(string(method), requestString, nil)

	if err != nil {
		return nil, err
	}

	for key, value := range defaultHeader {
		req.Header.Add(key, value)
	}

	for key, value := range con.additionalHeaders {
		req.Header.Add(key, value)
	}

	for key, value := range con.additionalHeaders {
		req.Header.Add(key, value)
	}

	req.Header.Add("Authorization", "Basic "+con.credentials)

	debugLogger.Printf("Requesting: %s", req.URL)

	resp, err = requestClient.Do(req)

	return
}

// Establishes a connection with API server
func (con *apiConnection) Connect(username string, password string) (ok error) {

	encodedCreds := new(bytes.Buffer)
	encoder := base64.NewEncoder(base64.StdEncoding, encodedCreds)
	encoder.Write([]byte(username + ":" + password))
	encoder.Close()
	con.credentials = encodedCreds.String()

	con.connected = true

	return
}

// ---------- Get Functions ----------

// Returns folder path of connection
func (con *apiConnection) GetPath() string {
	return fmt.Sprintf("%s", con.baseURL.Path)
}

// Returns full path of connection, including hostname
func (con *apiConnection) GetFullPath() string {
	return fmt.Sprintf("%s://%s%s", con.baseURL.Scheme, con.baseURL.Host, con.baseURL.Path)
}

// Returns if a connection can be established or not
func (con *apiConnection) IsConnected() bool {
	return con.connected
}

// Returns a value from additionalHeaders
func (con *apiConnection) GetHeader(Key string) (value string, ok bool) {
	value, ok = con.additionalHeaders[Key]
	return
}

// Returns base hostname of connection
func (con *apiConnection) GetBaseURL() string {
	return fmt.Sprintf("%s://%s", con.baseURL.Scheme, con.baseURL.Host)
}

// ---------- Add/Set Functions ----------

// Adds a header value to additionalHeaders
func (con *apiConnection) SetHeader(Key string, Value string) {
	con.ChangeHeader(Key, Value)
}

// Sets folder path of connection URL
func (con *apiConnection) SetPath(path string) {
	path = strings.TrimSpace(path)

	// Strip off trailing /'s
	for path[len(path)-1:] == `/` {
		path = path[:len(path)-1]
	}

	con.baseURL.Path = path
}

// Sets base URL of connection with default path
func (con *apiConnection) SetBaseURL(baseURL string) {
	u, err := url.ParseRequestURI(baseURL)
	if err != nil {
		log.Fatal(err)
	}
	con.baseURL = *u
}

// ---------- Modify Functions ----------

// Modifis a value in additionalHeaders
func (con *apiConnection) ChangeHeader(Key string, Value string) {
	con.additionalHeaders[Key] = Value
}

// ---------- Remove Functions ----------

// Deletes a value in additionalHeaders
func (con *apiConnection) RemoveHeader(Key string) {
	delete(con.additionalHeaders, Key)
}
