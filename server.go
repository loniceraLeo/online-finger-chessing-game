package main

import (
	"flag"
	"fmt"
	"log"
	"net"
)

type player struct {
	name  string
	score int
	conn  net.Conn
}

type room struct {
	id         int //id ranges from 0 to 65535
	state      byte
	host, plyr *player
}

var (
	ready map[int]chan byte
	//rooms   []*room
	rooms map[int]*room
	count int
)

//state of rooms
const (
	waitting = iota
	playing
)

//NOROOMID is the reply info for req
const (
	NOROOMID = iota
	ROOMISFULL
	OK
)

func init() {
	flag.IntVar(&count, "c", 3, "count is the finite score for game")
	flag.Parse()
	rooms = make(map[int]*room)
	ready = make(map[int]chan byte)
	rooms[1] = &room{id: 1, state: 1, host: nil, plyr: nil}
}

func main() {
	serve("localhost", "8080")
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

		if err != nil {
			log.Fatal(err)
			return
		}
		go processFrames(conn)
	}
}

func clear(tab map[int]*room) {
	for a := range tab {
		_, err := tab[a].host.conn.Write([]byte(""))
		if err != nil {
			delete(tab, a)
		}
	}
}

func processFrames(conn net.Conn) {
	defer func() {
		conn.Close()
	}()
	for {
		info := make([]byte, 50)
		n, err := conn.Read(info)
		if err != nil {
			return
		}
		info = info[:n]
		if len(info) == 0 {
			conn.Close()
			return
		}
		if info[0] == 2 {
			r := make([]byte, 0)
			for a := range rooms {
				t := make([]byte, 3)
				t[0], t[1], t[2] = byte(rooms[a].id>>8), byte(rooms[a].id%256), rooms[a].state
				r = append(r, t...)
			}
			conn.Write(r)
		} else if info[0] == 1 {
			roomid := int(info[1])<<8 + int(info[2])
			host := newPlyr(conn)
			rooms[roomid] = &room{roomid, 0, host, nil}
			ready[roomid] = make(chan byte)
			defer func() {
				delete(rooms, roomid)
				delete(ready, roomid)
			}()
			//<-ready[roomid]
			<-ready[roomid]
			return
		} else {
			roomid := int(info[1])<<8 + int(info[2])
			i := 0
			find := false
			for i = range rooms {
				if rooms[i].id == roomid {
					find = true
					break
				}
			}
			if find == false {
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
				rooms[i].host.conn.Write(res)
				rooms[i].plyr.conn.Write(res)
				fmt.Println(1)
				//ready[i] <- 0
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
	rm.state = playing
	for rm.host.score != count && rm.plyr.score != count {
		var (
			x, y int
		)
		ch := make(chan byte, 1)
		go func() {
			b := make([]byte, 1)
			rm.host.conn.Read(b)
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
			rm.host.conn.Write(b)
			continue
		}
		if x-y == -1 || x-y == 2 {
			rm.host.score++
		} else if x == y {
			//do nothing
		} else {
			rm.plyr.score++
		}
		if rm.plyr.score == count || rm.host.score == count {
			b := make([]byte, 1)
			if rm.plyr.score == count {
				b[0] = 127
			} else {
				b[0] = 128
			}
			rm.plyr.conn.Write(b)
			rm.host.conn.Write(b)
		} else {
			b := make([]byte, 2)
			b[0], b[1] = byte(rm.host.score), byte(rm.plyr.score)
			rm.host.conn.Write(b)
			rm.plyr.conn.Write(b)
		}
	}
}
