package main

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/bitly/go-simplejson"
)

func main() {
	start := time.Now()
	hostContent := ""
	domain := ""
	var hkIp string
	ip, err1 := GetLocalPublicIp()
	if err1 != nil {
		fmt.Println(err1.Error())
	} else {
		resp, err := httpGet("http://ip.taobao.com/service/getIpInfo.php?ip=" + ip)
		if nil != err {
			panic(err)
		}
		// fmt.Println(resp)
		json, err2 := simplejson.NewJson(resp)
		if nil != err2 {
			panic(err2)
		}
		region := json.Get("data").Get("region").MustString()
		region = strings.TrimRight(region, "省")
		isp := json.Get("data").Get("isp").MustString()
		isp = region + isp

		dnsFile := "./AppleDNS/List.md"
		fd, err3 := os.Open(dnsFile)
		if nil != err3 {
			panic(err3)
		}
		defer fd.Close()
		buf := bufio.NewReader(fd)
		for {
			line, fileError := buf.ReadString('\n')
			if nil != fileError || io.EOF == fileError {
				break
			}
			if strings.HasPrefix(line, "##") && strings.Contains(line, ".com") {
				if "" != hkIp {
					hostContent += hkIp + " " + domain + "\n"
				}
				domain = strings.TrimLeft(strings.TrimRight(line, "\n"), "## ")

			} else if strings.Contains(line, isp) {
				hostContent += strings.Fields(strings.TrimLeft(line, isp))[0] + " " + domain + "\n"
				hkIp = ""
			} else if strings.Contains(line, "香港") {
				hkIp = strings.Fields(strings.TrimLeft(line, "香港 "))[0]
			}
		}
		fmt.Println(hostContent)

	}

	// fmt.Println(newUrl)
	fmt.Println(time.Now().Sub(start))
}

func httpGet(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		// handle error
		panic(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
		panic(err)
	}

	return body, err
}

func GetLocalPublicIp() (string, error) {
	timeout := time.Nanosecond * 30
	conn, err := net.DialTimeout("tcp", "ns1.dnspod.net:6666", timeout*time.Second)
	defer func() {
		if x := recover(); x != nil {
			log.Println("Can't get public ip", x)
		}
		if conn != nil {
			conn.Close()
		}
	}()
	if err == nil {
		var bytes []byte
		deadline := time.Now().Add(timeout * time.Second)
		err = conn.SetDeadline(deadline)
		if err != nil {
			return "", err
		}
		bytes, err = ioutil.ReadAll(conn)
		if err == nil {
			return string(bytes), nil
		}
	}
	return "", err
}
