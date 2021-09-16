package main

import (
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"time"
	"encoding/json"
)

//Variáveis globais interessantes para o processo
var err string
var myPort string          //porta do meu servidor
var nServers int           //qtde de outros processo
var CliConn []*net.UDPConn //vetor com conexões para os servidores
var SharedResource *net.UDPConn // Endereço da conexão com o SharedResource
//dos outros processos
var ServConn *net.UDPConn //conexão do meu servidor (onde recebo
//mensagens dos outros processos)

var processId int					// id do processo utilizado
var logicalClock int = 0 	// relógio lógico local do processo, inicia com 0
var T int                 // Clock da ultima requisição enviada por este processo
var State string = "RELEASED"
var NumReceivedReplies int
var gotAllReplies bool = false
var RequestsQueue [] int

type MailStruct struct {
		Id int								// Identificador do processo
    SenderClock int       // Clock do sender, clock do processo
		Message string				// Mensagem a ser enviada
}

// Instância de struct MailStruct que será utilizada para enviar mensagens
// nesse processo
var MyMail MailStruct

func CheckError(err error) {			// Funções de captura e plot de erros
	if err != nil {
		fmt.Println("Erro: ", err)
		os.Exit(0)
	}
}
func PrintError(err error) {			// Funções de captura e plot de erros
	if err != nil {
		fmt.Println("Erro: ", err)
	}
}

// A função "Max" será utilizada no protocolo de relógio lógico
func Max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func SendReply(DestinationProcess int){
  MyMail.SenderClock = logicalClock
  MyMail.Message = "REPLY"
  doClientJob(DestinationProcess-1, MyMail)
  MyMail.Message = ""
}

func doServerJob() {
	//Loop infinito mesmo
	buf := make([]byte, 1024)
	for {
		n, addr, err := ServConn.ReadFromUDP(buf)

		if err != nil {
			fmt.Println("Error: ", err)
		}

		ReceivedMail := string(buf[0:n])
		fmt.Printf("Mail Received: %s | from %s\n", ReceivedMail, addr)

		// Essa variável auxiliar serve para transformar a string do json em uma struct
		var auxReceivedMail MailStruct
		json.Unmarshal([]byte(ReceivedMail), &auxReceivedMail)

		// O relógio lógico do processo é atualizado
		logicalClock = Max(auxReceivedMail.SenderClock, logicalClock) + 1

		fmt.Printf("LogicalClock updated to: %d\n", logicalClock)

    if auxReceivedMail.Message == "REQUEST" {
      if (State == "HELD") || (State == "WANTED" && T < auxReceivedMail.SenderClock) {
        if (State == "WANTED") {
          fmt.Printf("My state is: %s | My T is: %d | The Sender T is: %d | So, i will reply LATER the process %d\n", State, T, auxReceivedMail.SenderClock, auxReceivedMail.Id)
        }
        RequestsQueue = append(RequestsQueue, auxReceivedMail.Id)
        fmt.Printf("RequestsQueue: %v\n", RequestsQueue)
      } else {
        	time.Sleep(time.Second * 1) // Tive que botar esse sleep para conseguir cair no caso abaixo digitando no terminal =)
          if (State == "WANTED") {
            fmt.Printf("My state is: %s | My T is: %d | The Sender T is: %d | So, i will reply NOW the process %d\n", State, T, auxReceivedMail.SenderClock, auxReceivedMail.Id)
          }
          SendReply(auxReceivedMail.Id)
      }
    }

    if auxReceivedMail.Message == "REPLY" {
      NumReceivedReplies += 1
      fmt.Printf("NumReceivedReplies: %d\n", NumReceivedReplies)
    }
    if NumReceivedReplies == nServers - 1 {
      gotAllReplies = true
    }

  }
}

func doClientJob(otherProcess int, ReceivedClock MailStruct) {
	//Enviar uma mensagem (com valor i) para o servidor do processo
	//otherServer
	msg,_ := json.Marshal(ReceivedClock)
	fmt.Printf("Sending mail: %s | to process: %d\n",string(msg), otherProcess+1)
	//msg := strconv.Itoa(ReceivedClock)
	buf := []byte(msg)
	//_, err := Conn.Write(buf)
	_, err := CliConn[otherProcess].Write(buf)
	if err != nil {
		fmt.Println(msg, err)
	}
}

