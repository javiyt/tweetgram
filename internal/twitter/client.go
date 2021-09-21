package twitter

import gt "github.com/javiyt/go-twitter/twitter"

type Client struct {
	tc *gt.Client
}

func NewTwitterClient(tc *gt.Client) *Client {
	return &Client{tc: tc}
}

func (c *Client) SendUpdate(string) error {
	return nil
}
