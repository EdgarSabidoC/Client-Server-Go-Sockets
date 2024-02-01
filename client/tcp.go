package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
)

// sendFileTCP envía un archivo a través de una conexión TCP a una dirección IP y puerto especificados.
// Devuelve un error si ocurre algún problema durante el proceso.
func SendTCPFile(filePath string, ipConn string, portConn string) error {
	// Se abre el archivo:
	file, err := os.Open(filePath)
	if err != nil {
		return fmt.Errorf("error al abrir el archivo: %v", err)
	}
	defer file.Close()

	// Se genera la conexión:
	conn, err := net.Dial("tcp", ipConn+":"+portConn)
	if err != nil {
		return fmt.Errorf("error al establecer la conexión: %v", err)
	}
	defer conn.Close()

	// Se obtienen los datos del archivo:
	fileInfo, err := file.Stat()
	if err != nil {
		return fmt.Errorf("error al obtener información del archivo: %v", err)
	}

	fileName := fileInfo.Name()
	fileData, err := io.ReadAll(file)
	if err != nil {
		return fmt.Errorf("error al leer los datos del archivo: %v", err)
	}

	// Se calcula el hash SHA-256 del archivo:
	hash := sha256.Sum256(fileData)

	// Se crea la estructura del mensaje:
	msg := FileMessage{
		FileName: fileName,
		Data:     fileData,
		Hash:     hash,
	}

	// Se codifica y envía el mensaje al servidor:
	err = sendTCPMessage(conn, &msg)
	if err != nil {
		return fmt.Errorf("error al enviar el mensaje: %v", err)
	}

	// Lee la respuesta del servidor:
	response := make([]byte, 1)
	_, err = conn.Read(response)
	if err != nil {
		return fmt.Errorf("error al leer la respuesta del servidor: %v", err)
	}

	if response[0] == 1 {
		fmt.Println("El archivo se guardó correctamente.")
	} else {
		return fmt.Errorf("el archivo no se pudo guardar correctamente")
	}

	return nil
}

// sendTCPMessage envía un mensaje que contiene la información de un archivo a través de una conexión.
// Devuelve un error si ocurre algún problema durante el proceso.
func sendTCPMessage(conn net.Conn, msg *FileMessage) error {
	// Se codifica la estructura del mensaje y se envía a través de la conexión:
	_, err := conn.Write([]byte{0}) // Indicador de inicio del mensaje
	if err != nil {
		fmt.Println("[ERROR] al enviar indicador de inicio del mensaje: ", err)
		return err
	}

	// Se escribe el nombre del archivo en la conexión:
	fileNameLenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(fileNameLenBuf, uint32(len(msg.FileName)))
	_, err = conn.Write(fileNameLenBuf)
	if err != nil {
		fmt.Println("[ERROR] al enviar la longitud del nombre del archivo: ", err)
		return err
	}

	// Se envía el nombre del archivo:
	_, err = conn.Write([]byte(msg.FileName))
	if err != nil {
		fmt.Println("[ERROR] al enviar el nombre del archivo: ", err)
		return err
	}

	// Se envía el tamaño del buffer de los datos del archivo en la conexión:
	dataLenBuf := make([]byte, 4)
	binary.BigEndian.PutUint32(dataLenBuf, uint32(len(msg.Data)))
	_, err = conn.Write(dataLenBuf)
	if err != nil {
		fmt.Println("[ERROR] al enviar la longitud de los datos del archivo: ", err)
		return err
	}

	// Se envían los datos del archivo:
	_, err = conn.Write(msg.Data)
	if err != nil {
		fmt.Println("[ERROR] al enviar los datos del archivo: ", err)
		return err
	}

	// Se envía el hash del archivo en la conexión:
	_, err = conn.Write(msg.Hash[:])
	if err != nil {
		fmt.Println("[ERROR] al enviar el hash del archivo: ", err)
		return err
	}

	return nil
}
