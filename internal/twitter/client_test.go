package twitter_test

import (
	"net/http"
	"testing"
	"time"

	"github.com/dghubble/oauth1"
	"github.com/jarcoal/httpmock"
	gt "github.com/javiyt/go-twitter/twitter"
	"github.com/javiyt/tweettgram/internal/twitter"
	"github.com/stretchr/testify/require"
)

func TestSendUpdate(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

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
	})

	t.Run("it should send status update to twitter API", func(t *testing.T) {
		require.NoError(t, client.SendUpdate("testing"))
	})
}
