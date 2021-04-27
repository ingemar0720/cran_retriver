package fetch

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"database/sql"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"strings"
	"sync"
)

type Package struct {
	Name            string
	Version         string
	MD5sum          string
	DatePublication sql.NullString
	Title           sql.NullString
	Description     sql.NullString
	Author          Developer
	Maintainer      Developer
}

type Developer struct {
	Name  string
	Email sql.NullString
}

func downloadPackages(pkgs []Package, baseURL string) []Package {
	result := []Package{}
	wg := sync.WaitGroup{}
	pkgChans := make(chan Package, 10)
	for _, p := range pkgs {
		wg.Add(1)
		go func(p Package) {
			downloadPkgAsync(&p, baseURL)
			pkgChans <- p
			wg.Done()
		}(p)
	}
	go func() {
		for np := range pkgChans {
			result = append(result, np)
		}
	}()
	wg.Wait()
	return result
}

func downloadPkgAsync(p *Package, baseURL string) {
	client := http.DefaultClient
	request, err := http.NewRequest("GET", baseURL+p.Name+"_"+p.Version+".tar.gz", nil)
	if err != nil {
		fmt.Printf("compose request to download package %v fail, error %v", baseURL+p.Name+"_"+p.Version+".tar.gz", err)
		return
	}
	request.Header.Add("Accept-Encoding", "gzip")
	resp, err := client.Do(request)
	if err != nil || resp.StatusCode != 200 {
		fmt.Printf("download package %v fail, error %v, statusCode %v\n",
			baseURL+p.Name+"_"+p.Version+".tar.gz", err, resp.StatusCode)
		return
	}
	defer resp.Body.Close()
	reader, err := gzip.NewReader(resp.Body)
	if err != nil {
		fmt.Printf("create gzip reader fail, err: %v\n", err)
		return
	}
	defer reader.Close()
	if err := parseCompressedFile(reader, p); err != nil {
		fmt.Printf("parseCompressedFile fail, error %v\n", err)
		return
	}
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
			p.DatePublication = strToNullStr(strings.TrimPrefix(line, "Date/Publication: "))
		case strings.Contains(line, "Title: "):
			p.Title = strToNullStr(strings.TrimPrefix(line, "Title: "))
		case strings.Contains(line, "Description: "):
			p.Description = strToNullStr(strings.TrimPrefix(line, "Description: "))
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
		developer.Email = strToNullStr(u.Address)
	} else {
		developer.Name = developerStr
	}
	return developer
}

func strToNullStr(str string) sql.NullString {
	return sql.NullString{
		String: str,
		Valid:  true,
	}
}
