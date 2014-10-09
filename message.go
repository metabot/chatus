package chatus

import (
	"bytes"
	"encoding/xml"
	"errors"
	"io"
	"text/template"
	"regexp"
	"strconv"
)

type Pullable interface {
	PullTemplate() *template.Template
}

type Pushable interface {
	PushTemplate() *template.Template
}

func ToPullResp(msg Pullable) (string, error) {
	var buf bytes.Buffer
	if err := msg.PullTemplate().Execute(&buf, msg); err != nil {
		return "", err
	}

	return buf.String(), nil
}

func ToPushResp(msg Pushable) (string, error) {
	var buf bytes.Buffer
	if err := msg.PushTemplate().Execute(&buf, msg); err != nil {
		return "", err
	}

	return buf.String(), nil
}

//Struct to parse incoming messages/events
type InMessage struct {
	XMLName      xml.Name `xml:"xml" json:"-"`
	Id           uint64   `xml:"MsgId" json:"omitempty"`
	To           string   `xml:"ToUserName" json:"touser"`
	From         string   `xml:"FromUserName" json:"from,omitempty"`
	Time         int64    `xml:"CreateTime" json:"omitempty"`
	Type         string   `xml:"MsgType" json:"msgtype"`
	Content      string   `xml:"Content,omitempty" json:"text>content,omitempty"`
	MediaId      string   `xml:"MediaId,omitempty" json:"omitempty"`      //optional msgType: image|voice|video
	PUrl         string   `xml:"PicUrl,omitempty" json:"omitempty"`       //optional msgType: image
	Format       string   `xml:"Format,omitempty" json:"omitempty"`       //optional msgType: voice
	ThumbMediaId string   `xml:"ThumbMediaId,omitempty" json:"omitempty"` //optional msgType: video
	LocX         float64  `xml:"Location_X,omitempty" json:"omitempty"`   //optional msgType: location
	LocY         float64  `xml:"Location_Y,omitempty" json:"omitempty"`   //optional msgType: location
	Scale        uint     `xml:"Scale,omitempty" json:"omitempty"`        //optional msgType: location
	Label        string   `xml:"Label,omitempty" json:"omitempty"`        //optional msgType: location
	Title        string   `xml:"Title,omitempty" json:"omitempty"`        //optional msgType: link
	Desc         string   `xml:"Description,omitempty" json:"omitempty"`  //optional msgType: link
	Url          string   `xml:"url,omitempty" json:"omitempty"`          //optional msgType: link
	Event        string   `xml:"Event,omitempty"`                         //optional msgType: event
	EventKey     string   `xml:"EventKey,omitempty"`                      //optional msgType: event
	Lat          float64  `xml:"Latitude,omitempty"`                      //optional msgType: event
	Lng          float64  `xml:"Longitude,omitempty"`                     //optional msgType: event
	Precision    float64  `xml:"Precision,omitempty"`
	Ticket       string   `xml:"Ticket,omitempty"`
}

func ParseInMessage(r io.Reader) (*InMessage, error) {
	msg := &InMessage{}
	decoder := xml.NewDecoder(r)
	if err := decoder.Decode(msg); err != nil {
		return nil, err
	}

	return msg, nil
}


const (
	basePllStr = `<ToUserName>{{.To}}</ToUserName><FromUserName>{{.From}}</FromUserName>` +
		`<CreateTime>{{.Time}}</CreateTime><MsgType>{{.Type}}</MsgType>`

	basePshStr = `"touser":"{{.To}}", "msgtype":"{{.Type}}",`

	txtPllStr = `<xml>` + basePllStr +
		`<Content>{{.Content}}</Content></xml>`

	txtPshStr = `{` + basePshStr +
		`"text":{"content":"{{.Content}}"}}`

	imgPllStr = `<xml>` + basePllStr +
		`<Image><MediaId>{{.MediaId}}</MediaId></Image></xml>`

	imgPshStr = `{` + basePshStr +
		`"image":{"media_id":"{{.MediaId}}"}}`

	newsPllStr = `<xml>` + basePllStr +
		`<ArticleCount>{{len .Articles}}</ArticleCount><Articles>` +
		`{{range $_, $a := .Articles}}<item><Title>{{.Title}}</Title>` +
		`<Description>{{.Desc}}</Description><PicUrl><{{.PUrl}}></PicUrl><Url>{{.Url}}</Url></item>{{end}}</Articles></xml>`

	newsPshStr = `{` + basePshStr +
		`"news":{"articles" : [ {{range $index, $a := .Articles}}` +
		`{{ if $index}},{{end}}` +
		`{"title" : "{{.Title}}", "description" : "{{.Desc}}", "url" : "{{.Url}}", "picurl" : "{{.PUrl}}"}` +
		`{{end}}]}}`
)

