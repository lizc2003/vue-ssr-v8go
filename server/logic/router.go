package logic

import (
	"github.com/lizc2003/vue-ssr-v8go/server/common/util"
	"net/http"
)

func GetHttpHandler(env string, publicDir string) http.Handler {
	fileServer := http.FileServer(http.Dir(publicDir))

	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
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
	})
}
