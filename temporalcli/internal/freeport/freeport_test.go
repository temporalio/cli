package freeport_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/temporalio/cli/temporalcli/internal/freeport"
)

func TestFreePort_NoDouble(t *testing.T) {
	host := "127.0.0.1"
	portSet := make(map[int]bool)

	for i := 0; i < 2000; i++ {
		p, err := freeport.GetFreePort(host)
		if err != nil {
			t.Fatalf("Error: %s", err)
			break
		}

		if _, exists := portSet[p]; exists {
			t.Fatalf("Port %d has been assigned more than once", p)
		}

		// Add port to the set
		portSet[p] = true
	}
}

func TestFreePort_CanBindImmediatelySameProcess(t *testing.T) {
	host := "127.0.0.1"

	for i := 0; i < 500; i++ {
		p, err := freeport.GetFreePort(host)
		if err != nil {
			t.Fatalf("Error: %s", err)
			break
		}
		err = tryListenAndDialOn(host, p)
		if err != nil {
			t.Fatalf("Error: %s", err)
			break
		}
	}
}

// This function is used as part of unit tests, to ensure that the port
func tryListenAndDialOn(host string, port int) error {
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}
	defer l.Close()

	r, err := net.DialTCP("tcp", nil, l.Addr().(*net.TCPAddr))
	if err != nil {
		panic(err)
	}
	defer r.Close()

	c, err := l.Accept()
	if err != nil {
		panic(err)
	}
	defer c.Close()

	return nil
}
