package main

import (
	"fmt"
	"io"
	"net"
	"strconv"
)

func adminStart() {
	go func() {
		l, err := net.Listen("tcp", ":"+strconv.Itoa(int(conf.AdminPort)))
		if err != nil {
			Log(LogLevelError, err)
		}
		defer l.Close()
		for {
			conn, err := l.Accept()
			if err != nil {
				Log(LogLevelError, err)
			}
			go handleCommand(conn)
		}
	}()
}

func handleCommand(c net.Conn) {
	defer c.Close()
	for {
		inBuf := make([]byte, 128)
		n, err := c.Read(inBuf)
		if err != nil {
			if err != io.EOF {
				fmt.Fprintf(c, "read command error: %s\n", err)
			}
			continue
		}
		cmd := string(inBuf[:n-2])
		switch cmd {
		case "shutdown":
			_, err = shutdown()
			if err != nil {
				fmt.Fprintln(c, "shutdown server failed")
			}
			break
		case "stats":
			outBuf := ServerStatsReport()
			fmt.Fprintln(c, outBuf)
		case "quit": // close connection
			return
		default:
			fmt.Fprintln(c, "no known command")
		}
	}
}

func shutdown() (string, error) {
	return "shutdown successfully", nil
}

func ServerStatsReport() string {
	return StatsReport()
}
