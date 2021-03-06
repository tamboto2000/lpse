package lpse

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/tamboto2000/lpse/html_parse"
	"golang.org/x/net/html"
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
	Code                  string   `json:"code,omitempty"`
	PackageName           string   `json:"packageName,omitempty"`
	Agency                string   `json:"agency,omitempty"`
	Stage                 string   `json:"stage,omitempty"`
	StageURl              string   `json:"stageURl,omitempty"`
	ProcurementSystem     string   `json:"procurementSystem,omitempty"`
	HPSStr                string   `json:"hpsStr,omitempty"`
	Status                string   `json:"status,omitempty"`
	SPSEVer               string   `json:"spseVer,omitempty"`
	Category              string   `json:"category,omitempty"`
	RUPCode               string   `json:"rupCode,omitempty"`
	FundSource            string   `json:"fundSource,omitempty"`
	CreatedAt             Date     `json:"createdAt,omitempty"`
	Description           string   `json:"description,omitempty"`
	WorkUnit              string   `json:"workUnit,omitempty"`
	FiscalYear            string   `json:"fiscalYear,omitempty"`
	Ceiling               float64  `json:"ceiling,omitempty"`
	HPS                   float64  `json:"hps,omitempty"`
	PaymentMethod         string   `json:"paymentMethod,omitempty"`
	WorkLocations         []string `json:"workLocations,omitempty"`
	BusinessQualification string   `json:"businessQualification,omitempty"`

	// TODO
	// Get and format 'Syarat Kualifikasi'

	cl *Client
}

type Date struct {
	Year    int    `json:"year,omitempty"`
	Month   string `json:"month,omitempty"`
	MontInt int    `json:"monthInt,omitempty"`
	Day     int    `json:"day,omitempty"`
}

// Package categories
var (
	Procurement               = "PENGADAAN_BARANG"
	Construction              = "PEKERJAAN_KONSTRUKSI"
	BusinessEntityConsultancy = "KONSULTANSI"
	IndividualConsultancy     = "KONSULTANSI_PERORANGAN"
	Others                    = "JASA_LAINNYA"
)

var monthMap = map[string]string{
	"Januari":   "January",
	"Februari":  "February",
	"Maret":     "March",
	"April":     "April",
	"Mei":       "May",
	"Juni":      "June",
	"Juli":      "July",
	"Agustus":   "August",
	"September": "September",
	"Oktober":   "October",
	"November":  "November",
	"Desember":  "December",
}

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
			StageURl:          rawPkg.getStageURL(),
			HPSStr:            rawPkg.getHPSStr(),
			ProcurementSystem: rawPkg.getProcSys(),
			Status:            rawPkg.getStatus(),
			Category:          rawPkg.getCategory(),
			CreatedAt:         rawPkg.getDate(),
			SPSEVer:           rawPkg.getSpseVer(),
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

func (pkg *Package) Announcement() error {
	req, err := http.NewRequest("GET", host+"/eproc4/lelang/"+pkg.Code+"/pengumumanlelang", nil)
	if err != nil {
		return err
	}

	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip")
	req.Header.Set("Referer", host+"/")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")

	cookies := pkg.cl.cookies.get()
	for _, c := range cookies {
		req.AddCookie(c)
	}

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode > 200 {
		return errors.New(string(body))
	}

	pkg.cl.cookies.set(resp.Cookies())
	parseAnnouncement(body, pkg)

	return nil
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

func parseAnnouncement(raw []byte, pkg *Package) {
	rootNode, err := html_parse.ParseBytes(raw)
	if err != nil {
		panic(err.Error())
	}

	// extract RUPCode and FundSource
	newNodes := rootNode.SearchAllNode(html.ElementNode, "tr", "", "", "")
	for _, newNode := range newNodes {
		if innerNode1 := newNode.SearchNode(html.ElementNode, "th", "", "", ""); innerNode1 != nil {
			if innerNode2 := innerNode1.SearchNode(html.TextNode, "", "", "", ""); innerNode2 != nil {
				if innerNode2.Data == "Rencana Umum Pengadaan" {
					if innerNode3 := newNode.SearchNode(html.ElementNode, "table", "", "class", "table table-condensed"); innerNode3 != nil {
						innerNode4s := innerNode3.SearchAllNode(html.ElementNode, "td", "", "", "")
						if len(innerNode4s) > 0 {
							pkg.RUPCode = innerNode4s[0].Childs[0].Data
							pkg.FundSource = innerNode4s[2].Childs[0].Data
						}
					}
				}

				// CreatedAt
				if innerNode2.Data == "Tanggal Pembuatan" {
					if innerNode3 := newNode.SearchNode(html.ElementNode, "td", "", "", ""); innerNode3 != nil {
						dateStr := strings.TrimSpace(innerNode3.Childs[0].Data)
						dateSplit := strings.Split(dateStr, " ")
						dateSplit[1] = monthMap[dateSplit[1]]
						date, err := time.Parse("02 January 2006", strings.Join(dateSplit, " "))
						if err == nil {
							pkg.CreatedAt = Date{
								Year:    date.Year(),
								Month:   date.Month().String(),
								MontInt: int(date.Month()),
								Day:     date.Day(),
							}
						}
					}
				}

				// FiscalYear
				if innerNode2.Data == "Tahun Anggaran" {
					if innerNode3 := newNode.SearchNode(html.ElementNode, "td", "", "", ""); innerNode3 != nil {
						pkg.FiscalYear = strings.TrimSpace(innerNode3.Childs[0].Data)
					}
				}

				if innerNode2.Data == "Nilai Pagu Paket" {
					innerNode3s := newNode.SearchAllNode(html.ElementNode, "td", "", "", "")
					if len(innerNode3s) == 2 {
						// Ceiling
						innerNode3 := innerNode3s[0]
						priceStr := innerNode3.Childs[0].Data
						priceStr = strings.ReplaceAll(priceStr, ".", "")
						priceStr = strings.ReplaceAll(priceStr, ",", ".")
						priceStr = strings.ReplaceAll(priceStr, "Rp ", "")
						price, err := strconv.ParseFloat(priceStr, 64)
						if err == nil {
							pkg.Ceiling = price
						}

						// HPS
						innerNode3 = innerNode3s[0]
						priceStr = innerNode3.Childs[0].Data
						priceStr = strings.ReplaceAll(priceStr, ".", "")
						priceStr = strings.ReplaceAll(priceStr, ",", ".")
						priceStr = strings.ReplaceAll(priceStr, "Rp ", "")
						price, err = strconv.ParseFloat(priceStr, 64)
						if err == nil {
							pkg.HPS = price
						}
					}
				}

				// PaymentMethod
				if innerNode2.Data == "Jenis Kontrak" {
					if innerNode3 := newNode.SearchNode(html.ElementNode, "td", "", "", ""); innerNode3 != nil {
						pkg.PaymentMethod = innerNode3.Childs[0].Data
					}
				}

				// WorkLocations
				if innerNode2.Data == "Lokasi Pekerjaan" {
					if innerNode3s := newNode.SearchAllNode(html.ElementNode, "li", "", "", ""); innerNode3s != nil {
						for _, innerNode3 := range innerNode3s {
							pkg.WorkLocations = append(pkg.WorkLocations, innerNode3.Childs[0].Data)
						}
					}
				}
			}
		}
	}
}
