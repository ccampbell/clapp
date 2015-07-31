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
    PreviousPercent float64
    CurrentPercent float64
    EmptyShape string
    FillShape string
    FillColor string
    BackgroundColor string
    doneChannel chan bool
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

func (self *ProgressBar) Init(c *Context) {
    c.Mu.Lock()
    c.progressBarChannel = make(chan []string)
    c.Mu.Unlock()
    go func() {
        linesToAdd := make([]string, 0)
        previousLine := ""
        for {
            c.Mu.RLock()
            ch := c.progressBarChannel
            c.Mu.RUnlock()
            select {
                case lines := <- ch:
                    for _, l := range lines {
                        if l == "CANCEL" {
                            c.Print("")
                            close(ch)
                            return
                        }

                        linesToAdd = append(linesToAdd, l)
                    }
                    break
                default:
                    if len(linesToAdd) == 0 {
                        continue
                    }

                    nextLine := linesToAdd[0]
                    if nextLine == "DONE" {
                        c.Print("")
                        close(c.progressBarChannel)
                        self.doneChannel <- true
                        return
                    }

                    linesToAdd = linesToAdd[1:]
                    c.PrintInline("\r" + strings.Repeat(" ", len(previousLine)))
                    c.PrintInline("\r" + nextLine)
                    previousLine = nextLine
                    time.Sleep(50 * time.Millisecond)
                    break
            }
        }
    } ()
}

func (self *ProgressBar) Render(c *Context) {
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
