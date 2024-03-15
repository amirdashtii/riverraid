package main

import (
	"fmt"
	"math/rand"
	"strings"
	"time"

	"github.com/nsf/termbox-go"
)

// Location represents a position in the game world.
type Location struct {
	x, y int
}

// Bullet represents a bullet fired by the player.
type Bullet struct {
	location Location
}

// PlayerStatus represents the status of a player.
type PlayerStatus int

const (
	Alive PlayerStatus = iota
	Dead
	DeadBody
	Quit
	Paused
)

const (
	messageHitByRock  = "A rock has hit you!"
	messageHitByEnemy = "An enemy has delivered a hit to you!"
	messageOutOfFuel  = "You have run out of fuel!"
)

// Player represents the player in the game.
type Player struct {
	symbol   rune     // Symbol representing the player
	location Location // Current location of the player
	score    int
	fuel     int
	status   PlayerStatus
	message  string
	lives    int // Number of remaining lives for the player
}

func newPlayer() *Player {
	maxX, maxY := termbox.Size()
	return &Player{
		symbol:   'A',
		location: Location{x: maxX / 2, y: maxY - 5},
		score:    0,
		fuel:     100,
		status:   Alive,
		lives:    3, // Set initial number of lives

	}
}

// River represents the river obstacles in the game.
type River struct {
	l int // Left boundary of the river
	r int // Right boundary of the river
}

var shouldExecute bool

// ThingStatus represents the status of an enemy.
type ThingStatus int

const (
	ThingAlive ThingStatus = iota
	ThingDeadBody
	ThingDead
)

// Enemy represents an enemy in the game.
type Enemy struct {
	location Location // Current location of the enemy
	status   ThingStatus
	symbol   rune
}

type Fuel struct {
	location Location
	status   ThingStatus
	length   int
	symbol   string
}

// World represents the game world.
type World struct {
	player    Player   // The player
	widthBox  int      // widthBox  of the game world
	heightBox int      // heightBox  of the game world
	nextStart int      // Next start position of the river
	nextEnd   int      // Next end position of the river
	river     []River  // List of river obstacles
	bullets   []Bullet // List of bullets fired by the player
	enemies   []Enemy
	fuels     []Fuel
}

func newWorld() *World {
	maxX, maxY := termbox.Size()

	world := World{
		player:    *newPlayer(),
		widthBox:  maxX,
		heightBox: maxY - 3,
		nextEnd:   maxX/2 + 10,
		nextStart: maxX/2 - 10,
		river:     make([]River, maxY),
		bullets:   []Bullet{},
		enemies:   []Enemy{},
		fuels:     []Fuel{},
	}
	for y := world.heightBox - 1; y >= 0; y-- {
		world.river[y] = River{l: maxX/2 - 5, r: maxX/2 + 5}
	}
	return &world
}

func hit(l1, l2 Location) bool {
	return l1.x == l2.x && l1.y == l2.y
}

func printText(s string, x, y int, fg, bg termbox.Attribute) {
	for _, ch := range s {
		termbox.SetCell(x, y, ch, fg, bg)
		x++
	}
}

// draw function is responsible for rendering the game world.
func draw(w *World) {
	drawMap(w)
	drawStatusBar(w)
	drawPlayer(w)
	drawBullets(w)
	drawEnemies(w)
	drawFuel(w)
}

// drawMap function draws the river obstacles on the screen.
func drawMap(w *World) {
	_, maxY := termbox.Size()

	for y := 0; y < w.heightBox; y++ {
		for l := 0; l <= w.widthBox; l++ {
			termbox.SetCell(l, y, ' ', termbox.ColorDefault, termbox.ColorGreen)
		}
		for re := w.river[y].l; re < w.river[y].r; re++ {
			termbox.SetCell(re, y, ' ', termbox.ColorDefault, termbox.ColorBlue)
		}
	}
	for y := w.heightBox; y <= maxY; y++ {
		printText(strings.Repeat(" ", w.widthBox), 0, y, termbox.ColorBlack, termbox.ColorDarkGray)
	}
}

