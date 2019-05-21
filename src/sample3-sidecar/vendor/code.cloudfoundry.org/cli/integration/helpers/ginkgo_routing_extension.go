package helpers

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

type resp struct {
	status int
	body   []byte
}

var responses map[string]resp
var seenRoutes map[string]bool

// AddHandler adds a mock handler to the server making a request specified by "method" to the
// endpoint specified by "pathAndQuery", returning a response with status code "status" and
// response body "body".
func AddHandler(ser *ghttp.Server, method string, pathAndQuery string, status int, body []byte) {
	u, err := url.Parse(pathAndQuery)
	if err != nil {
		panic(err)
	}
	if len(responses) == 0 {
		responses = make(map[string]resp)
		seenRoutes = make(map[string]bool)
	}

	responses[key(method, u)] = resp{status, body}

	if !seenRoutes[key(method, u)] {
		ser.RouteToHandler(method, u.Path, func(w http.ResponseWriter, r *http.Request) {
			res, ok := responses[key(r.Method, r.URL)]
			if !ok {
				Expect(errors.New("Unexpected request: " + key(r.Method, r.URL))).ToNot(HaveOccurred())
			}
			w.WriteHeader(res.status)
			_, err := w.Write(res.body)
			Expect(err).ToNot(HaveOccurred())
		})
		seenRoutes[key(method, u)] = true
	}
}

func key(method string, url fmt.Stringer) string {
	return strings.ToLower(method + url.String())
}
