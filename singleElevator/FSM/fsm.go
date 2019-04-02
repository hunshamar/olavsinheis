package main

import (
	"../../config"
	"../../elevatorDriver/elevStruct"
	"../../elevatorDriver/elevio"
	"../../elevatorDriver/orderFunctions"
	"../LocalTunnel/connector"
	"../LocalTunnel/localMsg"
	"flag"
	"fmt"
	"time"
)

//////	ELEVATOR FSM //////

/*
	- The elevator FSM and FSM_msgHandler are locally connected and will restart the other it crashes
*/

func main() {
	fmt.Println("Starting elevator")
	port := flag.Int("port", 15657, "port for the elevator")
	flag.Parse()

	if *port == 0 {
		*port = 15657
	}

	elevio.Init(fmt.Sprintf("localhost:%v", *port))
	e := elevStruct.ElevatorInit()
	updateStatesToMsgHandler := make(chan elevStruct.Elevator)

	newButton := make(chan elevStruct.ButtonEvent)
	newOrders := make(chan elevStruct.ButtonEvent)
	floorArrivals := make(chan int)
	updateLights := make(chan elevStruct.LightType)
	receiveLocal, msgChanLocal := connector.EstablishLocalTunnel(
		config.MSGHANDLER_PATH, config.RECEIVE_PORT_FSM, config.SEND_PORT_FSM)

	go elevio.PollButtons(newButton, newOrders)
	go elevio.PollFloorSensor(floorArrivals)
	go local_msg_transceiver(e.Duplicate(), receiveLocal, msgChanLocal, newOrders, updateStatesToMsgHandler, updateLights)
	elevator_fsm(e, newOrders, floorArrivals, updateStatesToMsgHandler, updateLights, newButton)
}

