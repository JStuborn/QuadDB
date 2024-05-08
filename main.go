package main

import (
	"CyberDefenseEd/QuadDB/routes"
	"CyberDefenseEd/QuadDB/util"
	"crypto/rand"
	"crypto/sha256"
	"flag"
	"fmt"
	"os"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v3"
)

type Config struct {
	Port    int    `yaml:"port"`
	DataDir string `yaml:"data_dir"`
	AESKey  string `yaml:"aes_key"`
}

func main() {
	var config Config
	configFile := "./config/config.yaml"
	if _, err := os.Stat(configFile); err == nil {
		data, err := os.ReadFile(configFile)
		if err != nil {
			util.Error("Error reading config file:", err)
			return
		}
		err = yaml.Unmarshal(data, &config)
		if err != nil {
			util.Error("Error parsing config file:", err)
			return
		}

		util.Info("Found a valid config file, defaulting to that!")
	} else {
		config.Port = 9010
		config.DataDir = "./data"
		config.AESKey = ""
	}

	port := flag.Int("port", config.Port, "Port number")
	dataDir := flag.String("data-dir", config.DataDir, "Directory to store data files")
	aesKey := flag.String("aes-key", config.AESKey, "AES encryption key")
	generateAESKey := flag.Bool("generate-aes-key", false, "Generate a new AES key")
	flag.Parse()

	if *aesKey == "" {
		util.Error("We need an AES key to encrypt our database!")
		os.Exit(0)
	}

	if *generateAESKey {
		key, err := generateKey()
		if err != nil {
			util.Error("Error generating password:", err)
			return
		}
		*aesKey = fmt.Sprintf("%x", key)
		err = writeKeyToFile(*aesKey, "aes.temp.txt")
		if err != nil {
			util.Error("Error writing AES key to file:", err)
			return
		}

		util.Info(fmt.Sprintf("Generated password: %x\nWe saved this key in password.temp.txt, incase you need it again.\n**You will not be able to recover any data without this key!**", key))
		os.Exit(0)
	}

	// Hash the AES key using SHA-256 to allow all strings as keys
	hash := sha256.New()
	hash.Write([]byte(*aesKey))
	aesKeyBytes := hash.Sum(nil)

	util.Info(fmt.Sprintf("Using key hash - %x", aesKeyBytes))

	if len(aesKeyBytes) != 16 && len(aesKeyBytes) != 24 && len(aesKeyBytes) != 32 {
		util.Error("Invalid AES key length. Key must be 16, 24, or 32 bytes long.", aesKey)
		return
	}

	err := os.MkdirAll(*dataDir, 0755)
	if err != nil {
		panic(err)
	}

	gin.SetMode(gin.ReleaseMode)
	router := gin.New()

	// Return 500s instead of panic exiting because it kills process :sad:
	router.Use(gin.Recovery())

	util.Info("Creating routes...")
	routes.SetupRoutes(router, *dataDir, aesKeyBytes)
	routes.RegisterSwaggerRoutes(router)

	util.Info(fmt.Sprintf("Quad-Server Started - 127.0.0.1:%d", *port))
	if err := router.Run(fmt.Sprintf(":%d", *port)); err != nil {
		fmt.Printf("Error running server: %v\n", err)
		return
	}
}

func generateKey() ([]byte, error) {
	key := make([]byte, 32) // 32 bytes is a good default for AES keys
	_, err := rand.Read(key)
	return key, err
}

func writeKeyToFile(key string, filename string) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()
	_, err = f.WriteString(key)
	return err
}
