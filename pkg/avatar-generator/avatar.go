package avatargenerator

import (
	"fmt"
	"log"
	"os"
	"strings"

	"github.com/fogleman/gg"
)

const (
	ImageWidth      = 256
	ImageHeight     = 256
	BackgroundColor = "ffffff"
	TextColor       = "888888"
	FontSize        = 100
	FontPath        = "./Roboto-Regular.ttf"
)

func GetInitial(firstName, lastName string) string {
	if len(firstName) == 0 || len(lastName) < 2 {
		fmt.Println("Invalid names provided.")
		os.Exit(1)
	}
	return strings.ToUpper(string(firstName[0]) + string(lastName[0]))
}

func CreateProfileImage(initials, outputPath string) {
	dc := gg.NewContext(ImageWidth, ImageHeight)

	dc.SetHexColor(BackgroundColor)
	dc.Clear()

	if err := dc.LoadFontFace(FontPath, FontSize); err != nil {
		log.Fatalf("Could not load font: %v", err)
	}

	dc.SetHexColor(TextColor)

	dc.DrawStringAnchored(initials, ImageWidth/2, ImageHeight/2, 0.5, 0.5)

	dc.SavePNG(outputPath)
}
