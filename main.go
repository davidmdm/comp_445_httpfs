package main

import (
	"comp445/la2/httpfs/http"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
)

var directory *string
var verbose *bool

var f2m = map[string]*sync.Mutex{}

func main() {

	port := flag.String("p", "8080", "Port")
	verbose = flag.Bool("v", false, "")
	directory = flag.String("d", ".", "directory to write files to")

	flag.Parse()

	if *verbose {
		fmt.Println("Running server in verbose mode\n")
	}

	filenames, err := getFileNames(*directory)
	if err != nil {
		log.Fatal("Could not initialize server directory")
	}

	for _, file := range filenames {
		f2m[file] = &sync.Mutex{}
	}

	ln, err := net.Listen("tcp", fmt.Sprintf(":%s", *port))
	if err != nil {
		log.Fatalf("error listening on port %s: %v", *port, err)
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

		filepath := *directory + req.URL
		mutex := f2m[filepath]

		if mutex != nil {
			mutex.Lock()
			defer mutex.Unlock()
		} else {
			f2m[filepath] = &sync.Mutex{}
			f2m[filepath].Lock()
			defer f2m[filepath].Unlock()
		}
		f, err := os.Create(filepath)
		if err != nil {
			log.Printf("could not open file %s for writing: %v", req.URL[1:], err)
			return
		}
		defer f.Close()

		var r io.Reader
		if *verbose {
			r = io.TeeReader(req, os.Stdout)
		} else {
			r = req
		}

		if _, err = io.CopyN(f, r, int64(l)); err != nil {
			log.Printf("error writing to file: %v", err)
			return
		}

		if *verbose {
			fmt.Print("\n\n")
		}

		if err = res.SendStatus(200); err != nil {
			log.Printf("could not send response: %v", err)
		}

	} else if req.Method == "GET" {

		if req.URL == "/" {
			files, err := getFileNames(*directory)
			if err != nil {
				log.Printf("could not read directory: %v", err)
				return
			}
			if err = res.Send(strings.Join(files, "\r\n") + "\r\n"); err != nil {
				log.Printf("could not send response: %v", err)
			}
		} else {
			if err = res.SendFile(req.URL[1:]); err != nil {
				log.Printf("could not send response: %v", err)
			}
		}

	}

}

func getFileNames(directory string) ([]string, error) {
	files, err := ioutil.ReadDir(directory)
	if err != nil {
		return nil, err
	}
	var filenames = []string{}
	for _, file := range files {
		if !file.IsDir() {
			filenames = append(filenames, file.Name())
		}
	}
	return filenames, nil
}
