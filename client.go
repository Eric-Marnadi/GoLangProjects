package main

import(
	"fmt"
	"net"
	"time"
	"sync"
	"github.com/golang/protobuf/proto"
)
type Queue struct{
	requests []int
}
func (q *Queue) add(rate float64){
	t := time.Now()
	for {
		if(time.Now().Sub(t).Seconds() >= rate){
			q.requests = append(q.requests, 0)
			t = time.Now()
		}
	}
}
func (q *Queue) remove(rate float64){
	t := time.Now()
	for {
		if(time.Now().Sub(t).Seconds() >= rate){
			if(tokens > 0){
				tokens--;
				sent++;
			} else {
				dropped++;
			}
			if(len(q.requests) > 0){
				q.requests = q.requests[1:]
			}
			t = time.Now()
		}
	}
}
func refillTokens(){
	for{
		if(time.Now().Sub(lastRefill).Seconds() >= 10){
			mutex.Lock()
			conn, err := net.Dial("tcp", ":1000")
			if err != nil {
				panic(err)
			}
			clientMessage := &ClientMessage{Sent: int64(sent), Dropped: int64(dropped)}
			data, err := proto.Marshal(clientMessage)
			if err != nil {
				panic(err)
			}
			sent, dropped = 0, 0
			conn.Write(data)
			buf := make([]byte, 64)
	
			n, err := conn.Read(buf)
			if err != nil {
				panic(err)
			}
			protoData := &ServerMessage{}
			err = proto.Unmarshal(buf[0:n], protoData)
			if err != nil {
				panic(err)
			}
			tokens += int(protoData.GetNumTokens())
			lastRefill = time.Now()
			conn.Close()
			mutex.Unlock()
			fmt.Println(tokens)
		}
	}
}
var tokens = 0
var lastRefill = time.Now()
var mutex = sync.RWMutex{}
var wg = sync.WaitGroup{}
var sent = 0
var dropped = 0
func main(){
	wg.Add(3)
	q := Queue{make([]int, 0, 10)}
	go q.add(3)
	go q.remove(3)
	go refillTokens()
	wg.Wait()
}
