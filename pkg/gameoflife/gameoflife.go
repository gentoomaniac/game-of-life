package gameoflife

import (
	"math/rand"
	"time"

	log "github.com/rs/zerolog/log"

	"github.com/hajimehoshi/ebiten/v2"
	"github.com/hajimehoshi/ebiten/v2/inpututil"
)

type Game struct {
	paused bool
	Width  int
	Height int

	highLife    bool
	randomRaise *float64

	tick  int64
	Cells [][]bool

	pixels       []byte
	doubleBuffer []byte

	rng *rand.Rand
}

func NewGame(width, height int, density float64, highLife bool, seed *int64, randomraise *float64) *Game {
	realSeed := time.Now().UnixMicro()
	if seed != nil {
		realSeed = *seed
	}
	source := rand.NewSource(realSeed)
	rng := rand.New(source)

	g := &Game{
		paused:      false,
		Width:       width,
		Height:      height,
		highLife:    highLife,
		randomRaise: randomraise,
		rng:         rng,

		doubleBuffer: make([]byte, width*height*4),
		pixels:       make([]byte, width*height*4),
	}

	// initial randomness
	cells := make([]bool, width*height)
	for i := 0; i < len(cells); i++ {
		if rng.Float64() <= density {
			cells[i] = true
		}
	}
	g.Cells = append(g.Cells, cells)
	return g
}

func (g *Game) Update() error {
	if inpututil.IsKeyJustPressed(ebiten.KeyQ) {
		return ebiten.Termination
	}

	g.handleMouse()

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.paused = !g.paused
	}

	if !g.paused || (g.paused && inpututil.IsKeyJustPressed(ebiten.KeyN)) {
		if int(g.tick) == len(g.Cells)-1 {
			g.Cells = append(g.Cells, g.updateCells(g.Cells[g.tick]))
		}
		g.tick++
		g.drawState(g.Cells[g.tick])
	}

	if g.paused && inpututil.IsKeyJustPressed(ebiten.KeyP) {
		if g.tick > 0 {
			g.tick--
			g.drawState(g.Cells[g.tick])
		}
	}

	return nil
}

func (g *Game) handleMouse() {
	// Left klick to make alive
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()

		if x >= 0 && x < g.Width && y >= 0 && y < g.Height {
			idx := y*g.Width + x
			g.Cells[len(g.Cells)-1][idx] = true

			pIdx := idx * 4
			g.doubleBuffer[pIdx] = 0xff   // R
			g.doubleBuffer[pIdx+1] = 0xff // G
			g.doubleBuffer[pIdx+2] = 0xff // B
			g.doubleBuffer[pIdx+3] = 0xff // A
		}

		copy(g.pixels, g.doubleBuffer)
	}

	// Right click to kill
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		x, y := ebiten.CursorPosition()

		if x >= 0 && x < g.Width && y >= 0 && y < g.Height {
			idx := y*g.Width + x
			g.Cells[len(g.Cells)-1][idx] = false

			pIdx := idx * 4
			g.doubleBuffer[pIdx] = 0
			g.doubleBuffer[pIdx+1] = 0
			g.doubleBuffer[pIdx+2] = 0
			g.doubleBuffer[pIdx+3] = 0xff
		}
		copy(g.pixels, g.doubleBuffer)
	}
}

func (g *Game) drawState(state []bool) {
	for currentIdx := range state {
		pIdx := currentIdx * 4
		if state[currentIdx] {
			// Alive
			g.doubleBuffer[pIdx] = 0xff
			g.doubleBuffer[pIdx+1] = 0xff
			g.doubleBuffer[pIdx+2] = 0xff
			g.doubleBuffer[pIdx+3] = 0xff
		} else {
			// Dead
			g.doubleBuffer[pIdx] = 0
			g.doubleBuffer[pIdx+1] = 0
			g.doubleBuffer[pIdx+2] = 0
			g.doubleBuffer[pIdx+3] = 0xff
		}
	}
	copy(g.pixels, g.doubleBuffer)
}

func (g *Game) updateCells(state []bool) []bool {
	newState := make([]bool, len(state))

	for y := 0; y < g.Height; y++ {
		for x := 0; x < g.Width; x++ {
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

			currentIdx := y*g.Width + x
			alive := state[currentIdx]

			// Apply Game of Life Rules
			if alive && (neighbors == 2 || neighbors == 3) {
				newState[currentIdx] = true
			} else if !alive && neighbors == 3 {
				newState[currentIdx] = true
			} else if g.highLife && !alive && neighbors == 6 {
				newState[currentIdx] = true
			} else if g.randomRaise != nil {
				if g.rng.Float64() <= *g.randomRaise {
					newState[currentIdx] = true
				}
			} else {
				newState[currentIdx] = false
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

func Show(width, height, scale int, density float64, highLife bool, seed *int64, tps int, randomraise *float64) {
	ebiten.SetWindowSize(width*scale, height*scale)
	ebiten.SetWindowTitle("Game of Life - Go Visualization")

	// Set TPS (Ticks Per Second) to control simulation speed
	// 10 ticks per second is easier to watch than the default 60
	ebiten.SetTPS(tps)

	if err := ebiten.RunGame(NewGame(width, height, density, highLife, seed, randomraise)); err != nil {
		log.Fatal().Err(err)
	}
}
