package main

import (
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"github.com/faiface/pixel/text"
	"golang.org/x/image/colornames"
	"golang.org/x/image/font/basicfont"
	"image"
	_ "image/png"
	"os"
	"strconv"
	"time"
)

const (
	WindowHeight = 800
	WindowWidth  = 800
	empty        = 0
	circle       = 1
	cross        = 2
	frameWidth   = 11
	frameHeight  = 11
	shapesDir    = "resources/shapes.png"
)

func getSheet(filePath string) (pixel.Picture, error) {
	imgFile, err := os.Open(filePath)
	defer imgFile.Close()
	if err != nil {
		fmt.Println("Cannot read file:", filePath, err)
		return nil, err
	}
	img, _, err := image.Decode(imgFile)
	if err != nil {
		fmt.Println("Cannot decode file: ", filePath, err)
		return nil, err
	}
	sheet := pixel.PictureDataFromImage(img)
	return sheet, nil
}

//Get a rectangle's coordinates after dividing the picture
func getFrame(frameWidth float64, frameHeight float64, xGrid int, yGrid int) pixel.Rect {
	return pixel.R(
		float64(xGrid)*frameWidth,
		float64(yGrid)*frameHeight,
		float64(xGrid+1)*frameWidth,
		float64(yGrid+1)*frameHeight,
	)
}

//Get a rect coordinates by index after dividing the picture to grids
func getRectInGrid(width float64, height float64, totalX int, totalY int, x int, y int) pixel.Rect {
	gridWidth := width / float64(totalX+1)
	gridHeight := height / float64(totalY+1)
	return pixel.R(float64(x)*gridWidth, float64(y)*gridHeight, float64(x+1)*gridWidth, float64(y+1)*gridHeight)
}

type block struct {
	frame pixel.Rect
	sheet pixel.Picture
	gridX int
	gridY int
}

type board struct {
	sheet pixel.Picture
}

