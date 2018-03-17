package http

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"regexp"
)

type Request struct {
	Method   string
	URL      string
	Headers  map[string]string
	Protocol string
	Version  string
	*bufio.Reader
}

var requestLine = regexp.MustCompile(`^(\w+) ([?=&/\w\-]+) (HTTP|HTTPS)/(\d.\d)\r?\n$`)
var header = regexp.MustCompile(`^([\w-]+): ([\*:/\.\w/]+)\r?\n$`)

func Parse(conn net.Conn) (*Request, error) {

	req := &Request{
		Headers: map[string]string{},
		Reader:  bufio.NewReader(conn),
	}

	line, err := req.ReadString('\n')
	if err != nil {
		log.Printf("Could not parse request-line: %v\n", err)
		return nil, fmt.Errorf("could not parse request-line: %v", err)
	}

	matches := requestLine.FindStringSubmatch(line)
	if len(matches) != 5 {
		return nil, fmt.Errorf("request line is invalid: %s", line)
	}
	if matches[1] == "POST" || matches[1] == "GET" {
		req.Method = matches[1]
	} else {
		return nil, fmt.Errorf("Method not supported: %s", matches[1])
	}
	req.URL = matches[2]
	req.Protocol = matches[3]
	req.Version = matches[4]

	for {
		hl, err := req.ReadString('\n')
		if err != nil {
			return nil, fmt.Errorf("error parsing headers: %v", err)
		}
		if hl == "\r\n" {
			break
		}
		matches := header.FindStringSubmatch(hl)
		if len(matches) != 3 {
			return nil, fmt.Errorf("malformed headers: %s", hl)
		}
		req.Headers[matches[1]] = matches[2]
	}

	return req, nil
}
