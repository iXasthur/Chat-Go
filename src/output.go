package main

import "fmt"

func printHeader(client Client, newLineOffset int){
	for i := 0; i < newLineOffset; i++ {
		fmt.Println()
	}
	fmt.Println("Client info:")
	fmt.Println("IP address: " + client.ip.String())
	fmt.Println("Nickname: " + client.name)
	fmt.Print("Peers: ")
	for _, peer := range client.peers {
		fmt.Print(peer.name+ " ")
	}
	fmt.Println()
	for i := 0; i < newLineOffset; i++ {
		fmt.Println()
	}
}