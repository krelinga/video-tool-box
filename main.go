package main

import "log"

func main() {
    if err := appMain(); err != nil {
        log.Fatal(err)
    }
}
