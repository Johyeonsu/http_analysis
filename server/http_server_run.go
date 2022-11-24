package main

import (
	"context"
	"crypto/tls"
	"log"
	"net"
	"net/http"

	"github.com/lucas-clemente/quic-go"
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

	var err error
	log.Printf("[HTTP3] Serving on https://%s", portNum)
	certs := make([]tls.Certificate, 1)
	certs[0], err = tls.LoadX509KeyPair(certFile, keyFile)
	if err != nil {
		log.Fatalf("%+v", err)
	}

	quicConf := &quic.Config{}
	tlsConf := &tls.Config{
		Certificates: certs,
	}

	// quicConf.Tracer = qlog.NewTracer(func(_ logging.Perspective, connID []byte) io.WriteCloser {
	// 	filename := fmt.Sprintf("server_%x.qlog", connID)
	// 	f, err := os.Create(filename)
	// 	if err != nil {
	// 		log.Fatal(err)
	// 	}
	// 	log.Printf("Creating qlog file %s.\n", filename)
	// 	return utils.NewBufferedWriteCloser(bufio.NewWriter(f), f)
	// })

	tlsConf = http3.ConfigureTLSConfig(tlsConf)
	server := http3.Server{
		Handler:    handler,
		Addr:       portNum,
		QuicConfig: quicConf,
		TLSConfig:  tlsConf,
	}

	if err := server.ListenAndServe(); err != nil {
		// if err := http3.ListenAndServeQUIC(portNum, certFile, keyFile, handler); err != nil {
		log.Fatalf("%v", err)
	}

}
