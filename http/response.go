package http

import (
	"fmt"
	"io"
	"net"
	"os"
	"strings"
	"time"
)

var status2Message = map[int]string{
	200: "OK",
	400: "BAD REQUEST",
	404: "NOT FOUND",
	500: "INTERNAL SERVER ERROR",
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

func (res *response) Status(status int) {
	res.status = status
}

func (res response) SendFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return res.SendStatus(404)
		}
		return res.SendStatus(500)
	}
	stats, err := f.Stat()
	if err != nil {
		res.SendStatus(500)
	}
	res.Set("Content-Length", fmt.Sprintf("%d", stats.Size()))

	resp := fmt.Sprintf(
		"HTTP/1.0 %d %s\r\n%s\r\n",
		200,
		status2Message[200],
		res.headers,
	)

	r := io.MultiReader(strings.NewReader(resp), f)

	if _, err = io.Copy(res, r); err != nil {
		return fmt.Errorf("error writing file to tcp connection: %v", err)
	}

	return nil
}

func (res response) SendStatus(status int) error {
	res.Set("Content-Length", "0")
	resp := fmt.Sprintf(
		"HTTP/1.0 %d %s\r\n%s\r\n",
		status,
		status2Message[status],
		res.headers,
	)
	_, err := fmt.Fprintf(res, resp)
	if err != nil {
		return fmt.Errorf("error sending response: %v", err)
	}
	return nil
}

func NewResponse(conn net.Conn) *response {
	defaultHeaders := map[string]string{
		"Connection": "keep-alive",
		"Date":       time.Now().Format(time.UnixDate),
	}
	return &response{
		headers: defaultHeaders,
		Conn:    conn,
	}
}
