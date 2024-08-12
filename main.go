// TODO: Push after commit
// TODO: TUI
// TODO: Tab completions
// FIX: "view" outputs
// FIX: git diff colors
// FIX: remove the git repo inside the git repo from showing
// up as a change
package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strings"
	"strconv"
	"slices"
)

func main() {
  if checkDir() {
    getFiles("")
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
  case "":
  case "view":
    if len(input) > 1 {
      fmt.Println(getFiles(input[1]))
    } else {
      getFiles("")
    }
  case "add":
    if len(input) == 1 {
      files := getFiles("untracked")
      files = append(files[:], getFiles("changed")[:]...)
      gitAdd(addFiles(files, "normal"))
      fmt.Printf("Commit message? ")
      message, err := bufio.NewReader(os.Stdin).ReadString('\n')
      if err == nil {
	message = strings.TrimSuffix(message, "\n")
	commitFiles(message)
	fmt.Printf("Push changes? [Y/n]")
	message, err := bufio.NewReader(os.Stdin).ReadString('\n')
	if err != nil {
	  fmt.Println()
	} else {
	  message = strings.ToLower(strings.TrimSuffix(message, "\n"))
	  if message == "y" || message == "" {
	    pushFiles()
	  } else {
	    fmt.Println()
	  }
	}
      } else {
	fmt.Println()
      }
    } else if len(input) == 2 {
      files := getFiles("untracked")
      files = append(files[:], getFiles("changed")[:]...)
      gitAdd(addFiles(files, input[1]))
      fmt.Printf("Commit message? ")
      message, err := bufio.NewReader(os.Stdin).ReadString('\n')
      if err == nil {
	message = strings.TrimSuffix(message, "\n")
	commitFiles(message)
      } else {
	fmt.Println()
      }
    } else {
      if input[1] == "normal" {
	gitAdd(input[2:])
      } else {
	fmt.Println("Work in progress")
      }
    }
  case "restore":
    if len(input) == 1 {
      restoreFiles(addFiles(getFiles("added"), "normal"))
    } else {
      restoreFiles(input[1:])
    }
  case "commit":
    if len(input) > 1 {
      commitFiles(strings.Join(input[1:], " "))
    } else {
      fmt.Println("No commit message specified")
    }
  case "diff":
    if len(input) > 1 {
      changes := getDiff(input[1])
      changes_str := strings.Join(changes, "\n")
      fmt.Println(changes_str)
    } else {
      fmt.Println("No file specified")
    }
  case "push":
    pushFiles()
  case "fetch":
    fetchFiles()
  case "pull":
    pullFiles()
  case "exit":
    os.Exit(0)
  case "help":
    help()
  default:
    fmt.Printf("Invalid option: %v. Use 'help' to see available commands\n", input[0])
  }
}

func help() {
  fmt.Printf("Available commands:\n\n")
  fmt.Println("\tview [untracked, added, changed, all, none]:")
  fmt.Println("\t\tUsed to the status of files.")
  fmt.Println("\tadd [files, none]:")
  fmt.Println("\t\tUsed to add files to commit")
  fmt.Println("\tdiff [file]:")
  fmt.Println("\t\tUsed to view modifications made to the files")
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
  status := exec.Command("git", "status", "-s")
  if unstaged, err := status.CombinedOutput(); err == nil {
    git_status = strings.Split(string(unstaged), "\n")
  } else {
    log.Fatal(err)
  }
  return
}

func getFiles(state string) (files []string) {
  var startstring string
  switch state {
  case "untracked":
    startstring = "?? "
  case "added":
    startstring = "M"
  case "changed":
    startstring = " M "
  case "":
    untracked_files := strings.Join(getFiles("untracked"), "\n")
    changed_files := strings.Join(getFiles("changed"), "\n")
    added_files := strings.Join(getFiles("added"), "\n")
    fmt.Printf("\nUntracked files: \n%v\n\nChanged files: \n%v\n\nFiles to commit: \n%v\n\n",
      untracked_files,
      changed_files,
      added_files)
    return
  default:
    fmt.Printf("Invalid option '%v'", state)
    return
  }
  available_files := getStatus()
  for index, value := range available_files {
    if strings.HasPrefix(value, startstring) {
      available_files[index] = strings.TrimPrefix(available_files[index], startstring)
      available_files[index] = strings.TrimSpace(available_files[index])
      files = append(files, available_files[index])
    }
  }
  return
}

func gitAdd(files []string) {
  for _, value := range files {
    add := exec.Command("git", "add", value)
    err := add.Run()
    if err != nil {
      log.Fatal(err)
    } else {
      log.Printf("Added: %v\n", value)
    }
  }
}

// TODO: Allow usage of '*'
func addFiles(files []string, mode string) (addedfiles []string) {
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
  var indcies []int
  for _, value := range inputsplit {
    intvalue, err := strconv.Atoi(value)
    indcies = append(indcies[:], intvalue - 1)
    if err != nil {
      fmt.Println("Invalid input.")
      addFiles(files, mode)
    }
  }
  for index := range len(files) {
    switch mode {
    case "normal":
      if slices.Contains(indcies, index) {
	addedfiles = append(addedfiles, files[index])
      }
    case "exclude":
      if slices.Contains(indcies, index) {
	continue
      }
      addedfiles = append(addedfiles, files[index])
    }
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

func getDiff(filename string) (changes []string) {
  diff := exec.Command("git", "diff", "--minimal", filename)
  changed, err := diff.CombinedOutput()
  if err != nil || string(changed) == "" {
    log.Println("An error occured, file does not exist or there are simply no changes")
  } else {
    changed_lines := strings.Split(string(changed), "\n")
    changes = changed_lines[4:]
    for index, value := range changes {
      if strings.HasPrefix("+", value) {
	changes[index] = "\033[32m GREEN" + changes[index] + "\033[0m" 
      }
    }
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

func fetchFiles() {
  fetch := exec.Command("git", "fetch")
  message, err := fetch.CombinedOutput()
  if err == nil {
    log.Println(message)
  } else {
    log.Println(err)
  }
}

func pullFiles() {
  pull := exec.Command("git", "pull")
  message, err := pull.CombinedOutput()
  if err == nil {
    log.Println(message)
  } else {
    log.Println(err)
  }
}
