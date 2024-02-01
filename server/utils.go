package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"
	"path/filepath"
	"time"
)

// MsgSuccess y MsgFailure representan códigos de mensaje para indicar el estado de una operación.
const (
	MsgSuccess = 1 // 1 indica una operación exitosa
	MsgFailure = 0 // 0 indica una falla durante la operación
)

// ConnConfig contiene la configuración del servidor.
type ConnConfig struct {
	Host            string   `json:"ip"`              // Dirección IP del servidor
	TcpPort         int      `json:"tcpPort"`         // Puerto TCP del servidor
	UdpPort         int      `json:"udpPort"`         // Puerto UDP del servidor
	ChunkSize       int      `json:"chunkSize"`       // Tamaño del fragmento para transferencias de archivos
	ImagePath       string   `json:"imagePath"`       // Ruta para archivos de imágenes
	AudioPath       string   `json:"audioPath"`       // Ruta para archivos de audio
	VideoPath       string   `json:"videoPath"`       // Ruta para archivos de video
	TextPath        string   `json:"textPath"`        // Ruta para archivos de texto
	ImageExtensions []string `json:"imageExtensions"` // Extensiones de archivos de imágenes permitidas
	AudioExtensions []string `json:"audioExtensions"` // Extensiones de archivos de audio permitidas
	VideoExtensions []string `json:"videoExtensions"` // Extensiones de archivos de video permitidas
	TextExtensions  []string `json:"textExtensions"`  // Extensiones de archivos de texto permitidas
}

// GlobalConfig contiene la configuración global del servidor.
var GlobalConfig = ConnConfig{
	Host:            "localhost", // Dirección IP predeterminada
	TcpPort:         8080,        // Puerto TCP predeterminado
	UdpPort:         8000,        // Puerto UDP predeterminado
	ChunkSize:       1024,        // Tamaño predeterminado del fragmento
	ImagePath:       "Multimedia/Images",
	AudioPath:       "Multimedia/Audios",
	VideoPath:       "Multimedia/Videos",
	TextPath:        "Multimedia/Texts",
	ImageExtensions: []string{".jpg", ".jpeg", ".png"},
	AudioExtensions: []string{".mp3", ".wav", ".mid"},
	VideoExtensions: []string{".mp4", ".avi", ".flv"},
	TextExtensions:  []string{".txt"},
}

// FileMessage representa un mensaje multimedia.
type FileMessage struct {
	FileName string   // Nombre del archivo
	Data     []byte   // Datos del archivo
	Hash     [32]byte // Hash del archivo
}

// contains verifica si un valor está presente en un slice de strings.
func contains(value string, array []string) string {
	for _, v := range array {
		if value == v {
			return v
		}
	}
	return ""
}

// GetFileType devuelve la extensión, la ruta y la validez del tipo de archivo.
func GetFileType(fileName string) (string, string, bool) {
	ext := filepath.Ext(fileName)
	filePath := ""
	valid := true

	switch ext {
	case contains(ext, GlobalConfig.ImageExtensions):
		filePath = GlobalConfig.ImagePath
	case contains(ext, GlobalConfig.AudioExtensions):
		filePath = GlobalConfig.AudioPath
	case contains(ext, GlobalConfig.VideoExtensions):
		filePath = GlobalConfig.VideoPath
	case contains(ext, GlobalConfig.TextExtensions):
		filePath = GlobalConfig.TextPath
	default:
		valid = false
	}
	return ext, filePath, valid
}

// SetIfNotEmpty asigna el valor src a dest si src no está vacío.
func SetIfNotEmpty(dest *string, src string) {
	if src != "" {
		*dest = src
	}
}

// SetIfNotEmptyExtensions reemplaza el contenido de dest con los elementos de src si src no es nil.
func SetIfNotEmptyExtensions(dest *[]string, src []string) {
	if src != nil {
		*dest = append([]string(nil), src...)
	}
}

