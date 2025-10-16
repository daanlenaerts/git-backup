package main

import (
	"fmt"
	"net/http"
	"net/url"
	"os"
)

func SendTelegramMessage(message string) {
	telegramBotToken := os.Getenv("TELEGRAM_BOT_TOKEN")
	telegramChatID := os.Getenv("TELEGRAM_CHAT_ID")

	if telegramBotToken == "" || telegramChatID == "" {
		fmt.Println("TELEGRAM_BOT_TOKEN or TELEGRAM_CHAT_ID is not set")
		return
	}

	telegramAPIURL := fmt.Sprintf("https://api.telegram.org/bot%s/sendMessage", telegramBotToken)

	data := url.Values{
		"chat_id": {telegramChatID},
		"text":    {"git-backup: " + message},
	}

	resp, err := http.PostForm(telegramAPIURL, data)
	if err != nil {
		fmt.Println("Error sending telegram message:", err)
		return
	}

	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		fmt.Println("Error sending telegram message:", resp.Status)
	}
}
