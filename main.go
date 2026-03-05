package main

import (
	"fmt"
	"log"
	"math/rand"
	"strings"
	"time"

	tea "github.com/charmbracelet/bubbletea"
)

const Height = 20
const Width = 40

type DirectionState int

const (
	DirectionStateY DirectionState = iota
	DirectionStateX
	DirectionStateYReverse
	DirectionStateXReverse
)

var OpositeDirectionMap = map[DirectionState]DirectionState{
	DirectionStateY:        DirectionStateYReverse,
	DirectionStateYReverse: DirectionStateY,
	DirectionStateX:        DirectionStateXReverse,
	DirectionStateXReverse: DirectionStateX,
}

type TickMsg struct{}

type Point struct {
	X int
	Y int
}

type DirectionStatePoint struct {
	Point
	DirectionState
}

var DirectionStatePointKeyMap = map[string]DirectionStatePoint{
	"right": {
		Point: Point{
			X: 1,
		},
		DirectionState: DirectionStateX,
	},
	"left": {
		Point: Point{
			X: -1,
		},
		DirectionState: DirectionStateXReverse,
	},
	"up": {
		Point: Point{
			Y: -1,
		},
		DirectionState: DirectionStateYReverse,
	},
	"down": {
		Point: Point{
			Y: 1,
		},
		DirectionState: DirectionStateY,
	},
}

type SnakePoint struct {
	PositionX      int
	PositionY      int
	DirectionState DirectionState
}

type Model struct {
	Ticking              bool
	Width                int
	Height               int
	Snake                []SnakePoint
	FoodPoint            Point
	NextUpdateSnakePoint int
}

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

func getFoodPint() Point {
	y := rand.Intn(Height - 1)
	x := rand.Intn(Width - 1)

	return Point{Y: y, X: x}
}

func (m *Model) isSnake(point Point, snakeMap map[Point]struct{}) bool {
	_, exists := snakeMap[point]
	return exists
}

func (m *Model) isFood(point Point) bool {
	return point == m.FoodPoint
}

func (m *Model) isBorder(point Point) bool {
	return point.X == m.Width-1 || point.X == 0
}

func initialModel() *Model {
	snake := make([]SnakePoint, 0, Width*Height)
	snake = append(snake, SnakePoint{
		PositionX:      1,
		PositionY:      0,
		DirectionState: DirectionStateX,
	})

	return &Model{
		Width:     Width,
		Height:    Height,
		Snake:     snake,
		FoodPoint: getFoodPint(),
	}
}

func (m *Model) Init() tea.Cmd {
	return tick()
}

func (m *Model) View() string {
	var b strings.Builder

	snakeMap := make(map[Point]struct{}, len(m.Snake))
	for _, p := range m.Snake {
		snakeMap[Point{X: p.PositionX, Y: p.PositionY}] = struct{}{}
	}

	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {

			point := Point{X: x, Y: y}

			switch {
			case m.isSnake(point, snakeMap):
				b.WriteString("██")

			case m.isFood(point):
				b.WriteString("● ")

			case m.isBorder(point):
				b.WriteString("| ")

			default:
				b.WriteString("  ")
			}

		}
		b.WriteRune('\n')
	}

	return b.String()
}

func (m *Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case TickMsg:
		if !m.Ticking {
			m.Ticking = true
			return m, tick()
		}

		head := m.Snake[len(m.Snake)-1]
		switch head.DirectionState {
		case DirectionStateX:
			head.PositionX += 1

		case DirectionStateXReverse:
			head.PositionX -= 1

		case DirectionStateYReverse:
			head.PositionY -= 1

		case DirectionStateY:
			head.PositionY += 1
		}

		if head.PositionX < 0 || head.PositionX >= m.Width-1 || head.PositionY < 0 || head.PositionY >= m.Height {
			fmt.Println("YOU SUCK")
			return m, tea.Quit
		}

		if head.PositionX == m.FoodPoint.X && head.PositionY == m.FoodPoint.Y {
			tale := m.Snake[0]
			newSnakePoint := SnakePoint{
				PositionX:      tale.PositionX - 1,
				PositionY:      tale.PositionY - 1,
				DirectionState: tale.DirectionState,
			}

			m.Snake = append([]SnakePoint{newSnakePoint}, m.Snake...)
			m.FoodPoint = getFoodPint()
		}

		m.Snake = append(m.Snake, head)
		m.Snake = m.Snake[1:len(m.Snake)]

		return m, tick()

	case tea.KeyMsg:
		msgStr := msg.String()
		if msgStr == "q" || msgStr == "ctrl+c" {
			return m, tea.Quit
		}

		dStatePoint, exists := DirectionStatePointKeyMap[msgStr]

		if exists {
			head := m.Snake[len(m.Snake)-1]
			if head.DirectionState == dStatePoint.DirectionState {
				return m, nil
			}

			opositeDirection := OpositeDirectionMap[head.DirectionState]
			if opositeDirection == dStatePoint.DirectionState {
				return m, nil
			}
			m.Ticking = false

			x := head.PositionX + dStatePoint.X
			y := head.PositionY + dStatePoint.Y

			head.DirectionState = dStatePoint.DirectionState
			head.PositionX = x
			head.PositionY = y

			if head.PositionX == m.FoodPoint.X && head.PositionY == m.FoodPoint.Y {
				tale := m.Snake[0]
				newSnakePoint := SnakePoint{
					PositionX:      tale.PositionX - 1,
					PositionY:      tale.PositionY - 1,
					DirectionState: tale.DirectionState,
				}

				m.Snake = append([]SnakePoint{newSnakePoint}, m.Snake...)
				m.FoodPoint = getFoodPint()
			}

			m.Snake = append(m.Snake, head)
			m.Snake = m.Snake[1:len(m.Snake)]

			return m, nil
		}
	}

	return m, nil
}

func main() {
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
