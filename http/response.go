package http

import (
	"flag"
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
	verbose bool
	net.Conn
}

func (res response) Set(name, value string) {
	res.headers[name] = value
}

func (res *response) Status(status int) *response {
	res.status = status
	return res
}

func (res response) Send(data string) error {
	res.Set("Content-Length", fmt.Sprintf("%d", len(data)))

	resp := strings.NewReader(fmt.Sprintf(
		"HTTP/1.0 %d %s\r\n%s\r\n%s",
		res.status,
		status2Message[res.status],
		res.headers,
		data,
	))

	var r io.Reader
	if res.verbose {
		r = io.TeeReader(resp, os.Stdout)
	} else {
		r = resp
	}

	if _, err := io.Copy(res, r); err != nil {
		return fmt.Errorf("error writing to tcp connection: %v", err)
	}

	return nil
}

func (res response) SendFile(path string) error {
	f, err := os.Open(path)
	if err != nil {
		if os.IsNotExist(err) {
			return res.Status(404).Send(fmt.Sprintf("Could not find file %s", path))
		}
		return res.SendStatus(500)
	}
	defer f.Close()

	stats, err := f.Stat()
	if err != nil {
		res.SendStatus(500)
	}
	res.Set("Content-Length", fmt.Sprintf("%d", stats.Size()))

	resp := io.MultiReader(
		strings.NewReader(
			fmt.Sprintf(
				"HTTP/1.0 %d %s\r\n%s\r\n",
				200,
				status2Message[200],
				res.headers,
			),
		),
		f,
	)

	var r io.Reader
	if res.verbose {
		r = io.TeeReader(resp, os.Stdout)
	} else {
		r = resp
	}

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
	if res.verbose {
		fmt.Println(resp)
	}
	_, err := fmt.Fprintf(res, resp)
	if err != nil {
		return fmt.Errorf("error sending response: %v", err)
	}
	return nil
}

// NewResponse returns a pointer to a response object given a net.Conn
func NewResponse(conn net.Conn) *response {

	v := flag.Lookup("v").Value.(flag.Getter).Get().(bool)

	defaultHeaders := map[string]string{
		"Connection": "close",
		"Date":       time.Now().Format(time.UnixDate),
	}
	return &response{
		headers: defaultHeaders,
		Conn:    conn,
		status:  200,
		verbose: v,
	}
}
