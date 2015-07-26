# Clapp

Clapp is a simple tool for writing command line apps using Go. Rather than
being based around flags and subcommands, it takes a similar approach
to writing a web application. You define a series of patterns and handler
functions for the patterns. The handler gets passed a `clapp.Context` which you can use to access the arguments and flags passed into the command.

To install run

```
go get github.com/ccampbell/clapp
```

## Example

This is a sample application to demonstrate some of the functionality

```go

package main

import(
    "github.com/ccampbell/clapp"
    "os"
)

func HelloWorldHandler(c *clapp.Context) {
    c.Print("Hello World!")
}

func InvalidHandler(c *clapp.Context) {
    c.FailWithCode("This failed to run!", 123)
}

func SayNameHandler(c *clapp.Context) {
    c.Print("Say %s", c.Arg("name"))
}

func RunTaskHandler(c *clapp.Context) {
    c.Print("Command is %s", c.Arg("command"))
    c.Print("Run task named: %s", c.Arg("name"))

    // Flags are automatically available even if they are not defined
    if c.Flag("debug") != "" {
        c.Print("Now we are running in debug mode")
    }
}

func main() {
    app := clapp.New("hello")
    app.Version = "1.0"

    // You can overwrite the default title/version and usage using these
    // variables.
    // app.Intro = "Sweet Hello App - version " + app.Version
    // app.Usage = "This is the usage"
    app.HandleFunc("world", HelloWorldHandler, "Print hello world")

    // You can use named variables
    app.HandleFunc("say [name]", SayNameHandler, "Says a person's name")

    // You can use regular expressions
    app.HandleFunc("task command:^(add|remove|restart)$ [name]", RunTaskHandler, "Perform a task on a name (command: add, remove, restart)")
    app.HandleFunc("invalid", InvalidHandler)

    // It is not required to define flags but any that you define will
    // show up in the main usage output.
    //
    // If you want to set a default value then you do have to call DefineFlag
    // and pass the default value as the third parameter.
    app.DefineFlag("--config", "Path to config file", "config.json")
    app.DefineFlag("--verbose", "Show verbose output")

    app.Run(os.Args)
}
```
