package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/joho/godotenv"
)

var (
  userName string
  token string
  channelID string
) 

func Init() {
  if err := godotenv.Load(); err != nil {
    log.Fatalln("Error loading .env file")
  }
  userName = os.Getenv("USERNAME")
  token = os.Getenv("BOTTOKEN")
  channelID = os.Getenv("CHANNELID")
}

func main() {
  // debug
//   apiDefaultURL := "https://spla3.yuu26.com/api/regular/now"
//   if err := request2api(apiDefaultURL); err != nil {
//     log.Fatalln(err)
//   }
//   os.Exit(1)

  Init()

  discord, err := discordgo.New("Bot " + token)
  if err != nil {
    log.Fatalln(err)
  }

  if err = discord.Open(); err != nil {
    log.Fatalln(err)
  }

  NotifySpla3Match(discord)

  stop := make(chan os.Signal, 1)
  signal.Notify(stop, syscall.SIGINT, os.Interrupt, os.Kill)
  ticket := time.NewTicker(2 * time.Hour)

  for {
    select {
      case <-ticket.C:
        NotifySpla3Match(discord)
      case <-stop:
        goto GOTO_FINISH
    }
  }

  GOTO_FINISH:
  if err = discord.Close(); err != nil {
    log.Fatalln(err)
  }
  fmt.Println("Bot Down!")
}

func NotifySpla3Match(s *discordgo.Session) {
  var builder strings.Builder

  urls := []string{
    "https://spla3.yuu26.com/api/regular/now",
    "https://spla3.yuu26.com/api/regular/next",
    "https://spla3.yuu26.com/api/bankara-open/now",
    "https://spla3.yuu26.com/api/bankara-open/next",
    "https://spla3.yuu26.com/api/bankara-challenge/now",
    "https://spla3.yuu26.com/api/bankara-challenge/next",
  }

  for i, url := range urls {
    data, err := request2api(url)
    if err != nil {
      log.Println(err)
    }

    switch i / 2 {
      case 0:
        builder.WriteString("__=== regular ===__")
      case 1:
        builder.WriteString("__=== bankara-open ===__")
      case 2:
        builder.WriteString("__=== bankara-challenge ===__")
    }

    switch i % 2 {
    case 0:
      builder.WriteString(" : Now \n")
    case 1:
      builder.WriteString(" : Next \n")
    }

    builder.WriteString(data.Results[0].Rule.Name + "\n")

    for _, stage := range data.Results[0].Stages {
      builder.WriteString(" - " + stage.Name + "\n")
    }

    builder.WriteByte('\n')
  }

  if err := sendMessage(s, channelID, builder.String()); err != nil {
    log.Println(err)
  }
}

func sendMessage(s *discordgo.Session, channelID string, msg string) error {
  _, err := s.ChannelMessageSend(channelID, msg)
  log.Println(" > " + msg)
  if err != nil {
    return err
  }
  return nil
}

type data struct {
	Results []struct {
		StartTime time.Time `json:"start_time"`
		EndTime   time.Time `json:"end_time"`
		Rule      struct {
			Key  string `json:"key"`
			Name string `json:"name"`
		} `json:"rule"`
		Stages []struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"stages"`
	} `json:"results"`
}

func request2api(url string) (*data, error) {
  results := new(data)
  
  req, err := http.NewRequest("GET", url, nil)
  if err != nil {
    return results, err
  }

  req.Header.Set("User-Agent", userName)

  client := new(http.Client)
  resp, err := client.Do(req)
  if err != nil {
    return results, err
  }
  defer resp.Body.Close()

  if resp.StatusCode != 200 {
    return results, errors.New("HTTP code not OK")
  }

  body, _ := io.ReadAll(resp.Body)
  if err := json.Unmarshal(body, &results); err != nil {
    return results, err
  }

  fmt.Printf("%+v\n", results)

  return results, nil
}

