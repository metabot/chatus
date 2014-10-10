package main

import (
	"github.com/zenazn/goji"
	"github.com/zenazn/goji/web"
	"net/http"
	"fmt"
	"log"
	. "github.com/metabot/chatus"
	"time"
)


//load Configuration for WeChat
var ws = &WechatStation {
		Id: "example",
		Token: "foobar",
		ApiAccess: ApiAccess{
			Appid: "YOUR_APPID",
			Secret: "YOUR_SeCRet",
			ApiURL: "YOUR_APIURL",
		},
}

func main() {
	ws.AddProcessor(EchoTxtMsgProcessor)
	ws.AddProcessor(ClickEventProcessor)

	feeder := web.New()
	feeder.Use(validate)
	goji.Handle("/v0/:station", feeder)

	feeder.Get("/v0/:station", echo)
	feeder.Post("/v0/:station", receive)

	goji.Serve()
}

func validate(c *web.C, h http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		st := c.URLParams["station"]

		if st != ws.Id {
			http.Error(w, "no such station:"+st, http.StatusNotFound)
			return
		}

		timestamp := r.FormValue("timestamp")
		nonce := r.FormValue("nonce")
		signature := r.FormValue("signature")

		if err := ws.IsValid(timestamp, nonce, signature); err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		c.Env["wechatStation"] = ws
		h.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}

func echo(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, r.FormValue("echostr"))
}


func receive(c web.C, w http.ResponseWriter, r *http.Request) {
	s := c.Env["wechatStation"].(*WechatStation)

	resp, err := s.Process(r.Body)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	log.Println("Response: ", resp)

	fmt.Fprintln(w, resp)
}


var (
	EchoTxtMsgProcessor = &Processor{
	Type: "text.",
	Handle: func (i *InMessage) (string, error) {
		//just echo
		o, err := i.ToTxtMsg()
		if err != nil {
			return "", err
		}

		o.To, o.From, o.Time = i.From, i.To, time.Now().Unix()
		return ToPullResp(o)
	},
}


	ClickEventProcessor = &Processor{
	Type: "event.CLICK",
	Handle: func (_ *InMessage) (string, error) {
		return "", nil
	},
}



	DefaultProcessor = &Processor {
	Type: "*",
	Handle: func (i *InMessage) (string, error) {
		tm := TextMessage{
			Header {
				From: i.To,
				To:   i.From,
				Time: time.Now().Unix(),
				Type: "text"},
			"暂不支持该功能",
		}
		return ToPullResp(tm)
	},
}
)

