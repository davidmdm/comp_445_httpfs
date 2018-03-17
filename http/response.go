package http

import (
	"fmt"
	"net"
	"time"
)

var status2Message = map[string]string{
	"200": "OK",
	"400": "BAD REQUEST",
	"404": "NOT FOUND",
	"500": "INTERNAL SERVER ERROR",
}

type headers map[string]string

func (h headers) String() (ret string) {
	for k, v := range h {
		ret += fmt.Sprintf("%s: %s\r\n", k, v)
	}
	return
}

type response struct {
	status  int
	headers headers
	net.Conn
}

func (res response) Set(name, value string) {
	res.headers[name] = value
}

func (res response) SendStatus(int) error {
	res.Set("Content-Length", "0")
	resp := fmt.Sprintf("HTTP/1.0 %s ")
	_, err := fmt.Fprintf(res)
}

func NewResponse(conn net.Conn) *response {

	defaultHeaders := map[string]string{
		"Connection": "close",
		"Date":       time.Now().Format(time.UnixDate),
	}

	return &response{
		headers: defaultHeaders,
		Conn:    conn,
	}
}
