package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"

	"github.com/garfunkel/go-tvdb"
	"github.com/kennygrant/sanitize"
)

var showPatterns = []string{`(.*?)[\W\s][sS]?(\d{1,2})[xXeE]?(\d{1,2}).*`}

func digitsCleanup(num int) string {
	str := strconv.Itoa(num)
	if len(str) == 1 {
		return "0" + str
	}
	return str
}

func strCleanupNonWord(str string) string {
	regExp := regexp.MustCompile(`[\W_]+`)
	return strings.Trim(regExp.ReplaceAllString(str, " "), `        `)
}

func renameShow(path string, showName string, season string, episode string,
	episodeName string) (string, error) {
	ext := filepath.Ext(path)
	dir := filepath.Dir(path)
	newName := sanitize.Path(showName + ".S" + season + "E" + episode + "." +
		episodeName + ext)
	newPath := filepath.Join(dir, newName)
	if _, err := os.Stat(newPath); err == nil {
		return newName, errors.New("Destination file already exists.")
	}
	return newName, os.Rename(path, newPath)
}

func ShowInfo(filename string) (string, int, int, error) {
	for _, curPattern := range showPatterns {
		if matched, _ := regexp.MatchString(curPattern, filename); matched {
			showName := ""
			season := 0
			episode := 0
			regExp := regexp.MustCompile(curPattern)
			splitMatch := regExp.FindStringSubmatch(filename)
			showName = strCleanupNonWord(splitMatch[1])
			season, _ = strconv.Atoi(splitMatch[2])
			episode, _ = strconv.Atoi(splitMatch[3])
			return showName, season, episode, nil
		}
	}
	return "", 0, 0, errors.New("No filename pattern match found.")
}

func queryTVDB(showName string) (*tvdb.Series, error) {
	seriesList, err := tvdb.GetSeries(showName)
	if err != nil {
		return nil, err
	}
	if len(seriesList.Series) == 0 {
		return nil, errors.New("No match found.")
	}
	for index, series := range seriesList.Series {
		fmt.Println("\nTVDB Match: ", index+1, " Show: ", series.SeriesName, " ID: ", series.ID)
	}
	fmt.Println("\nSelect match:")
	i := -1
	fmt.Scanf("%d", &i)
	if i-1 > len(seriesList.Series) {
		return nil, errors.New("Invalid selection.")
	}
	selectedSeries := seriesList.Series[i-1]
	if err := selectedSeries.GetDetail(); err != nil {
		return nil, err
	}
	return selectedSeries, nil
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
			if currentSeries, err = queryTVDB(showName); err != nil {
				fmt.Println(err)
				continue
			} else {
				nameToShow[showName] = currentSeries
			}
		} else {
			currentSeries = nameToShow[showName]
		}
		if tvdbSeason, ok := currentSeries.Seasons[uint64(season)]; !ok {
			fmt.Println("Season not found.")
			continue
		} else {
			if len(tvdbSeason) >= (episode - 1) {
				newName, err := renameShow(path, currentSeries.SeriesName,
					digitsCleanup(season),
					digitsCleanup(episode),
					tvdbSeason[episode-1].EpisodeName)
				if err != nil {
					fmt.Println(err)
					continue
				}
				fmt.Println("New Name: ", newName)
			} else {
				fmt.Println("Episode not found.")
				continue
			}
		}
	}
}
