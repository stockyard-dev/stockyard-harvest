# Stockyard Harvest

**Self-hosted harvest and crop tracking for farms**

Part of the [Stockyard](https://stockyard.dev) family of self-hosted tools.

## Quick Start

```bash
curl -fsSL https://stockyard.dev/tools/harvest/install.sh | sh
```

Or with Docker:

```bash
docker run -p 9810:9810 -v harvest_data:/data ghcr.io/stockyard-dev/stockyard-harvest
```

Open `http://localhost:9810` in your browser.

## Configuration

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `9810` | HTTP port |
| `DATA_DIR` | `./harvest-data` | SQLite database directory |
| `STOCKYARD_LICENSE_KEY` | *(empty)* | License key for unlimited use |

## Free vs Pro

| | Free | Pro |
|-|------|-----|
| Limits | 5 records | Unlimited |
| Price | Free | Included in bundle or $29.99/mo individual |

Get a license at [stockyard.dev](https://stockyard.dev).

## License

Apache 2.0
