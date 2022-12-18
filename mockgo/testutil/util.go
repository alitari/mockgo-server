package testutil

import (
	"log"
	"math/rand"
	"testing"
	"time"
)

// RunAndCheckCoverage use this in order to let the test fail when a coverage is not reached
func RunAndCheckCoverage(testPackage string, m *testing.M, treshold float64) int {
	rand.Seed(time.Now().UnixNano())
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

const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

// RandString create a random string with n letters
func RandString(n int) string {
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Int63()%int64(len(letterBytes))]
	}
	return string(b)
}
