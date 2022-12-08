package proxy

import (
	"errors"
	"io"
	"log"
	"net"
	"net/url"
	"strings"
)

type Server interface {
	Name() string
	Addr() string
	Handshake(underlay net.Conn) (io.ReadWriter, *TargetAddr, error)
	Stop()
}

type ServerCreator func(url *url.URL) (Server, error)

var serverMap = make(map[string]ServerCreator)

func RegisterServer(name string, c ServerCreator) {
	serverMap[name] = c
}

// calls registered creator to create proxy servers
func ServerFromURL(s string) (Server, error) {
	u, err := url.Parse(s)
	if err != nil {
		log.Printf("can not parse server url %s err: %s", s, err)
		return nil, err
	}
	c, ok := serverMap[strings.ToLower(u.Scheme)]
	if ok {
		return c(u)
	}

	return nil, errors.New("unknown server scheme '" + u.Scheme + "'")
}

type TargetAddr struct {
	Name string
	IP   net.IP
	Port int
}