func drawStatusBar(w *World) {
	// Print player score on the terminal
	formattedText := fmt.Sprintf("Score: %+v", w.player.score)
	printText(formattedText, w.widthBox/4, w.heightBox+1, termbox.ColorWhite, termbox.ColorDarkGray)
	formattedText = fmt.Sprintf("Lives: %+v", w.player.lives)
	printText(formattedText, w.widthBox/2-4, w.heightBox+2, termbox.ColorWhite, termbox.ColorDarkGray)
	formattedText = fmt.Sprintf("q: quit, p:pause %+v", w.player.score)
	printText(formattedText, w.widthBox/4*3, w.heightBox+1, termbox.ColorWhite, termbox.ColorDarkGray)

	printText(" F U E L  ", w.widthBox/2-5, w.heightBox+1, termbox.ColorBlack, termbox.ColorCyan)
	fu := w.player.fuel / 10
	switch {
	case fu == 0:
		w.player.status = Dead
		w.player.symbol = 'X'
		w.player.message = messageOutOfFuel
	case fu <= 3:
		printText(" F U E L  "[:fu], w.widthBox/2-5, w.heightBox+1, termbox.ColorBlack, termbox.ColorRed)
	case fu > 3:
		printText(" F U E L  "[:fu], w.widthBox/2-5, w.heightBox+1, termbox.ColorBlack, termbox.ColorYellow)
	}
}

// moveBullets function updates the position of bullets and removes bullets when they collide with obstacles.
func moveBullets(w *World) {
mainloop:
	for i := len(w.bullets) - 1; i >= 0; i-- {
		// Move the bullet up
		w.bullets[i].location.y--
		// Check if the bullet collides with an obstacle (green area)
		if w.bullets[i].location.x <= w.river[w.bullets[i].location.y].l ||
			w.bullets[i].location.x >= w.river[w.bullets[i].location.y].r ||
			w.bullets[i].location.y == 0 {
			// Remove the bullet if it collides with an obstacle
			w.bullets = append(w.bullets[:i], w.bullets[i+1:]...)
		} else {
			for j := len(w.enemies) - 1; j >= 0; j-- {
				switch w.enemies[j].status {
				case ThingAlive:
					if hit(w.bullets[i].location, w.enemies[j].location) ||
						hit(Location{w.bullets[i].location.x, w.bullets[i].location.y - 1}, w.enemies[j].location) {
						w.enemies[j].status = ThingDeadBody
						w.enemies[j].symbol = 'X'
						w.bullets = append(w.bullets[:i], w.bullets[i+1:]...)
						w.player.score += 10
						continue mainloop
					}

				case ThingDeadBody:
					w.enemies[j].status = ThingDead

				case ThingDead:
					w.enemies = append(w.enemies[:j], w.enemies[j+1:]...)
				}
			}
			for j := len(w.fuels) - 1; j >= 0; j-- {
				switch w.fuels[j].status {
				case ThingAlive:
					if w.bullets[i].location.x == w.fuels[j].location.x &&
						w.bullets[i].location.y <= w.fuels[j].location.y &&
						w.bullets[i].location.y >= w.fuels[j].location.y-4 {
						w.fuels[j].status = ThingDeadBody
						w.fuels[j].symbol = " X X"
						w.bullets = append(w.bullets[:i], w.bullets[i+1:]...)
						w.player.score += 10
						continue mainloop
					}
				case ThingDeadBody:
					w.fuels[j].status = ThingDead
				case ThingDead:
					w.fuels = append(w.fuels[:j], w.fuels[j+1:]...)
				}
			}
		}
	}
}

// drawBullets function draws the bullets fired by the player on the screen.
func drawBullets(w *World) {
	for _, bullet := range w.bullets {
		termbox.SetCell(bullet.location.x, bullet.location.y, '|', termbox.ColorDefault, termbox.ColorBlue)
	}
}

func drawEnemies(w *World) {
	for _, enemy := range w.enemies {
		termbox.SetCell(enemy.location.x, enemy.location.y, enemy.symbol, termbox.ColorDefault, termbox.ColorBlue)
	}
}

