package logic

import (
	"errors"
	"net"
	"net/http"
	"net/http/httputil"
	"net/url"
	"path"
	"sort"
	"strings"
)

type LocationReverseProxy struct {
	path  string
	proxy *httputil.ReverseProxy
}

var gReverseProxies []*LocationReverseProxy

func InitReverseProxy(locs []ProxyLocation) error {
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

func GetReverseProxy(urlPath string) *httputil.ReverseProxy {
	for _, proxy := range gReverseProxies {
		if strings.HasPrefix(urlPath, proxy.path) {
			return proxy.proxy
		}
	}
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

			if r.In.Host != "" {
				host, _, err := net.SplitHostPort(r.In.Host)
				if err == nil {
					r.Out.Header.Set("X-Forwarded-Host", host)
				}
			}
			proto := "http"
			if r.In.TLS != nil {
				proto = "https"
			}
			r.Out.Header.Set("X-Forwarded-Proto", proto)
		},

		ModifyResponse: func(resp *http.Response) error {
			fwdHost := resp.Request.Header.Get("X-Forwarded-Host")
			if fwdHost == "" {
				return nil
			}
			proto := resp.Request.Header.Get("X-Forwarded-Proto")
			isHttp := (proto == "http")

			cookies := resp.Header.Values("Set-Cookie")
			if len(cookies) == 0 {
				return nil
			}
			newCookies := make([]string, 0, len(cookies))
			for _, sc := range cookies {
				parts := strings.Split(sc, ";")
				replaced := false
				sz := len(parts)
				for i := 0; i < sz; i++ {
					p := strings.ToLower(strings.TrimSpace(parts[i]))
					if strings.HasPrefix(p, "domain=") {
						parts[i] = "Domain=" + fwdHost
						replaced = true
					}
					if isHttp {
						if strings.HasPrefix(p, "secure") ||
							strings.HasPrefix(p, "samesite=") {
							parts[i] = ""
						}
					}
				}
				if replaced {
					idx := 0
					for i := 0; i < sz; i++ {
						if parts[i] == "" {
							continue
						}
						parts[idx] = strings.TrimSpace(parts[i])
						idx++
					}
					newCookies = append(newCookies, strings.Join(parts[:idx], "; "))
				} else {
					newCookies = append(newCookies, sc)
				}
			}
			resp.Header.Del("Set-Cookie")
			for _, nc := range newCookies {
				resp.Header.Add("Set-Cookie", nc)
			}
			return nil
		},
	}
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
