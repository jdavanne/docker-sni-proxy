package main

import (
	"errors"
	"regexp"
)

func GetHostnameHTTP(data string) (string, error) {
	re := regexp.MustCompile("\r\n(?i)Host: *(\\S+) *\r\n")
	hosts := re.FindAllStringSubmatch(data, -1)
	if len(hosts) > 0 {
		return hosts[0][1], nil
	}
	return "", errors.New("Host header not found")
}
