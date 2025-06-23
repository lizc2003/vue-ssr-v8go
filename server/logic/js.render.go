// Copyright 2020-present, lizc2003@gmail.com
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package logic

import (
	"strings"
)

type RenderResult struct {
	Html    string `json:"html"`
	Meta    string `json:"meta"`
	State   string `json:"state"`
	Modules string `json:"modules"`
}

const renderJsName = "render.js"

const renderJsContent = `
(function() {
	let ctx = $RENDER_CONTEXT;
	v8goRenderToString(ctx).then((html) => {
		let meta = '';
		let state = '';
		let modules = '';
		if (typeof ctx.htmlMeta === 'string') {
			meta = ctx.htmlMeta;
		}
		if (ctx.htmlState) {
			state = JSON.stringify(ctx.htmlState);
		}
		if (typeof ctx.htmlModules === 'string') {
			modules = ctx.htmlModules;
		}

		v8goGo.sendMessage(10, ctx.renderId, html, meta, state, modules);
		ctx = null;
	}).catch((err) => {
		v8goGo.sendMessage(11, ctx.renderId, err.stack, '', '', '');
		ctx = null;
	})
})()
`

var renderJsPart1 string
var renderJsPart2 string
var renderJsLength int

func init() {
	s := "$RENDER_CONTEXT"
	idx := strings.Index(renderJsContent, s)
	renderJsPart1 = renderJsContent[:idx]
	renderJsPart2 = renderJsContent[idx+len(s):]
	renderJsLength = len(renderJsPart1) + len(renderJsPart2)
}
