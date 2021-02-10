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
var FlutterStorageBaseUrl string
var PubHostedUrl string

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
		_, err = cfg.Section("default").NewKey("proxy_url", "http://YouHttpProxyUrl:Port")
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
		_, _ = fmt.Scanln(&s)
		main()
	}

	http.DefaultTransport = &http.Transport{Proxy: http.ProxyURL(proxyUrl)}
	openProxy(serverAddrString)
}

func openProxy(addr string) {

	var httpUrl = "http://" + addr + "/"
	FlutterStorageBaseUrl = httpUrl + "storage"
	PubHostedUrl = httpUrl + "pub"

	httpClient = http.Client{}
	/// open Proxy
	http.HandleFunc("/", proxyHandleFunc)
	fmt.Println("server start with " + addr +
		"\n FLUTTER_STORAGE_BASE_URL=" + FlutterStorageBaseUrl +
		"\n PUB_HOSTED_URL=" + PubHostedUrl)
	fmt.Println("--------------------------------------------------------------------------------")
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
	var resp *http.Response
	var err error
	log.Println(request.Method + " " + request.URL.String())
	switch request.Method {
	case "GET":
		resp, err = httpClient.Get(request.URL.String())
		break
	case "POST":
		resp, err = httpClient.Post(request.URL.String(), request.Header.Get("Content-Type"), request.Body)
	}

	if err != nil {
		log.Println(err)
		writer.WriteHeader(500)
		return
	}
	defer func() {
		if resp.Body != nil {
			_ = resp.Body.Close()
		}
	}()
	writer.Header().Set("Content-Type", resp.Header.Get("Content-Type"))
	body, err := ioutil.ReadAll(resp.Body)
	var sBody = string(body)
	sBody = strings.ReplaceAll(sBody, "https://pub.dartlang.org/", PubHostedUrl+"/")
	_, err = writer.Write([]byte(sBody))
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
