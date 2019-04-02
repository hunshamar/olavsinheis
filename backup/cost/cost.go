package cost

import (
	"../../network/msgStruct"
	"../../elevatorDriver/orderFunctions"
	"../../elevatorDriver/elevStruct"
	"../../config"
	"fmt"
)


/*
Cost module which calculates the optimal elevator for an order and distributes the orders 
*/

//constants for doing the calculations in TimeToServeRequest
//not to be used anywhere else
const TRAVEL_TIME = 4
const DOOR_OPEN_TIME = 5


type ElevatorStatus struct {
	SingleElevator elevStruct.Elevator
	CostValue      int
}

type MasterMap struct {
	Elevators  map[string]ElevatorStatus
	LightsHall elevStruct.LightType
}

func (mastMap *MasterMap) ConstructMessage(msgNumber int) (msg msgStruct.MsgFromMaster) {
	msg = msgStruct.MsgFromMaster{Orders: make(map[string]elevStruct.OrderType)}
	for elevID, value := range mastMap.Elevators {
		msg.Orders[elevID] = value.SingleElevator.Orders
	}
	msg.LightsHall = mastMap.LightsHall
	return msg
}

func TimeToServeRequest(e_old elevStruct.Elevator, button elevStruct.ButtonType, floor int) int {
	e := e_old
	e.Orders[button][floor] = true
	duration := 0
	//"simulates" how long an elevator will use to execute an order
	switch e.Behaviour {
	case elevStruct.E_Idle:
		e.Dir = orderFunctions.ChooseDirection(e)
		if e.Dir == elevStruct.MD_Stop {
			return duration
		}
		break
	case elevStruct.E_Moving:
		duration += TRAVEL_TIME / 2
		e.Floor += int(e.Dir)
		break
	case elevStruct.E_DoorOpen:
		duration -= DOOR_OPEN_TIME / 2
	}

	for {
		if orderFunctions.ShouldStop(e) == true {
			e = ClearOrdersAtCurrentFloorCost(e)
			if e.Floor == floor {
				return duration
			}
			duration += DOOR_OPEN_TIME
			e.Dir = orderFunctions.ChooseDirection(e)
		}
		e.Floor += int(e.Dir)
		duration += TRAVEL_TIME
	}
}

//simulated clear order function to be sure to not mess with actual elevator states
func ClearOrdersAtCurrentFloorCost(e elevStruct.Elevator) elevStruct.Elevator {
	e2 := e.Duplicate()
	for btn := 0; btn < 3; btn++ {
		if e2.Orders[btn][e2.Floor] != false {
			e2.Orders[btn][e2.Floor] = false
		}
	}
	return e2
}

func (mastermap *MasterMap) ChooseElevator() {
	minimum := 0
	i := 0 
	var elevator_id string
	var elevator_status_chosen ElevatorStatus
	//iterate through each button and floor
	for button := 0; button < 2; button++ {
		for floor := 0; floor < config.NUMFLOORS; floor++ {
			i = 0 //set to zero every time we iterate through elevators for a new elevator
			if mastermap.LightsHall[button][floor] == true { //check if an order exist in this floor
				for elevator_ID, elevator_status := range mastermap.Elevators {
					//calculate the fastest elevator for the order by iterating through elevators
					temp := TimeToServeRequest(elevator_status.SingleElevator, elevStruct.ButtonType(button), floor)
					temp = temp + elevator_status.CostValue
					if elevator_status.SingleElevator.Orders[button][floor] == true {
						temp -= 70 // if it already is in the order list, it should be prioritized
					}
					
					if i == 0 || temp < minimum { 
						//if i==0 it is the first time iterating and the minimum is set equal to temp  
						minimum = temp
						elevator_id = elevator_ID
						elevator_status_chosen = elevator_status
					}
					i++
					if elevator_status_chosen.SingleElevator.Floor == floor {
						mastermap.LightsHall[button][floor] = false
					}
				}
				if mastermap.Elevators != nil {
					elev := mastermap.Elevators[elevator_id]
					if elev.SingleElevator.Orders[button][floor] == true {
						elev.CostValue += 6
					} else {
						elev.SingleElevator.Orders[button][floor] = true
					}
					mastermap.Elevators[elevator_id] = elev
				}
			}
		}
	}
}


func (mastermap *MasterMap) UpdateLightMatrix(id string) {
	var elevator elevStruct.Elevator
	for _, elevator_status := range mastermap.Elevators {
		elevator = elevator_status.SingleElevator
		if elevator.Behaviour == elevStruct.E_DoorOpen {
			//if door is open in a floor we can turn of the lights in the corresponding floor
			mastermap.LightsHall[0][elevator.Floor] = false
			mastermap.LightsHall[1][elevator.Floor] = false
			elevator_status.CostValue = 0
			fmt.Println("Lights set to false in floor: ", elevator.Floor)
		}
	}
	for button := 0; button < 2; button++ {
		for floor := 0; floor < config.NUMFLOORS; floor++ {
			if mastermap.Elevators[id].SingleElevator.LightMatrix[button][floor] == true && mastermap.LightsHall[button][floor] == false {
				//if we receive a button update from an elevator we update the master light matrix 
				//this is in order to have all the hall lights show the same
				mastermap.LightsHall[button][floor] = true
			}
		}
	}
}
