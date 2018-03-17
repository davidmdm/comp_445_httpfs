package main

import (
	"comp445/la2/httpfs/http"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
)

var options = map[string]string{}

func main() {

	// v := flag.Bool("v", false, "Verbose mode")
	p := flag.String("p", "8080", "Port")
	d := flag.String("d", ".", "directory to write files to")

	flag.Parse()

	options["d"] = *d

	ln, err := net.Listen("tcp", fmt.Sprintf(":%s", *p))
	if err != nil {
		log.Fatalf("error listening on port %s: %v", *p, err)
	}
	defer ln.Close()

	for {
		conn, err := ln.Accept()
		if err != nil {
			log.Printf("error accepting connection: %v", err)
			continue
		}
		go handleConnection(conn)
	}

}

func handleConnection(conn net.Conn) {
	defer conn.Close()

	req, err := http.Parse(conn)
	if err != nil {
		log.Printf("error parsing request: %v", err)
		return
	}

	res := http.NewResponse(conn)

	if req.Method == "POST" {

		if req.URL == "/" {
			if err = res.Status(400).Send("Cannot upload new file under path`/`... please choose a filename\r\n"); err != nil {
				log.Printf("could not send response: %v", err)
			}
			return
		}

		if _, prs := req.Headers["Content-Length"]; !prs {
			if err = res.Status(400).Send("Content-Length header is required"); err != nil {
				log.Printf("could not send response: %v", err)
			}
			return
		}

		l, err := strconv.Atoi(req.Headers["Content-Length"])
		if err != nil {
			log.Printf("could not read content-length: %v. value: %v", err, req.Headers["Content-Length"])
			return
		}
		f, err := os.Create(options["d"] + req.URL)
		if err != nil {
			log.Printf("could not open file %s for writing: %v", req.URL[1:], err)
			return
		}

		defer f.Close()

		if _, err = io.CopyN(f, req, int64(l)); err != nil {
			log.Printf("error writing to file: %v", err)
			return
		}

		if err = res.SendStatus(200); err != nil {
			log.Printf("could not send response: %v", err)
		}

	} else if req.Method == "GET" {

		if req.URL == "/" {
			files, err := filepath.Glob("*")
			if err != nil {
				log.Printf("could not read directory: %v", err)
				return
			}
			if err = res.Send(strings.Join(files, "\r\n")); err != nil {
				log.Printf("could not send response: %v", err)
			}
		} else {
			if err = res.SendFile(req.URL[1:]); err != nil {
				log.Printf("could not send response: %v", err)
			}
		}

	}

}
