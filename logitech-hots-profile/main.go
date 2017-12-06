package main

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"
	"regexp"
)

const BASEDIR = `C:\Users\ikon\AppData\Local\Logitech\Logitech Gaming Software\profiles`
const GAMEDIR = `E:\Heroes of the Storm`

//This will update Logitech's Heroes Of The Storm profile with the new game
//version's path.
func main() {
	profileFiles, err := filepath.Glob(BASEDIR + `\*.xml`)
	if err != nil {
		log.Fatalln("No profile xml files found in " + BASEDIR)
	}

	gameBinaryDir, err := filepath.Glob(GAMEDIR + `\Versions\Base*`)
	if err != nil {
		log.Fatalln("No game version found in " + GAMEDIR)
	}

	re := regexp.MustCompile(`(?i)base(\d+)`)
	currentVersion := re.FindStringSubmatch(gameBinaryDir[0])[1]

	for _, f := range profileFiles {
		dat, err := ioutil.ReadFile(f)
		if err != nil {
			log.Printf("ERROR! Can't read %s: %s\n", f, err.Error())
		}

		re := regexp.MustCompile(`(?i)versions\\base\d+\\`)
		newDat := re.ReplaceAllString(string(dat),
			fmt.Sprintf(`versions\base%s\`, currentVersion))

		err = ioutil.WriteFile(f, []byte(newDat), 0644)
		if err != nil {
			log.Printf("ERROR! Can't write %s: %s\n", f, err.Error())
		}
	}

	fmt.Println("Done")
	fmt.Scanln()
}
