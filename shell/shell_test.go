package main

import (
	"bytes"
	"database/sql"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	_ "github.com/mattn/go-sqlite3"
	"golang.org/x/crypto/bcrypt"
)

const shellPath = "./mysh"

func TestMain(m *testing.M) {
	cmd := exec.Command("go", "build", "-o", shellPath)
	if err := cmd.Run(); err != nil {
		fmt.Printf("Failed to build shell: %v\n", err)
		os.Exit(1)
	}

	db, err := sql.Open("sqlite3", "./test_shell.db")
	if err != nil {
		panic(err)
	}
	defer db.Close()

	_, err = db.Exec(`CREATE TABLE IF NOT EXISTS users (
		username TEXT PRIMARY KEY,
		password_hash TEXT NOT NULL
	)`)
	if err != nil {
		panic(err)
	}

	exitCode := m.Run()

	os.Remove(shellPath)
	os.Remove("./test_shell.db")
	os.Exit(exitCode)
}

func runShell(t *testing.T, input string) (string, string, error) {
	cmd := exec.Command(shellPath)
	stdin, err := cmd.StdinPipe()
	if err != nil {
		t.Fatalf("Failed to create stdin pipe: %v", err)
	}
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		t.Fatalf("Failed to create stdout pipe: %v", err)
	}
	stderr, err := cmd.StderrPipe()
	if err != nil {
		t.Fatalf("Failed to create stderr pipe: %v", err)
	}

	err = cmd.Start()
	if err != nil {
		t.Fatalf("Failed to start shell: %v", err)
	}

	_, err = io.WriteString(stdin, input+"\nexit\n")
	if err != nil {
		t.Fatalf("Failed to write to stdin: %v", err)
	}
	err = stdin.Close()
	if err != nil {
		t.Fatalf("Failed to close stdin: %v", err)
	}

	outBuf, err := io.ReadAll(stdout)
	if err != nil {
		t.Fatalf("Failed to read stdout: %v", err)
	}
	errBuf, err := io.ReadAll(stderr)
	if err != nil {
		t.Fatalf("Failed to read stderr: %v", err)
	}

	err = cmd.Wait()

	return string(outBuf), string(errBuf), err
}
func TestBuiltinCommands(t *testing.T) {
	t.Run("Echo", func(t *testing.T) {
		out, _, _ := runShell(t, "echo 'hello $USER'")
		if !strings.Contains(out, "hello $USER") {
			t.Errorf("Echo failed, got: %s", out)
		}
	})

	t.Run("Pwd", func(t *testing.T) {
		out, _, _ := runShell(t, "pwd")
		wd, _ := os.Getwd()
		if !strings.Contains(out, wd) {
			t.Errorf("Pwd failed, got: %s", out)
		}
	})

	t.Run("TypeBuiltin", func(t *testing.T) {
		out, _, _ := runShell(t, "type echo")
		if !strings.Contains(out, "shell builtin") {
			t.Errorf("Type builtin failed, got: %s", out)
		}
	})
}

func TestRedirection(t *testing.T) {
	tmpDir := t.TempDir()

	t.Run("OutputRedirection", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "test.txt")
		runShell(t, fmt.Sprintf("echo content > %s", testFile))
		data, _ := os.ReadFile(testFile)
		if !strings.Contains(string(data), "content") {
			t.Error("Output redirection failed")
		}
	})

	t.Run("AppendRedirection", func(t *testing.T) {
		testFile := filepath.Join(tmpDir, "append.txt")
		runShell(t, fmt.Sprintf("echo line1 >> %s", testFile))
		runShell(t, fmt.Sprintf("echo line2 >> %s", testFile))
		data, _ := os.ReadFile(testFile)
		if strings.Count(string(data), "line") != 2 {
			t.Error("Append redirection failed")
		}
	})
}

func TestUserManagement(t *testing.T) {
	db, _ := sql.Open("sqlite3", "./test_shell.db")
	defer db.Close()

	t.Run("AddUser", func(t *testing.T) {
		out, _, _ := runShell(t, "adduser new 123")
		if !strings.Contains(out, "user created successfully") {
			t.Error("Adduser failed")
		}
	})

	t.Run("Login", func(t *testing.T) {
		out, _, _ := runShell(t, "login new 123")
		if !strings.Contains(out, "") {
			t.Error("Login failed")
		}

		out, _, _ = runShell(t, "login new wrongpass")
		if !strings.Contains(out, "login: incorrect password") {
			t.Error("Login error handling failed")
		}
	})
}

func TestHistory(t *testing.T) {
	db, _ := sql.Open("sqlite3", "./test_shell.db")
	defer db.Close()

	t.Run("LoggedInHistory", func(t *testing.T) {

		out, _, _ := runShell(t, "login parsa 123\necho command1\necho command2\nhistory")
		if !strings.Contains(out, "| echo command1 | 1 |") || !strings.Contains(out, "| echo command2 | 1 |") {
			t.Errorf("History recording failed, got: %s", out)
		}
	})

	t.Run("HistoryClean", func(t *testing.T) {
		runShell(t, "login parsa 123\nhistory clean")
		out, _, _ := runShell(t, "history")
		if !strings.Contains(out, "empty command history") {
			t.Error("History clean failed")
		}
	})
}

func TestExternalCommands(t *testing.T) {
	t.Run("SimpleCommand", func(t *testing.T) {
		out, _, _ := runShell(t, "ls")
		if !strings.Contains(out, "shell.go") && !strings.Contains(out, "shell_test.go") {
			t.Error("External command execution failed")
		}
	})

	t.Run("WithRedirection", func(t *testing.T) {
		tmpFile := filepath.Join(t.TempDir(), "ls_output.txt")
		runShell(t, fmt.Sprintf("ls > %s", tmpFile))
		data, _ := os.ReadFile(tmpFile)
		if len(data) == 0 {
			t.Error("External command redirection failed")
		}
	})
}

func TestErrorHandling(t *testing.T) {
	t.Run("InvalidCommand", func(t *testing.T) {
		_, errOut, _ := runShell(t, "invalid_command")
		if !strings.Contains(errOut, "not found") {
			t.Error("Invalid command handling failed")
		}
	})

	t.Run("SyntaxError", func(t *testing.T) {
		_, errOut, _ := runShell(t, "echo >")
		if !strings.Contains(errOut, "syntax error") {
			t.Error("Syntax error handling failed")
		}
	})
}

func TestComplexWorkflow(t *testing.T) {
	db, _ := sql.Open("sqlite3", "./test_shell.db")
	defer db.Close()

	hashed, _ := bcrypt.GenerateFromPassword([]byte("testpass"), bcrypt.DefaultCost)
	db.Exec("INSERT INTO users VALUES (?, ?)", "workflowuser", hashed)

	commands := []string{
		"login workflowuser testpass",
		"echo 'Hello World' > greeting.txt",
		"cat greeting.txt",
		"history",
		"logout",
	}

	var b bytes.Buffer
	for _, cmd := range commands {
		b.WriteString(cmd + "\n")
	}

	out, _, _ := runShell(t, b.String())
	if !strings.Contains(out, "Hello") {
		t.Error("Complex workflow failed")
	}
}
