package main

import (
	"bufio"
	"bytes"
	"database/sql"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

var builtins = map[string]bool{
	"exit":    true,
	"echo":    true,
	"cat":     true,
	"type":    true,
	"pwd":     true,
	"cd":      true,
	"login":   true,
	"logout":  true,
	"adduser": true,
	"history": true,
	"ls":      true,
}

func main() {
	db := initDB()
	defer db.Close()

	reader := bufio.NewReader(os.Stdin)
	currentUser := ""
	sessionHistory := []string{}

	for {
		// Display prompt
		prompt := "$ "
		if currentUser != "" {
			prompt = currentUser + ":$ "
		}
		fmt.Print(prompt)

		// Read input
		line, err := reader.ReadString('\n')
		if err != nil {
			fmt.Fprintln(os.Stderr, "Error reading input:", err)
			continue
		}

		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Split into arguments
		args := splitArgs(line)
		if len(args) == 0 {
			continue
		}

		cmd := args[0]
		cmdArgs := args[1:]

		// Update history
		if currentUser != "" {
			_, err = db.Exec("INSERT INTO command_history (username, command) VALUES (?, ?)", currentUser, line)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Failed to save history: %v\n", err)
			}
		} else {
			sessionHistory = append(sessionHistory, line)
		}

		// Handle commands
		switch cmd {
		case "exit":
			handleExit(cmdArgs)
		case "echo":
			handleEcho(cmdArgs)
		case "cat":
			handleCat(cmdArgs)
		case "type":
			handleType(cmdArgs)
		case "pwd":
			handlePwd(cmdArgs)
		case "cd":
			handleCd(cmdArgs)
		case "login":
			handleLogin(cmdArgs, db, &currentUser)
		case "logout":
			currentUser = ""
		case "adduser":
			handleAddUser(cmdArgs, db)
		case "history":
			if len(cmdArgs) > 0 && cmdArgs[0] == "clean" {
				handleHistoryClean(currentUser, db, &sessionHistory)
			} else {
				handleHistory(currentUser, db, sessionHistory)
			}
		case "ls":
			handleLs(cmdArgs)
		default:
			if builtins[cmd] {
				fmt.Fprintf(os.Stderr, "%s: built-in command not implemented\n", cmd)
				continue
			}
			executeExternalCommand(cmd, cmdArgs)
		}
	}
}

// Database Initialization
func initDB() *sql.DB {
	db, err := sql.Open("sqlite3", "./shell.db")
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		username TEXT PRIMARY KEY,
		password_hash TEXT NOT NULL
	)`)
	if err != nil {
		panic(err)
	}

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS command_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		username TEXT,
		command TEXT,
		timestamp DATETIME DEFAULT CURRENT_TIMESTAMP
	)`)
	if err != nil {
		panic(err)
	}

	return db
}
func handleExit(args []string) {
	stdoutFile, stderrFile, processedArgs := processRedirection(args)
	defer func() {
		if stdoutFile != nil {
			stdoutFile.Close()
		}
		if stderrFile != nil {
			stderrFile.Close()
		}
	}()

	if len(processedArgs) > 1 {
		if stderrFile != nil {
			fmt.Fprintln(stderrFile, "exit: too many arguments")
		} else {
			fmt.Println("exit: too many arguments")
		}
		return
	}

	code := 0
	if len(processedArgs) == 1 {
		_, err := fmt.Sscanf(processedArgs[0], "%d", &code)
		if err != nil {
			if stderrFile != nil {
				fmt.Fprintln(stderrFile, "exit: invalid status code")
			} else {
				fmt.Println("exit: invalid status code")
			}
			return
		}
	}

	fmt.Printf("exit status %d\n", code)
	os.Exit(code)
}

func handlePwd(args []string) {
	stdoutFile, stderrFile, _ := processRedirection(args)
	defer func() {
		if stdoutFile != nil {
			stdoutFile.Close()
		}
		if stderrFile != nil {
			stderrFile.Close()
		}
	}()

	dir, err := os.Getwd()
	if err != nil {
		fmt.Fprintln(os.Stderr, "pwd:", err)
		return
	}
	if stdoutFile != nil {
		fmt.Fprintln(stdoutFile, dir)
	} else {
		fmt.Println(dir)
	}
}
func handleEcho(args []string) {
	stdoutFile, stderrFile, processedArgs := processRedirection(args)
	defer func() {
		if stdoutFile != nil {
			stdoutFile.Close()
		}
		if stderrFile != nil {
			stderrFile.Close()
		}
	}()

	var output strings.Builder
	for _, arg := range processedArgs {
		if strings.HasPrefix(arg, "\"") && strings.HasSuffix(arg, "\"") {
			stripped := arg[1 : len(arg)-1]
			processed := processEscapes(stripped)
			output.WriteString(replaceEnvVars(processed))
		} else if !strings.HasPrefix(arg, "'") || !strings.HasSuffix(arg, "'") {
			output.WriteString(replaceEnvVars(arg))
		} else {
			stripped := arg[1 : len(arg)-1]
			processed := processEscapes(stripped)
			output.WriteString(processed)
			// output.WriteString(arg)
		}
		output.WriteString(" ")
	}

	result := strings.TrimSpace(output.String())

	if stdoutFile != nil {
		fmt.Fprint(stdoutFile, result+"\n")
	} else {
		fmt.Println(result)
	}
}

