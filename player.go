package main

import (
	"math"

	eba "github.com/hajimehoshi/ebiten/v2/audio"
)

const (
	PosMinX, PosMinY = 23.0,  48.0
	PosMaxX, PosMaxY = 233.0, 272.0

	MaxSpeed = 280.0
	Accel    = 700.0
	Friction = 1000.0
	Gravity  = 400.0
	JumpAcc  = 650.0
	FallAcc  = 1200.0
	JumpVel  = 240.0

	RetainForce = 0.8
	WallJumpMul = 0.7
)

type Player struct {
	Sprite
	Vel      Vec2
	CanJump  bool
	Flipped  bool
}

func (p *Player) UpdateInput(pLeft, pRight bool, delta float64, sound map[string]*eba.Player) {
	inputX  := 0.0
	if pLeft {
		if p.Pos.X >= PosMaxX {
			p.Vel.X = -math.Abs(p.Vel.X) * RetainForce
			sound["hit"].Rewind()
			sound["hit"].Play()
		}
		inputX -= 1
		if p.Vel.X > 0 {
			inputX -= 1
		}
		p.Flipped = true
	}
	if pRight {
		if p.Pos.X <= PosMinX {
			p.Vel.X = math.Abs(p.Vel.X) * RetainForce
			sound["hit"].Rewind()
			sound["hit"].Play()
		}
		inputX += 1
		if p.Vel.X < 0 {
			inputX += 1
		}
		p.Flipped = false
	}

	p.Vel.X += inputX * Accel * delta
	if inputX > 0 {
		p.Vel.X = math.Min(MaxSpeed, p.Vel.X)
	} else if inputX < 0 {
		p.Vel.X = math.Max(-MaxSpeed, p.Vel.X)
	} else if p.Vel.X > 0 {
		p.Vel.X -= Friction * delta
		p.Vel.X = math.Max(0, p.Vel.X)
	} else if p.Vel.X < 0 {
		p.Vel.X += Friction * delta
		p.Vel.X = math.Min(0, p.Vel.X)
	}
}

func (p *Player) UpdateGravity(delta float64) {
	if p.Vel.Y < 0 {
		p.Vel.Y = math.Min(p.Vel.Y + JumpAcc * delta, Gravity)
	} else {
		p.Vel.Y = math.Min(p.Vel.Y + FallAcc * delta, Gravity)
	}
}

func (p *Player) UpdateCollision(pJump bool, plats *[9]Platform, lastNum uint64, delta float64, sound map[string]*eba.Player) (uint64, float64) {
	if p.Vel.Y < 0 {
		return 0, -1
	}

	newNum := uint64(0)
	score  := float64(-1)

	foundPlat := false
	for i := 0; i < len(plats) && !foundPlat; i++ {
		if (p.Pos.Y < plats[i].Pos.Y - 2) || (p.Pos.Y > plats[i].Pos.Y + 2) {
			continue
		}
		if math.Abs(p.Pos.X - (plats[i].Pos.X + plats[i].Size * 4)) > (plats[i].Size * 4 + 6) {
			if math.Abs(p.Pos.X - (plats[i].Pos.X + plats[i].Size * 4)) < (plats[i].Size * 4 + 9) {
				p.CanJump = true
			}
			continue
		}
		if plats[i].Num > lastNum {
			scoreAdd := float64(plats[i].Num - lastNum)
			score = math.Pow(scoreAdd, 1 + (scoreAdd) / 5) * 10
			newNum = plats[i].Num
		}
		p.Pos.Y = plats[i].Pos.Y
		p.Vel.Y = 0
		p.CanJump = true
		foundPlat = true

		if p.CanJump == false {
			sound["land"].Rewind()
			sound["land"].Play()
		}
	}

	if foundPlat == false {
		p.CanJump = false
	}
	if pJump && p.CanJump {
		jumpMul := 1.0
		if p.Pos.X <= PosMinX || p.Pos.X >= PosMaxX {
			jumpMul = WallJumpMul
		}
		p.Vel.Y = -JumpVel * (1 + math.Abs(p.Vel.X) / MaxSpeed * jumpMul)
		p.CanJump = false

		sound["jump"].Rewind()
		sound["jump"].Play()
	}

	return newNum, score
}

func (p *Player) UpdatePos(delta float64) {
	p.Pos.X += p.Vel.X * delta
	p.Pos.Y += p.Vel.Y * delta
	p.Pos.X = math.Max(PosMinX, math.Min(PosMaxX, p.Pos.X))
	p.Pos.Y = math.Max(PosMinY, math.Min(PosMaxY, p.Pos.Y))
}
