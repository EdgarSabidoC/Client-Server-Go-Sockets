package main

import (
	"crypto/sha256"
	"encoding/binary"
	"fmt"
	"io"
	"net"
	"os"
	"path/filepath"
)

// HandleTCP envuelve a handleTCPClient para manejar la recepción de archivos a través de una conexión TCP.
func HandleTCP(conn net.Conn) {
	status := handleTCPClient(conn)
	// Se envía el estado de error de la operación al cliente:
	err := sendTCPResponse(conn, status)
	if err != nil {
		fmt.Println("[ERROR] al enviar respuesta del estado de la operación al cliente: ", err)
	}
}

// handleTCPClient maneja la recepción de archivos a través de una conexión TCP.
func handleTCPClient(conn net.Conn) byte {
	// Se recibe la estructura del mensaje que contiene el nombre, los datos y el hash del archivo:
	var fileMsg FileMessage
	err := readMessage(conn, &fileMsg)
	if err != nil {
		fmt.Println("[ERROR] leyendo el mensaje:", err)
		return MsgFailure
	}

	// Crear un directorio para guardar el archivo basado en su tipo (imagen, audio, vídeo o texto):
	fileType, filePath, valid := GetFileType(fileMsg.FileName)
	if !valid {
		fmt.Println("[ERROR] extensión de archivo no válida.")
		return MsgFailure
	}

	dir := filepath.Join(filePath, fileType)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Println("[ERROR] creando directorio:", err)
		return MsgFailure
	}

	// Se verifica si el hash del archivo coincide con el enviado por el cliente:
	err = CompareHash256(sha256.Sum256(fileMsg.Data), fileMsg.Hash)
	if err != nil {
		fmt.Println("[ERROR]:", err)
		return MsgFailure
	}

	// Se crea un archivo para guardar el archivo recibido:
	outPath := filepath.Join(dir, fileMsg.FileName)
	// Se crea el archivo:
	err = CreateFile(outPath, &fileMsg)
	if err != nil {
		fmt.Println("[ERROR]:", err)
		return MsgFailure
	}

	// Si se guardó correctamente el archivo, se imprime en la terminal de logs:
	WriteLog(outPath)

	return MsgSuccess
}

// readMessage decodifica la estructura del mensaje desde la conexión TCP.
func readMessage(conn net.Conn, msg *FileMessage) error {
	// Se decodifica la estructura del mensaje desde la conexión:
	_, err := io.ReadFull(conn, []byte{0}) // Indicador de inicio del mensaje
	if err != nil {
		return err
	}

	// Se lee el nombre del archivo:
	fileNameLenBuf := make([]byte, 4)
	_, err = io.ReadFull(conn, fileNameLenBuf)
	if err != nil {
		fmt.Println(err)
		return err
	}
	fileNameLen := int(binary.BigEndian.Uint32(fileNameLenBuf))

	fileNameBuf := make([]byte, fileNameLen)
	_, err = io.ReadFull(conn, fileNameBuf)
	if err != nil {
		fmt.Println(err)
		return err
	}
	msg.FileName = string(fileNameBuf)

	// Se leen los datos del archivo:
	dataLenBuf := make([]byte, 4)
	_, err = io.ReadFull(conn, dataLenBuf)
	if err != nil {
		fmt.Println(err)
		return err
	}
	dataLen := int(binary.BigEndian.Uint32(dataLenBuf))

	msg.Data = make([]byte, dataLen)
	_, err = io.ReadFull(conn, msg.Data)
	if err != nil {
		fmt.Println(err)
		return err
	}

	// Se lee el hash del archivo:
	_, err = io.ReadFull(conn, msg.Hash[:])
	if err != nil {
		fmt.Println(err)
		return err
	}

	return nil
}

// sendTCPResponse envía un mensaje de éxito (1) o error (0) al cliente UDP.
func sendTCPResponse(conn net.Conn, status byte) error {
	_, err := conn.Write([]byte{status})
	if err != nil {
		return err
	}
	return nil
}
