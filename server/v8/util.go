package v8

import (
	"errors"
	"github.com/tommie/v8go"
	"net"
	"net/http"
	"time"
)

func newHttpClient() *http.Client {
	return &http.Client{
		Timeout: 8 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:          200,
			MaxIdleConnsPerHost:   100,
			IdleConnTimeout:       15 * time.Second,
			DisableCompression:    true,
			ResponseHeaderTimeout: 6 * time.Second,
			DialContext: (&net.Dialer{
				Timeout: 2 * time.Second,
			}).DialContext,
		},
	}
}

func ToJsError(err error) error {
	var jsErr *v8go.JSError
	if errors.As(err, &jsErr) {
		err = errors.New(jsErr.StackTrace)
	}
	return err
}
