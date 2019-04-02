package msgStruct

import (
	"../../elevatorDriver/elevStruct"
)

type MsgFromMaster struct {
	Orders     map[string]elevStruct.OrderType
	LightsHall elevStruct.LightType
	Number     int
}

type MsgFromElevator struct {
	ElevId string
	States elevStruct.Elevator
	Number int
}

