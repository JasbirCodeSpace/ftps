package main

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net"
	"os"
	"strconv"
	"strings"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/hex"
)

func getDIR(conn net.Conn){
	files, err := ioutil.ReadDir(ROOT)

	if err!=nil{
		conn.Write([]byte(err.Error()))
		fmt.Println(err)
		return
	}
	fileINFO := ""
	for _, file := range files{
		if file.IsDir() {
			fileINFO += "D" + file.Name() + "|"
		}else{
			fileINFO += "F" + file.Name()+"|"
		}

	}
	conn.Write([]byte(fileINFO))
}


func changeDIR(conn net.Conn, dir string){
	if dir==".."{
		splits := strings.Split(ROOT,"/")
		ROOT = strings.Join(splits[:len(splits)-1],"/" )
		fmt.Println(ROOT)
	}else{
		tempROOT := ""
		if ROOT == "/"{
			tempROOT = ROOT + dir
		}else{
			tempROOT = ROOT  + "/"+ dir
		}
		fmt.Println(tempROOT)
		_, err := ioutil.ReadDir(tempROOT)
		if err!=nil{
			conn.Write([]byte(err.Error()))
			return
		}
		ROOT = tempROOT
	}
	conn.Write([]byte(ROOT))
}

func SendFile(conn net.Conn, name string){
	// key, _ := ioutil.ReadFile("key.txt")
	// fmt.Println(key)
	inputFile, err := os.Open(ROOT + "/" + name)
	defer inputFile.Close()

	if err !=nil {
		conn.Write([]byte(err.Error()))
		return
	}else{
		stats,_ := inputFile.Stat()
		//send file Size
		conn.Write([]byte(strconv.FormatInt(stats.Size(),10)))
	}
	buffer := make([]byte, BUFFERSIZE)
	for {
		_, err := inputFile.Read(buffer)
		if err == io.EOF{
			break
		}
		conn.Write(buffer)
	}
	fmt.Println("File Sent")

}

func GetFile(conn net.Conn, name string, fileSize int64){
	// key, _ := ioutil.ReadFile("key.txt")
	// fmt.Println(key)
	outputFile, err := os.Create("../filestore/serverDir/" + name)

	if err != nil {
		fmt.Println(err)
	}

	defer outputFile.Close()
	var fileSizeReceived int64
	for {
		if (fileSize - fileSizeReceived) < BUFFERSIZE {
			io.CopyN(outputFile, conn, (fileSize - fileSizeReceived))
			conn.Read(make([]byte, (fileSizeReceived+BUFFERSIZE)-fileSize))
			break
		}
		io.CopyN(outputFile, conn, BUFFERSIZE)
		fileSizeReceived += BUFFERSIZE
	}
	fmt.Println("File Received successfully")
}

func deleteDIR(conn net.Conn, name string){
	fmt.Println(ROOT + "/" +name)
	err := os.Remove(ROOT + "/" +name)
	if err != nil{
		conn.Write([]byte(err.Error()))
		log.Println(err)
		return
	}
	conn.Write([]byte("File Successfully Deleted"))
}

func encrypt(stringToEncrypt string, keyString string) (encryptedString string) {

	//Since the key is in string, we need to convert decode it to bytes
	key, _ := hex.DecodeString(keyString)
	plaintext := []byte(stringToEncrypt)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}
	
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	//Create a nonce. Nonce should be from GCM
	nonce := make([]byte, aesGCM.NonceSize())
	if _, err = io.ReadFull(rand.Reader, nonce); err != nil {
		panic(err.Error())
	}

	//Encrypt the data using aesGCM.Seal
	//Since we don't want to save the nonce somewhere else in this case, we add it as a prefix to the encrypted data. The first nonce argument in Seal is the prefix.
	ciphertext := aesGCM.Seal(nonce, nonce, plaintext, nil)
	return fmt.Sprintf("%x", ciphertext)
}

func decrypt(encryptedString string, keyString string) (decryptedString string) {

	key, _ := hex.DecodeString(keyString)
	enc, _ := hex.DecodeString(encryptedString)

	//Create a new Cipher Block from the key
	block, err := aes.NewCipher(key)
	if err != nil {
		panic(err.Error())
	}

	//Create a new GCM
	aesGCM, err := cipher.NewGCM(block)
	if err != nil {
		panic(err.Error())
	}

	//Get the nonce size
	nonceSize := aesGCM.NonceSize()

	//Extract the nonce from the encrypted data
	nonce, ciphertext := enc[:nonceSize], enc[nonceSize:]

	//Decrypt the data
	plaintext, err := aesGCM.Open(nil, nonce, ciphertext, nil)
	if err != nil {
		panic(err.Error())
	}

	return fmt.Sprintf("%s", plaintext)
}