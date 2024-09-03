// TODO: REFACTOR
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
  "github.com/chzyer/readline"
  "log"
  "os"
  "os/exec"
  "slices"
  "strconv"
  "strings"
)

func main() {
  if isGitDir() {
    getFiles("")
    loop()
  } else {
    log.Fatal("Current directory is not a git repository.")
  }
}

func isGitDir() bool {
  full_path, _ := os.Getwd()
  sep_paths := strings.Split(full_path, "/")
  for index := len(sep_paths); index >= 0; index-- {
    path_to_git := strings.Join(sep_paths[0:index], "/") + "/.git"
    _, err := os.ReadDir(path_to_git)
    if err == nil {
      return true
    }
  }
  return false
}

func loop() {
  prompt, _ := readline.New(">> ")
  defer prompt.Close()
  for {
    input, err := prompt.Readline()
    if err != nil {
      fmt.Printf("\n%v\n", err)
      break
    }
    split_input := strings.Split(input, " ")
    runCommand(split_input)
  }
}

func runCommand(argv []string) {
  argc := len(argv)
  command := strings.TrimSpace(argv[0])
  switch command {
  case "": // If there's no input do nothing
  case "view": // Viewing files according to their status
    // TODO: User passing more than one argument
    if argc > 1 { // View only the argument passed 
      fmt.Println(getFiles(argv[1]))
    } else { // If the command is executed without args view *
      getFiles("")
    }
  case "add": // Staging files
    files := append(getFiles("untracked"), getFiles("changed")[:]...)
    // Input prompts
    commit_prompt, _ := readline.New("Commit message? ")
    push_prompt, _ := readline.New("Push changes? [Y/n] ")
    if argc == 1 {
      gitAdd(selectFiles(files, "normal"))
      commit_message, err := commit_prompt.Readline()
      if err != nil {
	fmt.Println()
	return
      }
      commitFiles(strings.TrimSuffix(commit_message, "\n"))
      push_confirmation, err := push_prompt.Readline()
      if err != nil {
	fmt.Println()
	return
      }
      push_confirmation = strings.ToLower(strings.TrimSuffix(push_confirmation, "\n"))
      if push_confirmation == "y" || push_confirmation == "" {
	pushFiles()
      }
    } else if argc > 1 {
      gitAdd(argv[1:])
    }
  case "restore":
    if argc == 1 {
      restoreFiles(selectFiles(getFiles("added"), "normal"))
    } else {
      restoreFiles(argv[1:])
    }
  case "commit":
    if argc > 1 {
      commitFiles(strings.Join(argv[1:], " "))
    } else {
      fmt.Println("No commit message specified")
    }
  case "diff":
    if argc > 1 {
      changes := getDiff(argv[1])
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
    fmt.Printf("Invalid option: %v. Use 'help' to see available commands\n", command)
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
func selectFiles(files []string, mode string) (selected_files []string) {
  if len(files) == 0 {
    return
  }
  var inputsplit []string
  for index, value := range files {
    fmt.Printf("%v: %v\n", index+1, value)
  }
  fmt.Println("Enter indcies of files to select.")
  reader := bufio.NewReader(os.Stdin)
  fmt.Print("--> ")
  input, err := reader.ReadString('\n')
  if err != nil {
    fmt.Println()
    return
  }
  if input == "\n" {
    selected_files = files
    return
  }
  input = strings.TrimSuffix(input, "\n")
  inputsplit = strings.Split(input, " ")
  var indcies []int
  for _, value := range inputsplit {
    intvalue, err := strconv.Atoi(value)
    indcies = append(indcies[:], intvalue-1)
    if err != nil {
      fmt.Println("Invalid input.")
      selectFiles(files, mode)
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
