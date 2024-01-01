package main

import (
	"encoding/json"
	"html"
	"os"
	"os/signal"
	"regexp"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/dgraph-io/ristretto"
	"github.com/valyala/fasthttp"
)

const Browser = "Chrome"
const UserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/121.0.0.0 Safari/537.36 Edg/121.0.0.0"

const GiftAPI = "https://discordapp.com/api/v8/entitlements/gift-codes/"

var (
	GIT_COMMIT  string // from build flags
	config      Config
	reGiftLink  = regexp.MustCompile("(discord.com/gifts/|discordapp.com/gifts/|discord.gift/)([a-zA-Z0-9]+)")
	reNitroType = regexp.MustCompile(` "name": "([ a-zA-Z]+)", "features"`)
	cache, _    = ristretto.NewCache(&ristretto.Config{
		NumCounters: 1e7,
		MaxCost:     1 << 30,
		BufferItems: 64,
	})
)

func main() {
	withTimeLog("Git Commit:" + GIT_COMMIT)

	if newConfig, err := ReadConfig(); err != nil {
		withTimeFail("[x] Error reading config file," + err.Error())
		return
	} else {
		config = newConfig
	}

	if config.Token == "" {
		withTimeFail("[x] Token is empty.")
		return
	}

	dg, _ := discordgo.New(config.Token)
	dg.UserAgent = UserAgent
	dg.Identify.Properties.OS = "windows"
	dg.Identify.Properties.Browser = Browser

	if err := dg.Open(); err != nil {
		withTimeFail("[x] Error opening connection," + err.Error())
		// if strings.HasPrefix(err.Error(), "websocket: close 4004") {
		// 	withTimeFail("Invalid token provided.")
		// 	panic(err)
		// }
		return
	}

	dg.UpdateStatusComplex(discordgo.UpdateStatusData{
		Status: "dnd",
		AFK:    false,
	})
	dg.AddHandler(messageCreate)

	stop := make(chan os.Signal, 1)
	signal.Notify(stop, syscall.SIGINT, syscall.SIGTERM, os.Interrupt)
	<-stop
}

func printMessageLog(s *discordgo.Session, m *discordgo.MessageCreate) {
	guildName := ""

	// in dm
	if guildId := m.GuildID; guildId == "" {
		return
	} else if guild, _ := s.State.Guild(guildId); guild != nil {
		guildName = guild.Name
	}

	msg := "<cyan>[" + html.EscapeString(guildName) + "]</>\t "
	msg += "<yellow><" + html.EscapeString(m.Author.Username) + "></>\t "
	msg += "<gray>" + html.EscapeString(strings.ReplaceAll(m.Content, "\n", "\n\t\t")) + "</>"
	withTimeLog(msg)
}

func messageCreate(s *discordgo.Session, m *discordgo.MessageCreate) {
	guildID := m.GuildID

	// in dm
	if !config.DMEnable && guildID == "" {
		return
	}
	if contains[string]([]string{}, guildID) {
		return
	}

	// Already own Nitro
	if s.State.User.PremiumType != 0 {
		if !m.Author.Bot {
			printMessageLog(s, m)
		}
		return
	}

	if reGiftLink.Match([]byte(m.Content)) {
		start := time.Now()
		codes := reGiftLink.FindStringSubmatch(m.Content)
		if len(codes) < 2 {
			return
		}

		code := codes[2]
		if len(code) < 16 {
			withTimeLog("<red>[=] Invalid code obtained: " + code + "</>")
			return
		}
		if _, found := cache.Get(code); found {
			withTimeLog("<red>[=] Duplicate code obtained: " + code + "</>")
			return
		}

		req := fasthttp.AcquireRequest()
		req.Header.SetContentType("application/json")
		req.Header.Set("Authorization", GiftAPI+code+"/redeem")
		req.Header.Set("User-Agent", UserAgent)
		req.Header.Set("Accept-Language", "en-US")
		req.Header.Set("Accept-Encoding", "gzip, deflate, br")

		paymentSourceID := "1"
		req.SetBody([]byte(`{"channel_id":` + m.ChannelID + `,"payment_source_id": ` + paymentSourceID + `}`))
		req.Header.SetMethod(fasthttp.MethodPost)
		req.SetRequestURI(GiftAPI)

		res := fasthttp.AcquireResponse()
		if err := fasthttp.Do(req, res); err != nil {
			return
		}
		diff := int64(time.Since(start) / time.Millisecond)
		fasthttp.ReleaseRequest(req)

		body := res.Body()
		bodyString := string(body)
		fasthttp.ReleaseResponse(res)

		printMessageLog(s, m)

		response := Response{}
		if err := json.Unmarshal([]byte(bodyString), &response); err != nil {
			withTimeLog("<red>[=] Invalid response obtained: " + bodyString + "</>")
			return
		}

		withTimeLog("<yellow>[-] " + response.Message + "</> Delay: " + strconv.FormatInt(diff, 10) + "ms")
		if strings.Contains(bodyString, "nitro") {
			nitroType := ""
			if reNitroType.Match([]byte(bodyString)) {
				nitroType = reNitroType.FindStringSubmatch(bodyString)[1]
			}
			withTimeLog("<green>[+] Nitro applied : </><cyan>" + nitroType + "</>")
		}

		cache.Set(code, true, 1)
	}
}
