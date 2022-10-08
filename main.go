package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"time"

	"github.com/gdamore/tcell/v2"
)

const (
	SnakeSymbol     = 0x2588
	AppleSymbol     = tcell.RuneDiamond
	GameFrameWidth  = 30
	GameFrameHeight = 15
	GameFrameSymbol = '#'
)

type Point struct {
	row, col int
}

type Apple struct {
	point  *Point
	symbol rune
}

type Snake struct {
	parts          []*Point
	velRow, velCol int
	symbol         rune
}

/*
type GameObject struct {
	row, col, width, height int
	velRow, velCol          int
	symbol                  rune
}
*/

var (
	screen       tcell.Screen
	isGamePaused bool //? true by default
	isGameOver   bool
	score        int
	debugLog     string
	deletePoint  []*Point
	snake        *Snake //* empty struct
	apple        *Apple //* empty struct
	//gameObjects  []*GameObject //*empty slice
)

func main() {
	rand.Seed(time.Now().UnixNano())
	InitScreen()
	InitGameState()
	InputChan := InitUserInput()

	for !isGameOver {
		HandleUserInput(ReadInput(InputChan))
		UpdateStae()
		DrawState()

		time.Sleep(75 * time.Millisecond)
	}
	screenWide, screenHight := screen.Size()
	PrintStringCentered(screenWide/2, screenHight/2, "Game Over!!")
	PrintStringCentered(screenWide/2, screenHight/2+1, fmt.Sprintf("your Score is %d", score))
	screen.Show()
	time.Sleep(3 * time.Second)
	screen.Fini()

}

func InitScreen() {
	var err error
	screen, err = tcell.NewScreen()
	if err != nil {
		log.Fatalf("%+v", err)
	}
	if err := screen.Init(); err != nil {
		log.Fatalf("%+v", err)
	}
	defStyle := tcell.StyleDefault.
		Background(tcell.ColorBlack).
		Foreground(tcell.ColorWhite)
	screen.SetStyle(defStyle)
	screen.Clear()
}

func InitGameState() {
	//? filling struct object
	snake = &Snake{
		parts: []*Point{
			{7, 5}, //! tail
			{7, 4},
			{7, 3},
			{6, 3},
			{5, 3}, //! head
		},
		velRow: -1,
		velCol: 0,
		symbol: SnakeSymbol,
	}

	apple = &Apple{
		point:  &Point{10, 10},
		symbol: AppleSymbol,
	}

	//gameObjects = []*GameObject{}
}

func DrawState() {
	if isGamePaused {
		return
	}
	//screen.Clear()
	ClearScreen()
	PrintString(0, 0, debugLog) //? debug helper
	PrintGameFrame()
	DrawSnake()
	DrawApple()
	/*
		for _, obj := range gameObjects {
			PrintChar(obj.row, obj.col, obj.width, obj.height, obj.symbol)
		}
	*/
	screen.Show()
}

func ClearScreen() {
	for _, p := range deletePoint {
		DeleteCharInGameFrame(p.row, p.col)
	}
	deletePoint = []*Point{}
}

func DrawSnake() {
	for _, p := range snake.parts {
		PrintCharInGameFrame(p.row, p.col, 1, 1, snake.symbol)
		deletePoint = append(deletePoint, p)
	}
}

func DrawApple() {
	PrintCharInGameFrame(apple.point.row, apple.point.col, 1, 1, apple.symbol)
	deletePoint = append(deletePoint, apple.point)
}

//* drawing frame
func PrintGameFrame() {
	//  get size and print frame
	//* get top left corner gameframe
	topLeftRow, topLeftCol := GetGameFrameTopLeft()
	row, col := topLeftRow-1, topLeftCol-1
	width, hight := GameFrameWidth+2, GameFrameHeight+2
	//*	print unfilterd frame
	PrintUnfilledRect(row, col, width, hight, GameFrameSymbol)
}

func InitUserInput() chan string {
	//? create chanel for passing variabel string

	inputChan := make(chan string)
	go func() {
		for {
			switch ev := screen.PollEvent().(type) {
			case *tcell.EventKey:
				inputChan <- ev.Name()
			}
		}
	}()
	return inputChan
}

func ReadInput(inputChan chan string) string {
	//? get string form chanel
	var key string
	select {
	case key = <-inputChan:
	default:
		key = ""
	}
	return key
}

