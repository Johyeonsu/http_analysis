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
	httpMethod string
	// postBody      string
	saveOutput    bool
	outputFile    string
	insecure      bool
	addr          string
	httpv         int
	testClientNum int
	testReqNum    int
	handlerReq    string
)

func init() {
	flag.StringVar(&httpMethod, "X", "GET", "HTTP method to use")
	// flag.StringVar(&postBody, "d", "", "the body of a POST or PUT request; from file use @filename")
	flag.BoolVar(&saveOutput, "O", false, "save body as remote filename")
	flag.StringVar(&outputFile, "o", "", "output file for body")
	flag.BoolVar(&insecure, "k", false, "allow insecure SSL connections")
	flag.StringVar(&addr, "addr", "", "host:port")
	flag.IntVar(&httpv, "http", 3, "http Version")
	flag.IntVar(&testClientNum, "c", 1, "test client num")
	flag.IntVar(&testReqNum, "r", 1, "test request num")
	flag.StringVar(&handlerReq, "req", "pageload", "request file, page etc..")

	flag.Usage = usage
}

func main() {
	flag.Parse()

	t := time.Now().Format("0102_030405")
	logfn := fmt.Sprintf("client_http%d_%s.log", httpv, t)
	log.SetFlags(0)

	var qlogfn string
	fpLog, err := os.OpenFile(logfn, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer fpLog.Close()
	multiWriter := io.MultiWriter(fpLog, os.Stdout)
	log.SetOutput(multiWriter)

	url := parseURL(addr)
	var keyLog io.Writer
	f, err := os.OpenFile(sslLogFile, os.O_APPEND|os.O_CREATE|os.O_RDWR, 0644)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	defer f.Close()
	keyLog = f

	client := &http.Client{}
	if httpv == 1 {
		transport := &http.DefaultTransport
		client.Transport = newDumpTransport(*transport)

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

	} else if httpv == 3 {
		var qconf quic.Config

		qconf.Tracer = qlog.NewTracer(func(_ logging.Perspective, connID []byte) io.WriteCloser {
			filename := fmt.Sprintf("client_%x.qlog", connID)
			f, err := os.Create(filename)
			if err != nil {
				log.Fatal(err)
			}
			log.Printf("Creating qlog file %s.\n", filename)
			qlogfn = filename
			return NewBufferedWriteCloser(bufio.NewWriter(f), f)
		})
		transport := &http3.RoundTripper{
			TLSClientConfig: tlsConfig(keyLog),
			QuicConfig:      &qconf,
		}
		client.Transport = newDumpTransport(transport)
		defer transport.Close()
	}

	if httpv == 1 {
		url = parseURL(fmt.Sprintf("http://http-pf.kro.kr:707%d/%s", httpv, handlerReq))
	} else {
		url = parseURL(fmt.Sprintf("https://http-pf.kro.kr:707%d/%s", httpv, handlerReq))
	}

	var wg1 sync.WaitGroup
	wg1.Add(testClientNum)
	cn := 0
	t_start := time.Now()
	for cn < testClientNum {
		go func(cn int) {
			sendRequest(url, client, keyLog, httpv, testReqNum, cn)
			wg1.Done()
		}(cn)
		cn++

	}
	wg1.Wait()
	loginfo := fmt.Sprintf("HTTP%d TOTAL RUN", httpv)
	printTimeDuration(loginfo, t_start, time.Now())

	if httpv == 3 {
		log.Printf("\npython3 log_exc.py %s\n", qlogfn)
		// cmd := exec.Command("python3", "log_exc.py", qlogfn)
		// out, err := cmd.Output()
		// if err != nil {
		// 	log.Fatalf("%+v", err)
		// }
		// log.Printf(string(out))

	}
}
