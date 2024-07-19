package main

import (
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Message struct {
	username  string
	way       string
	othername string
	message   string
}

func main() {
	var mu sync.Mutex
	msg := new(Message)
	userMap := make(map[string]net.Conn)
	checkUser := make(map[string]bool)
	addr, _ := net.ResolveTCPAddr("tcp4", "127.0.0.1:8080")
	lis, _ := net.ListenTCP("tcp", addr)
	fmt.Println("服务器正在运行中.....")
	for {
		conn, _ := lis.Accept()
		go func(net.Conn) {
			defer conn.Close()
			for {
				buf := make([]byte, 1024)
				n, _ := conn.Read(buf)
				newbuf := string(buf[:n])
				meg := strings.Split(newbuf, "_")
				msg.username = meg[0]
				msg.way = meg[1]
				msg.othername = meg[2]
				msg.message = meg[3]

				//mu.lock()
				if msg.username != "" {
					userMap[msg.username] = conn
					checkUser[msg.username] = true
				}
				//userMap[msg.username] = conn
				//mu.Unlock()

				if msg.message == "checkUser" {
					if checkUser[msg.username] == true {
						conn.Write([]byte(msg.username + "___userzai"))
						continue
					}
				}

				if msg.message == "@" {
					count := 0
					for user, _ := range userMap {
						time.Sleep(100 * time.Millisecond)
						conn.Write([]byte("___" + user))
						count++
					}
					conn.Write([]byte("___" + strconv.Itoa(count)))
					continue
				}

				res := fmt.Sprintf(msg.username + "_" + msg.way + "_" + msg.othername + "_" + msg.message)
				fmt.Println("服务器接受到解析后的信息", res)

				mu.Lock()
				if msg.way == "@" && msg.othername != "" {
					for user, userconn := range userMap {
						if user == msg.othername {
							if msg.message == "C:\\Users\\gaoji\\Desktop\\rubbishDunp" {
								filepath := msg.message
								res, _ := PathExists(filepath)
								if res {
									userconn.Write([]byte("___文件接受成功,来自于" + msg.username))
								} else {
									userconn.Write([]byte("___文件接受成功"))
								}
							} else {
								userconn.Write([]byte("___" + msg.othername + msg.message + "(这是悄悄话)"))
							}
						}
					}
					fmt.Println("信息私发成功")
					msg.othername = ""
					msg.way = ""
				} else {
					for user, userconn := range userMap {
						fmt.Println(user, userconn)
						if user != msg.username {
							userconn.Write([]byte(res))
							fmt.Println("发送成功", res)
						}
					}
				}
				mu.Unlock()
			}
		}(conn)
	}
}

func PathExists(path string) (bool, error) {
	_, err := os.Stat(path)
	if err == nil { //文件或目录存在
		return true, nil
	}
	if os.IsNotExist(err) {
		return false, nil
	}
	return false, err
}
