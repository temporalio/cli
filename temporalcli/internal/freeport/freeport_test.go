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
	defer c.Close()
	if err != nil {
		panic(err)
	}

	return nil
}

// func main() {
// 	if (len(os.Args) > 1) {
// 		port, err := strconv.Atoi(os.Args[1])
// 		if err != nil {
// 			panic(err)
// 		}
// 		tryListenAndDialOn("127.0.0.1", port)
// 		return
// 	}

//     portSet := make(map[int]bool)

//     for i := 0; i < 5000; i++ {
//         p, err := GetFreePort("127.0.0.1")
// 		if err != nil {
// 			fmt.Printf("Error: %s\n", err)
// 			continue
// 		}
// 		fmt.Printf("... %d\n", p)

//         if _, exists := portSet[p]; exists {
//             fmt.Printf("Port %d has been assigned more than once\n", p)
//         }

// 		tryListenAndDialOn(p)

// 		// cmd := exec.Command(os.Args[0], strconv.Itoa(p))
// 		// cmd.Stdout = os.Stdout
// 		// cmd.Stderr = os.Stderr
// 		// err = cmd.Start()
// 		// if err != nil {
// 		// 	panic(err)
// 		// }
// 		// err = cmd.Wait()
// 		// if err != nil {
// 		// 	panic(err)
// 		// }

//         // Add port to the set
//         portSet[p] = true
//     }
// }

// func main() {
// 	if (len(os.Args) > 1) {
// 		port, err := strconv.Atoi(os.Args[1])
// 		if err != nil {
// 			panic(err)
// 		}
// 		listenAndDial(port)
// 		return
// 	}

//     portSet := make(map[int]bool)

//     for i := 0; i < 5000; i++ {
//         p, err := GetFreePort()
// 		if err != nil {
// 			fmt.Printf("Error: %s\n", err)
// 			continue
// 		}
// 		fmt.Printf("... %d\n", p)

//         if _, exists := portSet[p]; exists {
//             fmt.Printf("Port %d has been assigned more than once\n", p)
//         }

// 		// Test 2
// 		listenAndDial(p)

// 		// Test 3
// 		// cmd := exec.Command(os.Args[0], strconv.Itoa(p))
// 		// cmd.Stdout = os.Stdout
// 		// cmd.Stderr = os.Stderr
// 		// err = cmd.Start()
// 		// if err != nil {
// 		// 	panic(err)
// 		// }
// 		// err = cmd.Wait()
// 		// if err != nil {
// 		// 	panic(err)
// 		// }

//         // Add port to the set
//         portSet[p] = true
//     }
// }
