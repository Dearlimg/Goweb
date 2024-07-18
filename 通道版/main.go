package main

func main() {
	Server := NewServer("127.0.0.1", 8080)
	Server.Start()
}
