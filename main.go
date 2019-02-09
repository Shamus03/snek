package main

import (
	"fmt"
	"math/rand"
	"time"

	termbox "github.com/nsf/termbox-go"
)

var (
	directionUp    = vector{0, -1}
	directionDown  = vector{0, 1}
	directionLeft  = vector{-1, 0}
	directionRight = vector{1, 0}
)

type vector struct {
	x, y int
}

func (v vector) Add(w vector) vector {
	return vector{v.x + w.x, v.y + w.y}
}

type snakeGame struct {
	player            snake
	fruits            []vector
	width, height     int
	loop              bool
	paused            bool
	message           string
	messageExpiration time.Time
}

func (g *snakeGame) showMessage(text string) {
	g.message = text
	g.messageExpiration = time.Now().Add(time.Second * 3)
}

func (g *snakeGame) reset(width, height int) *snakeGame {
	g.player.reset(width/2, height/2)
	g.fruits = nil
	g.width = width
	g.height = height
	g.fruits = []vector{g.randomPos()}
	g.messageExpiration = time.Now()
	return g
}

func (g *snakeGame) tick() {
	if g.paused || g.player.dead {
		return
	}

	// Move the player.
	g.player.body = append([]vector{
		g.player.body[0].Add(g.player.direction),
	}, g.player.body[0:len(g.player.body)-1]...)

	// Loop the player around or check for collision with the walls.
	if g.loop {
		if g.player.body[0].x < 0 {
			g.player.body[0].x = g.width - 1
		}
		if g.player.body[0].x >= g.width {
			g.player.body[0].x = 0
		}
		if g.player.body[0].y < 0 {
			g.player.body[0].y = g.height - 1
		}
		if g.player.body[0].y >= g.height {
			g.player.body[0].y = 0
		}
	} else if g.player.body[0].x < 0 ||
		g.player.body[0].x >= g.width ||
		g.player.body[0].y < 0 ||
		g.player.body[0].y >= g.height {
		g.player.dead = true
		g.showMessage("You ran into a wall!")
	}

	// Check for collision with body.
	for _, s := range g.player.body[1:] {
		if s == g.player.body[0] {
			g.player.dead = true
			g.showMessage("You ran into yourself!")
		}
	}

	// Check for fruit collision.
	for i := range g.fruits {
		if g.fruits[i] == g.player.body[0] {
			g.fruits[i] = g.randomPos()
			g.player.body = append(g.player.body, g.player.body[len(g.player.body)-1])
		}
	}
}

func (g snakeGame) randomPos() vector {
	filled := make(map[vector]bool)
	for _, s := range append(g.player.body, g.fruits...) {
		filled[s] = true
	}
	var choices []vector
	for x := 0; x < g.width; x++ {
		for y := 0; y < g.height; y++ {
			p := vector{x, y}
			if _, ok := filled[p]; !ok {
				choices = append(choices, p)
			}
		}
	}
	return choices[rand.Int()%len(choices)]
}

type snake struct {
	body      []vector
	direction vector
	dead      bool
}

func (s *snake) changeDirection(d vector) {
	if len(s.body) > 1 && s.body[0].Add(d) == s.body[1] {
		// Prevent the snake from turning back on itself.
		return
	}
	s.direction = d
}

func (s *snake) reset(x, y int) *snake {
	s.body = []vector{{x, y}}
	s.direction = directionUp
	s.dead = false
	return s
}

func draw(g snakeGame) {
	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	// Draw the walls.
	for x := 1; x < g.width+1; x++ {
		termbox.SetCell(x, 1, '-', termbox.ColorWhite, termbox.ColorDefault)
		termbox.SetCell(x, g.height+2, '-', termbox.ColorWhite, termbox.ColorDefault)
	}
	for y := 2; y < g.height+2; y++ {
		termbox.SetCell(0, y, '|', termbox.ColorWhite, termbox.ColorDefault)
		termbox.SetCell(g.width+1, y, '|', termbox.ColorWhite, termbox.ColorDefault)
	}

	// Draw the game message.
	if time.Now().Before(g.messageExpiration) {
		termboxDrawText(0, 0, g.message, termbox.ColorWhite, termbox.ColorDefault)
	}

	// Draw the fruits.
	for _, f := range g.fruits {
		termbox.SetCell(f.x+1, f.y+2, '*', termbox.ColorRed, termbox.ColorDefault)
	}

	// Draw the player.
	for i := len(g.player.body) - 1; i >= 0; i-- {
		s := g.player.body[i]
		var color termbox.Attribute
		if i == 0 {
			color = termbox.ColorYellow
		} else {
			color = termbox.ColorWhite
		}
		termbox.SetCell(s.x+1, s.y+2, '@', color, termbox.ColorDefault)
	}

	if g.paused {
		termboxDrawCenteredText(g.width/2+1, g.height/2+1, "Paused", termbox.ColorDefault, termbox.ColorRed)
	}

	termbox.Flush()
}

func termboxDrawText(x, y int, text string, fg, bg termbox.Attribute) {
	for i, ch := range text {
		termbox.SetCell(x+i, y, ch, fg, bg)
	}
}

func termboxDrawCenteredText(x, y int, text string, fg, bg termbox.Attribute) {
	termboxDrawText(x-(len(text)/2), y, text, fg, bg)
}

func main() {
	err := termbox.Init()
	if err != nil {
		panic(err)
	}
	defer termbox.Close()

	termbox.Clear(termbox.ColorDefault, termbox.ColorDefault)

	events := make(chan termbox.Event)
	go func() {
		for {
			events <- termbox.PollEvent()
		}
	}()
	speed := time.Millisecond * 200
	ticker := time.NewTicker(speed)

	var game snakeGame

loop:
	for {
		select {
		case ev := <-events:
			switch ev.Type {
			case termbox.EventKey:
				switch {
				// Game controls
				case ev.Key == termbox.KeyCtrlC:
					break loop
				case ev.Key == termbox.KeyCtrlR:
					game.reset(game.width, game.height)
				case ev.Ch == 'p':
					game.paused = !game.paused
				case ev.Ch == 'l':
					game.loop = !game.loop
					game.showMessage(fmt.Sprintf("Loop: %t", game.loop))
				case ev.Ch == '+':
					speed = speed * 3 / 4
					ticker = time.NewTicker(speed)
					game.showMessage("Speed increased!")
				case ev.Ch == '-':
					speed = speed * 4 / 3
					ticker = time.NewTicker(speed)
					game.showMessage("Speed decreased...")
				// Movement
				case ev.Key == termbox.KeyArrowUp:
					game.player.changeDirection(directionUp)
				case ev.Key == termbox.KeyArrowDown:
					game.player.changeDirection(directionDown)
				case ev.Key == termbox.KeyArrowLeft:
					game.player.changeDirection(directionLeft)
				case ev.Key == termbox.KeyArrowRight:
					game.player.changeDirection(directionRight)
				}
			case termbox.EventResize:
				game.reset(ev.Width-2, ev.Height-3)
			}
		case <-ticker.C:
			game.tick()
			draw(game)
		}
	}
}
