package clapp

import(
    "errors"
    "os"
    "regexp"
    "strings"
    "time"
)

type App struct {
    Name string
    Description string
    Version string
    Intro string
    Usage string
    HandlerKeys []string
    Handlers map[string]func(*Context)
    CommandKeys []string
    Commands map[string]string
    FlagKeys []string
    Flags map[string]string
    FlagDefaults map[string]string
    Aliases map[string]string
}

func stripDashes(arg string) string {
    return strings.TrimLeft(arg, "-")
}

func ParseFlags(args []string, aliases map[string]string) map[string]string {
    var results map[string]string
    var last string

    results = make(map[string]string)

    for _, arg := range args {
        if strings.HasPrefix(arg, "-") && strings.Contains(arg, "=") {
            arg = stripDashes(arg)
            bits := strings.Split(arg, "=")
            results[bits[0]] = bits[1]
            continue
        }

        if strings.HasPrefix(arg, "-") {
            if val, ok := aliases[arg]; ok {
                arg = val
            }

            arg = stripDashes(arg)
            last = arg
            results[arg] = "1"
            continue
        }

        if last != "" {
            results[last] = arg
            last = ""
            continue
        }
    }

    return results
}

func New(n string) *App {
    app := &App{
        Name: n,
        Handlers: make(map[string]func(*Context)),
        Commands: make(map[string]string),
        Flags: make(map[string]string),
        FlagDefaults: make(map[string]string),
        Aliases: make(map[string]string),
    }
    return app
}

func patternToWords(pattern string) []string {
    re := regexp.MustCompile("\\[.*?\\]")
    pattern = string(re.ReplaceAllFunc([]byte(pattern), func(b []byte) []byte {
        newWord := strings.Replace(string(b), " ", "_", -1)
        return []byte(newWord)
    }))

    return strings.Split(pattern, " ")
}

func handlerMatches(pattern string, args []string) (map[string]string, bool) {
    var matches = make(map[string]string)
    var ok = true

    words := patternToWords(pattern)
    newArgs := make([]string, 0)
    for i, arg := range args {

        // Do not include the first argument or any flags
        if i > 0 && !strings.HasPrefix(arg, "-") {
            newArgs = append(newArgs, arg)
        }
    }

    if len(words) != len(newArgs) {
        return matches, false
    }

    for i, arg := range newArgs {
        word := words[i]
        if strings.HasPrefix(word, "[") && strings.HasSuffix(word, "]") {
            matches[strings.TrimRight(strings.TrimLeft(word, "["), "]")] = arg
            continue
        }

        if strings.Contains(word, ":") {
            bits := strings.Split(word, ":")
            re, err := regexp.Compile(bits[1])

            if err == nil && re.MatchString(arg) {
                matches[bits[0]] = arg
                continue
            }
        }

        if word != arg {
            ok = false
            break
        }
    }

    return matches, ok
}

func findMatchingHandler(a *App, args []string, c *Context) (func(*Context), error) {
    var f func(*Context)

    for _, k := range a.HandlerKeys {
        matches, ok := handlerMatches(k, args)
        if ok {
            c.Args = matches
            return a.Handlers[k], nil
        }
    }

    return f, errors.New("Unable to find matching handler")
}

func prettyPattern(p string) string {
    finalWords := make([]string, 0)
    words := strings.Split(p, " ")

    for _, v := range words {
        finalWord := v
        if strings.Contains(v, ":") {
            finalWord = "{" + strings.Split(v, ":")[0] + "}"
        }
        finalWords = append(finalWords, finalWord)
    }

    return strings.Join(finalWords, " ")
}

func (self *App) HandleFunc(pattern string, handler func(*Context), desc...string) {
    self.HandlerKeys = append(self.HandlerKeys, pattern)
    self.Handlers[pattern] = handler

    if len(desc) == 1 {
        self.CommandKeys = append(self.CommandKeys, prettyPattern(pattern))
        self.Commands[prettyPattern(pattern)] = desc[0]
    }
}

func (self *App) DefineFlag(flag...string) {
    if len(flag) < 2 {
        return
    }

    flagName := stripDashes(flag[0])

    if flag[1] != "" {
        self.FlagKeys = append(self.FlagKeys, flagName)
        self.Flags[flagName] = flag[1]
    }

    if len(flag) > 2 {
        self.FlagDefaults[flagName] = flag[2]
    }
}

// Adds a short flag as an alias for a larger flag
//
// For example:
//
// ```
// app.DefineFlag("--verbose", "Show verbose output")
// app.AddAlias("-v", "--verbose")
// ```
func (self *App) AddAlias(alias, flag string) {
    self.Aliases[alias] = stripDashes(flag)
}

func (self *App) Run(args []string) {
    f := ParseFlags(args, self.Aliases)
    p := ProgressBar {
        Width: 50,
        Duration: 500 * time.Millisecond,
        Type: "linear",
        EmptyShape: "-",
        FillShape: "#",
        FillColor: "white",
        BackgroundColor: "white",
    }

    c := Context {
        Flags: f,
        App: self,
        ProgressBar: &p,
    }

    if c.Flag("h") == "1" || c.Flag("help") == "1" {
        c.ShowUsage()
        return
    }

    if c.Flag("version") == "1" {
        c.ShowVersion()
        return
    }

    handler, err := findMatchingHandler(self, args, &c)
    if err == nil {
        handler(&c)
        return
    }

    if len(args) > 1 {
        c.ShowUsageWithMessage("“" + strings.Join(args, " ") + "”" + " is not a valid command. Make sure you typed it correctly.")
        os.Exit(1)
    }

    c.ShowUsage()
}
