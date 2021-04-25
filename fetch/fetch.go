package fetch

import (
	"bufio"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type FetchService struct {
	baseURL      string
	pkgMap       map[string]string
	numberOfPkgs int
}

func NewFetchService(url string, numberOfPkgs int) FetchService {
	f := FetchService{
		baseURL:      url,
		pkgMap:       make(map[string]string),
		numberOfPkgs: numberOfPkgs,
	}
	return f
}

func (f *FetchService) FetchPkgList() []Package {
	client := http.DefaultClient
	resp, err := client.Get(f.baseURL + "/PACKAGES")
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("fetch package list fail, error: %v, status code %v\n", err, resp.StatusCode)
		return []Package{}
	}
	pkgs := parsePkgResponse(resp.Body, f.pkgMap, f.numberOfPkgs)
	downloadPackages(pkgs, f.baseURL)
	return pkgs
}

func parsePkgResponse(data io.Reader, pkgMap map[string]string, numberOfPkgs int) []Package {
	pkgs := []Package{}
	scanner := bufio.NewScanner(data)
	count := 0
	for scanner.Scan() {
		if count > numberOfPkgs {
			break
		}
		pkgline := scanner.Text()
		if strings.Contains(pkgline, "Package: ") {
			newPkg := Package{}
			newPkg.name = strings.Split(pkgline, ": ")[1]
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "Version: ") {
					newPkg.version = strings.Split(line, ": ")[1]
				} else if strings.Contains(line, "MD5sum: ") {
					newPkg.md5sum = strings.Split(line, ": ")[1]
					break
				}
			}
			pkgs = append(pkgs, newPkg)
			count += 1
		}
	}
	return pkgs
}
