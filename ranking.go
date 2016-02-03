package main

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"sort"
	"sync"
	"time"
)

type RankingBoard struct {
	List      map[string]RankingItem
	Questions []QuestionConfig
	Start     time.Time
	mu        *sync.Mutex
}

type RankingItemList []RankingItem

type RankingItem struct {
	IpAddress  string
	Name       string
	Score      []int
	TotalScore int
}

func NewRankingBoard(start time.Time, qs []QuestionConfig) *RankingBoard {
	return &RankingBoard{
		List:      make(map[string]RankingItem),
		Start:     start,
		Questions: qs,
		mu:        &sync.Mutex{},
	}
}

func NewRankingBoardFromFile(path string) (*RankingBoard, error) {
	l := &RankingBoard{}
	buf, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(buf, l)
	if err != nil {
		return nil, err
	}
	l.mu = &sync.Mutex{}
	return l, nil
}

func (rb *RankingBoard) Append(ipaddr string, number, score int) (rankup, befirst bool) {
	team := Ip2Team(ipaddr)
	rb.mu.Lock()
	defer rb.mu.Unlock()

	changed := false
	_, ok := rb.List[team]
	oldRank := rb.Rank(team)
	if !ok {
		changed = true
		rb.List[team] = rb.createNewItem(ipaddr)
		rb.List[team].Score[number] = score
	}
	if rb.List[team].Score[number] < score {
		changed = true
		rb.List[team].Score[number] = score
	}

	if changed {
		// TODO: dirty
		m := rb.List[team]
		m.TotalScore = rb.List[team].totalScore()
		rb.List[team] = m

		err := rb.Save("ranking_backup.json")
		if err != nil {
			log.Println(err.Error())
		}

		newRank := rb.Rank(team)
		if oldRank < newRank {
			if newRank == 1 {
				return true, true
			}
			return true, false
		}
	}
	return false, false
}

func (rb *RankingBoard) createNewItem(ipaddr string) RankingItem {
	return RankingItem{
		IpAddress: ipaddr,
		Name:      Ip2Team(ipaddr),
		Score:     make([]int, len(rb.Questions)),
	}
}

func (rb *RankingBoard) Get() []RankingItem {
	list := make(RankingItemList, 0)
	for _, i := range rb.List {
		list = append(list, i)
	}
	sort.Sort(list)
	return list
}

func (rb *RankingBoard) Rank(name string) int {
	for rank, i := range rb.Get() {
		if i.Name == name {
			return rank
		}
	}
	return 0xff
}

func (rb RankingBoard) Save(path string) error {
	buf, err := json.Marshal(rb)
	if err != nil {
		return err
	}
	err = ioutil.WriteFile(path, buf, 0666)
	if err != nil {
		return err
	}
	return nil
}

func (rb RankingBoard) SLA() string {
	list := rb.Get()
	return list[0].IpAddress
}

func (ri RankingItem) totalScore() int {
	s := 0
	for _, v := range ri.Score {
		s += v
	}
	return s
}

func (l RankingItemList) Len() int {
	return len(l)
}

func (l RankingItemList) Less(i, j int) bool {
	return l[i].TotalScore > l[j].TotalScore
}

func (l RankingItemList) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}
