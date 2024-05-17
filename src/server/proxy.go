package server

import (
	"net/http/httputil"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/url"
)

func NewReverseProxy(target *url.URL) *httputil.ReverseProxy {
	rewriteFunc := func(r *httputil.ProxyRequest) {
		r.SetURL(target.URL)
	}
	return &httputil.ReverseProxy{Rewrite: rewriteFunc}
}
