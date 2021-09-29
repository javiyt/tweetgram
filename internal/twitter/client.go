package twitter

import (
	"fmt"
	gt "github.com/javiyt/go-twitter/twitter"
	"io"
	"net/http"
	"strings"
)

type Client struct {
	tc *gt.Client
}

func NewTwitterClient(tc *gt.Client) *Client {
	return &Client{tc: tc}
}

func (c *Client) SendUpdate(s string) error {
	if _, resp, err := c.tc.Statuses.Update(s, &gt.StatusUpdateParams{}); err != nil {
		buf := new(strings.Builder)
		_, _ = io.Copy(buf, resp.Body)
		return fmt.Errorf(
			"error sending status update: %w. Response status code: %v and body: %s",
			err,
			resp.StatusCode,
			buf.String(),
		)
	}

	return nil
}

func (c *Client) SendUpdateWithPhoto(s string, pic []byte) error {
	uploadResult, _, err := c.tc.Media.Upload(pic, http.DetectContentType(pic))
	if err != nil {
		return err
	}

	if _, resp, err := c.tc.Statuses.Update(s, &gt.StatusUpdateParams{
		MediaIds: []int64{uploadResult.MediaID},
	}); err != nil {
		buf := new(strings.Builder)
		_, _ = io.Copy(buf, resp.Body)
		return fmt.Errorf(
			"error sending status update: %w. Response status code: %v and body: %s",
			err,
			resp.StatusCode,
			buf.String(),
		)
	}

	return nil
}
