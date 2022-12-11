package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"
	"time"

	"github.com/bryanwux/V3Ray/common"
	"github.com/bryanwux/V3Ray/proxy"
	"github.com/bryanwux/V3Ray/proxy/direct"
)

var (
	version  = "0.1.0"
	codename = "An simple implementation of V2Ray"

	// Flag
	f = flag.String("f", "client.json", "config file name")
)

const (
	// router mode
	whitelist = "whitelist"
	blacklist = "blacklist"
)

// Version
func printVersion() {
	fmt.Printf("V3Ray %v (%v), %v %v %v\n", version, codename, runtime.Version(), runtime.GOOS, runtime.GOARCH)
}

// Config
type Config struct {
	Local  string `json:"local"`
	Route  string `json:"route"`
	Remote string `json:"remote"`
}

func loadConfig(configName string) (*Config, error) {
	path := common.GetPath(configName)
	if len(path) > 0 {
		configJson, err := os.Open(path)

		if err != nil {
			return nil, fmt.Errorf("can not open config file %v, %v", configName, err)
		}
		defer configJson.Close()

		bytes, _ := ioutil.ReadAll(configJson)
		config := &Config{}
		// parse local json config file
		if err = json.Unmarshal(bytes, config); err != nil {
			return nil, fmt.Errorf("can not parse config file %v, %v, something wrong with the JSON", configName, err)
		}
		return config, nil
	}
	return nil, fmt.Errorf("can not load config file %v", configName)
}

func main() {
	printVersion()
	flag.Parse()

	// read config file, default to client mode
	conf, err := loadConfig(*f)
	if err != nil {
		log.Printf("can not load config file: %v", err)
		os.Exit(-1)
	}

	// initialize server and client with config json
	localServer, err := proxy.ServerFromURL(conf.Local)
	if err != nil {
		log.Printf("can not create local server: %v", err)
		os.Exit(-1)
	}
	defer localServer.Stop()

	remoteClient, err := proxy.ClientFromURL(conf.Remote)
	if err != nil {
		log.Printf("can not create remote client: %v", err)
		os.Exit(-1)
	}

	directClient, _ := proxy.ClientFromURL("direct://")
	//matcher := common.NewMather(conf.Route)

	listener, err := net.Listen("tcp", localServer.Addr())
	if err != nil {
		log.Printf("can not listen on %v: %v", localServer.Addr(), err)
		os.Exit(-1)
	}
	log.Printf("%v listening TCP on %v", localServer.Name(), localServer.Addr())

	go func() {
		for {
			lc, err := listener.Accept()
			if err != nil {
				errStr := err.Error()
				if strings.Contains(errStr, "closed") {
					break
				}
				log.Printf("failed to accepted connection: %v", err)

				if strings.Contains(errStr, "too many") {
					time.Sleep(time.Millisecond * 500)
				}
				continue
			}

			go func() {
				defer lc.Close()
				var client proxy.Client
				wlc, targetAddr, err := localServer.Handshake(lc)
				if err != nil {
					log.Printf("failed in handshake from %v: %v", localServer.Addr(), err)
					return
				}

				// route matching
				if conf.Route == whitelist { // whitelist mode, if matching then connect directly, else use proxy server

				} else if conf.Route == blacklist { // blacklis mode, if matching then use proxy server , else connect directly

				} else {
					client = remoteClient // use proxy server globally
				}
				log.Printf("%v to %v", client.Name(), targetAddr)

				// connect to remote client
				dialAddr := remoteClient.Addr()
				if _, ok := client.(*direct.Direct); ok {
					dialAddr = targetAddr.String()
				}
				rc, err := net.Dial("tcp", dialAddr)
				if err != nil {
					log.Printf("failed to dail to %v: %v", dialAddr, err)
					return
				}
				defer rc.Close()

				wrc, err := client.Handshake(rc, targetAddr.String())
				if err != nil {
					log.Printf("failed in handshake to %v: %v", dialAddr, err)
					return
				}

				//traffic forwarding
				go io.Copy(wrc, wlc)
				io.Copy(wlc, wrc)
			}()
		}
	}()

	// run in background
	{
		osSignals := make(chan os.Signal, 1)
		signal.Notify(osSignals, os.Interrupt, os.Kill, syscall.SIGTERM)
		<-osSignals
	}
}
