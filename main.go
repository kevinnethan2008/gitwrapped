package main 

import (
	"fmt"
	"os/exec"
	"strings"
	"time"
	"sort"
	"strconv"
	"text/template"
	"os"
)

const svgTemplate = `<svg width="400" height="300" xmlns="http://www.w3.org/2000/svg">
  <rect width="400" height="300" fill="#1a1a2e"/>
  <text x="20" y="40" font-size="20" fill="white">Your Git Wrapped</text>
  <text x="20" y="80" font-size="14" fill="#aaa">Total commits: {{.TotalCommits}}</text>
  <text x="20" y="110" font-size="14" fill="#aaa">Busiest hour: {{.BusiestHour}}:00</text>
  <text x="20" y="140" font-size="14" fill="#aaa">Longest streak: {{.LongestStreak}} days</text>
  <text x="20" y="170" font-size="14" fill="#aaa">Most-changed file: {{.TopFile}} ({{.TopFileChurn}} lines)</text>
</svg>`


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

type cardData struct {
	TotalCommits int
	BusiestHour int
	LongestStreak int
	TopFile string
	TopFileChurn int
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

func bucket(Commit []Commit) (int,int) {

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
	BusiestHour := 0
	maxCount := 0
	for hour, count := range hourCount {
		if count > maxCount {
		maxCount = count
		BusiestHour = hour
		}
	}
	return longestStreak, BusiestHour
}

func churn(Output string) (string,int) {

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
	return stats[0].name, stats[0].churn
}

func main() {

	cmd := exec.Command("git", "log", "--pretty=format:%H|%an|%ad|%s")

	output, err := cmd.Output()
	if err != nil {
	fmt.Println("Error running git log:",err)
	return
	}

	longestStreak,busiestHour := bucket(logger(string(output)))

	cmd1 := exec.Command("git", "log", "--numstat", "--pretty=format:COMMIT:%H")
	output1, err1 := cmd1.Output()
	if err1 != nil {
	fmt.Println("Error running numstat")
	return
	}
	name, churn := churn(string(output1))
	
	data := cardData {
		TotalCommits: len(logger(string(output))),
		BusiestHour: busiestHour,
		LongestStreak: longestStreak,
		TopFile: name,
		TopFileChurn: churn,
	}

	tmpl, err := template.New("card").Parse(svgTemplate)
	if err != nil {
	fmt.Println("Template errror:",err)
	return
	}

	file, err := os.Create("wrapped.svg")
	if err != nil {
	fmt.Println("File error", err)
	return
	}
	defer file.Close()

	err = tmpl.Execute(file, data)
	if err != nil {
	fmt.Println("Execute Errror:", err)
	return
	}

	fmt.Println("Wrote wrapped.svg")


}

