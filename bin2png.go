package main

import (
	"bytes"
	"crypto/sha256"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"math"
	"os"
	"strings"
)

// readBytesFromFile reads all the bytes from a file and returns them as a slice.
func readBytesFromFile(file *os.File) ([]byte, error) {
	stat, err := file.Stat()
	if err != nil {
		return nil, err
	}

	bytes := make([]byte, stat.Size())
	_, err = file.Read(bytes)
	if err != nil {
		return nil, err
	}

	return bytes, nil
}

// calculateImageDimensions calculates the width and height of an image based on the number of bytes.
func calculateImageDimensions(numBytes int) (int, int) {
	//width := 1

	width := int(math.Sqrt(float64(numBytes)))
	height := width

	return width, height
}

// saveImageAsJPEG saves an image as a JPEG file with the given name.
func saveImageAsJPEG(filename string, img image.Image) error {
	//JPEG compression alters the colors even with 100% quality...this is
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	err = jpeg.Encode(out, img, &jpeg.Options{Quality: 100})
	if err != nil {
		return err
	}

	return nil
}

func saveImageAsPNG(filename string, img image.Image) error {
	out, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer out.Close()

	err = png.Encode(out, img)
	if err != nil {
		return err
	}

	return nil
}

func writeBytesToFile(filename string, data []byte) error {
	outfile, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer outfile.Close()
	outfile.Write(data)

	return nil
}

func Unpack_Image(jpegfile string) ([]byte, []byte, []byte) {

	var byteArray []byte
	var checksum []byte
	var EOF_Series []byte
	file, err := os.Open(jpegfile)

	defer file.Close()
	//print("DECODED AS:\n")
	decodedImg, _, err := image.Decode(file)
	if err != nil {
		panic(err)
	}
	var Pixel color.NRGBA
	// Convert the decoded image to RGBA model
	rgba := image.NewNRGBA(decodedImg.Bounds())
	for y := decodedImg.Bounds().Min.Y; y < decodedImg.Bounds().Max.Y; y++ {
		for x := decodedImg.Bounds().Min.X; x < decodedImg.Bounds().Max.X; x++ {
			rgba.Set(x, y, decodedImg.At(x, y))
			Pixel = rgba.NRGBAAt(x, y)
			byteArray = append(byteArray, Pixel.R, Pixel.G, Pixel.B)
		}
	}

	checksum = byteArray[:32]
	EOF_Series = byteArray[32:44]
	byteArray = byteArray[44:]

	return checksum, EOF_Series, byteArray

}

func Pack_Binary(bytes []byte, outputfile string) error {
	Color := "Red"
	var red, green, blue uint8 = 0, 0, 0
	//var OldR, OldG, OldB uint8 = 0, 0, 0
	var x, y int = 0, 0
	var Pixel color.NRGBA
	// Create a new RGB image with dimensions based on the number of bytes.
	width, height := calculateImageDimensions(len(bytes))
	img := image.NewNRGBA(image.Rect(0, 0, width, height))

	// Set the pixel values for each byte in the file.
	for _, b := range bytes {
		if x == width {
			y = y + 1
			x = 0
		}
		switch Color {
		case "Red":
			red = uint8(b)
			//OldR = red
			Color = "Green"
		case "Green":
			green = uint8(b)
			//OldG = green
			Color = "Blue"
		case "Blue":
			blue = uint8(b)
			//OldB = blue
			Color = "Red"

			Pixel.R = red
			Pixel.G = green
			Pixel.B = blue
			Pixel.A = 255
			img.SetNRGBA(x, y, Pixel)
			x = x + 1
		}
	}
	//Pixel.R = red
	//Pixel.G = green
	//Pixel.B = blue
	//Pixel.A = 255
	//img.SetNRGBA(x, y, Pixel)

	// Save the RGB image as a JPEG file.
	err := saveImageAsPNG(outputfile, img)
	if err != nil {
		fmt.Printf("Error saving image: %v\n", err)
		return err
	}

	//fmt.Println("Image saved successfully as output.jpg")
	return nil
}

func Recover(bytes2 []byte, AlteredBytes bool) []byte {
	var bytes_to_write []byte
	for i := 0; i < len(bytes2); i++ {
		bytes_to_write = append(bytes_to_write, bytes2[i]) //TODO:Imperfect. FIXME!
	}

	var count int = 0
	for i := len(bytes2) - 1; i > 0; i-- {
		if bytes_to_write[i] == 0xAA {
			count++
			break
		}
		count++
	}

	bytes_to_write = bytes_to_write[:len(bytes_to_write)-count]

	if AlteredBytes {
		bytes_to_write = bytes_to_write[:len(bytes_to_write)-1]
		//print("RECOVED ALTERED-->", len(bytes_to_write))
	}

	return bytes_to_write

}

