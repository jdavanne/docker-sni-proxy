package main

import (
	"fmt"
	"testing"
)

func ExampleNormal() {
	host, _ := GetHostnameHTTP("GET /host HTTP/1.1\r\nHost: zouzou\r\n\r\n")
	fmt.Println(host)
	// Output: zouzou
}

func ExampleShiftHost() {
	host, _ := GetHostnameHTTP("GET /host HTTP/1.1\r\nContent-Type: application/json\r\nHost: zouzou\r\n\r\n")
	fmt.Println(host)
	// Output: zouzou
}

func TestHostMissing(t *testing.T) {
	_, err := GetHostnameHTTP("GET /host HTTP/1.1\r\nContent-Type: application/json\r\n\r\n")
	if err == nil {
		t.Errorf("Should be error")
	}
}
