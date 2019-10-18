// /\\//\\//\\//\\//\\//\\//\\//\\ //
// Look for the top marker ^^^^^   //
// without a new-line              //
// and the bottom marker vvvvvv    //
// /\\//\\//\\//\\//\\//\\//\\//\\ //
// including the return/new-line   //

package main

import (
    "bufio"
    "fmt"
    "log"
    "os"
    "path/filepath"
)

type error interface {
    Error() string
}
type ProjectFiles struct {
    FileName []string
}

func main() {
    file, err := os.Open("citation_text.txt")
    if err != nil {
        log.Fatal(err)
    }
    defer file.Close()

    var s []string
    scanner := bufio.NewScanner(file)
    for scanner.Scan() {
        s = append(s, fmt.Sprintf("%s\n", scanner.Text()))
    }
    if err = scanner.Err(); err != nil {
        log.Fatal(err)
    }
    fmt.Println(s)

    // Loop through the files, add the citation to the top of each file
    // Ideally do this if there is no header or force override...but that
    //  will be later
    root_dir := "."

    var file_list ProjectFiles
	err = filepath.Walk(root_dir, func(path string, _ os.FileInfo, err error) error {
        if err != nil {
            return err
        }
        file_list = append(file_list, path)
        return err
    })
    if err != nil {
        log.Println(err)
        return 1, err
    }
    return 0, nil
}