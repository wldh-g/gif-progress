package cmd

import (
	"fmt"
	"image/color"
	"image/gif"
	"os"
	"path/filepath"

	gif_progress "github.com/nwtgck/gif-progress"
	"github.com/spf13/cobra"
	"gopkg.in/go-playground/colors.v1"
)

var barTop bool
var barHeight int
var barFPS int
var barColorString string
var outFile string
var inFile string

func init() {
	cobra.OnInitialize()
	RootCmd.Flags().BoolVarP(&barTop, "bar-top", "t", false, "Bar is on top")
	RootCmd.Flags().IntVar(&barHeight, "bar-height", 5, "Bar height")
	RootCmd.Flags().StringVarP(&barColorString, "bar-color", "c", "#ccc", "Bar color")
	RootCmd.Flags().IntVarP(&barFPS, "bar-fps", "f", 0, "Bar FPS (output gif fps will be changed if set)")
	RootCmd.Flags().StringVarP(&outFile, "out", "o", "", "Output gif")
	RootCmd.Flags().StringVarP(&inFile, "in", "i", "", "Input gif")
}

var RootCmd = &cobra.Command{
	Use:   os.Args[0],
	Short: "gif-progress",
	Long:  "Attach progress bar to animated GIF",
	RunE: func(cmd *cobra.Command, args []string) error {
		fmt.Println(inFile, outFile)

		r := os.Stdin
		w := os.Stdout

		if len(inFile) != 0 {
			// Open file and set source reader to infile
			f, err := os.Open(filepath.Clean(inFile))
			if err != nil {
				return err
			}
			defer f.Close()
			r = f
		}

		if len(outFile) != 0 {
			// Open file and set dest writer to outfile
			f, err := os.OpenFile(filepath.Clean(outFile), os.O_TRUNC|os.O_CREATE|os.O_RDWR, 0600)
			if err != nil {
				return err
			}
			defer f.Close()
			w = f
		}

		// Parse color
		hex, err := colors.ParseHEX(barColorString)
		if err != nil {
			return err
		}
		c := hex.ToRGB()
		barColor := color.RGBA{c.R, c.G, c.B, 255}
		inGif, err := gif.DecodeAll(r)
		if err != nil {
			return err
		}
		// Add progress bar to gif
		var outGif *gif.GIF
		if barFPS == 0 {
			outGif = gif_progress.AddProgressBar(inGif, barTop, barHeight, barColor)
		} else {
			outGif = gif_progress.AddProgressBarFPS(inGif, barTop, barHeight, barColor, barFPS)
		}
		// Write gif
		err = gif.EncodeAll(w, outGif)
		return err
	},
}
