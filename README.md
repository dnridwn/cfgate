# cfgate

Keep your Tencent Cloud Lighthouse firewall in sync with Cloudflare's IP ranges.

`cfgate` fetches Cloudflare's official IP ranges and updates the firewall rules of a Tencent Cloud Lighthouse instance so that only Cloudflare IPs are allowed to reach your HTTP (port 80) and HTTPS (port 443) services. This keeps your origin locked down to Cloudflare's proxy and automatically stays current as Cloudflare publishes new ranges.

## How it works

1. Fetches Cloudflare's IPv4 (and optionally IPv6) CIDRs from the [Cloudflare IP API](https://api.cloudflare.com/client/v4/ips).
2. Builds `ACCEPT` firewall rules for **TCP ports 80 and 443**, one per Cloudflare CIDR, tagged with the description `"Auto-set by cfgate"`.
3. Lists the current firewall rules on the target Tencent Cloud Lighthouse instance.
4. Preserves all existing rules that are **not** TCP 80 or 443, and merges them with the new Cloudflare rules.
5. Applies the merged ruleset back to the instance using `ModifyFirewallRules`, including the latest `FirewallVersion` to avoid clobbering concurrent changes.

## Why

When you proxy a site through Cloudflare, you typically want your origin server to accept HTTP/HTTPS traffic **only** from Cloudflare's IPs — so attackers can't bypass Cloudflare's DDoS protection and WAF by hitting your origin directly. `cfgate` automates keeping that allow-list up to date, since Cloudflare's published ranges change over time.

## Prerequisites

- A Tencent Cloud Lighthouse instance.
- Tencent Cloud API credentials (`SecretId` and `SecretKey`) with permissions to manage Lighthouse firewall rules.
- Go 1.27+ (only if building from source).

## Install

### From releases

Download the appropriate tarball for your platform from the [Releases](../../releases) page, then verify it against the published `checksums.txt`.

```sh
tar -xzf cfgate_vX.Y.Z_linux_amd64.tar.gz
chmod +x cfgate
```

### From source

```sh
go build -o cfgate ./cmd/
```

## Usage

```sh
export TENCENTCLOUD_SECRET_ID="your-secret-id"
export TENCENTCLOUD_SECRET_KEY="your-secret-key"

./cfgate -r <region> -i <instance-id>
```

To also allow Cloudflare's IPv6 ranges:

```sh
./cfgate -r <region> -i <instance-id> --include-ipv6
```

Check the version:

```sh
./cfgate version
```

## Environment variables

| Variable                  | Required | Description                          |
| ------------------------- | -------- | ------------------------------------ |
| `TENCENTCLOUD_SECRET_ID`  | Yes      | Tencent Cloud API SecretId           |
| `TENCENTCLOUD_SECRET_KEY` | Yes      | Tencent Cloud API SecretKey          |

## Flags

| Flag              | Required | Default | Description                                      |
| ----------------- | -------- | ------- | ------------------------------------------------ |
| `-r, --region`    | Yes      | —       | Tencent Cloud region of the target instance     |
| `-i, --instance-id` | Yes    | —       | ID of the target Lighthouse instance             |
| `--include-ipv6`  | No       | `false` | Also add rules for Cloudflare's IPv6 CIDRs       |

## What it does to your firewall

`cfgate` rewrites the firewall rules for **TCP ports 80 and 443** to accept only Cloudflare's IP ranges, while preserving all your other existing rules (non-80/443-TCP rules).

> ⚠️ **Warning:** Tencent Cloud's `ModifyFirewallRules` is effectively a full replacement of the firewall ruleset. `cfgate` mitigates this by reading the current rules first and merging, but you should review your rules after running it the first time to make sure the result matches your expectations.
