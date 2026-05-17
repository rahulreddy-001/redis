package server

import (
	"log"
	"net"
	"redis/internal/core"
	"syscall"
	"time"
)

func RunAsyncTCPServer() {
	log.Println("started async tcp server")

	maxClients := 20000
	conClients := 0

	// read epoll events
	var events []syscall.EpollEvent = make([]syscall.EpollEvent, maxClients)

	// create server socket
	serverFD, err := syscall.Socket(syscall.AF_INET, syscall.O_NONBLOCK|syscall.SOCK_STREAM, 0)
	logPanic(`error opening socket connection `, err)
	defer syscall.Close(serverFD)

	err = syscall.SetNonblock(serverFD, true)
	logPanic(`error setting sock to non blocking mode `, err)

	// bind socket the port
	ipv4 := net.ParseIP("0.0.0.0")
	err = syscall.Bind(serverFD, &syscall.SockaddrInet4{
		Port: 8004,
		Addr: [4]byte{ipv4[0], ipv4[1], ipv4[2], ipv4[3]},
	})
	logPanic(`error binding port to the socket`, err)

	// listen to socket
	err = syscall.Listen(serverFD, maxClients)
	logPanic(`error listening the socket `, err)

	// create epoll instance
	epollFD, err := syscall.EpollCreate1(0)
	logPanic(`error creating epoll instance `, err)

	// listen to epollin events for serverFD
	var socketServerEvent syscall.EpollEvent = syscall.EpollEvent{
		Events: syscall.EPOLLIN,
		Fd:     int32(serverFD),
	}
	err = syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, serverFD, &socketServerEvent)
	logPanic(`error listening to epoll in events for the serverFD `, err)

	lastCleanUpTime := time.Now()
	for {
		if time.Now().After(lastCleanUpTime.Add(time.Duration(10) * time.Second)) {
			for core.Clean() {
				continue
			}
			lastCleanUpTime = time.Now()
		}
		nevents, e := syscall.EpollWait(epollFD, events[:], -1)
		if e != nil {
			continue
		}
		for i := 0; i < nevents; i++ {
			if int(events[i].Fd) == serverFD { // new connection to server socket
				// accept the incomming event
				fd, _, err := syscall.Accept(serverFD)
				if err != nil {
					log.Println(`error accepting connection to server socket `, err)
					continue
				}
				log.Println("accepted connection from FD ", fd)
				conClients++
				syscall.SetNonblock(serverFD, true)

				// register the incomming client fd into epoll events
				var socketClientEvent syscall.EpollEvent = syscall.EpollEvent{
					Events: syscall.EPOLLIN,
					Fd:     int32(fd),
				}
				err = syscall.EpollCtl(epollFD, syscall.EPOLL_CTL_ADD, fd, &socketClientEvent)
				if err != nil {
					log.Println(`error listening to epoll in events for the clientFD `, err)
				}
			} else { // event from the existing connection

				conn := &core.FDCommon{FD: int64(events[i].Fd)}
				cmds, err := readCommand(conn)
				if err != nil {
					log.Println("closed connection from FD ", events[i].Fd)
					syscall.Close(int(events[i].Fd))
					conClients--
					continue
				}
				respond(cmds, conn)
			}
		}
	}

}
