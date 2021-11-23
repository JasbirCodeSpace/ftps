package main

import (
	"ftpserver/server/credentials"
	"crypto/tls"
	"bufio"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)
const (
	PORT = "9090"
	BUFFERSIZE = 4096
)

var  ROOT = "../filestore"
//dynamic root dir
func init(){
	cdir, _ := os.Getwd()
	splits := strings.Split(cdir, "/")
	ROOT = strings.Join(splits[:len(splits)-1],"/" )+ ROOT

}

const serverKey = `-----BEGIN EC PARAMETERS-----
BggqhkjOPQMBBw==
-----END EC PARAMETERS-----
-----BEGIN EC PRIVATE KEY-----
MHcCAQEEIBGcLjOMtmcOtWGTN/VYLFULNEDxXa3T5YI73q72HUXjoAoGCCqGSM49
AwEHoUQDQgAEAxgzx96lQzbvC+S2lNkGzFORalrFJ0jC5GWcKoSOyPOnaCRIlUwT
2lZ9IMDkicWeHpAlPLqm0UxSgPFhEtLTlw==
-----END EC PRIVATE KEY-----
`

const serverCert = `-----BEGIN CERTIFICATE-----
MIICGTCCAb+gAwIBAgIUGqDt34TV7GOlqIxUJeNcBJZuGFYwCgYIKoZIzj0EAwIw
YjELMAkGA1UEBhMCSU4xDjAMBgNVBAgMBURlbGhpMQ4wDAYDVQQHDAVEZWxoaTES
MBAGA1UECgwJbG9jYWxob3N0MQswCQYDVQQLDAJDTjESMBAGA1UEAwwJZnRwc2Vy
dmVyMB4XDTIxMTEyMzE0NTI1OVoXDTMxMTEyMTE0NTI1OVowYjELMAkGA1UEBhMC
SU4xDjAMBgNVBAgMBURlbGhpMQ4wDAYDVQQHDAVEZWxoaTESMBAGA1UECgwJbG9j
YWxob3N0MQswCQYDVQQLDAJDTjESMBAGA1UEAwwJZnRwc2VydmVyMFkwEwYHKoZI
zj0CAQYIKoZIzj0DAQcDQgAEAxgzx96lQzbvC+S2lNkGzFORalrFJ0jC5GWcKoSO
yPOnaCRIlUwT2lZ9IMDkicWeHpAlPLqm0UxSgPFhEtLTl6NTMFEwHQYDVR0OBBYE
FPmkknrazmMylpZOCXZkaNjOfad9MB8GA1UdIwQYMBaAFPmkknrazmMylpZOCXZk
aNjOfad9MA8GA1UdEwEB/wQFMAMBAf8wCgYIKoZIzj0EAwIDSAAwRQIgRUxLR7F2
hVC3+o33XbesY2K65aDExNKLJIEG/q3zeBMCIQCa5k9ijdiKSJEc5YyQGt+Nx4V/
VI3LodE91HHz1aCHig==
-----END CERTIFICATE-----
`
func main(){

	cer, err := tls.X509KeyPair([]byte(serverCert), []byte(serverKey))
	if err != nil {
		log.Fatal(err)
	}
	config := &tls.Config{Certificates: []tls.Certificate{cer}}


	server, err := tls.Listen("tcp", ":"+PORT, config)

	defer server.Close()

	if err != nil{
		log.Fatal(err)
	}
	log.Println("TCP server is UP @ localhost:", PORT)


	for{
		connection, err := server.Accept()

		if err != nil{
			log.Println("Client Connection failed")
		}

		go HandleServer(connection)
	}
}


func GetCred() *credentials.CredArr{

	f, _ := os.Open("credential.json")
	var creds credentials.CredArr
	creds.FromJSON(f)
	return &creds
}

func AuthenticateClient(conn net.Conn){
	reader := bufio.NewScanner(conn)
	validated := false
	conn.Write( []byte("Connection Established"))
	CREDS := GetCred()

	//validate user
	for !validated {
		reader.Scan()
		uname := reader.Text()
		reader.Scan()
		passwd := reader.Text()

		for _, cred := range *CREDS{
			if cred.Username == uname && cred.Password == passwd{
				validated=true
				log.Println("Client Validated")
				break
			}
		}
		if validated{
			conn.Write([]byte("1"))
			break
		}
		conn.Write([]byte("0"))
	}
}

func HandleServer(conn net.Conn){
	defer conn.Close()
	AuthenticateClient(conn)


	buffer := make([]byte, BUFFERSIZE)
	for {
		n,_ := conn.Read(buffer)
		command := strings.TrimSpace(string(buffer[:n]))
		commandArr := strings.Split(command," ")

		switch strings.ToLower(commandArr[0]) {

		case "upload":
			log.Println("UPLOAD Request")
			n, _ := conn.Read(buffer)
			fileSize,err := strconv.ParseInt(string(buffer[:n]),10, 64)
			log.Println(fileSize)
			if err!=nil || fileSize ==-1{
				log.Println(err)
				log.Println("FILE ERROR")
				continue
			}
			GetFile(conn,commandArr[1], fileSize)


		case "download":
			log.Println("Download")
			SendFile(conn, commandArr[1])


		case "ls":
			log.Println("ls")
			getDIR(conn)


		case "pwd":
			log.Println("pwd")
			conn.Write([]byte(ROOT))


		case "cd":
			changeDIR(conn, commandArr[1])

		case "delete":
			deleteDIR(conn, commandArr[1])


		case "close":
			log.Println("CLOSE")
			return
		}
	}

}