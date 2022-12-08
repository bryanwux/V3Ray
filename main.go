package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"runtime"

	"github.com/bryanwux/V3Ray/common"
	"github.com/bryanwux/V3Ray/proxy"
)

var (
	version  = "0.1.0"
	codename = "An simple implementation of V2Ray"

	// Flag
	f = flag.String("f", "client.json", "config file name")
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

	// initialize local server with config
	localServer, err := proxy.ServerFromURL(conf.Local)
	if err != nil {
		log.Printf("can not create local server: %v", err)
		os.Exit(-1)
	}
	defer localServer.Stop()

}
