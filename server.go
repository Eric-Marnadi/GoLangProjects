package main

import(
	"fmt"
	"net"
	"time"
	"sync"
	"github.com/golang/protobuf/proto"
)

type Client struct{
	tokensUsed int
	lastRefill time.Time
}
func handleConnection(conn net.Conn){
	data := make([]byte, 64)
	
	n, err := conn.Read(data)
	if err != nil {
		panic(err)
	}
	protoData := &ClientMessage{}
	if err := proto.Unmarshal(data[0:n], protoData); err != nil {
		panic(err)
	}
	sent := protoData.GetSent()
	dropped := protoData.GetDropped()
	tokensNeeded := int(sent + dropped)
	_, exists := clients[conn.RemoteAddr()]
	if(!exists){
		clients[conn.RemoteAddr()] = Client{0, time.Now()}
	}
	client := clients[conn.RemoteAddr()]
	
	if(tokens > tokensNeeded && (client.tokensUsed <= 10 || time.Now().Sub(client.lastRefill).Seconds() >= 10)){
		client.tokensUsed += tokensNeeded
		tokens -= tokensNeeded
		client.lastRefill = time.Now()
		serverMessage := &ServerMessage{NumTokens: int64(tokensNeeded)}
		data, err := proto.Marshal(serverMessage)
		if err != nil {
			panic(err)
		}
		conn.Write(data)
	}
}
func refillTokens(){
	if(time.Now().Sub(lastRefill).Seconds() >= 60){
		mutex.Lock()
		tokens = 100
		lastRefill = time.Now()
		fmt.Println("Refilled 100 tokens")
		mutex.Unlock()
	}
}
var clients = make(map[net.Addr]Client)
var tokens = 100
var lastRefill = time.Now()
var mutex = sync.RWMutex{}
func main(){
	
	PORT := ":1000"
	listen, err := net.Listen("tcp", PORT)
	if err != nil {
		fmt.Println(err)
	}
	defer listen.Close()
	go func(){
		for{
			if(time.Now().Sub(lastRefill).Seconds() >= 60){
				mutex.Lock()
				tokens = 100
				lastRefill = time.Now()
				fmt.Println("Refilled 100 tokens")
				mutex.Unlock()
			}
		}
	}()
	for {
		conn, err := listen.Accept()
		if err != nil {
			fmt.Println(err)
		}
		go handleConnection(conn)		
	}

}
