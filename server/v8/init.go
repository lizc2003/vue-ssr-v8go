package v8

import (
	"github.com/lizc2003/v8go"
	"github.com/lizc2003/vue-ssr-v8go/server/common/defs"
	"github.com/lizc2003/vue-ssr-v8go/server/common/util"
	"os"
	"strings"
)

var gInitJs string
var gInitJsCache *v8go.CompilerCachedData
var gServerFileName string
var gServerJs string
var gServerJsCache *v8go.CompilerCachedData

const (
	gInitJsName   = "init.js"
	gServerJsName = "server.js"
)

func initVm(env string, serverDir string, useStrict bool) error {
	flags := []string{"--max-heap-size=8192"}
	if useStrict {
		flags = append(flags, "--use_strict")
	} else {
		flags = append(flags, "--nouse_strict")
	}
	v8go.SetFlags(flags...)

	bDev := false
	nodeEnv := "production"
	if env == defs.EnvDev {
		nodeEnv = "development"
		bDev = true
	}

	gInitJs = strings.Replace(initJsContent, "$NODE_ENV", nodeEnv, 1)
	gInitJs += xmlHttpRequestJsContent

	var err error
	gInitJsCache, err = CompileJsScript(gInitJs, gInitJsName)
	if err != nil {
		return err
	}

	if serverDir != "" {
		gServerFileName = serverDir + "/" + gServerJsName
		if !bDev {
			content, err := os.ReadFile(gServerFileName)
			if err != nil {
				return err
			}
			gServerJs = util.UnsafeBytes2Str(content)
			gServerJsCache, err = CompileJsScript(gServerJs, gServerJsName)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

const initJsContent = `
globalThis.process = { env: { NODE_ENV: '$NODE_ENV' }};
globalThis.v8goJs = {};
globalThis.dumpObject = (function() {
	const maxDepth = 20;

	function _dumpObject(obj, depth, seen) {
		if (depth > maxDepth) {
			return '';
		}

		if (typeof obj !== 'object' || obj === null) {
			return JSON.stringify(obj);
		}

		let dump = '';
		let indent = '';
		if (depth == 0) {
			let objectString = obj.toString();
			if (objectString.length > 0) {
				dump += objectString + '\n' + '-'.repeat(objectString.length) + '\n';
			} else {
				if (Array.isArray(obj) && obj.length == 0) {
					return indent + '[]';
				}
			}
		} else {
			indent = ' '.repeat(depth * 2);
		}

		if (seen.has(obj)) {
			return indent + '[Circular Reference]';
		}
		seen.set(obj, true);

		let isFirst = true;
		for (let key in obj) {
			const item = obj[key];
			const type = typeof item;
			let typeInfo = type;

			if (isFirst) {
				isFirst = false;
			} else {
				dump += ',\n';
			}

			dump += indent + key + ' => (' + typeInfo + ')';

			if (item === null) {
				dump += ' null';
			} else if (type ==='object' || type === 'function') {
				const isArray = Array.isArray(item);
				let isFunction = false;
				if (isArray) {
					dump += ' [';
				} else if (type === 'function') {
					dump += ' {';
					isFunction = true;
				} else {
					// we're assuming toString() yields sane values
					let itemString = item.toString();
					if (itemString !== '[object Object]') {
						dump += ' ' + itemString;
					}
					dump += ' {';
				}

				if (!isFunction) {
					let objDump = '';
					try {
						objDump = _dumpObject(item, depth + 1, seen);
					} catch (e) {
						// object cannot be dumped, probably a non-null pointer
					}
					if (objDump !== '') {
						dump += '\n' + objDump + '\n' + indent;
					}
				}

				if (isArray) {
					dump += ']';
				} else if (isFunction) {
					dump += '}';
				} else {
					dump += '}';
				}
			} else {
				dump += ' ' + item;
			}
		}
		return dump;
	}

	return function(obj) {
		const seen = new WeakMap()
		return _dumpObject(obj, 0, seen);
	}
})();
`
