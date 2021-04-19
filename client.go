package lpse

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io/ioutil"
	"net/http"
	"regexp"
	"strings"
)

const host = "https://lpse.pu.go.id"
const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:86.0) Gecko/20100101 Firefox/86.0"

type Client struct {
	cookies    *cookies
	authToken1 string
	authToken2 string
}

func NewClient() *Client {
	return &Client{cookies: newCookies()}
}

func (cl *Client) Init() error {
	req, err := http.NewRequest("GET", host+"/eproc4/lelang", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	buff, err := decompressResponseBody(resp)
	if err != nil {
		return err
	}

	body, err := ioutil.ReadAll(buff)
	if err != nil {
		return err
	}

	if resp.StatusCode > 200 {
		return errors.New(string(body))
	}

	// extract auth token
	for _, c := range resp.Cookies() {
		if c.Name == "SPSE_SESSION" {
			if token := extractAuthTokenPart1(c.Raw); token != "" {
				cl.authToken1 = token
			} else {
				return errors.New("authenticity token (1) not found")
			}

			if token := extractAuthTokenPart2(c.Raw); token != "" {
				cl.authToken2 = token
			} else {
				return errors.New("authenticity token (2) not found")
			}

			break
		}
	}

	cl.cookies.set(resp.Cookies())

	return nil
}

func decompressResponseBody(resp *http.Response) (*bytes.Buffer, error) {
	buff := new(bytes.Buffer)
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			return nil, err
		}

		buff.ReadFrom(reader)
	default:
		buff.ReadFrom(resp.Body)
	}

	return buff, nil
}

func extractAuthTokenPart1(val string) string {
	rgx := regexp.MustCompile(`___AT=([a-zA-Z0-9_.-]*)`)
	matches := rgx.FindAllString(val, 1)
	if matches != nil {
		token := matches[0]
		token = strings.ReplaceAll(token, "___AT=", "")
		return token
	}

	return ""
}

func extractAuthTokenPart2(val string) string {
	rgx := regexp.MustCompile(`___TS=([a-zA-Z0-9_.-]*)`)
	matches := rgx.FindAllString(val, 1)
	if matches != nil {
		token := matches[0]
		token = strings.ReplaceAll(token, "___TS=", "")
		return token
	}

	return ""
}
