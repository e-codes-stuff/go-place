package main

import (
	"bytes"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

const (
	IMAGE_HEIGHT = 256
	IMAGE_WIDTH  = 256
)

var DB *gorm.DB
var board = image.NewRGBA(image.Rect(0, 0, IMAGE_HEIGHT, IMAGE_WIDTH))

func main() {
	router := gin.Default()
	router.SetTrustedProxies(nil)

	var err error
	DB, err = gorm.Open(sqlite.Open("pixels.db"), &gorm.Config{})

	if err != nil {
		log.Fatal("Failed to connect to DB: ", err)
		return
	}

	setupDB(DB)

	rows, err := DB.Model(&RequestPutPixel{}).Rows()

	if err != nil {
		log.Fatal("could not scan rows: ", err)
	}

	// iterate through all changes to the image and apply each one
	for rows.Next() {
		var pixel RequestPutPixel
		DB.ScanRows(rows, &pixel)

		board.SetRGBA(pixel.X, pixel.Y, color.RGBA{pixel.R, pixel.G, pixel.B, 255})
	}

	rows.Close()

	router.GET("/image", getImage)
	router.POST("/image", putPixel)

	router.Run("localhost:8080")
}

func setupDB(db *gorm.DB) {
	db.AutoMigrate(RequestPutPixel{})
}

func getImage(c *gin.Context) {
	var im bytes.Buffer
	png.Encode(&im, board)

	c.DataFromReader(http.StatusOK, int64(im.Len()), "image/png", &im, map[string]string{})
}

type RequestPutPixel struct {
	// gorm fields added manually with `json:"-"` to ensure they are ignored when binding
	ID        uint           `gorm:"primaryKey" json:"-"`
	CreatedAt time.Time      `json:"-"`
	UpdatedAt time.Time      `json:"-"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	X int  `json:"x"`
	Y int  `json:"y"`
	R byte `json:"red"`
	G byte `json:"green"`
	B byte `json:"blue"`
}

func putPixel(c *gin.Context) {
	var pixel RequestPutPixel

	if err := c.BindJSON(&pixel); err != nil {
		fmt.Println("Error putting pixel:", err)
		return
	}

	if pixel.X > IMAGE_WIDTH-1 || pixel.Y > IMAGE_HEIGHT-1 || pixel.X < 0 || pixel.Y < 0 {
		return
	}

	// skip entire operation if pixel unchanged
	currentColor := board.At(pixel.X, pixel.Y)
	if (currentColor == color.RGBA{pixel.R, pixel.G, pixel.B, 255}) {
		return
	}

	fmt.Println(&pixel)

	DB.Create(&pixel)

	// update the current image
	board.SetRGBA(pixel.X, pixel.Y, color.RGBA{pixel.R, pixel.G, pixel.B, 255})
}
