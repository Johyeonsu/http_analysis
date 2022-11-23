package main

import (
	"context"
	"log"
	"net"
	"net/http"

	"github.com/lucas-clemente/quic-go/http3"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

const certFile = "../cert/certificate.crt"
const keyFile = "../cert/private.key"
const keyServerAddr = "serverAddr"

func runHttp1(portNum string) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	handler := setHandler()

	server := &http.Server{
		Addr:    portNum,
		Handler: handler,
		BaseContext: func(l net.Listener) context.Context {
			ctx = context.WithValue(ctx, keyServerAddr, l.Addr().String())
			return ctx
		},
	}

	log.Printf("[HTTP1] Serving on http://%s", portNum)

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("%v", err)
	}
	defer cancelCtx()
}

func runHttp2(portNum string) {
	ctx, cancelCtx := context.WithCancel(context.Background())
	handler := setHandler()

	h2s := &http2.Server{}
	h1s := &http.Server{
		Addr:    portNum,
		Handler: h2c.NewHandler(handler, h2s),
		BaseContext: func(l net.Listener) context.Context {
			ctx = context.WithValue(ctx, keyServerAddr, l.Addr().String())
			return ctx
		},
	}

	log.Printf("[HTTP2] Serving on https://%s", portNum)

	if err := h1s.ListenAndServeTLS(certFile, keyFile); err != nil {
		log.Fatalf("%v", err)
	}
	defer cancelCtx()
}

func runHttp3(portNum string) {
	handler := setHandler()

	log.Printf("[HTTP3] Serving on https://%s", portNum)

	if err := http3.ListenAndServeQUIC(portNum, certFile, keyFile, handler); err != nil {
		log.Fatalf("%v", err)
	}

}
