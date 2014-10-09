package chatus

import (
	"time"
	"github.com/metabot/utils"
	"io"
	"log"
	"encoding/json"
	"net/url"
	"strings"
	"net/http"
	"io/ioutil"
	"errors"
	"fmt"
)


type WechatStation struct {
	processors map[string]*Processor
	defaultProcessor *Processor
	id string
	token string
	ApiAccess
}


func (s *WechatStation) AddProcessor(processor *Processor) {
	if (s.processors == nil) {
		s.processors = make (map[string]*Processor)
	}

	s.processors[processor.Type] = processor
}


func (s *WechatStation) IsValid(timestamp string, nonce string, signature string) error {
	if signature == utils.GenerateSignature(timestamp, nonce, s.token) {
		return nil
	}
	return utils.ErrInvalidRequest
}


func (s *WechatStation) Process(in io.Reader) (string ,error) {
	//parse to InMessage
	msg, err := ParseInMessage(in)
	if err != nil {
		return "", err
	}

	log.Println("Parsed Message: ", msg)

	if p, present := s.processors[processorKey(msg)]; present {
		return p.handle(msg)
	} else {
		return s.defaultProcessor.handle(msg)
	}
}

func processorKey(m *InMessage) string {
	ks := []string{m.Type, m.Event}
	return strings.Join(ks, ".")
}


//Processor to process one type of message/event
// Type:  {{InMessage.Type}}.{{InMessage.Event}}
type Processor struct {
	Type string
	handle  func(*InMessage) (string, error)
}

var (
	EchoTxtMsgProcessor = &Processor{
	Type: "text.",
	handle: func (i *InMessage) (string, error) {
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
	handle: func (_ *InMessage) (string, error) {
		return "", nil
	},
}



	DefaultProcessor = &Processor {
	Type: "*",
	handle: func (i *InMessage) (string, error) {
		tm := TextMessage{
			header {
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


////////////////
// distributor
////////////////

type pushResp struct {
	Errcode int    `json:"errcode"`
	Errmsg  string  `json:"errmsg"`
}

type ApiAccess struct {
	Appid   string
	Secret  string
	ApiURL  string
	AccessToken *accessToken
}

type accessToken struct {
	Token     string `json:"access_token"`
	ExpiresIn int  `json:"expires_in"`
	ExpiresAt time.Time
}


func (s *WechatStation) accessToken() (*accessToken, error) {
	if s.AccessToken!= nil && s.AccessToken.ExpiresAt.After(time.Now()) {
		return s.AccessToken, nil
	}

	//build URL
	u, err := url.Parse(s.ApiURL)
	if err != nil {
		return "", err
	}

	u.Path = "cgi-bin/token"
	q := u.Query()
	q.Set("appid", s.Appid)
	q.Set("secret", s.Secret)
	q.Set("grant_type", "client_credential")
	u.RawQuery = q.Encode()

	//get access_token and expiration time
	resp, err := http.Get(u.String())
	if err != nil {
		return "", err
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}
	if resp.StatusCode > 400 {
		return "", errors.New(string(body))
	}

	atk, err := parseAccessToken(string(body))

	s.AccessToken = atk

	return atk, nil
}

func parseAccessToken(in string) (*accessToken, error) {
	atk := &accessToken{}
	err := json.Unmarshal([]byte(in), atk)
	if err != nil {
		return nil, err
	}

	atk.ExpiresAt = time.Now().Add(time.Duration(atk.ExpiresIn) * time.Second)

	return atk, nil
}



func (s *WechatStation) Send(data string) error {
	at, err := s.accessToken()
	if err != nil {
		return errors.New("failed to get access token: " + err.Error())
	}

	u := fmt.Sprintf("%s/cgi-bin/message/custom/send?access_token=%s", s.ApiURL, at.Token)
	resp, err := http.Post(u, "text/json", strings.NewReader(data))
	if err != nil {
		return errors.New("failed to send message: " + err.Error())
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.New("failed to read message: " + err.Error())
	}

	if resp.StatusCode > 400 {
		return errors.New(fmt.Sprintf("%d:%s", resp.StatusCode, string(body)))
	}
	pr := &pushResp{}
	err = json.Unmarshal(body, pr)
	if err != nil {
		return err
	}

	if pr.Errcode != 0 {
		return errors.New(pr.Errmsg)
	}

	return nil
}


