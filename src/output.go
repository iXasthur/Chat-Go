package main

import "fmt"

func printHeader(newLineOffset int){
	for i := 0; i < newLineOffset; i++ {
		fmt.Println()
	}
	fmt.Println("Client info:")
	fmt.Println("IP address: " + client.ip.String())
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
	for i := 0 ; i<len(client.history); i++  {
		fmt.Print(client.history[i].time + " ")
		//fmt.Print(client.history[i].name + "(" + client.history[i].ip + ")")
		fmt.Print("[" + client.history[i].name + "]")
		fmt.Print(": " + client.history[i].text)
		fmt.Println()
	}
	fmt.Println()
}

func resetChatWindow(){
	clearScreen()
	printHeader(1)
	printHistory()
	fmt.Print("> ")
}