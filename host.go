package httptest

import (
	"io"
	"net/http"
	"strings"
)

// ---------------------------------------------------------------------------

type hostsMgr struct {
	hostToPortals map[string]string
}

func (p *hostsMgr) initHostsMgr() {

	p.hostToPortals = make(map[string]string)
}

func (p *hostsMgr) hostOf(url string) (host string, url2 string, ok bool) {

	var istart int

	// http://host/xxx or https://host/xxx
	if strings.HasPrefix(url[4:], "://") {
		istart = 7
	} else if strings.HasPrefix(url[4:], "s://") {
		istart = 8
	} else {
		return
	}

	n := strings.IndexByte(url[istart:], '/')
	if n < 1 {
		return
	}

	host = url[istart:istart+n]
	portal, ok := p.hostToPortals[host]
	if ok {
		url2 = url[:istart] + portal + url[istart+n:]
	}
	return
}

func (p *hostsMgr) newRequest(method, url string, body io.Reader) (req *http.Request, err error) {

	if host, url2, ok := p.hostOf(url); ok {
		req, err = http.NewRequest(method, url2, body)
		req.Host = host
		return
	}
	return http.NewRequest(method, url, body)
}

func (p *hostsMgr) SetHost(host string, portal string) {

	p.hostToPortals[host] = portal
}

// ---------------------------------------------------------------------------

