package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/Arc-Services/Arc/shared/cache"
)

type Client struct {
	BaseURL      string
	ClientID     string
	Timeout      time.Duration
	client       *http.Client
	CacheEnabled bool
	CacheTTL     time.Duration
}

type Request struct {
	Action string      `json:"action"`
	Key    string      `json:"key,omitempty"`
	Value  interface{} `json:"value,omitempty"`
	Prefix string      `json:"prefix,omitempty"`
	TTL    int64       `json:"ttl,omitempty"`
	Limit  int         `json:"limit,omitempty"`
	Ops    []Op        `json:"ops,omitempty"`
}

type Op struct {
	Action string      `json:"action"`
	Key    string      `json:"key"`
	Value  interface{} `json:"value,omitempty"`
	TTL    int64       `json:"ttl,omitempty"`
}

type Response struct {
	Value   interface{}            `json:"value,omitempty"`
	Keys    []string               `json:"keys,omitempty"`
	Results interface{}            `json:"results,omitempty"`
	Count   int                    `json:"count,omitempty"`
	Success bool                   `json:"success,omitempty"`
	Error   string                 `json:"error,omitempty"`
	Message string                 `json:"message,omitempty"`
	Raw     map[string]interface{} `json:"-"`
}

func New(baseURL string) *Client {
	return &Client{
		BaseURL:      baseURL + "/internal/private/api/v1/database/",
		Timeout:      10 * time.Second,
		client:       &http.Client{Timeout: 10 * time.Second},
		CacheEnabled: true,
		CacheTTL:     5 * time.Minute,
	}
}

func (c *Client) getKey(action, key string) string {
	return fmt.Sprintf("db:%s:%s:%s", c.ClientID, action, key)
}

