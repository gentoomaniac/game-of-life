package gameoflife

import (
	"math/rand"

	log "github.com/rs/zerolog/log"

	"github.com/hajimehoshi/ebiten/v2"
)

type Game struct {
	Pause        bool
	Width        int
	Height       int
	Cells        [][]bool
	pixels       []byte
	doubleBuffer []byte
}

func NewGame(width, height int) *Game {
	g := &Game{
		Pause:        false,
		Width:        width,
		Height:       height,
		doubleBuffer: make([]byte, width*height*4),
		pixels:       make([]byte, width*height*4),
	}

	// initial randomness
	cells := make([]bool, width*height)
	for i := 0; i < len(cells); i++ {
		if rand.Intn(2) == 1 {
			cells[i] = true
		}
	}
	g.Cells = append(g.Cells, cells)
	return g
}

// Update handles the logic (Game of Life rules go here)
func (g *Game) Update() error {
	if !g.Pause {
		g.Cells = append(g.Cells, g.updateCells(g.Cells[len(g.Cells)-1]))
	}

	return nil
}

func (g *Game) updateCells(state []bool) []bool {
	newState := make([]bool, len(state))

	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
			// Count alive neighbors
			neighbors := 0

			// Check all 8 surrounding cells
			for dy := -1; dy <= 1; dy++ {
				for dx := -1; dx <= 1; dx++ {
					if dx == 0 && dy == 0 {
						continue
					} // Skip self

					// Wrap coordinates (Toroidal logic)
					nx := (x + dx + g.Width) % g.Width
					ny := (y + dy + g.Height) % g.Height

					// Convert 2D neighbor coord back to 1D index
					idx := ny*g.Width + nx

					if state[idx] {
						neighbors++
					}
				}
			}

			// Calculate current cell index
			currentIdx := y*g.Width + x
			alive := state[currentIdx]

			// Apply Game of Life Rules
			if alive && (neighbors == 2 || neighbors == 3) {
				newState[currentIdx] = true
			} else if !alive && neighbors == 3 {
				newState[currentIdx] = true
			} else {
				newState[currentIdx] = false
			}

			// update the pixel buffer here as well since this
			// only has to be done when something actually changes
			// and not on every draw
			// To avoid partial updates we use a double buffer and update
			// the real buffer in Draw()
			pIdx := currentIdx * 4
			if newState[currentIdx] {
				// White
				g.doubleBuffer[pIdx] = 0xff
				g.doubleBuffer[pIdx+1] = 0xff
				g.doubleBuffer[pIdx+2] = 0xff
				g.doubleBuffer[pIdx+3] = 0xff
			} else {
				// Black
				g.doubleBuffer[pIdx] = 0
				g.doubleBuffer[pIdx+1] = 0
				g.doubleBuffer[pIdx+2] = 0
				g.doubleBuffer[pIdx+3] = 0xff
			}
		}
	}

	copy(g.pixels, g.doubleBuffer)

	return newState
}

func (g *Game) Draw(screen *ebiten.Image) {
	screen.WritePixels(g.pixels)
}

// Layout defines the logical screen size
func (g *Game) Layout(outsideWidth, outsideHeight int) (int, int) {
	return g.Width, g.Height
}

func Show(width, height, scale int) {
	ebiten.SetWindowSize(width*scale, height*scale)
	ebiten.SetWindowTitle("Game of Life - Go Visualization")

	// Set TPS (Ticks Per Second) to control simulation speed
	// 10 ticks per second is easier to watch than the default 60
	ebiten.SetTPS(10)

	if err := ebiten.RunGame(NewGame(width, height)); err != nil {
		log.Fatal().Err(err)
	}
}
