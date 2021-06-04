package handlers

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"
)

type postData struct {
	key   string
	value string
}

var theTests = []struct {
	name               string
	url                string
	method             string
	params             []postData
	expectedStatusCode int
}{
	{"home", "/", "GET", []postData{}, http.StatusOK},
	{"about", "/about", "GET", []postData{}, http.StatusOK},
	{"contact", "/contact", "GET", []postData{}, http.StatusOK},
	{"gq", "/rooms/generals-quarters", "GET", []postData{}, http.StatusOK},
	{"ms", "/rooms/majors-suite", "GET", []postData{}, http.StatusOK},
	{"sa", "/search-availability", "GET", []postData{}, http.StatusOK},
	{"mr", "/make-reservation", "GET", []postData{}, http.StatusOK},
	{"psa", "/search-availability", "POST", []postData{
		{key: "start", value: "2021/06/03"},
		{key: "end", value: "2021/06/04"},
	}, http.StatusOK},
	{"psa-json", "/search-availability-json", "POST", []postData{
		{key: "start", value: "2021/06/03"},
		{key: "end", value: "2021/06/04"},
	}, http.StatusOK},
	{"pmr", "/make-reservation", "POST", []postData{
		{key: "first_name", value: "John"},
		{key: "last_name", value: "Doe"},
		{key: "email", value: "john.doe@email.com"},
		{key: "phone", value: "55 19 996539944"},
	}, http.StatusOK},
}

func TestHandlers(t *testing.T) {
	routes := getRoutes()
	ts := httptest.NewTLSServer(routes)
	defer ts.Close()

	for _, e := range theTests {
		if e.method == "GET" {
			res, err := ts.Client().Get(ts.URL + e.url)
			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}

			if res.StatusCode != e.expectedStatusCode {
				t.Errorf("For %s, expected %d but got %d", e.name, e.expectedStatusCode, res.StatusCode)
			}
		} else if e.method == "POST" {
			values := url.Values{}
			for _, x := range e.params {
				values.Add(x.key, x.value)
			}

			res, err := ts.Client().PostForm(ts.URL+e.url, values)
			if err != nil {
				t.Log(err)
				t.Fatal(err)
			}

			if res.StatusCode != e.expectedStatusCode {
				t.Errorf("For %s, expected %d but got %d", e.name, e.expectedStatusCode, res.StatusCode)
			}
		}
	}
}
