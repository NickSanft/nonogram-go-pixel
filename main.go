package main

import (
	"fmt"
	"github.com/faiface/pixel"
	"github.com/faiface/pixel/imdraw"
	"github.com/faiface/pixel/pixelgl"
	"image"
	_ "image/png"
	"os"
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

	worldMap := [][]uint8{
		{0, 0, 0, 0, 0},
		{0, 1, 0, 1, 1},
		{2, 1, 2, 2, 0},
		{0, 0, 0, 0, 0},
		{1, 1, 1, 1, 1},
	}
	World.worldMap = worldMap


	sheet, err := getSheet(shapesDir)

	imd := imdraw.New(sheet)
	brd := &board{}

	World.brd = brd

	for !win.Closed() {
		imd.Clear()
		brd.draw(imd)
		imd.Draw(win)
		win.Update()
	}
}



func main() {
	pixelgl.Run(run)
}
