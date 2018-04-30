package main

import (
	"crypto/rand"
	"crypto/sha256"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"io/ioutil"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/ssh"
	"golang.org/x/image/font"
	"golang.org/x/image/font/inconsolata"
	"golang.org/x/image/math/fixed"
)

var smooth int
var validBs = make([]Board, 0)

func main() {
	cores := flag.Int("cores", 4, "cores to use")
	smoo := flag.Int("smooth", 20, "diff, lower is harder")
	flag.Parse()

	smooth = *smoo
	fmt.Print("\n")

	// validBs := make([]Board, 0)
	for n := 0; n < 289; n++ {
		b := Board{}
		/*
			X = 9
			Y = 17

						XDim = 17
						YDim = 9
		*/

		// for y := 17; y != 0; y-- {
		// 	for x := 0; x < 16; x++ {
		// 		px := regular8x16.Mask.At(y, (n*17)+x)
		// 		r, _, _, _ := px.RGBA()
		// 		if r == 0 {
		// 			// fmt.Print(" ")
		// 		} else {
		// 			// fmt.Print("#")
		// 		}
		// 	}
		// 	// fmt.Print("\n")
		// }
		pixles := false
		xoffset := 4
		for x := xoffset; x < 14; x++ {
			for y := 0; y < 17; y++ {
				px := regular8x16.Mask.At(y, (n*17)+x)
				r, _, _, _ := px.RGBA()
				if r < 10 {
					fmt.Print(" ")
				} else {
					nx := x - xoffset
					if nx < 9 {
						fmt.Print("#")
						pixles = true
						// fmt.Printf("S X:%d Y:%d", x-4, y)
						b.Tiles[nx][y+3] = 5
					}
				}
			}
			fmt.Print("\n")
		}
		fmt.Printf("-----------\n")
		fmt.Print(b.String())
		// for x := 2; x < 9; x++ {
		// 	for y := 0; y < 17; y++ {
		// 		px := bold8x16.Mask.At(y, (n*17)+x)
		// 		r, _, _, _ := px.RGBA()
		// 		if r != 0 {
		// 			b.Tiles[x][y] = 4
		// 			// fmt.Printf("%d , %d \n", x, y)
		// 		}
		// 		// else {
		// 		// b.Tiles[y][x] = 0
		// 		// }
		// 	}
		// }
		b.Lowestscore = smooth
		if !pixles {
			b.Skip = true
		}
		validBs = append(validBs, b)
		// bold8x16.Mask.At
	}

	for k, v := range validBs {
		rend := v.String()
		white := color.RGBA{255, 255, 255, 255}
		img := image.NewRGBA(image.Rect(0, 0, 200, 200))
		draw.Draw(img, img.Bounds(), &image.Uniform{white}, image.ZP, draw.Src)

		lines := strings.Split(rend, "\n")
		for n, ln := range lines {
			addLabel(img, 0, 16+(n*16), ln)
		}

		os.Mkdir(fmt.Sprintf("./%d", k), 0777)
		f, err := os.Create(fmt.Sprintf("./%d/%d.png", k, 0))
		if err != nil {
			panic(err)
		}

		if err := png.Encode(f, img); err != nil {
			panic(err)
		}
		f.Close()

	}

	// nb := Board{}
	// eee := bold8x16.Mask.

	// fmt.Print(validBs[11].String())

	_, _, _, lastBoard := makeKey()

	stats := make(chan int)
	results := make(chan ArtKey)
	dista := make([]chan Board, *cores)
	for core := 0; core < *cores; core++ {
		dista[core] = make(chan Board, 1)
	}

	for core := 0; core < *cores; core++ {
		go worker(lastBoard, stats, results, dista[core])
	}

	go func() {
		for {
			tn := 0
			for core := 0; core < *cores; core++ {
				i := <-stats
				tn += i
			}
			fmt.Printf("%d keys per second\n", tn)
			tn = 0
		}
	}()

	imgc := 0
	for {
		aK := <-results
		rend := aK.Render.String()
		validBs[aK.GlyfID].Lowestscore = aK.Score

		white := color.RGBA{255, 255, 255, 255}
		img := image.NewRGBA(image.Rect(0, 0, 200, 200))
		draw.Draw(img, img.Bounds(), &image.Uniform{white}, image.ZP, draw.Src)

		lines := strings.Split(rend, "\n")
		for n, ln := range lines {
			addLabel(img, 0, 16+(n*16), ln)
		}

		os.Mkdir(fmt.Sprintf("./%d", aK.GlyfID), 0777)
		f, err := os.Create(fmt.Sprintf("./%d/%d.png", aK.GlyfID, aK.Score))
		if err != nil {
			panic(err)
		}

		if err := png.Encode(f, img); err != nil {
			panic(err)
		}
		f.Close()

		imgc++
		fmt.Print(rend)
		fmt.Print(validBs[aK.GlyfID].String())

		// using publicKey from above.
		// though NewPublicKey takes an interface{}, it must be a pointer to a key.
		pub, err := ssh.NewPublicKey(aK.Public)
		if err != nil {
			// do something
		}
		pubBytes := ssh.MarshalAuthorizedKey(pub)
		ioutil.WriteFile(fmt.Sprintf("./%d/%d.pub", aK.GlyfID, aK.Score), pubBytes, 0777)

		// privb := x
		// for core := 0; core < *cores; core++ {
		// 	// select {
		// 	// case dista[core] <- aK.Render:
		// 	// default:
		// 	// 	fmt.Print("!")
		// 	// }
		// 	dista[core] <- aK.Render
		// }

	}
}

