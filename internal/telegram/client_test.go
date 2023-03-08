package telegram_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"os"
	"sync/atomic"
	"testing"
	"time"

	tbBotMock "github.com/javiyt/tweetgram/mocks/telegram"

	"github.com/javiyt/tweetgram/mocks/telebot"

	"github.com/jarcoal/httpmock"
	"github.com/javiyt/tweetgram/internal/bot"
	"github.com/javiyt/tweetgram/internal/telegram"
	"github.com/stretchr/testify/require"
	tb "gopkg.in/telebot.v3"
)

const (
	botToken            = "asdfg:12345"
	botImageHandleToken = "qwert:98765"
	botSendToken        = "zxcvb:54321"
)

var (
	testMessageSent     atomic.Value
	testLongMessageSent atomic.Value
	photoSent           atomic.Value
	firstLongMessage    atomic.Value
)

func TestMain(m *testing.M) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	meJson, _ := os.ReadFile("testdata/me.json")
	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://api.telegram.mock/bot%s/getMe", botToken),
		httpmock.NewStringResponder(
			200,
			string(meJson),
		),
	)
	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://api.telegram.mock/bot%s/getUpdates", botToken),
		httpmock.NewStringResponder(
			200,
			" {\"ok\":true,\"result\":[]}",
		),
	)
	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://api.telegram.mock/bot%s/setMyCommands", botToken),
		httpmock.NewStringResponder(
			200,
			" {\"ok\":true,\"result\":true}",
		),
	)

	updateJson, _ := os.ReadFile("testdata/image.json")
	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://api.telegram.mock/bot%s/getUpdates", botImageHandleToken),
		httpmock.NewStringResponder(
			200,
			string(updateJson),
		),
	)

	registerResponders(botSendToken, &testMessageSent, &testLongMessageSent, &photoSent, &firstLongMessage)

	os.Exit(m.Run())
}

func TestBot_Start(t *testing.T) {
	tlgmbot, err := tb.NewBot(tb.Settings{
		URL:   "https://api.telegram.mock",
		Token: botToken,
		Poller: &tb.LongPoller{
			Timeout: 10 * time.Second,
		},
		Offline: true,
	})
	require.NoError(t, err)

	go telegram.NewBot(tlgmbot).Start()

	require.Eventually(t, func() bool {
		countInfo := httpmock.GetCallCountInfo()
		_, ok := countInfo["POST https://api.telegram.mock/botasdfg:12345/getMe"]

		return ok
	}, time.Second, time.Millisecond)
}

func TestBot_Stop(t *testing.T) {
	tbBot := tbBotMock.NewTbBot(t)
	tbBot.On("Stop")
	telegram.NewBot(tbBot).Stop()
	tbBot.AssertCalled(t, "Stop")
}

