# 📊 Vnstat Telegram Monitor

A minimal and efficient Go program that sends a monthly traffic usage report via Telegram using `vnstat`. The report includes:

- RX/TX traffic for a specified interface (default: `eth0`)
- TX usage compared to a defined limit (default: 1 TiB)
- Graphical progress bar and emoji status
- Public IPv4 address and location info (city, region, ISP)

## 🔧 Configuration

This tool is designed to run on a schedule via `cron`. It uses the following environment variables:

| Variable            | Description                                  | Default     |
|---------------------|----------------------------------------------|-------------|
| `TELEGRAM_BOT_TOKEN`| Your Telegram bot token (required)           | -           |
| `TELEGRAM_CHAT_ID`  | Telegram chat ID to send the message to      | -           |
| `INTERFACE`         | Network interface to monitor                 | `eth0`      |
| `LIMIT_GIB`         | TX traffic limit in GiB                      | `1024` (1 TiB) |

## 🧪 Example Output

📊VNSTAT
Usage on eth0 in May:

⬇️ RX: 710.5 GB

⬆️ TX: 675.2 GB (limit: 1024 GiB)

Total: 1385.7 GB

TX Limit: 🟡 65.96% used

[████████████░░░░░░░░░░░░]

🌐 Public IP: 123.123.123.123 🇨🇦

📍 Location: Montréal, Quebec

🏢 ISP: AS215311 Regxa Company for Information Technology Ltd


## 🐧 Dependencies

- [`vnstat`](https://humdi.net/vnstat/) must be installed and tracking the specified interface.
- Tested on Linux systems.

## 📅 Usage with crontab

To schedule this script to run monthly:

```bash
0 0 1 * * /path/to/vnstat-telegram-monitor
```
🛡️ License

MIT License
