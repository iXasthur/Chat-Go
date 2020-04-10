package main

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"sort"
	"strconv"
)

// 1 byte always 213
// 2 byte always 137
// 3 byte always (	14 - rdy to chat UDP,
//					114 - send client TCP, 115 - message TCP, 116 - disconnect TCP, 117 - history request TCP,
//					118 history message TCP	)
type packageFirstBytesTemplateType struct {
	rdyToChatUDP []byte

	clientDataTCP []byte
	messageTCP []byte
	disconnectTCP []byte
	historyRequestTCP []byte
	historyMessageTCP []byte
}
var packageFirstBytesTemplates = packageFirstBytesTemplateType{
	rdyToChatUDP: 		[]byte{213, 137, 14},

	clientDataTCP:		[]byte{213, 137, 114},
	messageTCP: 		[]byte{213, 137, 115},
	disconnectTCP: 		[]byte{213, 137, 116},
	historyRequestTCP: 	[]byte{213, 137, 117},
	historyMessageTCP: 	[]byte{213, 137, 118},
}

type Message struct {
	kind []byte
	name string
	ip net.IP
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
var receivedHistory = false


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

func addMessageToHistory(msg Message){
	//msg := Message{
	//	name: name,
	//	ip: ip.String(),
	//	time: "0:0:0",
	//	text: text,
	//}
	client.history = append(client.history, msg)
}

func addPeer(name string, ip net.IP){
	peer := Peer{
		name: name,
		ip:   ip,
	}
	client.peers = append(client.peers, peer)
}

func removePeer(peer Peer){
	var i int
	for i = 0; i<len(client.peers); i++ {
		if client.peers[i].ip.String() == peer.ip.String() {
			if i == len(client.peers) - 1 {
				client.peers = client.peers[:i]
			} else {
				client.peers = append(client.peers[:i], client.peers[i:]...)
			}
			break
		}
	}
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

func receivedBroadcastMessageUDP(b []byte) {
	msg := parsePackage(b)

	if bytes.Compare(msg.kind, packageFirstBytesTemplates.rdyToChatUDP) == 0 {
		if msg.ip.String() != client.ip.String() {
			if findPeerByIP(msg.ip).ip == nil {
				addPeer(msg.name, msg.ip)
				addMessageToHistory(msg)
				resetChatWindow()

				replyMsg := Message{
					kind: packageFirstBytesTemplates.clientDataTCP,
					name: client.name,
					ip:   client.ip,
					time: getTimeString(),
					text: "Add me!\n",
				}
				p := createPackage(replyMsg)
				sendMessageTCP(p, msg.ip)
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
			break
		}
		receivedBroadcastMessageUDP(buf[:n])
	}
}

// ----------Package----------
// 1-3 bytes are type of package
// 4-7 bytes of ip
// 8 byte length of name in bytes
// 9-* name (name must be <=255 bytes)
// *+1 byte length of text in bytes
// *+1-** text
// **+1 byte length of time in bytes
// **+1-** time string
// ----------------------------
func parsePackage(b []byte) Message {
	buff := Message{
		kind: []byte{},
		name: "",
		ip:   nil,
		time: "",
		text: "",
	}
	if len(b) > 3 {
		nameLength := b[7]
		textLengthPos := 8 + nameLength
		textLength := b[textLengthPos]
		textStartPos := textLengthPos + 1
		timeLengthPos := textStartPos+textLength
		timeLength := b[timeLengthPos]
		timeStartPos := timeLengthPos + 1

		buff.kind = b[:3]
		buff.ip = net.IP(b[3:7])
		buff.name = string(b[8:8+nameLength])
		buff.text = string(b[textStartPos:textStartPos+textLength])
		buff.time = string(b[timeStartPos:timeStartPos+timeLength])
	}
	return buff
}

func createPackage(msg Message) []byte{
	var buff = msg.kind

	var ipBytes = []byte(msg.ip)
	var nameLength = byte(len(msg.name))
	var nameBytes = []byte(msg.name)
	var textLength = byte(len(msg.text))
	var textBytes = []byte(msg.text)
	var timeLength = byte(len(msg.time))
	var timeBytes = []byte(msg.time)

	buff = append(buff, ipBytes...)
	buff = append(buff, nameLength)
	buff = append(buff, nameBytes...)
	buff = append(buff, textLength)
	buff = append(buff, textBytes...)
	buff = append(buff, timeLength)
	buff = append(buff, timeBytes...)


	return buff
}

func sortHistory(){
	sort.Slice(client.history, func(i, j int) bool {
		if client.history[i].time < client.history[j].time {
			return true
		} else {
			return false
		}
	})
}

func receiveMessageTCP(b []byte){
	msg := parsePackage(b)
	if bytes.Compare(msg.kind, packageFirstBytesTemplates.messageTCP) == 0 {
		addMessageToHistory(msg)
		resetChatWindow()
	} else
	if bytes.Compare(msg.kind, packageFirstBytesTemplates.clientDataTCP) == 0 {
		addPeer(msg.name, msg.ip)
		resetChatWindow()

		if !receivedHistory {
			historyRequestMsg := Message{
				kind: packageFirstBytesTemplates.historyRequestTCP,
				name: client.name,
				ip:   client.ip,
				time: getTimeString(),
				text: "requested history!\n",
			}
			p := createPackage(historyRequestMsg)
			sendMessageTCP(p, msg.ip)
			receivedHistory = true
		}
	} else
	if bytes.Compare(msg.kind, packageFirstBytesTemplates.disconnectTCP) == 0 {
		removePeer(findPeerByIP(msg.ip))
		addMessageToHistory(msg)
		resetChatWindow()
	} else
	if bytes.Compare(msg.kind, packageFirstBytesTemplates.historyRequestTCP) == 0 {
		for _, historyMsg := range client.history {
			buff := historyMsg
			buff.kind = packageFirstBytesTemplates.historyMessageTCP
			p := createPackage(buff)
			sendMessageTCP(p, msg.ip)
		}
	} else
	if bytes.Compare(msg.kind, packageFirstBytesTemplates.historyMessageTCP) == 0 {
		addMessageToHistory(msg)
		sortHistory()
		resetChatWindow()
	}
}

func shoutOutUDP(connection net.PacketConn, client *Client){
	msg := Message{
		kind: packageFirstBytesTemplates.rdyToChatUDP,
		name: client.name,
		ip:   client.ip,
		time: getTimeString(),
		text: "joined chat!\n",
	}
	p := createPackage(msg)
	sendUDPBroadcast(connection, p)
}

func handleRequest(conn net.Conn) {
	// Make a buffer to hold incoming data.
	buf := make([]byte, 1024)
	// Read the incoming connection into the buffer.
	length, err := conn.Read(buf)
	if err != nil {
		fmt.Println("Error reading:", err.Error())
	}

	receiveMessageTCP(buf[:length])

	//// Send a response back to person contacting us.
	//conn.Write([]byte("Message received."))

	// Close the connection when you're done with it.
	conn.Close()
}

func startTCPServer(l net.Listener){
	for {
		// Listen for an incoming connection.
		conn, err := l.Accept()
		if err != nil {
			fmt.Println("Error accepting: ", err.Error())
			break
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

func disconnectTCP(){
	msg := Message{
		kind: packageFirstBytesTemplates.disconnectTCP,
		name: client.name,
		ip:   client.ip,
		time: getTimeString(),
		text: "left chat!\n",
	}
	buff := createPackage(msg)
	sendMessageToPeersTCP(buff)
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

	// Listen for incoming connections.
	listenerTCP, err := net.Listen("tcp4", client.ip.String()+":"+strconv.Itoa(client.portTCP))
	if err != nil {
		fmt.Println("Error listening:", err.Error())
		os.Exit(1)
	}
	// Close the listener when the application closes.
	defer listenerTCP.Close()

	go startTCPServer(listenerTCP)

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
		case "/upd\n":
			{
				fmt.Println("Updating chat")
				resetChatWindow()
			}
		case "/exit\n":
			{
				fmt.Println("Exiting chat")
				disconnectTCP()
				resetChatWindow()
				return
			}
		default:
			{
				// Send msg to peers
				fmt.Println("Sending message")
				msg := Message{
					kind: packageFirstBytesTemplates.messageTCP,
					name: client.name,
					ip:   client.ip,
					time: getTimeString(),
					text: text,
				}
				addMessageToHistory(msg)
				buff := createPackage(msg)
				sendMessageToPeersTCP(buff)
				resetChatWindow()
				//sendMessageTCP(connectionTCP, []byte(text))
			}
		}
	}
}