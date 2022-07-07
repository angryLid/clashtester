package speedtest

import (
	"context"
	"crypto/tls"
	"encoding/json"
	"encoding/xml"
	"errors"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"time"
)

var httpClient *http.Client

func init() {
	proxy, _ := url.Parse("http://localhost:7890")
	tr := &http.Transport{
		Proxy:           http.ProxyURL(proxy),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	httpClient = &http.Client{
		Transport: tr,
		Timeout:   time.Second * 30, //超时时间
	}
}

type Client struct {
	inner *http.Client

	user *User
}

func NewClient(c *http.Client) *Client {
	return &Client{inner: c}
}
func Setup() *Client {
	proxy, _ := url.Parse("http://localhost:7890")
	tr := &http.Transport{
		Proxy:           http.ProxyURL(proxy),
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}

	httpClient := &http.Client{
		Transport: tr,
		Timeout:   time.Second * 30, //超时时间
	}

	client := NewClient(httpClient)

	client.FetchUserInfo()
	return client
}
func (c *Client) GetInnerClient() *http.Client {
	return c.inner
}

func (c *Client) CurrentUser() *User {
	return c.user
}

func (c *Client) FetchUserInfo() (*User, error) {
	req, err := http.NewRequest(http.MethodGet, speedTestConfigUrl, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.inner.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var users Users
	bytes, _ := ioutil.ReadAll(resp.Body)
	// err = ioutil.WriteFile("./settings.xml", bytes, 0644)
	if err != nil {
		log.Fatal(err)
	}
	xml.Unmarshal(bytes, &users)
	if len(users.Users) == 0 {
		return nil, errors.New("failed to fetch user information")
	}
	c.user = &users.Users[0]
	return &users.Users[0], nil
}

func (c *Client) FetchServerList(ctx context.Context) (ServerList, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, speedTestServersUrl, nil)
	if err != nil {
		return nil, err
	}

	resp, err := c.inner.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	var serverList ServerList

	decoder := json.NewDecoder(resp.Body)

	if err := decoder.Decode(&serverList); err != nil {
		return serverList, err
	}
	if len(serverList) == 0 {
		return nil, errors.New("failed to fetch servers")
	}

	return serverList, nil
}
