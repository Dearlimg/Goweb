package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/hex"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"net"
	"os"
	"strings"
	"sync"
	"time"
)

func searchUser(name string) bool {
	db, err := sql.Open("mysql", "root:1234@tcp(127.0.0.1:3306)/first")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	stmt, err := db.Prepare("select username from clients")
	if err != nil {
		fmt.Println(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	for rows.Next() {
		var username string
		rows.Scan(&username)
		if username == name {
			return true
		}
	}
	return false
}

func login() {
	db, err := sql.Open("mysql", "root:1234@tcp(127.0.0.1:3306)/first")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	stmt, err := db.Prepare("select username,password from clients")
	if err != nil {
		fmt.Println(err)
	}
	defer stmt.Close()

	var username1, password1 string
	fmt.Println("请输入账号:")
	fmt.Scan(&username1)
	fmt.Println("请输入密码:")
	fmt.Scan(&password1)

	rows, _ := stmt.Query()
	for rows.Next() {
		var username string
		var password string
		rows.Scan(&username, &password)

		md5password := fmt.Sprintf("%x", md5.Sum([]byte(password1)))
		//fmt.Println(username, password)
		if username == username1 && password == md5password {
			realmain(username)
		}
	}
	fmt.Println("登录失败")
	start()

	/*
		rows, err := stmt.Query()
		for rows.Next() {
			var username string
			var password string
			rows.Scan(&username, &password)

			md5password := fmt.Sprintf("%x", md5.Sum([]byte(password1)))
			//fmt.Println(username, password)
			if username == username1 && password == md5password {
				userMap[username1] = true
				fmt.Println("登录成功")
				realmain(username)
			}
		}
		fmt.Println("登录失败")
		start()

	*/
}

func register() {
	db, err := sql.Open("mysql", "root:1234@tcp(127.0.0.1:3306)/first")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	stmt, err := db.Prepare("INSERT INTO clients(username,password) values(?,?)")
	if err != nil {
		panic(err)
	}
	defer stmt.Close()

	var username, password string
	fmt.Println("请输入账号:")
	fmt.Scan(&username)
	fmt.Println("请输入密码:")
	fmt.Scan(&password)

	ok := searchUser(username)

	if ok == true {
		fmt.Println("账号已经被注册,请换一个")
		start()
	}

	md5Pwd := md5.Sum([]byte(password))
	strmd5 := hex.EncodeToString(md5Pwd[:])

	_, err = stmt.Exec(username, strmd5)
	if err != nil {
		panic(err)
	}
	fmt.Println("注册成功")
	username = ""
	password = ""
	start()
}

var wg sync.WaitGroup

func showInformation(username1 string) {
	db, err := sql.Open("mysql", "root:1234@tcp(127.0.0.1:3306)/first")
	if err != nil {
		fmt.Println(err)
	}
	defer db.Close()

	stmt, err := db.Prepare("select * from clients")
	if err != nil {
		fmt.Println(err)
	}
	defer stmt.Close()

	rows, err := stmt.Query()
	for rows.Next() {
		var id int
		var username string
		var password string
		rows.Scan(&id, &username, &password)
		if username == username1 {
			fmt.Println("当前账号信息:"+username, password)
		}
	}
}

func checkUser(conn net.Conn, username string) {
	msg := []byte(username + "___" + "checkUser")
	conn.Write(msg)
}

var mu sync.Mutex

func realmain(username string) {
	wg.Add(2)
	var meg message
	meg.username = username

	addr, _ := net.ResolveTCPAddr("tcp4", "127.0.0.1:8080")
	conn, _ := net.DialTCP("tcp4", nil, addr)

	checkUser(conn, meg.username)

	fmt.Println("链接成功")
	defer conn.Close()

	now := time.Now()

	conn.Write([]byte(meg.username + "___" + "欢迎用户" + meg.username + "于" + now.Format("2006-01-02 15-04-05") + "加入聊天室"))

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
			if meg.message == "/show" {
				showInformation(username)
				fmt.Println("如果你想修改密码,请输入M(不想就随便乱输然后,就可以直接对话了)")
				var modify string
				fmt.Scanln(&modify)
				if modify == "M" {
					var newPassword string
					fmt.Println("请输入新密码:")
					fmt.Scanln(&newPassword)
					updatePassword(username, newPassword)
				}
				continue
			}
			if meg.message == "/quit" {
				start()
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
			//msg1:=str[0]
			msg4 := str[3]
			if msg4 == "userzai" {
				fmt.Println("用户已经存在,请重新登录")
				mu.Lock()
				start()
				mu.Unlock()
				break
			}

			fmt.Println(str)
		}
	}(conn)
	wg.Wait()
}

func Instructions() {
	fmt.Println("聊天室,/q找人私聊,/file发送文件,/show展示账号密码,/@显示在线人员已经人数,默认群聊")
}

type message struct {
	username  string
	way       string
	othername string
	message   string
}

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
		start()
		break
	case 4:
		os.Exit(0)
	}
}

func main() {
	start()
}

func updatePassword(username, newPassword string) error {
	db, err := sql.Open("mysql", "root:1234@tcp(127.0.0.1:3306)/first")
	if err != nil {
		return err
	}
	defer db.Close()

	stmt, err := db.Prepare("UPDATE clients SET password = ? WHERE username = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()

	_, err = stmt.Exec(newPassword, username)
	if err != nil {
		return err
	}

	fmt.Printf("用户 %s 的密码已成功更新为 %s\n", username, newPassword)
	return nil
}
