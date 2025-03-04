package main

import (
	eb "github.com/hajimehoshi/ebiten/v2"
)

type Vec2 struct {
	X, Y float64
}

type Sprite struct {
	Image *eb.Image
	Pos   Vec2
}
