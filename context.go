package clapp

import(
    "fmt"
    "github.com/mgutz/ansi"
    "os"
    "strings"
    "sync"
    "time"
)

type Context struct {
    Mu sync.RWMutex
    App *App
    spinnerChannel chan bool
    progressBarChannel chan []string
    ProgressBar *ProgressBar
    Flags map[string]string
    Args map[string]string
}

func padRight(str, pad string, length int) string {
    for {
        str += pad
        if len(str) > length {
            return str[0:length]
        }
    }
}

func getMaxLength(m map[string]string) int {
    maxLength := 0
    for k, _ := range m {
        if len(k) > maxLength {
            maxLength = len(k)
        }
    }
    return maxLength
}

func (self *Context) PrintIntro() {
    if self.App.Intro != "" {
        self.Print(self.App.Intro)
        return
    }

    self.Print("%s v%s", self.App.Name, self.App.Version)
}

func (self *Context) PrintUsage() {
    if self.App.Usage != "" {
        self.Print(self.App.Usage)
        return
    }

    boldUnderline := ansi.ColorFunc("white+bu")
    self.Print("\n%s", boldUnderline("COMMANDS"))

    maxLength1 := getMaxLength(self.App.Commands)
    maxLength2 := getMaxLength(self.App.Flags) + 2
    maxLength := maxLength1
    if maxLength2 > maxLength1 {
        maxLength = maxLength2
    }

    for _, k := range self.App.CommandKeys {
        command := padRight(k, " ", maxLength + 3)
        self.Print(command + self.App.Commands[k])
    }

    if len(self.App.FlagKeys) > 0 {
        self.Print("\n%s", boldUnderline("FLAGS"))
        for _, k := range self.App.FlagKeys {
            flag := padRight("--" + k, " ", maxLength + 3)
            desc := self.App.Flags[k]

            if self.App.FlagDefaults[k] != "" {
                desc += " (default: " + self.App.FlagDefaults[k] + ")"
            }

            self.Print(flag + desc)
        }
    }
}

func (self *Context) StartSpinner(text...string) {
    self.spinnerChannel = make(chan bool)
    go func() {
        glyphs := [8]string{"|", "/", "-", "\\", "|", "/", "-", "\\"}
        for {
            select {
                case <- self.spinnerChannel:
                    self.PrintInline("\r")
                    if len(text) > 0 {
                        self.PrintInline(strings.Repeat(" ", len(text[0]) + 2))
                    }
                    self.PrintInline("\r")
                    return

                default:
                    for _, glyph := range glyphs {
                        msg := fmt.Sprintf("%s", glyph)
                        if len(text) > 0 {
                            msg = fmt.Sprintf("\r%s %s ", text[0], glyph)
                        }
                        self.PrintInline(msg)
                        time.Sleep(150 * time.Millisecond)
                    }
            }
        }
    } ()
}

func (self *Context) StopSpinner() {
    self.spinnerChannel <- true
    close(self.spinnerChannel)
}

func (self *Context) StartProgress() {
    self.ProgressBar.Init(self)
    self.ProgressBar.Render(self)
}

func (self *Context) SetProgressPercent(percent interface{}) {
    p := float64(0)
    switch percent.(type) {
        case float64:
            p = percent.(float64)
            break
        case float32:
            p = float64(percent.(float32))
            break
        case int:
            p = float64(percent.(int))
            break
    }

    self.ProgressBar.Mu.Lock()
    self.ProgressBar.CurrentPercent = p
    self.ProgressBar.Mu.Unlock()
    self.ProgressBar.Render(self)
}

func (self *Context) CancelProgress() {
    self.ProgressBar.Cancel(self)
}

func (self *Context) StopProgress() {
    self.ProgressBar.Stop(self)
}

func (self *Context) ShowUsage() {
    self.PrintIntro()
    if self.App.Description != "" {
        self.Print("\n%s", self.App.Description)
    }
    self.PrintUsage()
}

func (self *Context) ShowUsageWithMessage(m string) {
    self.PrintIntro()
    error := ansi.ColorFunc("88")
    self.Print("\n%s", error(m))
    self.PrintUsage()
}

func (self *Context) ShowVersion() {
    self.Print(self.App.Version)
}

func output(forceLineBreak bool, messages...interface{}) {
    if len(messages) == 0 {
        return
    }

    first := messages[0].(string)

    if forceLineBreak {
        first += "\n"
    }

    rest := make([]interface{}, 0)

    if len(messages) > 1 {
        rest = messages[1:len(messages)]
    }

    fmt.Printf(first, rest...)
}

func (self *Context) Print(messages...interface{}) {
    forceLineBreak := true
    output(forceLineBreak, messages...)
}

func (self *Context) PrintInline(messages...interface{}) {
    forceLineBreak := false
    output(forceLineBreak, messages...)
}

func (self *Context) Fail(msg string) {
    error := ansi.ColorFunc("88")
    self.Print("%s", error(msg))
    os.Exit(1)
}

func (self *Context) FailWithCode(msg string, code int) {
    error := ansi.ColorFunc("88")
    self.Print("%s", error(msg))
    os.Exit(code)
}

func (self *Context) Flag(name string) string {
    if strings.HasPrefix(name, "-") {
        name = stripDashes(name)
    }

    val := self.Flags[name]

    if val == "" && self.App.FlagDefaults[name] != "" {
        return self.App.FlagDefaults[name]
    }

    return self.Flags[name]
}

func (self *Context) Arg(name string) string {
    val, ok := self.Args[name]

    if !ok && strings.Contains(name, " ") {
        name = strings.Replace(name, " ", "_", -1)
        val, ok = self.Args[name]
    }

    if !ok {
        return ""
    }

    return val
}
