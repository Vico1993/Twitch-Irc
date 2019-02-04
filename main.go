package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

type Channel struct {
	ID          int    `json:"_id"`
	DisplayName string `json:"display_name"`
	Followers   int    `json:"followers"`
	Views       int    `json:"views"`
}

type Streams struct {
	ID      int     `json:"_id"`
	Game    string  `json:"game"`
	Viewers int     `json:"viewers"`
	Channel Channel `json:"channel"`
}

type TwitchResponse struct {
	Total   int       `json:"_total"`
	Streams []Streams `json:"streams"`
}

type Client struct {
	httpClient *http.Client
}

func NewClient(httpClient *http.Client) *Client {
	if httpClient == nil {
		httpClient = http.DefaultClient
	}
	c := &Client{httpClient: httpClient}

	return c
}

func (c *Client) newRequest(method, path string, body interface{}) (*http.Request, error) {
	var buf io.ReadWriter
	if body != nil {
		buf = new(bytes.Buffer)
		err := json.NewEncoder(buf).Encode(body)
		if err != nil {
			return nil, err
		}
	}
	req, err := http.NewRequest(method, baseURL+path, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Client-ID", clientID)
	// req.Header.Set("Authorization:", token)
	return req, nil
}

func (c *Client) do(req *http.Request, v interface{}) (*http.Response, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	err = json.NewDecoder(resp.Body).Decode(v)
	return resp, err
}

func main() {
	cl := NewClient(nil)
	r, err := cl.newRequest("GET", "amouranth", nil)
	if err != nil {
		fmt.Println("Error here : " + err.Error())
		return
	}

	rjson := TwitchResponse{}
	_, err = cl.do(r, &rjson)
	if err != nil {
		fmt.Println("Error here too : " + err.Error())
		return
	}

	for _, s := range rjson.Streams {
		fmt.Println(s.Viewers)
		fmt.Println(s.Game)
		fmt.Println(s.ID)
	}
}
