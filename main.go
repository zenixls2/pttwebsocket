package main

import (
	"bufio"
	"flag"
	"fmt"
	"github.com/gorilla/websocket"
	"github.com/nsf/termbox-go"
	"golang.org/x/text/encoding/traditionalchinese"
	"golang.org/x/text/transform"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
)

var addr = flag.String("addr", "ws.ptt.cc", "http service address")

var decoder = traditionalchinese.Big5.NewDecoder()

//var b2uMap []rune

func b2u(message []byte) []byte {
	if len(message) == 0 {
		return []byte{}
	}
	unit := 1
	for i := 0; i < len(message); i += unit {
		if message[i] >= 0x81 && message[i] <= 0xfe {
			if i+1 < len(message) {
				ch1 := message[i+1]
				if ch1 >= 0x40 && ch1 <= 0x7e ||
					ch1 >= 0xa1 && ch1 <= 0xfe {
					unit = 2
					r := b2uMap[int(message[i])<<8|int(ch1)]
					if r == 0 {
						panic(fmt.Sprintf("%d %d %d",
							message[i], ch1, int(message[i])<<8+int(ch1)))
					} else {
						os.Stdout.Write([]byte(string(r)))
					}
				} else {
					unit = 1
					os.Stdout.Write([]byte{message[i]})
				}
			} else {
				unit = 1
				return []byte{message[i]}
			}
		} else if message[i] >= ' ' {
			os.Stdout.Write([]byte{byte(message[i])})
			unit = 1
		}
	}
	os.Stdout.Sync()
	return []byte{}
}

func output(message []byte) []byte {
	result, _, err := transform.Bytes(decoder, message)
	if err != nil {
		os.Stdout.Write(message)
		os.Stdout.Sync()
	} else {
		os.Stdout.Write(result)
		os.Stdout.Sync()
	}
	return []byte{}
}

func outputControlCode(buf []byte) bool {
	os.Stdout.Write(buf)
	os.Stdout.Sync()
	return true
}

// DEPRECATED
func loadb2u() {
	b2uMap = make([]rune, 65535, 65535)
	file, err := os.Open("uao250-b2u.big5.txt")
	if err != nil {
		fmt.Println(err)
		panic(err)
	}
	defer file.Close()
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		row := scanner.Text()
		results := strings.Split(strings.TrimSpace(row), " ")
		val1, err := strconv.ParseInt(results[0][2:], 16, 32)
		if err != nil {
			panic(err)
		}
		val2, err := strconv.ParseInt(results[1][2:], 16, 32)
		if err != nil {
			panic(err)
		}
		b2uMap[int(val1)] = rune(val2)
	}
}

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
	go func() {
		defer c.Close()
		defer close(done)
		buffer := []byte{}
		CR := false
		CSI := false
		COMMAND := false
		for {
			_, message, err := c.ReadMessage()
			if err != nil {
				fmt.Println(err)
				return
			}

			for _, i := range message {
				if i < ' ' {
					switch i {
					case '\x1b':
						COMMAND = true
						buffer = b2u(buffer)
						os.Stdout.Write([]byte{i})
						break
					case '\n':
						buffer = b2u(buffer)
						if !CR {
							os.Stdout.Write([]byte{'\n'})
						}
						CR = false
						break
					case '\r':
						buffer = b2u(buffer)
						CR = true
						os.Stdout.Write([]byte{'\n'})
						break
					default:
						buffer = append(buffer, i)
					}
				} else {
					if COMMAND {
						if i < 0x81 {
							if i == '[' {
								CSI = true
							}
							if CSI {
								switch {
								case i >= '0' && i <= '9' ||
									i == '?' ||
									i == '[' ||
									i == ';' ||
									i == '=':
									os.Stdout.Write([]byte{i})
									break
								case i == '~' ||
									i >= 'a' && i <= 'z' ||
									i >= 'A' && i <= 'Z':
									CSI = false
									COMMAND = false
									os.Stdout.Write([]byte{i})
									break
								default:
									CSI = false
									COMMAND = false
									if i != ' ' {
										panic(fmt.Sprintf("Not catched chr %d", int(i)))
									}
								}
							} else {
								panic(fmt.Sprintf("Not catched chr %d", int(i)))
							}
						} else {
							CSI = false
							COMMAND = false
							buffer = append(buffer, i)
						}
					} else {
						buffer = append(buffer, i)
					}
				}
			}
			buffer = b2u(buffer)
		}
	}()

	defer termbox.Close()
	for {
		defer c.Close()
		var msg string
		switch ev := termbox.PollEvent(); ev.Type {
		// refer to http://www.novell.com/documentation/extend52/Docs/help/Composer/books/TelnetAppendixB.html
		case termbox.EventKey:
			switch ev.Key {
			case termbox.KeyEsc:
				msg = "\u001b"
				break
			case termbox.KeyArrowDown:
				os.Stdout.Write([]byte{27, 91, 49, 65})
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
