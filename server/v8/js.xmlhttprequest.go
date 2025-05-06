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

package v8

const xmlHttpRequestJsContent = `
v8goJs.xhrMgr = (function() {
	const xhrMap = new Map();
	const statusCodes = {
		100:'Continue',101:'Switching Protocols',102:'Processing',200:'OK',201:'Created',202:'Accepted',203:'Non-Authoritative Information',204:'No Content',205:'Reset Content',206:'Partial Content',207:'Multi-Status',208:'Already Reported',226:'IM Used',300:'Multiple Choices',301:'Moved Permanently',302:'Found',303:'See Other',304:'Not Modified',305:'Use Proxy',306:'Switch Proxy',307:'Temporary Redirect',308:'Permanent Redirect',400:'Bad Request',401:'Unauthorized',402:'Payment Required',403:'Forbidden',404:'Not Found',405:'Method Not Allowed',406:'Not Acceptable',407:'Proxy Authentication Required',408:'Request Timeout',409:'Conflict',410:'Gone',411:'Length Required',412:'Precondition Failed',413:'Request Entity Too Large',414:'Request-URI Too Long',415:'Unsupported Media Type',416:'Requested Range Not Satisfiable',417:'Expectation Failed',418:'I\'m a teapot',419:'Authentication Timeout',420:'Method Failure',420:'Enhance Your Calm',422:'Unprocessable Entity',423:'Locked',424:'Failed Dependency',426:'Upgrade Required',428:'Precondition Required',429:'Too Many Requests',431:'Request Header Fields Too Large',440:'Login Timeout',444:'No Response',449:'Retry With',450:'Blocked by Windows Parental Controls',451:'Unavailable For Legal Reasons',451:'Redirect',494:'Request Header Too Large',495:'Cert Error',496:'No Cert',497:'HTTP to HTTPS',498:'Token expired/invalid',499:'Client Closed Request',499:'Token required',500:'Internal Server Error',501:'Not Implemented',502:'Bad Gateway',503:'Service Unavailable',504:'Gateway Timeout',505:'HTTP Version Not Supported',506:'Variant Also Negotiates',507:'Insufficient Storage',508:'Loop Detected',509:'Bandwidth Limit Exceeded',510:'Not Extended',511:'Network Authentication Required',520:'Origin Error',521:'Web server is down',522:'Connection timed out',523:'Proxy Declined Request',524:'A timeout occurred',598:'Network read timeout error',599:'Network connect timeout error'
	};

	return {
		getStatusText: function (status) {
			if (statusCodes.hasOwnProperty(status)) {
				return statusCodes[status];
			} else {
				return status.toString();
			}
		},

		addObject: function (xhrId, obj) {
			xhrMap.set(xhrId, obj);
		},

		sendEvent: function (msg) {
			let evt = msg.event;
			let obj = xhrMap.get(msg.xhr_id);
			if (obj === undefined) {
				return;
			}
		
			if (evt === "onfinish") {
				xhrMap.delete(msg.xhr_id);
			} else if (evt === "onerror") {
				obj._onErrorCallback(msg.error)
			} else if (evt === "onstart") {
				obj._onStartCallback()
			} else if (evt === "onheader") {
				obj._onHeaderCallback(msg.status, msg.headers)
			} else if (evt === "onend") {
				obj._onEndCallback(msg.response)
			}
		}
	};
})();

globalThis.XMLHttpRequest = function() {
	let method, url,
		xhr, callEventListeners,
		listeners = ['readystatechange', 'abort', 'error', 'loadend', 'progress', 'load'],
		privateListeners = {},
		responseHeaders = {},
		headers = {},
		hasSent = false,
		thisXhrId = 0;

	callEventListeners = function(listeners, evt) {
		let i;
		let listenerFound = false;
		if (typeof listeners === 'string') {
			listeners = [listeners];
		}
		listeners.forEach(function(e) {
			if (typeof xhr['on' + e] === 'function') {
				listenerFound = true;
				xhr['on' + e].call(xhr, evt);
			}
			if (privateListeners.hasOwnProperty(e)) {
				for (i = 0; i < privateListeners[e].length; i++) {
					if (privateListeners[e][i] !== undefined) {
						listenerFound = true;
						privateListeners[e][i].call(xhr, evt);
					}
				}
			}
		});
		return listenerFound;
	};

	xhr = {
		UNSENT:		   0,
		OPENED:		   1,
		HEADERS_RECEIVED: 2,
		LOADING:		  3,
		DONE:			 4,

		status: 0,
		statusText: '',
		responseText: '',
		response: '',
		readyState: 0,
		timeout: 8000,

		open: function(_method, _url, _async, _user, _password) {
			let async = (typeof _async !== "boolean" ? true : _async)
			if (_method === undefined || _url === undefined) {
				throw TypeError('Failed to execute \'open\' on \'XMLHttpRequest\': 2 arguments required, but only '+(_method || _url)+' present.');
			}
			if (!async) {
				throw TypeError('Failed to execute \'open\' on \'XMLHttpRequest\': Synchronous request is not supported.');
			}
			method = _method;
			url = _url;
			xhr.readyState = xhr.OPENED;
			callEventListeners('readystatechange');
		},

		send: function(post) {
			if (xhr.readyState !== xhr.OPENED) {
				throw {
					name: 'InvalidStateError',
					message: 'Failed to execute \'send\' on \'XMLHttpRequest\': The object\'s state must be OPENED.'
				}
			}
			if (hasSent) {
				throw {
					name: 'InvalidStateError',
					message: 'Failed to execute \'send\' on \'XMLHttpRequest\': The object has been sent.'
				}
			}
			hasSent = true;

			let options = {
				cmd: "open",
				url: url,
				method: method,
				headers: headers,
				timeout: xhr.timeout
			};
 
			if (post) {
				let ptype = typeof(post);
				if (ptype == 'object') {
					try {
						let data = ""
						for (let key in post) {
							if (Object.prototype.hasOwnProperty.call(post, key)) {
								if (data.length > 0) {
									data += "&";
								}
								data += key.toString() + '=' + encodeURIComponent(post[key]);
							}
						}
						post = data;
					} catch(e) {
						post = "";
					}
				} else if (ptype !== "string") {
					post = "";
				}
				if (post.length > 0) {
					//console.log("post data:" + post)
					options.headers['Content-Length'] = post.length.toString();
					options.post = post
				}
			}
 
			thisXhrId = parseInt(v8goGo.handleXhrCmd(JSON.stringify(options)));
			if (thisXhrId > 0) {
				v8goJs.xhrMgr.addObject(thisXhrId, xhr)
			} else {
				throw TypeError('Failed to execute \'send\' on \'XMLHttpRequest\'');
			}
		},

		abort: function() {
			let tmpId = thisXhrId;
			if (tmpId > 0) {
				thisXhrId = 0;
				callEventListeners(['abort', 'loadend']);

				let options = {
					cmd: "abort",
					xhr_id: tmpId
				};
				v8goGo.handleXhrCmd(JSON.stringify(options));
			}
		},

		setRequestHeader: function(header, value) {
			if (header === undefined || value === undefined) {
				console.error(' Failed to execute \'setRequestHeader\' on \'XMLHttpRequest\': 2 arguments required, but only ' + (headers || value) + ' present.');
				return
			}
			headers[header] = value;
		},

		addEventListener: function(type, fn) {
			if (listeners.indexOf(type) !== -1) {
				privateListeners[type].push(fn);
			}
		},

		removeEventListener: function(type, fn) {
			let index;
			if (privateListeners[type] === undefined) {
				return;
			}
			while ((index = privateListeners[type].indexOf(fn)) !== -1) {
				privateListeners[type].splice(index, 1);
			}
		},

		getAllResponseHeaders: function() {
			let res = '';
			for (let key in responseHeaders) {
				res += key + ': ' + responseHeaders[key] + '\r\n';
			}
			if (res.length >= 2) {
				res = res.slice(0, -2);
			}
			return res;
		},

		getResponseHeader: function(key) {
			if (responseHeaders.hasOwnProperty(key)) {
				return responseHeaders[key];
			}
			return null;
		},

		_onStartCallback: function() {
			if (thisXhrId > 0) {
				callEventListeners('loadstart');
			}
		},

		_onHeaderCallback: function(status, headers) {
			if (thisXhrId > 0) {
				let total = 0;
				let contentLength = parseInt(headers['Content-Length']);
				if (contentLength > 0) {
					total = contentLength;
				}

				xhr.status = status;
				xhr.statusText = v8goJs.xhrMgr.getStatusText(status);
				responseHeaders = headers;
				xhr.readyState = xhr.HEADERS_RECEIVED;
				callEventListeners('readystatechange');
				xhr.readyState = xhr.LOADING;
				callEventListeners('readystatechange');
				
				if (total > 0) {
					callEventListeners('progress', {lengthComputable:true, loaded:0, total:total});
				} else {
					callEventListeners('progress', {lengthComputable:false, loaded:0, total:0});
				}
			}
		},

		_onErrorCallback: function(err) {
			if (thisXhrId > 0) {
				thisXhrId = 0;
				if (callEventListeners('error') === false) {
					console.error(err);
				}
				callEventListeners('loadend');
			}
		},

		_onEndCallback: function(response) {
			if (thisXhrId > 0) {
				thisXhrId = 0;
				xhr.readyState = xhr.DONE;
				xhr.responseText = response;
				xhr.response = response;
				callEventListeners(['readystatechange', 'load', 'loadend']);
			}
		}
	};

	for (let i = 0; i < listeners.length; i++) {
		xhr['on' + listeners[i]] = null;
		privateListeners[listeners[i]] = [];
	}

	return xhr;
};
`
