package twitter_test

import (
	"bytes"
	"math/rand"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/dghubble/oauth1"
	"github.com/jarcoal/httpmock"
	gt "github.com/javiyt/go-twitter/twitter"
	"github.com/javiyt/tweetgram/internal/twitter"
	"github.com/stretchr/testify/require"
)

func TestSendUpdate(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	rand.Seed(time.Now().UnixNano())
	letterRunes := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ ")
	b := make([]rune, 300)
	for i := range b {
		b[i] = letterRunes[rand.Intn(len(letterRunes))]
	}
	longTweet := string(b)

	httpmock.RegisterResponder(
		"POST",
		"https://api.twitter.com/1.1/statuses/update.json",
		func(req *http.Request) (*http.Response, error) {
			_ = req.ParseForm()

			var resp *http.Response
			var err error
			if req.Form.Get("status") == "testing" {
				resp, err = httpmock.NewJsonResponse(200, gt.Tweet{
					ID:        1050118621198921700,
					IDStr:     "1050118621198921728",
					CreatedAt: time.Now().UTC().Format(time.RubyDate),
					Text:      "testing",
					FullText:  "testing",
				})
				if err != nil {
					return httpmock.NewStringResponse(http.StatusInternalServerError, ""), nil
				}
			} else if  req.Form.Get("status") == longTweet[:277] + "..." {
				resp, err = httpmock.NewJsonResponse(200, gt.Tweet{
					ID:        1445823463904798049,
					IDStr:     "1445823463904798049",
					CreatedAt: time.Now().UTC().Format(time.RubyDate),
					Text:      longTweet[:280],
					FullText:  longTweet[:280],
				})
				if err != nil {
					return httpmock.NewStringResponse(http.StatusInternalServerError, ""), nil
				}
			} else if req.Form.Get("status") == longTweet[277:] && req.Form.Get("in_reply_to_status_id") == "1445823463904798049" {
				resp, err = httpmock.NewJsonResponse(200, gt.Tweet{
					ID:        1445823463904798051,
					IDStr:     "1445823463904798051",
					CreatedAt: time.Now().UTC().Format(time.RubyDate),
					Text:      longTweet[280:],
					FullText:  longTweet[280:],
				})
				if err != nil {
					return httpmock.NewStringResponse(http.StatusInternalServerError, ""), nil
				}
			} else {
				return httpmock.NewStringResponse(http.StatusForbidden, ""), nil
			}

			return resp, nil
		},
	)

	httpmock.RegisterResponder(
		"POST",
		"https://upload.twitter.com/1.1/media/upload.json",
		func(req *http.Request) (*http.Response, error) {
			_ = req.ParseForm()

			var resp *http.Response
			var err error
			if req.Form.Get("media_type") == "image/png" || req.Form.Get("media_id") == "12345" {
				resp, err = httpmock.NewJsonResponse(200, gt.MediaUploadResult{MediaID: 12345})
				if err != nil {
					return httpmock.NewStringResponse(http.StatusInternalServerError, ""), nil
				}
			} else {
				return httpmock.NewStringResponse(http.StatusForbidden, ""), nil
			}

			return resp, nil
		},
	)

	httpClient := oauth1.NewConfig("consumerKey", "consumerSecret").
		Client(oauth1.NoContext, oauth1.NewToken("accessToken", "accessSecret"))

	client := twitter.NewTwitterClient(gt.NewClient(httpClient))

	t.Run("it should fail when error happens on Twitter API", func(t *testing.T) {
		require.EqualError(
			t,
			client.SendUpdate("it should fail"),
			"error sending status update: EOF. Response status code: 403 and body: ",
		)
		require.Equal(t, 1, httpmock.GetTotalCallCount())
		httpmock.ZeroCallCounters()
	})

	t.Run("it should not send status update when status is empty", func(t *testing.T) {
		require.NoError(
			t,
			client.SendUpdate(""),
		)
		require.Zero(t, httpmock.GetTotalCallCount())
		httpmock.ZeroCallCounters()
	})

	t.Run("it should fail when invalid character in status update", func(t *testing.T) {
		require.EqualError(
			t,
			client.SendUpdate("test \uFFFE"),
			"error sending status update: Invalid chararcter [\uFFFE] found at byte offset 5",
		)
		require.Zero(t, httpmock.GetTotalCallCount())
		httpmock.ZeroCallCounters()
	})


	t.Run("it should send status update to Twitter API", func(t *testing.T) {
		require.NoError(t, client.SendUpdate("testing"))
		require.Equal(t, 1, httpmock.GetTotalCallCount())
		httpmock.ZeroCallCounters()
	})

	t.Run("it should send long status update to Twitter API", func(t *testing.T) {
		require.NoError(t, client.SendUpdate(longTweet))
		require.Equal(t, 2, httpmock.GetTotalCallCount())
		httpmock.ZeroCallCounters()
	})

	t.Run("it should fail when media type not allowed by Twitter", func(t *testing.T) {
		file, _ := os.Open("testdata/icon_gopher.jpg")
		defer func() { _ = file.Close() }()
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(file)

		require.EqualError(
			t,
			client.SendUpdateWithPhoto("testing", buf.Bytes()),
			"error sending status update: EOF. Response status code: 403 and body: ",
		)
	})

	t.Run("it should fail sending status update with photo to Twitter API", func(t *testing.T) {
		file, _ := os.Open("testdata/test.png")
		defer func() { _ = file.Close() }()
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(file)

		require.EqualError(
			t,
			client.SendUpdateWithPhoto("it should fail", buf.Bytes()),
			"error sending status update: EOF. Response status code: 403 and body: ",
		)
	})

	t.Run("it should send status update with photo to Twitter API", func(t *testing.T) {
		file, _ := os.Open("testdata/test.png")
		defer func() { _ = file.Close() }()
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(file)

		require.NoError(
			t,
			client.SendUpdateWithPhoto("testing", buf.Bytes()),
		)
	})
}
