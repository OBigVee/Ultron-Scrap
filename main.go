package main

import (
	"fmt"
	"log"
	"time"

	"github.com/anthdm/hollywood/actor"
)

func main() {
    fmt.Println("Ultron! Do you read the world?")

    e, err := actor.NewEngine(actor.NewEngineConfig())
    if err != nil {
        log.Fatal(err)
    }

    e.SpawnFunc(func (c *actor.Context)  {
        switch msg := c.Message().(type){
        case actor.Started:
            fmt.Println("started")
            _ = msg
        }
    }, "foo")
        time.Sleep(time.Second * 10)
}