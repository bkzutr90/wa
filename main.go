package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"

	_ "github.com/mattn/go-sqlite3"
	"github.com/mdp/qrterminal"
	"github.com/sirupsen/logrus"
	"google.golang.org/protobuf/proto"

	"go.mau.fi/whatsmeow"
	"go.mau.fi/whatsmeow/store"
	"go.mau.fi/whatsmeow/store/sqlstore"
	"go.mau.fi/whatsmeow/types"
	"go.mau.fi/whatsmeow/types/events"

	waProto "go.mau.fi/whatsmeow/binary/proto"
	waLog "go.mau.fi/whatsmeow/util/log"
)

type VEZZA struct {
	VClient *whatsmeow.Client
}

var Client VEZZA
var Log *logrus.Logger

func (vh *VEZZA) register() {
	vh.VClient.AddEventHandler(vh.MessageHandler)
}

func (vh *VEZZA) newClient(d *store.Device, l waLog.Logger) {
	vh.VClient = whatsmeow.NewClient(d, l)
}

func (vh *VEZZA) SendMessageV2(evt interface{}, msg *string) {
	v := evt.(*events.Message)
	resp := &waProto.Message{
		ExtendedTextMessage: &waProto.ExtendedTextMessage{
			Text: msg,
			ContextInfo: &waProto.ContextInfo{
				StanzaId:    &v.Info.ID,
				Participant: proto.String(v.Info.MessageSource.Sender.String()),
			},
		},
	}
	vh.VClient.SendMessage(v.Info.Sender, "", resp)
}

func (vh *VEZZA) SendTextMessage(jid types.JID, text string) {
	vh.VClient.SendMessage(jid, "", &waProto.Message{Conversation: proto.String(text)})
}

func (vh *VEZZA) MessageHandler(evt interface{}) {
	switch v := evt.(type) {
	case *events.Message:
		cok := evt.(*events.Message)
		fmt.Println(cok.Info.Chat)

		txt := strings.ToLower(v.Message.GetConversation())
		to := cok.Info.Chat
		if strings.HasPrefix(txt, "about") {
			vh.SendTextMessage(to, "This bot created by JustFerianss")
		}
        if strings.HasPrefix(txt, "imsakiyah") {
            str := strings.Replace(txt, "imsakiyah ", "", 1)
            url := "https://api.vhtear.com/jadwalsholat?query=" + url.QueryEscape(str) + "&apikey=NOT-PREMIUM"
            res, err := http.Get(url)
            if err != nil {
                fmt.Println(err)
            } else {
                defer res.Body.Close()
                body, _ := ioutil.ReadAll(res.Body)
                dear := "Imsakiyah-2022\n"
                dear += "\n*Tanggal* : " + fmt.Sprintf("%s", gjson.Get(string(body), "result.tanggal").String()) + " WIB"
                dear += "\n*Shubuh* : " + fmt.Sprintf("%s", gjson.Get(string(body), "result.Shubuh").String()) + " WIB"
                dear += "\n*Zduhur* : " + fmt.Sprintf("%s", gjson.Get(string(body), "result.Zduhur").String()) + " WIB"
                dear += "\n*Ashr* : " + fmt.Sprintf("%s", gjson.Get(string(body), "result.Ashr").String()) + " WIB"
                dear += "\n*Magrib* : " + fmt.Sprintf("%s", gjson.Get(string(body), "result.Magrib").String()) + " WIB"
                dear += "\n*Isya* : " + fmt.Sprintf("%s", gjson.Get(string(body), "result.Isya").String()) + " WIB"
                dear += "\n\n*Kota* : " + fmt.Sprintf("%s", gjson.Get(string(body), "result.kota").String())
                dear += "\n*Tanggal* : " + fmt.Sprintf("%s", gjson.Get(string(body), "result.tanggal").String()) + " WIB"
                dear += "\n" + fmt.Sprintf("%s", gjson.Get(string(body), "result.ramadhan").String())
                vh.SendTextMessage(to, dear)
            }
        }
        if strings.HasPrefix(txt, "/artinama") {
            str := strings.Replace(txt, "/artinama ", "", 1)
            url := "https://api.vhtear.com/ramalan_nama?nama=" + url.QueryEscape(str) + "&apikey=NOT-PREMIUM"
            res, err := http.Get(url)
            if err != nil {
                fmt.Println(err)
            } else {
                defer res.Body.Close()
                body, _ := ioutil.ReadAll(res.Body)
                dear := "Berikut arti nama And\n\n" + fmt.Sprintf("%s", gjson.Get(string(body), "result.hasil").String())
                dear += "\n\nSource : " + fmt.Sprintf("%s", gjson.Get(string(body), "result.source").String())
                vh.SendTextMessage(to, dear)
		return
	}
}

func main() {
	dbLog := waLog.Stdout("Database", "DEBUG", true)
	container, err := sqlstore.New("sqlite3", "file:commander.db?_foreign_keys=on", dbLog)
	if err != nil {
		panic(err)
	}

	deviceStore, err := container.GetFirstDevice()
	if err != nil {
		panic(err)
	}

	clientLog := waLog.Stdout("Client", "DEBUG", true)
	Client.newClient(deviceStore, clientLog)
	Client.register()

	if Client.VClient.Store.ID == nil {
		qrChan, _ := Client.VClient.GetQRChannel(context.Background())
		err = Client.VClient.Connect()
		if err != nil {
			panic(err)
		}
		for evt := range qrChan {
			if evt.Event == "code" {
				qrterminal.GenerateHalfBlock(evt.Code, qrterminal.L, os.Stdout)
			} else {
				fmt.Println("Login event:", evt.Event)
			}
		}

	} else {
		err = Client.VClient.Connect()
		fmt.Println("Login Success")
		if err != nil {
			panic(err)
		}
	}
	c := make(chan os.Signal)
	signal.Notify(c, os.Interrupt, syscall.SIGTERM)
	<-c

	Client.VClient.Disconnect()
}
