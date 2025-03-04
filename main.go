package main

import (
	"bytes"
	"image"
	"io"
	"log"
	"net/http"
	"os"
	"runtime"

	eb "github.com/hajimehoshi/ebiten/v2"
	eba "github.com/hajimehoshi/ebiten/v2/audio"
	ebw "github.com/hajimehoshi/ebiten/v2/audio/wav"
	ebu "github.com/hajimehoshi/ebiten/v2/ebitenutil"
	ebt "github.com/hajimehoshi/ebiten/v2/text/v2"
)

func OpenFile(name string) []byte {
	if runtime.GOOS != "js" {
		d, err := os.ReadFile(name)
		if err != nil {
			log.Fatal(err)
		}
		return d
	}
	res, err := http.Get(name)
	if err != nil {
		log.Fatal(err)
	}
	d, err := io.ReadAll(res.Body)
	if err != nil {
		log.Fatal(err)
	}
	return d
}

func main() {
	const updateTPS = 240

	eb.SetWindowSize(512, 480)
	eb.SetWindowTitle("Jump Tower")
	eb.SetWindowResizingMode(eb.WindowResizingModeEnabled)
	eb.SetTPS(updateTPS)
	eb.SetVsyncEnabled(true)

	graphics, _, err := ebu.NewImageFromFile("asset/ingame.png")
	if err != nil {
		log.Fatal(err)
	}
	playerImage := graphics.SubImage(image.Rect(0,  0, 16, 24)).(*eb.Image)
	wallImage   := graphics.SubImage(image.Rect(16, 8, 32, 24)).(*eb.Image)
	platImage   := graphics.SubImage(image.Rect(20, 0, 28, 8)).(*eb.Image)
	platLImage  := graphics.SubImage(image.Rect(16, 0, 20, 8)).(*eb.Image)
	platRImage  := graphics.SubImage(image.Rect(28, 0, 32, 8)).(*eb.Image)

	fontBytes := OpenFile("asset/font.ttf")
	fontSource, err := ebt.NewGoTextFaceSource(bytes.NewReader(fontBytes))
	if err != nil {
		log.Fatal(err)
	}

	game := &Game{
		Delta:  1.0 / float64(updateTPS),
		Plats:  CreatePlatforms(),
		Player: Player{Sprite: Sprite{playerImage, Vec2{128, 224}}},
		Wall:  wallImage,
		Plat:  platImage,
		PlatL: platLImage,
		PlatR: platRImage,

		Font:    &ebt.GoTextFace{Source: fontSource, Size: 8},
		FontBig: &ebt.GoTextFace{Source: fontSource, Size: 16},

		AudioContext: eba.NewContext(44100),
		Sound: map[string]*eba.Player{},
	}

	audioNames   := [4]string{"hit", "jump", "land", "lose"}
	audioPlayers := [4]*eba.Player{}

	for i := 0; i < 4; i++ {
		data := OpenFile("asset/" + audioNames[i] + ".wav")
		dec, err := ebw.DecodeWithSampleRate(44100, bytes.NewReader(data))
		if err != nil {
			log.Fatal(err)
		}
		audioPlayers[i], err = game.AudioContext.NewPlayer(dec)
		if err != nil {
			log.Fatal(err)
		}
		game.Sound[audioNames[i]] = audioPlayers[i]
	}

	if err := eb.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