type ArtKey struct {
	Render  Board
	Private ed25519.PrivateKey
	Public  ed25519.PublicKey
	Marshal []byte
	GlyfID  int
	Score   int
}

func worker(starting Board, statsout chan int, resultsout chan ArtKey, workin chan Board) {
	tn := time.Now()
	keysd := 0
	lastBoard := starting

	for {
		pub, priv, mar, nBoard := makeKey()
		keysd++
		s, key := compareBoardScore(nBoard)

		if key != 0 {
			lastBoard = nBoard
			newB := ArtKey{
				Render:  lastBoard,
				Private: priv,
				Public:  pub,
				Marshal: mar,
				GlyfID:  key,
				Score:   s,
			}

			select {
			case resultsout <- newB:
			default:
				fmt.Print(":(")
			}

		}

		if keysd%100 == 0 {
			select {
			case msg := <-workin:
				lastBoard = msg
				continue
			default:
				if time.Since(tn) > time.Second {
					tn = time.Now()
					statsout <- keysd
					keysd = 0
				}
			}

		}
	}
}

type scores struct {
	lol [1000]int
	sync.RWMutex
}

var gs scores

func compareBoardScore(after Board) (score, key int) {
	for k, v := range validBs {
		if v.Skip {
			continue
		}
		diff := 0

		for x, row := range v.Tiles {
			for y, c := range row {
				if c == 0 && after.Tiles[x][y] != 0 {
					diff++
				}

				if c != 0 && after.Tiles[x][y] == 0 {
					diff++
				}
			}
		}
		if validBs[k].Lowestscore > diff {
			return diff, k
		}
	}

	return 100, 0
}

func makeKey() (publicKey ed25519.PublicKey, privateKey ed25519.PrivateKey, marshal []byte, board Board) {
	pub, pk, err := ed25519.GenerateKey(rand.Reader) // make a key

	if err != nil {
		panic(err) // bad practice, but this is a small demo
	}

	spub, _ := ssh.NewPublicKey(pub)  // turn it into a ssh pub key
	spubb := spub.Marshal()           // make it a wire format
	spubbhash := sha256.Sum256(spubb) // get the hash needed

	b := GenerateSubtitled(spubbhash[:], "ED25519 256", "SHA256")
	// fmt.Print(b.String())

	return pub, pk, spubb, b
}

func addLabel(img *image.RGBA, x, y int, label string) {
	col := color.RGBA{0, 0, 0, 255}
	point := fixed.Point26_6{fixed.Int26_6(x * 64), fixed.Int26_6(y * 64)}

	d := &font.Drawer{
		Dst:  img,
		Src:  image.NewUniform(col),
		Face: inconsolata.Bold8x16,
		Dot:  point,
	}
	d.DrawString(label)
}
