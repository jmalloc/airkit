package main

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

func main() {
	var m sync.Mutex
	var g sync.WaitGroup

	for i := 1; i < 256; i++ {
		i := i // capture loop variable

		g.Add(1)
		go func() {
			defer g.Done()
			addr := fmt.Sprintf("10.0.100.%d:2025", i)
			conn, err := net.DialTimeout("tcp", addr, 1*time.Second)

			m.Lock()
			defer m.Unlock()

			if err != nil {
				// if x, ok := err.(net.Error); ok && x.Timeout() {
				// return
				// }
				fmt.Print(".")
				// fmt.Printf("%s: %T %s\n", addr, err, err)
			} else {
				conn.Close()
				fmt.Println("")
				fmt.Printf("%s: OK\n", addr)
				os.Exit(0)
			}
		}()
	}

	g.Wait()
}
