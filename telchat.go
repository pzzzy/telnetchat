package main

import (
	"bufio"
	"log"
	"net"
)

type client struct {
	name   string
	reader *bufio.Reader
	writer *bufio.Writer
}

var (
	entering = make(chan client)
	leaving  = make(chan client)
	messages = make(chan string)
)

func main() {
	listener, err := net.Listen("tcp", "localhost:8000")
	if err != nil {
		log.Fatal(err)
	}
	go broadcaster()
	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Print(err)
			continue
		}
		go handleConn(conn)
	}
}

func broadcaster() {
	clients := make(map[client]bool)
	for {
		select {
		case msg := <-messages:
			for cli := range clients {
				cli.writer.WriteString(msg)
				cli.writer.Flush()
			}
		case cli := <-entering:
			clients[cli] = true
			cli.writer.WriteString("Current clients:\n")
			for c := range clients {
				cli.writer.WriteString(" - " + c.name + "\n")
			}
			cli.writer.Flush()
		case cli := <-leaving:
			delete(clients, cli)
		}
	}
}

func handleConn(conn net.Conn) {
	cli := client{
		name:   conn.RemoteAddr().String(),
		reader: bufio.NewReader(conn),
		writer: bufio.NewWriter(conn),
	}
	cli.writer.WriteString("Enter your name: ")
	cli.writer.Flush()
	name, _, err := cli.reader.ReadLine()
	if err != nil {
		log.Print(err)
		return
	}
	cli.name = string(name)

	cli.writer.WriteString("Welcome, " + cli.name + "!\n")
	cli.writer.Flush()
	messages <- cli.name + " has joined\n"
	entering <- cli

	input := bufio.NewScanner(cli.reader)
	for input.Scan() {
		messages <- cli.name + ": " + input.Text() + "\n"
	}

	leaving <- cli
	messages <- cli.name + " has left\n"
	conn.Close()
}

