package main

import (
	"log"
	"net/http"

	"github.com/lucas-clemente/quic-go/http3"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const certFile = "../cert/certificate.crt"
const keyFile = "../cert/private.key"

func runHttp1(portNum string) {
	handler := setHandler()
	server := &http.Server{
		Addr:    portNum,
		Handler: handler,
	}
	log.Printf("[HTTP1] Serving on http://%s", portNum)
	// return server
	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("%v", err)
	}
}

func runHttp2(portNum string) {
	handler := setHandler()

	h2s := &http2.Server{}
	h1s := &http.Server{
		Addr:    portNum,
		Handler: h2c.NewHandler(handler, h2s),
		// TLSConfig: ,
	}

	log.Printf("[HTTP2] Serving on https://%s", portNum)

	if err := h1s.ListenAndServeTLS(certFile, keyFile); err != nil {
		log.Fatalf("%v", err)
	}
}

func runHttp3(portNum string) {
	handler := setHandler()

	log.Printf("[HTTP3] Serving on https://%s", portNum)

	if err := http3.ListenAndServe(portNum, certFile, keyFile, handler); err != nil {
		log.Fatalf("%v", err)
	}

}

// func runHttp1(portNum string) error {
// 	handler := setHandler()
// 	server := &http.Server{
// 		Addr:    portNum,
// 		Handler: handler,
// 	}
// 	log.Printf("[HTTP1] Serving on http://%s", portNum)
// 	// return server
// 	return server.ListenAndServe()
// }
