package main

import (
	"crypto/rand"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/schollz/progressbar/v3"
)

const (
	name    = "NetMeter - Network performance metering"
	site    = "https://github.com/dassump/netmeter"
	author  = "Daniel Dias de Assumpção <dassump@gmail.com>"
	version = "v1.0.0"
)

var (
	listen         bool
	listen_key     = "listen"
	listen_default = false
	listen_info    = "Listen server"

	host         string
	host_key     = "host"
	host_default = "0.0.0.0:12345"
	host_info    = "Host address and port"

	size         int64
	size_key     = "size"
	size_default = int64(10)
	size_info    = "Size in MiB"

	progress         bool
	progress_key     = "progress"
	progress_default = false
	progress_info    = "Show progress bar"

	quit = make(chan os.Signal, 1)
)

func init() {
	flag.BoolVar(&listen, listen_key, listen_default, listen_info)
	flag.StringVar(&host, host_key, host_default, host_info)
	flag.Int64Var(&size, size_key, size_default, size_info)
	flag.BoolVar(&progress, progress_key, progress_default, progress_info)

	flag.Usage = func() {
		fmt.Printf(
			"%s\n%s\n\nAuthor: %s\nVersion: %s\n\n",
			name, site, author, version,
		)
		fmt.Fprintf(flag.CommandLine.Output(), "Usage of %s:\n", os.Args[0])
		flag.PrintDefaults()
	}

	flag.Parse()

	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)
}

func main() {
	var err error

	if listen {
		err = server()
	} else {
		err = client()
	}

	if err != nil {
		log.Println(err)
		os.Exit(1)
	}
}

func server() error {
	listener, err := net.Listen("tcp", host)
	if err != nil {
		return err
	}
	defer listener.Close()

	log.Println("Listener:", listener.Addr())

	go func() {
		for {
			conn, err := listener.Accept()
			if err != nil {
				continue
			}

			log.Println("Accept:", conn.RemoteAddr())

			go func() {
				defer conn.Close()
				buf := make([]byte, 4096)

				for {
					_, err := conn.Read(buf)
					switch err {
					case nil:
						continue
					case io.EOF:
						log.Println("Closed:", conn.RemoteAddr(), err)
						return
					default:
						log.Println("Error:", conn.RemoteAddr(), err)
					}
				}
			}()
		}
	}()

	<-quit

	return nil
}

func client() error {
	conn, err := net.Dial("tcp", host)
	if err != nil {
		return err
	}
	defer conn.Close()

	start := time.Now()
	total := size * (1024 * 1024)
	written := int64(0)
	errs := int64(0)

	go func() {
		buf := make([]byte, 4096)

		var writer io.Writer
		if progress {
			bar := progressbar.DefaultBytes(-1, conn.RemoteAddr().String())
			defer bar.Clear()
			writer = io.MultiWriter(conn, bar)
		} else {
			writer = conn
		}

		for i := int64(0); i < total; i += int64(4096) {
			conn.SetDeadline(time.Now().Add(100 * time.Millisecond))

			_, _ = rand.Read(buf)
			n, err := writer.Write(buf)
			written += int64(n)
			if err != nil {
				errs++
			}
		}

		quit <- os.Interrupt
	}()

	<-quit

	elapsed := time.Since(start)

	fmt.Printf("total_bytes:%d\n", total)
	fmt.Printf("written_bytes:%d\n", written)
	fmt.Printf("written_percent:%f\n", float64(written*100)/float64(total))
	fmt.Printf("elapsed_seconds:%f\n", elapsed.Seconds())
	fmt.Printf("bytes_per_second:%f\n", float64(written)/(float64(elapsed)/float64(time.Second)))
	fmt.Printf("total_errors:%d\n", errs)

	return nil
}
