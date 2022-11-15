package main

import (
	"bytes"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"time"
)

const defaultPath = "./public/"
const defaultFile = "video.mp4"
const defaultFilePath = defaultPath + defaultFile
const indexFilePath = defaultPath + "index.html"
const uploadPath = defaultPath + "upload/"
const downloadPath = "/home/hyeonsu/Downloads/"
const testImgNum = 1

const keyServerAddr = "serverAddr"

func printNow() string {
	return time.Now().Format("15:04:05.000000")
}

func setHandler() http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/", http.FileServer(http.Dir("public")))
	mux.HandleFunc("/imgload", imgloadHandler)

	// mux.HandleFunc("/", imgloadHandler)

	mux.HandleFunc("/imgpusher", imgPusherHandler)
	mux.HandleFunc("/upload", uploadHandler)
	mux.HandleFunc("/download", downloadHandler)

	return mux
}

func imgPusherHandler(w http.ResponseWriter, r *http.Request) {
	for i := 1; i <= testImgNum; i++ {
		var err error
		path := fmt.Sprintf("%s%d%s", "img/z-768px-Sign-check-icon-", i, ".png")
		image, err := ioutil.ReadFile(path)
		if err != nil {
			panic(err)
		}
		w.Header().Set("Link", fmt.Sprintf("</%s>; rel=preload; as=image", path))
		w.Header().Set("Content-Type", "image/png")
		w.Write(image)
	}
}

func imgloadHandler(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	fmt.Printf("%s [%s] %s, %s, %s\n", printNow(), ctx.Value(keyServerAddr), r.Method, r.URL, r.Proto)

//	pusher, ok := w.(http.Pusher)
//	if ok {
//		fmt.Printf("%s [%s] Push %s\n", printNow(), ctx.Value(keyServerAddr), r.URL)
//		pusher.Push("/imgpusher", nil)
//		fmt.Printf("%s [%s] %s\n", printNow(), ctx.Value(keyServerAddr), "Send Push")
//	}

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

func uploadHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("%s, %s, %s, %s", printNow(), r.Method, r.URL, r.Proto)

	if r.Method == http.MethodPost {
		uploadFile, header, err := r.FormFile("upload_file")
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			fmt.Fprint(w, err)
			return
		}

		filepath := fmt.Sprintf("%s%s", uploadPath, header.Filename)
		file, err := os.Create(filepath)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			fmt.Fprint(w, err)
			return
		}

		defer file.Close()
		io.Copy(file, uploadFile)
		fmt.Fprintln(w, filepath)

		result := fileCompare(defaultFilePath, uploadPath+defaultFile)
		if !result {
			w.WriteHeader(http.StatusNotFound)
		} else {
			w.WriteHeader(http.StatusAccepted)
		}
	}

	if r.Method == http.MethodGet {
		io.WriteString(w, `<html><body><form action="/upload" method="post" enctype="multipart/form-data">
					<input type="file" name="upload_file"><br>
					<input type="submit">
					</form></body></html>`)
	}
}

func downloadHandler(w http.ResponseWriter, r *http.Request) {
	result := fileCompare(defaultFilePath, downloadPath+defaultFile)

	if !result {
		w.WriteHeader(http.StatusNotFound)
	} else {
		w.WriteHeader(http.StatusOK)
	}
	fmt.Fprintln(w, r)
}

func fileCompare(ori string, new string) bool {
	oriBytes, err := ioutil.ReadFile(ori)
	if err != nil {
		panic(err)
	}
	newBytes, err2 := ioutil.ReadFile(new)
	if err2 != nil {
		panic(err2)
	}

	result := bytes.Equal(oriBytes, newBytes)
	return result
}

func fileExist(fn string) bool {
	_, err := os.Stat(fn)
	var result bool = true

	if err != nil {
		result = false
		log.Printf("%+v", err)
	}
	return result
}
