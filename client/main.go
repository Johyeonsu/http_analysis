package main

import (
	"bufio"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"sync"
	"time"

	"github.com/lucas-clemente/quic-go"
	"github.com/lucas-clemente/quic-go/http3"
	"github.com/lucas-clemente/quic-go/logging"
	"github.com/lucas-clemente/quic-go/qlog"
	"golang.org/x/net/http2"
)

const sslLogFile = "/home/hyeonsu/.ssl-key.log"

var (
	// Command line flags.
	httpMethod      string
	postBody        string
	followRedirects bool
	onlyHeader      bool
	insecure        bool
	httpHeaders     headers
	saveOutput      bool
	outputFile      string
	showVersion     bool
	clientCertFile  string
	fourOnly        bool
	sixOnly         bool
	addr            string
	httpv           int
	testClientNum   int
	handlerReq      string
)

func init() {
	flag.StringVar(&httpMethod, "X", "GET", "HTTP method to use")
	flag.StringVar(&postBody, "d", "", "the body of a POST or PUT request; from file use @filename")
	flag.BoolVar(&followRedirects, "L", false, "follow 30x redirects")
	flag.BoolVar(&onlyHeader, "I", false, "don't read body of request")
	flag.BoolVar(&insecure, "k", false, "allow insecure SSL connections")
	flag.Var(&httpHeaders, "H", "set HTTP header; repeatable: -H 'Accept: ...' -H 'Range: ...'")
	flag.BoolVar(&saveOutput, "O", false, "save body as remote filename")
	flag.StringVar(&outputFile, "o", "", "output file for body")
	flag.BoolVar(&showVersion, "v", false, "print version number")
	flag.StringVar(&clientCertFile, "E", "", "client cert file for tls config")
	flag.BoolVar(&fourOnly, "4", false, "resolve IPv4 addresses only")
	flag.BoolVar(&sixOnly, "6", false, "resolve IPv6 addresses only")
	flag.StringVar(&addr, "addr", "", "host:port")
	flag.IntVar(&httpv, "http", 3, "http Version")
	flag.IntVar(&testClientNum, "t", 1, "test client")
	flag.StringVar(&handlerReq, "req", "imgload", "request file, page etc..")

	flag.Usage = usage
}

func main() {

	flag.Parse()

	// Set log Format
	t := time.Now().Format("0102_030405")
	logfn := fmt.Sprintf("client_http%d_%s.log", httpv, time.Now().Format("0102_030405"))
	log.SetFlags(0)

	fpLog, err := os.OpenFile(logfn, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer fpLog.Close()
	multiWriter := io.MultiWriter(fpLog, os.Stdout)
	log.SetOutput(multiWriter)

	url := parseURL(addr)
	client := &http.Client{}

	var keyLog io.Writer
	f, err := os.OpenFile(sslLogFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	defer f.Close()
	keyLog = f

	var wg sync.WaitGroup
	wg.Add(testClientNum)
	cn := 0
	for cn < testClientNum {
		go func() {

			if httpv == 1 {
				transport := &http.DefaultTransport
				client.Transport = newDumpTransport(*transport)
				if addr == "" {
					url = parseURL(fmt.Sprintf("http://http-pf.kro.kr:7071/%s", handlerReq))
				}

			} else if httpv == 2 {
				transport := &http2.Transport{
					TLSClientConfig: &tls.Config{
						InsecureSkipVerify: insecure,
						RootCAs:            CaCert(),
						KeyLogWriter:       keyLog,
						MinVersion:         tls.VersionTLS12,
					},
				}
				client.Transport = newDumpTransport(transport)

				if addr == "" {
					url = parseURL(fmt.Sprintf("https://http-pf.kro.kr:7072/%s", handlerReq))
				}

			} else if httpv == 3 {
				var qconf quic.Config
				qconf.Tracer = qlog.NewTracer(func(_ logging.Perspective, connID []byte) io.WriteCloser {
					filename := fmt.Sprintf("client_%s_%x.qlog", t, connID)
					f, err := os.Create(filename)
					if err != nil {
						log.Fatal(err)
					}
					log.Printf("Creating qlog file %s.\n", filename)
					return NewBufferedWriteCloser(bufio.NewWriter(f), f)
				})

				transport := &http3.RoundTripper{
					TLSClientConfig: tlsConfig(keyLog),
					QuicConfig:      &qconf,
				}

				client.Transport = newDumpTransport(transport)
				defer transport.Close()
				if addr == "" {
					url = parseURL(fmt.Sprintf("https://http-pf.kro.kr:7073/%s", handlerReq))
				}
			} else {
				log.Fatalf("Only support http Version 1, 2, 3")
			}
			sendRequest(url, client, keyLog, httpv)
			wg.Done()
		}()
		cn++

	}
	wg.Wait()
	time.Sleep(3 * time.Second)
}
