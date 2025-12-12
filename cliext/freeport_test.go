package cliext_test

import (
	"fmt"
	"net"
	"testing"

	"github.com/temporalio/cli/cliext"
)

func TestFreePort_NoDouble(t *testing.T) {
	host := "127.0.0.1"
	portSet := make(map[int]bool)
	for i := 0; i < 2000; i++ {
		p, err := cliext.GetFreePort(host)
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
		p, err := cliext.GetFreePort(host)
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

func TestFreePort_IPv4Unspecified(t *testing.T) {
	host := "0.0.0.0"
	p, err := cliext.GetFreePort(host)
	if err != nil {
		t.Fatalf("Error: %s", err)
		return
	}
	err = tryListenAndDialOn(host, p)
	if err != nil {
		t.Fatalf("Error: %s", err)
		return
	}
}

func TestFreePort_IPv6Unspecified(t *testing.T) {
	host := "::"
	p, err := cliext.GetFreePort(host)
	if err != nil {
		t.Fatalf("Error: %s", err)
		return
	}
	err = tryListenAndDialOn(host, p)
	if err != nil {
		t.Fatalf("Error: %s", err)
		return
	}
}

// This function is used as part of unit tests, to ensure that the port
// is available for listening and dialing.
func tryListenAndDialOn(host string, port int) error {
	host = cliext.MaybeEscapeIPv6(host)
	l, err := net.Listen("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		return err
	}
	defer l.Close()

	tcpAddr, err := net.ResolveTCPAddr("tcp", fmt.Sprintf("%s:%d", host, port))
	if err != nil {
		panic(err)
	}
	r, err := net.DialTCP("tcp", nil, tcpAddr)
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
