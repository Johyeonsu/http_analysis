package main

import (
	"flag"
	"log"
	"runtime"
	"sync"
)

func main() {

	runtime.SetBlockProfileRate(1)

	httpv := flag.Int("http", 1, "http Version")
	addr := flag.String("addr", ":7071", "host:port")
	all := flag.Bool("all", false, "open 3 port, ignore addr option")
	flag.Parse()

	if *all {
		var wg sync.WaitGroup
		wg.Add(3)

		go func() {
			runHttp1(":7071")
			wg.Done()
		}()

		go func() {
			runHttp2(":7072")
			wg.Done()
		}()

		go func() {
			runHttp3(":7073")
			wg.Done()
		}()
		wg.Wait()
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
