package v8

import (
	"errors"
	"github.com/lizc2003/v8go"
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

func CompileJsScript(code string, scriptName string) (*v8go.CompilerCachedData, error) {
	var ret *v8go.CompilerCachedData
	iso := v8go.NewIsolate()
	v8ctx := v8go.NewContext(iso)

	script, err := iso.CompileUnboundScript(code, scriptName, v8go.CompileOptions{})
	if err == nil {
		_, err = script.Run(v8ctx)
		if err == nil {
			ret = script.CreateCodeCache()
		}
	}

	v8ctx.Close()
	iso.Dispose()
	return ret, err
}
