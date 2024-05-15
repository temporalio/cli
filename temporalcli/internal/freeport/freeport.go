package freeport

import "fmt"

func MustGetFreePort(host string) int {
	port, err := GetFreePort(host)
	if err != nil {
		panic(fmt.Errorf("failed assigning ephemeral port: %w", err))
	}
	return port
}
