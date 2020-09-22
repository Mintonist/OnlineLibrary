package main

import (
	"log"
	"os"
	"path/filepath"
	"time"

	"github.com/kvark128/av3715/internal/config"
	"github.com/kvark128/av3715/internal/events"
	"github.com/kvark128/av3715/internal/gui"
	"github.com/kvark128/av3715/internal/manager"
	daisy "github.com/kvark128/daisyonline"
)

// Supported mime types of content
const (
	MP3_FORMAT = "audio/mpeg"
	LKF_FORMAT = "audio/x-lkf"
	LGK_FORMAT = "application/lgk"
)

const ProgramVersion = "2020.09.15"

// General client configuration of DAISY-online
var readingSystemAttributes = daisy.ReadingSystemAttributes{
	Manufacturer: "Kvark <kvark128@yandex.ru>",
	Model:        "OnlineLibrary",
	Version:      ProgramVersion,
	Config: daisy.Config{
		SupportsMultipleSelections:        false,
		PreferredUILanguage:               "ru-RU",
		SupportedContentFormats:           daisy.SupportedContentFormats{},
		SupportedContentProtectionFormats: daisy.SupportedContentProtectionFormats{},
		SupportedMimeTypes:                daisy.SupportedMimeTypes{MimeType: []daisy.MimeType{daisy.MimeType{Type: LKF_FORMAT}, daisy.MimeType{Type: LGK_FORMAT}, daisy.MimeType{Type: MP3_FORMAT}}},
		SupportedInputTypes:               daisy.SupportedInputTypes{Input: []daisy.Input{daisy.Input{Type: daisy.TEXT_ALPHANUMERIC}, daisy.Input{Type: daisy.AUDIO}}},
		RequiresAudioLabels:               false,
	},
}

func main() {
	appData := filepath.Join(os.Getenv("APPDATA"), "AV3715")
	config.Initialize(appData)

	os.MkdirAll(config.Conf.UserData, os.ModeDir)
	if fl, err := os.Create(filepath.Join(config.Conf.UserData, "av3715.log")); err == nil {
		log.SetOutput(fl)
		defer fl.Close()
	}
	log.SetPrefix("\n")
	log.SetFlags(log.Lmsgprefix | log.Ltime | log.Lshortfile)

	log.Printf("Starting av3715 version %s", ProgramVersion)
	eventCH := make(chan events.Event, 16)

	if err := gui.Initialize(eventCH); err != nil {
		log.Fatal(err)
	}

	mng := manager.NewManager(daisy.NewClient(config.Conf.ServiceURL, time.Second*3), &readingSystemAttributes)
	go mng.Start(eventCH)

	eventCH <- events.LIBRARY_LOGON
	gui.RunMainWindow()

	config.Save()
	eventCH <- events.PLAYER_STOP
	eventCH <- events.LIBRARY_LOGOFF
	close(eventCH)

	mng.Wait()
	log.Printf("Exiting")
}