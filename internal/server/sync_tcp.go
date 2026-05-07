package server

import (
	"fmt"
	"io"
	"log"
	"net"
	"redis/internal/core"
	"strings"
)

func logPanic(message string, err error) {
	if err != nil {
		log.Panicln(message, err)
	}
}

func readCommand(c io.ReadWriter) (core.Cmd, error) {
	var buf []byte = make([]byte, 512)
	n, err := c.Read(buf)
	if err != nil {
		return core.Cmd{}, err
	}
	tokens, err := core.DecoreArrayString(buf[:n])
	if err != nil {
		return core.Cmd{}, err
	}
	return core.Cmd{
		Cmd:  strings.ToUpper(tokens[0]),
		Args: tokens[1:],
	}, nil
}

func respondError(err error, c io.ReadWriter) {
	c.Write([]byte(fmt.Sprintf("-%s\r\n", err)))
}

func respond(cmd core.Cmd, c io.ReadWriter) {
	err := core.EvalAndRespond(cmd, c)
	if err != nil {
		respondError(err, c)
	}
}

func RunTCPServer() {
	log.Println("started sync tcp server")

	lsnr, err := net.Listen("tcp", "0.0.0.0:8004")

	logPanic("error starting the tcp server ", err)
	log.Println("started listening on port ", 8004)

	for {
		c, err := lsnr.Accept()
		logPanic("error accepting the tcp connections ", err)
		log.Println("accepted connection from ", c.RemoteAddr())

		go func(c net.Conn) {
			for {
				cmd, err := readCommand(c)
				if err != nil {
					c.Close()
					log.Println("closed connection from ", c.RemoteAddr())
					break
				}
				respond(cmd, c)
			}
		}(c)
	}
}
