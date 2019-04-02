package orderFunctions

import (
	."../elevStruct"
	"../elevio"
	"../../config"
)
func ordersAbove(e Elevator) bool {
	for floor := e.Floor + 1; floor < config.NUMFLOORS; floor++ {
		for btn := 0; btn < 3; btn++ {
			if e.Orders[btn][floor] {
				return true
			}
		}
	}
	return false
}
func ordersBelow(e Elevator) bool {
	for floor := 0; floor < e.Floor; floor++ {
		for btn := 0; btn < 3; btn++ {
			if e.Orders[btn][floor] {
				return true
			}
		}
	}
	return false
}
func ChooseDirection(e Elevator) MotorDirection {
	switch e.Dir {
	case MD_Down:
		if ordersBelow(e) {
			return MD_Down
		} else if ordersAbove(e) {
			return MD_Up
		} else {
			return MD_Stop
		}
	case MD_Up:
		if ordersAbove(e) {
			return MD_Up
		} else if ordersBelow(e) {
			return MD_Down
		} else {
			return MD_Stop
		}
	case MD_Stop:
		if ordersBelow(e) {
			return MD_Down
		} else if ordersAbove(e) {
			return MD_Up
		} else {
			return MD_Stop
		}
	default:
		return MD_Stop
	}

}

func ShouldStop(e Elevator) bool {
	switch e.Dir {
	case MD_Down:
		return e.Orders[BT_HallDown][e.Floor] ||
			e.Orders[BT_Cab][e.Floor] ||
			!ordersBelow(e)
	case MD_Up:
		return e.Orders[BT_HallUp][e.Floor] ||
			e.Orders[BT_Cab][e.Floor] ||
			!ordersAbove(e)
	case MD_Stop:
		fallthrough
	default:
		return true
	}
}

func ClearOrdersAtCurrentFloor(e Elevator) Elevator {
	e2 := e.Duplicate()
	for btn := 0; btn < 3; btn++ {
		if e2.Orders[btn][e2.Floor] {
			e2.Orders[btn][e2.Floor] = false
		}
	}
	return e2
}
func ClearLightsAtCurrentFloor(e Elevator) Elevator {
	e2 := e.Duplicate()
	for btn := 0; btn < 2; btn++ {
		if e2.LightMatrix[btn][e2.Floor] {
			e2.LightMatrix[btn][e2.Floor] = false
		}
	}
	elevio.SetButtonLamp(BT_Cab, e.Floor, false)
	return e2
}

func UpdateLights(e Elevator) {
	for floor := 0; floor < config.NUMFLOORS; floor++ {
		for btn := 0; btn < 2; btn++ {
			if e.LightMatrix[btn][floor] {
				elevio.SetButtonLamp(ButtonType(btn), floor, true)
			} else {
				elevio.SetButtonLamp(ButtonType(btn), floor, false)
			}
		}
		if e.Orders[int(BT_Cab)][floor] {
			elevio.SetButtonLamp(BT_Cab, floor, true)
		} else {
			elevio.SetButtonLamp(BT_Cab, floor, false)
		}
	}
}
