package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"strings"

	tgbotapi "github.com/go-telegram-bot-api/telegram-bot-api/v5"
)

type bnResp struct {
	Price float64 `json:"price, string"`
	Code  int64   `json:"code"`
}
type wallet map[string]float64

var db = map[int64]wallet{}

func main() {
	bot, err := tgbotapi.NewBotAPI("5729925133:AAEgZdT5-F8XVfz76mZItVKgBJzIkyLMQQ0")
	if err != nil {
		log.Panic(err)
	}

	bot.Debug = true

	log.Printf("Authorized on account %s", bot.Self.UserName)

	u := tgbotapi.NewUpdate(0)
	u.Timeout = 60

	updates := bot.GetUpdatesChan(u)

	for update := range updates {
		if update.Message != nil {

			command := strings.Split(update.Message.Text, " ")

			switch command[0] {
			case "ADD":
				if len(command) == 3 {

					amount, err := strconv.ParseFloat(command[2], 64)
					if err != nil {
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "введен неверный формат валюты"))
					}

					if _, ok := db[update.Message.Chat.ID]; !ok {
						db[update.Message.Chat.ID] = wallet{}
					}

					db[update.Message.Chat.ID][command[1]] += amount
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
						fmt.Sprintf("валюта добавлена + %f", amount)))
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
						fmt.Sprintf("баланс кошелька %f", db[update.Message.Chat.ID][command[1]])))

				} else {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "неверная команда"))
					break
				}

			case "SUB":
				if len(command) == 3 {

					amount, err := strconv.ParseFloat(command[2], 64)
					if err != nil {
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "введен неверный формат валюты"))
					}

					if _, ok := db[update.Message.Chat.ID]; !ok {
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "валюта не найдена"))
						continue
					}

					check := db[update.Message.Chat.ID][command[1]] - amount
					if check >= 0 {
						db[update.Message.Chat.ID][command[1]] -= amount
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
							fmt.Sprintf("валюта списана - %f", amount)))
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID,
							fmt.Sprintf("баланс кошелька %f", db[update.Message.Chat.ID][command[1]])))
					} else {
						bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "нельзя списать валюту, так как баланс отрицательный"))
					}

				} else {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "неверная команда"))
					break
				}
			case "DEL":
				if len(command) == 2 {
					delete(db[update.Message.Chat.ID], command[1])
				} else {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "неверная команда"))
					break
				}

			case "SHOW":
				if len(command) == 1 {
					msg := ""
					var sum float64 = 0

					for key, value := range db[update.Message.Chat.ID] {
						price, _ := getPrice(key)
						sum += value * price
						msg += fmt.Sprintf("%s: %f [%.2f]\n", key, value, value*price)
					}
					msg += fmt.Sprintf("Total: %.2f\n", sum)
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, msg))
				} else {
					bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "неверная команда"))
					break
				}
			default:
				bot.Send(tgbotapi.NewMessage(update.Message.Chat.ID, "команда не найдена"))
			}

		}
	}

}

func getPrice(symbol string) (price float64, err error) {
	resp, err := http.Get(fmt.Sprintf("https://api.binance.com/api/v3/ticker/price?symbol=%sUSDT", symbol))
	if err != nil {
		return
	}
	defer resp.Body.Close()
	var jsonResp bnResp
	err = json.NewDecoder(resp.Body).Decode(&jsonResp)
	if err != nil {
		return
	}

	if jsonResp.Code != 0 {
		err = errors.New("неверный символ")
	}
	price = jsonResp.Price
	return
}
