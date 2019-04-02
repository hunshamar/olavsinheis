package localMsg

import (
	"../../../elevatorDriver/elevStruct"
	"encoding/json"
)

// A simple module to 'stringify' the message sent locally using the tunnel module
type ElevatorMsg struct {
	ElevStruct elevStruct.Elevator
	Number     int
}

func DecodeElevatorMsg(str string) (outMsg ElevatorMsg) {
	json.Unmarshal([]byte(str), &outMsg)
	return
}

func EncodeElevatorMsg(msg ElevatorMsg) string {
	bytes, _ := json.Marshal(msg)
	return string(bytes)
}
