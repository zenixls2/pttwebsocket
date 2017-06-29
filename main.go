package main

import (
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/nsf/termbox-go"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
	"net/http"
	"net/url"
	"os"
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
	termbox.Init()
	termbox.SetOutputMode(termbox.Output256)
	decoder := traditionalchinese.Big5.NewDecoder()
	go func() {
		defer c.Close()
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				fmt.Println(err)
				return
			}
			result, _, err := transform.Bytes(decoder, message)
			if err != nil {
				os.Stdout.Write(message)
				os.Stdout.Sync()

			} else {
				os.Stdout.Write(result)
				os.Stdout.Sync()
			}
		}
	}()

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
