package main

import (
	"math/rand"
	"sync"
	"time"

	"github.com/nsf/termbox-go"
)

// Location represents a position in the game world.
type Location struct {
	x int
	y int
}

// Bullet represents a bullet fired by the player.
type Bullet struct {
	location Location
}

// Player represents the player in the game.
type Player struct {
	symbol   rune     // Symbol representing the player
	location Location // Current location of the player
	died     bool     // Flag indicating if the player is dead
}

// River represents the river obstacles in the game.
type River struct {
	l int // Left boundary of the river
	r int // Right boundary of the river
}

// World represents the game world.
type World struct {
	player    Player   // The player
	river     []River  // List of river obstacles
	height    int      // Height of the game world
	width     int      // Width of the game world
	nextStart int      // Next start position of the river
	nextEnd   int      // Next end position of the river
	bullets   []Bullet // List of bullets fired by the player
}

func newWorld() *World {
	maxX, maxY := termbox.Size()

	world := World{
		player: Player{
			symbol:   'A',
			location: Location{x: maxX / 2, y: maxY - 2},
			died:     false,
		},
		width:     maxX,
		height:    maxY,
		nextEnd:   maxX/2 + 10,
		nextStart: maxX/2 - 10,
		river:     make([]River, maxY)}

	for y := maxY - 1; y >= 0; y-- {
		world.river[y] = River{l: maxX/2 - 5, r: maxX/2 + 5}
	}
	for y := maxY - 1; y >= 0; y-- {
		if y <= 2*maxY/3 {
			if world.nextEnd < world.river[y+1].r {
				world.river[y].r = world.river[y+1].r - 1
			}
			if world.nextEnd > world.river[y+1].r {
				world.river[y].r = world.river[y+1].r + 1
			}
			if world.nextStart < world.river[y+1].l {
				world.river[y].l = world.river[y+1].l - 1
			}
			if world.nextStart > world.river[y+1].l {
				world.river[y].l = world.river[y+1].l + 1
			}
			if world.nextStart == world.river[y+1].l {
				world.river[y].l = world.nextStart
			}
			if world.nextEnd == world.river[y+1].r {
				world.river[y].r = world.nextEnd
			}

			// Randomize river boundaries
			if world.nextStart == world.river[y].l || world.nextEnd == world.river[y].r || (world.river[y].l+5) >= world.river[y].r {
				if rand.Intn(10) > 8 {
					world.nextStart = rand.Intn(40) - 20 + world.nextStart
					world.nextEnd = 50 - rand.Intn(40) + world.nextStart
				}
			}

		}
	}
	return &world
}

// draw function is responsible for rendering the game world.
func draw(wg *sync.WaitGroup, w *World) {
	defer wg.Done()
	for !w.player.died {
		termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

		// Draw the map
		drawMap(w)

		// Draw the bullets
		drawBullets(w)

		// Draw the player
		drawPlayer(w)

		termbox.HideCursor()
		termbox.Flush()
		time.Sleep(200 * time.Millisecond)
	}
}

// drawMap function draws the river obstacles on the screen.
func drawMap(w *World) {

	for y := 0; y < len(w.river); y++ {
		for lx := 0; lx < w.river[y].l; lx++ {
			termbox.SetCell(lx, y, ' ', termbox.ColorDefault, termbox.ColorGreen)
		}
		for rx := w.river[y].r; rx < w.width; rx++ {
			termbox.SetCell(rx, y, ' ', termbox.ColorDefault, termbox.ColorGreen)
		}
		for re := w.river[y].l; re < w.river[y].r; re++ {
			termbox.SetCell(re, y, ' ', termbox.ColorDefault, termbox.ColorBlue)
		}
	}
}

// moveBullets function updates the position of bullets and removes bullets when they collide with obstacles.
func moveBullets(w *World) {
	for i := 0; i < len(w.bullets); i++ {
		// Move the bullet up
		w.bullets[i].location.y--

		// Check if the bullet collides with an obstacle (green area)

		if w.bullets[i].location.x <= w.river[w.bullets[i].location.y].l ||
			w.bullets[i].location.x >= w.river[w.bullets[i].location.y].r ||
			w.bullets[i].location.y == 0 {
			// Remove the bullet if it collides with an obstacle
			w.bullets = append(w.bullets[:i], w.bullets[i+1:]...)
			i-- // Decrement the index as the bullet is removed

		}
	}
}

// drawBullets function draws the bullets fired by the player on the screen.
func drawBullets(w *World) {
	for _, bullet := range w.bullets {
		termbox.SetCell(bullet.location.x, bullet.location.y, '|', termbox.ColorDefault, termbox.ColorBlue)
	}
}

// drawPlayer function draws the player on the screen.
func drawPlayer(w *World) {
	termbox.SetChar(w.player.location.x, w.player.location.y, w.player.symbol)
}

// physics function simulates the physics of the game world.
func physics(wg *sync.WaitGroup, w *World) {
	defer wg.Done()
	for !w.player.died {
		// Check if player is out of bounds
		if w.player.location.x < w.river[w.player.location.y].l ||
			w.player.location.x >= w.river[w.player.location.y].r {
			w.player.died = true
		}

		// Shift the river obstacles
		for y := w.height - 1; y > 0; y-- {
			w.river[y] = w.river[y-1]
		}

		// Update river boundaries
		if w.nextEnd < w.river[0].r {
			w.river[0].r -= 1
		}
		if w.nextEnd > w.river[0].r {
			w.river[0].r += 1
		}
		if w.nextStart < w.river[0].l {
			w.river[0].l -= 1
		}
		if w.nextStart > w.river[0].l {
			w.river[0].l += 1
		}

		// Randomize river boundaries
		if w.nextStart == w.river[0].l || w.nextEnd == w.river[0].r || (w.river[0].l+10) >= w.river[0].r {
			if rand.Intn(10) > 8 {
				w.nextStart = rand.Intn(40) - 20 + w.nextStart
				w.nextEnd = 50 - rand.Intn(40) + w.nextStart
			}
		}
		time.Sleep(200 * time.Millisecond)
	}
}

// listenToKeyboard function listens to keyboard input and updates the player's position accordingly.
func listenToKeyboard(w *World) {
	for !w.player.died {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Ch {
			case 'q':
				w.player.died = true
			case 'w':
				if w.player.location.y > 1 {
					w.player.location.y -= 1
				}
			case 's':
				if w.player.location.y < w.height-2 {
					w.player.location.y += 1
				}
			case 'd':
				if w.player.location.x < w.width-2 {
					w.player.location.x += 1
				}
			case 'a':
				if w.player.location.x > 1 {
					w.player.location.x -= 1
				}
			default:
				switch ev.Key {
				// TODO  همزمانی تیر و حرکت
				case termbox.KeySpace:
					// Shoot bullet when space key is pressed
					newBullet := Bullet{location: Location{x: w.player.location.x, y: w.player.location.y}}
					w.bullets = append(w.bullets, newBullet)
				}
			}
		case termbox.EventError:
			panic(ev.Err)
		}
	}
}

func main() {
	// Initialize the screen
	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()

	// Initialize the game
	world := newWorld()

	var wg sync.WaitGroup
	wg.Add(2)

	// Start drawing and physics goroutines
	go draw(&wg, world)
	go physics(&wg, world)

	// Start a separate goroutine for moving bullets
	go func() {
		for {
			moveBullets(world)
			time.Sleep(100 * time.Millisecond) // Adjust the sleep time for bullet speed
		}
	}()

	// Listen to keyboard input
	go listenToKeyboard(world)

	wg.Wait()
}
