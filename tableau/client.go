package tableau

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/hashicorp/go-cleanhttp"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
)

const (
	ErrCodeInternal = "-1" // Internal error.
	jsonMediaType   = "application/json"
)

const (
	libraryVersion = "v0.0.1"
	userAgent      = "go-tableau/" + libraryVersion
)

// Client encapsulates a client that talks to the Tableau API
type Client struct {
	client *http.Client

	UserAgent string

	headers map[string]string

	baseURL *url.URL

	SiteID string

	DataSources *dataSourcesService
	Projects    *projectsService
}

// NewClient instantiates an instance of the Tableau API client.
func NewClient(serverAddr, personalAccessTokenName, personalAccessTokenSecret, site string) (*Client, error) {
	baseURL, err := url.Parse(serverAddr + "/api/3.4/")
	if err != nil {
		return nil, err
	}

	c := &Client{
		client:    cleanhttp.DefaultClient(),
		baseURL:   baseURL,
		UserAgent: userAgent,
		headers:   make(map[string]string, 0),
	}

	err = c.signIn(personalAccessTokenName, personalAccessTokenSecret, site)
	if err != nil {
		return nil, err
	}
	c.DataSources = &dataSourcesService{client: c}
	c.Projects = &projectsService{client: c}
	return c, nil
}

type signInRequest struct {
	Credentials credentials `json:"credentials"`
}

type credentials struct {
	Name        string `json:"name"`
	Password    string `json:"password"`
	TokenName   string `json:"personalAccessTokenName"`
	TokenSecret string `json:"personalAccessTokenSecret"`
	Site        site   `json:"site"`
}

type site struct {
	ID         string `json:"id"`
	ContentUrl string `json:"contentUrl"`
}

type signInResponse struct {
	Credentials struct {
		Site                      site   `json:"site"`
		Token                     string `json:"token"`
		EstimatedTimeToExpiration string `json:"estimatedTimeToExpiration"`
	}
}

// sign in to Tableau API and fetch token for futures requests.
func (c *Client) signIn(personalAccessTokenName, personalAccessTokenSecret, siteName string) error {
	signInRequest := signInRequest{
		Credentials: credentials{
			TokenName:   personalAccessTokenName,
			TokenSecret: personalAccessTokenSecret,
			Site: site{
				ContentUrl: siteName,
			},
		},
	}

	req, err := c.newRequest(http.MethodPost, "auth/signin", signInRequest)
	if err != nil {
		return errors.Wrap(err, "error creating request auth/signin")
	}

	resp := &signInResponse{}
	err = c.do(context.TODO(), req, resp)
	if err != nil {
		return err
	}
	c.headers["X-Tableau-Auth"] = resp.Credentials.Token
	c.SiteID = resp.Credentials.Site.ID
	return nil
}

// do makes an HTTP request and populates the given struct v from the response.
func (c *Client) do(ctx context.Context, req *http.Request, v interface{}) error {
	req = req.WithContext(ctx)
	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	return c.handleResponse(ctx, res, v)
}

// handleResponse makes an HTTP request and populates the given struct v from
// the response.  This is meant for internal testing and shouldn't be used
// directly. Instead please use `Client.do`.
func (c *Client) handleResponse(ctx context.Context, res *http.Response, v interface{}) error {
	out, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return err
	}

	if res.StatusCode >= 400 {
		// errorResponse represents an error response from the API
		type errorResponse struct {
			Error struct {
				Summary string `json:"summary"`
				Detail  string `json:"detail"`
				Code    string `json:"code"`
			}
		}

		errorRes := &errorResponse{}
		err = json.Unmarshal(out, errorRes)
		if err != nil {
			var jsonErr *json.SyntaxError
			if errors.As(err, &jsonErr) {
				return &Error{
					msg:  "malformed error response body received",
					Code: ErrCodeInternal,
					Meta: map[string]string{
						"body":        string(out),
						"err":         jsonErr.Error(),
						"http_status": http.StatusText(res.StatusCode),
					},
				}
			}
			return err
		}

		if *errorRes == (errorResponse{}) {
			return &Error{
				msg:  "internal error, response body doesn't match error type signature",
				Code: ErrCodeInternal,
				Meta: map[string]string{
					"body":        string(out),
					"http_status": http.StatusText(res.StatusCode),
				},
			}
		}

		return &Error{
			msg:  errorRes.Error.Summary + ": " + errorRes.Error.Detail,
			Code: errorRes.Error.Code,
		}
	}

	// this means we don't care about unmarshaling the response body into v
	if v == nil || res.StatusCode == http.StatusNoContent {
		return nil
	}

	err = json.Unmarshal(out, &v)
	if err != nil {
		var jsonErr *json.SyntaxError
		if errors.As(err, &jsonErr) {
			return &Error{
				msg:  "malformed response body received",
				Code: ErrCodeInternal,
				Meta: map[string]string{
					"body":        string(out),
					"http_status": http.StatusText(res.StatusCode),
				},
			}
		}
		return err
	}

	return nil
}

func (c *Client) newRequest(method string, path string, body interface{}) (*http.Request, error) {
	u, err := c.baseURL.Parse(path)
	if err != nil {
		return nil, err
	}

	var req *http.Request
	switch method {
	case http.MethodGet:
		req, err = http.NewRequest(method, u.String(), nil)
		if err != nil {
			return nil, err
		}
	default:
		buf := new(bytes.Buffer)
		if body != nil {
			err = json.NewEncoder(buf).Encode(body)
			if err != nil {
				return nil, err
			}
		}

		req, err = http.NewRequest(method, u.String(), buf)
		if err != nil {
			return nil, err
		}

		req.Header.Set("Content-Type", jsonMediaType)
	}

	req.Header.Set("Accept", jsonMediaType)
	req.Header.Set("User-Agent", c.UserAgent)

	for k, v := range c.headers {
		req.Header.Set(k, v)
	}

	return req, nil
}

// Error represents common errors originating from the Client.
type Error struct {
	// msg contains the human readable string
	msg string

	// code specifies the error code. i.e; NotFound, RateLimited, etc...
	Code string

	// Meta contains additional information depending on the error code. As an
	// example, if the Code is "ErrResponseMalformed", the map will be: ["body"]
	// = "body of the response"
	Meta map[string]string
}

// Error returns the string representation of the error.
func (e *Error) Error() string { return e.msg }
