package main

import (
	"fmt"
	"strings"
	"time"
)

type Map [][]bool

type Game struct {
	list  []Question
	start time.Time
}

type Question struct {
	hMap     Map
	openTime time.Duration
	flag     string
}

func NewQuestion(strMap string, flag string, openTime int) (Question, error) {
	hMap, err := parseMapString(strMap, " ")
	if err != nil {
		return Question{}, nil
	}

	return Question{
		hMap:     hMap,
		openTime: time.Duration(openTime) * time.Second,
		flag:     flag,
	}, nil
}

func (q Question) IsOpen(start, when time.Time) bool {
	if start.Add(q.openTime).Before(when) {
		return true
	}
	return false
}

func (q Question) Try(answer Map) (int, int, string, error) {
	worngs := 0
	if len(q.hMap) != len(answer) || len(answer) == 0 {
		return 0, 0, "", fmt.Errorf("invalid image size")
	}
	if len(q.hMap[0]) != len(answer[0]) {
		return 0, 0, "", fmt.Errorf("invalid image size")
	}
	for i := range answer {
		for j := range answer[i] {
			if len(q.hMap) <= i {
				return 0, 0, "", fmt.Errorf("invalid image size")
			}
			if len(q.hMap[i]) <= j {
				return 0, 0, "", fmt.Errorf("invalid image size")
			}
			if answer[i][j] != q.hMap[i][j] {
				worngs += 1
			}
		}
	}
	height := len(q.hMap)
	width := len(q.hMap[0])
	if q.CanGetFlag(worngs) {
		return height*width - worngs, worngs, q.flag, nil
	}
	return height*width - worngs, worngs, "", nil
}

func (q Question) CanGetFlag(worngs int) bool {
	height := len(q.hMap)
	width := len(q.hMap[0])
	if float64(worngs)/float64(height*width) < 0.1 {
		return true
	}
	return false
}

func NewGame(start time.Time, questions []QuestionConfig) (*Game, error) {
	g := &Game{
		list:  []Question{},
		start: start,
	}
	for _, m := range questions {
		q, err := NewQuestion(m.Map, m.Flag, m.Open)
		if err != nil {
			return nil, err
		}
		g.list = append(g.list, q)
	}
	return g, nil
}

func (g *Game) IsOpen(number int) bool {
	if number < 0 ||
		number >= len(g.list) {
		return false
	}
	if !g.list[number].IsOpen(g.start, time.Now()) {
		return false
	}
	return true
}

func (g *Game) Try(answer Map, number int) (int, int, string, error) {
	if !g.IsOpen(number) {
		return 0, 0x8fffffff, "", fmt.Errorf("invalid number")
	}
	return g.list[number].Try(answer)
}

func parseMapString(m string, sep string) (Map, error) {
	lines := strings.Split(m, sep)
	width := 0
	println(len(lines), len(lines[0]))

	newMap := make(Map, 0)
	for lNumber, line := range lines {
		if width != len(line) && width != 0 {
			return nil, fmt.Errorf("%d charactors expected but got %d at line %d", width, len(line), lNumber)
		}
		width = len(line)
		l := make([]bool, width)
		for i, dot := range line {
			l[i] = (dot != '0')
		}
		newMap = append(newMap, l)
	}
	return newMap, nil
}
