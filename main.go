package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
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

	reader := bufio.NewReader(os.Stdin)

	go func() {
		defer c.Close()
		defer close(done)
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				fmt.Println(err)
				return
			}
			fmt.Println(string(message))
		}
	}()
	for {
		defer c.Close()
		bt, err := reader.ReadByte()
		if err != nil {
			fmt.Println(err)
			return
		}
		// Necessary or not? I am not sure
		msg := []byte{bt}
		if bt == LF {
			msg = []byte{CR, LF}
		}
		err = c.WriteMessage(websocket.BinaryMessage, msg)
		if err != nil {
			fmt.Println(err)
			return
		}
	}
}
