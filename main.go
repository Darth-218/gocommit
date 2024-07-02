// TODO: TUI
// TODO: Tab completions
// TODO: 'view' command representation
// TODO: Fix repo detection
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
  end := len(paths)
  for end >= 0 {
    currentpath := paths[0:end]
    currentpathjoined := strings.Join(currentpath, "/") + "/.git"
    _, err := os.ReadDir(currentpathjoined)
    if err == nil {
      isgit = true
      return
    }
    end--
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
  command := strings.TrimSpace(input[0])
  switch command {
  case "view":
    if len(input) > 1 {
      getFiles(input[1])
    } else {
      getFiles("")
    }
  case "add":
    if len(input) == 1 {
      files := getFiles("untracked")
      files = append(files[:], getFiles("changed")[:]...)
      gitAdd(addFiles(files))
    } else {
      gitAdd(input[1:])
    }
  case "restore":
    if len(input) == 1 {
      restoreFiles(addFiles(getFiles("added")))
    } else {
      restoreFiles(input[1:])
    }
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
  case "help":
    help()
  default:
    fmt.Printf("Invalid option: %v. Use 'help' to see available commands\n", input[0])
  }
}

func help() {
  fmt.Println("Available commands:")
  fmt.Println("\tview [untracked, added, changed, all, none]:")
  fmt.Println("\t\tUsed to the status of files.")
  fmt.Println("\tadd [files, none]:")
  fmt.Println("\t\tUsed to add files to commit")
  fmt.Println("\trestore [files, none]:")
  fmt.Println("\t\tUsed to restore added files")
  fmt.Println("\tcommit [message]:")
  fmt.Println("\t\tUsed to commit added files")
  fmt.Println("\tpush:")
  fmt.Println("\t\tUsed to push committed files")
  fmt.Println("\texit:")
  fmt.Println("\t\tExits.")
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
    files[index] = strings.TrimSpace(files[index])
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

func gitAdd(files []string) {
  for index, _ := range files {
    add := exec.Command("git", "add", files[index])
    err := add.Run()
    if err != nil {
      log.Fatal(err)
    } else {
      log.Printf("Added: %v\n", files[index])
    }
  }
}

// TODO: Allow usage of '*'
func addFiles(files []string) (addedfiles []string) {
  if len(files) == 0 {
    return
  }
  var inputsplit []string
    for index, value := range files {
      fmt.Printf("%v: %v\n", index + 1, value)
    }
    fmt.Println("Enter the index of the files to select.")
    reader := bufio.NewReader(os.Stdin)
    fmt.Print("--> ")
    input, err := reader.ReadString('\n')
    if err != nil {
      fmt.Println()
      return
    }
    if input == "\n" {
      addedfiles = files
      return
    }
    input = strings.TrimSuffix(input, "\n")
    inputsplit = strings.Split(input, " ")
    for _, value := range inputsplit {
      intvalue, err := strconv.Atoi(value)
      if err != nil {
	fmt.Println("Invalid input.")
	addFiles(files)
      }
      addedfiles = append(addedfiles, files[intvalue - 1])
    }
    return
}

func restoreFiles(files []string) {
  for _, value := range files {
    restore := exec.Command("git", "restore", "--staged", value)
    err := restore.Run()
    if err != nil {
      log.Println("Failed to restore files, are they added?")
    } else {
      log.Printf("Restored file '%v'.\n", value)
    }
  }
}

func getCommitid() (commitid string) {
  logs := exec.Command("git", "log")
  if ids, err := logs.CombinedOutput(); err == nil {
    commitids := strings.Split(string(ids), "\n")
    commitid = commitids[0]
  } else {
    log.Fatal(err)
  }
  return
}

func commitFiles(commitmessage string) {
  commit := exec.Command("git", "commit", "-m", commitmessage)
  message, err := commit.CombinedOutput()
  if strings.Contains(string(message), commitmessage) == false {
    log.Println("No changes to commit.")
    return
  }
  if err != nil {
    log.Fatal(err)
  } else {
    log.Println("Files committed ->", getCommitid())
  }
}

func pushFiles() {
  push := exec.Command("git", "push")
  message, err := push.CombinedOutput()
  if string(message) == "Everything up-to-date\n" {
    log.Println("No changes to push")
    return
  }
  if err != nil {
    log.Fatal(err)
  } else {
    log.Println("Pushed changes")
  }
}
