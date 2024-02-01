package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
)

// SendUDPFile envía un archivo a través de una conexión UDP al servidor especificado.
func SendUDPFile(filePath string, ipConn string, portConn string) error {
	// Abre el archivo:
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer file.Close()

	// Obtiene los datos del archivo:
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("error al obtener información del archivo: %v", err)
	}

	fileName := fileInfo.Name()
	fileData, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error al leer los datos del archivo: %v", err)
	}

	// Calcula el hash SHA-256 del archivo:
	hash := sha256.Sum256(fileData)

	// Crea la estructura del mensaje:
	msg := FileMessage{
		FileName: fileName,
		Data:     fileData,
		Hash:     hash,
	}

	// Resuelve la dirección del servidor:
	serverAddr, err := net.ResolveUDPAddr("udp", ipConn+":"+portConn)
	if err != nil {
		return fmt.Errorf("error al resolver la dirección UDP: %v", err)
	}

	// Crea la conexión UDP:
	conn, err := net.DialUDP("udp", nil, serverAddr)
	if err != nil {
		return fmt.Errorf("error al establecer la conexión UDP: %v", err)
	}
	defer conn.Close()

	// Codifica y envía el mensaje al servidor:
	err = sendUDPMessage(conn, &msg)
	if err != nil {
		return fmt.Errorf("error al enviar el mensaje por UDP: %v", err)
	}

	// Respuesta del servidor:
	response := make([]byte, 1)
	_, err = conn.Read(response)
	if err != nil {
		return fmt.Errorf("error al leer la respuesta del servidor: %v", err)
	}

	if response[0] == MsgSuccess {
		fmt.Println("El archivo se guardó correctamente.")
	} else {
		return fmt.Errorf("el archivo no se pudo guardar correctamente")
	}

	return nil
}

// sendUDPMessage envía un mensaje que contiene la información de un archivo a través de una conexión UDP.
// Devuelve un error si ocurre algún problema durante el proceso.
func sendUDPMessage(conn *net.UDPConn, msg *FileMessage) error {
	// Codifica la estructura del mensaje y lo envía a través de la conexión:
	_, err := conn.Write([]byte{0}) // Indicador de inicio del mensaje
	if err != nil {
		fmt.Println("[ERROR] al enviar indicador de inicio del mensaje: ", err)
		return err
	}

	// Escribe el nombre del archivo en la conexión:
	fileNameLenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(fileNameLenBuf, uint32(len(msg.FileName)))
	_, err = conn.Write(fileNameLenBuf)
	if err != nil {
		fmt.Println("[ERROR] al enviar la longitud del nombre del archivo: ", err)
		return err
	}

	// Envía el nombre del archivo:
	_, err = conn.Write([]byte(msg.FileName))
	if err != nil {
		fmt.Println("[ERROR] al enviar el nombre del archivo: ", err)
		return err
	}

	// Envía el tamaño total del archivo:
	totalSizeBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(totalSizeBuf, uint32(len(msg.Data)))
	_, err = conn.Write(totalSizeBuf)
	if err != nil {
		fmt.Println("[ERROR] al enviar la longitud total del archivo: ", err)
		return err
	}

	// Define el tamaño del fragmento
	chunkSize := 1024 // 1024, 8192
	dataLen := len(msg.Data)

	// Enviar fragmentos del archivo
	for i := 0; i < dataLen; i += chunkSize {
		end := i + chunkSize
		if end > dataLen {
			end = dataLen
		}

		// Envía un fragmento del archivo:
		_, err = conn.Write(msg.Data[i:end])
		if err != nil {
			fmt.Println("[ERROR] al enviar fragmento del archivo: ", err)
			return err
		}
	}

	// Envía el hash del archivo en la conexión:
	_, err = conn.Write(msg.Hash[:])
	if err != nil {
		fmt.Println("[ERROR] al enviar el hash del archivo: ", err)
		return err
	}

	return nil
}
