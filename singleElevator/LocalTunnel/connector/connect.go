package connector

import (
	"fmt"
	"log"
	"net"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"time"
)

/*
	This module has as funtion to connect two programs locally on a computer.
	They will be able to send and receive strings to each other,
	and will start up eachother if its 'partner' is not responding.
*/

func checkErr(err error) {
	if err != nil {
		panic(err)
	}
}

func listenLocal(port int, rxChan chan string, myPartnerIsAliveCh chan bool) {
	localAddress, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", port))
	checkErr(err)
	connection, err := net.ListenUDP("udp", localAddress)
	checkErr(err)
	buffer := make([]byte, 1024)
	var length int
	var msg string
	prevMsg := ""

	for err == nil {
		defer connection.Close()
		length, _, err = connection.ReadFromUDP(buffer)
		msg = string(buffer[:length])
		msg = strings.Replace(msg, "\x00", "", -1)
		switch msg {
		case "":
		case "I'm alive!":
			//This is a message both will send to eachother contionusly to tell they live
			myPartnerIsAliveCh <- true
		default:
			if prevMsg != msg {
				rxChan <- msg
				prevMsg = msg
			}
		}
	}
	fmt.Println("Error in local listener, ", err)
}

func sendLocal(port int, tx chan string) {
	dst, err := net.ResolveUDPAddr("udp", fmt.Sprintf("127.0.0.1:%d", port))
	checkErr(err)
	conn, err := net.DialUDP("udp", nil, dst)
	checkErr(err)
	for msg := range tx {
		buffer := make([]byte, 1024)
		copy(buffer[:], msg)
		_, err = conn.Write(buffer)
		if err != nil {
			break
		}
	}
	fmt.Println("Error in local sender, ", err)
}

func EstablishLocalTunnel(partnerName string, rxPort int, txPort int) (chan string, chan string) {
	msgChan := make(chan string)
	msgChanSendAndRepeat := make(chan string)

	go func() {
		var currentMsg string
		for {
			select {
			case a := <-msgChanSendAndRepeat:
				currentMsg = a
			case <-time.After(20 * time.Millisecond):
				msgChan <- currentMsg
			}
		}
	}()

	go sendLocal(txPort, msgChan)
	go func() {
		for {
			time.Sleep(20 * time.Millisecond)
			msgChan <- "I'm alive!"
		}
	}()

	rxChan := make(chan string)
	partnerAliveCh := make(chan bool)

	go listenLocal(rxPort, rxChan, partnerAliveCh)
	go keepAlive(partnerName, rxChan, partnerAliveCh)

	return rxChan, msgChanSendAndRepeat
}

func keepAlive(partnerName string, rxChan chan string, parnerAliveCh chan bool) {
	watchDogCnt := -2000   // starts out negative to give the partner extra time to respond initially
	hasConnection := false // is only used to give informative printouts

	go func() {
		for {
			select {
			case <-time.After(200 * time.Millisecond):
				switch {
				case watchDogCnt > 1000:
					if hasConnection {
						fmt.Printf("Connection lost with %v\n", partnerName)
						rxChan <- "Connection lost"
					}
					hasConnection = false
					go startProgram(partnerName)
					watchDogCnt = -30000
				case watchDogCnt > 200:
					fmt.Printf("Watchdog for %v, at %v ms\n", partnerName, watchDogCnt)
					fallthrough
				default:
					watchDogCnt += 200
				}
			case <-parnerAliveCh:
				watchDogCnt = 0
				if !hasConnection {
					fmt.Printf("Connection established with %v\n", partnerName)
					rxChan <- "Connection established"
					hasConnection = true
				}
			}
		}
	}()
}

func startProgram(name string) {
	fmt.Println("Starting ", name)

	cmd := exec.Command("gnome-terminal", "-x", name)
	if runtime.GOOS == "windows" {
		cmd = exec.Command("cmd.exe", "/C", "start", "go", "run", name)
		cmd.Stdout = os.Stdout //trengs disse?
		cmd.Stderr = os.Stderr //trengs disse?
	}
	if err := cmd.Run(); err != nil {
		log.Fatal(err)
	}
}
