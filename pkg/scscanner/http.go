package scscanner

import (
	"crypto/tls"
	"fmt"
	"io/ioutil"
	"net/http"
)

type Response struct {
	Server      string
	ContentType string
	StatusCode  int
	Body        []byte
}

type HttpErr struct {
	err error
}

type HTTPHeader struct {
	Name  string
	Value string
}

type HTTPClient struct {
	client    *http.Client
	userAgent string
	headers   []HTTPHeader
	cookies   string
	method    string
	//	host      string
}

func AddTraversal(url string) []string {
	var traversal_urls_list []string
	short_payloads_list := []string{"../", "..%2f", "..%2f%26", "..", "..\\"}
	for _, payload := range short_payloads_list {
		traversal_urls_list = append(traversal_urls_list, url+payload)
	}
	return traversal_urls_list
}

func NewHTTPClient(opt *Options) (*HTTPClient, error) {
	//var proxyURLFunc func(*http.Request) (*url.URL, error)
	var client HTTPClient
	//proxyURLFunc = http.ProxyFromEnvironment

	if opt == nil {
		return nil, fmt.Errorf("options is nil")
	}

	// if opt.Proxy != "" {
	// 	proxyURL, err := url.Parse(opt.Proxy)
	// 	if err != nil {
	// 		return nil, fmt.Errorf("proxy URL is invalid (%w)", err)
	// 	}
	// 	proxyURLFunc = http.ProxyURL(proxyURL)
	// }

	var redirectFunc func(req *http.Request, via []*http.Request) error
	if !opt.FollowRedirect {
		redirectFunc = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else {
		redirectFunc = nil
	}

	client.client = &http.Client{
		Timeout:       opt.Timeout,
		CheckRedirect: redirectFunc,
		Transport: &http.Transport{
			//Proxy:               proxyURLFunc,
			MaxIdleConns:        100,
			MaxIdleConnsPerHost: 100,
			TLSClientConfig: &tls.Config{
				InsecureSkipVerify: opt.NoTLSValidation,
			},
		}}
	client.userAgent = opt.UserAgent
	client.headers = opt.Headers
	client.cookies = opt.Cookies
	client.method = opt.Method
	if client.method == "" {
		client.method = http.MethodGet
	}
	return &client, nil
}

func (client *HTTPClient) SetRedirects(flag bool) {
	var redirectFunc func(req *http.Request, via []*http.Request) error
	if !flag {
		redirectFunc = func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		}
	} else {
		redirectFunc = nil
	}
	client.client.CheckRedirect = redirectFunc
}

func (client *HTTPClient) CreateResponse(hostname string, urlPath string) (*Response, error) {
	req, err := http.NewRequest(client.method, hostname, nil)
	if err != nil {
		return nil, err
	}
	req.URL.Opaque = urlPath
	if client.cookies != "" {
		req.Header.Set("Cookie", client.cookies)
	}

	if client.userAgent != "" {
		req.Header.Set("User-Agent", client.userAgent)
	} else {
		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:78.0) Gecko/20100101 Firefox/78.0")
	}

	// add custom headers
	for _, h := range client.headers {
		req.Header.Set(h.Name, h.Value)
	}

	resp, err := client.client.Do(req)
	if err != nil {
		return &Response{}, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	var target_response Response
	target_response.Server = resp.Header.Get("Server")
	target_response.ContentType = resp.Header.Get("Content-Type")
	target_response.StatusCode = resp.StatusCode
	target_response.Body = body
	return &target_response, err
}
