package fetch

import (
	"archive/tar"
	"compress/gzip"
	"fmt"
	"io"
	"os"
	"testing"
)

func TestDecompressFile(t *testing.T) {
	reader, err := os.Open("/Users/ingemarsc/Downloads/control.tar.gz")
	if err != nil {
		t.Fatal(err)
	}
	defer reader.Close()

	archive, err := gzip.NewReader(reader)
	if err != nil {
		t.Fatal(err)
	}
	defer archive.Close()
	tarReader := tar.NewReader(archive)
	i := 0
	for {
		header, err := tarReader.Next()
		fmt.Println(header, err)
		if err == io.EOF {
			break
		}

		if err != nil {
			fmt.Println("parse fail", err)
			os.Exit(1)
		}

		name := header.Name

		switch header.Typeflag {
		case tar.TypeDir:
			continue
		case tar.TypeReg:
			fmt.Println("(", i, ")", "Name: ", name)

		default:
			fmt.Printf("%s : %c %s %s\n",
				"Yikes! Unable to figure out type",
				header.Typeflag,
				"in file",
				name,
			)
		}
		t.Errorf("")
		i++
	}
}
