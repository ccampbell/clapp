package clapp

import(
    "fmt"
    "github.com/fatih/color"
    "os"
    "strings"
)

type Context struct {
    App *App
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
        fmt.Println(self.App.Intro)
        return
    }

    fmt.Println(self.App.Name + " v" + self.App.Version)
}

func (self *Context) PrintUsage() {
    if self.App.Usage != "" {
        fmt.Println(self.App.Usage)
        return
    }

    c := color.New(color.Bold, color.Underline)
    c.Println("\nCOMMANDS")

    maxLength1 := getMaxLength(self.App.Commands)
    maxLength2 := getMaxLength(self.App.Flags) + 2
    maxLength := maxLength1
    if maxLength2 > maxLength1 {
        maxLength = maxLength2
    }

    for _, k := range self.App.CommandKeys {
        command := padRight(k, " ", maxLength + 10)
        fmt.Println(command + self.App.Commands[k])
    }

    if len(self.App.FlagKeys) > 0 {
        c.Println("\nFLAGS")
        for _, k := range self.App.FlagKeys {
            flag := padRight("--" + k, " ", maxLength + 10)
            desc := self.App.Flags[k]

            if self.App.FlagDefaults[k] != "" {
                desc += " (default: " + self.App.FlagDefaults[k] + ")"
            }

            fmt.Println(flag + desc)
        }
    }
}

func (self *Context) ShowUsage() {
    self.PrintIntro()
    self.PrintUsage()
}

func (self *Context) ShowVersion() {
    fmt.Println(self.App.Version)
}

func (self *Context) Fail(msg string) {
    error := color.New(color.FgRed, color.Bold).SprintFunc()
    fmt.Printf("%s\n", error(msg))
    os.Exit(1)
}

func (self *Context) FailWithCode(msg string, code int) {
    error := color.New(color.FgRed, color.Bold).SprintFunc()
    fmt.Printf("%s\n", error(msg))
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
    return self.Args[name]
}
