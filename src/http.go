package main

import (
	"errors"
	"regexp"
)

var re = regexp.MustCompile("(?s)^[^\r\n]+HTTP/1.1.*\r\n(?i)Host: *(\\S+) *\r\n")

func GetHostnameHTTP(data string) (string, error) {
	hosts := re.FindAllStringSubmatch(data, -1)
	if len(hosts) > 0 {
		return hosts[0][1], nil
	}
	return "", errors.New("HTTP/1.1 Host header not found")
}
