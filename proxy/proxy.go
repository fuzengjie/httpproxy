package proxy

import (
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"httpproxy/config"
	ss "github.com/shadowsocks/shadowsocks-go/shadowsocks"
	log "github.com/sirupsen/logrus"
)

type ProxyServer struct {
	ssServer string
	ssCipher *ss.Cipher
}

func NewProxyServer() *http.Server {
	p := &ProxyServer{
		ssServer: config.Conf.SsServer,
	}

	var err error
	p.ssCipher, err = ss.NewCipher(config.Conf.SsCipher, config.Conf.SsPassword)
	if err != nil {
		panic("init cipher error: " + err.Error())
	}

	return &http.Server{
		Addr:    config.Conf.Listen,
		Handler: p,
	}
}

func (p *ProxyServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	defer func() {
		if err := recover(); err != nil {
			http.Error(w, fmt.Sprint(err), http.StatusInternalServerError)
			log.Debugf("panic: 代理服务错误:%v\n", err)
		}
	}()
	auth := p.Auth(r)
	if !auth {
		log.Errorf("client %s proxy auth faild",r.RemoteAddr)
		w.Header().Add("Proxy-Authenticate","Basic realm=User Auth")
		w.Header().Add("Server","FuZZ inc")
		w.WriteHeader(407)
		w.Write([]byte("用户验证失败"))
	} else {
		if r.Method == "CONNECT" {
			p.HandleConnect(w, r)
		} else {
			p.HandleHttp(w, r)
		}
	}

}

// 处理HTTPS、HTTP2代理请求
func (p *ProxyServer) HandleConnect(w http.ResponseWriter, r *http.Request) {
	log.Infof("https handler:  %s %s", r.Method, r.Host)

	hj, _ := w.(http.Hijacker)
	conn, _, err := hj.Hijack()
	if err != nil {
		log.Error(err)
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	ssConn, err := ss.Dial(r.URL.Host, p.ssServer, p.ssCipher.Copy())
	if err != nil {
		log.Error("ss dial: ", err)
		http.Error(w, err.Error(), http.StatusBadGateway)
		return
	}

	conn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))

	go ss.PipeThenClose(conn, ssConn, nil)
	ss.PipeThenClose(ssConn, conn, nil)
}

// 处理HTTP代理请求
func (p *ProxyServer) HandleHttp(w http.ResponseWriter, r *http.Request) {
	log.Infof("http handler: %s %s", r.Method, r.URL)

	// ss proxy
	tr := http.Transport{
		Dial: func(network, addr string) (net.Conn, error) {
			log.Infof("dial ss %v/%v", addr, network)
			return ss.Dial(addr, p.ssServer, p.ssCipher.Copy())
		},
	}

	// transport
	resp, err := tr.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		log.Error("request error: ", err)
		return
	}
	defer resp.Body.Close()

	// copy headers
	for k, values := range resp.Header {
		for _, v := range values {
			w.Header().Add(k, v)
		}
	}
	w.WriteHeader(resp.StatusCode)

	// copy body
	n, err := io.Copy(w, resp.Body)
	if err != nil && err != io.EOF {
		log.Errorf("copy response body error: %v", err)
	}

	log.Infof("copied %v bytes from %v.", n, r.Host)
}

func (p *ProxyServer) Auth(r *http.Request) bool {
	if len(config.Conf.Auth) == 0 {
		log.Infof("config not found auth")
		return true
	}
	auth := r.Header.Get("Proxy-Authorization")
	log.Error("auth:",auth,)
	auth = strings.Replace(auth, "Basic ", "", 1)
	if auth == "" {
		return false
	}
	data, err := base64.StdEncoding.DecodeString(auth)
	if err != nil {
		log.Debug("Fail to decoding Proxy-Authorization, %v, got an error of %v", auth, err)
		return false
	}
	userPasswdPair := strings.Split(string(data), ":")
	if len(userPasswdPair) != 2 {
		return false
	}
	var user, passwd string
	user = userPasswdPair[0]
	passwd = userPasswdPair[1]
	for _,auth := range config.Conf.Auth {
		if auth.User == user && auth.Pwd == passwd {
			log.Debugf("proxy user:%s auth success",user)
			return true
		}
	}
	return false
}
