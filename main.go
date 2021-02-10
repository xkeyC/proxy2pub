package main

import (
	"fmt"
	"gopkg.in/ini.v1"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"strings"
)

var httpClient http.Client

func main() {
	/// check Conf
	if !exists("proxy2pub.ini") {
		file, err := os.Create("proxy2pub.ini")
		if err != nil {
			log.Fatal(err)
		}
		_ = file.Close()
		cfg, err := ini.Load("proxy2pub.ini")
		if err != nil {
			log.Fatal(err)
		}
		_, err = cfg.Section("default").NewKey("server_addr", "127.0.0.1:59776")
		_, err = cfg.Section("default").NewKey("proxy_url", "http://127.0.0.1:7890")
		if err != nil {
			log.Fatal(err)
		}
		err = cfg.SaveTo("proxy2pub.ini")
		if err != nil {
			log.Fatal(err)
		}
	}

	/// load Conf
	cfg, err := ini.Load("proxy2pub.ini")
	if err != nil {
		log.Fatal(err)
	}
	proxyKey, err := cfg.Section("default").GetKey("proxy_url")
	serverKey, err := cfg.Section("default").GetKey("server_addr")
	if err != nil {
		log.Fatal(err)
	}
	var proxyUrlString = proxyKey.Value()
	var serverAddrString = serverKey.Value()

	proxyUrl, err := url.Parse(proxyUrlString)
	if err != nil {
		fmt.Println("HTTP PROXY URL HAS ERROR:\n" + err.Error() + "\nPress enter to try again:")
		var s string
		_, _ = fmt.Scanln("%q", &s)
		main()
	}

	http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
	openProxy(serverAddrString)
}

func openProxy(addr string) {
	httpClient = http.Client{}
	/// open Proxy
	http.HandleFunc("/", proxyHandleFunc)
	fmt.Println("server start with " + addr)
	err := http.ListenAndServe(addr, nil)
	if err != nil {
		log.Println(err)
		openProxy(addr)
	}
}

func proxyHandleFunc(writer http.ResponseWriter, request *http.Request) {
	var domain = ""
	if strings.Contains(request.URL.Path, "/pub") {
		request.URL.Path = strings.Replace(request.URL.Path, "/pub", "", 1)
		domain = "pub.dev"
	} else if strings.Contains(request.URL.Path, "/storage") {
		request.URL.Path = strings.Replace(request.URL.Path, "/storage", "", 1)
		domain = "storage.googleapis.com"
	} else {
		return
	}
	request.URL.Host = domain
	request.URL.Scheme = "https"
	resp, err := httpClient.Get(request.URL.String())
	if err != nil {
		log.Println(err)
		writer.WriteHeader(500)
		return
	}
	defer func() {
		_ = resp.Body.Close()
	}()
	body, err := ioutil.ReadAll(resp.Body)
	_, err = writer.Write(body)
	if err != nil {
		log.Println(err)
		writer.WriteHeader(500)
	}
}

func exists(path string) bool {
	_, err := os.Stat(path)
	if err != nil {
		if os.IsExist(err) {
			return true
		}
		return false
	}
	return true
}