func EOF(Original_hashing []byte, filename string, bytesO []byte, EOF_Series []byte) []byte {
	var count int = 0
	var Appended bool = false
	var temp_bytes []byte
	fmt.Println(Original_hashing)
	for i := (len(bytesO) - 1); i >= 0; i-- {
		for l := 0; l < (len(EOF_Series) - 1); l++ {
			j := (len(EOF_Series) - 1 - l) //last byte of the series
			if strings.Contains(filename, "txt") {
				j = 0
			}
			//fmt.Println(i)
			if bytesO[i] == EOF_Series[j] { //scanning every byte in the series against the byte.
				//fmt.Println("HERE")
				for k := 0; k < (len(EOF_Series) - 1); k++ {
					if bytesO[i-k] == EOF_Series[j] { //scanning the bytes before bytes[i]
						if i-k == 0 {
							panic("ERROR:Could not find a match!")
						}
						count++
					}
					//fmt.Println(j)
					//if count == (len(EOF_Series) - k - 1) { //j == 0 {
					//fmt.Println("HERE!")

					for m := -12 + j; m < len(EOF_Series[j:]); m++ {
						hasher := sha256.New()
						//fmt.Println("ZERO!")
						temp_bytes = bytesO[:(i - m)]
						temp_bytes = append(temp_bytes, EOF_Series[j:]...)

						hasher.Write(temp_bytes)
						hashSum := hasher.Sum(nil)
						temp_bytes = nil
						//fmt.Println(hashSum)
						if bytes.Equal(Original_hashing, hashSum) {
							fmt.Println("Found!")
							bytesO = bytesO[:(i - m)]
							bytesO = append(bytesO, EOF_Series[j:]...)
							Appended = true
							break
						}
						hasher.Reset()
					}
					if !Appended {
						for m := -12 + j; m < len(EOF_Series[j:]); m++ {
							hasher := sha256.New()
							//fmt.Println("ZERO!")
							temp_bytes = bytesO[:(i + m)]
							temp_bytes = append(temp_bytes, EOF_Series[j:]...)

							hasher.Write(temp_bytes)
							hashSum := hasher.Sum(nil)
							temp_bytes = nil
							//fmt.Println(hashSum)
							if bytes.Equal(Original_hashing, hashSum) {
								fmt.Println("Found!")
								bytesO = bytesO[:(i + m)]
								bytesO = append(bytesO, EOF_Series[j:]...)
								Appended = true
								break
							}
							hasher.Reset()
						}
					}

					//}
					count = 0
					if Appended {
						break
					}
				}
				if Appended {
					break
				}
			}
			if Appended {
				break
			}
		}
		if Appended {
			break
		}
	}
	return bytesO
}

func main() {
	var e string
	var r string
	var bytes2 []byte
	var bytes3 []byte
	//var AlteredBytes bool = false

	flag.StringVar(&e, "e", "", "encode file")
	flag.StringVar(&r, "r", "", "recover file")

	flag.Parse()

	if len(os.Args) != 3 {
		fmt.Println("Usage: bin2png -e/-r <input_binary_file/input_encoded_file>")
		return
	}

	if e != "" {
		inputFile := e
		outputfile := inputFile + ".png"
		unpacked_file := outputfile + ".out"
		var Original_hashing []byte
		var EOF_Series []byte

		// Open the binary file for reading.
		file, err := os.Open(inputFile)
		if err != nil {
			fmt.Printf("Error opening file: %v\n", err)
			return
		}
		defer file.Close()

		// Read all the bytes from the file.
		bytesO, err := readBytesFromFile(file)
		if err != nil {
			fmt.Printf("Error reading bytes: %v\n", err)
			return
		}

		hasher := sha256.New()
		hasher.Write(bytesO)
		hashSum := hasher.Sum(nil)

		//fmt.Println("ORIGNIAL_HASHING =", hashSum)

		/*if (len(bytesO)+len(hashSum))%2 == 0 {
			//add a byte to the original file to match len(bytes)%3
			bytesO = append(bytesO, 0xAA)
			AlteredBytes = true
			//fmt.Printf("ALTERED!!!!-->%d\n", len(bytes))
		}*/
		hashSum = append(hashSum, bytesO[len(bytesO)-12:]...) //12 bytes buffer saved as the markers of the end of the file.
		hashSum = append(hashSum, bytesO...)

		if Pack_Binary(hashSum, outputfile) == nil {
			Original_hashing, EOF_Series, bytes2 = Unpack_Image(outputfile)
		}

		if !bytes.Equal(hashSum[:32], Original_hashing) {
			os.Exit(1)
		}

		bytes2 = EOF(Original_hashing, inputFile, bytes2, EOF_Series) //Original_hashing

		//bytes_to_write := Recover(bytes2, AlteredBytes)

		err = writeBytesToFile(unpacked_file, bytes2)
		if err != nil {
			panic(err)
		}
	}
	if r != "" {
		var err error
		var Original_hashing []byte
		var EOF_Series []byte
		inputFile := r
		Original_hashing, EOF_Series, bytes3 = Unpack_Image(inputFile)
		//fmt.Println("Original_HASHING =", Original_hashing)
		unpacked_file := inputFile + ".out"

		//bytes_to_write := Recover(bytes2, AlteredBytes)
		bytes3 = EOF(Original_hashing, inputFile, bytes3, EOF_Series)

		hasher := sha256.New()
		hasher.Write(bytes3)
		hashSum := hasher.Sum(nil)

		//fmt.Println("Adjusted Modified_HASHING =", hashSum)

		/*err = writeBytesToFile(unpacked_file, bytes3)
		if err != nil {
			panic(err)
		}*/

		if bytes.Equal(Original_hashing, hashSum) {
			fmt.Println("Recovered the file successfully!")
			err = writeBytesToFile(unpacked_file, bytes3)
			if err != nil {
				panic(err)
			}
		} /*else {
			hasher2 := sha256.New()
			for i := 1; i < len(bytes3); i++ {

				hasher2.Write(bytes3[:len(bytes3)-i])
				hashSum2 := hasher2.Sum(nil)

				fmt.Println("Modified_HASHING =", hashSum2)

				if bytes.Equal(Original_hashing, hashSum2) {
					err = writeBytesToFile(unpacked_file, bytes3[:len(bytes3)-i])
					if err != nil {
						panic(err)
					}
				}
				hashSum2 = nil
				hasher2.Reset()
			}
		}*/

	}

}
