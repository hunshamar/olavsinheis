package main

import (
	"../../config"
	"../../network/bcast"
	"../../network/localip"
	"../../network/msgStruct"
	"../../network/peers"
	"../../elevatorDriver/elevStruct"
	"../LocalTunnel/connector"
	"../LocalTunnel/localMsg"
	"flag"
	"fmt"
	"reflect"
	"strings"
	"time"
)

///////	MSG HANDLER //////

type relationship int

const (
	Crashed      relationship = 2
	Connected                 = 1
	Disconnected              = 0
)

type pendingType int

const (
	ToFSM_Not_Connected pendingType = 2
	ToFSM_Connected                 = 1
	ToMaster                        = 0
)

type MsgHandler struct {
	MyElevStates elevStruct.Elevator
	MsgToElev    localMsg.ElevatorMsg
	MsgFromElev  localMsg.ElevatorMsg

	MsgToMaster   msgStruct.MsgFromElevator
	MsgFromMaster msgStruct.MsgFromMaster

	RelationElevator relationship
	RelationMaster   relationship
}

type MsgFromHandlerToHandler struct {
	Id     string
	States elevStruct.Elevator
	Number int
}

func main() {
	H := MsgHandler{MsgFromElev: localMsg.ElevatorMsg{Number: 0},
		RelationElevator: Disconnected, RelationMaster: Disconnected}
	pendingUpdates := make(chan pendingType)
	var myName string

	flag.StringVar(&myName, "name", "", "id of this peer")
	flag.Parse()

	if myName == "" {
		myName, _ = localip.LocalIP()
	}

	id := "Elev - " + myName

	H.MsgToMaster.ElevId = id
	H.MsgToMaster.Number = 0

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)

	go peers.Transmitter(config.PEER_PORT, id, peerTxEnable, peerTxEnable)
	go peers.Receiver(config.PEER_PORT, peerUpdateCh)

	netTx := make(chan msgStruct.MsgFromElevator)
	netRx := make(chan msgStruct.MsgFromMaster)

	var prevMsg msgStruct.MsgFromMaster
	repeatTx := make(chan msgStruct.MsgFromElevator)

	go bcast.Transmitter(config.BROAD_PORT_HANDLER_TX, netTx)
	go bcast.Receiver(config.BROAD_PORT_HANDLER_RX, netRx)

	go func() {
		var currMsg msgStruct.MsgFromElevator
		for {
			select {
			case msg := <-repeatTx:
				currMsg = msg
			case <-time.After(config.SEND_INTERVAL):
				netTx <- currMsg
			}
		}
	}()

	receiveLocal, msgChanLocal := connector.EstablishLocalTunnel(
		config.FSM_PATH, config.RECEIVE_PORT_HANDLER, config.SEND_PORT_HANDLER)

	go func() {
		for update := range pendingUpdates {
			switch update {
			case ToFSM_Not_Connected:
				msgChanLocal <- localMsg.EncodeElevatorMsg(H.MsgToElev)
			case ToFSM_Connected:
				msgChanLocal <- localMsg.EncodeElevatorMsg(localMsg.ElevatorMsg{ElevStruct: H.MsgToElev.ElevStruct})
			case ToMaster:
				repeatTx <- H.MsgToMaster
			}
		}
	}()

	for {
		select {
		case p := <-peerUpdateCh:
			for _, name := range p.Lost {
				if strings.HasPrefix(name, "MASTER") {
					fmt.Println("LOST CONNECTION TO MASTER")
					H.RelationMaster = Disconnected
				}
			}
			if strings.HasPrefix(p.New, "MASTER") {
				fmt.Printf("Connected to master \n")
				H.RelationMaster = Connected
			}

		case a := <-netRx:
			if !reflect.DeepEqual(a, prevMsg) {

				H.MsgFromMaster = a

				//update lightmatrix of elevator
				if H.MsgFromMaster.Number > H.MsgToMaster.Number {
					H.MsgToMaster.Number = H.MsgFromMaster.Number
				}

				NewLightMatrix := a.LightsHall
				H.MsgToElev.ElevStruct.LightMatrix = NewLightMatrix
				NewOrders := a.Orders[id]
				H.MsgToElev.ElevStruct.Orders = NewOrders
				pendingUpdates <- ToFSM_Connected
				prevMsg = a
			}

		case a := <-receiveLocal:
			switch a {
			case "Connection lost":
				H.MsgToElev = H.MsgFromElev // if the FSM crash the latest info is handy
				H.RelationElevator = Crashed
			case "Connection established":
				if H.RelationElevator == Disconnected { // I.e. not crashed and thus a freshly new connection
					H.RelationElevator = Connected
					switch H.RelationMaster {
					case Disconnected:
						pendingUpdates <- ToFSM_Not_Connected
					case Connected:
						pendingUpdates <- ToFSM_Connected
					}
				}
			default:
				msg := localMsg.DecodeElevatorMsg(a)
				if msg.Number > H.MsgFromElev.Number { // check if it is new info
					H.MsgFromElev = msg
					if H.RelationMaster == Connected {
						H.MsgToMaster.States = msg.ElevStruct
						H.MsgToMaster.Number++

						pendingUpdates <- ToMaster
					} else {
						H.MsgToElev = H.MsgFromElev
						H.MsgToElev.ElevStruct.Orders[0] = H.MsgToElev.ElevStruct.LightMatrix[0]
						H.MsgToElev.ElevStruct.Orders[1] = H.MsgToElev.ElevStruct.LightMatrix[1]
						pendingUpdates <- ToFSM_Not_Connected
					}
					H.MyElevStates = msg.ElevStruct

				} else { // old info imply a crash has occured
					H.RelationElevator = Crashed
					pendingUpdates <- ToFSM_Not_Connected
				}
				H.RelationElevator = Connected
			}
		}
	}
}
