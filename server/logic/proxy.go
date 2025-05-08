package logic

import (
	"errors"
	"net/http/httputil"
	"net/url"
	"path"
	"sort"
	"strings"
)

// /////////////////////////////////////////
type LocationReverseProxy struct {
	path  string
	proxy *httputil.ReverseProxy
}

var gReverseProxies []*LocationReverseProxy

func initReverseProxy(locs []ProxyLocation) error {
	if len(locs) == 0 {
		return nil
	}
	sort.Sort(LocationList(locs))

	var proxies []*LocationReverseProxy
	for _, loc := range locs {
		u, err := url.Parse(loc.Target)
		if err != nil {
			return err
		}
		var w1, w2 string
		sz := len(loc.Rewrite)
		if sz > 0 && sz != 2 {
			return errors.New("invalid rewrite format")
		}
		if sz == 2 {
			w1 = loc.Rewrite[0]
			w2 = loc.Rewrite[1]
		}
		proxies = append(proxies, &LocationReverseProxy{
			path:  loc.Path,
			proxy: makeReverseProxy(u, w1, w2),
		})
	}

	gReverseProxies = proxies
	return nil
}

func makeReverseProxy(target *url.URL, pattern, replaceVal string) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Rewrite: func(r *httputil.ProxyRequest) {
			inPath := r.In.URL.Path
			if pattern != "" {
				inPath = strings.Replace(inPath, pattern, replaceVal, 1)
			}
			r.Out.URL.Scheme = target.Scheme
			r.Out.URL.Host = target.Host
			r.Out.URL.Path = path.Join(target.Path, inPath)
			r.Out.Host = target.Host
		},
	}
}

func getReverseProxy(urlPath string) *httputil.ReverseProxy {
	for _, proxy := range gReverseProxies {
		if strings.HasPrefix(urlPath, proxy.path) {
			return proxy.proxy
		}
	}
	return nil
}

type ProxyLocation struct {
	Path    string   `toml:"path"`
	Target  string   `toml:"target"`
	Rewrite []string `toml:"rewrite"`
}

type ProxyConfig struct {
	Locations []ProxyLocation `toml:"location"`
}

type LocationList []ProxyLocation

func (a LocationList) Len() int {
	return len(a)
}

func (a LocationList) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}

func (a LocationList) Less(i, j int) bool {
	return len(a[i].Path) > len(a[j].Path)
}
