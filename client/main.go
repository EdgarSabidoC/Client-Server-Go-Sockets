package main

import (
	"flag"
	"fmt"
	"os"
	"strconv"
)

/**
 * Función principal
 */
func main() {
	// Se verifica la cantidad de argumentos:
	var filePath string
	if len(os.Args) < 2 {
		fmt.Println("error ruta de archivo no encontrada.")
		os.Exit(1)
	}

	// Se crean las banderas para la IP y el número de puerto:
	ip := flag.String("ip", ConnHost, "IP address")
	port := flag.String("p", strconv.Itoa(ConnPort), "Port number")
	protocol := flag.String("t", ConnType, "Protocol type")

	// Se hace el parseo de las banderas:
	flag.Parse()

	// Se obtiene la ruta del archivo:
	filePath = flag.Arg(0)

	// Se validan la ruta del archivo, la IP y el puerto:
	if !IsValidFilePath(filePath) {
		fmt.Println("Ruta del archivo no válida:", filePath)
		os.Exit(1)
	}

	if !IsValidIP(*ip) {
		fmt.Println("Dirección IP no válida:", *ip)
		os.Exit(1)
	}

	if !IsValidPort(*port, *protocol) {
		fmt.Println("Número de puerto no válido:", *port)
		os.Exit(1)
	}

	if *protocol == "tcp" {
		// Se envía el archivo por el protocolo TCP:
		err := SendTCPFile(filePath, *ip, *port)
		if err != nil {
			fmt.Println("Error al enviar el archivo:", err)
			os.Exit(1)
		}
	} else if *protocol == "udp" {
		// Se envía el archivo por el protocolo UDP:
		err := SendUDPFile(filePath, *ip, *port)
		if err != nil {
			fmt.Println("Error al enviar el archivo:", err)
			os.Exit(1)
		}
	}
}
