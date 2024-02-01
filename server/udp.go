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

var chunkSize int = GlobalConfig.ChunkSize // Tamaño de los trozos de datos, normal 1024, 8192

// HandleUDP envuelve a handleUDPClient para manejar la recepción de archivos a través de una conexión UDP.
func HandleUDP(conn *net.UDPConn) {
	status, clientAddr := handleUDPClient(conn)
	// Se envía el estado de error de la operación al cliente:
	if clientAddr != nil && !sendUDPResponse(conn, clientAddr, status) {
		fmt.Println("[ERROR] al enviar respuesta del estado de la operación al cliente.")
	}
}

// handleUDPClient maneja la recepción de archivos a través de una conexión UDP.
func handleUDPClient(conn *net.UDPConn) (byte, *net.UDPAddr) {
	// Se recibe la estructura del mensaje que contiene el nombre, los datos y el hash del archivo:
	var fileMsg FileMessage
	_, clientAddr, err := readUDPMessage(conn, &fileMsg)
	if err != nil {
		fmt.Println("[ERROR] leyendo el mensaje:", err)
		return MsgFailure, clientAddr
	}

	// Se crea un directorio para guardar el archivo basado en su tipo (imagen, audio, vídeo o texto):
	fileType, filePath, valid := GetFileType(fileMsg.FileName)
	if !valid {
		fmt.Println("[ERROR] extensión de archivo no válida.")
		return MsgFailure, clientAddr
	}
	dir := filepath.Join(filePath, fileType)
	err = os.MkdirAll(dir, os.ModePerm)
	if err != nil {
		fmt.Println("[ERROR] creando directorio:", err)
		return MsgFailure, clientAddr
	}

	// Se verifica si el hash del archivo coincide con el enviado por el cliente:
	err = CompareHash256(sha256.Sum256(fileMsg.Data), fileMsg.Hash)
	if err != nil {
		fmt.Println(err)
		return MsgFailure, clientAddr
	}

	// Se crea un archivo para guardar el archivo recibido:
	outPath := filepath.Join(dir, fileMsg.FileName)
	// Se crea el archivo:
	err = CreateFile(outPath, &fileMsg)
	if err != nil {
		fmt.Println(err)
		return MsgFailure, clientAddr
	}

	// Si se guardó correctamente el archivo, se imprime en la terminal de logs:
	WriteLog(outPath)

	return MsgSuccess, clientAddr
}

// readUDPMessage decodifica la estructura del mensaje desde la conexión UDP, que puede contener fragmentos.
func readUDPMessage(conn *net.UDPConn, msg *FileMessage) (int, *net.UDPAddr, error) {
	// Decodificar la estructura del mensaje desde la conexión
	_, addr, err := conn.ReadFromUDP([]byte{0}) // Indicador de inicio del mensaje
	if err != nil {
		return 0, nil, err
	}

	// Se lee el nombre del archivo:
	fileNameLenBuf := make([]byte, 4)
	_, err = io.ReadFull(conn, fileNameLenBuf)
	if err != nil {
		return 0, nil, err
	}
	fileNameLen := int(binary.BigEndian.Uint32(fileNameLenBuf))

	fileNameBuf := make([]byte, fileNameLen)
	_, err = io.ReadFull(conn, fileNameBuf)
	if err != nil {
		return 0, nil, err
	}
	msg.FileName = string(fileNameBuf)

	// Se lee el tamaño total del archivo:
	totalSizeBuf := make([]byte, 4)
	_, err = io.ReadFull(conn, totalSizeBuf)
	if err != nil {
		return 0, nil, err
	}
	totalSize := int(binary.BigEndian.Uint32(totalSizeBuf))

	// Se reciben los fragmentos y se reconstruye el archivo:
	var receivedData []byte // Almacena los datos recibidos
	receivedDataSize := 0   // Variable para rastrear la cantidad total de bytes leídos

	for receivedDataSize < totalSize {
		remainingSize := totalSize - receivedDataSize
		readSize := chunkSize
		if remainingSize < chunkSize {
			readSize = remainingSize
		}

		// Se leen los datos del fragmento del archivo:
		dataBuf := make([]byte, readSize)
		_, err := io.ReadFull(conn, dataBuf)
		if err != nil {
			return receivedDataSize, nil, err
		}

		// Se agregan los datos del fragmento al archivo reconstruido:
		receivedData = append(receivedData, dataBuf...)
		receivedDataSize += len(dataBuf)
	}

	// Se lee el hash del archivo:
	_, err = io.ReadFull(conn, msg.Hash[:])
	if err != nil {
		return receivedDataSize, nil, err
	}

	// Se asignan los datos reconstruidos al mensaje:
	msg.Data = receivedData

	return receivedDataSize, addr, nil
}

// sendUDPResponse envía un mensaje de éxito (1) o error (0) al cliente UDP.
func sendUDPResponse(conn *net.UDPConn, clientAddr *net.UDPAddr, status byte) bool {
	_, err := conn.WriteToUDP([]byte{status}, clientAddr)
	if err != nil {
		fmt.Println("[ERROR] enviando respuesta al cliente UDP: ", err)
		return false
	}
	return true
}
