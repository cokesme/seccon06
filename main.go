package main

import (
	"flag"
	"html/template"
	"io/ioutil"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gopkg.in/yaml.v2"
)

var (
	iBreaker *IntervalBreaker
	game     *Game
	config   *Config
	ranking  *RankingBoard

	pathConfig = flag.String("config", "config_example.yaml", "path to config.yaml")
	addr       = flag.String("addr", ":8080", "receive address")
)

type Config struct {
	Questions []QuestionConfig
	Game      struct {
		Start    time.Time
		End      time.Time
		Interval float64
	}
}

type QuestionConfig struct {
	Map  string
	Flag string
	Open int
}

func main() {
	flag.Parse()
	loadConfig(*pathConfig)

	var err error
	game, err = NewGame(config.Game.Start, config.Questions)
	if err != nil {
		panic(err)
	}
	ranking = NewRankingBoard(config.Game.Start, config.Questions)
	iBreaker = NewIntervalBreaker(time.Duration(config.Game.Interval * float64(time.Second)))
	tmpl := template.New("html")
	tmpl = template.Must(tmpl.New("index.html").Parse(tmplIndexHtml))

	r := gin.Default()
	r.SetHTMLTemplate(tmpl)
	r.GET("/", viewIndex).
		GET("/teamflag.txt", viewTeamflag).
		POST("/answer/:number", viewAnswer).
		Static("/css", "css")
	r.Run(*addr)
}

func loadConfig(path string) error {
	b, err := ioutil.ReadFile(path)
	if err != nil {
		return err
	}
	config = &Config{}
	err = yaml.Unmarshal(b, config)
	if err != nil {
		return err
	}
	return nil
}

