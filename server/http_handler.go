package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/handlers"
)

const testImgNum = 5

func printNow() string {
	return time.Now().Format("15:04:05.000000")
}

func setHandler() http.Handler {
	mux := http.NewServeMux()

	mux.Handle("/",
		handlers.CombinedLoggingHandler(
			os.Stdout, http.FileServer(http.Dir("public"))))

	mux.HandleFunc("/pageload", pageloadHandler)
	mux.HandleFunc("/imgpusher", imgPusherHandler)

	return mux
}

func imgPusherHandler(w http.ResponseWriter, r *http.Request) {
	for i := 1; i <= testImgNum; i++ {
		var err error
		path := "img.png"
		image, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}
		w.Header().Set("Link", fmt.Sprintf("</%s>; rel=preload; as=image", path))
		w.Header().Set("Content-Type", "image/png")
		w.Write(image)
	}
}

func pageloadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fmt.Printf("%s [%s] %s, %s, %s\n", printNow(), ctx.Value(keyServerAddr), r.Method, r.URL, r.Proto)

	pusher, ok := w.(http.Pusher)
	if ok {
		fmt.Printf("%s [%s] Push %s\n", printNow(), ctx.Value(keyServerAddr), r.URL)
		pusher.Push("/imgpusher", nil)
		fmt.Printf("%s [%s] %s\n", printNow(), ctx.Value(keyServerAddr), "Send Push")
	}

	w.Header().Add("Content-Type", "text/html")

	fmt.Fprintf(w, `<!DOCTYPE html><head><title>HTTP</title></head><body>
	<p><h1>HTTP TEST</h1>
	<h3>Loading %d icon image...</h3></p>`, testImgNum)

	for i := 1; i <= testImgNum; i++ {
		fmt.Fprintf(w, `<span><img src="img/z-768px-Sign-check-icon-%d.png" width=30 height=30></span>`, i)
	}

	fmt.Fprintf(w, "</body></html>")
	fmt.Printf("%s [%s] %s\n", printNow(), ctx.Value(keyServerAddr), "Send Response")
}
