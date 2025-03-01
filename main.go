package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"io"
	"log"
	"math"
	"math/rand"
	"net/http"
	"os"
	"runtime"

	eb "github.com/hajimehoshi/ebiten/v2"
	eba "github.com/hajimehoshi/ebiten/v2/audio"
	ebw "github.com/hajimehoshi/ebiten/v2/audio/wav"
	ebu "github.com/hajimehoshi/ebiten/v2/ebitenutil"
	ebt "github.com/hajimehoshi/ebiten/v2/text/v2"
)

type Vec2 struct { X, Y float64 }

type Sprite struct {
	Image *eb.Image
	Pos   Vec2
}

type Player struct {
	Sprite
	Vel      Vec2
	Speed    float64
	MoveStep float64
	StopStep float64
	Gravity  float64
	JumpAcc  float64
	FallAcc  float64
	Jump     float64
	CanJump  bool
	Flipped  bool
}

type Platform struct {
	Pos  Vec2
	Size float64
	Num  uint64
}

type Input struct {
	Left, Right, Jump eb.Key
}

type Game struct {
	PosMin Vec2
	PosMax Vec2
	Delta  float64
	Input  Input
	Plats  [9]Platform
	Player Player

	UpSpeed, UpOffset float64
	Score, WaitTimer  float64
	State, ColOffset  uint8
	LastNum uint64

	Scores    []float64
	HighScore float64

	Wall, Plat, PlatL, PlatR *eb.Image
	Font, FontBig            *ebt.GoTextFace

	AudioContext          *eba.Context
	Hit, Jump, Land, Lose *eba.Player
}