func (brd *board) load(sheet pixel.Picture) {
	brd.sheet = sheet
}
func (brd *board) draw(t pixel.Target) error {
	emptyFrame := getFrame(frameWidth, frameHeight, 1, 0)
	crossFrame := getFrame(frameWidth, frameHeight, 0, 0)
	circleFrame := getFrame(frameWidth, frameHeight, 2, 0)
	worldMap := World.worldMap
	for i := 0; i < len(worldMap); i++ {
		for y := 0; y < len(worldMap[0]); y++ {
			var x = abs(i-len(worldMap)) - 1
			switch {
			case worldMap[i][y] == empty:
				b := block{frame: emptyFrame, gridX: x, gridY: y + 1, sheet: brd.sheet}
				b.draw(t)
			case worldMap[i][y] == circle:
				b := block{frame: circleFrame, gridX: x, gridY: y + 1, sheet: brd.sheet}
				b.draw(t)
			case worldMap[i][y] == cross:
				b := block{frame: crossFrame, gridX: x, gridY: y + 1, sheet: brd.sheet}
				b.draw(t)
			}
		}
	}
	return nil
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (blk block) draw(t pixel.Target) {
	sprite := pixel.NewSprite(nil, pixel.Rect{})
	sprite.Set(blk.sheet, blk.frame)
	pos := getRectInGrid(WindowWidth, WindowHeight, len(World.worldMap[0]), len(World.worldMap), blk.gridY, blk.gridX)
	sprite.Draw(t, pixel.IM.
		ScaledXY(pixel.ZV, pixel.V(
			pos.W()/sprite.Frame().W(),
			pos.H()/sprite.Frame().H(),
		)).
		Moved(pos.Center()),
	)
}

type world struct {
	brd         *board
	worldMap    [][]uint8
	solutionMap [][]uint8
}

var World = &world{}

func run() {
	cfg := pixelgl.WindowConfig{
		Title:  "I'm back for more nonograms!",
		Bounds: pixel.R(0, 0, WindowWidth, WindowHeight),
		VSync:  true,
	}
	win, err := pixelgl.NewWindow(cfg)
	if err != nil {
		panic(err)
	}

	solutionMap := [][]uint8{
		{1, 1, 1, 1, 0},
		{0, 1, 0, 1, 1},
		{0, 1, 0, 0, 0},
		{0, 0, 0, 0, 0},
		{1, 1, 1, 1, 1},
	}
	World.solutionMap = solutionMap

	worldMap := make([][]uint8, len(solutionMap))
	for i := range worldMap {
		worldMap[i] = make([]uint8, len(solutionMap[0]))
	}
	World.worldMap = worldMap

	sheet, err := getSheet(shapesDir)

	imd := imdraw.New(sheet)
	brd := &board{}

	World.brd = brd

	for !win.Closed() {
		if checkForWin() {
			win.Clear(colornames.Black)
			drawText(win, "YOU WON!!", win.Bounds().Center())
			win.Update()
			time.Sleep(2000 * time.Millisecond)
			break
		}
		checkMouseClicks(win, worldMap)
		getSolutionNumbers(win)

		imd.Clear()
		brd.draw(imd)
		imd.Draw(win)
		win.Update()
	}
}

func checkMouseClicks(win *pixelgl.Window, worldMap [][]uint8) {
	var leftPressed = win.JustPressed(pixelgl.MouseButtonLeft)
	var rightPressed = win.JustPressed(pixelgl.MouseButtonRight)
	if leftPressed || rightPressed {
		var pos = win.MousePosition()
		for i := 0; i < len(worldMap); i++ {
			for y := 0; y < len(worldMap[0]); y++ {
				var gridPos = getRectInGrid(WindowWidth, WindowHeight, len(World.worldMap[0]), len(World.worldMap), y+1, i)
				if gridPos.Contains(pos) {
					var x = abs(i-len(worldMap)) - 1
					var gridValue = worldMap[x][y]
					if leftPressed {
						if gridValue == empty {
							worldMap[x][y] = circle
						} else {
							worldMap[x][y] = empty
						}
					}
					if rightPressed {
						if gridValue == empty {
							worldMap[x][y] = cross
						} else {
							worldMap[x][y] = empty
						}
					}
				}
			}
		}
	}
}

func checkForWin() bool {
	var solutionMap = World.solutionMap
	var worldMap = World.worldMap
	for i := 0; i < len(solutionMap); i++ {
		for j := 0; j < len(solutionMap[0]); j++ {
			if solutionMap[i][j] == 1 && worldMap[i][j] != 1 {
				return false
			} else if solutionMap[i][j] != 1 && worldMap[i][j] == 1 {
				return false
			}
		}
	}
	return true
}

func getSolutionNumbers(win *pixelgl.Window) {
	var solutionMap = World.solutionMap

	for i := 0; i < len(solutionMap); i++ {
		var result = ""
		var xCount = 0

		for j := 0; j < len(solutionMap[0]); j++ {
			var currentVal = solutionMap[i][j]
			if currentVal == 1 {
				xCount++
			} else if xCount > 0 {
				result = result + strconv.Itoa(xCount) + " "
				//println(xCount)
				xCount = 0
			}
		}
		if xCount > 0 {
			result = result + strconv.Itoa(xCount) + " "
			//println(xCount)
			xCount = 0
		}
		//var rect = getRectInGrid(WindowWidth, WindowHeight, len(World.worldMap[0]), len(World.worldMap), 0, i)
		//println(result)
		var rect = getRectInGrid(WindowWidth, WindowHeight, len(solutionMap[0]), len(solutionMap), 0, abs(i-len(solutionMap))-1)
		drawText(win, result, rect.Min)
		//println("END OF LINE")
	}
	//println("END OF X")

	for i := 0; i < len(solutionMap); i++ {
		var result = ""
		var yCount = 0

		for j := 0; j < len(solutionMap[0]); j++ {
			var currentVal = solutionMap[j][i]
			if currentVal == 1 {
				yCount++
			} else if yCount > 0 {
				result = result + strconv.Itoa(yCount) + " "
				//println(yCount)
				yCount = 0
			}
		}
		if yCount > 0 {
			result = result + strconv.Itoa(yCount) + " "
			yCount = 0
		}
		//println(result)
		var rect = getRectInGrid(WindowWidth, WindowHeight, len(solutionMap[0]), len(solutionMap), i+1, len(solutionMap))
		drawText(win, result, rect.Min)
		//println("END OF LINE")
	}

}

func drawText(win *pixelgl.Window, textToPrint string, v pixel.Vec) {
	basicAtlas := text.NewAtlas(basicfont.Face7x13, text.ASCII)
	basicTxt := text.New(v, basicAtlas)
	basicTxt.Color = colornames.Red
	fmt.Fprintln(basicTxt, textToPrint)
	basicTxt.Draw(win, pixel.IM.Scaled(basicTxt.Orig, 4))
}

func main() {
	pixelgl.Run(run)
}
