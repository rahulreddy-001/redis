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

func readCommand(c io.ReadWriter) (core.Cmds, error) {
	var buf []byte = make([]byte, 512)
	n, err := c.Read(buf)
	if err != nil {
		return core.Cmds{}, err
	}
	commands, err := core.DecodeManyCmds(buf[:n])
	cmds := make(core.Cmds, len(commands))
	if err != nil {
		return cmds, err
	}
	for i := 0; i < len(commands); i++ {
		tokens, err := core.ToArrayString(commands[i])
		if err != nil {
			return cmds, err
		}
		cmds[i] = &core.Cmd{
			Cmd:  strings.ToUpper(tokens[0]),
			Args: tokens[1:],
		}
	}
	return cmds, nil
}

func respondError(err error, c io.ReadWriter) {
	c.Write([]byte(fmt.Sprintf("-%s\r\n", err)))
}

func respond(cmds core.Cmds, c io.ReadWriter) {
	err := core.EvalAndRespondCmds(cmds, c)
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
				cmds, err := readCommand(c)
				if err != nil {
					c.Close()
					log.Println("closed connection from ", c.RemoteAddr())
					break
				}
				respond(cmds, c)
			}
		}(c)
	}
}