func elevator_fsm(e elevStruct.Elevator, newOrders <-chan elevStruct.ButtonEvent,
	floorArrivals <-chan int, updateStatesToMsgHandler chan elevStruct.Elevator,
	updateLights <-chan elevStruct.LightType, newButton <-chan elevStruct.ButtonEvent) {

	var doorTimer <-chan time.Time
	errorCh := make(chan string)

	prevElevator := e
	go func() {
		for err := range errorCh {
			fmt.Println("Error in FSM - ", err)
		}
	}()

	//initate the elevator at a floor
	initialFloor := elevio.GetFloor()
	if initialFloor == -1 {
		elevio.SetMotorDirection(elevStruct.MD_Down)
	} else {
		elevio.SetMotorDirection(elevStruct.MD_Stop)
	}
	e.Behaviour = elevStruct.E_Initial

	for {
	sel:
		select {
		case a := <-newButton: //when a new button is pressed at this elevator
			switch e.Behaviour {
			case elevStruct.E_Initial:
				//do nothing
			case elevStruct.E_Idle:
				if e.Floor == a.Floor { //if we receive a order at the floor the elevator is located at
					elevio.SetDoorOpenLamp(true)
					doorTimer = time.After(config.DOOR_OPEN_TIME) //activate door open timer
					e.Behaviour = elevStruct.E_DoorOpen
				} else {
					//handle internal order
					if a.Button == elevStruct.BT_Cab {
						e.Orders[a.Button][a.Floor] = true
					} else {
						e.LightMatrix[a.Button][a.Floor] = true
					}
				}

			case elevStruct.E_Moving:
				if a.Button == elevStruct.BT_Cab {
					e.Orders[a.Button][a.Floor] = true
				} else {
					e.LightMatrix[a.Button][a.Floor] = true
				}

			case elevStruct.E_DoorOpen:
				if e.Floor == a.Floor {
					doorTimer = time.After(config.DOOR_OPEN_TIME) // reset door open timer if we are at the ordered floor
				} else {
					//handle internal order
					if a.Button == elevStruct.BT_Cab {
						e.Orders[a.Button][a.Floor] = true
					} else {
						e.LightMatrix[a.Button][a.Floor] = true
					}
				}
			}
		case a := <-newOrders: //when we receive a new order from the message handler
			switch e.Behaviour {
			case elevStruct.E_Idle:
				if e.Floor == a.Floor {
					elevio.SetDoorOpenLamp(true)
					doorTimer = time.After(config.DOOR_OPEN_TIME)
					e.Behaviour = elevStruct.E_DoorOpen
				} else {
					//handle internal order
					e.Orders[a.Button][a.Floor] = true
					if a.Button == elevStruct.BT_Cab {
						elevio.SetButtonLamp(elevStruct.BT_Cab, a.Floor, true)
					}
					e.Dir = orderFunctions.ChooseDirection(e)
					elevio.SetMotorDirection(e.Dir)
					e.Behaviour = elevStruct.E_Moving
				}

			case elevStruct.E_Moving, elevStruct.E_Initial:
				e.Orders[a.Button][a.Floor] = true
				if a.Button == elevStruct.BT_Cab {
					elevio.SetButtonLamp(elevStruct.BT_Cab, a.Floor, true)
					break sel
				}

			case elevStruct.E_DoorOpen:
				if e.Floor == a.Floor {
					doorTimer = time.After(config.DOOR_OPEN_TIME)
				} else {
					//handle internal order
					e.Orders[a.Button][a.Floor] = true
					if a.Button == elevStruct.BT_Cab {
						elevio.SetButtonLamp(elevStruct.BT_Cab, a.Floor, true)
					}
				}
			}
		case a := <-floorArrivals:
			e.Floor = a
			elevio.SetFloorIndicator(e.Floor)

			switch e.Behaviour {
			case elevStruct.E_Idle:
				errorCh <- "Arrived on floor when idle!"
				fallthrough

			case elevStruct.E_Moving, elevStruct.E_Initial:
				if orderFunctions.ShouldStop(e) {
					elevio.SetMotorDirection(elevStruct.MD_Stop)
					elevio.SetDoorOpenLamp(true)
					e = orderFunctions.ClearOrdersAtCurrentFloor(e)
					e = orderFunctions.ClearLightsAtCurrentFloor(e)
					doorTimer = time.After(config.DOOR_OPEN_TIME)
					e.Behaviour = elevStruct.E_DoorOpen
				}

			case elevStruct.E_DoorOpen:
				errorCh <- "Arrived on floor with doors open!"
			}

		case <-doorTimer: //when the door timer is done
			switch e.Behaviour {
			case elevStruct.E_Initial:
				//do nothing

			case elevStruct.E_Idle:
				//do nothing

			case elevStruct.E_Moving:
				errorCh <- "Door timer when moving!"

			case elevStruct.E_DoorOpen:
				e.Dir = orderFunctions.ChooseDirection(e)
				elevio.SetDoorOpenLamp(false)
				elevio.SetMotorDirection(e.Dir)
				if e.Dir == elevStruct.MD_Stop {
					e.Behaviour = elevStruct.E_Idle
				} else {
					e.Behaviour = elevStruct.E_Moving
				}
			}

		case a := <-updateLights:
			e.LightMatrix = a
			orderFunctions.UpdateLights(e)

		case <-time.After(20 * time.Millisecond):
			if prevElevator != e {
				updateStatesToMsgHandler <- e.Duplicate() // This will in turn update the FSM_msgHandler
				prevElevator = e.Duplicate() //save the previous elevator
			}
		}
	}
}

func local_msg_transceiver(elev elevStruct.Elevator, receive <-chan string, msgChan chan<- string,
	newOrders chan elevStruct.ButtonEvent, updateStatesToMsgHandler chan elevStruct.Elevator, updateLights chan<- elevStruct.LightType) {
	msgToHandler := localMsg.ElevatorMsg{ElevStruct: elev, Number: 0} //initialize a message for sending to message handler
	msgFromHandler := localMsg.ElevatorMsg{}                          //initialize an empty message we fill with information from message handler
	for {
		select {
		case e := <-updateStatesToMsgHandler: //send the updated states to the message handler
			msgToHandler.ElevStruct = e
			msgToHandler.Number++
			msgChan <- localMsg.EncodeElevatorMsg(msgToHandler)

		case a := <-receive: //receive updated order matrix from master
			msg := localMsg.DecodeElevatorMsg(a)
			newOrderButtons, newOrderFloors := msgToHandler.ElevStruct.Differences(msg.ElevStruct)

			for i, button := range newOrderButtons {
				newOrders <- elevStruct.ButtonEvent{
					Floor:  newOrderFloors[i],
					Button: elevStruct.ButtonType(button),
				}
			}
			if msg.Number > msgToHandler.Number {
				msgToHandler.Number = msg.Number
			}
			updateLights <- msg.ElevStruct.LightMatrix //update light matrix according to master
			msgFromHandler = msg

		case <-time.After(1000 * time.Millisecond):
			updateLights <- msgFromHandler.ElevStruct.LightMatrix //update light matrix every tenth of a second
		}
	}
}
