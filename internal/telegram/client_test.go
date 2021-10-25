package telegram_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/jarcoal/httpmock"
	"github.com/javiyt/tweetgram/internal/bot"
	"github.com/javiyt/tweetgram/internal/telegram"
	"github.com/stretchr/testify/require"
	tb "gopkg.in/tucnak/telebot.v2"
	"io/ioutil"
	"math/rand"
	"net/http"
	"os"
	"sync/atomic"
	"testing"
	"time"
)

func TestMain(m *testing.M) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	meJson, _ := ioutil.ReadFile("testdata/me.json")
	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.mock/botasdfg:12345/getMe",
		httpmock.NewStringResponder(
			200,
			string(meJson),
		),
	)
	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.mock/botasdfg:12345/getUpdates",
		httpmock.NewStringResponder(
			200,
			" {\"ok\":true,\"result\":[]}",
		),
	)
	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.mock/botasdfg:12345/setMyCommands",
		httpmock.NewStringResponder(
			200,
			" {\"ok\":true,\"result\":true}",
		),
	)

	os.Exit(m.Run())
}

func TestBot_Start(t *testing.T) {
	tlgmbot, err := tb.NewBot(tb.Settings{
		URL:   "https://api.telegram.mock",
		Token: "asdfg:12345",
		Poller: &tb.LongPoller{
			Timeout: 10 * time.Second,
		},
	})
	require.NoError(t, err)

	bt := telegram.NewBot(tlgmbot)
	defer bt.Stop()

	go bt.Start()

	require.Eventually(t, func() bool {
		countInfo := httpmock.GetCallCountInfo()
		_, ok := countInfo["POST https://api.telegram.mock/botasdfg:12345/getMe"]
		return ok
	}, time.Second, time.Millisecond)
}

func TestBot_SetCommands(t *testing.T) {
	tlgmbot, err := tb.NewBot(tb.Settings{
		URL:   "https://api.telegram.mock",
		Token: "asdfg:12345",
		Poller: &tb.LongPoller{
			Timeout: 10 * time.Second,
		},
	})
	require.NoError(t, err)

	bt := telegram.NewBot(tlgmbot)

	err = bt.SetCommands([]bot.TelegramBotCommand{
		{
			Text:        "a",
			Description: "desc",
		},
	})

	require.Eventually(t, func() bool {
		countInfo := httpmock.GetCallCountInfo()
		_, ok := countInfo["POST https://api.telegram.mock/botasdfg:12345/setMyCommands"]
		return ok
	}, time.Second, time.Millisecond)
	require.NoError(t, err)
}

func TestBot_Handle(t *testing.T) {
	updateJson, _ := ioutil.ReadFile("testdata/image.json")
	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.mock/botasdfg:12345/getUpdates",
		httpmock.NewStringResponder(
			200,
			string(updateJson),
		),
	)

	tlgmbot, err := tb.NewBot(tb.Settings{
		URL:   "https://api.telegram.mock",
		Token: "asdfg:12345",
		Poller: &tb.LongPoller{
			Timeout: 10 * time.Second,
		},
	})
	require.NoError(t, err)

	bt := telegram.NewBot(tlgmbot)

	var handled atomic.Value
	handled.Store(false)

	bt.Handle(tb.OnPhoto, func(m *bot.TelegramMessage) {
		handled.Store(m.Photo.Caption == "image")
		bt.Stop()
	})

	go bt.Start()

	require.Eventually(t, func() bool {
		return handled.Load().(bool)
	}, time.Second, time.Millisecond)
}

