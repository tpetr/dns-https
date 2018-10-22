# dns-https

A simple [DNS-over-HTTPS](https://en.wikipedia.org/wiki/DNS_over_HTTPS) proxy.

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

Example:

```bash
sudo dns-https -l 127.0.0.1:53 https://9.9.9.9/dns-query https://cloudflare-dns.com/dns-query
```

This command makes `dns-https` listen locally on the DNS port (53, hence `sudo`) and randomly proxy DNS requests to Quad9 or Cloudflare's DNS-over-HTTPS endpoints. To make your computer use `dns-https`, set your DNS server to `127.0.0.1`.