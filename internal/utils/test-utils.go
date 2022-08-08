package utils

import (
	"log"
	"testing"
)

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
