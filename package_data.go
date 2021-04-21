package lpse

import (
	"regexp"
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
