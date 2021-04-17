package lpse

import (
	"bytes"
	"compress/gzip"
	"errors"
	"io/ioutil"
	"net/http"
)

const host = "https://lpse.pu.go.id"
const userAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64; rv:86.0) Gecko/20100101 Firefox/86.0"

type rawPackage struct {
	Draw            string     `json:"draw"`
	RecordsTotal    int64      `json:"recordsTotal"`
	RecordsFiltered int64      `json:"recordsFiltered"`
	Data            [][]string `json:"data"`
}

type Packages struct {
	Page     int       `json:"page"`
	PageSize int       `json:"pageSize"`
	Tenders  []Package `json:"tenders"`
}

type Package struct {
	Code        string  `json:"code,omitempty"`
	PackageName string  `json:"packageName,omitempty"`
	Agency      string  `json:"agency,omitempty"`
	Stage       string  `json:"stage,omitempty"`
	StageURl    string  `json:"stageURl,omitempty"`
	HPS         float64 `json:"hps,omitempty"`
}

type Client struct {
	cookies *cookies
}

func NewClient() *Client {
	return &Client{cookies: newCookies()}
}

func (cl *Client) Init() error {
	req, err := http.NewRequest("GET", host, nil)
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

	cl.cookies.set(resp.Cookies())

	return nil
}

func (cl *Client) reqPackage(page, pageSize int) (*rawPackage, error) {
	req, err := http.NewRequest("GET", host, nil)
	if err != nil {
		return nil, err
	}

	req.Header.Set("Accept", "application/json, text/javascript, */*; q=0.01")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("X-Requested-With", "XMLHttpRequest")

	return nil, nil
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
