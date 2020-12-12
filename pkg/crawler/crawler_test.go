package crawler

import (
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httptest"
	"testing"
)

var serverIndexResponse = []byte("pong")

func TestFetchProviders(t *testing.T) {
	ts := newTestServer()
	defer ts.Close()

	doc := requestPage(ts.URL + "/html")
	providers := fetchProviders(doc)
	if len(providers) != 51 {
		t.Fatalf("Should be 51 providers, found %d", len(providers))
	}
}

func newTestServer() *httptest.Server {
	htmlPage, _ := ioutil.ReadFile("testdata/duyaoss3.html")
	img, err := ioutil.ReadFile("testdata/sample_img.png")
	if err != nil {
		log.Fatal(err)
	}

	mux := http.NewServeMux()

	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		w.Write(serverIndexResponse)
	})

	mux.HandleFunc("/html", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.Write(htmlPage)
	})

	mux.HandleFunc("/img", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "image/png")
		w.Write(img)
	})

	return httptest.NewServer(mux)
}
