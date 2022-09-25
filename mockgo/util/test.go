package util

import (
	"bytes"
	"io"
	"log"
	"math/rand"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func RequestCall(t *testing.T, httpClient http.Client, method, url string, header map[string][]string, body string, expectedStatus int, assertBody func(responseBody string)) {
	t.Logf("calling url: '%s' ...", url)

	var bodyReader io.Reader
	if len(body) > 0 {
		bodyReader = bytes.NewBufferString(body)
	}
	request, err := http.NewRequest(method, url, bodyReader)
	assert.NoError(t, err)
	request.Header = header
	response, err := httpClient.Do(request)
	assert.NoError(t, err)
	defer response.Body.Close()
	t.Logf("response status: '%s'", response.Status)
	assert.Equal(t, expectedStatus, response.StatusCode)
	respBytes, err := io.ReadAll(response.Body)
	respBody := string(respBytes)
	assert.NoError(t, err)
	t.Logf("response body:\n '%s'", respBody)
	if assertBody != nil {
		assertBody(respBody)
	}
}

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}

func RunAndCheckCoverage(testPackage string, m *testing.M, treshold float64) int {

	code := m.Run()

	if code == 0 && testing.CoverMode() != "" {
		coverage := testing.Coverage()
		if coverage < treshold {
			log.Printf("%s tests passed, but coverage must be above %2.2f%%, but it is %2.2f%%\n", testPackage, treshold*100, coverage*100)
			code = -1
		}
	}
	return code
}