// SetIfNotEmptyInt asigna el valor src a dest si src no es 0.
func SetIfNotEmptyInt(dest *int, src int) {
	if src != 0 {
		*dest = src
	}
}

// GetLocalIP devuelve la dirección IP local de la máquina.
func GetLocalIP() (string, error) {
	addrs, err := net.InterfaceAddrs()
	if err != nil {
		return "", err
	}

	for _, addr := range addrs {
		ipNet, ok := addr.(*net.IPNet)
		if ok && !ipNet.IP.IsLoopback() && ipNet.IP.To4() != nil {
			return ipNet.IP.String(), nil
		}
	}

	return "", fmt.Errorf("no se encontró una dirección IP local")
}

// ReadConfigFile lee el archivo de configuración JSON o establece valores predeterminados si no existe.
func ReadConfigFile(filename string) error {
	// Intenta abrir el archivo de configuración JSON:
	file, err := os.Open(filename)
	if err != nil {
		// Si hay un error al abrir el archivo se utilizan valores predeterminados:
		file.Close()
		return nil
	}
	defer file.Close()

	// Decodifica el contenido del archivo en una estructura de configuración:
	var config ConnConfig
	err = json.NewDecoder(file).Decode(&config)
	if err != nil {
		return err
	}

	// Establece valores globales basados en la configuración leída o utiliza valores predeterminados:
	SetIfNotEmpty(&GlobalConfig.Host, config.Host)
	if GlobalConfig.Host == "" {
		// Obtiene la dirección IP local si no se especifica en la configuración:
		localIP, err := GetLocalIP()
		if err != nil {
			return err
		}
		GlobalConfig.Host = localIP
	}
	SetIfNotEmptyInt(&GlobalConfig.TcpPort, config.TcpPort)
	SetIfNotEmptyInt(&GlobalConfig.UdpPort, config.UdpPort)
	SetIfNotEmptyInt(&GlobalConfig.ChunkSize, config.ChunkSize)
	SetIfNotEmpty(&GlobalConfig.ImagePath, config.ImagePath)
	SetIfNotEmpty(&GlobalConfig.AudioPath, config.AudioPath)
	SetIfNotEmpty(&GlobalConfig.VideoPath, config.VideoPath)
	SetIfNotEmpty(&GlobalConfig.TextPath, config.TextPath)
	SetIfNotEmptyExtensions(&GlobalConfig.ImageExtensions, config.ImageExtensions)
	SetIfNotEmptyExtensions(&GlobalConfig.AudioExtensions, config.AudioExtensions)
	SetIfNotEmptyExtensions(&GlobalConfig.VideoExtensions, config.VideoExtensions)
	SetIfNotEmptyExtensions(&GlobalConfig.TextExtensions, config.TextExtensions)

	return nil
}

// WriteLog escribe un mensaje en la terminal con una marca de tiempo.
func WriteLog(filePath string) {
	fmt.Print("[" + time.Now().Format("2006-01-02 15:04:05") + "] ")
	fmt.Println("Archivo subido exitosamente:", filePath)
}

// CompareHash256 compara dos hashes SHA-256 (32 bytes).
func CompareHash256(hash1, hash2 [32]byte) error {
	if hash1 != hash2 {
		return fmt.Errorf("verificación de hash fallida")
	}
	return nil
}

// CreateFile crea un archivo a partir del mensaje enviado por el cliente.
func CreateFile(outPath string, fileMsg *FileMessage) error {
	out, err := os.Create(outPath)
	if err != nil {
		out.Close() // Cierra la conexión del archivo
		return fmt.Errorf("al crear archivo en el servidor")
	}

	// Escribe los datos del archivo en el archivo:
	_, err = out.Write(fileMsg.Data)
	if err != nil {
		out.Close() // Cierra la conexión del archivo
		return fmt.Errorf("al escribir datos del archivo")
	}
	defer out.Close()
	return nil
}
