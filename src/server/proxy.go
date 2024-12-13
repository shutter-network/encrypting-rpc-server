package server

import (
	"net/http"
	"net/http/httputil"
	"net/url"
)

type StripCORSHeaders struct {
	Transport http.RoundTripper
}

func (s *StripCORSHeaders) RoundTrip(req *http.Request) (*http.Response, error) {
	resp, err := s.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}

	resp.Header.Del("Access-Control-Allow-Origin")
	resp.Header.Del("Access-Control-Allow-Methods")
	resp.Header.Del("Access-Control-Allow-Headers")

	return resp, nil
}

func NewReverseProxy(target *url.URL) *httputil.ReverseProxy {
	transport := &http.Transport{}

	stripTransport := &StripCORSHeaders{
		Transport: transport,
	}

	rewriteFunc := func(r *httputil.ProxyRequest) {
		r.SetURL(target)
	}

	proxy := &httputil.ReverseProxy{
		Rewrite:   rewriteFunc,
		Transport: stripTransport,
	}

	return proxy
}