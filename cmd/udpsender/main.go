package main

import (
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
)

func main() {
	u, err := net.ResolveUDPAddr("udp", "localhost:42069")
	if err != nil {
		log.Fatal(err)
	}
	c, err := net.DialUDP("udp", nil, u)
	if err != nil {
		log.Fatal(err)
	}
	defer c.Close() // nolint

	b := bufio.NewReader(os.Stdin)
	for {
		fmt.Printf("> ")
		s, err := b.ReadString(byte('\n'))
		if err != nil {
			log.Fatal(err)
		}
		_, err = c.Write([]byte(s))
		if err != nil {
			log.Println(err)
		}
	}
}
