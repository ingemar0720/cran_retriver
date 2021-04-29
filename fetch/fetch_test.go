package fetch

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
)

func TestFetchService_FetchPkgList(t *testing.T) {
	var hasError bool
	mockServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if hasError == true {
			w.WriteHeader(http.StatusInternalServerError)
		}
		http.ServeFile(w, r, r.URL.Path[1:])
	}))
	tests := []struct {
		name              string
		server            *httptest.Server
		givenNumberOfPkgs int
		givenServerError  bool
		want              []Package
	}{
		{
			name:              "happy path",
			server:            mockServer,
			givenNumberOfPkgs: 2,
			want: []Package{
				{
					Name:            "A3",
					Version:         "1.0.0",
					MD5sum:          "027ebdd8affce8f0effaecfcd5f5ade2",
					DatePublication: "2015-08-16 23:05:52",
					Title:           "Accurate, Adaptable, and Accessible Error Metrics for Predictive",
					Description:     "Supplies tools for tabulating and analyzing the results of predictive models. The methods employed are applicable to virtually any predictive model and make comparisons between different methodologies straightforward.",
					Author:          Developer{Name: "Scott Fortmann-Roe"},
					Maintainer: Developer{
						Name:  "Scott Fortmann-Roe",
						Email: "scottfr@berkeley.edu",
					},
				},
				{
					Name:            "aaSEA",
					Version:         "1.1.0",
					MD5sum:          "0f9aaefc1f1cf18b6167f85dab3180d8",
					DatePublication: "2019-11-09 16:20:02 UTC",
					Title:           "Amino Acid Substitution Effect Analyser",
					Description:     "Given a protein multiple sequence alignment, it is daunting task to assess the effects of substitutions along sequence length. 'aaSEA' package is intended to help researchers to rapidly analyse property changes caused by single, multiple and correlated amino acid substitutions in proteins. Methods for identification of co-evolving positions from multiple sequence alignment are as described in :  Pelé et al., (2017) <doi:10.4172/2379-1764.1000250>.",
					Author:          Developer{Name: "Raja Sekhara Reddy D.M"},
					Maintainer: Developer{
						Name:  "Raja Sekhara Reddy D.M",
						Email: "raja.duvvuru@gmail.com",
					},
				},
			},
		},
		{
			name:              "server error and fail to download",
			server:            mockServer,
			givenNumberOfPkgs: 2,
			givenServerError:  true,
			want:              []Package{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.givenServerError == true {
				hasError = true
			}
			f := &FetchService{
				baseURL:      mockServer.URL + "/fixture/",
				numberOfPkgs: tt.givenNumberOfPkgs,
			}
			if got := f.FetchPkgList(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("FetchService.FetchPkgList() = %v, want %v", got, tt.want)
			}
			hasError = false
		})
	}
}