func openFile(name string) []byte {
	if runtime.GOOS == "js" {
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
	d, err := os.ReadFile(name)
	if err != nil {
		log.Fatal(err)
	}
	return d
}

func createPlatforms() [9]Platform {
	plats := [9]Platform{}
	for i := 0; i < 8; i++ {
		pPosX := rand.Intn(28)
		pSize := 1 + rand.Intn(28 - pPosX)
		plats[8 - i] = Platform{Vec2{float64(pPosX * 8 + 16), float64(i * 32 - 32)}, float64(pSize), uint64(8 - i)}
	}
	plats[0] = Platform{Vec2{16, 224}, 28, 0}

	return plats
}

func (g *Game) ResetState() {
	g.Scores = append(g.Scores, g.Score)
	if g.Score > g.HighScore {
		g.HighScore = g.Score
	}

	g.State     = 0
	g.ColOffset = 0
	g.UpSpeed   = 0
	g.UpOffset  = 0
	g.Score     = 0
	g.WaitTimer = 0
	g.LastNum   = 0

	g.Player.Pos = Vec2{128, 224}
	g.Plats = createPlatforms()
}

func (g *Game) UpdateMenu() {
	g.WaitTimer += g.Delta
	if (g.WaitTimer < 1) {
		return
	}

	if eb.IsKeyPressed(eb.KeyLeft) || eb.IsKeyPressed(eb.KeyRight) || eb.IsKeyPressed(eb.KeyUp) {
		g.State = 1
	}
}

func (g *Game) UpdatePlay() {
	inputX, lerpStep := 0.0, g.Player.StopStep * g.Delta
	if eb.IsKeyPressed(g.Input.Left) {
		if g.Player.Pos.X >= g.PosMax.X {
			g.Player.Vel.X = -math.Abs(g.Player.Vel.X) * 0.8

			g.Hit.Rewind()
			g.Hit.Play()
		}
		inputX -= 1
		lerpStep = g.Player.MoveStep * g.Delta
		g.Player.Flipped = true
	}
	if eb.IsKeyPressed(g.Input.Right) {
		if g.Player.Pos.X <= g.PosMin.X {
			g.Player.Vel.X = math.Abs(g.Player.Vel.X) * 0.8

			g.Hit.Rewind()
			g.Hit.Play()
		}
		inputX += 1
		lerpStep = g.Player.MoveStep * g.Delta
		g.Player.Flipped = false
	}
	g.Player.Vel.X = g.Player.Vel.X * (1 - lerpStep) + inputX * g.Player.Speed * lerpStep

	gravityStep := g.Player.FallAcc * g.Delta
	if g.Player.Vel.Y < 0 {
		gravityStep = g.Player.JumpAcc * g.Delta
	}
	g.Player.Vel.Y = math.Min(g.Player.Vel.Y + gravityStep, g.Player.Gravity)

	if g.Player.Vel.Y > 0 {
		foundPlat := false
		for i := 0; i < len(g.Plats); i++ {
			if (g.Player.Pos.Y < g.Plats[i].Pos.Y - 2) || (g.Player.Pos.Y - g.Plats[i].Pos.Y > 2) {
				continue
			}
			if math.Abs(g.Player.Pos.X - (g.Plats[i].Pos.X + g.Plats[i].Size * 4)) > (g.Plats[i].Size * 4 + 6) {
				if math.Abs(g.Player.Pos.X - (g.Plats[i].Pos.X + g.Plats[i].Size * 4)) < (g.Plats[i].Size * 4 + 9) {
					g.Player.CanJump = true
				}
				continue
			}

			if g.Plats[i].Num > g.LastNum {
				scoreAdd := float64(g.Plats[i].Num - g.LastNum)
				g.Score += math.Pow(scoreAdd, 1 + (scoreAdd) / 5) * 10
				g.LastNum = g.Plats[i].Num
			}

			if g.Player.CanJump == false {
				g.Land.Rewind()
				g.Land.Play()
			}

			g.Player.Pos.Y = g.Plats[i].Pos.Y
			g.Player.Vel.Y = 0
			g.Player.CanJump = true
			foundPlat = true
			break
		}
		if foundPlat == false {
			g.Player.CanJump = false
		}
	}

	if eb.IsKeyPressed(g.Input.Jump) && g.Player.CanJump {
		g.Player.Vel.Y = -g.Player.Jump * (1 + math.Abs(g.Player.Vel.X) / g.Player.Speed)
		g.Player.CanJump = false

		g.Jump.Rewind()
		g.Jump.Play()
	}

	for i := 0; i < len(g.Plats); i++ {
		if g.Player.Pos.Y <= g.PosMin.Y {
			g.Plats[i].Pos.Y -= g.Player.Vel.Y * g.Delta
		}
		g.Plats[i].Pos.Y += g.UpSpeed * g.Delta
		if (g.Plats[i].Pos.Y < 240) {
			continue
		}
		pPosX := rand.Intn(28)
		g.Plats[i].Pos.X = float64(pPosX * 8 + 16)
		g.Plats[i].Pos.Y -= 288
		g.Plats[i].Size = float64(rand.Intn(28 - pPosX) + 1)
		g.Plats[i].Num += 9

		if (g.Plats[i].Num % 5 == 0) && (g.ColOffset < 60) {
			g.ColOffset += 1
		}

		if g.UpSpeed == 0 {
			g.UpSpeed = 7.9
		}
		g.UpSpeed += 0.1
	}
	g.UpOffset += g.UpSpeed * g.Delta
	if g.Player.Pos.Y <= g.PosMin.Y {
		g.UpOffset -= g.Player.Vel.Y * g.Delta
	}

	g.Player.Pos.X += g.Player.Vel.X * g.Delta
	g.Player.Pos.Y += g.Player.Vel.Y * g.Delta
	g.Player.Pos.X = math.Max(g.PosMin.X, math.Min(g.PosMax.X, g.Player.Pos.X))
	g.Player.Pos.Y = math.Max(g.PosMin.Y, math.Min(g.PosMax.Y, g.Player.Pos.Y))

	if g.Player.Pos.Y != g.PosMax.Y {
		return
	}
	g.Lose.Rewind()
	g.Lose.Play()
	g.ResetState()
}

func (g *Game) Update() error {
	switch g.State {
	case 0:
		g.UpdateMenu()
	case 1:
		g.UpdatePlay()
	}

	return nil
}

func (g *Game) DrawMenu(screen *eb.Image) {
	topts := &ebt.DrawOptions{}
	topts.PrimaryAlign = ebt.AlignCenter

	if g.HighScore == 0 {
		topts.GeoM.Translate(128, 176)
		ebt.Draw(screen, "Press arrow keys to play", g.Font, topts)
		topts.GeoM.Translate(0, -100)
		ebt.Draw(screen, "JUMP TOWER", g.FontBig, topts)
	} else {
		if g.WaitTimer >= 1 {
			topts.GeoM.Translate(128, 176)
			ebt.Draw(screen, "Press arrow keys to play", g.Font, topts)
		}
		topts.GeoM.Reset()
		topts.GeoM.Translate(128, 116)
		ebt.Draw(screen, fmt.Sprintf("High score: %.0f", g.HighScore), g.Font, topts)
		topts.GeoM.Translate(0, -10)
		ebt.Draw(screen, fmt.Sprintf("Your score: %.0f", g.Scores[len(g.Scores) - 1]), g.Font, topts)
		topts.GeoM.Translate(0, -30)
		ebt.Draw(screen, "YOU LOSE!", g.FontBig, topts)
	}
}

func (g *Game) DrawPlay(screen *eb.Image) {
	topts := &ebt.DrawOptions{}
	topts.GeoM.Translate(8, 4)
	ebt.Draw(screen, fmt.Sprintf("Score: %.0f", g.Score), g.Font, topts)
}

func (g *Game) Draw(screen *eb.Image) {
	screen.Fill(color.RGBA{60 + g.ColOffset, 60, 120 - g.ColOffset, 255})
	opts := eb.DrawImageOptions{}

	opts.GeoM.Translate(0, float64(int(g.UpOffset + 0.5) % 16 - 16))
	for i := 0; i < 16; i++ {
		screen.DrawImage(g.Wall, &opts)
		opts.GeoM.Translate(240, 0)
		screen.DrawImage(g.Wall, &opts)
		opts.GeoM.Translate(-240, 16)
	}

	for i := 0; i < len(g.Plats); i++ {
		opts.GeoM.Reset()
		opts.GeoM.Translate(g.Plats[i].Pos.X, g.Plats[i].Pos.Y)
		screen.DrawImage(g.PlatL, &opts)
		opts.GeoM.Translate(4, 0)
		for j := 0; j < int(g.Plats[i].Size) - 1; j++ {
			screen.DrawImage(g.Plat, &opts)
			opts.GeoM.Translate(8, 0)
		}
		screen.DrawImage(g.PlatR, &opts)
	}

	opts.GeoM.Reset()
	opts.GeoM.Translate(-float64(g.Player.Image.Bounds().Dx() / 2), -float64(g.Player.Image.Bounds().Dy()))
	if (g.Player.Flipped) {
		opts.GeoM.Scale(-1, 1)
	}
	opts.GeoM.Translate(g.Player.Pos.X, g.Player.Pos.Y)
	screen.DrawImage(g.Player.Image, &opts)

	// debugInfo := fmt.Sprintf(" TPS: %.2f\n FPS: %.2f", eb.ActualTPS(), eb.ActualFPS())
	// ebu.DebugPrint(screen, debugInfo)

	switch g.State {
	case 0:
		g.DrawMenu(screen)
	case 1:
		g.DrawPlay(screen)
	}
}

func (g *Game) Layout(ww, wh int) (int, int) {
	return 256, 240
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

	fontBytes := openFile("asset/font.ttf")
	fontSource, err := ebt.NewGoTextFaceSource(bytes.NewReader(fontBytes))
	if err != nil {
		log.Fatal(err)
	}

	game := &Game{
		PosMin: Vec2{23, 48},
		PosMax: Vec2{233, 272},
		Delta:  1.0 / float64(updateTPS),
		Input:  Input{eb.KeyLeft, eb.KeyRight, eb.KeyUp},
		Plats:  createPlatforms(),
		Player: Player{
			Sprite:   Sprite{playerImage, Vec2{128, 224}},
			Speed:    280,
			MoveStep: 3,
			StopStep: 10,
			Gravity:  400,
			JumpAcc:  650,
			FallAcc:  1200,
			Jump:     240,
		},
		Wall:  wallImage,
		Plat:  platImage,
		PlatL: platLImage,
		PlatR: platRImage,

		WaitTimer: 1,

		Font:    &ebt.GoTextFace{Source: fontSource, Size: 8},
		FontBig: &ebt.GoTextFace{Source: fontSource, Size: 16},

		AudioContext: eba.NewContext(44100),
	}

	audioSources := [4]string{"asset/hit.wav", "asset/jump.wav", "asset/land.wav", "asset/lose.wav"}
	audioPlayers := [4]*eba.Player{}

	for i := 0; i < 4; i++ {
		data := openFile(audioSources[i])
		dec, err := ebw.DecodeWithSampleRate(44100, bytes.NewReader(data))
		if err != nil {
			log.Fatal(err)
		}
		audioPlayers[i], err = game.AudioContext.NewPlayer(dec)
		if err != nil {
			log.Fatal(err)
		}
	}
	game.Hit, game.Jump, game.Land, game.Lose = audioPlayers[0], audioPlayers[1], audioPlayers[2], audioPlayers[3]


	if err := eb.RunGame(game); err != nil {
		log.Fatal(err)
	}
}
