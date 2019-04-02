package main

import (
	"../connector"
	"fmt"
	"time"
)

func main() {


	fmt.Println("Program start! 'prg2.go'")

	receive, msgChan := connector.EstablishLocalTunnel("prg1.go", 55555, 44444)

	_ = msgChan //For å sende en beskjed legg streng i kanalen

	go func() {
		for {
			time.Sleep(time.Second)
			msgChan <- "hei"
		}
	}()
	for {
		select {
		case a := <-receive:
			fmt.Printf("Received:\t %v\n", a)
		}
	}
}
