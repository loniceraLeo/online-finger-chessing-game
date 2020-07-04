package main

import (
	"flag"
	"fmt"
	"io"
	"log"
	"net"
)

type player struct {
	name  string
	score int
	conn  net.Conn
}

type room struct {
	id           int //id ranges from 0 to 65535
	state        byte
	hoster, plyr *player
}

var (
	ready   []chan byte
	rooms   []*room
	roomLen int
	count   int
)

//state of rooms
const (
	waitting = iota
	playing
)

//NOROOMID is
const (
	NOROOMID = iota
	ROOMISFULL
	OK
)

func init() {
	flag.IntVar(&count, "c", 3, "count is the finite score for game")
	flag.Parse()
	rooms = make([]*room, 100)
	ready = make([]chan byte, 100)
	roomLen = 1
	rooms[0] = &room{id: 1, state: 1, hoster: nil, plyr: nil}
}

func main() {
	serve("127.0.0.1", "8080")
}

func serve(addr, port string) {
	fmt.Println("server is on listening")
	local := net.JoinHostPort(addr, port)
	listener, err := net.Listen("tcp", local)
	defer listener.Close()
	if err != nil {
		log.Fatal(err)
		return
	}

	for {
		conn, err := listener.Accept()
		defer conn.Close()

		if err != nil {
			log.Fatal(err)
			return
		}
		go processFrames(conn)
	}
}

func processFrames(conn net.Conn) {
	for {
		info := make([]byte, 50)
		n, err := conn.Read(info)
		if err == io.EOF {
			return
		}
		info = info[:n]
		if len(info) == 0 {
			conn.Close()
			fmt.Println(conn)
			return
		}
		fmt.Println(info)
		if info[0] == 2 {
			r := make([]byte, 0)
			for i := 0; i < roomLen; i++ {
				t := make([]byte, 3)
				t[0], t[1], t[2] = byte(rooms[i].id>>8), byte(rooms[i].id%256), rooms[i].state
				r = append(r, t...)
			}
			conn.Write(r)
		} else if info[0] == 1 {
			roomid := int(info[1])<<8 + int(info[2])
			hoster := newPlyr(conn)
			rooms[roomLen] = &room{roomid, 0, hoster, nil}
			ready[roomLen] = make(chan byte, 1)
			defer func() {
				rooms[roomLen] = nil
				ready[roomLen] = nil
				roomLen--
			}()
			roomLen++
			<-ready[roomLen-1]
			<-ready[roomLen-1]
			return
		} else {
			roomid := int(info[1])<<8 + int(info[2])
			i := 0
			for i = 0; i < roomLen; i++ {
				if rooms[i].id == roomid {
					break
				}
			}
			if i == roomLen {
				res := make([]byte, 1)
				res[0] = NOROOMID
				conn.Write(res)
				return
			} else if rooms[i].state == playing {
				res := make([]byte, 1)
				res[0] = ROOMISFULL
				conn.Write(res)
				return
			} else {
				rooms[i].plyr = newPlyr(conn)
				res := make([]byte, 1)
				res[0] = OK
				rooms[i].hoster.conn.Write(res)
				rooms[i].plyr.conn.Write(res)
				fmt.Println(1)
				ready[i] <- 0
				playground(rooms[i], i)
				ready[i] <- 0
				return
			}
		}
	}
}
func newPlyr(conn net.Conn) *player {
	p := new(player)
	p.name, p.score, p.conn = "default-user", 0, conn
	return p
}

func playground(rm *room, i int) {
	for rm.hoster.score != count && rm.plyr.score != count {
		var (
			x, y int
		)
		ch := make(chan byte, 1)
		go func() {
			b := make([]byte, 1)
			rm.hoster.conn.Read(b)
			x = int(b[0])
			ch <- 0
		}()
		go func() {
			b := make([]byte, 1)
			rm.plyr.conn.Read(b)
			y = int(b[0])
			ch <- 0
		}()
		<-ch
		<-ch
		if (x < 0 || x > 2) || (y < 0 || y > 2) {
			b := make([]byte, 1)
			b[0] = 1
			rm.plyr.conn.Write(b)
			rm.hoster.conn.Write(b)
			continue
		}
		if x-y == -1 || x-y == 2 {
			rm.hoster.score++
		} else if x == y {
			//do nothing
		} else {
			rm.plyr.score++
		}
		if rm.plyr.score == count || rm.hoster.score == count {
			b := make([]byte, 1)
			if rm.plyr.score == count {
				b[0] = 127
			} else {
				b[0] = 128
			}
			rm.plyr.conn.Write(b)
			rm.hoster.conn.Write(b)
		} else {
			b := make([]byte, 2)
			b[0], b[1] = byte(rm.hoster.score), byte(rm.plyr.score)
			rm.hoster.conn.Write(b)
			rm.plyr.conn.Write(b)
		}
	}
}
