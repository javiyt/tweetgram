package twitter_test

import (
	"testing"

	gt "github.com/javiyt/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/javiyt/tweettgram/internal/twitter"
	"github.com/stretchr/testify/require"
)

func TestSendUpdate(t *testing.T) {
	httpClient := oauth1.NewConfig("consumerKey", "consumerSecret").
		Client(oauth1.NoContext, oauth1.NewToken("accessToken", "accessSecret"))

	client := twitter.NewTwitterClient(gt.NewClient(httpClient))

	require.Nil(t, client.SendUpdate("testing"))
}
