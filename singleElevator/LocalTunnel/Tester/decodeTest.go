package main

import(
	"../decoding"
	"fmt"
)

func main(){
	message := decoding.Msg{Number: 1}
	str := decoding.EncodeMsg(message)
	outMsg := decoding.DecodeMsg(str)
	fmt.Printf("%v\n", outMsg)
	return
}