var txtPllTmpl = template.Must(template.New("txtPll").Parse(txtPllStr))
var txtPshTmpl = template.Must(template.New("txtPsh").Parse(txtPshStr))
var imgPllTmpl = template.Must(template.New("imgPll").Parse(imgPllStr))
var imgPshTmpl = template.Must(template.New("imgPsh").Parse(imgPshStr))
var newsPllTmpl = template.Must(template.New("newsPll").Parse(newsPllStr))
var newsPshTmpl = template.Must(template.New("newsPsh").Parse(newsPshStr))

type header struct {
	Id   uint64
	From string
	To   string
	Time int64
	Type string
}

type TextMessage struct {
	header
	Content string
}

func (in InMessage) ToTxtMsg() (*TextMessage, error) {
	if in.Type != "text" {
		return nil, errors.New("cannot convert " + in.Type + " into text message")
	}

	txtMsg := &TextMessage{}
	txtMsg.Id, txtMsg.From, txtMsg.To, txtMsg.Type, txtMsg.Content = in.Id, in.From, in.To, in.Type, in.Content

	return txtMsg, nil
}

func (_ TextMessage) PullTemplate() *template.Template {
	return txtPllTmpl
}

func (_ TextMessage) PushTemplate() *template.Template {
	return txtPshTmpl
}

type ImageMessage struct {
	header
	MediaId string
}

func (_ ImageMessage) PullTemplate() *template.Template {
	return imgPllTmpl
}

func (_ ImageMessage) PushTemplate() *template.Template {
	return imgPshTmpl
}

type VoiceMessage struct {
	header
	MediaId string
}

type Article struct {
	Title string `json:"title,omitempty"`
	Desc  string `json:"description,omitempty"`
	Url   string `json:"url,omitempty"`
	PUrl  string `json:"picurl,omitempty"`
}

type NewsMessage struct {
	header
	Articles []Article
}

func (_ NewsMessage) PullTemplate() *template.Template {
	return newsPllTmpl
}

func (_ NewsMessage) PushTemplate() *template.Template {
	return newsPshTmpl
}

type UserInfo struct {
	Subscribe uint   `json:"subscribe"`
	Id        string `json:"openid,omitempty"`
	NickName  string `json:"nickname,omitempty"`
	Gender    uint   `json:"sex,omitempty"`
	Language  string `json:"language"`
	City      string `json:"city"`
	Province  string `json:"province"`
	Country   string `json:"country"`
	Avatar    string `json:"headimgurl,omitempty"`
	SubTime   int64  `json:"subscribe_time,omitempty`
}

type GeoEvent struct {
	UId string
	Lat float64
	Lng float64
}

func (msg *InMessage) ToGeoEvent() (*GeoEvent, error) {
	if msg.Event != "LOCATION" {
		return nil, errors.New("not a geo event ")
	}

	ge := &GeoEvent{}

	ge.UId, ge.Lat, ge.Lng = msg.From, msg.Lat, msg.Lng

	return ge, nil
}


type SubscribeEvent struct {
	UId string
	Key int  //channel keys
}

var reSubKey = regexp.MustCompile("qrscene_([0-9]+)$")

func (msg *InMessage) ToSubscribeEvent() (*SubscribeEvent, error) {
	if msg.Event != "subscribe" {
		return nil, errors.New("not a subscribe event ")
	}

	se := &SubscribeEvent{}

	se.UId, se.Key = msg.From, 0

	if msg.EventKey != "" {
		//parse the event key
		matches := reSubKey.FindStringSubmatch(msg.EventKey)
		if len(matches) == 2 {
			se.Key,_ = strconv.Atoi(matches[1])
		}
	}

	return se, nil
}


type ScanEvent struct {
	UId string
	Key int
}


func (msg *InMessage) ToScanEvent() (*ScanEvent, error) {
	if msg.Event != "SCAN" {
		return nil, errors.New("not a SCAN event ")
	}

	se := &ScanEvent{}

	se.UId, se.Key = msg.From, 0

	//parse the event key
	se.Key,_ = strconv.Atoi(msg.EventKey)

	return se, nil
}


type ViewEvent struct {
	UId string
	Url string
}
