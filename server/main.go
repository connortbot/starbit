package main

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"log"
	"math/big"
	"net"
	"time"

	pb "starbit/proto"
	"starbit/server/game"

	"github.com/quic-go/quic-go"
	"google.golang.org/grpc"
)

func generateSelfSignedCert() tls.Certificate {
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		log.Fatalf("Failed to generate private key: %v", err)
	}

	// vaid for 1 year
	notBefore := time.Now()
	notAfter := notBefore.Add(365 * 24 * time.Hour)

	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	serialNumber, err := rand.Int(rand.Reader, serialNumberLimit)
	if err != nil {
		log.Fatalf("Failed to generate serial number: %v", err)
	}

	template := x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Organization: []string{"Starbit Game"},
		},
		NotBefore:             notBefore,
		NotAfter:              notAfter,
		KeyUsage:              x509.KeyUsageKeyEncipherment | x509.KeyUsageDigitalSignature,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
		BasicConstraintsValid: true,
		DNSNames:              []string{"*"},
	}

	derBytes, err := x509.CreateCertificate(rand.Reader, &template, &template, &key.PublicKey, key)
	if err != nil {
		log.Fatalf("Failed to create certificate: %v", err)
	}

	cert := tls.Certificate{
		Certificate: [][]byte{derBytes},
		PrivateKey:  key,
	}
	return cert
}

func generateTLSConfig() *tls.Config {
	// simple, insecure TLS for non-prod only
	return &tls.Config{
		Certificates: []tls.Certificate{generateSelfSignedCert()},
		NextProtos:   []string{"starbit-quic"},
	}
}

func main() {
	// TCP setup - bind to all interfaces
	lis, err := net.Listen("tcp", "0.0.0.0:50051")
	if err != nil {
		log.Fatalf("failed to listen: %v", err)
	}
	defer lis.Close()

	// UDP setup - bind to all interfaces
	listener, err := quic.ListenAddr("localhost:50052", generateTLSConfig(), nil)
	if err != nil {
		log.Fatal("QUIC listen error:", err)
	}
	defer listener.Close()
	log.Println("QUIC server listening on localhost:50052")

	// create shared game state
	gameState := game.NewState()

	// create and start UDP server
	udpServer := game.NewUDPServer(listener)
	udpServer.SetState(gameState)

	// create and start TCP server
	tcpServer := game.NewServer()
	tcpServer.SetState(gameState)
	udpServer.SetTCPServer(tcpServer)

	go udpServer.Start()

	// Set up and start the TCP server
	s := grpc.NewServer()
	pb.RegisterGameServer(s, tcpServer)
	log.Printf("TCP server listening at %v", lis.Addr())
	if err := s.Serve(lis); err != nil {
		log.Fatalf("failed to serve: %v", err)
	}
}
