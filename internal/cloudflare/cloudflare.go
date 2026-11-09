package cloudflare

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
)

type CloudflareIPsResponse struct {
	Success  bool     `json:"success"`
	Messages []string `json:"messages"`
	Errors   []string `json:"errors"`
	Result   struct {
		Ipv4Cidrs []string `json:"ipv4_cidrs"`
		Ipv6Cidrs []string `json:"ipv6_cidrs"`
		Etag      string   `json:"etag"`
	} `json:"result"`
}

func GetIPs() ([]string, []string, error) {
	resp, err := http.Get("https://api.cloudflare.com/client/v4/ips")
	if err != nil {
		return nil, nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, nil, err
	}

	var result CloudflareIPsResponse
	if err := json.Unmarshal(body, &result); err != nil {
		return nil, nil, err
	}

	if !result.Success {
		return nil, nil, fmt.Errorf(strings.Join(result.Errors, "|"))
	}

	return result.Result.Ipv4Cidrs, result.Result.Ipv6Cidrs, nil
}
