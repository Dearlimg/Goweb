package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"math/rand"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

func login() {
	db, err := sql.Open("mysql", "root:1234@tcp(127.0.0.1:3306)/first")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	var username string
	var password string
	fmt.Printf("请输入账号:")
	fmt.Scanf("%s", &username)
	fmt.Printf("请输入密码:")
	fmt.Scanf("%s", &password)

	stmt, err := db.Prepare("select username from clients")
	if err != nil {
		fmt.Println(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	if err != nil {
		fmt.Println(err)
	}
	defer rows.Close()
	for rows.Next() {
		var username string
		var password string
		rows.Scan(&username, &password)
		if username == username && password == password {
			realmain(username)
		}
	}

}

func register() {
	db, err := sql.Open("mysql", "root:1234@tcp(127.0.0.1:3306)/first")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	stmt, err := db.Prepare("insert into clients(username,password) values(?,?)")
	if err != nil {
		fmt.Println(err)
	}
	defer stmt.Close()

	rand.Seed(time.Now().UnixNano())

	var username string
	var password string
	fmt.Println("请输入注册账号:")
	fmt.Scanf("%s", &username)
	fmt.Println("请输入密码:")
	fmt.Scanf("%s", &password)

	_, err = stmt.Exec(username, password)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println("注册成功!")
	start()
}

func Instructions() {
	fmt.Println("聊天室,/q找人私聊,/file发送文件,默认群聊")
}

type message struct {
	username  string
	way       string
	othername string
	message   string
}

var wg sync.WaitGroup

func start() {
	fmt.Println("-----------------------------------")
	fmt.Println("      welcome to chatRoom")
	fmt.Println("-----------------------------------")
	fmt.Println("1.登录")
	fmt.Println("2.注册")
	fmt.Println("3.说明")
	fmt.Println("4.退出")
	fmt.Println("-----------------------------------")
	var n int
	fmt.Scanf("%d", &n)
	switch n {
	case 1:
		login()
		break
	case 2:
		register()
		start()
		break
	case 3:
		Instructions()
		break
	case 4:
		os.Exit(0)
	}
}

func main() {
	start()
}

func realmain(username string) {
	wg.Add(2)
	var meg message
	meg.username = username
	addr, _ := net.ResolveTCPAddr("tcp4", "127.0.0.1:8080")
	conn, _ := net.DialTCP("tcp4", nil, addr)
	fmt.Println("链接成功")
	defer conn.Close()

	now := time.Now()

	conn.Write([]byte(meg.username + "___" + "欢迎用户" + meg.username + "于" + now.Format("2006-01-02 15:04:05") + "加入聊天室"))

	go func(conn net.Conn, meg message) {
		defer wg.Done()
		fmt.Println("请输入要发送的信息:(之提醒一次)")
		for {
			meg.message = ""
			meg.othername = ""
			meg.way = ""
			//fmt.Println("请输入要发送的信息:")
			fmt.Scanln(&meg.message)
			if meg.message == "/q" {
				conn.Write([]byte("___" + "@"))
				fmt.Println("输入@+name可以私聊")
				fmt.Scanln(&meg.way, &meg.othername)
				fmt.Println("选择成功,可以输入内容了")
				fmt.Scanln(&meg.message)
				conn.Write([]byte(meg.username + "_" + meg.way + "_" + meg.othername + "_" + meg.message))
				continue
			}
			if meg.message == "/file" {
				conn.Write([]byte("___" + "@"))
				fmt.Println("输入@+name可以发文件")
				fmt.Scanln(&meg.way, &meg.othername)
				fmt.Println("选择成功,可以输入文件地址了")
				fmt.Scanln(&meg.message)
				conn.Write([]byte(meg.username + "_" + meg.way + "_" + meg.othername + "_" + meg.message))
				fmt.Println("文件发送成功")
				continue
			}
			res := fmt.Sprintf(meg.username + "_" + meg.way + "_" + meg.othername + "_" + meg.message)
			//fmt.Println(res)
			_, _ = conn.Write([]byte(res))
		}
	}(conn, meg)

	go func(conn net.Conn) {
		defer wg.Done()
		for {
			buf := make([]byte, 1024)
			n, err := conn.Read(buf)
			if err != nil {
				fmt.Println(err)
			}

			str := strings.Split(string(buf[:n]), "_")
			fmt.Println(str)
		}
	}(conn)
	wg.Wait()
}
