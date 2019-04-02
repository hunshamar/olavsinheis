package main

import (
	"../config"
	"../network/bcast"
	"../network/idGenerator"
	"../network/msgStruct"
	"../network/peers"
	"./cost"
	"flag"
	"fmt"
	"strings"
	"time"
)

////// BACKUP //////

/*
	module for backup and master. If the backup is set to be a master,
	it sends its states in addition to storing them
*/

func main() {
	time.Sleep(time.Millisecond)
	backupMap := make(map[string]cost.ElevatorStatus)
	backupStates := cost.MasterMap{Elevators: backupMap} // storing backup for all of the elevators currently connected
	masterMsg := msgStruct.MsgFromMaster{Number: 0}      // initialize the message to be sent if the backup is a master
	var myName string
	flag.StringVar(&myName, "name", "", "id of this peer")
	flag.Parse()

	if myName == "" {
		myName = idGenerator.GetRandomID()
	}

	id := "Backup - " + myName // start as backup to master

	peerUpdateCh := make(chan peers.PeerUpdate)
	peerTxEnable := make(chan bool)
	terminateTransmitter := make(chan bool) // a true value will be used to terminate the peer transmission
	go peers.Transmitter(config.PEER_PORT, id, peerTxEnable, terminateTransmitter)
	go peers.Receiver(config.PEER_PORT, peerUpdateCh)

	sendMsgChan := make(chan msgStruct.MsgFromMaster)
	repeatTx := make(chan msgStruct.MsgFromMaster)
	receiveMsgChan := make(chan msgStruct.MsgFromElevator)

	prevMsgMap := make(map[string]msgStruct.MsgFromElevator)

	go bcast.Transmitter(config.BROAD_PORT_MASTER_TX, sendMsgChan)

	go func() {
		var currMsg msgStruct.MsgFromMaster
		for {
			select {
			case msg := <-repeatTx:
				currMsg = msg
			case <-time.After(config.SEND_INTERVAL):
				if strings.HasPrefix(id, "MASTER") {
					sendMsgChan <- currMsg
				}
			}
		}
	}()

	go bcast.Receiver(config.BROAD_PORT_MASTER_RX, receiveMsgChan)

	for {
		select {
		/*
			We assume only one set of nodes with more than one connection
			on the network can exist at a time.
		*/
		case p := <-peerUpdateCh:
			var nonElevatorPeers []string
			for _, peer := range p.Peers {
				if strings.HasPrefix(peer, "Backup") || strings.HasPrefix(peer, "MASTER") {
					nonElevatorPeers = append(nonElevatorPeers, peer)
				}
			}

		sw: // useful to be able to break the switch
			switch {
			case len(nonElevatorPeers) <= 1:
				if strings.HasPrefix(id, "MASTER") {
					id = "Backup - " + myName
					terminateTransmitter <- true // terminate to give new name on network
					go peers.Transmitter(config.PEER_PORT, id, peerTxEnable, terminateTransmitter)
				}
			case true: //will only apply if it is not alone on network
				for _, name := range nonElevatorPeers {
					if strings.HasPrefix(name, "MASTER") {
						break sw // if there already is a master everything is set
					}
				}

				idOfMasterToBe := ""
				var tempCode string
				for _, name := range nonElevatorPeers {
					tempCode = strings.Replace(name, "Backup - ", "", -1)
					if idOfMasterToBe < tempCode {
						idOfMasterToBe = tempCode // this is to designate one and only one master
					}
				}
				if strings.HasSuffix(id, idOfMasterToBe) {
					// if the program was chosen master
					id = "MASTER - " + myName
					terminateTransmitter <- true // terminate to give new name on network
					go peers.Transmitter(config.PEER_PORT, id, peerTxEnable, terminateTransmitter)
				}
			}

			if strings.HasPrefix(p.New, "Elev") {
				prevMsgMap[p.New] = msgStruct.MsgFromElevator{}
			}

			for _, name := range p.Lost {
				if strings.HasPrefix(name, "Elev") {
					// if the connection to an elevator is lost its costValue is set very high
					tempElevatorStatus := backupStates.Elevators[name]
					tempElevatorStatus.CostValue = 1000
					backupStates.Elevators[name] = tempElevatorStatus
				}
			}

		case a := <-receiveMsgChan:
			if a.ElevId == "" {
				break
			}
			if prevMsgMap[a.ElevId] != a && (prevMsgMap[a.ElevId].Number <= a.Number) { // will ignore old received messages
				backupStates.Elevators[a.ElevId] = cost.ElevatorStatus{SingleElevator: a.States}
				backupStates.UpdateLightMatrix(a.ElevId)
				if !strings.HasPrefix(id, "Backup") {
					backupStates.ChooseElevator()
					masterMsg = backupStates.ConstructMessage(masterMsg.Number)
					masterMsg.Number = a.Number
					repeatTx <- masterMsg
				}
				prevMsgMap[a.ElevId] = a
			}

		case <-time.After(1000 * time.Millisecond):
			// primarly used to be able to assign a order to a new elevator

			if !strings.HasPrefix(id, "Backup") {
				backupStates.ChooseElevator()
				masterMsg = backupStates.ConstructMessage(masterMsg.Number)
				repeatTx <- masterMsg
			}
			//// Print status ////

			fmt.Printf("\nI'm: %v\n", id)
			fmt.Printf("PrevMSG\n")
			for id, val := range prevMsgMap {
				fmt.Printf("ID: %v, \t", id)
				val.States.PrintOrders()
				fmt.Printf("\n")
			}
			for name, elevINFO := range backupMap {
				fmt.Printf("NAME: %v, cost val: %v\n", name, elevINFO.CostValue)
			}

			fmt.Println("Sending: %v", masterMsg)

			//////////////////////
		}
	}
}
