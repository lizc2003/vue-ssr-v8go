package util

import (
	"bytes"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/lizc2003/vue-ssr-v8go/server/common/tlog"
	"io"
	"net"
	"net/http"
	"strings"
	"time"
)

func NewHttpClient(skipSSLVerify bool) *http.Client {
	return &http.Client{
		Timeout: 35 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:          200,
			MaxIdleConnsPerHost:   100,
			IdleConnTimeout:       60 * time.Second,
			DisableCompression:    true,
			ResponseHeaderTimeout: 30 * time.Second,
			DialContext: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).DialContext,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipSSLVerify},
		},
	}
}

func NewLongHttpClient(skipSSLVerify bool) *http.Client {
	return &http.Client{
		Timeout: 60 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:          200,
			MaxIdleConnsPerHost:   100,
			IdleConnTimeout:       60 * time.Second,
			DisableCompression:    true,
			ResponseHeaderTimeout: 60 * time.Second,
			DialContext: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).DialContext,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipSSLVerify},
		},
	}
}

func NewHttpClientWithShortTimeout(skipSSLVerify bool) *http.Client {
	return &http.Client{
		Timeout: 1200 * time.Millisecond,
		Transport: &http.Transport{
			MaxIdleConns:          200,
			MaxIdleConnsPerHost:   100,
			IdleConnTimeout:       60 * time.Second,
			DisableCompression:    true,
			ResponseHeaderTimeout: 1000 * time.Millisecond,
			DialContext: (&net.Dialer{
				Timeout: 300 * time.Millisecond,
			}).DialContext,
			TLSClientConfig: &tls.Config{InsecureSkipVerify: skipSSLVerify},
		},
	}
}

func NewHttpClientDisableKeepAlive(skipSSLVerify bool) *http.Client {
	return &http.Client{
		Timeout: 35 * time.Second,
		Transport: &http.Transport{
			MaxIdleConns:          200,
			MaxIdleConnsPerHost:   100,
			IdleConnTimeout:       60 * time.Second,
			DisableCompression:    true,
			ResponseHeaderTimeout: 30 * time.Second,
			DialContext: (&net.Dialer{
				Timeout: 10 * time.Second,
			}).DialContext,
			DisableKeepAlives: true,
			TLSClientConfig:   &tls.Config{InsecureSkipVerify: skipSSLVerify},
		},
	}
}

func HttpPost(client *http.Client, url string, headers map[string]string, params interface{}, ret interface{}) error {
	if client == nil {
		client = NewHttpClient(false)
	}

	var reqBody io.Reader
	switch params.(type) {
	case []byte:
		b := params.([]byte)
		reqBody = bytes.NewReader(b)
	default:
		reqJSON, _ := json.Marshal(params)
		reqBody = bytes.NewReader(reqJSON)
	}

	// buf := new(strings.Builder)
	// n, err := io.Copy(buf, reqBody)
	// tlog.Debugf("reqBody: %s n: %d err: %v", buf.String(), n, err)

	httpReq, err := http.NewRequest("POST", url, reqBody)
	if err != nil {
		tlog.Error(err, url)
		return err
	}

	httpReq.Header.Set("Content-Type", "application/json;charset=UTF-8")
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}
	//发送req
	resp, err := client.Do(httpReq)
	if err != nil {
		tlog.Error(err, url)
		return err
	}

	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			tlog.Error(err, url)
			return err
		}
		if IsHttpStatusSuccess(resp.StatusCode) {
			switch ret.(type) {
			case *string:
				*(ret.(*string)) = UnsafeBytes2Str(body)
			default:
				if err = json.Unmarshal(body, ret); err != nil {
					tlog.Infof("json.Unmarshal fail, body: %s", UnsafeBytes2Str(body))
					tlog.Error(err, url)
					return err
				}
			}
			return nil
		} else {
			tlog.Infof("status fail, body: %s", UnsafeBytes2Str(body))
			switch ret.(type) {
			case *string:
				*(ret.(*string)) = UnsafeBytes2Str(body)
			}
			errMsg := fmt.Sprintf("Http status error: %d, %s", resp.StatusCode, url)
			err = errors.New(errMsg)
			tlog.Error(err)
			return err
		}
	}
	err = fmt.Errorf("Http no body: %s", url)
	tlog.Error(err)
	return err
}

func HttpPostWithForm(client *http.Client, url string, headers map[string]string, params map[string]string, ret interface{}) error {
	if client == nil {
		client = NewHttpClient(false)
	}

	var hreq http.Request
	hreq.ParseForm()
	for k, v := range params {
		hreq.Form.Add(k, v)
	}
	httpReq, err := http.NewRequest("POST", url, strings.NewReader(hreq.Form.Encode()))
	if err != nil {
		tlog.Error(err, url)
		return err
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	for k, v := range headers {
		httpReq.Header.Set(k, v)
	}
	resp, err := client.Do(httpReq)
	if err != nil {
		tlog.Error(err, url)
		return err
	}

	if resp != nil && resp.Body != nil {
		defer resp.Body.Close()
		body, err := io.ReadAll(resp.Body)
		if err != nil {
			tlog.Error(err, url)
			return err
		}

		if IsHttpStatusSuccess(resp.StatusCode) {
			switch ret.(type) {
			case *string:
				*(ret.(*string)) = UnsafeBytes2Str(body)
			case *[]byte:
				*(ret.(*[]byte)) = body
			default:
				if err := json.Unmarshal(body, ret); err != nil {
					tlog.Infof("json.Unmarshal fail, body: %s", UnsafeBytes2Str(body))
					tlog.Error(err, url)
					return err
				}
			}
			return nil
		} else {
			tlog.Infof("status fail, body: %s", UnsafeBytes2Str(body))
			switch ret.(type) {
			case *string:
				*(ret.(*string)) = UnsafeBytes2Str(body)
			case *[]byte:
				*(ret.(*[]byte)) = body
			}
			errMsg := fmt.Sprintf("Http status error: %d, %s", resp.StatusCode, url)
			err = errors.New(errMsg)
			tlog.Error(err)
			return err
		}
	}
	err = fmt.Errorf("Http no body: %s", url)
	tlog.Error(err)
	return err
}

func IsHttpStatusSuccess(code int) bool {
	return code >= http.StatusOK && code <= http.StatusIMUsed
}
