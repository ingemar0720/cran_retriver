package fetch

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"strings"
	"sync"

	"github.com/pkg/errors"
)

type Package struct {
	Name            string
	Version         string
	MD5sum          string
	DatePublication string
	Title           string
	Description     string
	Author          Developer
	Maintainer      Developer
}

type Developer struct {
	Name  string
	Email string
}

func downloadPackages(pkgs []Package, baseURL string) []Package {
	result := []Package{}
	wg := sync.WaitGroup{}
	pkgChans := make(chan Package, 10)
	for _, p := range pkgs {
		wg.Add(1)
		go func(p Package) {
			if err := downloadPkgAsync(&p, baseURL); err == nil {
				pkgChans <- p
			}
		}(p)
	}
	go func() {
		for np := range pkgChans {
			result = append(result, np)
			wg.Done()
		}
	}()
	wg.Wait()
	fmt.Printf("found %d package\n", len(result))
	return result
}

func downloadPkgAsync(p *Package, baseURL string) error {
	client := http.DefaultClient
	request, err := http.NewRequest("GET", baseURL+p.Name+"_"+p.Version+".tar.gz", nil)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("compose request to download package %v fail, error %v", baseURL+p.Name+"_"+p.Version+".tar.gz", err))
	}
	request.Header.Add("Accept-Encoding", "gzip")
	resp, err := client.Do(request)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("download package %v fail, error %v\n", baseURL+p.Name+"_"+p.Version+".tar.gz", err))
	}
	if resp.StatusCode != 200 {
		return errors.Wrap(err, fmt.Sprintf("download package %v fail, error %v, statusCode %v\n",
			baseURL+p.Name+"_"+p.Version+".tar.gz", err, resp.StatusCode))
	}
	defer resp.Body.Close()
	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("create gzip reader fail, err: %v\n", err))
	}
	defer reader.Close()
	if err := parseCompressedFile(reader, p); err != nil {
		return errors.Wrap(err, fmt.Sprintf("parseCompressedFile fail, error %v\n", err))
	}
	return nil
}

func parseCompressedFile(reader *gzip.Reader, p *Package) error {
	tarReader := tar.NewReader(reader)

	for {
		header, err := tarReader.Next()
		if err == io.EOF {
			break
		}
		if err != nil {
			return fmt.Errorf("parse tar file fail, error: %v", err)
		}
		switch header.Typeflag {
		case tar.TypeReg:
			if strings.Contains(header.Name, "DESCRIPTION") {
				parseDescription(tarReader, p)
			}
		default:
			continue
		}
	}
	return nil
}

func parseDescription(reader io.Reader, p *Package) {
	scanner := bufio.NewScanner(reader)
	for scanner.Scan() {
		switch line := scanner.Text(); {
		case strings.Contains(line, "Package: "):
			p.Name = strings.TrimPrefix(line, "Package: ")
		case strings.Contains(line, "Version: "):
			p.Version = strings.TrimPrefix(line, "Version: ")
		case strings.Contains(line, "Date/Publication: "):
			p.DatePublication = strings.TrimPrefix(line, "Date/Publication: ")
		case strings.Contains(line, "Title: "):
			p.Title = strings.TrimPrefix(line, "Title: ")
		case strings.Contains(line, "Description: "):
			p.Description = strings.TrimPrefix(line, "Description: ")
		case strings.Contains(line, "Author: "):
			p.Author = parseDeveloper(line, "Author: ")
		case strings.Contains(line, "Maintainer: "):
			p.Maintainer = parseDeveloper(line, "Maintainer: ")
		default:
			continue
		}
	}
}

func parseDeveloper(str string, tag string) Developer {
	developerStr := strings.TrimPrefix(str, tag)
	developer := Developer{}
	u, err := mail.ParseAddress(developerStr)
	if err == nil {
		developer.Name = u.Name
		developer.Email = u.Address
	} else {
		developer.Name = developerStr
	}
	return developer
}
