package main

import (
	"math/rand"
	"time"

	"github.com/nsf/termbox-go"
)

type Location struct {
	x int
	y int
}
type Player struct {
	symbol   rune
	location Location
	died     bool
}
type River struct {
	l int
	r int
}
type World struct {
	player    Player
	river     []River
	height    int
	width     int
	nextStart int
	nextEnd   int
	// pointsChan chan (int)
}

var runeEvents = make(chan rune, 1)

func draw(w World) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)
	// draw the map
	for y := 0; y < w.height; y++ {
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

	termbox.HideCursor()

	//draw the player
	termbox.SetChar(w.player.location.x, w.player.location.y, w.player.symbol)
	termbox.Flush()
}

func physics(w World) World {
	// check if player died
	if w.player.location.x < w.river[w.player.location.y].l ||
		w.player.location.x >= w.river[w.player.location.y].r {
		w.player.died = true
	}

	// shift the map
	for y := w.height - 1; y > 0; y-- {
		w.river[y] = w.river[y-1]
	}

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

	if w.nextStart == w.river[0].l || w.nextEnd == w.river[0].r || (w.river[0].l+10) >= w.river[0].r {
		if rand.Intn(10) > 8 {

			w.nextStart = rand.Intn(40) - 20 + w.nextStart
			w.nextEnd = 50 - rand.Intn(40) + w.nextStart
		}
	}

	return w
}

func main() {

	// init the screen
	if err := termbox.Init(); err != nil {
		panic(err)
	}
	defer termbox.Close()
	maxX, maxY := termbox.Size()

	// init the game
	rivers := []River{}

	for i := maxY; i >= 0; i-- {
		rivers = append(rivers, River{l: maxX/2 - 5, r: maxX/2 + 5})
	}

	world := World{
		player: Player{
			symbol:   'A',
			location: Location{x: maxX / 2, y: maxY - 2},
			died:     false,
		},
		river:     rivers,
		width:     maxX,
		height:    maxY,
		nextEnd:   maxX/2 + 10,
		nextStart: maxX/2 - 10,
	}

	go func() {
		for {
			runeEvents <- termbox.PollEvent().Ch
		}
	}()

mainloop:
	for !world.player.died {
		// read and apply keyboard
		select {
		case key := <-runeEvents:
			switch key {
			case 'q':
				break mainloop
			case 'w':
				if world.player.location.y > 1 {
					world.player.location.y -= 1
				}
			case 's':
				if world.player.location.y < maxY-2 {
					world.player.location.y += 1
				}
			case 'd':
				if world.player.location.x < maxX-2 {
					world.player.location.x += 1
				}
			case 'a':
				if world.player.location.x > 1 {
					world.player.location.x -= 1
				}
			}
			for len(runeEvents) > 0 {
				<-runeEvents
			}
		case <-time.After(10 * time.Millisecond):
		}

		// physics
		world = physics(world)

		//draw
		draw(world)
		time.Sleep(200 * time.Millisecond)

	}

}
