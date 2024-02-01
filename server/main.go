package main

import (
	"fmt"
	"net"
	"strconv"
)

func main() {
	// Lee el archivo de configuración de la conexión:
	err := ReadConfigFile("config.json")
	if err != nil {
		fmt.Println("[ERROR] al leer el archivo de configuración:", err)
		return
	}
	fmt.Println("Logs:")
	host := GlobalConfig.Host
	tcpPort := strconv.Itoa(GlobalConfig.TcpPort)
	udpPort := strconv.Itoa(GlobalConfig.UdpPort)

	// Inicia el listener del TCP con una goroutine:
	go func() {
		fmt.Println("> Arrancando servidor TCP en " + host + ":" + tcpPort)
		tcpListener, err := net.Listen("tcp", host+":"+tcpPort)
		if err != nil {
			fmt.Println("[ERROR] al iniciar el listener del protocolo TCP: ", err)
			return
		}
		defer tcpListener.Close()

		fmt.Println("\t>> Servidor TCP escuchando en puerto: " + tcpPort)

		for {
			conn, err := tcpListener.Accept()
			if err != nil {
				fmt.Println("[ERROR] aceptando conexión: ", err)
				continue
			}
			go HandleTCP(conn)
		}
	}()

	// Inicia el listener del UDP con una goroutine:
	go func() {
		fmt.Println("> Arrancando servidor UDP en " + host + ":" + udpPort)
		addr := net.UDPAddr{
			Port: GlobalConfig.UdpPort,
			IP:   net.ParseIP(host),
		}
		udpListener, err := net.ListenUDP("udp", &addr)
		if err != nil {
			fmt.Println("[ERROR] al iniciar el listener del protocolo UDP: ", err)
			return
		}
		defer udpListener.Close()

		fmt.Println("\t>> Servidor UDP escuchando en puerto: " + udpPort)

		for {
			HandleUDP(udpListener)
		}
	}()

	// Permite esperar indefinidamente:
	select {}
}
