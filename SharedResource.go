package main

import (
	"fmt"
	"net"
	"os"
//	"strings"
)

func CheckError(err1 error) {
	if err1 != nil {
		fmt.Println("Erro: ", err1)
		os.Exit(0)
	}
}

func main() {
	Address, err := net.ResolveUDPAddr("udp", ":10000")
	CheckError(err)
	Connection, err := net.ListenUDP("udp", Address)
	CheckError(err)
	defer Connection.Close()
  buf := make([]byte, 1024)
	for {
		//Loop infinito para receber mensagem e escrever todo
		//conteúdo (processo que enviou, relógio recebido e texto)
		//na tela
    n, addr, err := Connection.ReadFromUDP(buf)
    fmt.Println("Received mail:", string(buf[0:n]), " from ", addr)

    if err != nil {
      fmt.Println("Error: ", err)
    }
	}
}
