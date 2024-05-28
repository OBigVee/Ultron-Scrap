package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"

	"time"

	"github.com/anthdm/hollywood/actor"
	"golang.org/x/net/html"
)


type VisitReq struct{
    links []string
}

type Manager struct{}

func NewManger() actor.Producer{
    return func () actor.Receiver {
        return &Manager{}
    }
}

func (m *Manager) Receive(c *actor.Context){
    switch msg := c.Message().(type){
    case VisitReq:
        m.handlevisitReqeust(msg)
    case actor.Started:
        slog.Info("manager started")
    case actor.Stopped:
    }
}

func (m *Manager) handlevisitReqeust(msg VisitReq) error {
    for _, link := range msg.links  {
        slog.Info("(Visiting url)", "url", link)
    }
    return nil
}


func extracLinks(body io.Reader) []string {
    links := make([]string, 0)
    tokenizer := html.NewTokenizer(body)

    for {
        tokenType := tokenizer.Next()
        if tokenType == html.ErrorToken {
            return links
        }

        if tokenType == html.StartTagToken {
            token := tokenizer.Token()
            if token.Data == "a" {
                for _, attr := range token.Attr {
                    if attr.Key == "href" {
                        links = append(links, attr.Val)
                    }
                }
            }
        }
    }
 //   return links
}
func main() {
    fmt.Println("Ultron! Do you read the world?")
    baseUrl, err :=  url.Parse("https://github.com/obigvee")
    if err != nil{
        log.Fatal(err)
    }
    resp, err := http.Get(baseUrl.String())
    if err != nil{
        log.Fatal(err)
    }
    // b, err := io.ReadAll(resp.Body)

    // if err != nil{
    //     log.Fatal(err)
    // }
    links := extracLinks((resp.Body))
    for _, link := range links{
        linksUrl, err :=  url.Parse(link)
        if err != nil{
            log.Fatal(err)
        }
        actualLink := baseUrl.ResolveReference(linksUrl)
        fmt.Println(actualLink)

    }

    // fmt.Println(extracLinks(resp.Body))
    return 

    e, err := actor.NewEngine(actor.NewEngineConfig())
    if err != nil {
        log.Fatal(err)
    }

    // spawn it amd return a PID
    pid := e.Spawn(NewManger(), "manager")
    time.Sleep(time.Millisecond * 500)
    // send a message to the pid
    e.Send(pid, VisitReq{links: []string{"https://obigvee.com"}})


    // e.SpawnFunc(func (c *actor.Context)  {
    //     switch msg := c.Message().(type){
    //     case actor.Started:
    //         fmt.Println("started")
    //         _ = msg
    //     }
    // }, "foo")

        time.Sleep(time.Second * 10)
}