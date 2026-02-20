# meteocli — MeteoSwiss CLI

A command-line interface for [MeteoSwiss](https://www.meteoswiss.admin.ch/), the Swiss Federal Office of Meteorology and Climatology.

Get current weather conditions, multi-day forecasts, and active warnings for any Swiss postal code — directly in your terminal.

> **Note:** This tool uses the unofficial MeteoSwiss app backend API, which is reverse-engineered from the official iOS/Android app. It is not affiliated with or endorsed by MeteoSwiss.

## Install / Build

```bash
go build -o ./dist/meteocli ./cmd/meteocli
```

Or install directly:

```bash
go install github.com/user/meteocli/cmd/meteocli@latest
```

## Quick Start

```bash
# Current weather in Zurich (PLZ 8000)
meteocli weather --zip 8000

# 7-day forecast for Bern
meteocli forecast --zip 3000

# 3-day forecast for Geneva, JSON output
meteocli forecast --zip 1200 --days 3 --json

# Active weather warnings for Switzerland
meteocli warnings

# Only warnings at level 3 (Considerable) and above
meteocli warnings --min-level 3
```

## Commands

### `weather`

Shows current observed conditions for a Swiss postal code.

```
meteocli weather --zip <PLZ>
```

Output includes: current temperature, weather description, and a summary of today's high/low and precipitation.

### `forecast`

Shows a multi-day (up to 10 days) forecast.

```
meteocli forecast --zip <PLZ> [--days N]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--zip` | required | Swiss postal code (1000–9999) |
| `--days` | 7 | Number of days to display (1–10) |

### `warnings`

Lists all active MeteoSwiss weather warnings.

```
meteocli warnings [--min-level N]
```

| Flag | Default | Description |
|------|---------|-------------|
| `--min-level` | 1 | Minimum warning level (1=Minor … 5=Very high) |

## Global Flags

| Flag | Description |
|------|-------------|
| `--json` | Output machine-readable JSON instead of formatted text |
| `--version` | Print version and exit |

## Warning Levels

| Level | Label |
|-------|-------|
| 1 | Minor |
| 2 | Moderate |
| 3 | Considerable |
| 4 | High |
| 5 | Very high |

## Warning Types

Wind, Thunderstorm, Rain, Snow, Slippery roads, Frost, Heat, Avalanche, Fire danger, Flooding, UV.

## Postal Codes

Swiss postal codes run from 1000 to 9999. A few examples:

| City | PLZ |
|------|-----|
| Zurich | 8000 |
| Bern | 3000 |
| Geneva | 1200 |
| Basel | 4000 |
| Lausanne | 1000 |
| Lucerne | 6000 |
| St. Gallen | 9000 |

## License

MIT — see [LICENSE](LICENSE).