func replaceEnvVars(s string) string {
	re := regexp.MustCompile(`\$([A-Za-z_][A-Za-z0-9_]*)`)
	return re.ReplaceAllStringFunc(s, func(m string) string {
		return os.Getenv(m[1:])
	})
}
func handleCat(args []string) {
	stdoutFile, stderrFile, processedArgs := processRedirection(args)
	defer func() {
		if stdoutFile != nil {
			stdoutFile.Close()
		}
		if stderrFile != nil {
			stderrFile.Close()
		}
	}()

	if len(processedArgs) == 0 {
		if stderrFile != nil {
			fmt.Fprintln(stderrFile, "cat: missing file argument")
		} else {
			fmt.Fprintln(os.Stderr, "cat: missing file argument")
		}
		return
	}
	for _, file := range processedArgs {
		content, err := os.ReadFile(file)
		if err != nil {
			if stderrFile != nil {
				fmt.Fprintf(stderrFile, "cat: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "cat: %v\n", err)
			}
			continue
		}
		if stdoutFile != nil {
			fmt.Fprint(stdoutFile, string(content))
		} else {
			fmt.Print(string(content))
		}
	}
}
func handleType(args []string) {
	stdoutFile, stderrFile, processedArgs := processRedirection(args)
	defer func() {
		if stdoutFile != nil {
			stdoutFile.Close()
		}
		if stderrFile != nil {
			stderrFile.Close()
		}
	}()

	if len(processedArgs) == 0 {
		if stderrFile != nil {
			fmt.Fprintln(stderrFile, "type: missing argument")
		} else {
			fmt.Println("type: missing argument")
		}
		return
	}
	cmd := processedArgs[0]
	if builtins[cmd] {
		output := fmt.Sprintf("%s is a shell builtin\n", cmd)
		if stdoutFile != nil {
			fmt.Fprint(stdoutFile, output)
		} else {
			fmt.Print(output)
		}
		return
	}

	path := os.Getenv("PATH")
	for _, dir := range filepath.SplitList(path) {
		fullPath := filepath.Join(dir, cmd)
		if _, err := os.Stat(fullPath); err == nil {
			output := fmt.Sprintf("%s is %s\n", cmd, fullPath)
			if stdoutFile != nil {
				fmt.Fprint(stdoutFile, output)
			} else {
				fmt.Print(output)
			}
			return
		}
	}
	output := fmt.Sprintf("%s: command not found\n", cmd)
	if stderrFile != nil {
		fmt.Fprint(stderrFile, output)
	} else {
		fmt.Print(output)
	}
}

func handleCd(args []string) {
	stdoutFile, stderrFile, processedArgs := processRedirection(args)
	defer func() {
		if stdoutFile != nil {
			stdoutFile.Close()
		}
		if stderrFile != nil {
			stderrFile.Close()
		}
	}()
	target := ""
	if len(processedArgs) == 0 {
		home, err := os.UserHomeDir()
		if err != nil {
			if stderrFile != nil {
				fmt.Fprintf(stderrFile, "cd: %v\n", err)
			} else {
				fmt.Fprintf(os.Stderr, "cd: %v\n", err)
			}
			return
		}
		target = home
	} else {
		target = processedArgs[0]
	}

	if err := os.Chdir(target); err != nil {
		if stderrFile != nil {
			fmt.Fprintf(stderrFile, "cd: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "cd: %v\n", err)
		}
	}
}

// User Management
func handleAddUser(args []string, db *sql.DB) {
	stdoutFile, stderrFile, processedArgs := processRedirection(args)
	defer func() {
		if stdoutFile != nil {
			stdoutFile.Close()
		}
		if stderrFile != nil {
			stderrFile.Close()
		}
	}()

	if len(processedArgs) < 1 || len(processedArgs) > 2 {
		if stderrFile != nil {
			fmt.Fprintln(stderrFile, "adduser: invalid arguments")
		} else {
			fmt.Println("adduser: invalid arguments")
		}
		return
	}
	username := processedArgs[0]
	password := ""
	if len(processedArgs) == 2 {
		password = processedArgs[1]
	}

	hashed, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		if stderrFile != nil {
			fmt.Fprintf(stderrFile, "Error creating user: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "Error creating user: %v\n", err)
		}
		return
	}

	_, err = db.Exec("INSERT INTO users (username, password_hash) VALUES (?, ?)", username, hashed)
	if err != nil {
		if stderrFile != nil {
			fmt.Fprintln(stderrFile, "duplicate user exists with this username")
		} else {
			fmt.Println("duplicate user exists with this username")
		}
	} else {
		if stdoutFile != nil {
			fmt.Fprintln(stdoutFile, "user created successfully")
		} else {
			fmt.Println("user created successfully")
		}
	}
}

func handleLogin(args []string, db *sql.DB, currentUser *string) {
	stdoutFile, stderrFile, processedArgs := processRedirection(args)
	defer func() {
		if stdoutFile != nil {
			stdoutFile.Close()
		}
		if stderrFile != nil {
			stderrFile.Close()
		}
	}()

	if len(processedArgs) < 1 || len(processedArgs) > 2 {
		if stderrFile != nil {
			fmt.Fprintln(stderrFile, "login: invalid arguments")
		} else {
			fmt.Println("login: invalid arguments")
		}
		return
	}
	username := processedArgs[0]
	password := ""
	if len(processedArgs) == 2 {
		password = processedArgs[1]
	}

	var storedHash string
	err := db.QueryRow("SELECT password_hash FROM users WHERE username = ?", username).Scan(&storedHash)
	if err != nil {
		if stderrFile != nil {
			fmt.Fprintln(stderrFile, "login: user not found")
		} else {
			fmt.Println("login: user not found")
		}
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(storedHash), []byte(password))
	if err != nil {
		if stderrFile != nil {
			fmt.Fprintln(stderrFile, "login: incorrect password")
		} else {
			fmt.Println("login: incorrect password")
		}
		return
	}

	*currentUser = username
	if stdoutFile != nil {
		fmt.Fprintln(stdoutFile, "login successful")
	} else {
		fmt.Println("login successful")
	}
}

// History Management
func handleHistoryClean(currentUser string, db *sql.DB, sessionHistory *[]string) {
	if currentUser != "" {
		_, err := db.Exec("DELETE FROM command_history WHERE username = ?", currentUser)
		if err != nil {
			fmt.Fprintf(os.Stderr, "history clean: %v\n", err)
			return
		}
	} else {
		*sessionHistory = []string{}
	}
}

func handleHistory(currentUser string, db *sql.DB, sessionHistory []string) {
	type historyEntry struct {
		command string
		count   int
	}

	var entries []historyEntry

	if currentUser != "" {
		rows, err := db.Query(`
			SELECT command, COUNT(*) as count 
			FROM command_history 
			WHERE username = ? 
			GROUP BY command 
			ORDER BY count DESC, MAX(timestamp) DESC
		`, currentUser)
		if err != nil {
			fmt.Fprintf(os.Stderr, "history: %v\n", err)
			return
		}
		defer rows.Close()

		for rows.Next() {
			var cmd string
			var cnt int
			if err := rows.Scan(&cmd, &cnt); err != nil {
				fmt.Fprintf(os.Stderr, "history: %v\n", err)
				continue
			}
			entries = append(entries, historyEntry{cmd, cnt})
		}
	} else {
		counts := make(map[string]int)
		for _, cmd := range sessionHistory {
			counts[cmd]++
		}

		for cmd, cnt := range counts {
			entries = append(entries, historyEntry{cmd, cnt})
		}

		sort.Slice(entries, func(i, j int) bool {
			if entries[i].count == entries[j].count {
				return entries[i].command < entries[j].command
			}
			return entries[i].count > entries[j].count
		})
	}
	if len(entries) == 1 {
		fmt.Printf("empty command history\n")
	}
	for _, e := range entries {
		if e.command != "history" {
			fmt.Printf("| %s | %d |\n", e.command, e.count)
		}
	}
}

func handleLs(args []string) {
	stdoutFile, stderrFile, processedArgs := processRedirection(args)
	defer func() {
		if stdoutFile != nil {
			stdoutFile.Close()
		}
		if stderrFile != nil {
			stderrFile.Close()
		}
	}()

	dir := "."
	if len(processedArgs) > 0 {
		dir = processedArgs[0]
	}

	files, err := os.ReadDir(dir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "ls: %v\n", err)
		return
	}

	var output strings.Builder
	for _, file := range files {
		output.WriteString(file.Name())
		output.WriteString(" ")
	}
	output.WriteString("\n")

	result := output.String()
	if stdoutFile != nil {
		fmt.Fprint(stdoutFile, result)
	} else {
		fmt.Print(result)
	}
}

// External Command Execution
func executeExternalCommand(cmdName string, args []string) {
	stdoutFile, stderrFile, newArgs := processRedirection(args)
	defer func() {
		if stdoutFile != nil {
			stdoutFile.Close()
		}
		if stderrFile != nil {
			stderrFile.Close()
		}
	}()

	cmd := exec.Command(cmdName, newArgs...)
	cmd.Stdin = os.Stdin

	if stdoutFile != nil {
		cmd.Stdout = stdoutFile
	} else {
		cmd.Stdout = os.Stdout
	}

	if stderrFile != nil {
		cmd.Stderr = stderrFile
	} else {
		cmd.Stderr = os.Stderr
	}

	if err := cmd.Run(); err != nil {
		if stderrFile != nil {
			fmt.Fprintf(stderrFile, "error executing command: %v\n", err)
		} else {
			fmt.Fprintf(os.Stderr, "error executing command: %v\n", err)
		}
	}
}
func splitArgs(line string) []string {
	var args []string
	var buf bytes.Buffer
	inSingle, inDouble, escape := false, false, false

	for _, r := range line {
		if escape {
			if inDouble {
				switch r {
				case '$', '\'', '"', '\\', 'n':
					buf.WriteRune(r)
				default:
					buf.WriteRune('\\')
					buf.WriteRune(r)
				}
			} else {
				buf.WriteRune('\\')
				buf.WriteRune(r)
			}
			escape = false
			continue
		}

		if r == '\\' {
			if inDouble {
				escape = true
				continue
			} else if inSingle {
				buf.WriteRune(r)
			} else {
				buf.WriteRune(r)
			}
			continue
		}

		if r == '\'' && !inDouble {
			if inSingle {
				inSingle = false
				buf.WriteRune(r)
				args = append(args, buf.String())
				buf.Reset()
			} else {
				inSingle = true
				if buf.Len() > 0 {
					args = append(args, buf.String())
					buf.Reset()
				}
				buf.WriteRune(r)
			}
			continue
		} else if r == '"' && !inSingle {
			inDouble = !inDouble
			buf.WriteRune(r)
			if !inDouble {
				args = append(args, buf.String())
				buf.Reset()
			}
			continue
		} else if (r == ' ' || r == '\t') && !inSingle && !inDouble {
			if buf.Len() > 0 {
				args = append(args, buf.String())
				buf.Reset()
			}
		} else {
			buf.WriteRune(r)
		}
	}

	if buf.Len() > 0 {
		args = append(args, buf.String())
	}

	return args
}
func processEscapes(s string) string {
	var result strings.Builder
	escape := false

	for _, r := range s {
		if escape {
			switch r {
			case '$', '\'', '"', '\\', 'n', '`':
				result.WriteRune(r)
			default:
				result.WriteRune('\\')
				result.WriteRune(r)
			}
			escape = false
		} else if r == '\\' {
			escape = true
		} else {
			result.WriteRune(r)
		}
	}

	if escape {
		result.WriteRune('\\')
	}

	return result.String()
}

func processRedirection(args []string) (stdoutFile, stderrFile *os.File, processedArgs []string) {
	var newArgs []string

	for i := 0; i < len(args); {
		arg := args[i]
		switch arg {
		case ">", ">>", "1>", "1>>":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "syntax error: no file specified for output redirection")
				return nil, nil, nil
			}
			filename := args[i+1]
			flag := os.O_WRONLY | os.O_CREATE
			if arg == ">" || arg == "1>" {
				flag |= os.O_TRUNC
			} else {
				flag |= os.O_APPEND
			}
			file, err := os.OpenFile(filename, flag, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error opening file: %v\n", err)
				return nil, nil, nil
			}
			stdoutFile = file
			i += 2
		case "2>", "2>>":
			if i+1 >= len(args) {
				fmt.Fprintln(os.Stderr, "syntax error: no file specified for error redirection")
				return nil, nil, nil
			}
			filename := args[i+1]
			flag := os.O_WRONLY | os.O_CREATE
			if arg == "2>" {
				flag |= os.O_TRUNC
			} else {
				flag |= os.O_APPEND
			}
			file, err := os.OpenFile(filename, flag, 0644)
			if err != nil {
				fmt.Fprintf(os.Stderr, "error opening file: %v\n", err)
				return nil, nil, nil
			}
			stderrFile = file
			i += 2
		default:
			newArgs = append(newArgs, arg)
			i++
		}
	}
	return stdoutFile, stderrFile, newArgs
}
