package core

import (
	"errors"
	"net/http"
	"net/url"
)

type Method string

type Core struct {
	user string
	pass string
}

const (
	MethodGet  Method = http.MethodGet
	MethodPost Method = http.MethodPost
)

var (
	ErrMethodNotAllowed error = errors.New("method not allowed")
)

func (c Core) CallApi(method Method, namespace, apiName string, data url.Values) (*http.Response, error) {
	urlPath := "http://sms.mysmsmasking.com/" + namespace + "/" + apiName + ".php"
	data.Add("username", c.user)
	data.Add("password", c.pass)

	switch method {
	case MethodGet:
		return http.Get(urlPath + "?" + data.Encode())
	case MethodPost:
		return http.PostForm(urlPath, data)
	}

	return nil, ErrMethodNotAllowed
}

func New(username, password string) Core {
	return Core{
		user: username,
		pass: password,
	}
}
