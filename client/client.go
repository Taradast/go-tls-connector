package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net"
	"os"
	"time"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		Address string `yaml:"address"`
		Port    int    `yaml:"port"`
	} `yaml:"server"`

	Client struct {
		ListenAddress string `yaml:"listen_address"`
		ListenPort    int    `yaml:"listen_port"`
	} `yaml:"client"`

	TLS struct {
		CACert     string `yaml:"ca_cert"`
		ClientCert string `yaml:"client_cert"`
		ClientKey  string `yaml:"client_key"`
	} `yaml:"tls"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var cfg Config
	err = yaml.Unmarshal(data, &cfg)
	if err != nil {
		return nil, err
	}

	return &cfg, nil
}

func formatBytes(n int64) string {
	if n < 1024 {
		return fmt.Sprintf("%d B", n)
	} else if n < 1024*1024 {
		return fmt.Sprintf("%.2f KB", float64(n)/1024)
	}
	return fmt.Sprintf("%.2f MB", float64(n)/(1024*1024))
}

func forward(localConn net.Conn, serverAddr string, tlsConfig *tls.Config) {
	defer localConn.Close()

	localAddr := localConn.RemoteAddr().String()
	log.Printf("[LOCAL] New connection from %s", localAddr)

	serverConn, err := tls.Dial("tcp", serverAddr, tlsConfig)
	if err != nil {
		log.Printf("[ERROR] Connection to server failed: %v", err)
		return
	}
	log.Printf("[SERVER] Connected to %s", serverAddr)
	defer serverConn.Close()

	bytesReport := make(chan string)

	go func() {
		n, err := io.Copy(serverConn, localConn)
		if err != nil {
			log.Printf("[ERROR] Sending to server failed: %v", err)
		}
		bytesReport <- fmt.Sprintf("%s — Sent to server: %s",
			time.Now().Format("15:04:05"), formatBytes(n))
	}()

	go func() {
		n, err := io.Copy(localConn, serverConn)
		if err != nil {
			log.Printf("[ERROR] Receiving from server failed: %v", err)
		}
		bytesReport <- fmt.Sprintf("%s — Received from server: %s",
			time.Now().Format("15:04:05"), formatBytes(n))
	}()

	for i := 0; i < 2; i++ {
		log.Println(<-bytesReport)
	}

	log.Printf("[END] Connection with %s closed", localAddr)
}

func main() {
	config, err := loadConfig("../conf/client_config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	clientAddr := fmt.Sprintf("%s:%d",
		config.Client.ListenAddress,
		config.Client.ListenPort,
	)

	serverAddr := fmt.Sprintf("%s:%d",
		config.Server.Address,
		config.Server.Port,
	)

	cert, err := tls.LoadX509KeyPair(config.TLS.ClientCert, config.TLS.ClientKey)
	if err != nil {
		log.Fatal(err)
	}

	caCertPool := x509.NewCertPool()
	caCert, err := os.ReadFile(config.TLS.CACert)
	if err != nil {
		log.Fatal(err)
	}
	caCertPool.AppendCertsFromPEM(caCert)

	tlsConfig := &tls.Config{
		Certificates:       []tls.Certificate{cert},
		RootCAs:            caCertPool,
		InsecureSkipVerify: false,
		MinVersion:         tls.VersionTLS12,
	}

	// Старт локального слушателя
	localListener, err := net.Listen("tcp", clientAddr)
	if err != nil {
		log.Fatal(err)
	}
	defer localListener.Close()

	log.Printf("[READY] Client listening on %s", clientAddr)
	log.Printf("[CONFIG] Server: %s", serverAddr)

	for {
		localConn, err := localListener.Accept()
		if err != nil {
			log.Printf("[ERROR] Accept failed: %v", err)
			continue
		}

		go forward(localConn, serverAddr, tlsConfig)
	}
}
