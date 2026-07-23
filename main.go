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

const svgTemplate = `<svg width="400" height="450" xmlns="http://www.w3.org/2000/svg" font-family="Helvetica, Arial, sans-serif">
  <rect width="400" height="450" fill="#1a1a2e" rx="12"/>

  <text x="20" y="30" font-size="20" fill="white">Your Git Wrapped</text>
  <rect x="20" y="38" width="60" height="3" fill="#e94560" rx="1.5"/>

  {{range .Bars}}
  <rect x="{{.X}}" y="{{sub 250 .Height}}" width="10" height="{{.Height}}" fill="#e94560"/>
  {{if eq (mod .Hour 3) 0}}
  <text x="{{.X}}" y="265" font-size="9" fill="#666">{{.Hour}}</text>
  {{end}}
  {{end}}

  <line x1="20" y1="250" x2="380" y2="250" stroke="#444" stroke-width="1"/>

  <text x="20" y="290" font-size="14" fill="#888">Total commits: <tspan fill="#e94560" font-weight="bold">{{.TotalCommits}}</tspan></text>
  <text x="20" y="315" font-size="14" fill="#888">Busiest hour: <tspan fill="#e94560" font-weight="bold">{{.BusiestHour}}:00</tspan></text>
  <text x="20" y="340" font-size="14" fill="#888">Busiest day: <tspan fill="#e94560" font-weight="bold">{{.BusiestDay}}</tspan></text>
  <text x="20" y="365" font-size="14" fill="#888">Longest streak: <tspan fill="#e94560" font-weight="bold">{{.LongestStreak}}</tspan> days</text>
  <text x="20" y="390" font-size="14" fill="#888">Most-changed file: <tspan fill="#e94560" font-weight="bold">{{.TopFile}}</tspan> ({{.TopFileChurn}} lines)</text>
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
	BusiestDay string
	LongestStreak int
	TopFile string
	TopFileChurn int
	Bars []HourBar
}

type HourBar struct {
	Hour int
	Count int
	Height int
	X int
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

func bucket(Commit []Commit) (int,int,string,[]HourBar) {

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

	busiestDay := time.Sunday
	maxDayCount := 0
	for day, count := range weekdayCounts {
	if count > maxDayCount {
			maxDayCount = count
			busiestDay = day
		}
	}
	var bars []HourBar
	for hour :=0; hour<24; hour++ {
		count := hourCount[hour]
		height := 0
		if maxCount > 0 {
			height = int(float64(count) / float64(maxCount) * 100 )
		}
		bars = append(bars, HourBar {
			Hour: hour,
			Count: count,
			Height: height,
			X: 20 + hour*15,
		})
	}
	return longestStreak, BusiestHour, busiestDay.String(), bars
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

	longestStreak,busiestHour,busiestDay,bars := bucket(logger(string(output)))

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
		BusiestDay: busiestDay,
		LongestStreak: longestStreak,
		TopFile: name,
		TopFileChurn: churn,
		Bars: bars,
	}

	funcMap := template.FuncMap{
		"sub": func(a,b int) int { return a - b },
		"mod": func(a,b int) int { return a % b },
	}

	tmpl, err := template.New("card").Funcs(funcMap).Parse(svgTemplate)
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

