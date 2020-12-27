package crawler

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
)

var serverIndexResponse = []byte("pong")
var sampleProvider = Provider{
	Name:   "4.ssrcloud （CNIX 中转机场 高性价比）",
	ImgURL: "https://user-images.githubusercontent.com/34016863/100780954-58ac3280-3445-11eb-928d-a75c71dcfe7f.png#vwid=1359&amp;vhei=8700",
}

func TestFetchProviders(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	doc := RequestPage(ts.URL + "/html")
	providers := FetchProviders(doc)
	if len(providers) != 51 {
		t.Errorf("Found %d out of 51 providers.", len(providers))
	}

	subgroupCount := 0
	for _, provider := range providers {
		if provider.Subgroup != nil {
			subgroupCount++
		}
	}
	if subgroupCount != 7 {
		t.Errorf("Found %d out of 7 subgroups.", subgroupCount)
	}
}

func TestFetchImage(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	img := FetchImage(sampleProvider.ImgURL)

	width, height := img.Bounds().Max.X, img.Bounds().Max.Y
	if width != 1359 {
		t.Errorf("Width is %d and should be 1359", width)
	}
	if height != 8700 {
		t.Errorf("Height is %d and should be 8700", width)
	}

}

func newTestServer() *httptest.Server {
	htmlPage, _ := ioutil.ReadFile("testdata/duyaoss3.html")

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write(serverIndexResponse)
	})

	mux.HandleFunc("/html", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(htmlPage)
	})

	return httptest.NewServer(mux)
}