func drawFuel(w *World) {
	for _, fuel := range w.fuels {
		termbox.SetCell(fuel.location.x, fuel.location.y, rune(fuel.symbol[3]), termbox.ColorDefault, termbox.ColorWhite)
		termbox.SetCell(fuel.location.x, fuel.location.y-1, rune(fuel.symbol[2]), termbox.ColorDefault, termbox.ColorCyan)
		termbox.SetCell(fuel.location.x, fuel.location.y-2, rune(fuel.symbol[1]), termbox.ColorDefault, termbox.ColorWhite)
		termbox.SetCell(fuel.location.x, fuel.location.y-3, rune(fuel.symbol[0]), termbox.ColorDefault, termbox.ColorCyan)
	}
}

// drawPlayer function draws the player on the screen.
func drawPlayer(w *World) {
	termbox.SetChar(w.player.location.x, w.player.location.y, w.player.symbol)
}
func startgame(w *World) {
	for i := w.heightBox / 3 * 2; i > 0; i-- {
		draw(w)
		ShiftRiver(w)
		moveAddItems(w)
		termbox.Flush()
		time.Sleep(10 * time.Millisecond)
	}
	w.player.status = Paused
}

func moveAddItems(w *World) {
	// Move enemies and add new enemies
	for i := len(w.enemies) - 1; i >= 0; i-- {
		if w.enemies[i].location.y >= w.heightBox-1 {
			w.enemies = append(w.enemies[:i], w.enemies[i+1:]...)
		} else {
			w.enemies[i].location.y++
		}
	}
	if rand.Intn(10) > 5 {
		x := rand.Intn(w.river[0].r-w.river[0].l) + w.river[0].l
		newEnemy := Enemy{location: Location{x: x, y: 0}, symbol: 'E', status: ThingAlive}
		w.enemies = append(w.enemies, newEnemy)
	}

	for i := len(w.fuels) - 1; i >= 0; i-- {
		if w.fuels[i].location.y >= w.heightBox-1 {
			w.fuels = append(w.fuels[:i], w.fuels[i+1:]...)
		} else {
			w.fuels[i].location.y++
		}
	}
	if rand.Intn(10) > 8 {
		x := rand.Intn(w.river[0].r-w.river[0].l) + w.river[0].l
		newFuel := Fuel{location: Location{x: x, y: 0}, length: 4, symbol: "FUEL", status: ThingAlive}
		w.fuels = append(w.fuels, newFuel)
	}
}

func ShiftRiver(w *World) { // Shift the river obstacles
	for y := w.heightBox - 1; y > 0; y-- {
		w.river[y] = w.river[y-1]
	}

	// Update river boundaries
	if w.nextEnd < w.river[0].r {
		w.river[0].r--
	}
	if w.nextEnd > w.river[0].r {
		w.river[0].r++
	}
	if w.nextStart < w.river[0].l {
		w.river[0].l--
	}
	if w.nextStart > w.river[0].l {
		w.river[0].l++
	}

	// Randomize river boundaries
	if w.nextStart == w.river[0].l || w.nextEnd == w.river[0].r || (w.river[0].l+10) >= w.river[0].r {
		if rand.Intn(10) > 8 {
			w.nextStart = rand.Intn(40) - 20 + w.nextStart
			w.nextEnd = 50 - rand.Intn(40) + w.nextStart
		}
	}
}

// physics function simulates the physics of the game world.
func physics(w *World) {
	shouldExecute = !shouldExecute
	if shouldExecute {
		w.player.fuel--
		// Check player boundaries and enemy collisions
		if w.player.location.x < w.river[w.player.location.y].l ||
			w.player.location.x >= w.river[w.player.location.y].r {
			w.player.status = Dead
			w.player.symbol = 'X'
			w.player.message = messageHitByRock
		} else {
			for i := len(w.enemies) - 1; i >= 0; i-- {
				if hit(w.enemies[i].location, w.player.location) {
					w.player.status = Dead
					w.player.symbol = 'X'
					w.player.message = messageHitByEnemy
					break
				}
			}
		}
		for i := len(w.fuels) - 1; i >= 0; i-- {
			if w.player.location.x == w.fuels[i].location.x &&
				w.player.location.y <= w.fuels[i].location.y &&
				w.player.location.y >= w.fuels[i].location.y-4 {
				w.player.fuel += 10
			}
			if w.player.fuel > 100 {
				w.player.fuel = 100
			}
		}
		ShiftRiver(w)
		moveAddItems(w)
	}
	moveBullets(w)
	time.Sleep(100 * time.Millisecond)
}

