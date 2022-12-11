package common

import (
	"bufio"
	"net"
	"os"
	"strings"

	"github.com/yl2chen/cidranger"
)

type Matcher struct {
	NetRanger cidranger.Ranger
	IPMap     map[string]net.IP
	DomainMap map[string]string
}

// check if a cidr/ip/domain name is in a predefined list
func NewMatcher(configName string) *Matcher {
	ranger := cidranger.NewPCTrieRanger()
	ipMap := make(map[string]net.IP)
	domainMap := make(map[string]string)

	path := GetPath(configName)
	if len(path) > 0 {
		if cf, err := os.Open(path); err == nil {
			defer cf.Close()
			scanner := bufio.NewScanner(cf)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if len(line) == 0 {
					continue
				}
				// A cidr address, add it to ranger
				if strings.Contains(line, "/") {
					if _, net, err := net.ParseCIDR(line); err == nil {
						ranger.Insert(cidranger.NewBasicRangerEntry(*net))
					}
					continue
				}
				// An ip address, add it to ipMap
				if ip := net.ParseIP(line); ip != nil {
					ipMap[line] = ip
					continue
				}
				domainMap[line] = line

			}
		}
	}

	return &Matcher{
		NetRanger: ranger,
		IPMap:     ipMap,
		DomainMap: domainMap,
	}
}

// check if a host is in a predefined list
func (m *Matcher) Check(host string) bool {
	if ip := net.ParseIP(host); ip != nil {
		if m.NetRanger != nil {
			if contains, _ := m.NetRanger.Contains(ip); contains {
				return true
			}
		}

		if m.IPMap != nil {
			if _, contains := m.IPMap[host]; contains {
				return true
			}
		}
	}

	tokens := strings.Split(host, ".")
	if len(tokens) > 1 {
		suffix := tokens[len(tokens)-1]
		for i := len(tokens) - 2; i >= 0; i-- {
			suffix = tokens[i] + "." + suffix
			if _, found := m.DomainMap[suffix]; found {
				return true
			}
		}
	}

	return false
}
