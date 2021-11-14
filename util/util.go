package util

import (
	"log"
	"net"
	"path/filepath"
	"strings"
)

const (
	defaultDataDir = "tomato_dat"
	confName       = "conf.json"
)

func GetDataDir() string {
	return defaultDataDir
}

func GetConfigPath() string {
	return filepath.Join(GetDataDir(), confName)
}

// Get IP from certain(first) network interface
//
func GetDefaultUri() (string, error) {
	ifaces, err := net.Interfaces()
	if err != nil {
		log.Println(err)
		return "", err
	}

	var targetIp net.IP
	for _, iface := range ifaces {
		if !strings.Contains(iface.Name, "eth") {
			continue
		}

		addrs, err := iface.Addrs()
		if err != nil {
			log.Println(err)
			continue
		}

		for _, addr := range addrs {
			switch v := addr.(type) {
			case *net.IPNet:
				if targetIp == nil {
					targetIp = v.IP
				}
			case *net.IPAddr:
				if targetIp == nil {
					targetIp = v.IP
				}
				break
			}
		}

		if targetIp != nil {
			break
		}

	}
	return "http://" + targetIp.String() + ":8000", nil
}
