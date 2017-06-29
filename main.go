package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/nsf/termbox-go"
	"net/http"
	"net/url"
)

var addr = flag.String("addr", "ws.ptt.cc", "http service address")

const (
	CR = byte('\r')
	LF = byte('\n')
)

func main() {
	flag.Parse()
	u := url.URL{Scheme: "wss", Host: *addr, Path: "/bbs"}
	fmt.Println("connecting to ", u.String())

	header := http.Header{}
	header.Add("Origin", "https://robertabcd.github.io")
	c, _, err := websocket.DefaultDialer.Dial(u.String(), header)
	if err != nil {
		fmt.Println("dail error: ", err)
		return
	}
	defer c.Close()

	done := make(chan struct{})

	go func() {
		defer c.Close()
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Print(string(message))
		}
	}()
	termbox.Init()
	defer termbox.Close()
	for {
		defer c.Close()
		var msg string
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				msg = "\u001b"
				break
			case termbox.KeyArrowDown:
				msg = "\u001b[B"
				break
			case termbox.KeyArrowLeft:
				msg = "\u001b[D"
				break
			case termbox.KeyArrowRight:
				msg = "\u001b[C"
				break
			case termbox.KeyArrowUp:
				msg = "\u001b[A"
				break
			case termbox.KeyEnter:
				msg = "\u001bOM"
				break
			case termbox.KeyTab:
				msg = "\u0009"
				break
			case termbox.KeyPgdn:
				msg = "\u001b[6~"
				break
			case termbox.KeyPgup:
				msg = "\u001b[5~"
				break
			default:
				msg = string(ev.Ch)
				break
			}
			break
		default:
			msg = string(ev.Ch)
			break
		}
		bmsg := []byte(msg)
		err := c.WriteMessage(websocket.BinaryMessage, bmsg)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
