package mode

import (
	"github.com/pingc0y/URLFinder/queue"
	"sync"
)

type Config struct {
	Proxy      string              `yaml:"proxy"`
	Timeout    int                 `yaml:"timeout"`
	Thread     int                 `yaml:"thread"`
	UrlSteps   int                 `yaml:"urlSteps"`
	JsSteps    int                 `yaml:"jsSteps"`
	Max        int                 `yaml:"max"`
	Headers    map[string]string   `yaml:"headers"`
	JsFind     []string            `yaml:"jsFind"`
	UrlFind    []string            `yaml:"urlFind"`
	InfoFind   map[string][]string `yaml:"infoFiler"`
	Risks      []string            `yaml:"risks"`
	JsFiler    []string            `yaml:"jsFiler"`
	UrlFiler   []string            `yaml:"urlFiler"`
	JsFuzzPath []string            `yaml:"jsFuzzPath"`
}

type Link struct {
	Url      string
	Baseurl  string
	Status   string
	Size     string
	Title    string
	Redirect string
	Source   string
}

type Info struct {
	Phone   []string
	Email   []string
	Baseurl string
	IDcard  []string
	JWT     []string
	Other   []string
	Source  string
}

type Scan struct {
	UrlQueue   *queue.Queue
	Ch         chan []string
	Wg         sync.WaitGroup
	Thread     int
	Output     string
	Proxy      string
	JsResult   []Link
	UrlResult  []Link
	InfoResult []Link
}
