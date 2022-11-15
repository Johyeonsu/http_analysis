package main

import (
	"context"
	"flag"
	"fmt"
	"log"
	"net"
	"net/http"
	"runtime"

	"github.com/lucas-clemente/quic-go/http3"
	"golang.org/x/net/http2"
	"golang.org/x/net/http2/h2c"
)

func main() {

	runtime.SetBlockProfileRate(1)

	httpv := flag.Int("http", 1, "http Version")
	addr := flag.String("addr", ":7071", "host:port")
	all := flag.Bool("all", false, "open 3 port, ignore addr option")
	flag.Parse()

	if *all {
		ctx, cancelCtx := context.WithCancel(context.Background())
		const keyServerAddr = "serverAddr"

		handler := setHandler()
		h1server := &http.Server{
			Addr:    ":7071",
			Handler: handler,
			BaseContext: func(l net.Listener) context.Context {
				ctx = context.WithValue(ctx, keyServerAddr, l.Addr().String())
				return ctx
			},
		}

		h2s := &http2.Server{}
		h2server := &http.Server{
			Addr:    ":7072",
			Handler: h2c.NewHandler(handler, h2s),
			BaseContext: func(l net.Listener) context.Context {
				ctx = context.WithValue(ctx, keyServerAddr, l.Addr().String())
				return ctx
				// TLSConfig: ,
			},
		}

		// h3server := &http3.Server{}
		// h3server.ServerContextKey = keyServerAddr

		go func() {
			err := h1server.ListenAndServe()
			if err != nil {
				fmt.Printf("%+v", err)
			}
			defer cancelCtx()
		}()

		go func() {
			err := h2server.ListenAndServeTLS(certFile, keyFile)
			if err != nil {
				fmt.Printf("%+v", err)
			}
			defer cancelCtx()
		}()

		go func() {
			err := http3.ListenAndServe(":7073", certFile, keyFile, handler)
			if err != nil {
				fmt.Printf("%+v", err)
			}
			defer cancelCtx()
		}()
		<-ctx.Done()

	} else {
		switch *httpv {
		case 1:
			runHttp1(*addr)
		case 2:
			runHttp2(*addr)
		case 3:
			runHttp3(*addr)
		default:
			log.Fatalf("Only support http Version 1, 2, 3")
		}
	}

}
