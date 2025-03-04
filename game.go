package main

import (
	"fmt"
	"image/color"

	eb  "github.com/hajimehoshi/ebiten/v2"
	eba "github.com/hajimehoshi/ebiten/v2/audio"
	ebt "github.com/hajimehoshi/ebiten/v2/text/v2"
)

type Game struct {
	Delta  float64
	Plats  [9]Platform
	Player Player

	UpSpeed, UpOffset float64
	Score, WaitTimer  float64
	State, ColOffset  uint8
	LastNum uint64

	Scores    []float64
	HighScore float64

	Wall, Plat, PlatL, PlatR *eb.Image
	Font, FontBig *ebt.GoTextFace

	AudioContext *eba.Context
	Sound map[string]*eba.Player
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

	g.Player.Pos = Vec2{X: 128, Y: 224}
	g.Player.Vel = Vec2{X: 0, Y: 0}
	g.Plats = CreatePlatforms()
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
	g.Player.UpdateInput(eb.IsKeyPressed(eb.KeyLeft), eb.IsKeyPressed(eb.KeyRight), g.Delta, g.Sound)
	g.Player.UpdateGravity(g.Delta)
	newNum, score := g.Player.UpdateCollision(eb.IsKeyPressed(eb.KeyUp), &(g.Plats), g.LastNum, g.Delta, g.Sound)
	if score != -1 {
		g.LastNum = newNum
		g.Score  += score
	}
	g.Player.UpdatePos(g.Delta)

	for i := 0; i < len(g.Plats); i++ {
		up := g.Plats[i].Update(g.Player.Pos.Y <= PosMinY, g.Player.Vel.Y, g.UpSpeed, g.Delta)
		if up == 0 {
			continue
		}
		g.UpSpeed += up
		if (g.ColOffset < 60) && (int(g.UpSpeed * 10) % 5 == 0) {
			g.ColOffset++
		}
	}
	g.UpOffset += g.UpSpeed * g.Delta
	if g.Player.Pos.Y <= PosMinY {
		g.UpOffset -= g.Player.Vel.Y * g.Delta
	}

	if g.Player.Pos.Y == PosMaxY {
		g.ResetState()

		g.Sound["lose"].Rewind()
		g.Sound["lose"].Play()
	}
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

	if g.WaitTimer >= 0.5 {
		topts.GeoM.Translate(128, 176)
		ebt.Draw(screen, "Press arrow keys to play", g.Font, topts)
	}

	if g.HighScore == 0 {
		topts.GeoM.Reset()
		topts.GeoM.Translate(128, 76)
		ebt.Draw(screen, "JUMP TOWER", g.FontBig, topts)
		return
	}

	topts.GeoM.Reset()
	topts.GeoM.Translate(128, 116)
	ebt.Draw(screen, fmt.Sprintf("High score: %.0f", g.HighScore), g.Font, topts)
	topts.GeoM.Translate(0, -10)
	ebt.Draw(screen, fmt.Sprintf("Your score: %.0f", g.Scores[len(g.Scores) - 1]), g.Font, topts)
	topts.GeoM.Translate(0, -30)
	ebt.Draw(screen, "YOU LOSE!", g.FontBig, topts)
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
