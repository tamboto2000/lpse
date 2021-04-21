package lpse

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
)

type rawPackageList struct {
	Draw            string        `json:"draw"`
	RecordsTotal    int           `json:"recordsTotal"`
	RecordsFiltered int           `json:"recordsFiltered"`
	Data            []packageData `json:"data"`
}

type Packages struct {
	Page               int        `json:"page"`
	PageSize           int        `json:"pageSize"`
	PageCount          int        `json:"pageCount"`
	ItemsTotal         int        `json:"itemsTotal"`
	ItemsFilteredTotal int        `json:"itemsFilteredTotal"`
	Packages           []*Package `json:"packages"`
	Error              error      `json:"-"`
	agency             string
	search             string
	category           string

	cl *Client
}

type Package struct {
	Code        string `json:"code,omitempty"`
	PackageName string `json:"packageName,omitempty"`
	Agency      string `json:"agency,omitempty"`
	Stage       string `json:"stage,omitempty"`
	StageURl    string `json:"stageURl,omitempty"`
	HPS         int    `json:"hps,omitempty"`
	// Sistem Pengadaan
	ProcurementSystem string `json:"procurementSystem,omitempty"`
	HPSStr            string `json:"hpsStr,omitempty"`
	Status            string `json:"status,omitempty"`
	// Nilai Pagu Paket
	Ceiling     int    `json:"ceiling,omitempty"`
	Category    string `json:"category,omitempty"`
	Description string `json:"description,omitempty"`
	CreatedAt   string `json:"createdAt,omitempty"`

	cl *Client
}

// Package categories
var (
	Procurement               = "PENGADAAN_BARANG"
	Construction              = "PEKERJAAN_KONSTRUKSI"
	BusinessEntityConsultancy = "KONSULTANSI"
	IndividualConsultancy     = "KONSULTANSI_PERORANGAN"
	Others                    = "JASA_LAINNYA"
)

func (cl *Client) Packages(pageSize int, agency, search, category string) *Packages {
	return &Packages{
		PageSize: pageSize,
		agency:   agency,
		search:   search,
		category: category,
		cl:       cl,
	}
}

func (pkgs *Packages) Next(page int) bool {
	if page == pkgs.PageCount {
		return false
	}

	pkgs.Page = page
	start := (page - 1) * pkgs.PageSize

	rawPkgs, err := pkgs.cl.reqPackageList(page, start, pkgs.PageSize, pkgs.agency, pkgs.search, pkgs.category)
	if err != nil {
		pkgs.Error = err
		return false
	}

	pkgs.Packages = make([]*Package, 0)
	for _, rawPkg := range rawPkgs.Data {
		pkgs.Packages = append(pkgs.Packages, &Package{
			Code:              rawPkg.getCode(),
			PackageName:       rawPkg.getPkgName(),
			Agency:            rawPkg.getAgency(),
			Stage:             rawPkg.getStage(),
			HPSStr:            rawPkg.getHPSStr(),
			ProcurementSystem: rawPkg.getProcSys(),
			Status:            rawPkg.getStatus(),
			Category:          rawPkg.getCategory(),
			cl:                pkgs.cl,
		})
	}

	pageCount := rawPkgs.RecordsFiltered / pkgs.PageSize
	if pageCount*pkgs.PageSize < rawPkgs.RecordsFiltered {
		pageCount++
	}

	pkgs.PageCount = pageCount
	pkgs.ItemsTotal = rawPkgs.RecordsTotal
	pkgs.ItemsFilteredTotal = rawPkgs.RecordsFiltered

	return true
}

func (cl *Client) reqPackageList(draw, start, length int, agency, search, category string) (*rawPackageList, error) {
	req, err := http.NewRequest("GET", host+"/eproc4/dt/lelang", nil)
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

	req.URL.RawQuery = "columns[0][data]=0&columns[0][name]=&columns[0][searchable]=true&columns[0][orderable]=true&columns[0][search][value]=&columns[0][search][regex]=false&columns[1][data]=1&columns[1][name]=&columns[1][searchable]=true&columns[1][orderable]=true&columns[1][search][value]=&columns[1][search][regex]=false&columns[2][data]=2&columns[2][name]=&columns[2][searchable]=true&columns[2][orderable]=true&columns[2][search][value]=&columns[2][search][regex]=false&columns[3][data]=3&columns[3][name]=&columns[3][searchable]=false&columns[3][orderable]=false&columns[3][search][value]=&columns[3][search][regex]=false&columns[4][data]=4&columns[4][name]=&columns[4][searchable]=true&columns[4][orderable]=true&columns[4][search][value]=&columns[4][search][regex]=false&order[0][column]=0&order[0][dir]=desc&search[regex]=false"
	q := req.URL.Query()

	q.Add("authenticityToken", cl.authToken1)
	q.Add("_", cl.authToken2)
	q.Add("draw", strconv.Itoa(draw))
	q.Add("start", strconv.Itoa(start))
	q.Add("length", strconv.Itoa(length))

	if agency != "" {
		q.Add("rkn_nama", agency)
	}

	if search != "" {
		q.Add("search[value]", search)
	}

	if category != "" {
		q.Add("kategori", category)
	}

	req.URL.RawQuery = q.Encode()
	cookies := cl.cookies.get()
	for _, c := range cookies {
		req.AddCookie(c)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode > 200 {
		return nil, errors.New(string(body))
	}

	rawPackageList := new(rawPackageList)
	if json.Unmarshal(body, rawPackageList); err != nil {
		return nil, err
	}

	cl.cookies.set(resp.Cookies())

	return rawPackageList, nil
}
