package main

import (
	"math/rand"
)

type Platform struct {
	Pos  Vec2
	Size float64
	Num  uint64
}

func (p *Platform) Update(goingUp bool, upAmount, upSpeed, delta float64) float64 {
	if goingUp {
		p.Pos.Y -= upAmount * delta
	}
	p.Pos.Y += upSpeed * delta
	if (p.Pos.Y < 240) {
		return 0
	}

	pPosX := rand.Intn(28)
	p.Pos  = Vec2{float64(pPosX * 8 + 16), p.Pos.Y - 288}
	p.Size = float64(rand.Intn(28 - pPosX) + 1)
	p.Num += 9

	if upSpeed == 0 {
		return 8
	}
	return 0.1
}

func CreatePlatforms() [9]Platform {
	plats := [9]Platform{}
	for i := uint64(0); i < 8; i++ {
		platVal := rand.Intn(28)
		platPosX := float64(platVal * 8 + 16)
		platPosY := float64(i * 32) - 32
		platSize := float64(rand.Intn(28 - platVal) + 1)
		plats[8 - i] = Platform{Vec2{platPosX, platPosY}, platSize, 8 - i}
	}
	plats[0] = Platform{Vec2{16, 224}, 28, 0}

	return plats
}
