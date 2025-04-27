package logic

import (
	"net/http"
	"strings"

	"github.com/lizc2003/vue-ssr-v8go/server/common/util"
)

func GetHttpHandler(env string, assetsPrefix string) http.Handler {
	e := util.NewGinEngine(env)

	publicDir := getDistPath() + clientPath
	e.StaticFile("/robots.txt", publicDir+"/robots.txt")
	e.StaticFile("/favicon.ico", publicDir+"/favicon.ico")
	e.NoRoute(HandleSsrRequest)

	assetsServer := getAssetsServer(publicDir, assetsPrefix)
	return http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
		if strings.HasPrefix(request.URL.Path, assetsPrefix) {
			assetsServer.ServeHTTP(writer, request)
		} else {
			e.ServeHTTP(writer, request)
		}
	})
}

func getAssetsServer(publicDir string, assetsPrefix string) http.Handler {
	assetsDir := publicDir + assetsPrefix
	assetsServer := http.FileServer(http.Dir(assetsDir))
	return http.StripPrefix(assetsPrefix, assetsServer)
}
