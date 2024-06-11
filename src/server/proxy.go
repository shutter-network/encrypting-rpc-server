package server

import (
	"net/http/httputil"

	"github.com/shutter-network/rolling-shutter/rolling-shutter/medley/encodeable/url"
)

func NewReverseProxy(target *url.URL) *httputil.ReverseProxy {

	rewriteFunc := func(r *httputil.ProxyRequest) {
		Logger.Info().Msg("NewReverseProxy - Proxy Request" + r.In.Method)

		r.SetURL(target.URL)
	}
	return &httputil.ReverseProxy{Rewrite: rewriteFunc}
}
