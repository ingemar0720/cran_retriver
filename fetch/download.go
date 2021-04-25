package fetch

import (
	"archive/tar"
	"bufio"
	"compress/gzip"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type Package struct {
	name            string
	version         string
	md5sum          string
	datePublication string
	title           string
	description     string
	author          string
	maintainer      string
}

func downloadPackages(pkgs []Package, baseURL string) {
	for i, p := range pkgs {
		client := http.DefaultClient
		request, err := http.NewRequest("GET", baseURL+p.name+"_"+p.version+".tar.gz", nil)
		if err != nil {
			fmt.Printf("compose request to download package %v fail, error %v", baseURL+p.name+"_"+p.version+".tar.gz", err)
			continue
		}
		request.Header.Add("Accept-Encoding", "gzip")
		resp, err := client.Do(request)
		if err != nil || resp.StatusCode != 200 {
			fmt.Printf("download package %v fail, error %v, statusCode %v\n",
				baseURL+p.name+"_"+p.version+".tar.gz", err, resp.StatusCode)
			continue
		}
		defer resp.Body.Close()
		reader, err := gzip.NewReader(resp.Body)
		if err != nil {
			fmt.Printf("create gzip reader fail, err: %v\n", err)
			continue
		}
		defer reader.Close()
		if err := parseCompressedFile(reader, &p); err != nil {
			fmt.Printf("parseCompressedFile fail, error %v\n", err)
			continue
		}
		pkgs[i] = p
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
			p.name = strings.TrimPrefix(line, "Package: ")
		case strings.Contains(line, "Version: "):
			p.version = strings.TrimPrefix(line, "Version: ")
		case strings.Contains(line, "Date/Publication: "):
			p.datePublication = strings.TrimPrefix(line, "Date/Publication: ")
		case strings.Contains(line, "Title: "):
			p.title = strings.TrimPrefix(line, "Title: ")
		case strings.Contains(line, "Description: "):
			p.description = strings.TrimPrefix(line, "Description: ")
		case strings.Contains(line, "Author: "):
			p.author = strings.TrimPrefix(line, "Authors: ")
		case strings.Contains(line, "Maintainer: "):
			p.maintainer = strings.TrimPrefix(line, "Maintainers: ")
		}
	}
}
