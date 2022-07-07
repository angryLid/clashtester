package clash

import (
	"encoding/json"
	"net/http"
	"strings"
)

type Axios struct {
	URL    string
	Secret string
}

func NewAxios() *Axios {
	c := ReadConfig()
	return &Axios{
		URL:    c.ExternalController,
		Secret: c.Secret,
	}
}

func (a *Axios) Get(url string, body interface{}) (*http.Response, error) {
	return a.request(http.MethodGet, url, nil)
}

func (a *Axios) Patch(url string, body map[string]interface{}) (*http.Response, error) {
	return a.request(http.MethodPatch, url, body)
}

func (a *Axios) Put(url string, body map[string]any) (*http.Response, error) {
	return a.request(http.MethodPut, url, body)
}

func (a *Axios) request(method string, url string, body map[string]interface{}) (*http.Response, error) {
	json, _ := json.Marshal(body)
	reader := strings.NewReader(string(json))
	req, err := http.NewRequest(method, a.URL+url, reader)
	if err != nil {
		return nil, err
	}
	req.Header.Add("Authorization", a.Secret)

	// log.Printf("Axios Client [Base URL %s] [Secret %s]\n", a.URL, a.Secret)
	return http.DefaultClient.Do(req)
}
