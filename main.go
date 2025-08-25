package main

import (
	"fmt"
	"image"
	"image/png"
	"os"
	"os/exec"
	"strings"
	"syscall"
	"time"
	"unsafe"

	"github.com/getlantern/systray"
	"github.com/go-vgo/robotgo"
	"github.com/kbinani/screenshot"
	"github.com/ncruces/zenity"
)

var last = ""
var started = false
var speakToggle = false

func PlayWAV() {
	winmm := syscall.NewLazyDLL("winmm.dll")
	playSound := winmm.NewProc("PlaySoundW")

	const (
		SND_FILENAME  = 0x00020000
		SND_ASYNC     = 0x0001
		SND_NODEFAULT = 0x0002
	)

	playWavFile := func(filename string) {
		ret, _, _ := playSound.Call(
			uintptr(unsafe.Pointer(syscall.StringToUTF16Ptr(filename))),
			0,
			SND_FILENAME|SND_ASYNC,
		)
		if ret == 0 {
			fmt.Println("Failed to play WAV file")
		}
	}

	playWavFile("media/beep.wav")
}

func GetChatCoordinates() (int, int, int, int) {
	PlayWAV()
	time.Sleep(7 * time.Second)
	x1, y1 := robotgo.Location()

	PlayWAV()
	time.Sleep(7 * time.Second)
	x2, y2 := robotgo.Location()

	if x1 > x2 {
		zenity.Error("Make sure the first set of coordinates are the top left, followed by the bottom right",
			zenity.Title("Coordinate Help"), zenity.ErrorIcon)

		x1, y1, x2, y2 = GetChatCoordinates()
	}

	return x1, y1, x2, y2
}

func CheckWithLastText(t string) bool {
	if speakToggle {
		if t == last {
			return false
		} else {
			return true
		}
	} else {
		return true
	}
}

func SSAndRead(x1 int, y1 int, x2 int, y2 int) {
	img, err := screenshot.CaptureRect(image.Rect(x1, y1, x2, y2))
	if err != nil {
		panic(err)
	}

	os.Remove("image.png")
	f, err := os.Create("image.png")
	if err != nil {
		zenity.Error(err.Error(),
			zenity.Title("Image Creation Error"), zenity.ErrorIcon)
	}
	defer f.Close()

	png.Encode(f, img)

	cmd := exec.Command("tesseract/tesseract.exe", "image.png", "read", "-l", "eng")

	err = cmd.Run()
	if err != nil {
		zenity.Error(err.Error(),
			zenity.Title("Screen Read Error"), zenity.ErrorIcon)
	}

	b_text, err := os.ReadFile("read.txt")
	s_text := string(b_text)

	text := strings.ReplaceAll(s_text, "'", "")

	if err != nil {
		zenity.Error(err.Error(), zenity.Title("File Read Error"), zenity.ErrorIcon)
	}

	if CheckWithLastText(string(text)) {
		fmt.Println(text)

		psScript := fmt.Sprintf(`
		Add-Type -AssemblyName System.Speech;
		$speak = New-Object System.Speech.Synthesis.SpeechSynthesizer;
		$speak.Speak('%s');
		`, text)

		cmd = exec.Command("powershell.exe", "-Command", psScript)
		err = cmd.Run()

		if err != nil {
			zenity.Error(err.Error(),
				zenity.Title("Vocalisation Error"), zenity.ErrorIcon)
		}

		last = string(text)
	}
}

func main() {
	fmt.Println("starting...")
	go func() {
		systray.Run(onReady, onExit)
	}()

	version_curr := "0.0.0"
	fmt.Println("FPSCR Version " + version_curr)

	for {
		if started {
			break
		}
	}

	x1, y1, x2, y2 := GetChatCoordinates()

	for {
		if started {
			SSAndRead(x1, y1, x2, y2)
		} else {
			break
		}
	}
}

func onReady() {
	ico, _ := os.ReadFile("media/16.ico")
	systray.SetIcon(ico)
	systray.SetTitle("FPSCR")
	systray.SetTooltip("FPS Chat Reader")

	mStart := systray.AddMenuItem("Start", "Start the program")
	mQuit := systray.AddMenuItem("Quit", "Quit the program")

	//mSpeakToggle := systray.AddMenuItemCheckbox("Speak only when text changed", "", false)

	go func() {
		for {
			select {
			case <-mStart.ClickedCh:
				started = true
			case <-mQuit.ClickedCh:
				started = false
				os.Exit(0)
				/*case <-mSpeakToggle.ClickedCh:
				speakToggle = mSpeakToggle.Checked()

				fmt.Println("Checkbox was clicked")

				if mSpeakToggle.Checked() {
					mSpeakToggle.SetTitle("✓OK Speak only when text changed")
				} else {
					mSpeakToggle.SetTitle("Speak only when text changed")
				}
				*/
			}
		}
	}()
}

func onExit() {

}
