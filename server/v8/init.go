package v8

import (
	"fmt"
	"github.com/lizc2003/vue-ssr-v8go/server/common/defs"
)

var gGlobalJs string

func initVm(env string) error {
	nodeEnv := "production"
	if env == defs.EnvDev {
		nodeEnv = "development"
	}

	gGlobalJs = fmt.Sprintf(
		`
globalThis.v8goJs = {};
globalThis.process = { env: { NODE_ENV: "%s" }};
`,
		nodeEnv)
	gGlobalJs += xmlHttpRequestJsContent

	return nil
}
