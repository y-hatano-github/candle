package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/alexflint/go-arg"
	"github.com/charmbracelet/bubbles/timer"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/muesli/termenv"
)

// WIDTH is width of fire
const WIDTH = 16

// HEIGHT is height of file
const HEIGHT = 16

// FLUCTUATION is fluctuation of fire
const FLUCTUATION = 5

func main() {
	var args struct {
		DurationTime int64 `arg:"positional" help:"Duration in seconds (default: 0).\n                         0 keeps the fire burning indefinitely." default:"0"`
	}
	arg.MustParse(&args)

	m := initialModel(args.DurationTime)
	p := tea.NewProgram(m)
	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}

	m.o.CursorPrevLine(WIDTH + 1)
}

type model struct {
	timer     timer.Model
	st        time.Time
	err       error
	cnt       int
	v         *[WIDTH][HEIGHT]int
	c         []string
	dt        int64
	o         *termenv.Output
	r         *rand.Rand
	isLasting bool
}

func initialModel(d int64) model {
	isLasting := d <= 0
	var tm timer.Model
	if !isLasting {
		tm = timer.NewWithInterval(time.Duration(d)*time.Second, 50*time.Millisecond)
	}

	var v [WIDTH][HEIGHT]int
	for y := 0; y < HEIGHT; y++ {
		for x := 0; x < WIDTH; x++ {
			v[x][y] = 0
		}
	}

	output := termenv.NewOutput(os.Stdout)
	bgc := output.BackgroundColor()

	return model{
		timer: tm,
		st:    time.Now(),
		err:   nil,
		cnt:   0,
		v:     &v,
		c: []string{
			bgc.Sequence(true),
			"197",
			"203",
			"209",
			"215",
			"221",
			"227",
			"228",
			"229",
			"230",
			"231"},
		o:         output,
		dt:        d,
		r:         rand.New(rand.NewSource(time.Now().UnixNano())),
		isLasting: isLasting,
	}
}

func (m model) Fire() {
	sub := time.Since(m.st).Milliseconds()
	for x := 0; x < WIDTH; x++ {
		for y := 0; y < HEIGHT; y++ {

			if y < HEIGHT-2 {
				if x == WIDTH-1 {
					m.v[x][y] = (m.v[x][y] + m.v[x-1][y+1] + m.v[x][y+1] + m.v[x][y+2]) / 5
				} else if x == 0 {
					m.v[x][y] = (m.v[x][y] + m.v[x][y+1] + m.v[x+1][y+1] + m.v[x][y+2]) / 5
				} else {
					m.v[x][y] = (m.v[x][y] + m.v[x-1][y+1] + m.v[x][y+1] + m.v[x+1][y+1] + m.v[x][y+2]) / 5
				}
			} else if y < HEIGHT-1 {
				if x == WIDTH-1 {
					m.v[x][y] = (m.v[x][y] + m.v[x-1][y+1] + m.v[x][y+1]) / 5
				} else if x == 0 {
					m.v[x][y] = (m.v[x][y] + m.v[x][y+1] + m.v[x+1][y+1]) / 5
				} else if x < 3 || x > WIDTH-4 {
					m.v[x][y] = (m.v[x][y] + m.v[x-1][y+1] + m.v[x][y+1] + m.v[x+1][y+1]) / 5
				}
			} else {
				if x < 3 || x > WIDTH-4 {
					m.v[x][y] = m.v[x][y] / 5
				}
			}
			if m.r.Intn(FLUCTUATION) == 1 {
				m.v[x][y] = m.v[x][y] - 1
			}
			if m.v[x][y] < 0 {
				m.v[x][y] = 0
			}
		}
	}

	for x := 5; x < WIDTH-5; x++ {
		if !m.isLasting && sub > m.dt*1000-500 {
			m.v[x][HEIGHT-1] = 0
		} else {
			m.v[x][HEIGHT-1] = m.r.Intn(3) + len(m.c) - 3
		}
	}
	for x := 4; x < WIDTH-4; x++ {
		if !m.isLasting && sub > m.dt*1000-500 {
			m.v[x][HEIGHT-2] = 0
		} else {
			m.v[x][HEIGHT-2] = m.r.Intn(3) + len(m.c) - 3
		}
	}
}

func (m model) Init() tea.Cmd {
	if m.isLasting {
		return tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
			return t
		})
	}
	return m.timer.Init()
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			m.st = m.st.Add(-5 * time.Second)
			m.dt = 1
			if m.isLasting {
				m.isLasting = false
				m.timer = timer.NewWithInterval(time.Duration(m.dt)*time.Second, 50*time.Millisecond)
				return m, m.timer.Init()
			}
			return m, cmd
		}
	case time.Time: // lasting
		if m.isLasting {
			m.Fire()
			return m, tea.Tick(50*time.Millisecond, func(t time.Time) tea.Msg {
				return t
			})
		}
	case timer.TickMsg: // timer
		m.cnt = m.cnt + 1
		m.timer, cmd = m.timer.Update(msg)
		m.Fire()
		b := false
		for y := 0; y < HEIGHT; y++ {
			for x := 0; x < WIDTH; x++ {
				if m.v[x][y] != 0 {
					b = true
				}
			}
		}
		if !b {
			return m, tea.Quit
		}
	case timer.TimeoutMsg:
		return m, tea.Quit
	}

	return m, cmd
}

func (m model) View() string {
	var s string = "\n"
	for y := 0; y < HEIGHT; y++ {
		for x := 0; x < WIDTH; x++ {
			style := m.o.String(" ")
			s = fmt.Sprintf("%s%v", s, style.Background(termenv.TrueColor.Color(m.c[m.v[x][y]])))
		}
		s = s + "\n"
	}

	return s
}
