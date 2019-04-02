package elevStruct

import "fmt"
import "../../config"
/*
Module for structs and types representing a single elevator
*/

type behaviourType int

const (
	E_Initial   behaviourType = 3
	E_Moving                 = 2
	E_Idle                   = 1
	E_DoorOpen               = 0
)
type MotorDirection int

const (
	MD_Up   MotorDirection = 1
	MD_Down                = -1
	MD_Stop                = 0
)

type ButtonType int

const (
	BT_HallUp   ButtonType = 0
	BT_HallDown            = 1
	BT_Cab                 = 2
)

type ButtonEvent struct {
	Floor     int
	Button    ButtonType
}

type LightType [2][config.NUMFLOORS]bool
type OrderType [3][config.NUMFLOORS]bool

type Elevator struct {
	Floor         int
	Dir          MotorDirection
	Orders       OrderType
	Behaviour    behaviourType
	LightMatrix  LightType
}

func ElevatorInit() (e Elevator) {
	e = Elevator{Dir: MD_Stop}
	return
}
func (e *Elevator) Duplicate() (e2 Elevator) {
	e2 = *e
	return
}

// finds the difference between two order matrices and returns slices containing 
// the button and floor for the different orders
func (e *Elevator) Differences(e2 Elevator) ([]int, []int) { 
	buttons := make([]int, 0)
	floors := make([]int, 0)
	for i := 0; i < 3; i++ {
		for j := 0; j < config.NUMFLOORS; j++ {
			if e2.Orders[i][j] == true && e.Orders[i][j] == false {
				buttons = append(buttons, i)
				floors = append(floors, j)
			}
		}
	}
	return buttons, floors
}

func (e *Elevator) PrintOrders() {
	for i := 0; i < 3; i++ {
		fmt.Println()
		for j := 0; j < config.NUMFLOORS; j++ {
			fmt.Print(e.Orders[i][j], " ")
		}
	}
	fmt.Println()
}
func (e *Elevator) PrintLightMatrix() {
	for i := 0; i < 2; i++ {
		fmt.Println()
		for j := 0; j < config.NUMFLOORS; j++ {
			fmt.Print(e.LightMatrix[i][j], " ")
		}
	}
	fmt.Println()
}

func (e *Elevator) PrintStates() {
	fmt.Println("Floor: ", e.Floor)
	fmt.Println("Dir: ", int(e.Dir))
	fmt.Println("Orders:")
	e.PrintOrders()
	fmt.Println("Behavior: ", int(e.Behaviour))
	fmt.Println("LightMatrix: ")
	e.PrintLightMatrix()
}
