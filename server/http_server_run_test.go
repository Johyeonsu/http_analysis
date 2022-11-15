package main

import (
	"io/ioutil"
	"net/http"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestServerFunc(t *testing.T) {
	assert := assert.New(t)

	t.Run("http1 test", func(t *testing.T) {
		var wg sync.WaitGroup
		testAddr := "192.168.56.101:5000"
		go func() {
			defer wg.Done()
			runHttp1(testAddr)
		}()
		wg.Wait()

		res, _ := http.Get(testAddr)

		out, _ := ioutil.ReadAll(res.Body)
		defer res.Body.Close()
		result := fileCompare(string(out), indexFilePath)

		assert.Equal(true, result)

	})

}
