package clapp

import(
    "fmt"
    "github.com/mgutz/ansi"
    "strings"
    "sync"
    "time"
)

type ProgressBar struct {
    Mu sync.RWMutex
    Width int
    Duration time.Duration
    Type string
    PreviousPercent float64
    CurrentPercent float64
    EmptyShape string
    FillShape string
    FillColor string
    BackgroundColor string
    doneChannel chan bool
    renderingChannel chan bool
}

func (self *ProgressBar) GetLineForPercent(i int, displayPercent float64) string {
    line := ansi.Color("[", self.BackgroundColor)
    line += ansi.Color(strings.Repeat(self.FillShape, i), self.FillColor)
    line += ansi.Color(strings.Repeat(self.EmptyShape, self.Width - i), self.BackgroundColor)
    line += ansi.Color("] ", self.BackgroundColor) + strings.Replace(fmt.Sprintf("%.01f", displayPercent), ".0", "", -1)
    line += "%% "
    return line
}

func (self *ProgressBar) GetBlockCountForPercent(p float64) int {
    percentagesPerBlock := float64(100) / float64(self.Width)
    blocks := p / percentagesPerBlock
    return int(blocks)
}

func (self *ProgressBar) GetLinesForPercentRange(start float64, end float64) []string {
    var lines []string
    startBlockCount := self.GetBlockCountForPercent(start)
    endBlockCount := self.GetBlockCountForPercent(end)
    i := startBlockCount
    for i <= endBlockCount {
        percentToUse := start
        if i == endBlockCount {
            percentToUse = end
        }

        line := self.GetLineForPercent(i, percentToUse)
        lines = append(lines, line)
        i += 1
    }

    return lines
}

func easeIn(t float64, b float64, c float64, d float64) float64 {
    t = t / d
    return c*t*t + b
}

func easeOut(t float64, b float64, c float64, d float64) float64 {
    t = t / d
    return -c *t*(t-2) + b
}

func (self *ProgressBar) Init(c *Context) {
    c.Mu.Lock()
    c.progressBarChannel = make(chan []string)
    self.renderingChannel = make(chan bool, 1)
    c.Mu.Unlock()
    go func() {
        previousLine := ""
        var fn func(float64, float64, float64, float64) float64

        for {
            c.Mu.RLock()
            ch := c.progressBarChannel
            c.Mu.RUnlock()
            select {
                case lines := <- ch:
                    for i, l := range lines {
                        if l == "CANCEL" {
                            c.Print("")
                            <- self.renderingChannel
                            close(ch)
                            return
                        }

                        if l == "DONE" {
                            c.Print("")
                            close(c.progressBarChannel)
                            self.doneChannel <- true
                            <- self.renderingChannel
                            return
                        }

                        c.PrintInline("\r" + strings.Repeat(" ", len(previousLine)))
                        c.PrintInline("\r" + l)
                        previousLine = l

                        // Counterintuitive but since we are using sleep times to do the animations the function should be the opposite
                        switch self.Type {
                            case "ease-in":
                                fn = easeOut
                                break
                            case "ease-out":
                                fn = easeIn
                                break
                        }

                        if fn != nil {
                            dur := fn(float64(i + 1), float64(0), float64(self.Duration), float64(len(lines)))
                            prevDur := fn(float64(i), float64(0), float64(self.Duration), float64(len(lines)))
                            time.Sleep(time.Duration(dur) - time.Duration(prevDur))
                            continue
                        }

                        time.Sleep(self.Duration / time.Duration(len(lines)))
                    }
                    <- self.renderingChannel
                    break
            }
        }
    } ()
}

func (self *ProgressBar) Render(c *Context) {
    self.renderingChannel <- true
    lines := self.GetLinesForPercentRange(self.PreviousPercent, self.CurrentPercent)
    c.progressBarChannel <- lines
    self.Mu.Lock()
    self.PreviousPercent = self.CurrentPercent
    self.Mu.Unlock()
}

func (self *ProgressBar) Cancel(c *Context) {
    c.progressBarChannel <- []string{ "CANCEL" }
}

func (self *ProgressBar) Stop(c *Context) {
    self.Mu.Lock()
    self.doneChannel = make(chan bool, 1)
    self.Mu.Unlock()
    c.progressBarChannel <- []string{ "DONE" }
    <- self.doneChannel
    close(self.doneChannel)
}
