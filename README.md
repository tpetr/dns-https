# dns-https

A simple [DNS-over-HTTPS](https://en.wikipedia.org/wiki/DNS_over_HTTPS) proxy written in Go.

## Installation

Binaries can be found on the [releases page](https://github.com/tpetr/dns-https/releases).

Or you can install via [Homebrew](https://brew.sh/): `brew install tpetr/tap/dns-https`

## Usage

```
Usage of dns-https:
  -f, --fallback string    fallback DNS address for resolving upstreams (default "1.1.1.1:53")
  -l, --listen string      address to listen on (default ":10053")
  -t, --timeout duration   HTTP timeout (default 5s)
  -v, --verbose            enable verbose logging
```

## Example

```bash
sudo dns-https -v -l 127.0.0.1:53 https://dns.quad9.net/dns-query https://cloudflare-dns.com/dns-query
```

This command configures `dns-https` listen locally on the standard DNS port (53, hence `sudo`) and randomly proxy DNS requests to [Quad9](https://www.quad9.net/doh-quad9-dns-servers/) or [Cloudflare's](https://developers.cloudflare.com/1.1.1.1/dns-over-https/) DNS-over-HTTPS endpoints.

To test this out, run `dig google.com @127.0.0.1` in another terminal window. You should see output like this from `dns-https`:

```
INFO[2018-10-22T13:15:45-04:00] Proxying to 2 URL(s) with a 5s timeout
INFO[2018-10-22T13:15:45-04:00] Listening on 127.0.0.1:53
DEBU[2018-10-22T13:16:39-04:00] Proxied in 209ms                              Domain=google.com. Id=33104 upstream="https://cloudflare-dns.com/dns-query"
```