func TestBot_Send(t *testing.T) {
	tlgmbot, err := tb.NewBot(tb.Settings{
		URL:   "https://api.telegram.mock",
		Token: "asdfg:12345",
		Poller: &tb.LongPoller{
			Timeout: 10 * time.Second,
		},
	})
	require.NoError(t, err)

	bt := telegram.NewBot(tlgmbot)

	var testMessageSent atomic.Value
	testMessageSent.Store(false)
	var testLongMessageSent atomic.Value
	testLongMessageSent.Store(false)
	var photoSent atomic.Value
	photoSent.Store(false)

	firstLongMessageSent := false
	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.mock/botasdfg:12345/sendMessage",
		func(req *http.Request) (*http.Response, error) {
			buf := new(bytes.Buffer)
			_, _ = buf.ReadFrom(req.Body)
			var requestBody struct {
				ChatID  string `json:"chat_id"`
				Text    string `json:"text"`
				ReplyTo string `json:"reply_to_message_id"`
			}
			_ = json.Unmarshal(buf.Bytes(), &requestBody)
			if requestBody.ChatID == "1234567890" && requestBody.Text == "test message" {
				testMessageSent.Store(true)
				messageSent, _ := ioutil.ReadFile("testdata/sendmessage.json")
				return httpmock.NewStringResponse(
					200,
					string(messageSent),
				), nil
			} else if len(requestBody.Text) == 4096 {
				firstLongMessageSent = true
				messageSent, _ := ioutil.ReadFile("testdata/sendmessage.json")
				return httpmock.NewStringResponse(
					200,
					string(messageSent),
				), nil
			} else if firstLongMessageSent && requestBody.ReplyTo == "59" {
				testLongMessageSent.Store(true)
				messageSent, _ := ioutil.ReadFile("testdata/sendmessage.json")
				return httpmock.NewStringResponse(
					200,
					string(messageSent),
				), nil
			}

			return httpmock.NewStringResponse(500, "response not found"),
				errors.New("response not found")
		},
	)

	photoUpdate, _ := ioutil.ReadFile("testdata/sendphoto.json")
	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.mock/botasdfg:12345/sendPhoto",
		func(req *http.Request) (*http.Response, error) {
			photoSent.Store(true)
			return httpmock.NewStringResponse(
				200,
				string(photoUpdate),
			), nil
		},
	)

	t.Run("it sends a text message", func(t *testing.T) {
		err = bt.Send("1234567890", "test message")

		require.NoError(t, err)
		require.Eventually(t, func() bool {
			return testMessageSent.Load().(bool)
		}, time.Second, time.Millisecond)
	})

	t.Run("it send a text message longer than expected", func(t *testing.T) {
		const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
		b := make([]byte, 5000)
		for i := range b {
			b[i] = letters[rand.Intn(len(letters))]
		}

		err = bt.Send("1234567890", string(b))

		require.NoError(t, err)
		require.Eventually(t, func() bool {
			return testLongMessageSent.Load().(bool)
		}, time.Second, time.Millisecond)
	})

	t.Run("it should send a picture", func(t *testing.T) {
		err = bt.Send("1234567890", bot.TelegramPhoto{
			Caption:  "test",
			FileID:   "123456",
			FileURL:  "http://image.url",
			FileSize: 1234,
		})

		require.NoError(t, err)
		require.Eventually(t, func() bool {
			return photoSent.Load().(bool)
		}, time.Second, time.Millisecond)
	})
}

func TestBot_GetFile(t *testing.T) {
	fileJson, _ := ioutil.ReadFile("testdata/getfile.json")
	httpmock.RegisterResponder(
		"POST",
		"https://api.telegram.mock/botasdfg:12345/getFile",
		httpmock.NewStringResponder(
			200,
			string(fileJson),
		),
	)

	icon, _ := ioutil.ReadFile("testdata/td_icon.png")
	httpmock.RegisterResponder(
		"GET",
		"https://api.telegram.mock/file/botasdfg:12345/photos/file_4.jpg",
		httpmock.NewBytesResponder(
			200,
			icon,
		),
	)

	tlgmbot, err := tb.NewBot(tb.Settings{
		URL:   "https://api.telegram.mock",
		Token: "asdfg:12345",
		Poller: &tb.LongPoller{
			Timeout: 10 * time.Second,
		},
	})
	require.NoError(t, err)

	bt := telegram.NewBot(tlgmbot)

	_, err = bt.GetFile("AZCDxruqG7J3iTM9")

	require.NoError(t, err)
}