func initConnections() {
	//myPort = os.Args[1]
	// Recebe o id e converte a string para int
	processId, _ = strconv.Atoi(os.Args[1])
	//CheckError(err)

	myPort = os.Args[processId+1]
	fmt.Printf("processId: %d\nPort: %s\n", processId, myPort)
	nServers = len(os.Args) - 2
	/*Esse 2 tira o nome (no caso Process) e tira a primeira porta (que
	  é a minha). As demais portas são dos outros processos*/
	CliConn = make([]*net.UDPConn, nServers)

	//Outros códigos para deixar ok a conexão do meu servidor (onde re-
	//cebo msgs). O processo já deve ficar habilitado a receber msgs.

	ServerAddr, err := net.ResolveUDPAddr("udp", "127.0.0.1"+myPort)
	CheckError(err)
	ServConn, err = net.ListenUDP("udp", ServerAddr)
	CheckError(err)
	//Outros códigos para deixar ok as conexões com os servidores dos
	//outros processos. Colocar tais conexões no vetor CliConn.
	for servidores := 0; servidores < nServers; servidores++ {
		ServerToConnect := os.Args[2+servidores]
		fmt.Printf("Connected with: %s \n", ServerToConnect)
		ServerAddr, err := net.ResolveUDPAddr("udp","127.0.0.1"+ServerToConnect)
		CheckError(err)
		Conn, err := net.DialUDP("udp", nil, ServerAddr)
		CliConn[servidores] = Conn
		CheckError(err)
	}
  ServerAddr, err = net.ResolveUDPAddr("udp", "127.0.0.1:10000")
  PrintError(err)
  SharedResource, err = net.DialUDP("udp", nil, ServerAddr)
  PrintError(err)
}

func readInput(ch chan string) {
	// Rotina não-bloqueante que “escuta” o stdin
	reader := bufio.NewReader(os.Stdin)
	for {
		text, _, _ := reader.ReadLine()
		ch <- string(text)
	}
}

func updateLogicalClock(ReceivedClock int) {
	if ReceivedClock > logicalClock {
		logicalClock = ReceivedClock + 1
	} else {
		logicalClock = logicalClock + 1
	}
}

func PrintState(){
  fmt.Println("State is: ", State)
}

func ReplyQueuedRequests() {
  for requestNum := 0; requestNum < len(RequestsQueue); requestNum++ {
    SendReply(RequestsQueue[requestNum])
  }
  RequestsQueue = nil
}

func UseCS() {
  State = "HELD"
  PrintState()

  fmt.Printf("Entrei na CS\n")

  MyMail.Message = "Hello, Peter!"
  MyMail.SenderClock = logicalClock
  T = logicalClock
  msg,_ := json.Marshal(MyMail)
  fmt.Printf("Enviado: %s para SharedResource\n",string(msg))

  buf := []byte(msg)
  //_, err := Conn.Write(buf)
  _, err := SharedResource.Write(buf)
  if err != nil {
    fmt.Println(msg, err)
  }
  time.Sleep(time.Second*6)

  State = "RELEASED"
  PrintState()
  fmt.Printf("Saí da CS\n")
  logicalClock += 1

  ReplyQueuedRequests()

}

func TryToUseCS() {
  gotAllReplies = false // Inicializa o contador de replies
  NumReceivedReplies = 0
  State = "WANTED"  // Atualiza o estado
  PrintState()

  logicalClock += 1
  T = logicalClock
  MyMail.SenderClock = logicalClock
  MyMail.Message = "REQUEST"
  for DestinationProcess:=0;DestinationProcess<nServers;DestinationProcess++{
    if DestinationProcess + 1 != MyMail.Id {
      doClientJob(DestinationProcess, MyMail)
    }
  }
  //gotAllReplies = true
  for (!gotAllReplies) {
  // Aguarda a var gotAllReplies ser atualizada pelo servidor, que roda em paralelo
  }

  UseCS()
}


func main() {
	initConnections()
	//O fechamento de conexões deve ficar aqui, assim só fecha
	//conexão quando a main morrer
	defer ServConn.Close()
	for i := 0; i < nServers; i++ {
		defer CliConn[i].Close()
	}

	fmt.Printf("LogicalClock: %d\n", logicalClock)
  //fmt.Printf("State: %s\n", State)

  // Identificador do remetente das mensagens
	MyMail.Id = processId

	ch := make(chan string) //canal que guarda itens lidos do teclado
	go readInput(ch)        //chamar rotina que ”escuta” o teclado

	go doServerJob()
	for { // Verificar (de forma não bloqueante) se tem algo no // stdin (input do terminal)
		select {
		case x, valid := <-ch:
      if valid {
        if (x == "x") {
          if State != "RELEASED" {
            fmt.Printf("x ignorado\n")
          } else {
            fmt.Printf("State is: %s\n", State)
            // fazer broadcast de request e aguardar as respostas
            go TryToUseCS()
          }

        } else {
          DestinationProcess, _ := strconv.Atoi(x)
          // Se estiver se enviando mensagem apenas atualiza o clock
          if (DestinationProcess == MyMail.Id) {
            logicalClock += 1
            fmt.Printf("LogicalClock updated to %d\n", logicalClock)
          } else {
          // Caso contrário envia mensagem para outros processos
            logicalClock += 1
            MyMail.SenderClock = logicalClock
            MyMail.Message = ""
            fmt.Printf("LogicalClock updated to %d\n", logicalClock)
            doClientJob(DestinationProcess-1, MyMail)
          }
        }

			} else {
				fmt.Println("Canal fechado!")
			}
		default:
			// Fazer nada...
			// Mas não fica bloqueado esperando o teclado
			time.Sleep(time.Second * 1)
		}
		// Esperar um pouco
		time.Sleep(time.Second * 1)
	}
}
