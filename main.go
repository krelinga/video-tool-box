package main

import (
    "log"
    "os"
)

func main() {
    if err := appMain(os.Args); err != nil {
        log.Fatal(err)
    }
}
