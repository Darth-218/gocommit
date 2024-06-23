package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"strconv"
)

func main() {
  if checkDir() {
    loop()
  } else {
    log.Fatal("Current directory is not a git repository.")
  }
}

func checkDir() (isgit bool) {
  isgit = false
  fullpath, _ := os.Getwd()
  paths := strings.Split(fullpath, "/")
  var end int = len(paths) - 1
  for end >= 0 {
    currentpath := paths[:end]; end--
    currentpathjoined := strings.Join(currentpath, "/") + "/.git"
    _, err := os.ReadDir(currentpathjoined)
    if err == nil {
      isgit = true
      return
    }
  }
  return
}

func loop() {
  reader := bufio.NewReader(os.Stdin)
  for true {
    fmt.Printf(">> ")
    input, err := reader.ReadString('\n')
    if err == nil {
      input = strings.TrimSuffix(input, "\n")
      inputsplit := strings.Split(input, " ")
      runCommand(inputsplit)
    } else {
      fmt.Printf("\n%v\n", err)
      break
    }
  }
}

func runCommand(input []string) {
  switch input[0] {
  case "view":
    if len(input) > 1 {
      fmt.Println(getFiles(input[1]))
    } else {
      getFiles("")
    }
  case "add":
    files := getFiles("untracked")
    files = append(files[:], getFiles("changed")[:]...)
    addFiles(files)
  case "commit":
    if len(input) > 1 {
      commitFiles(strings.Join(input[1:], " "))
    } else {
      fmt.Println("No commit message specified")
    }
  case "push":
    pushFiles()
  case "exit":
    os.Exit(0)
  default:
    fmt.Println("Invalid option:", input[0])
  }
}

func getStatus() (git_status []string) {
  status := exec.Command("git", "status")
  if unstaged, err := status.CombinedOutput(); err == nil {
    git_status = strings.Split(string(unstaged), "\n")
    for index, value := range git_status {
    git_status[index] = strings.TrimSpace(value)
    }
  } else {
    log.Fatal(err)
  }
  return
}

func getFiles(state string) (files []string) {
  var startstring, trim, alttrim string
  switch state {
  case "untracked":
    startstring = "Untracked files:"
  case "added":
    startstring = "Changes to be committed:"
    trim = "new file:   "
    alttrim = "modified:   "
  case "changed":
    startstring = "Changes not staged for commit:"
    trim = "modified:   "
    alttrim = "deleted:   "
  case "":
    fallthrough
  case "all":
    fmt.Printf("Untracked files: %v\nChanged files: %v\nFiles to commit: %v\n",
      getFiles("untracked"),
      getFiles("changed"),
      getFiles("added"))
    return
  default:
    fmt.Printf("Invalid option \"%v\"", state)
    return
  }
  files = getStatus()
  keystart := len(files)
  keyend := keystart
  for index, value := range files {
    if value == startstring {
      keystart = index
      continue
    }
    if len(value) == 0 && index > keystart {
      keyend = index
      break
    }
    files[index] = strings.TrimPrefix(files[index], trim)
    files[index] = strings.TrimPrefix(files[index], alttrim)
  }
  if keystart == len(files) {
    files = nil
    return
  }
  if keyend == 0 {
    keyend = len(files)
  }
  if startstring == "Changes not staged for commit:" {
    keystart += 1
  }
  files = files[keystart + 2:keyend]
  return
}

func addFiles(files []string) {
  for index, value := range files {
    fmt.Printf("%v: %v\n", index + 1, value)
  }
  fmt.Println("Enter the index of the files to add.")
  reader := bufio.NewReader(os.Stdin)
  fmt.Print("--> ")
  input, err := reader.ReadString('\n')
  var inputsplit []string
  if err == nil {
    input = strings.TrimSuffix(input, "\n")
    inputsplit = strings.Split(input, " ")
    for _, value := range inputsplit {
      intvalue, _ := strconv.Atoi(value)
      add := exec.Command("git", "add", files[intvalue - 1])
      err := add.Run()
      if err != nil {
        log.Fatal(err)
      } else {
        log.Printf("Added: %v\n", files[intvalue - 1])
      }
    }
  } else {
    fmt.Println()
  }
}

func commitFiles(commitmessage string) {
  commit := exec.Command("git", "commit", "-m", commitmessage)
  err := commit.Run()
  if err != nil {
    log.Fatal(err)
  } else {
    log.Println("Files committed")
  }
}

func pushFiles() {
  push := exec.Command("git", "push")
  err := push.Run()
  if err != nil {
    log.Fatal(err)
  } else {
    log.Println("Pushed changes ->")
  }
}
