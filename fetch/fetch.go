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
	numberOfPkgs int
}

func NewFetchService(url string, numberOfPkgs int) FetchService {
	f := FetchService{
		baseURL:      url,
		numberOfPkgs: numberOfPkgs,
	}
	return f
}

func (f *FetchService) FetchPkgList() []Package {
	client := http.DefaultClient
	resp, err := client.Get(f.baseURL + "PACKAGES")
	if err != nil {
		fmt.Printf("fetch package list fail, error: %v\n", err)
		return []Package{}
	}
	if resp.StatusCode != 200 {
		fmt.Printf("fetch package list fail, error: %v, status code %v\n", err, resp.StatusCode)
		return []Package{}
	}
	pkgs := parsePkgResponse(resp.Body, f.numberOfPkgs)
	return downloadPackages(pkgs, f.baseURL)
}

func parsePkgResponse(data io.Reader, numberOfPkgs int) []Package {
	pkgs := []Package{}
	scanner := bufio.NewScanner(data)
	count := 0
	for scanner.Scan() {
		if numberOfPkgs > 0 && count >= numberOfPkgs {
			break
		}
		pkgline := scanner.Text()
		if strings.Contains(pkgline, "Package: ") {
			newPkg := Package{}
			newPkg.Name = strings.Split(pkgline, ": ")[1]
			for scanner.Scan() {
				line := scanner.Text()
				if strings.Contains(line, "Version: ") {
					newPkg.Version = strings.Split(line, ": ")[1]
				} else if strings.Contains(line, "MD5sum: ") {
					newPkg.MD5sum = strings.Split(line, ": ")[1]
					break
				}
			}
			pkgs = append(pkgs, newPkg)
			count += 1
		}
	}
	return pkgs
}
