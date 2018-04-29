package main

import (
	"bytes"
	"crypto/rand"
	"crypto/sha256"
	"fmt"

	"golang.org/x/crypto/ed25519"
	"golang.org/x/crypto/ssh"
)

var smooth int

func main() {
	var Tiles [YDim][XDim]int64
	count := 0
	for {
		count++
		_, _, _, nBoard := makeKey()
		for ydim, row := range nBoard.Tiles {
			for xdim, c := range row {
				if c != 0 {
					Tiles[ydim][xdim]++
				}
			}
		}

		if count == 10e9 {
			break

		}
	}
	max, low := 0, 1000000000

	for _, row := range Tiles {
		for _, c := range row {
			cc := int(c)
			if cc > max {
				max = cc
			}
			if low > cc {
				low = cc
			}
		}
	}

	step := (max - low) / 10
	var buf bytes.Buffer

	for _, row := range Tiles {

		buf.WriteString("|")
		for _, c := range row {
			var s string
			// if c == start {
			// 	s = "S"
			// } else if c == end {
			// 	s = "E"
			// } else if int(c) < len(chars) {
			// 	s = chars[c]
			// } else {
			// 	s = chars[len(chars)-1]
			// }

			s = fmt.Sprintf("%d", int(c)/step)
			buf.WriteString(s)
		}
		buf.WriteString("|\n")
	}

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

// func main() {
// 	cores := flag.Int("cores", 4, "cores to use")
// 	smoo := flag.Int("smooth", 20, "diff, lower is harder")
// 	flag.Parse()

// 	smooth = *smoo
// 	fmt.Print("\n")

// 	_, _, _, lastBoard := makeKey()

// 	stats := make(chan int)
// 	results := make(chan ArtKey)
// 	dista := make([]chan Board, *cores)
// 	for core := 0; core < *cores; core++ {
// 		dista[core] = make(chan Board, 1)
// 	}

// 	for core := 0; core < *cores; core++ {
// 		go worker(lastBoard, stats, results, dista[core])
// 	}

// 	go func() {
// 		for {
// 			tn := 0
// 			for core := 0; core < *cores; core++ {
// 				i := <-stats
// 				tn += i
// 			}
// 			fmt.Printf("%d keys per second\n", tn)
// 			tn = 0
// 		}
// 	}()

// 	imgc := 0
// 	for {
// 		aK := <-results
// 		rend := aK.Render.String()

// 		white := color.RGBA{255, 255, 255, 255}
// 		img := image.NewRGBA(image.Rect(0, 0, 200, 200))
// 		draw.Draw(img, img.Bounds(), &image.Uniform{white}, image.ZP, draw.Src)

// 		lines := strings.Split(rend, "\n")
// 		for n, ln := range lines {
// 			addLabel(img, 0, 16+(n*16), ln)
// 		}

// 		f, err := os.Create(fmt.Sprintf("%d.png", imgc))
// 		if err != nil {
// 			panic(err)
// 		}

// 		if err := png.Encode(f, img); err != nil {
// 			panic(err)
// 		}
// 		f.Close()

// 		imgc++
// 		fmt.Print(rend)

// 		for core := 0; core < *cores; core++ {
// 			// select {
// 			// case dista[core] <- aK.Render:
// 			// default:
// 			// 	fmt.Print("!")
// 			// }
// 			dista[core] <- aK.Render
// 		}

// 	}
// }

// type ArtKey struct {
// 	Render  Board
// 	Private ed25519.PrivateKey
// 	Public  ed25519.PublicKey
// 	Marshal []byte
// }

// func worker(starting Board, statsout chan int, resultsout chan ArtKey, workin chan Board) {
// 	tn := time.Now()
// 	keysd := 0
// 	lastBoard := starting

// 	for {
// 		pub, priv, mar, nBoard := makeKey()
// 		keysd++
// 		if compareBoardScore(lastBoard, nBoard) < smooth {
// 			lastBoard = nBoard
// 			newB := ArtKey{
// 				Render:  lastBoard,
// 				Private: priv,
// 				Public:  pub,
// 				Marshal: mar,
// 			}

// 			select {
// 			case resultsout <- newB:
// 			default:
// 				fmt.Print(":(")
// 			}

// 		}

// 		if keysd%100 == 0 {
// 			select {
// 			case msg := <-workin:
// 				lastBoard = msg
// 				continue
// 			default:
// 				if time.Since(tn) > time.Second {
// 					tn = time.Now()
// 					statsout <- keysd
// 					keysd = 0
// 				}
// 			}

// 		}
// 	}
// }

// func compareBoardScore(before, after Board) int {
// 	diff := 0

// 	for x, row := range before.Tiles {
// 		for y, c := range row {
// 			if c == 0 && after.Tiles[x][y] != 0 {
// 				diff++
// 			}

// 			if c != 0 && after.Tiles[x][y] == 0 {
// 				diff++
// 			}
// 		}
// 	}

// 	return diff
// }

// func addLabel(img *image.RGBA, x, y int, label string) {
// 	col := color.RGBA{0, 0, 0, 255}
// 	point := fixed.Point26_6{fixed.Int26_6(x * 64), fixed.Int26_6(y * 64)}

// 	d := &font.Drawer{
// 		Dst:  img,
// 		Src:  image.NewUniform(col),
// 		Face: inconsolata.Bold8x16,
// 		Dot:  point,
// 	}
// 	d.DrawString(label)
// }
