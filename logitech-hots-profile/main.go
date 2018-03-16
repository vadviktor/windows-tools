package main

import (
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/spf13/viper"
)

var (
	fileBaseName string
)

func init() {
	fileBaseName = strings.TrimRight(filepath.Base(os.Args[0]),
		filepath.Ext(os.Args[0]))

	flag.Usage = func() {
		u := fmt.Sprintf(`Scan Logitech profiles and fix up HOTS version paths in the XMLs.

Create a config file named %s.json by filling in what is defined in its sample file.
`, fileBaseName)
		fmt.Fprint(os.Stderr, u)
	}
	flag.Parse()

	viper.SetConfigName(fileBaseName)
	viper.AddConfigPath(".")
	err := viper.ReadInConfig()
	if err != nil {
		log.Fatalf("%s\n", err.Error())
	}
}

//This will update Logitech's Heroes Of The Storm profile with the new game
//version's path.
func main() {
	profileFiles, err := filepath.Glob(viper.GetString("basedir") + `\*.xml`)
	if err != nil {
		log.Fatalln("No profile xml files found in " + viper.GetString("basedir"))
	}

	gameBinaryDir, err := filepath.Glob(viper.GetString("gamedir") + `\Versions\Base*`)
	if err != nil {
		log.Fatalln("No game version found in " + viper.GetString("gamedir"))
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
