package lpse

import (
	"regexp"
	"strconv"
	"strings"
)

type packageData []string

func (pkgData packageData) getCode() string {
	return pkgData[0]
}

func (pkgData packageData) getPkgName() string {
	pkgName := pkgData[1]
	rgx := regexp.MustCompile(`<span class='label label-warning'>[a-z A-Z]*</span>`)
	matches := rgx.FindAllString(pkgName, 1)
	if matches != nil {
		pkgName = strings.ReplaceAll(pkgName, matches[0], "")
	}

	return pkgName
}

func (pkgData packageData) getAgency() string {
	return pkgData[2]
}

func (pkgData packageData) getStage() string {
	return strings.ReplaceAll(pkgData[3], " [...]", "")
}

func (pkgData packageData) getHPSStr() string {
	return pkgData[4]
}

func (pkgData packageData) getProcSys() string {
	return pkgData[6] + " - " + pkgData[5] + " - " + pkgData[7]
}

func (pkgData packageData) getStatus() string {
	var status string
	pkgName := pkgData[1]
	rgx := regexp.MustCompile(`<span class='label label-warning'>[a-z A-Z]*</span>`)
	matches := rgx.FindAllString(pkgName, 1)
	if matches != nil {
		status = strings.ReplaceAll(matches[0], "<span class='label label-warning'>", "")
		status = strings.ReplaceAll(status, "</span>", "")
	}

	return status
}

func (pkgData packageData) getCategory() string {
	category := pkgData[8]
	rgx := regexp.MustCompile(` - [A-Z]* [0-9]*`)
	matches := rgx.FindAllString(category, 1)
	if matches != nil {
		category = strings.ReplaceAll(category, matches[0], "")
	}

	return category
}

func (pkgData packageData) getStageURL() string {
	return "https://lpse.pu.go.id/eproc4/lelang/" + pkgData.getCode() + "/jadwal"
}

func (pkgData packageData) getDate() Date {
	var date Date
	str := pkgData[8]
	rgx := regexp.MustCompile(` - [A-Z]* [0-9]*`)
	matches := rgx.FindAllString(str, 1)
	var dateStr string
	if matches != nil {
		dateStr = matches[0]
		dateStr = strings.ReplaceAll(dateStr, " - TA ", "")
		i, err := strconv.Atoi(dateStr)
		if err == nil {
			date.Year = i
		}
	}

	return date
}

func (pkgData packageData) getSpseVer() string {
	ver := pkgData[9]
	if i, err := strconv.Atoi(ver); err == nil {
		if i > 0 {
			return "spse 4." + strconv.Itoa(i)
		}

		return "spse 3"
	}

	return ""
}
