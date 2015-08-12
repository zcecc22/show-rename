package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/garfunkel/go-tvdb"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

var showPatterns = []string{`(?i)S(\d{1,2})E(\d{1,2})`, `(?i)(\d{1,2})X(\d{1,2})`,
	`(?i)Season (\d{1,2}) Episode (\d{1,2})`, `(?i)(\d{1})(\d{2})`}

func digitsCleanup(num int) string {
	str := strconv.Itoa(num)
	if len(str) == 1 {
		return "0" + str
	}
	return str
}

func strCleanupNonWord(str string) string {
	regExp := regexp.MustCompile(`\W+`)
	return strings.Trim(regExp.ReplaceAllString(str, " "), ` 	`)
}

func strCleanupSymbols(str string) string {
	regExp := regexp.MustCompile(`[\\/$&?\*]`)
	return regExp.ReplaceAllString(str, "_")
}

func renameShow(path string, showName string, season string, episode string,
	episodeName string) (string, error) {
	ext := filepath.Ext(path)
	dir := filepath.Dir(path)
	newName := strCleanupSymbols(showName + ".S" + season + "E" + episode + "." +
		episodeName + ext)
	newPath := filepath.Join(dir, newName)

	if _, err := os.Stat(newPath); err == nil {
		return newName, errors.New("Destination file already exists.")
	}

	return newName, os.Rename(path, newPath)
}

func ShowInfo(filename string) (string, int, int, error) {
	for _, curPattern := range showPatterns {
		if matched, err := regexp.MatchString(curPattern, filename); err == nil && matched {
			showName := ""
			season := 0
			episode := 0
			err = nil

			regExp := regexp.MustCompile(curPattern)
			showName = strCleanupNonWord(regExp.Split(filename, -1)[0])
			seasonEpisode := regExp.FindStringSubmatch(filename)[1:3]

			season, err = strconv.Atoi(seasonEpisode[0])
			episode, err = strconv.Atoi(seasonEpisode[1])

			return showName, season, episode, err
		} else if err != nil {
			return "", 0, 0, err
		}
	}
	return "", 0, 0, errors.New("No filename pattern match found.")
}

func main() {
	flag.Parse()
	nameToShow := make(map[string]*tvdb.Series)
	var currentSeries *tvdb.Series

	for _, path := range flag.Args() {
		fmt.Println("\nCurrent Name: ", filepath.Base(path))
		showName, season, episode, err := ShowInfo(filepath.Base(path))

		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println("Show: ", showName, " Season: ", season, " Episode: ", episode)

		if _, ok := nameToShow[showName]; !ok {
			seriesList, err := tvdb.GetSeries(showName)

			if err != nil {
				fmt.Println(err)
				continue
			}

			if len(seriesList.Series) == 0 {
				fmt.Println(errors.New("No match found."))
				continue
			}

			for index, series := range seriesList.Series {
				fmt.Println("\nTVDB Match: ", index+1, " Show: ", series.SeriesName, " ID: ", series.ID)
			}

			fmt.Println("\nSelect match:")
			i := -1
			fmt.Scanf("%d", &i)
			currentSeries = seriesList.Series[i-1]
			nameToShow[showName] = currentSeries
			if err := currentSeries.GetDetail(); err != nil {
				fmt.Println(err)
				continue
			}
		} else {
			currentSeries = nameToShow[showName]
		}

		newName, err := renameShow(path, currentSeries.SeriesName,
			digitsCleanup(season),
			digitsCleanup(episode),
			currentSeries.Seasons[uint64(season)][episode-1].EpisodeName)

		if err != nil {
			fmt.Println(err)
			continue
		}

		fmt.Println("New Name: ", newName)
	}
}
