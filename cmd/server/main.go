package main

import (
	"log"
	"net/http"

	"wheather-go/internal/config"
	"wheather-go/internal/notify"
	"wheather-go/internal/store"

	"github.com/bwmarrin/discordgo"
)

type App struct {
	cfg         config.Config
	store       *store.Store
	notifier    *notify.Notifier
	cnxNotifier *notify.Notifier
}

func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatal("Config error:", err)
	}

	st, err := store.New(cfg.DBPath)
	if err != nil {
		log.Fatal("Database init error:", err)
	}

	sess, err := discordgo.New("Bot " + cfg.WeatherBotKey)
	if err != nil {
		log.Fatal("Discord init error:", err)
	}

	app := &App{cfg: cfg, store: st}

	maesaiWebhook := cfg.WebhookMaesaiURL
	if cfg.Debug {
		maesaiWebhook = cfg.WebhookTestURL
	}
	app.notifier = notify.NewNotifier(maesaiWebhook, cfg.MaesaiChannel, sess)

	cnxWebhook := cfg.WebhookCNXURL
	if cfg.Debug {
		cnxWebhook = cfg.WebhookTestURL
	}
	app.cnxNotifier = notify.NewNotifier(cnxWebhook, cfg.CNXChannel, sess)

	sess.AddHandler(app.onMessageCreate)
	sess.Identify.Intents = discordgo.IntentsAllWithoutPrivileged | discordgo.IntentMessageContent

	if err = sess.Open(); err != nil {
		log.Fatal("Discord open error:", err)
	}
	defer sess.Close()

	app.startCron()
	app.registerRoutes()

	log.Printf("Server running on :%s", cfg.Port)
	if err := http.ListenAndServe(":"+cfg.Port, nil); err != nil {
		log.Fatal(err)
	}
}
