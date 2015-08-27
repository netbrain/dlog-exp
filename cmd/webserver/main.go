package main

import (
	"flag"
	"fmt"
	"net"
	"net/http"
	"strings"
	"time"

	"github.com/netbrain/dlog/client"
)

var servers string

var writeClient *client.WriteClient
var readClient *client.ReadClient

func init() {
	flag.StringVar(&servers, "servers", "", "dlog servers to connect to")
}

func log(w http.ResponseWriter, r *http.Request) {
	writeClient.Write([]byte(fmt.Sprintf("Time: %d\nIP: %s\n", time.Now().Nanosecond(), externalIP())))
}

func replay(w http.ResponseWriter, r *http.Request) {
	for c := range readClient.Replay() {
		w.Write(c)
	}
}

func main() {
	flag.PrintDefaults()
	flag.Parse()

	s := strings.Split(servers, ",")
	writeClient = client.NewWriteClient(s)
	readClient = client.NewReadClient(s)

	http.HandleFunc("/record", log)
	http.HandleFunc("/replay", replay)

	http.ListenAndServe(":80", nil)
}

func externalIP() string {
	ifaces, err := net.Interfaces()
	if err != nil {
		return ""
	}
	for _, iface := range ifaces {
		if iface.Flags&net.FlagUp == 0 {
			continue // interface down
		}
		if iface.Flags&net.FlagLoopback != 0 {
			continue // loopback interface
		}
		addrs, err := iface.Addrs()
		if err != nil {
			return ""
		}
		for _, addr := range addrs {
			var ip net.IP
			switch v := addr.(type) {
			case *net.IPNet:
				ip = v.IP
			case *net.IPAddr:
				ip = v.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			ip = ip.To4()
			if ip == nil {
				continue // not an ipv4 address
			}
			return ip.String()
		}
	}
	return ""
}
