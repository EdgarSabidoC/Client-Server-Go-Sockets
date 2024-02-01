package main

import (
	"net"
	"os"
)

// Datos de conexión predeterminados:
const (
	ConnHost = "localhost"
	ConnPort = 8080
	ConnType = "tcp"
)

const (
	MsgSuccess = 1
	MsgFailure = 0
)

// FileMessage representa la estructura del mensaje multimedia.
type FileMessage struct {
	FileName string
	Data     []byte
	Hash     [32]byte
}

// IsValidIP verifica si la cadena proporcionada es una dirección IP válida o "localhost".
// Devuelve true si es válida, de lo contrario, devuelve false.
func IsValidIP(ip string) bool {
	if ip == "localhost" {
		return true
	}
	parsedIP := net.ParseIP(ip)
	return parsedIP != nil
}

// IsValidPort verifica si el puerto proporcionado es un puerto válido para el tipo de conexión especificado.
// Devuelve true si es válido, de lo contrario, devuelve false.
func IsValidPort(port string, connType string) bool {
	if connType == "tcp" {
		_, err := net.ResolveTCPAddr("tcp", ":"+port)
		return err == nil
	} else if connType == "udp" {
		_, err := net.ResolveUDPAddr("udp", ":"+port)
		return err == nil
	}
	return false
}

// IsValidFilePath verifica si la ruta del archivo proporcionada es válida y existe.
// Devuelve true si es válida y existe, de lo contrario, devuelve false.
func IsValidFilePath(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}
