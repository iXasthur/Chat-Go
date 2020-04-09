package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"strconv"
)

// 1 byte always 213
// 2 byte always 137
// 3 byte always (	14 - rdy to chat UDP,
//					114 - send client TCP, 115 - message TCP)
type packageFirstBytesTemplateType struct {
	rdyToChatUDP []byte
	clientDataTCP []byte
	messageTCP []byte
}
var packageFirstBytesTemplates = packageFirstBytesTemplateType{
	rdyToChatUDP: []byte{213, 137, 14},
	clientDataTCP: []byte{213, 137, 114},
	messageTCP: []byte{213, 137, 115},
}

type Message struct {
	name string
	ip string
	time string
	text string
}
type Peer struct {
	name string
	ip net.IP
}
type Client struct {
	ip net.IP
	name string
	history []Message

	portUDP int
	portTCP int

	peers []Peer
}
var client = Client{
	ip:   getLocalIP(),
	name: "",
	history: []Message{},

	portUDP: 0,
	portTCP: 0,

	peers: []Peer{},
}


func sendUDPBroadcast(connection net.PacketConn, b []byte){
	for i := 0; i<=255; i++ {
		addrStr := "192.168."+strconv.Itoa(i)+".255"
		addr, err := net.ResolveUDPAddr("udp4", addrStr+":"+strconv.Itoa(client.portUDP))
		if err != nil {
			fmt.Println(err)
		}
		_, err = connection.WriteTo(b, addr)
		if err != nil {
			fmt.Println(err)
		}
	}
}

func addMessageToHistory(name string, ip net.IP, text string){
	msg := Message{
		name: name,
		ip: ip.String(),
		time: "0:0:0",
		text: text,
	}
	client.history = append(client.history, msg)
}

func addPeer(name string, ip net.IP){
	peer := Peer{
		name: name,
		ip:   ip,
	}
	client.peers = append(client.peers, peer)
}

func findPeerByIP(ip net.IP) Peer {
	buff := Peer{
		name: "",
		ip:   nil,
	}
	for _, peer := range client.peers {
		if ip.String() == peer.ip.String() {
			buff = peer
		}
	}
	return buff
}

func receivedBroadcastMessageUDP(b []byte){
	//fmt.Printf("Received this: %s\n", bytes)
	name, ip := parseUDPPackage(b)
	if ip.String() != client.ip.String() {
		if ip != nil && name != "" {
			if findPeerByIP(ip).ip == nil {
				addPeer(name, ip)
				addMessageToHistory(name, ip, "joined chat!\n")
				resetChatWindow()
			}
		}
	}
}

func listenBroadcastUDP(connection net.PacketConn){
	for {
		buf := make([]byte, 1024)
		n, _, err := connection.ReadFrom(buf)
		if err != nil {
			fmt.Println(err)
			return
		}
		receivedBroadcastMessageUDP(buf[:n])
	}
}

// 1-3 bytes are type of package
// 4-7 bytes of ip
// 8 byte length of name in bytes
// 9-* name
// name must be <=255 bytes
func parseUDPPackage(b []byte) (string, net.IP){
	if len(b) > 3 {
		if bytes.Compare(b[:3], packageFirstBytesTemplates.rdyToChatUDP) == 0 {
			buffIP := net.IP{b[3],b[4],b[5],b[6]}
			buffName := string(b[8:8+b[7]])
			return buffName, buffIP
		}
	}
	return "", nil
}

func createUDPPackage(firstBytesTemplate []byte, name string, ip net.IP) []byte{
	var buff = firstBytesTemplate

	if bytes.Compare(firstBytesTemplate, packageFirstBytesTemplates.rdyToChatUDP) == 0 {
		var ipBytes = []byte(ip)
		var nameLength = byte(len(name))
		var nameBytes = []byte(name)

		buff = append(buff, ipBytes...)
		buff = append(buff, nameLength)
		buff = append(buff, nameBytes...)
	}

	return buff
}