func (c *Client) signedRequest(clientID string, body []byte) (*http.Request, error) {
	httpReq, err := http.NewRequest("POST", c.BaseURL, bytes.NewBuffer(body))
	if err != nil {
		return nil, err
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.Header.Set("X-Arc-Client", clientID)

	sig, ts, nonce := SignRequest(clientID, body)
	httpReq.Header.Set(HeaderSignature, sig)
	httpReq.Header.Set(HeaderTimestamp, ts)
	httpReq.Header.Set(HeaderNonce, nonce)

	return httpReq, nil
}

func (c *Client) executeRequest(httpReq *http.Request) (*Response, error) {
	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	if resp.StatusCode >= 400 {
		var errResp Response
		json.Unmarshal(data, &errResp)
		return nil, fmt.Errorf("request failed: %s", errResp.Message)
	}

	var result Response
	if err := json.Unmarshal(data, &result); err != nil {
		return nil, err
	}

	var raw map[string]interface{}
	json.Unmarshal(data, &raw)
	result.Raw = raw
	result.Success = true

	return &result, nil
}

func (c *Client) do(req Request) (*Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := c.signedRequest(c.ClientID, body)
	if err != nil {
		return nil, err
	}

	return c.executeRequest(httpReq)
}

func (c *Client) GetFor(clientID, key string) (*Response, error) {
	if c.CacheEnabled {
		cacheKey := fmt.Sprintf("db:%s:get:%s", clientID, key)
		var res Response
		if err := cache.Get(cacheKey, &res); err == nil {
			return &res, nil
		}
	}

	response, err := c.doFor(clientID, Request{Action: "get", Key: key})
	if err != nil {
		return nil, err
	}

	if c.CacheEnabled && response.Success {
		cacheKey := fmt.Sprintf("db:%s:get:%s", clientID, key)
		cache.Set(cacheKey, response, c.CacheTTL)
	}

	return response, nil
}

func (c *Client) doFor(clientID string, req Request) (*Response, error) {
	body, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpReq, err := c.signedRequest(clientID, body)
	if err != nil {
		return nil, err
	}

	return c.executeRequest(httpReq)
}

func (c *Client) Get(key string) (*Response, error) {
	if c.CacheEnabled {
		cacheKey := c.getKey("get", key)
		var res Response

		if err := cache.Get(cacheKey, &res); err == nil {
			return &res, nil
		}
	}

	response, err := c.do(Request{Action: "get", Key: key})
	if err != nil {
		return nil, err
	}

	if c.CacheEnabled && response.Success {
		cacheKey := c.getKey("get", key)
		cache.Set(cacheKey, response, c.CacheTTL)
	}

	return response, nil
}

func (c *Client) Set(key string, value interface{}) (*Response, error) {
	response, err := c.do(Request{Action: "set", Key: key, Value: value})
	if err != nil {
		return nil, err
	}

	if c.CacheEnabled && response.Success {
		cacheKey := c.getKey("get", key)
		cache.Delete(cacheKey)
	}

	return response, nil
}

func (c *Client) SetTTL(key string, value interface{}, ttl int64) (*Response, error) {
	response, err := c.do(Request{Action: "set", Key: key, Value: value, TTL: ttl})
	if err != nil {
		return nil, err
	}

	if c.CacheEnabled && response.Success {
		cacheKey := c.getKey("get", key)
		cache.Delete(cacheKey)
	}

	return response, nil
}

func (c *Client) Delete(key string) (*Response, error) {
	response, err := c.do(Request{Action: "delete", Key: key})
	if err != nil {
		return nil, err
	}

	if c.CacheEnabled && response.Success {
		cacheKey := c.getKey("get", key)
		cache.Delete(cacheKey)
	}

	return response, nil
}

func (c *Client) Keys(prefix string) (*Response, error) {
	if c.CacheEnabled {
		cacheKey := fmt.Sprintf("db:%s:keys:%s", c.ClientID, prefix)
		var res Response
		if err := cache.Get(cacheKey, &res); err == nil {
			return &res, nil
		}
	}

	resp, err := c.do(Request{
		Action: "keys",
		Prefix: prefix,
	})
	if err != nil {
		return nil, err
	}

	if c.CacheEnabled && resp.Success {
		cacheKey := fmt.Sprintf("db:%s:keys:%s", c.ClientID, prefix)
		cache.Set(cacheKey, resp, 30*time.Second)
	}

	return resp, nil
}

func (c *Client) Scan(prefix string, limit int) (*Response, error) {
	if c.CacheEnabled {
		cacheKey := fmt.Sprintf("db:%s:scan:%s:%d", c.ClientID, prefix, limit)
		var res Response
		if err := cache.Get(cacheKey, &res); err == nil {
			return &res, nil
		}
	}

	resp, err := c.do(Request{
		Action: "scan",
		Prefix: prefix,
		Limit:  limit,
	})
	if err != nil {
		return nil, err
	}

	if c.CacheEnabled && resp.Success {
		cacheKey := fmt.Sprintf("db:%s:scan:%s:%d", c.ClientID, prefix, limit)
		cache.Set(cacheKey, resp, 30*time.Second)
	}

	return resp, nil
}

func (c *Client) Batch(ops []Op) (*Response, error) {
	response, err := c.do(Request{Action: "batch", Ops: ops})
	if err != nil {
		return nil, err
	}

	if c.CacheEnabled && response.Success {
		for _, op := range ops {
			if op.Action == "set" || op.Action == "delete" {
				cacheKey := c.getKey("get", op.Key)
				cache.Delete(cacheKey)
			}
		}
	}

	return response, nil
}

func (c *Client) Stats() (*Response, error) {
	return c.do(Request{Action: "stats"})
}

func (c *Client) SetCacheEnabled(enabled bool) {
	c.CacheEnabled = enabled
}

func (c *Client) SetCacheTTL(ttl time.Duration) {
	c.CacheTTL = ttl
}

func (c *Client) InvalidateCache() error {
	if !c.CacheEnabled {
		return nil
	}
	pattern := fmt.Sprintf("db:%s:*", c.ClientID)
	return cache.InvalidatePattern(pattern)
}