// listenToKeyboard function listens to keyboard input and updates the player's position accordingly.
var previouStatus PlayerStatus = Alive

func listenToKeyboard(w *World) {
	for w.player.status != Quit {
		switch ev := termbox.PollEvent(); ev.Type {
		case termbox.EventKey:
			switch ev.Ch {
			case 'q':
				w.player.status = Quit
			case 'p':
				if w.player.status != Paused {
					previouStatus = w.player.status
					w.player.status = Paused
				} else {
					w.player.status = previouStatus
				}
			case 'w':
				if w.player.location.y > 1 {
					w.player.location.y -= 1
				}
			case 's':
				if w.player.location.y < w.heightBox-1 {
					w.player.location.y += 1
				}
			case 'd':
				if w.player.location.x < w.widthBox-1 {
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
					if w.player.status == Paused {
						w.player.status = previouStatus
					}
					if w.player.status == DeadBody {
						// Place the player in the middle of the river and pause the game
						w.player.symbol = 'A'
						w.player.fuel = 100
						w.player.location.x = (w.river[w.heightBox-1].r + w.river[w.heightBox-1].l) / 2
						w.player.location.y = w.heightBox - 1
						w.player.status = Alive
					}
					newBullet := Bullet{location: Location{x: w.player.location.x, y: w.player.location.y}}
					w.bullets = append(w.bullets, newBullet)
				case termbox.KeyEsc:
					w.player.status = Quit
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
	startgame(world)

	// Listen to keyboard input
	go listenToKeyboard(world)

	shouldExecute = false
	for world.player.status != Quit {
		termbox.HideCursor()
		switch world.player.status {
		case Alive:
			draw(world)
			physics(world)
		case Dead:
			// Reduce remaining lives when the player dies
			if world.player.lives == 0 {
				// End the game if no remaining lives
				printText(fmt.Sprintf("You burned! Game over. Your score: %d", world.player.score), world.widthBox/2-20, world.heightBox/2, termbox.ColorDefault, termbox.ColorDefault)
				formattedScore := fmt.Sprintf("Your Score: %+v", world.player.score)
				printText(formattedScore, world.widthBox/2-len(formattedScore)/2, world.heightBox/2+1, termbox.ColorDefault, termbox.ColorDefault)
				} else {
				world.player.lives--
				world.player.status = DeadBody
				text := fmt.Sprintf("Oh no! %s", world.player.message)
				printText(text, world.widthBox/2-len(text)/2, world.heightBox/2, termbox.ColorDefault, termbox.ColorDefault)
				text = fmt.Sprintf("Your score: %d", world.player.score)
				printText(text, world.widthBox/2-len(text)/2, world.heightBox/2+1, termbox.ColorDefault, termbox.ColorDefault)
				text = fmt.Sprintf("Lives remaining: %d", world.player.lives)
				printText(text, world.widthBox/2-len(text)/2, world.heightBox/2+2, termbox.ColorDefault, termbox.ColorDefault)
				text = "Press space to continue..."
				printText(text, world.widthBox/2-len(text)/2, world.heightBox/2+3, termbox.ColorDefault, termbox.ColorDefault)

			}
			drawPlayer(world)
		case Paused:
			if world.player.lives == 0 {
				mess := "Paused. Press Space to continue."
				printText(mess, world.widthBox/2-len(mess)/2, world.heightBox/2, termbox.ColorDefault, termbox.ColorDefault)
				formattedScore := fmt.Sprintf("Your Score: %+v", world.player.score)
				printText(formattedScore, world.widthBox/2-len(formattedScore)/2, world.heightBox/2+1, termbox.ColorDefault, termbox.ColorDefault)
			}
		}
		termbox.Flush()
	}
}
