package main 

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
	"sort"
	"strconv"
)

type Commit struct {
	Hash string
	Author string
	Date time.Time
	Message string

}

type fileStats struct {
	name string
	churn int
}


func logger(Output string) []Commit{
	lines := strings.Split(strings.TrimSpace(Output), "\n")
	commits := make([]Commit,0,len(lines))

	for _, line := range lines{
	if line == "" {
	continue
	}

	parts := strings.SplitN(line,"|",4)
	if len(parts) != 4 {
	fmt.Println("too many substrings")
	continue
	}

	date, err := time.Parse("Mon Jan 2 15:04:05 2006 -0700", parts[2])
	if err != nil {
		fmt.Println("Date error:", err)
		continue
	}

	commit := Commit {
	Hash: parts[0],
	Author: parts[1],
	Date: date,
	Message: parts[3],
	}
	commits = append(commits,commit)
	fmt.Println(commit)
	fmt.Println(date.Format("2006-01-02 15:04:05"))
     }
	return commits
}

func bucket(Commit []Commit) {

	hourCount := make(map[int]int)
	weekdayCounts := make(map[time.Weekday]int)

	for _, c := range Commit {
	hourCount[c.Date.Hour()]++
	weekdayCounts[c.Date.Weekday()]++
	}
	for hour := 0; hour<24; hour++ {
	count := hourCount[hour]
	bar := strings.Repeat("#",count)
	fmt.Printf("%2d:00 | %s (%d)\n", hour, bar, count)
	}
	weekdays := []time.Weekday{ time.Sunday, time.Monday, time.Tuesday, time.Wednesday, time.Thursday, time.Friday, time.Saturday,}

	for _,day := range weekdays {
	count := weekdayCounts[day]
	bar := strings.Repeat("#",count)
	fmt.Printf("%-9s | %s (%d)\n", day, bar, count)
	}

	dataSet := make(map[string]bool)
	for _,c := range Commit {
	day := c.Date.Format("2006-01-02")
	dataSet[day] = true
	}

	var days []time.Time
	for dayStr := range dataSet {
	t, _ := time.Parse("2006-01-02", dayStr)
	days = append(days,t)
	}

	sort.Slice(days, func(i, j int) bool {
	return days[i].Before(days[j])
	})

	currentStreak := 1
	longestStreak := 1

	for i:=1; i<len(days); i++ {
	diff := days[i].Sub(days[i-1]).Hours() / 24

		if diff == 1 {
	 	currentStreak++
		} else {
		currentStreak = 1
		}
		if currentStreak > longestStreak {
		longestStreak = currentStreak
		}
	}
	fmt.Println("longest Streak:", longestStreak, "days")
}

func churn(Output string) {

fileChurn := make(map[string]int)
lines := strings.Split((Output), "\n")

	for _, line := range lines {

		if line == "" || strings.HasPrefix(line, "COMMIT:") {
			continue
		}

		fields := strings.Split(line, "\t")
		if len(fields) != 3 {
			continue 
		}
	
		added, err1 := strconv.Atoi(fields[0])
		removed, err2 := strconv.Atoi(fields[1])
		if err1 != nil || err2 != nil {
			continue
		}

		fileName := fields[2]
		fileChurn[fileName] += added + removed
	}
var stats []fileStats
	for name, churn := range fileChurn {
		stats = append(stats,fileStats{name,churn})
	}

	sort.Slice(stats, func(i, j int) bool {
	return stats[i].churn > stats[j].churn
	})

	for i:=0; i<5 && i<len(stats); i++ {
	fmt.Printf("%d. %s (%d lines changed)\n", i+1, stats[i].name, stats[i].churn)
	}
}

func main() {

	cmd := exec.Command("git", "log", "--pretty=format:%H|%an|%ad|%s")

	output, err := cmd.Output()
	if err != nil {
	fmt.Println("Error running git log:",err)
	return
	}

	bucket(logger(string(output)))

	cmd1 := exec.Command("git", "log", "--numstat", "--pretty=format:COMMIT:%H")
	output1, err1 := cmd1.Output()
	if err1 != nil {
	fmt.Println("Error running numstat")
	return
	}
	churn(string(output1))

}

