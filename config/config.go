package config
import "time"
const (
	////// ELEVSTRUCT  ///////
	DOOR_OPEN_TIME        = 3*time.Second
	NUMFLOORS             = 4

	////// MSG HANDLER ///////
	SEND_INTERVAL         = 30 * time.Millisecond
	FSM_PATH              = "../FSM/fsm.go" ///home/student/Desktop/project-gruppe-60/FSM/FSM"
	RECEIVE_PORT_HANDLER  = 55555
	SEND_PORT_HANDLER     = 44444
	BROAD_PORT_HANDLER_TX = 16977
	BROAD_PORT_HANDLER_RX = 16542
	PEER_PORT             = 15789

	//////     FSM     ///////
	MSGHANDLER_PATH       = "../msgHandler/msgHandler.go" //"/home/student/Desktop/project-gruppe-60/msgHandler/msgHandler" 
	SEND_PORT_FSM         = RECEIVE_PORT_HANDLER
	RECEIVE_PORT_FSM      = SEND_PORT_HANDLER

	//////    BACKUP   //////
	BROAD_PORT_MASTER_TX  = BROAD_PORT_HANDLER_RX
	BROAD_PORT_MASTER_RX  = BROAD_PORT_HANDLER_TX
)

