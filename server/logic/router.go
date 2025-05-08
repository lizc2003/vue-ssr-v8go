package logic

import (
	"net/http"
	"strings"

	"github.com/lizc2003/vue-ssr-v8go/server/common/util"
)

func GetHttpHandler(env string, publicDir string, assetsPrefix string) http.Handler {
	fileServer := http.FileServer(http.Dir(publicDir))

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.HasPrefix(request.URL.Path, assetsPrefix) {
			fileServer.ServeHTTP(writer, request)
		} else {
			proxy := GetReverseProxy(request.URL.Path)
			if proxy != nil {
				proxy.ServeHTTP(writer, request)
			} else {
				isExists, _ := util.FileExists(publicDir + request.URL.Path)
				if isExists {
					fileServer.ServeHTTP(writer, request)
				} else {
					HandleSsrRequest(writer, request)
				}
			}
		}
	})
}
