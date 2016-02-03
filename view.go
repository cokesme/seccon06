package main

import (
	//"bufio"
	"compress/gzip"
	"encoding/json"
	//"fmt"
	"github.com/gin-gonic/gin"
	"io/ioutil"
	"net"
	"net/http"
	"strconv"
	"sync"
	"time"
)

func viewIndex(c *gin.Context) {
	list := ranking.Get()
	c.HTML(http.StatusOK, "index.html", list)
	return
}

func viewTeamflag(c *gin.Context) {
	flag := getSLAFlag(ranking.SLA())
	c.String(http.StatusOK, flag)
	return
}

func viewAnswer(c *gin.Context) {
	ipaddr := getIpAddr(c.Request)
	if !iBreaker.Check(ipaddr) {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "request is too many.",
		})
		return
	}

	gzipr, err := gzip.NewReader(c.Request.Body)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "gzip error",
		})
		return
	}
	defer gzipr.Close()
	defer c.Request.Body.Close()

	buf, err := ioutil.ReadAll(gzipr)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "gzip reading error",
		})
		return
	}
	tryMap, err := parseMapString(string(buf), "\r\n")
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": "format error.(" + err.Error() + ")",
		})
		return
	}

	number, err := strconv.Atoi(c.Param("number"))
	number -= 1
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": err.Error(),
		})
		return
	}

	score, wrong, flag, err := game.Try(tryMap, number)
	if err != nil {
		c.JSON(http.StatusBadRequest, map[string]interface{}{
			"message": err.Error(),
		})
		return
	}
	rankup, beFst := ranking.Append(ipaddr, number, score)
	if rankup {
		SendToNirvana(ipaddr, beFst)
	}

	resp := map[string]interface{}{
		"wrong": wrong,
		"score": score,
	}
	if flag != "" {
		resp["flag"] = flag
	}
	c.JSON(http.StatusOK, resp)
}

func getIpAddr(r *http.Request) string {
	tcpaddr, err := net.ResolveTCPAddr("tcp", r.RemoteAddr)
	if err != nil {
		panic(err)
	}
	return tcpaddr.IP.String()
}

func parseJsonInput(r *http.Request) (Map, error) {
	jd := json.NewDecoder(r.Body)

	var req [][]int
	err := jd.Decode(&req)
	if err != nil {
		return nil, err
	}

	rMap := make(Map, 0)
	for _, reqLine := range req {
		rLine := make([]bool, 0)
		for _, reqDot := range reqLine {
			if reqDot == 0 {
				rLine = append(rLine, false)
			} else {
				rLine = append(rLine, true)
			}
		}
		rMap = append(rMap, rLine)
	}
	return rMap, nil
}

type IntervalBreaker struct {
	duration time.Duration
	memo     map[string]time.Time
	mu       *sync.Mutex
}

func NewIntervalBreaker(d time.Duration) *IntervalBreaker {
	return &IntervalBreaker{
		duration: d,
		memo:     make(map[string]time.Time),
		mu:       &sync.Mutex{},
	}
}

func (i *IntervalBreaker) Check(ipaddr string) bool {
	now := time.Now()
	team := Ip2Team(ipaddr)

	i.mu.Lock()
	defer i.mu.Unlock()
	if last, ok := i.memo[team]; ok {
		i.memo[team] = now
		if now.Sub(last) < i.duration {
			return false
		}
		return true
	} else {
		i.memo[team] = now
		return true
	}
}