func HandleUserInput(key string) {
	quit := func() {
		screen.Fini()
		os.Exit(0)
	}
	if key == "Rune[q]" { //? casting rune to string
		quit()
	} else if key == "Rune[p]" {
		isGamePaused = !isGamePaused //? true became false
	} else if (key == "Up" || key == "Rune[w]") && snake.velRow != 1 {
		snake.velRow = -1
		snake.velCol = 0
	} else if (key == "Left" || key == "Rune[a]") && snake.velCol != 1 {
		snake.velRow = 0
		snake.velCol = -1
	} else if (key == "Down" || key == "Rune[s]") && snake.velRow != -1 {
		snake.velRow = 1
		snake.velCol = 0
	} else if (key == "Right" || key == "Rune[d]") && snake.velCol != -1 {
		snake.velRow = 0
		snake.velCol = 1
	}

}

func UpdateStae() {
	if isGamePaused {
		return
	}
	UpdateSnake()
	UpdateApple()
}

func UpdateSnake() {
	//* add new elemet to snake / moving snake
	head := GetSnakeHead()
	snake.parts = append(snake.parts, &Point{
		row: head.row + snake.velRow,
		col: head.col + snake.velCol,
	})
	//* deleting the tail if apple not eaten and add score if apple eaten
	if !AppleIsInsideSnake() {
		snake.parts = snake.parts[1:]
	} else {
		score++
	}
	if SnakeHittingWall() || SnakeEatingItself() {
		isGameOver = true
	}
}

func SnakeHittingWall() bool {
	head := GetSnakeHead()
	return head.row < 0 ||
		head.row >= GameFrameHeight ||
		head.col < 0 ||
		head.col >= GameFrameWidth
}
func SnakeEatingItself() bool {
	head := GetSnakeHead()
	for _, p := range snake.parts[:SnakeheadIndex()] {
		if p.row == head.row && p.col == head.col {
			return true
		}
	}
	return false
}

func GetSnakeHead() *Point {
	head := snake.parts[len(snake.parts)-1]
	return head
}

func SnakeheadIndex() int {
	return len(snake.parts) - 1
}

func UpdateApple() {
	//* generate new apple possition untill it's outside the snake
	for AppleIsInsideSnake() {
		apple.point.row, apple.point.col =
			rand.Intn(GameFrameHeight), rand.Intn(GameFrameWidth)
	}
}

func AppleIsInsideSnake() bool {
	for _, p := range snake.parts {
		if p.row == apple.point.row && p.col == apple.point.col {
			return true
		}
	}
	return false
}

//! helper function

//* PrintFilledRect
func PrintChar(row, col, width, height int, char rune) {
	for r := 0; r < height; r++ {
		for c := 0; c < width; c++ {
			screen.SetContent(col+c, row+r, char, nil, tcell.StyleDefault)
		}
	}
}

func DeleteChar(row, col int) {
	screen.SetContent(col, row, ' ', nil, tcell.StyleDefault)
}

func PrintString(row, col int, text string) {
	for _, r := range text {
		screen.SetContent(col, row, r, nil, tcell.StyleDefault)
		col++
	}
}

func PrintStringCentered(col, row int, text string) {
	col = col - len(text)/2
	PrintString(row, col, text)
}

//* membuat pagar
func PrintUnfilledRect(row, col, width, height int, char rune) {
	//*print first row (print baris pertama)
	for c := 0; c < width; c++ {
		screen.SetContent(col+c, row, char, nil, tcell.StyleDefault)
	}

	//*for each row
	//* print first col & last col(print awalan dan akiran pada tiap baris)
	for r := 1; r < height-1; r++ {
		screen.SetContent(col, row+r, '|', nil, tcell.StyleDefault)
		screen.SetContent(col+width-1, row+r, '|', nil, tcell.StyleDefault)
	}
	//* print last row(print baris terakhir)
	for c := 0; c < width; c++ {
		screen.SetContent(col+c, row+height-1, char, nil, tcell.StyleDefault)
	}

}

//* print character / object inside frame
func PrintCharInGameFrame(row, col, width, hight int, char rune) {
	r, c := GetGameFrameTopLeft()
	PrintChar(row+r, col+c, width, hight, char)
}

func DeleteCharInGameFrame(row, col int) {
	r, c := GetGameFrameTopLeft()
	DeleteChar(row+r, col+c)
}

func GetGameFrameTopLeft() (int, int) {
	screenWidth, screenHeigth := screen.Size()
	return screenHeigth/2 - GameFrameHeight/2, screenWidth/2 - GameFrameWidth/2
}
