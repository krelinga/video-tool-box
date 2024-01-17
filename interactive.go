package main

import (
    "bufio"
    "fmt"
    "os"
    "strings"
    "unicode/utf8"
)

type interChoice struct {
    Text    string
    Key     rune
    Fn      func() error
}

func interPrompt(choices []*interChoice) error {
    keys := make(map[rune]*interChoice)
    for _, choice := range choices {
        if _, found := keys[choice.Key]; found {
            panic("duplicate Key in choices")
        }
        keys[choice.Key] = choice
    }
    prompt := func() string {
        parts := make([]string, len(choices), len(choices))
        for i, choice := range choices {
            parts[i] = choice.Text
        }
        return strings.Join(parts, ", ")
    }()
    scanner := bufio.NewScanner(os.Stdin)
    input := rune(0)
    for func() bool { _, found := keys[input]; return !found }() {
        fmt.Println(prompt)
        if !scanner.Scan() {
            return scanner.Err()
        }
        found := scanner.Text()
        if utf8.RuneCountInString(found) != 1 {
            // Display the prompt again and read more input
            continue
        }
        input, _ = utf8.DecodeRuneInString(found)
    }

    return keys[input].Fn()
}
