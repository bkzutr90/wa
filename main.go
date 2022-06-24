package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strings"
	"syscall"
    "net/url"
    "net/http"
    "io/ioutil"
    "github.com/tidwall/gjson"

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
        tod := cok.Info.Sender
        if strings.HasPrefix(txt, "/myid") {
            var tods = cok.Info.ID
            vh.SendTextMessage(to, "Cek personal message untuk melihat ID")
            vh.SendTextMessage(tod, "Your ID\n> " + tods)
		if strings.HasPrefix(txt, "/about") {
            ig := "This bot created by *Lyrics.AnimeMusic*\nBuild with Golang"
            ig += "\n\nFollow IG"
            ig+= "\ninstagram.com/lyrics.animemusic"
			vh.SendTextMessage(to, ig)
        }
        if strings.HasPrefix(txt, "/menu") {
            menus := "╭──────────"                                                 
            menus += "\n│                Menu"
            menus += "\n│──────────"
            menus += "\n│/menu
			menus += "\n│/apikey"
            menus += "\n│/myid
            menus += "\n│/about"
            menus += "\n│/quotes
            menus += "\n│/artinama (nama mu)"                                      
            menus += "\n│/jadwalsholat (kota mu)"                                  
            menus += "\n│──────────"
            menus += "\n│ ©      Just Ferianss"                                  
            menus += "\n╰──────────"                                               
            vh.SendTextMessage(to, menus)
        }
        if strings.HasPrefix(txt, "/jadwalsholat") {
            str := strings.Replace(txt, "/jadwalsholat ", "", 1)
            url := "https://api.xteam.xyz/jadwalsholat?kota=" + url.QueryEscape(str) + "&APIKEY=e3cf51d0639bfd38"
            res, err := http.Get(url)
            if err != nil {
                fmt.Println(err)
            } else {
                defer res.Body.Close()
                body, _ := ioutil.ReadAll(res.Body)
                dear := "╭─Jadwal Sholat" + fmt.Sprintf("%s", gjson.Get(string(body), "Kota").String()) + "\n\n"
                dear += "\n│*Shubuh* : " + fmt.Sprintf("%s", gjson.Get(string(body), "Subuh").String()) + " WIB"
                dear += "\n│*Zduhur* : " + fmt.Sprintf("%s", gjson.Get(string(body), "Dzuhur").String()) + " WIB"
                dear += "\n│*Ashr* : " + fmt.Sprintf("%s", gjson.Get(string(body), "Ashar").String()) + " WIB"
                dear += "\n│*Magrib* : " + fmt.Sprintf("%s", gjson.Get(string(body), "Magrib").String()) + " WIB"
                dear += "\n│*Isya* : " + fmt.Sprintf("%s", gjson.Get(string(body), "Isha").String()) + " WIB"
                dear += "\n\n*╰─Tanggal* : " + fmt.Sprintf("%s", gjson.Get(string(body), "Tanggal").String())
                vh.SendTextMessage(to, dear)
            }
        }
        if strings.HasPrefix(txt, "/quotes") {
            url := "https://st4rz.herokuapp.com/api/randomquotes"
            res, err := http.Get(url)
            if err != nil {
                fmt.Println(err)
            } else {
                defer res.Body.Close()
                body, _ := ioutil.ReadAll(res.Body)
                dear := "Random Quotes"
                dear += "\n\nAuthor\n> " + "*" + fmt.Sprintf("%s", gjson.Get(string(body), "author").String()) + "*"
                dear += "\n>Quotes\n> " + fmt.Sprintf("%s", gjson.Get(string(body), "quotes").String())
                vh.SendTextMessage(to, dear)
            }
        }
        if strings.HasPrefix(txt, "/apikey") {
            url := "https://api.xteam.xyz/cekey?&APIKEY=e3cf51d0639bfd38"
            res, err := http.Get(url)
            if err != nil {
                fmt.Println(err)
            } else {
                defer res.Body.Close()
                body, _ := ioutil.ReadAll(res.Body)
                dear := "Information APIKEY"
                dear += "\n\n>Username : " + fmt.Sprintf("%s", gjson.Get(string(body), "response.name").String())
                dear += "\n>Response : " + fmt.Sprintf("%s", gjson.Get(string(body), "response.totalhit").String()) + "/100"
                dear += "\n>Your APIKEY : " +  fmt.Sprintf("%s", gjson.Get(string(body), "response.apikey").String())
                dear += "*API Rest by XTEAM, VHTear*"
                vh.SendTextMessage(to, dear)
            }
        }
        if strings.HasPrefix(txt, "/artinama") {
            str := strings.Replace(txt, "/artinama ", "", 1)
            url := "https://api.xteam.xyz/primbon/artinama?q=" + url.QueryEscape(str) + "&apikey=NOT-PREMIUM"
            res, err := http.Get(url)
            if err != nil {
                fmt.Println(err)
            } else {
                defer res.Body.Close()
                body, _ := ioutil.ReadAll(res.Body)
                dear := "Berikut arti nama " + fmt.Sprintf("%s", gjson.Get(string(body), "result.nama").String())
                dear += "\n\n> " + fmt.Sprintf("%s", gjson.Get(string(body), "result.arti").String())
                dear += "\n> " + fmt.Sprintf("%s", gjson.Get(string(body), "result.maksud").String())
                vh.SendTextMessage(to, dear)
        }
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