func TestBot_SetCommands(t *testing.T) {
	tlgmbot, err := tb.NewBot(tb.Settings{
		URL:   "https://api.telegram.mock",
		Token: botToken,
		Poller: &tb.LongPoller{
			Timeout: 10 * time.Second,
		},
		Offline: true,
	})
	require.NoError(t, err)

	err = telegram.NewBot(tlgmbot).SetCommands([]bot.TelegramBotCommand{
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
	tlgmbot, err := tb.NewBot(tb.Settings{
		URL:   "https://api.telegram.mock",
		Token: botImageHandleToken,
		Poller: &tb.LongPoller{
			Timeout: 10 * time.Second,
		},
		Offline: true,
	})
	require.NoError(t, err)

	bt := telegram.NewBot(tlgmbot)

	var handled atomic.Value

	handled.Store(false)

	bt.Handle(tb.OnPhoto, func(m bot.TelegramMessage) error {
		handled.Store(m.Photo.Caption == "image")

		return nil
	})

	go bt.Start()

	require.Eventually(t, func() bool {
		b, ok := handled.Load().(bool)

		return ok && b
	}, time.Second, time.Millisecond)
}

func TestBot_Send(t *testing.T) {
	tlgmbot, _ := tb.NewBot(tb.Settings{URL: "https://api.telegram.mock", Token: botSendToken, Poller: &tb.LongPoller{
		Timeout: 10 * time.Second,
	}, Offline: true})

	bt := telegram.NewBot(tlgmbot)

	testMessageSent.Store(false)
	testLongMessageSent.Store(false)
	photoSent.Store(false)
	firstLongMessage.Store(false)

	t.Run("it should fail when unsupported message sent", func(t *testing.T) {
		require.EqualError(t, bt.Send("1234567890", tb.File{}), "unsupported type")
	})

	t.Run("it should fail when recipient could not be converted to float", func(t *testing.T) {
		require.EqualError(
			t,
			bt.Send("asdfg", "test message"),
			"strconv.ParseFloat: parsing \"asdfg\": invalid syntax",
		)
	})

	t.Run("it sends a text message", func(t *testing.T) {
		require.NoError(t, bt.Send("1234567890", "test message"))
		require.Eventually(t, checkResponderCalled(&testMessageSent), time.Second, time.Millisecond)
	})

	t.Run("it should fail sending a text message", func(t *testing.T) {
		require.EqualError(t, bt.Send("1234567890", "fail message"), "telegram:  (0)")
	})

	t.Run("it send a text message longer than expected", func(t *testing.T) {
		require.NoError(t, bt.Send("1234567890", string(generateRandomString())))
		require.Eventually(t, checkResponderCalled(&testLongMessageSent), time.Second, time.Millisecond)
	})

	t.Run("it should send a picture", func(t *testing.T) {
		require.NoError(t, bt.Send("1234567890", bot.TelegramPhoto{
			Caption:  "test",
			FileID:   "123456",
			FileURL:  "http://image.url",
			FileSize: 1234,
		}))
		require.Eventually(t, checkResponderCalled(&photoSent), time.Second, time.Millisecond)
	})
}

func TestBot_GetFile(t *testing.T) {
	tlgmbot, err := tb.NewBot(tb.Settings{
		URL:   "https://api.telegram.mock",
		Token: "asdfg:12345",
		Poller: &tb.LongPoller{
			Timeout: 10 * time.Second,
		},
		Offline: true,
	})
	require.NoError(t, err)

	t.Run("it should fail when downloading the file", func(t *testing.T) {
		httpmock.RegisterResponder(
			"POST",
			"https://api.telegram.mock/botasdfg:12345/getFile",
			httpmock.NewStringResponder(404, ""),
		)

		bt := telegram.NewBot(tlgmbot)

		_, err = bt.GetFile("AZCDxruqG7J3iTM9")

		require.EqualError(t, err, "telebot: unexpected end of JSON input")
	})

	t.Run("it should download the file", func(t *testing.T) {
		fileJson, _ := os.ReadFile("testdata/getfile.json")
		httpmock.RegisterResponder(
			"POST",
			"https://api.telegram.mock/botasdfg:12345/getFile",
			httpmock.NewStringResponder(
				200,
				string(fileJson),
			),
		)

		icon, _ := os.ReadFile("testdata/td_icon.png")
		httpmock.RegisterResponder(
			"GET",
			"https://api.telegram.mock/file/botasdfg:12345/photos/file_4.jpg",
			httpmock.NewBytesResponder(
				200,
				icon,
			),
		)

		bt := telegram.NewBot(tlgmbot)

		_, err = bt.GetFile("AZCDxruqG7J3iTM9")

		require.NoError(t, err)
	})
}

func TestBot_ErrorHandler(t *testing.T) {
	var handled atomic.Value

	handled.Store(false)

	eh := func(e error, m bot.TelegramMessage) {
		handled.Store(true)
	}

	tlgmbot, err := tb.NewBot(tb.Settings{
		URL:   "https://api.telegram.mock",
		Token: "asdfg:12345",
		Poller: &tb.LongPoller{
			Timeout: 10 * time.Second,
		},
		OnError: func(err error, ctx tb.Context) {
			eh(err, bot.TelegramMessage{
				SenderID:  fmt.Sprintf("%v", ctx.Sender().ID),
				Text:      ctx.Text(),
				Payload:   ctx.Message().Payload,
				IsPrivate: ctx.Chat().Private,
			})
		},
		Offline: true,
	})
	require.NoError(t, err)

	telegram.NewBot(tlgmbot)

	c := new(telebot.Context)
	c.On("Sender").Return(&tb.User{})
	c.On("Text").Return("test")
	c.On("Message").Return(&tb.Message{})
	c.On("Chat").Return(&tb.Chat{})

	tlgmbot.OnError(nil, c)

	require.Equal(t, true, handled.Load())
}

func registerResponders(token string, testMessageSent, testLongMessageSent, photoSent, firstLongMessage *atomic.Value) {
	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://api.telegram.mock/bot%s/sendMessage", token),
		func(req *http.Request) (*http.Response, error) {
			buf := new(bytes.Buffer)
			_, _ = buf.ReadFrom(req.Body)

			//nolint:tagliatelle
			var requestBody struct {
				ChatID  string `json:"chat_id"`
				Text    string `json:"text"`
				ReplyTo string `json:"reply_to_message_id"`
			}
			_ = json.Unmarshal(buf.Bytes(), &requestBody)
			firstLongMessageBool, ok := firstLongMessage.Load().(bool)

			if requestBody.ChatID == "1234567890" && requestBody.Text == "test message" {
				testMessageSent.Store(true)
				messageSent, _ := os.ReadFile("testdata/sendmessage.json")

				return httpmock.NewStringResponse(200, string(messageSent)), nil
			} else if requestBody.ChatID == "1234567890" && requestBody.Text == "fail message" {
				return httpmock.NewStringResponse(429, "{}"), nil
			} else if len(requestBody.Text) == 4096 {
				firstLongMessage.Store(true)
				messageSent, _ := os.ReadFile("testdata/sendmessage.json")

				return httpmock.NewStringResponse(200, string(messageSent)), nil
			} else if ok && firstLongMessageBool && requestBody.ReplyTo == "59" {
				testLongMessageSent.Store(true)
				messageSent, _ := os.ReadFile("testdata/sendmessage.json")

				return httpmock.NewStringResponse(200, string(messageSent)), nil
			}

			return httpmock.NewStringResponse(500, "response not found"),
				errors.New("response not found")
		},
	)

	photoUpdate, _ := os.ReadFile("testdata/sendphoto.json")

	httpmock.RegisterResponder(
		"POST",
		fmt.Sprintf("https://api.telegram.mock/bot%s/sendPhoto", token),
		func(req *http.Request) (*http.Response, error) {
			photoSent.Store(true)

			return httpmock.NewStringResponse(200, string(photoUpdate)), nil
		},
	)
}

func generateRandomString() []byte {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"

	b := make([]byte, 5000)

	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}

	return b
}

func checkResponderCalled(b *atomic.Value) func() bool {
	return func() bool {
		b, ok := b.Load().(bool)

		return ok && b
	}
}
