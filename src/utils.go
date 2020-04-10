package main

import (
	"fmt"
	"net"
	"strconv"
	"time"
)

// Get preferred outbound ip of this machine
func getLocalIP() net.IP {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		fmt.Println(err)
		return nil
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)

	return localAddr.IP
}

func getTimeString() string {
	h, m, s := time.Now().Clock()
	hStr := strconv.Itoa(h)
	mStr := strconv.Itoa(m)
	sStr := strconv.Itoa(s)
	msStr := strconv.Itoa(time.Now().Nanosecond()/1000000)

	if len(hStr) == 1 {
		hStr = "0" + hStr
	}

	if len(mStr) == 1 {
		mStr = "0" + mStr
	}

	if len(sStr) == 1 {
		sStr = "0" + sStr
	}

	if len(msStr) != 3 {
		var additionalZeros = 3 - len(msStr)
		for i := 0; i<additionalZeros ; i++ {
			msStr = "0" + msStr
		}
	}

	return hStr+":"+mStr+":"+sStr+":"+msStr
}