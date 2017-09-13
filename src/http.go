package main

import (
	"errors"
	"strings"

	log "github.com/sirupsen/logrus"
)

func GetHostnameHTTP(data string) (string, error) {
	parts := strings.SplitN(data, "\n", 3)
	if len(parts) < 3 {
		return "", errors.New("Host Header is missing")
	}
	parts2 := strings.SplitN(parts[1], ":", 3)
	if len(parts2) < 2 {
		return "", errors.New("Host Header is malformed : " + parts[1])
	}

	if strings.ToUpper(parts2[0]) != "HOST" {
		return "", errors.New("Second Header is not 'Host: xxxxxx' : " + parts[1])
	}

	hostname := strings.Trim(parts2[1], " ")
	log.Println("GetHostnameHTTP", hostname)
	return hostname, nil
}
