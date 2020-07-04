package main

import (
	"fmt"
	"log"
	"net"
	"syscall"

	"flag"
)

const version = "1.0.0"

//operations
const (
	newroom = 1 << iota
	roomlist
	enter
)

const (
	hostmode = iota
	playermode
)

var (
	remote string
	mode   byte
)

var (
	name string
)

func main() {
	flag.StringVar(&remote, "r", ":8080", "The server of game")
	flag.Parse()
	for {
		var ch int
		surface()
		fmt.Scan(&ch)
		sendReq(ch)
	}
}

func sendReq(method int, extra ...interface{}) {
	var (
		msg []byte
	)

	switch method {
	case 1:
		msg = doCreatePlayRoom()
	case 2:
		msg = doGetRoomList()
	case 3:
		msg = doEnterRoom()
	case 4:
		leave()
	}

	conn, err := net.Dial("tcp", remote)
	defer conn.Close()
	if err != nil {
		log.Fatal("连接失败,服务器可能已经关闭")
		return
	}
	conn.Write(msg)
	if method == 2 {
		res := make([]byte, 200)
		n, _ := conn.Read(res)
		res = res[:n]
		printList(res)
	} else if method == 1 {
		waitForplyr(conn)
		mode = hostmode
		fmt.Println("连接成功")
		playground(conn)
	} else {
		res := make([]byte, 1)
		conn.Read(res)
		switch res[0] {
		case 0:
			fmt.Println("房间不存在")
		case 1:
			fmt.Println("房间人数已满")
		case 2:
			mode = playermode
			fmt.Println("连接成功")
			playground(conn)
		}
	}
}

func playground(conn net.Conn) {
	for {
		var ch byte
		b := make([]byte, 2)
		fmt.Println("0.石头 1.剪刀 2.布")
		fmt.Scan(&ch)
		conn.Write([]byte{ch})
		fmt.Println("等待对方回应...")
		n, _ := conn.Read(b)
		b = b[:n]
		if len(b) == 1 && b[0] == 1 {
			fmt.Println("有人在装逼 这局不算(请输入0 1 2)")
			continue
		} else if len(b) == 1 {
			if mode == hostmode {
				if b[0] == 127 {
					fmt.Println("你输了!")
				} else {
					fmt.Println("你赢了!")
				}
			} else {
				if b[0] == 127 {
					fmt.Println("你赢了!")
				} else {
					fmt.Println("你输了!")
				}
			}
			break
		} else {
			if mode == hostmode {
				fmt.Printf("你: %v 分    对方: %v 分\n", b[0], b[1])
			} else {
				fmt.Printf("你: %v 分    对方: %v 分\n", b[1], b[0])
			}
		}
	}
}

func doCreatePlayRoom() []byte {
	var id int

	fmt.Println("房间号:")
	fmt.Scan(&id)
	msg := make([]byte, 3)
	msg[0] = newroom
	msg[1] = byte(id >> 8)
	msg[2] = byte(id % 256)

	return msg
}

func doGetRoomList() []byte {
	msg := make([]byte, 1)
	msg[0] = roomlist

	return msg
}

func doEnterRoom() []byte {
	var id int

	fmt.Println("房间号:")
	fmt.Scan(&id)
	msg := make([]byte, 3)
	msg[0] = enter
	msg[1] = byte(id >> 8)
	msg[2] = byte(id % 256)

	return msg
}

func printList(table []byte) {
	num := len(table) / 3
	fmt.Printf("房间数: %v\n", num)
	for i := 0; i < len(table); i += 3 {
		var state string
		if table[i+2] == 0 {
			state = "waiting"
		} else {
			state = "playing"
		}
		fmt.Printf("roomid: %v, state: %v \n", int(table[i])<<8+int(table[i+1]), state)
	}
}

func waitForplyr(conn net.Conn) byte {
	var r byte
	fmt.Println("waiting...")
	b := make([]byte, 1)
	conn.Read(b)
	r = b[0]
	return r
}

func leave() {
	fmt.Println("Bye~")
	syscall.Exit(0)
}

func ver() {
	fmt.Println(version)
}

func surface() {
	fmt.Println("<1.> 开房间")
	fmt.Println("<2.> 房间列表")
	fmt.Println("<3.> 加入房间")
	fmt.Println("<4.> 离开")
}

func login() {
	fmt.Println("尊姓大名:")
	fmt.Scan(&name)
}

func printName() {
	fmt.Println(name)
}
