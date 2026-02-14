# Configuration

Default path:

- `~/.config/attio-cli/config.json`

Override path:

- `ATTIO_CONFIG_PATH=/path/to/config.json`

## Format

```json
{
  "default_profile": "default",
  "profiles": {
    "default": {
      "api_key": "sk_live_...",
      "base_url": "https://api.attio.com"
    },
    "staging": {
      "api_key": "sk_test_...",
      "base_url": "https://staging-api.attio.com"
    }
  }
}
```

Notes:

- `api_key` in config is lowest-priority auth source.
- `base_url` in profile is overridden by `ATTIO_BASE_URL`.
- `default_profile` is used when `--profile` is not set.
