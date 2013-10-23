package heroku

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	neturl "net/url"
	"os"
	"path/filepath"
	"reflect"
	"strings"
)

const (
	DefaultHost = "api.heroku.com"
	userAgent   = "heroku-go"
)

var (
	netrcPath = filepath.Join(os.Getenv("HOME"), ".netrc")
)

type Client struct {
	Host  string // e.g. "api.heroku.com"
	Token string // OAuth token

	// If nil, http.DefaultClient is used.
	Client interface {
		Do(*http.Request) (*http.Response, error)
	}
}

// New returns a new Client for interacting with the Heroku API.
// It initializes Host and Token according to the host and
// password in url, which must use scheme https.
// If url is the empty string, DefaultHost is used.
func New(url string) (*Client, error) {
	if url == "" {
		return &Client{Host: DefaultHost}, nil
	}
	c := new(Client)
	u, err := neturl.Parse(url)
	if err != nil {
		return nil, err
	}
	if u.Scheme != "https" {
		return nil, errors.New("invalid scheme " + u.Scheme)
	}
	c.Host = u.Host
	if u.User != nil {
		c.Token, _ = u.User.Password()
	}
	return c, nil
}

// NewRequest returns a new http Request suitable for sending
// to the Heroku API.
//
//   Accept: application/vnd.heroku+json; version=3
//
// The type of data determines how to encode the request:
//
//   nil         no body
//   io.Reader   data is sent verbatim
//   else        data is encoded as application/json
//
func (c *Client) NewRequest(method, path string, data interface{}) (*http.Request, error) {
	var ctype string
	var body io.Reader
	switch v := data.(type) {
	case nil:
	case io.Reader:
		body = v
	case *bytes.Buffer:
		_ = v
	default:
		j, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		body = bytes.NewReader(j)
		ctype = "application/json"
	}

	req, err := http.NewRequest(method, "https://"+c.Host+path, body)
	if err != nil {
		return nil, err
	}
	req.SetBasicAuth("", c.Token)
	req.Header.Set("Accept", "application/vnd.heroku+json; version=3")
	req.Header.Set("User-Agent", userAgent)
	if ctype != "" {
		req.Header.Set("Content-Type", ctype)
	}
	return req, nil
}

// Submits an HTTP request, checks its response, and deserializes
// the response into w. The type of w determines how to handle
// the response body:
//
//   nil        discard
//   io.Writer  body is copied directly into w
//   else       body is decoded into w as json
//
func (c *Client) Do(req *http.Request, w interface{}) error {
	client := c.Client
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode/100 != 2 {
		return errors.New("Bad status: " + resp.Status)
	}
	switch v := w.(type) {
	case nil:
	case io.Writer:
		_, err = io.Copy(v, resp.Body)
	default:
		err = json.NewDecoder(resp.Body).Decode(w)
	}
	return err
}

func (c *Client) call(method, path string, r, w interface{}) error {
	req, err := c.NewRequest(method, path, r)
	if err != nil {
		return err
	}
	return c.Do(req, w)
}

// join joins path components a into a URL path
// beginning with "/".
func join(a []string) (path string) {
	for _, s := range a {
		// TODO(kr): escape each path component
		path += "/" + s
	}
	return
}

type Pather interface {
	Path() string
}

func dir(path string) (string, error) {
	if i := strings.LastIndex(path, "/"); i != -1 {
		return path[:i], nil
	}
	return "", errors.New("invalid path: " + path)
}

func (c *Client) Info(p Pather) error {
	return c.call("GET", p.Path(), nil, p)
}

func (c *Client) Create(p Pather) error {
	path, err := dir(p.Path())
	if err != nil {
		return err
	}
	return c.call("POST", path, p, p)
}

func (c *Client) Update(p Pather) error {
	return c.call("PATCH", p.Path(), p, nil)
}

func (c *Client) Destroy(p Pather) error {
	return c.call("DELETE", p.Path(), nil, nil)
}

// List lists objects, p must be a pointer to a slice of
// pointer types that implement Pather. List panics if
// p has an invalid type.
func (c *Client) List(p interface{}) error {
	path, err := dir(listPather(p).Path())
	if err != nil {
		return err
	}
	return c.call("GET", path, nil, p)
}

func listPather(v interface{}) Pather {
	t := reflect.TypeOf(v)
	t = mustElem(t, reflect.Ptr)
	t = mustElem(t, reflect.Slice)
	t = mustElem(t, reflect.Ptr)
	return reflect.New(t).Interface().(Pather)
}

func mustElem(t reflect.Type, k reflect.Kind) reflect.Type {
	if t.Kind() != k {
		panic("bad type")
	}
	return t.Elem()
}
