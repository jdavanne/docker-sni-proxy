package main

import (
	"fmt"
	"testing"
)

func ExampleGetHostnameHTTP() {
	host, _ := GetHostnameHTTP("GET /host HTTP/1.1\r\nHost: zouzou\r\n\r\n")
	fmt.Println(host)
	// Output: zouzou
}

func ExampleGetHostnameHTTP_ShiftHost() {
	host, _ := GetHostnameHTTP("GET /host HTTP/1.1\r\nContent-Type: application/json\r\nHost: zouzou\r\n\r\n")
	fmt.Println(host)
	// Output: zouzou
}

func ExampleGetHostnameHTTP_HostCase() {
	host, _ := GetHostnameHTTP("GET /host HTTP/1.1\r\nContent-Type: application/json\r\nhOST: zouzou\r\n\r\n")
	fmt.Println(host)
	// Output: zouzou
}

func TestHostMissing(t *testing.T) {
	_, err := GetHostnameHTTP("GET /host HTTP/1.1\r\nContent-Type: application/json\r\n\r\n")
	if err == nil {
		t.Errorf("Should be error")
	}
}
func TestNoyHTTP11(t *testing.T) {
	_, err := GetHostnameHTTP("GET /host HTTP/1.0\r\nContent-Type: application/json\r\nHost: zouzou\r\n\r\n")
	if err == nil {
		t.Errorf("Should be error")
	}
}
