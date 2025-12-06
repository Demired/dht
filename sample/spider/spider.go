package main

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log"
	"net/http"

	"github.com/Demired/dht"
)

func Run() {
	defer func() {
		if err := recover(); err != nil {
			log.Printf("bt:%s", err)
		}
	}()
	type file struct {
		Path   []interface{} `json:"path"`
		Length int           `json:"length"`
	}

	type bitTorrent struct {
		InfoHash string `json:"magnet"`
		Name     string `json:"name"`
		Files    []file `json:"files,omitempty"`
		// Length   int    `json:"length,omitempty"`
	}

	type downTorrent struct {
		IP       string `json:"ip"`
		Port     int    `json:"port"`
		InfoHash string `json:"magnet"`
	}

	downloader := dht.NewWire(128*1024, 1024, 100)
	go func() {
		for resp := range downloader.Response() {
			metadata, err := dht.Decode(resp.MetadataInfo)
			if err != nil {
				continue
			}
			info := metadata.(map[string]interface{})
			if _, ok := info["name"]; !ok {
				continue
			}
			bt := bitTorrent{
				InfoHash: hex.EncodeToString(resp.InfoHash),
				Name:     info["name"].(string),
			}
			if v, ok := info["files"]; ok {
				files := v.([]interface{})
				bt.Files = make([]file, len(files))
				for i, item := range files {
					f := item.(map[string]interface{})
					bt.Files[i] = file{
						Path:   f["path"].([]interface{}),
						Length: f["length"].(int),
					}
				}
				// } else if _, ok := info["length"]; ok {
				// 	bt.Length = info["length"].(int)
			}
			btJSON, _ := json.Marshal(bt)
			resf, _ := http.Post("https://clientapi.ipip.net/bt/file", "Content-Type:application/json", bytes.NewBuffer(btJSON))
			down := downTorrent{
				IP:       resp.IP,
				Port:     resp.Port,
				InfoHash: hex.EncodeToString(resp.InfoHash),
			}
			downJSON, _ := json.Marshal(down)
			resd, _ := http.Post("https://clientapi.ipip.net/bt/down", "Content-Type:application/json", bytes.NewBuffer(downJSON))
			resd.Body.Close()
			resf.Body.Close()
		}
	}()
	go downloader.Run()

	config := dht.NewCrawlConfig()
	config.Network = "udp"
	config.MaxNodes = 1000
	config.PrimeNodes = []string{
		"router.bittorrent.com:6881",
		"router.utorrent.com:6881",
		"dht.transmissionbt.com:6881",
		"2604:a880:2:d0::21d:c001:6881",
		"2001:19f0:5:1337:5400:1ff:fe90:7e6f:6881",
		"2600:3c0d::f03c:93ff:fe61:ae27:6881",
		"2001:41d0:203:4cca:5::6881",
	}
	config.Address = "[::]:6881"
	config.BlackListMaxSize = 3000
	config.OnGetPeers = func(infoHash, ip string, port int) {
		fmt.Printf("OnGetPeers infoHash: xxx ip: %s port: %d\n", ip, port)
	}
	config.OnAnnouncePeer = func(infoHash, ip string, port int) {
		fmt.Printf("OnAnnouncePeer: %s %s %d\n", infoHash, ip, port)
		// downloader.Request([]byte(infoHash), ip, port)
	}
	d := dht.New(config)
	d.Run()
}
