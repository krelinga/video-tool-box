package main

import (
    "fmt"
    cli "github.com/urfave/cli/v2"
    "os"
)

func main() {
    fmt.Println("Hello from main!")
    (&cli.App{}).Run(os.Args)
}
