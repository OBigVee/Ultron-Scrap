package util

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/anthdm/hollywood/actor"
	"golang.org/x/net/html"
)

type VisitReq struct {
	Links []string
}

type Manager struct{}

func NewManager() actor.Producer {
	return func() actor.Receiver {
		return &Manager{}
	}
}

func (m *Manager) Receive(c *actor.Context) {
	switch msg := c.Message().(type) {
	case VisitReq:
		m.handlevisitRequest(msg)
	case actor.Started:
		log.Println("manager started")
	case actor.Stopped:
	}
}

func (m *Manager) handlevisitRequest(msg VisitReq) error {
	for _, link := range msg.Links {
		log.Println("(Visiting url)", "url", link)
	}
	return nil
}

func ExtractLinks(body io.Reader) []string {
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
}
