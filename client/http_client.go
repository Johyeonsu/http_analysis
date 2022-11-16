package main

import (
	"bufio"
	"crypto/tls"
	"crypto/x509"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"time"
)

const certFile = "../cert/certificate.crt"

func CaCert() *x509.CertPool {
	caCert, err := ioutil.ReadFile(certFile)

	if err != nil {
		log.Fatalf("%+v", err)
	}
	caCertPool := x509.NewCertPool()
	caCertPool.AppendCertsFromPEM(caCert)
	return caCertPool
}

func tlsConfig(keyLog io.Writer) *tls.Config {
	return &tls.Config{
		InsecureSkipVerify: insecure,
		RootCAs:            CaCert(),
		KeyLogWriter:       keyLog,
		MinVersion:         tls.VersionTLS12,
	}
}

type DumpTransport struct {
	Transport http.RoundTripper
	Output    io.Writer
}

func newDumpTransport(t http.RoundTripper) http.RoundTripper {
	return &DumpTransport{Transport: t}
}

func (t *DumpTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	b, err := httputil.DumpRequest(req, true)
	if err != nil {
		return nil, err
	}
	log.Print(string(b))
	resp, err := t.Transport.RoundTrip(req)
	if err != nil {
		return nil, err
	}
	b, err = httputil.DumpResponse(resp, true)
	if err != nil {
		return nil, err
	}
	log.Println(len(string(b)), "bytes")
	return resp, nil
}

type bufferedWriteCloser struct {
	*bufio.Writer
	io.Closer
}

func NewBufferedWriteCloser(writer *bufio.Writer, closer io.Closer) io.WriteCloser {
	return &bufferedWriteCloser{
		Writer: writer,
		Closer: closer,
	}
}

func (h bufferedWriteCloser) Close() error {
	if err := h.Writer.Flush(); err != nil {
		return err
	}
	return h.Closer.Close()
}

func sendRequest(url *url.URL, client *http.Client, keyLog io.Writer, httpv int, testReqNum int, cn int) *http.Request {
	rn := 0
	req, err := http.NewRequest(httpMethod, url.String(), nil)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	req.Header.Add("Accept-Charset", "UTF-8;q=1, ISO-8859-1;q=0")
	req.Header.Add("Connection", "Keep-Alive")
	// req.Header.Add("Connection", "Close")

	for rn < testReqNum {
		if httpv == 3 {
			h3t := time.Now()
			_, err := client.Do(req)
			if err != nil {
				log.Fatalf("failed to read response: %+v", err)
			}
			printTimeDuration("HTTP3 TOTAL", h3t, time.Now())
		} else {
			httpTrace(url, client, req, keyLog, rn, cn)
		}
		rn++
	}

	return req
}
