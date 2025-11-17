package main

import (
	"fmt"
	"log"
	"net"

	"github.com/austin-weeks/http-from-scratch/internal/request"
)

func main() {
	n, err := net.Listen("tcp", "localhost:42069")
	if err != nil {
		log.Fatal(err)
	}
	for {
		c, err := n.Accept()
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("A connection has been accepted")
		r, err := request.RequestFromReader(c)
		if err != nil {
			log.Fatal(err)
		}
		fmt.Println("Request line:")
		fmt.Printf("- Method: %s\n", r.RequestLine.Method)
		fmt.Printf("- Target: %s\n", r.RequestLine.RequestTarget)
		fmt.Printf("- Version: %s\n", r.RequestLine.HttpVersion)

		fmt.Println("Headers:")
		for name, value := range r.Headers {
			fmt.Printf("- %s: %s\n", name, value)
		}

		fmt.Println("Body:")
		fmt.Println(string(r.Body))
	}
}
