package notify

import (
	"bytes"
	"encoding/json"
	"log"
	"net/http"

	"wheather-go/internal/weather"

	"github.com/bwmarrin/discordgo"
)

type Notifier struct {
	webhookURL string
	session    *discordgo.Session
	channelID  string
}

// NewNotifier creates a Notifier that sends to the given webhook URL and bot channel.
// Pass an empty webhookURL or channelID to disable that delivery path.
func NewNotifier(webhookURL, channelID string, session *discordgo.Session) *Notifier {
	return &Notifier{
		webhookURL: webhookURL,
		channelID:  channelID,
		session:    session,
	}
}

func (n *Notifier) Send(p Payload) error {
	if n.webhookURL == "" {
		log.Println("Discord webhook URL not set, skipping notification")
		return nil
	}
	body, err := json.Marshal(p)
	if err != nil {
		return err
	}
	resp, err := http.Post(n.webhookURL, "application/json", bytes.NewReader(body))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		log.Printf("Discord webhook returned status %d", resp.StatusCode)
	}
	return nil
}

func (n *Notifier) SendToChannel(p Payload) error {
	return n.SendToChannelID(n.channelID, p)
}

func (n *Notifier) SendToChannelID(channelID string, p Payload) error {
	// This path is used for request/response bot behavior where the destination channel is decided at runtime.
	if n.session == nil || channelID == "" {
		return nil
	}
	embeds := make([]*discordgo.MessageEmbed, 0, len(p.Embeds))
	for _, e := range p.Embeds {
		embeds = append(embeds, toDiscordgoEmbed(e))
	}
	_, err := n.session.ChannelMessageSendComplex(channelID, &discordgo.MessageSend{
		Content: p.Content,
		Embeds:  embeds,
	})
	return err
}

func (n *Notifier) broadcast(p Payload) {
	if err := n.Send(p); err != nil {
		log.Printf("Discord webhook error: %v", err)
	}
	if err := n.SendToChannel(p); err != nil {
		log.Printf("Bot channel send error: %v", err)
	}
}

func (n *Notifier) UrgentWeather(report *weather.WeatherReport) {
	n.broadcast(UrgentWeatherPayload(report))
	log.Println("Urgent weather alert sent to Discord")
}

func (n *Notifier) UrgentAQI(aqi *weather.AQIResponse) {
	n.broadcast(UrgentAQIPayload(aqi))
	log.Println("Urgent AQI alert sent to Discord")
}

func (n *Notifier) AllClear() {
	n.broadcast(AllClearPayload())
	log.Println("All Clear sent to Discord")
}

func (n *Notifier) PeriodicReport(report *weather.WeatherReport, risk string) {
	n.broadcast(PeriodicReportPayload(report, risk))
	log.Println("Periodic report sent to Discord")
}

func (n *Notifier) AQIReport(aqi *weather.AQIResponse) {
	n.broadcast(AQIReportPayload(aqi))
	log.Println("AQI report sent to Discord")
}

// Webhook-only helpers keep scheduled jobs from also echoing into bot channels.
func (n *Notifier) PeriodicReportWebhookOnly(report *weather.WeatherReport, risk string) {
	if err := n.Send(PeriodicReportPayload(report, risk)); err != nil {
		log.Printf("Discord webhook error: %v", err)
		return
	}
	log.Println("Periodic report sent to Discord webhook")
}

func (n *Notifier) AQIReportWebhookOnly(aqi *weather.AQIResponse) {
	if err := n.Send(AQIReportPayload(aqi)); err != nil {
		log.Printf("Discord webhook error: %v", err)
		return
	}
	log.Println("AQI report sent to Discord webhook")
}

// Channel-only helpers keep user-issued commands inside the same Discord conversation.
func (n *Notifier) PeriodicReportToChannel(channelID string, report *weather.WeatherReport, risk string) {
	if err := n.SendToChannelID(channelID, PeriodicReportPayload(report, risk)); err != nil {
		log.Printf("Bot channel send error: %v", err)
		return
	}
	log.Println("Periodic report sent to bot channel")
}

func (n *Notifier) AQIReportToChannel(channelID string, aqi *weather.AQIResponse) {
	if err := n.SendToChannelID(channelID, AQIReportPayload(aqi)); err != nil {
		log.Printf("Bot channel send error: %v", err)
		return
	}
	log.Println("AQI report sent to bot channel")
}

func toDiscordgoEmbed(e Embed) *discordgo.MessageEmbed {
	embed := &discordgo.MessageEmbed{
		Title:       e.Title,
		Description: e.Description,
		Color:       e.Color,
		Timestamp:   e.Timestamp,
	}
	for _, f := range e.Fields {
		embed.Fields = append(embed.Fields, &discordgo.MessageEmbedField{
			Name:   f.Name,
			Value:  f.Value,
			Inline: f.Inline,
		})
	}
	if e.Footer != nil {
		embed.Footer = &discordgo.MessageEmbedFooter{Text: e.Footer.Text}
	}
	return embed
}
