package main

import (
	"crypto/tls"
	"crypto/x509"
	"bufio"
	"fmt"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
)

const (
	PORT = "9090"
	HOST = "localhost"
	BUFFERSIZE =4096
)

var ROOT = "../fileStore"

//dynamic root dir
func init(){
	cdir, _ := os.Getwd()
	splits := strings.Split(cdir, "/")
	ROOT = strings.Join(splits[:len(splits)-1],"/" )+ ROOT
}

const rootCert = `-----BEGIN CERTIFICATE-----
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

	roots := x509.NewCertPool()
	ok := roots.AppendCertsFromPEM([]byte(rootCert))
	if !ok {
		log.Fatal("failed to parse root certificate")
	}
	config := &tls.Config{RootCAs: roots,InsecureSkipVerify: true}
	

	server, err := tls.Dial("tcp", HOST+":"+PORT, config)

	defer server.Close()

	if err != nil{
		log.Fatal(err)
	}

	log.Println("TCP server is Connected @ ",HOST,":", PORT)

	AuthClient(server)
	HandleClient(server)
}

func AuthClient(conn net.Conn){
	stdreader := bufio.NewReader(os.Stdin)

	buffer := make([]byte, BUFFERSIZE)
	n, _ := conn.Read(buffer)

	fmt.Println(string(buffer[:n]))
	fmt.Println("Authentication Required")

	validated := false
	for !validated{
		fmt.Printf("Username >> ")
		uname,_ := stdreader.ReadString('\n')
		fmt.Printf("Password >> ")
		passwd, _ := stdreader.ReadString('\n')
		conn.Write([]byte(uname))
		conn.Write([]byte(passwd))
		n, _ := conn.Read(buffer)

	if string(buffer[:n]) == "1"{
			fmt.Println("Authentication Successful")
			validated=true
			break
		}
		fmt.Println("Invalid credentials")
	}
}


func HandleClient(conn net.Conn){
	stdreader := bufio.NewReader(os.Stdin)
	buffer := make([]byte, BUFFERSIZE)

	for {
		fmt.Printf("ftp> ")
		cmd, _  := stdreader.ReadString('\n')
		cmd = strings.TrimSpace(strings.Trim(cmd, "\n"))
		cmdArr := strings.Split(cmd, " ")

		switch strings.ToLower(cmdArr[0]){

		case "upload":
			if len(cmdArr) ==1{
				fmt.Println("provide File name please")
				continue
			}
			conn.Write([]byte(cmd))
			SendFile(conn, cmdArr[1])


		case "download":
			if len(cmdArr) ==1{
				fmt.Println("provide File name please")
				continue
			}
			conn.Write([]byte(cmd))
			n, _ := conn.Read(buffer)
			fileSize , err := strconv.ParseInt(string(buffer[:n]), 10, 64)
			if err != nil{
				fmt.Println("ERROR: ", string(buffer[:n]))
				continue
			}
			DOWNLOAD(conn, cmdArr[1], fileSize)


		case "close":
			fmt.Println("Logging out")
			conn.Write([]byte(cmd))
			return


		case "exit":
			fmt.Println("Logging out")
			conn.Write([]byte("close"))
			return

		case "ls":
			conn.Write([]byte(cmd))
			n,_ :=  conn.Read(buffer)
			files := strings.Split(string(buffer[:n]),"|")
			fmt.Println(len(files)-1, "entities found!")
			for _, file := range files[:len(files)-1]{
				isDir, name := string(file[0]), file[1:]
				fmt.Println(isDir," | ",name)
			}


		case "pwd":
			conn.Write([]byte(cmd))
			n, _ := conn.Read(buffer)
			fmt.Println(string(buffer[:n]))


		case "cd":
			if len(cmdArr)==1{
				fmt.Println("argument required")
				continue
			}
			conn.Write([]byte(cmd))
			n, _ := conn.Read(buffer)
			fmt.Println(string(buffer[:n]))

		case "delete":
			if len(cmdArr) ==1 {
				fmt.Println("Enter file name as argument")
				continue
			}
			conn.Write([]byte(cmd))
			n, _ := conn.Read(buffer)
			fmt.Println(string(buffer[:n]))


		default:
			fmt.Println("Invalid Command, Supported: PWD | LS | CD | UPLOAD | DOWNLOAD | CLOSE | DELETE")
		}
	}
}