// 1-3 bytes are type of package
// 4 byte length of text in bytes
// 5-* text
func parseTCPPackage(b []byte) string {
	if len(b) > 3 {
		if bytes.Compare(b[:3], packageFirstBytesTemplates.messageTCP) == 0 {
			buffText := string(b[4:4+b[3]])
			return buffText
		}
	}
	return ""
}
func createTCPPackage(firstBytesTemplate []byte, text string) []byte {
	var buff = firstBytesTemplate

	if bytes.Compare(firstBytesTemplate, packageFirstBytesTemplates.messageTCP) == 0 {
		var textLength = byte(len(text))
		var textBytes = []byte(text)
		buff = append(buff, textLength)
		buff = append(buff, textBytes...)
	}

	return buff
}

func receivedMessageTCP(b []byte, ip string){
	str := parseTCPPackage(b)
	if str != "" {
		peer := findPeerByIP(net.IP(ip))
		addMessageToHistory(peer.name, peer.ip, str)
		resetChatWindow()
	}
}

func shoutOutUDP(connection net.PacketConn, client *Client){
	msg := createUDPPackage(packageFirstBytesTemplates.rdyToChatUDP, client.name, client.ip)
	sendUDPBroadcast(connection, msg)
}

func handleRequest(conn net.Conn) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	length, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}

	receivedMessageTCP(buf[:length], conn.RemoteAddr().String())
	//fmt.Println("Received message: " + string(buf[:length]))

	//// Send a response back to person contacting us.
	//conn.Write([]byte("Message received."))

	// Close the connection when you're done with it.
	conn.Close()
}

func startTCPServer(){
	// Listen for incoming connections.
	l, err := net.Listen("tcp4", client.ip.String()+":"+strconv.Itoa(client.portTCP))
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer l.Close()
	//fmt.Println("Listening on " + client.ip.String() + ":" + strconv.Itoa(client.portTCP))
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			os.Exit(1)
		}
		// Handle connections in a new goroutine.
		go handleRequest(conn)
	}
}

func sendMessageTCP(b []byte, ip net.IP){
	l, err := net.Dial("tcp4", ip.String()+":"+strconv.Itoa(client.portTCP))
	if err != nil {
		fmt.Println(err)
		return
	}
	l.Write([]byte(b))
	l.Close()
}

func sendMessageToPeersTCP(b []byte){
	for _, peer := range client.peers {
		sendMessageTCP(b, peer.ip)
	}
}

func main() {
	initClearFunctions()

	if client.ip == nil {
		fmt.Println("Unable to receive IP address.")
		fmt.Println("Terminating app.")
	}

	client.name = "WinUser _iXasthur" // Name must be <=255 in bytes
	client.portUDP = 8892
	client.portTCP = 8893


	//connectionTCP,err := net.ListenPacket("tcp", client.ip.String()+":"+strconv.Itoa(client.portTCP))
	//if err != nil {
	//	panic(err)
	//}
	//defer connectionTCP.Close()
	go startTCPServer()

	connectionUDP,err := net.ListenPacket("udp4", ":"+strconv.Itoa(client.portUDP))
	if err != nil {
		panic(err)
	}
	defer connectionUDP.Close()

	shoutOutUDP(connectionUDP, &client)
	//printHeader(1)
	//printHistory()
	resetChatWindow()

	go listenBroadcastUDP(connectionUDP)


	reader := bufio.NewReader(os.Stdin)
	for {

		text, _ := reader.ReadString('\n') // Text must be <=255 in bytes

		switch text {
		case "/upd\n":{
			fmt.Println("Updating chat")
			resetChatWindow()
		}
		case "/exit\n":{
			fmt.Println("Exiting chat")
			break
		}
		default:{
			// Send msg to peers
			fmt.Println("Sending message")
			addMessageToHistory(client.name, client.ip, text)
			buff := createTCPPackage(packageFirstBytesTemplates.messageTCP, text)
			sendMessageToPeersTCP(buff)
			resetChatWindow()
			//sendMessageTCP(connectionTCP, []byte(text))
		}
		}
	}
}