package main

import (
	"fmt"
	"strconv"
)

func printHeader(newLineOffset int){
	for i := 0; i < newLineOffset; i++ {
		fmt.Println()
	}

	fmt.Println("Client info:")
	fmt.Println("IP address: " + client.ip.String())
	fmt.Println("UDP port: " + strconv.Itoa(client.portUDP))
	fmt.Println("TCP port: " + strconv.Itoa(client.portTCP))
	fmt.Println("Nickname: " + client.name)
	fmt.Print("Peers: ")
	fmt.Println()
	for _, peer := range client.peers {
		fmt.Println(peer.name + "(" + peer.ip.String() + ")")
	}

	for i := 0; i < newLineOffset; i++ {
		fmt.Println()
	}
}

func printHistory(){
	for _, msg := range client.history {
		fmt.Print(msg.time + " ")
		fmt.Print(msg.name + "(" + msg.ip.String() + ")")
		//fmt.Print("[" + msg.name + "]")
		fmt.Print(": " + msg.text)
	}
	fmt.Println()
}

func resetChatWindow(){
	clearScreen()
	printHeader(1)
	printHistory()
	fmt.Print("> ")
}