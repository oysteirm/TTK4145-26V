// Use `go run foo.go` to run your program

package main

import (
    . "fmt"
    "runtime"
    //"time"
)

var incCh chan struct{}
var decCh chan struct{}
var getCh chan int
var done chan struct{}

func incrementing() {
    //TODO: increment i 1000000 times
    var j int
    for j = 0; j < 1000000; j++ {
        incCh <- struct{}{}
    }
    done <- struct{}{}
}

func decrementing() {
    //TODO: decrement i 1000000 times
    var k int
    for k = 0; k < 1000000; k++ {
        decCh <- struct{}{}
    }
    done <- struct{}{}
    
}


func server() {
    i := 0
        for {
            select {
            case <-incCh:
                i++
            case <-decCh:
                i--
            case getCh <- i:
            }
        }
}



func main() {
    // What does GOMAXPROCS do? What happens if you set it to 1?
    runtime.GOMAXPROCS(2)   

    incCh = make(chan struct{})
    decCh = make(chan struct{})
    getCh = make(chan int)
    done = make(chan struct{})
   
    // TODO: Spawn both functions as goroutines

    go server()
	go incrementing()
    go decrementing()

    // We have no direct way to wait for the completion of a goroutine (without additional synchronization of some sort)
    // We will do it properly with channels soon. For now: Sleep.
    <-done
    <-done
    Println("The magic number is:", <-getCh)
}