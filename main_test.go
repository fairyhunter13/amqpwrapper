// +build integration

package rabbitmq

import (
	"flag"
	"fmt"
	"os"
	"testing"
)

const (
	uriDialTemplate = "amqp://guest:guest@%s:5672"
)

var (
	uriDial string
)

func parseHost() {
	host := "localhost"
	if envHost := os.Getenv("RABBITMQ_HOST"); envHost != "" {
		host = envHost
	}
	uriDial = fmt.Sprintf(uriDialTemplate, host)
}

func TestMain(m *testing.M) {
	flag.Parse()
	parseHost()
	os.Exit(m.Run())
}
