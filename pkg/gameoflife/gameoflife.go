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

	highLife bool

	Cells [][]bool

	pixels       []byte
	doubleBuffer []byte

	rng *rand.Rand
}

func NewGame(width, height int, density float64, highLife bool, seed *int64) *Game {
	realSeed := time.Now().UnixMicro()
	if seed != nil {
		realSeed = *seed
	}
	source := rand.NewSource(realSeed)
	rng := rand.New(source)

	g := &Game{
		paused:   false,
		Width:    width,
		Height:   height,
		highLife: highLife,
		rng:      rng,

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

// Update handles the logic (Game of Life rules go here)
func (g *Game) Update() error {
	g.handleMouse()

	if inpututil.IsKeyJustPressed(ebiten.KeySpace) {
		g.paused = !g.paused
	}

	if !g.paused || (g.paused && inpututil.IsKeyJustPressed(ebiten.KeyN)) {
		g.Cells = append(g.Cells, g.updateCells(g.Cells[len(g.Cells)-1]))
	}

	return nil
}

func (g *Game) handleMouse() {
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonLeft) {
		x, y := ebiten.CursorPosition()

		if x >= 0 && x < g.Width && y >= 0 && y < g.Height {
			// 3. Calculate 1D Index
			idx := y*g.Width + x

			g.Cells[len(g.Cells)-1][idx] = true

			// 5. Update Visuals Instantly
			// We update the pixel buffer directly so we don't have to wait
			// for the next Render() cycle to see the change.
			pIdx := idx * 4
			g.doubleBuffer[pIdx] = 0xff   // R
			g.doubleBuffer[pIdx+1] = 0xff // G
			g.doubleBuffer[pIdx+2] = 0xff // B
			g.doubleBuffer[pIdx+3] = 0xff // A
		}

		copy(g.pixels, g.doubleBuffer)
	}

	// Right click to erase
	if ebiten.IsMouseButtonPressed(ebiten.MouseButtonRight) {
		x, y := ebiten.CursorPosition()

		if x >= 0 && x < g.Width && y >= 0 && y < g.Height {
			idx := y*g.Width + x
			g.Cells[len(g.Cells)-1][idx] = false

			// Update Visuals (Black)
			pIdx := idx * 4
			g.doubleBuffer[pIdx] = 0
			g.doubleBuffer[pIdx+1] = 0
			g.doubleBuffer[pIdx+2] = 0
			g.doubleBuffer[pIdx+3] = 0xff
		}
		copy(g.pixels, g.doubleBuffer)
	}
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
			} else if g.highLife && !alive && neighbors == 6 {
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

func Show(width, height, scale int, density float64, highLife bool, seed *int64) {
	ebiten.SetWindowSize(width*scale, height*scale)
	ebiten.SetWindowTitle("Game of Life - Go Visualization")

	// Set TPS (Ticks Per Second) to control simulation speed
	// 10 ticks per second is easier to watch than the default 60
	ebiten.SetTPS(10)

	if err := ebiten.RunGame(NewGame(width, height, density, highLife, seed)); err != nil {
		log.Fatal().Err(err)
	}
}
