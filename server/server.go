package main

import (
	"crypto/tls"
	"crypto/x509"
	"fmt"
	"io"
	"log"
	"net"
	"os"

	"gopkg.in/yaml.v3"
)

type Config struct {
	Server struct {
		ListenAddress string `yaml:"listen_address"`
		ListenPort    int    `yaml:"listen_port"`
	} `yaml:"server"`

	Backend struct {
		LAddress string `yaml:"laddress"`
		Port     int    `yaml:"port"`
	} `yaml:"backend"`

	TLS struct {
		CACert     string `yaml:"ca_cert"`
		ServerCert string `yaml:"server_cert"`
		ServerKey  string `yaml:"server_key"`
	} `yaml:"tls"`
}

func loadConfig(path string) (*Config, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	var cfg Config
	if err := yaml.Unmarshal(data, &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}

func handleClient(conn net.Conn, backendAddr string) {
	defer conn.Close()

	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		log.Println("Not a TLS connection")
		return
	}

	if err := tlsConn.Handshake(); err != nil {
		log.Println("TLS handshake error:", err)
		return
	}

	state := tlsConn.ConnectionState()
	clientName := "Unknown"
	if len(state.PeerCertificates) > 0 {
		clientName = state.PeerCertificates[0].Subject.CommonName
	}

	clientIP := conn.RemoteAddr().String() // IP:port

	log.Printf("Client connected: %s [%s]", clientName, clientIP)

	// Подключение к backend
	localConn, err := net.Dial("tcp", backendAddr)
	if err != nil {
		log.Println("Error connecting to backend:", err)
		return
	}
	defer localConn.Close()

	bytesReport := make(chan string)

	go func() {
		n, _ := io.Copy(localConn, conn)
		bytesReport <- fmt.Sprintf("Client %s [%s] sent %d bytes", clientName, clientIP, n)
	}()

	go func() {
		n, _ := io.Copy(conn, localConn)
		bytesReport <- fmt.Sprintf("Client %s [%s] received %d bytes", clientName, clientIP, n)
	}()

	for i := 0; i < 2; i++ {
		log.Println(<-bytesReport)
	}

	log.Printf("Client disconnected: %s [%s]", clientName, clientIP)
}

func main() {
	config, err := loadConfig("../conf/server_config.yaml")
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	serverAddr := fmt.Sprintf("%s:%d", config.Server.ListenAddress, config.Server.ListenPort)
	backendAddr := fmt.Sprintf("%s:%d", config.Backend.LAddress, config.Backend.Port)

	// TLS
	cert, err := tls.LoadX509KeyPair(config.TLS.ServerCert, config.TLS.ServerKey)
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
		Certificates: []tls.Certificate{cert},
		ClientCAs:    caCertPool,
		ClientAuth:   tls.RequireAndVerifyClientCert,
		MinVersion:   tls.VersionTLS12,
	}

	listener, err := tls.Listen("tcp", serverAddr, tlsConfig)
	if err != nil {
		log.Fatal(err)
	}
	defer listener.Close()
	log.Printf("Server listening on %s, forwarding to backend %s", serverAddr, backendAddr)

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println("Accept error:", err)
			continue
		}
		go handleClient(conn, backendAddr)
	}
}
