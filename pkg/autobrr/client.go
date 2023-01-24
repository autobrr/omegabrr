package autobrr

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httputil"
	"strconv"
	"time"

	"github.com/rs/zerolog/log"

	"github.com/pkg/errors"
)

type Client struct {
	Host   string
	APIKey string

	BasicUser string
	BasicPass string

	client *http.Client
}

func NewClient(host string, apikey string) *Client {
	if host == "" {
		log.Fatal().Msg("autobrr: missing host")
	} else if apikey == "" {
		log.Fatal().Msg("autobrr: missing apikey")
	}

	c := &Client{
		Host:   host,
		APIKey: apikey,
	}

	customTransport := http.DefaultTransport.(*http.Transport).Clone()
	customTransport.TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	c.client = &http.Client{
		Timeout:   60 * time.Second,
		Transport: customTransport,
	}

	return c
}

func (c *Client) SetBasicAuth(user, pass string) {
	if user != "" {
		c.BasicUser = user
	}
	if pass != "" {
		c.BasicPass = pass
	}
}

func (c *Client) baseClient() {

}

func (c *Client) Test(ctx context.Context) error {
	if _, err := c.GetFilters(ctx); err != nil {
		return err
	}
	return nil
}

func (c *Client) GetFilters(ctx context.Context) ([]Filter, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, c.Host+"/api/filters", nil)
	if err != nil {
		return nil, err
	}

	req.SetBasicAuth(c.BasicUser, c.BasicPass)
	req.Header.Add("X-API-Token", c.APIKey)

	res, err := c.client.Do(req)
	if err != nil {
		return nil, err
	}

	if res.StatusCode != http.StatusOK {
		return nil, errors.New("something went wrong")
	}

	body, err := io.ReadAll(res.Body)
	if err != nil {
		return nil, err
	}

	var filters []Filter

	if err := json.Unmarshal(body, &filters); err != nil {
		return nil, err
	}

	return filters, nil
}

func (c *Client) UpdateFilterByID(ctx context.Context, filterID int, filter UpdateFilter) error {
	id := strconv.Itoa(filterID)

	body, err := json.Marshal(filter)
	if err != nil {
		return err
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPatch, c.Host+"/api/filters/"+id, bytes.NewBuffer(body))
	if err != nil {
		return err
	}

	req.SetBasicAuth(c.BasicUser, c.BasicPass)
	req.Header.Add("X-API-Token", c.APIKey)

	res, err := c.client.Do(req)
	if err != nil {
		return err
	}
	defer res.Body.Close()

	dump, err := httputil.DumpResponse(res, true)
	if err != nil {
		return errors.Wrap(err, "could not dump response")
	}

	log.Trace().Msgf("update filter response: %v", string(dump))

	if res.StatusCode != http.StatusNoContent {
		return errors.New("bad status")
	}

	return nil
}

type Filter struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	UpdatedAt time.Time `json:"updated_at"`
	CreatedAt time.Time `json:"created_at"`
}

type UpdateFilter struct {
	Shows         string `json:"shows"`
	Albums        string `json:"albums"`
	MatchReleases string `json:"match_releases"`
}
