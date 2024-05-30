package main

import (
	// "fmt"
	//"context"
	"bytes"
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


type VisitFunc func (io.Reader) error

type Visit struct{
    link string
    visitFunc VisitFunc
}

type VisitReq struct{
    links [] string
    visitFunc VisitFunc
}

func NewVisitRequest(links []string) VisitReq  {
    return VisitReq{
        links: links,
        visitFunc: func(r io.Reader) error {
            fmt.Println("========================")
            fmt.Println("doing the visitFunc over the body ")
            b, err := io.ReadAll(r)
            if err != nil{
                return err
            }
            fmt.Println(string(b))
            fmt.Println("===================")
            return nil
        },
    }
}

type Visitor struct{
    managerPID *actor.PID
    URL *url.URL
    visitFn VisitFunc
}

func NewVisitor(url *url.URL, mpid *actor.PID, visitFn VisitFunc) actor.Producer{
    //baseURL, err := url.Parse(link)

    return func () actor.Receiver {
        return &Visitor{
            URL: url,
            managerPID: mpid,
            visitFn: visitFn,
        }
    }
}

func (v *Visitor) Receive(c *actor.Context){
    switch  c.Message().(type){
    // case VisitReq:
    //     slog.Info("vistor started work", "url", msg.links)
    case actor.Started:
        slog.Info("vister started", "url", v.URL)
        links, err := v.doVisit(v.URL.String(), v.visitFn)
        if err != nil {
            slog.Error("visit error", "err", err)
            return 
        }
        c.Send(v.managerPID, NewVisitRequest(links))
        c.Engine().Poison(c.PID())
    case actor.Stopped:
        slog.Info("vistor stopped", "url", v.URL) 
    }
}

func (v *Visitor) extracLinks(body io.Reader) ([]string, error) {
    links := make([]string, 0)
    tokenizer := html.NewTokenizer(body)

    for {
        tokenType := tokenizer.Next()
        if tokenType == html.ErrorToken {
            return links, nil
        }

        if tokenType == html.StartTagToken {
            token := tokenizer.Token()
            if token.Data == "a" {
                for _, attr := range token.Attr {
                    if attr.Key == "href" {
                        // if attr.Val[0] == '#' {
                        //     continue
                        // }
                        linksUrl, err :=  url.Parse(attr.Val)
                    if err != nil{
                        return links, err
                    }
                    actualLink := v.URL.ResolveReference(linksUrl)
                        links = append(links, actualLink.String())
                    }
                }
            }
        }
    }
}

func (v *Visitor) doVisit(link string, visit VisitFunc) ([]string, error){
    baseUrl, err :=  url.Parse(link)
    if err != nil{
        return []string{}, err
    }
    resp, err := http.Get(baseUrl.String())
    if err != nil{
        return []string{}, err
    }

    w := &bytes.Buffer{}
    r := io.TeeReader(resp.Body, w)

    
    links, err := v.extracLinks(r)
    if err != nil{
        return []string{}, err
    }


    if err := visit(w); err != nil{
        return []string{} , err
    }
    return links, nil
}

type Manager struct{
    visitedHistory map[string]bool
    visitors map [*actor.PID]bool
}

func NewManger() actor.Producer{
    return func () actor.Receiver {
        return &Manager{
            visitors: make(map[*actor.PID]bool),
            visitedHistory: make(map[string]bool),
        }
    }
}

func (m *Manager) Receive(c *actor.Context){
    switch msg := c.Message().(type){
    case VisitReq:
        m.handlevisitReqeust(c, msg)
    case actor.Started:
        slog.Info("manager started")
    case actor.Stopped:
    }
}

func (m *Manager) handlevisitReqeust(c *actor.Context ,msg VisitReq) error {
    for _, link := range msg.links  {
        if _, ok := m.visitedHistory[link]; !ok{
            slog.Info("(Visiting url)", "url", link)
            baseURL, err := url.Parse(link)
            if err != nil{
                return err
            }
        // spawn a child
        c.SpawnChild(NewVisitor(baseURL, c.PID(), msg.visitFunc), "visitor/"+link) 
        m.visitedHistory[link] = true
}
    }
    return nil
}


// for _, link := range links{
//     linksUrl, err :=  url.Parse(link)
//     if err != nil{
//         log.Fatal(err)
//     }
//     actualLink := baseUrl.ResolveReference(linksUrl)
//     fmt.Println(actualLink)

// }

func main() {
    e, err := actor.NewEngine(actor.NewEngineConfig())
    if err != nil {
        log.Fatal(err)
    }
    // spawn it amd return a PID
    pid := e.Spawn(NewManger(), "manager")

    time.Sleep(time.Millisecond * 200)
    
    // send a message to the pid
    e.Send(pid, NewVisitRequest([]string{"https://www.linkedin.com/in/victor-adedeji/"}))
    e.Send(pid, NewVisitRequest([]string{"https://www.jumia.com.ng/"}))
    //e.Send(pid, VisitReq{links: []string{"https://github.com/obigvee"}})
        time.Sleep(time.Second * 1000)
}