var tmplIndexHtml = `
<!DOCTYPE html>
<html>
<head>
	<meta http-equiv="X-UA-Compatible" content="IE=edge">
	<meta charset="utf-8">
	<title></title>
	<meta name="description" content="">
	<meta name="author" content="">
	<meta name="viewport" content="width=device-width, initial-scale=1">
	<link rel="stylesheet" href="css/github-markdown.css">
	<!--[if lt IE 9]>
	<script src="//cdnjs.cloudflare.com/ajax/libs/html5shiv/3.7.2/html5shiv.min.js"></script>
	<script src="//cdnjs.cloudflare.com/ajax/libs/respond.js/1.4.2/respond.min.js"></script>
	<![endif]-->
	<link rel="shortcut icon" href="">
</head>
<body><div class="markdown-body" style="width:650px; margin:0 auto; padding:45px;">
	<h1>Find the Image!</h1>
	<h2>Ranking</h2>
	<table style="width:100%;">
		<thead>
			<tr><td rowspan=2>Rank</td><td rowspan=2>Name</td><td colspan=4>SCORE</td></tr>
			<tr><td>image1</td><td>image2</td><td>image3</td><td>total</td></tr>
		</thead>
		<tbody>
			{{range .}}
			<tr><td>1</td><td>{{.Name}}</td><td>{{index .Score 0}}</td><td>{{index .Score 1}}</td><td>{{index .Score 2}}</td><td>{{.TotalScore}}</td></tr>
			{{end}}
		</tbody>
	</table>
	<h2>About this game</h2>
	<ul>
		<li>There are 3 hidden imgaes.</li>
		<li>Please find all complete images.</li>
		<li>Image size is 130 * 130, And dot is black or white(binary image).</li>
		<li>Server will return count of different dots from your sending candidate image.</li>
		<li>You can try only 1 request/sec.</li>
		<li>You will get SLA points while you are staying 1st.</li>
		<li>Server give you a flag if more than 95% is correct.</li>
	</ul>
	<h3>API: POST /answer/:(image number - 1 or 2 or 3)</h3>
	<h4>Request Body</h4>
	<ul>
		<li>Request body is your candidate image.</li>
		<li>Please compress by gzip.</li>
	</ul>
	example:
	<pre>0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000011111000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000011111111111000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000111111111111100000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000011111111111110000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000011111111111111000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000001111111111111000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000001111111111111100000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000001111111111111000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000011111111111111000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000111111111111110000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000001111111111111100000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000011111111111110000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000111111111111100000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000001111111111110000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000011111111111000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000111111111100000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000011111100000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000111111111110000000000000000000000000000000000000000001111111111111100000000000000000000000000000000000
0000000000000000000000000111111111111111110000000000000000000000000000000000001111111111111111111000000000000000000000000000000000
0000000000000000000000000111111111111111111000000000000000000000000000000001111111111111111111111110000000000000000000000000000000
0000000000000000000000000011111111111111111110000000000000000000000000000111111111111111111111111111000000000000000000000000000000
0000000000000000000000000001111111111111111111000000000100000000000000001111111111111111111111111111100000000000000000000000000000
0000000000000000000000000000011111111111111111000000001110000000000111111111000000000000011111111111110000000000000000000000000000
0000000000000000000000000000001111111111111111000000001110000000000011111111100000000000001111111111110000000000000000000000000000
0000000000000000000000000000000011111111111110000000001111000000000001111111110000000000000111111111110000000000000000000000000000
0000000000000000000000000000000000111111111100000000001111100000000000111111111000000000000011111111111000000000000000000000000000
0000000000000000000000000000000000000111100000000000001111100000000000011111111100000000000011111111111000000000000000000000000000
0000000000000000000000000000000000000000000000000000011111110000000000011111111100000000000011111111111000000000000000000000000000
0000000000000000000000000000000000000000000000000000011111111000000000011111111110000000000011111111111000000000000000000000000000
0000000000000000000000000000000000000000000000000000011111111000000000011111111110000000000011111111110000000000000000000000000000
0000000000000000000000000000000000000000000000000000011111111100000000001111111111110000000011111111110000000000000000000000000000
0000000000000000000000000000000000000000000000000000011111111100000000011111111111111100000011111111110000000000000000000000000000
0000000000000000000000000000000000000000000000000000011111111110000011111111111111111110000011111111110000000000000000000000000000
0000000000000000000000000000000000000000000000000000011111111110001111111111111111111110000011111111110000000000000000000000000000
0000000000000000000000000000000000000000000000000000011111111111111111111111111111111110000011111111110000000000000000000000000000
0000000000000000000000000000000000000000000000000000001111111111111111111111111111111110000011111111110000000000000000000000000000
0000000000000000000000000000000000000000000000000000001111111111111111111111111111111100000111111111100000000000000000000000000000
0000000000000000000000000000000000011111111100000000001111111111001111111111111111110000000111111111100000000000000000000000000000
0000000000000000000000000000000001111111111110000000001111111111000001111111111110000000000111111111100000000000000000000000000000
0000000000000000000000000000001111111111111111100000001111111111000000011111111110000000000111111111100000000000000000000000000000
0000000000000000000000000001111111111111111111100000001111111111000000011111111110000000001111111111000000000000000000000000000000
0000000000000000000000001111111111111111111111100000001111111111000000011111111110000000001111111111000000000000000000000000000000
0000000000000000000001111111111111111111111111000000000111111111000000011111111100000000001111111111000000000000000000000000000000
0000000000000000111111111111111111111111111110000000000111111111000000011111111100000000011111111110000000000000000000000000000000
0000000111111111111111111111111111111111111000000000000111111111000000111111111100000000011111111110000000000000000000000000000000
0000000011111111111111111111111111111111110000000000000011111111000000111111111000000000111111111100000000000000000000000000000000
0000000011111111111111111111111111111111100000000000000011111111000000111111111000111111111111111100000000000000000000000000000000
0000000001111111111111111111111111111110000000000000000001111111000001111111111111111111111111111100000000000000000000000000000000
0000000001111111111111111111111111111100000000000000000001111111000001111111111111111111111111111000000000000000000000000000000000
0000000000111111111111111111111111111000000000000000000000111111000111111111111111111111111111111000000000000000000000000000000000
0000000000011111111111111111111111110000000000000000000000011111111111111111111111111111111111110000000000000000000000000000000000
0000000000001111111111111111111111100000000000000000000000001111111111111111111111111111111111110000000000000000000000000000000000
0000000000000011111111111111111111000000000000000000000000001111111111111111111111111011111111100000000000000000000000000000000000
0000000000000000000000011111111110000000000000010000000000000111111111111111111111100000111111000000000000000000000000000000000000
0000000000000000000000111111111100000011000000011000000000000111111111111111111100000000011100000000000000000000000000000000000000
0000000000000000000001111111111000001111111110011000000000000011111111111111100000000000000011100000000000000000000000000000000000
0000000000000000000011111111110000001111111111111000000000000000111111111101111000000000011111111111000000000000000000000000000000
0000000000000000000111111111100000000111111111111100000000000000000111110011111111000000001111111111110000000000000000000000000000
0000000000000000001111111111000000001111111111111100000000000000001111110011111111000000001111111111110000000000000000000000000000
0000000000000000011111111110000000111111111111111100000000000000011111100011111111000000001111111111100000000000000000000000000000
0000000000000000111111111110000111111111111111111110000000000000111111000011111111000000001111111111100000000000000000000000000000
0000000000000001111111111100111111111111111111111110000000000001111111000011111111000000011111111111100000000000000000000000000000
0000000000000011111111111011111111111111111111111111000000000011111110000111111111000000011111111111000000000000000000000000000000
0000000000000111111111111111111111111111111111111111000000000011111110001111111110000000111111111111000000000000000000000000000000
0000000000001111111111111111111111111111110011111111100000000111111100001111111110000001111111111110000000000000000000000000000000
0000000000011111111111111111111111111111100001111111100000001111111000011111111100000011111111111110000000000000000000000000000000
0000000000011111111111111111111111111110000000111111100000011111111000011111111000000111111111111100000000000000000000000000000000
0000000000011111111111111111111111111100000000011111100000111111110000011111111000001111111111111000000000000000000000000000000000
0000000000011111111111111111111111110000000000001000000001111111100000111111110000011111111111111000000001100000000000000000000000
0000000000001111111111111111111110000000000000000000000011111111100000111111110000111111111111110000011111111000000000000000000000
0000000000001111111111111111110000000000000000000000000011111111000001111111110001111111111111100011111111111100000000000000000000
0000000000000111111111111110000000000000000000000000000111111110000001111111100001111111111111111111111111111110000000000000000000
0000000000000111111111110000000000000000000000000000001111111110000001111111100011111111111111111111111111111111000000000000000000
0000000000000011111100000000000000000000000000000000011111111100000001111111100011111111111111111111111111111111100000000000000000
0000000000000000100000000000000000000000000000000000111111111000000011111111100001111111111111111111111111111111100000000000000000
0000000000000000000000000000000000000000000000000001111111111000000011111111100001111111111111111110000111111111110000000000000000
0000000000000000000000000000000000000000000000000011111111110000000011111111000000111111111111000000001111111111100000000000000000
0000000000000000000000000000000000000000000000000111111111100000000011111111000000111111000000000000011111111111000000000000000000
0000000000000000000000000000000000000000000000001111111111000000000011111111000000000000000000000001111111111110000000000000000000
0000000000000000000000000000000000000000000000011111111110000000000011111111000000000000000000000011111111111000000000000000000000
0000000000000000000000000000000000000000000000111111111100000000000011111111000000000000000000000111111111000000000000000000000000
0000000000000000000000000000000000000000000001111111111000000000000011111111000000000000000000011111110000000000011000000000000000
0000000000000000000000000000000000000000110011111111111000000000000011111111000000000000000000111110000000000000011100000000000000
0000000000000000000000000000000000000001111111111111110000000000000001111111100000000000000000000000000000000000011110000000000000
0000000000000000000000000000000000000011111111111111100000000000000001111111100000000000000000000000000000000000011111000000000000
0000000000000000000000000000000000000011111111111111000000000000000001111111100000000000000000000000000000000000011111100000000000
0000000000000000000000000000000000000111111111111110000000000000000000111111110000000000000000000000000000000000011111100000000000
0000000000000000000000000000000000000111111111111100000000000000000000111111110000000000000000000000000000000000011111110000000000
0000000000000000000000000000000000000111111111111000000000000000000000011111111000000000000000000000000000000000111111111000000000
0000000000000000000000000000000000000111111111110000000000000000000000001111111110000000000000000000000000000000111111111100000000
0000000000000000000000000000000000000111111111100000000000000000000000001111111111000000000000000000000000000000111111111100000000
0000000000000000000000000000000000000011111111000000000000000000000000000111111111110000000000000000000000000001111111111110000000
0000000000000000000000000000000000000011111110000000000000000000000000000011111111111100000000000000000000000111111111111110000000
0000000000000000000000000000000000000011111100000000000000000000000000000001111111111111110000000000000001111111111111111111000000
0000000000000000000000000000000000000001110000000000000000000000000000000000111111111111111111111111111111111111111111111111000000
0000000000000000000000000000000000000001000000000000000000000000000000000000011111111111111111111111111111111111111111111111000000
0000000000000000000000000000000000000000000000000000000000000000000000000000001111111111111111111111111111111111111111111111000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000011111111111111111111111111111111111111111111000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000001111111111111111111111111111111111111111111000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000011111111111111111111111111111111111111110000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000001111111111111111111111111111111111111100000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000011111111111111111111111111111111100000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000011111111111111111111111111100000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000011111111111111111110000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000011111000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000
0000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000</pre>
	<h4>Response</h4>
	<ul>
		<li>"wrong" is count of wrong dots.</li>
		<li>"score" is count of correct dots.</li>
		<li>"flag" is a flag for attack point.</li>
	</ul>
	example:
	<pre>{
	"wrong": 5,
	"score": 16895,
	"flag": "SECCON{this is flag}"
}</pre>
	<h2>Hint</h2>
	<ul>
		<li>image1 is a charactor.</li>
		<li>image2 is an animal.</li>
		<li>image3 is some complex charactors.</li>
	</ul>
</div></body>
</html>
`

func Ip2Team(ipaddr string) string {
	for key, name := range map[string]string{
	/*
		// Players IP addresses
		"192.168.1.":  "scryptos",
		"192.168.2.":  "urandom",
		"192.168.3.":  "nw",
		"192.168.4.":  "katagaitai",
		"192.168.5.":  "Jinkai",
		"192.168.6.":  "Nem",
		"192.168.7.":  "Pwnladin",
		"192.168.8.":  "Cykorkinesis",
		"192.168.9.":  "217",
		"192.168.10.": "GoatskiN",
		"192.168.11.": "m1z0r3",
		"192.168.12.": "0x0",
		"192.168.13.": "PwnThyBytes",
		"192.168.14.": "Shellphish",
		"192.168.15.": "CodeRed",
		"192.168.16.": "KaSecon",
		"192.168.17.": "Bushwhackers",
		"192.168.18.": "TomoriNao",
	*/
	} {
		if strings.HasPrefix(ipaddr, key) {
			return name
		}
	}
	return "unknown team"
}
