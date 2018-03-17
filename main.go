package main

import (
	"comp445/la2/httpfs/http"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"strconv"
)

func main() {

	// v := flag.Bool("v", false, "Verbose mode")
	p := flag.String("p", "8080", "Port")
	// d := flag.String("d", "fs", "directory to write files to")

	flag.Parse()

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
	// f, err := os.Open("./data.txt")

	req, err := http.Parse(conn)
	if err != nil {
		log.Printf("error parsing request: %v", err)
		return
	}

	fmt.Printf("REQUEST: %v\n", req)
	res := http.NewResponse(conn)

	if req.Method == "POST" {
		l, err := strconv.Atoi(req.Headers["Content-Length"])
		if err != nil {
			log.Printf("invalid content length: %s", req.Headers["Content-Length"])
			return
		}
		f, err := os.Create(req.URL[1:])
		if err != nil {
			log.Printf("could not open file %s for writing: %v", req.URL[1:], err)
			return
		}
		if _, err = io.CopyN(f, req, int64(l)); err != nil {
			log.Printf("error writing to file: %v", err)
			return
		}
		res.SendStatus(200)
	} else if req.Method == "GET" {
		res.SendFile(req.URL[1:])
	}

	// if _, err = io.Copy(conn, f); err != nil {
	// 	log.Printf("error writing file to connection: %v", err)
	// 	return
	// }
}
