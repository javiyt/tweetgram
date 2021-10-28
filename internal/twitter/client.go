package twitter

import (
	"fmt"
	"io"
	"net/http"
	"strings"

	gt "github.com/javiyt/go-twitter/twitter"
	"github.com/javiyt/twitter-text-go/validate"
)

const (
	tweetMaxLength = 280
	joinString     = "..."
)

type Client struct {
	tc *gt.Client
}

func NewTwitterClient(tc *gt.Client) *Client {
	return &Client{tc: tc}
}

func (c *Client) SendUpdate(s string) error {
	return c.publishTweet(s, &gt.StatusUpdateParams{})
}

func (c *Client) SendUpdateWithPhoto(s string, pic []byte) error {
	uploadResult, resp, err := c.tc.Media.Upload(pic, http.DetectContentType(pic))

	defer func() { _ = resp.Body.Close() }()

	if err != nil {
		buf := new(strings.Builder)
		_, _ = io.Copy(buf, resp.Body)

		return fmt.Errorf(
			"error sending status update: %w. Response status code: %v and body: %s",
			err,
			resp.StatusCode,
			buf.String(),
		)
	}

	return c.publishTweet(s, &gt.StatusUpdateParams{MediaIds: []int64{uploadResult.MediaID}})
}

func (c *Client) publishTweet(s string, params *gt.StatusUpdateParams) error {
	err := validate.ValidateTweet(s)
	switch err.(type) {
	case validate.EmptyError:
		return nil
	case validate.InvalidCharacterError:
		return fmt.Errorf("error sending status update: %w", err)
	}

	var replyToID int64
	for _, ts := range c.chunks(s, tweetMaxLength-len(joinString)) {
		if replyToID > 0 {
			params.InReplyToStatusID = replyToID
		}

		if len(ts) == tweetMaxLength-len(joinString) {
			ts += joinString
		}

		tweet, resp, err := c.tc.Statuses.Update(ts, params)

		if err != nil {
			buf := new(strings.Builder)
			_, _ = io.Copy(buf, resp.Body)
			_ = resp.Body.Close()

			return fmt.Errorf(
				"error sending status update: %w. Response status code: %v and body: %s",
				err,
				resp.StatusCode,
				buf.String(),
			)
		}

		replyToID = tweet.ID
	}

	return nil
}

func (c *Client) chunks(s string, chunkSize int) []string {
	if chunkSize >= len(s) {
		return []string{s}
	}

	chunks := make([]string, 0, (len(s)-1)/chunkSize+1)
	currentLen := 0
	currentStart := 0

	for i := range s {
		if currentLen == chunkSize {
			chunks = append(chunks, s[currentStart:i])
			currentLen = 0
			currentStart = i
		}
		currentLen++
	}

	chunks = append(chunks, s[currentStart:])

	return chunks
}
