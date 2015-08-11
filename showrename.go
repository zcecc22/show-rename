package main

import (
	"errors"
	"fmt"
	"github.com/garfunkel/go-tvdb"
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
	return "", 0, 0, errors.New("No match returned.")
}

func main() {
	nameToShow := make(map[string]*tvdb.Series)
	var currentSeries *tvdb.Series
	showName, season, episode, err := ShowInfo("The.Big.Bang.Theory.Season 1 Episode 2.Release.1999.mp4")

	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Println(showName, " ", season, " ", episode, " ")

	seriesList, err := tvdb.GetSeries(showName)

	if err != nil {
		fmt.Println(err)
		return
	}

	if _, ok := nameToShow[showName]; !ok {
		for index, series := range seriesList.Series {
			fmt.Println("Match: ", index+1, " ", series.ID, " ", series.SeriesName)
		}
		fmt.Println("\nSelect match.")
		i := -1
		fmt.Scanf("%d", &i)
		currentSeries = seriesList.Series[i-1]
		nameToShow[showName] = currentSeries
		if err := currentSeries.GetDetail(); err != nil {
			fmt.Println(err)
			return
		}
	} else {
		currentSeries = nameToShow[showName]
	}

	fmt.Println(currentSeries.SeriesName)
	fmt.Println(digitsCleanup(season))
	fmt.Println(digitsCleanup(episode))
	fmt.Println(currentSeries.Seasons[uint64(season)][episode-1].EpisodeName)
	fmt.Println(currentSeries.Seasons[uint64(season)][episode-1].EpisodeNumber)

}
