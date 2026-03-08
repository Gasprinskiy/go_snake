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
	Point
	DirectionState DirectionState
}

type Model struct {
	Ticking              bool
	Width                int
	Height               int
	Snake                []SnakePoint
	SnakePointMap        map[Point]int
	FoodPoint            *Point
	NextUpdateSnakePoint int
}

func tick() tea.Cmd {
	return tea.Tick(time.Millisecond*100, func(t time.Time) tea.Msg {
		return TickMsg{}
	})
}

func (m *Model) setFoodPint() {
	y := rand.Intn(Height - 1)
	x := rand.Intn(Width - 1)

	point := Point{Y: y, X: x}
	_, inSnakeBody := m.SnakePointMap[point]
	if inSnakeBody {
		m.FoodPoint = nil
		return
	}

	m.FoodPoint = &point
}

func (m *Model) updateSnakePointMap() {
	m.SnakePointMap = make(map[Point]int, cap(m.Snake))

	for i, p := range m.Snake {
		m.SnakePointMap[p.Point] = i
	}
}

func (m *Model) isSnake(point Point) bool {
	_, exists := m.SnakePointMap[point]
	return exists
}

func (m *Model) isFood(point Point) bool {
	if m.FoodPoint == nil {
		return false
	}

	return point == *m.FoodPoint
}

func (m *Model) isBorder(point Point) bool {
	return point.X == m.Width-1 || point.X == 0
}

func (m *Model) onKeyMsgUpdate(msgStr string) (tea.Model, tea.Cmd) {
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

		x := head.X + dStatePoint.X
		y := head.Y + dStatePoint.Y

		head.DirectionState = dStatePoint.DirectionState
		head.X = x
		head.Y = y

		if m.hasBorderCollision(head) {
			fmt.Println("YOU SUCK")
			return m, tea.Quit
		}

		head = m.handleYOutOfBorder(head)

		m.handleSelfHarm(head)
		m.handleEat(head)
		m.handleMovement(head)

		return m, nil
	}

	return m, nil
}

func (m *Model) onTickMsg() (tea.Model, tea.Cmd) {
	if !m.Ticking {
		m.Ticking = true
		return m, tick()
	}

	head := m.Snake[len(m.Snake)-1]
	switch head.DirectionState {
	case DirectionStateX:
		head.X += 1

	case DirectionStateXReverse:
		head.X -= 1

	case DirectionStateYReverse:
		head.Y -= 1

	case DirectionStateY:
		head.Y += 1
	}

	if m.hasBorderCollision(head) {
		fmt.Println("YOU SUCK")
		return m, tea.Quit
	}
	head = m.handleYOutOfBorder(head)

	m.handleSelfHarm(head)
	m.handleEat(head)
	m.handleMovement(head)

	return m, tick()
}

func (m *Model) hasBorderCollision(head SnakePoint) bool {
	return head.X < 0 || head.X >= m.Width-1
}

func (m *Model) handleYOutOfBorder(head SnakePoint) SnakePoint {
	switch {
	case head.Y > m.Height:
		head.Y = 0

	case head.Y < 0:
		head.Y = m.Height
	}

	return head
}

func (m *Model) handleEat(head SnakePoint) {
	isHeadOnFoodPoint := head.X == m.FoodPoint.X && head.Y == m.FoodPoint.Y
	if !isHeadOnFoodPoint {
		return
	}

	tale := m.Snake[0]
	newSnakePoint := SnakePoint{
		Point: Point{
			X: tale.X - 1,
			Y: tale.Y - 1,
		},
		DirectionState: tale.DirectionState,
	}

	m.Snake = append([]SnakePoint{newSnakePoint}, m.Snake...)
	m.setFoodPint()
}

func (m *Model) handleSelfHarm(head SnakePoint) {
	bodyIndex, headInOwnBody := m.SnakePointMap[head.Point]
	if !headInOwnBody {
		return
	}

	m.Snake = m.Snake[bodyIndex:len(m.Snake)]
}

func (m *Model) handleMovement(head SnakePoint) {
	m.Snake = append(m.Snake, head)
	m.Snake = m.Snake[1:len(m.Snake)]
}

func initialModel() *Model {
	snakeStartPoint := Point{
		X: 1,
		Y: 0,
	}

	snake := make([]SnakePoint, 0, Width*Height)
	snake = append(snake, SnakePoint{
		Point:          snakeStartPoint,
		DirectionState: DirectionStateX,
	})
	snakeMap := map[Point]int{
		snakeStartPoint: 0,
	}

	m := &Model{
		Width:         Width,
		Height:        Height,
		Snake:         snake,
		SnakePointMap: snakeMap,
	}
	m.setFoodPint()

	return m
}

func (m *Model) Init() tea.Cmd {
	return tick()
}

func (m *Model) View() string {
	var b strings.Builder

	m.updateSnakePointMap()

	b.WriteString(fmt.Sprintf("SCORE: %d", len(m.Snake)-1))
	b.WriteRune('\n')
	b.WriteRune('\n')

	for y := 0; y < m.Height; y++ {
		for x := 0; x < m.Width; x++ {

			point := Point{X: x, Y: y}

			switch {
			case m.isSnake(point):
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
	if m.FoodPoint == nil {
		m.setFoodPint()
	}

	switch msg := msg.(type) {
	case TickMsg:
		return m.onTickMsg()

	case tea.KeyMsg:
		return m.onKeyMsgUpdate(msg.String())
	}

	return m, nil
}

func main() {
	p := tea.NewProgram(initialModel())

	if _, err := p.Run(); err != nil {
		log.Fatal(err)
	}
}
