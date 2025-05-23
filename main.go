package main

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/joho/godotenv"
)

type VnstatOutput struct {
	Interfaces []struct {
		Name    string `json:"name"`
		Traffic struct {
			Month []struct {
				Date struct {
					Year  int `json:"year"`
					Month int `json:"month"`
				} `json:"date"`
				RX uint64 `json:"rx"`
				TX uint64 `json:"tx"`
			} `json:"month"`
		} `json:"traffic"`
	} `json:"interfaces"`
}

type IPInfo struct {
	IP      string `json:"ip"`
	City    string `json:"city"`
	Region  string `json:"region"`
	Country string `json:"country"`
	Org     string `json:"org"`
}

func formatBytes(b uint64) string {
	gb := float64(b) / (1024 * 1024 * 1024)
	return fmt.Sprintf("%.1f GB", gb)
}

func getIPInfo() (IPInfo, error) {
	client := &http.Client{
		Transport: &http.Transport{
			DialContext: func(ctx context.Context, network, address string) (net.Conn, error) {
				d := net.Dialer{}
				return d.DialContext(ctx, "tcp4", address)
			},
		},
	}

	req, err := http.NewRequest("GET", "https://ipinfo.io", nil)
	if err != nil {
		return IPInfo{}, err
	}
	req.Header.Set("User-Agent", "curl/7.81.0")

	resp, err := client.Do(req)
	if err != nil {
		return IPInfo{}, err
	}
	defer resp.Body.Close()

	var info IPInfo
	err = json.NewDecoder(resp.Body).Decode(&info)
	if err != nil {
		return IPInfo{}, err
	}

	if ip := net.ParseIP(info.IP); ip != nil && ip.To4() == nil {
		info.IP = ""
	}
	return info, nil
}

func getVnstatData(interfaceName string) (string, uint64, uint64, string, error) {
	cmd := exec.Command("vnstat", "-i", interfaceName, "--json", "m", "1")
	output, err := cmd.Output()
	if err != nil {
		return "", 0, 0, "", err
	}
	var data VnstatOutput
	err = json.Unmarshal(output, &data)
	if err != nil || len(data.Interfaces) == 0 {
		return "", 0, 0, "", fmt.Errorf("invalid vnstat output")
	}
	iface := data.Interfaces[0]
	month := iface.Traffic.Month[0]
	date := fmt.Sprintf("%s %d", time.Month(month.Date.Month).String(), month.Date.Year)
	return iface.Name, month.RX, month.TX, date, nil
}

func sendTelegramMessage(token, chatID, message string) error {
	url := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", token)
	payload := map[string]string{
		"chat_id": chatID,
		"text":    message,
	}
	body, _ := json.Marshal(payload)
	resp, err := http.Post(url, "application/json", bytes.NewBuffer(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, err = ioutil.ReadAll(resp.Body)
	return err
}

func countryFlagEmoji(code string) string {
	if len(code) != 2 {
		return ""
	}
	r := []rune(strings.ToUpper(code))
	offset := rune(127397)
	return string(r[0]+offset) + string(r[1]+offset)
}

func getUsageEmoji(percent float64) string {
	switch {
	case percent < 50:
		return "🟢"
	case percent < 80:
		return "🟡"
	default:
		return "🔴"
	}
}

func getProgressBar(percent float64, width int) string {
	filled := int(percent / 100 * float64(width))
	if filled > width {
		filled = width
	}
	return fmt.Sprintf("[%s%s]", strings.Repeat("█", filled), strings.Repeat("░", width-filled))
}

func main() {
	err := godotenv.Load()
	if err != nil {
		log.Println("Error loading .env file")
	}

	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	chatID := os.Getenv("TELEGRAM_CHAT_ID")
	interfaceName := os.Getenv("INTERFACE")
	if interfaceName == "" {
		interfaceName = "eth0"
	}
	limitStr := os.Getenv("LIMIT_GIB")
	limitGiB := float64(1024) // Default is 1 TiB = 1024 GiB
	if limitStr != "" {
		if v, err := strconv.ParseFloat(limitStr, 64); err == nil {
			limitGiB = v
		}
	}
	if token == "" || chatID == "" {
		fmt.Println("Missing TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID")
		return
	}

	iface, rxBytes, txBytes, date, err := getVnstatData(interfaceName)
	if err != nil {
		fmt.Println("vnstat error:", err)
		return
	}

	txGiB := float64(txBytes) / (1024 * 1024 * 1024)
	percentUsed := (txGiB / limitGiB) * 100

	ipinfo, err := getIPInfo()
	if err != nil {
		fmt.Println("IP info error:", err)
		return
	}
	flag := countryFlagEmoji(strings.ToUpper(ipinfo.Country))
	emoji := getUsageEmoji(percentUsed)
	bar := getProgressBar(percentUsed, 20)

	msg := fmt.Sprintf(`📊 VNSTAT  
Usage on %s in %s:

⬇️ RX: %s  
⬆️ TX: %s (limit: %.0f GiB)  
Total: %s  

TX Limit: %s %.2f%% used  
%s  
🌐 Public IP: %s %s
📍 Location: %s, %s
🏢 ISP: %s`,
		iface, date,
		formatBytes(rxBytes),
		formatBytes(txBytes), limitGiB,
		formatBytes(rxBytes+txBytes),
		emoji, percentUsed,
		bar,
		ipinfo.IP, flag,
		ipinfo.City, ipinfo.Region,
		ipinfo.Org)

	err = sendTelegramMessage(token, chatID, msg)
	if err != nil {
		fmt.Println("Telegram error:", err)
	}
}
