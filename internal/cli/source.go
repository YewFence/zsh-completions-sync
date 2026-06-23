package cli

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"
)

const (
	httpTimeout = 30 * time.Second
	userAgent   = "zsh-completions-sync"
)

type CommandSource struct {
	Command []string
}

type LocalFileSource struct {
	Path string
}

type HTTPFileSource struct {
	URL string
}

type GitFileSource struct {
	Repository string
	Path       string
	Ref        string
}

type FileSource struct {
	File any
}

type SourceReadResult struct {
	Content []byte
	Error   string
}

func parseSource(config map[string]any) (any, bool) {
	command, hasCommand := parseCommand(config["command"])
	fileValue, hasFile := config["file"]
	if hasFile && fileValue != nil {
		fileSource, ok := parseFileSource(fileValue)
		if !ok {
			return nil, false
		}
		return FileSource{File: fileSource}, true
	}
	if !hasCommand {
		return nil, false
	}
	return CommandSource{Command: command}, true
}

func parseCommand(value any) ([]string, bool) {
	switch typed := value.(type) {
	case []any:
		if len(typed) == 0 {
			return nil, false
		}
		command := make([]string, 0, len(typed))
		for _, item := range typed {
			text, ok := item.(string)
			if !ok || text == "" {
				return nil, false
			}
			command = append(command, text)
		}
		return command, true
	case []string:
		if len(typed) == 0 {
			return nil, false
		}
		for _, item := range typed {
			if item == "" {
				return nil, false
			}
		}
		return append([]string(nil), typed...), true
	default:
		return nil, false
	}
}

func parseFileSource(value any) (any, bool) {
	switch typed := value.(type) {
	case string:
		if typed == "" {
			return nil, false
		}
		return parseFileString(typed)
	case map[string]any:
		return parseFileMapping(typed)
	default:
		return nil, false
	}
}

func parseFileString(value string) (any, bool) {
	if strings.HasPrefix(value, "git+") {
		return parseGitFileString(strings.TrimPrefix(value, "git+"))
	}

	parsed, err := url.Parse(value)
	if err == nil {
		switch parsed.Scheme {
		case "http", "https":
			return HTTPFileSource{URL: value}, true
		case "file":
			path, err := url.PathUnescape(parsed.Path)
			if err != nil {
				return nil, false
			}
			return LocalFileSource{Path: expandUser(filepath.FromSlash(path))}, true
		case "":
			return LocalFileSource{Path: expandUser(os.ExpandEnv(value))}, true
		default:
			return nil, false
		}
	}

	return LocalFileSource{Path: expandUser(os.ExpandEnv(value))}, true
}

func parseFileMapping(value map[string]any) (any, bool) {
	repository, _ := value["git"].(string)
	if repository == "" {
		repository, _ = value["repo"].(string)
	}
	if repository != "" {
		path, _ := value["path"].(string)
		ref, refOK := value["ref"].(string)
		if path == "" || (value["ref"] != nil && !refOK) {
			return nil, false
		}
		return GitFileSource{Repository: repository, Path: strings.TrimLeft(path, "/"), Ref: ref}, true
	}

	rawURL, _ := value["url"].(string)
	if rawURL != "" {
		parsed, err := url.Parse(rawURL)
		if err != nil {
			return nil, false
		}
		switch parsed.Scheme {
		case "http", "https":
			return HTTPFileSource{URL: rawURL}, true
		case "file":
			path, err := url.PathUnescape(parsed.Path)
			if err != nil {
				return nil, false
			}
			return LocalFileSource{Path: expandUser(filepath.FromSlash(path))}, true
		default:
			return nil, false
		}
	}

	path, _ := value["path"].(string)
	if path == "" {
		return nil, false
	}
	return LocalFileSource{Path: expandUser(os.ExpandEnv(path))}, true
}

func parseGitFileString(value string) (any, bool) {
	separatorIndex := strings.LastIndex(value, "//")
	if separatorIndex <= 0 {
		return nil, false
	}

	repository := value[:separatorIndex]
	pathAndQuery := value[separatorIndex+2:]
	path, rawQuery, _ := strings.Cut(pathAndQuery, "?")
	unescapedPath, err := url.PathUnescape(path)
	if err != nil {
		return nil, false
	}
	unescapedPath = strings.TrimLeft(unescapedPath, "/")
	if repository == "" || unescapedPath == "" {
		return nil, false
	}

	queryValues, _ := url.ParseQuery(rawQuery)
	refValues := queryValues["ref"]
	ref := ""
	if len(refValues) > 0 {
		ref = refValues[len(refValues)-1]
	}
	return GitFileSource{Repository: repository, Path: unescapedPath, Ref: ref}, true
}

func readSourceWithEnv(source any, env map[string]string) SourceReadResult {
	switch typed := source.(type) {
	case CommandSource:
		return readCommandSource(typed, env)
	case FileSource:
		return readFileSource(typed)
	default:
		return SourceReadResult{Error: "unknown source"}
	}
}

func readCommandSource(source CommandSource, env map[string]string) SourceReadResult {
	if !commandExists(source.Command) {
		return SourceReadResult{Error: fmt.Sprintf("command not found: %s", source.Command[0])}
	}

	result, err := runCommandWithEnv(source.Command, nil, env)
	if err != nil {
		return SourceReadResult{Error: fmt.Sprintf("failed to run command %s: %v", formatCommand(source.Command), err)}
	}
	if result.exitCode != 0 {
		return SourceReadResult{Error: processError(source.Command, result)}
	}
	return SourceReadResult{Content: result.stdout}
}

func readFileSource(source FileSource) SourceReadResult {
	switch typed := source.File.(type) {
	case LocalFileSource:
		return readLocalFile(typed)
	case HTTPFileSource:
		return readHTTPFile(typed)
	case GitFileSource:
		return readGitFile(typed)
	default:
		return SourceReadResult{Error: "unknown file source"}
	}
}

func commandExists(command []string) bool {
	executable := command[0]
	if strings.ContainsRune(executable, os.PathSeparator) {
		_, err := os.Stat(executable)
		return err == nil
	}
	_, err := exec.LookPath(executable)
	return err == nil
}

func readLocalFile(source LocalFileSource) SourceReadResult {
	content, err := os.ReadFile(source.Path)
	if err != nil {
		return SourceReadResult{Error: fmt.Sprintf("failed to read local file %s: %v", source.Path, err)}
	}
	return SourceReadResult{Content: content}
}

func readHTTPFile(source HTTPFileSource) SourceReadResult {
	request, err := http.NewRequest(http.MethodGet, source.URL, nil)
	if err != nil {
		return SourceReadResult{Error: fmt.Sprintf("failed to read HTTP file %s: %v", source.URL, err)}
	}
	request.Header.Set("User-Agent", userAgent)

	client := &http.Client{Timeout: httpTimeout}
	response, err := client.Do(request)
	if err != nil {
		return SourceReadResult{Error: fmt.Sprintf("failed to read HTTP file %s: %v", source.URL, err)}
	}
	defer func() {
		_ = response.Body.Close()
	}()

	if response.StatusCode < 200 || response.StatusCode >= 300 {
		return SourceReadResult{Error: fmt.Sprintf("failed to read HTTP file %s: HTTP status %s", source.URL, response.Status)}
	}

	content, err := io.ReadAll(response.Body)
	if err != nil {
		return SourceReadResult{Error: fmt.Sprintf("failed to read HTTP file %s: %v", source.URL, err)}
	}
	return SourceReadResult{Content: content}
}

func readGitFile(source GitFileSource) SourceReadResult {
	if _, err := exec.LookPath("git"); err != nil {
		return SourceReadResult{Error: "command not found: git"}
	}

	tempDir, err := os.MkdirTemp("", "zcs-git-*")
	if err != nil {
		return SourceReadResult{Error: fmt.Sprintf("failed to create temp dir: %v", err)}
	}
	defer func() {
		_ = os.RemoveAll(tempDir)
	}()

	repositoryDir := filepath.Join(tempDir, "repository")
	if cloneError := cloneGitRepository(source.Repository, repositoryDir); cloneError != "" {
		return SourceReadResult{Error: cloneError}
	}

	revision := "HEAD"
	if source.Ref != "" {
		if fetchError := fetchGitRef(repositoryDir, source.Ref); fetchError != "" {
			return SourceReadResult{Error: fetchError}
		}
		revision = "FETCH_HEAD"
	}

	return showGitFile(repositoryDir, revision, source.Path)
}

func cloneGitRepository(repository string, destination string) string {
	command := []string{"git", "clone", "--depth=1", "--filter=blob:none", "--no-checkout", repository, destination}
	result, err := runCommand(command, nil)
	if err != nil {
		return fmt.Sprintf("failed to run command %s: %v", formatCommand(command), err)
	}
	if result.exitCode == 0 {
		return ""
	}

	_ = os.RemoveAll(destination)
	fallbackCommand := []string{"git", "clone", "--depth=1", "--no-checkout", repository, destination}
	fallbackResult, err := runCommand(fallbackCommand, nil)
	if err != nil {
		return fmt.Sprintf("failed to run command %s: %v", formatCommand(fallbackCommand), err)
	}
	if fallbackResult.exitCode == 0 {
		return ""
	}
	return processError(fallbackCommand, fallbackResult)
}

func fetchGitRef(repositoryDir string, ref string) string {
	command := []string{"git", "-C", repositoryDir, "fetch", "--depth=1", "origin", ref}
	result, err := runCommand(command, nil)
	if err != nil {
		return fmt.Sprintf("failed to run command %s: %v", formatCommand(command), err)
	}
	if result.exitCode == 0 {
		return ""
	}
	return processError(command, result)
}

func showGitFile(repositoryDir string, revision string, path string) SourceReadResult {
	command := []string{"git", "-C", repositoryDir, "show", fmt.Sprintf("%s:%s", revision, path)}
	result, err := runCommand(command, nil)
	if err != nil {
		return SourceReadResult{Error: fmt.Sprintf("failed to run command %s: %v", formatCommand(command), err)}
	}
	if result.exitCode != 0 {
		return SourceReadResult{Error: processError(command, result)}
	}
	return SourceReadResult{Content: result.stdout}
}

type commandResult struct {
	stdout   []byte
	stderr   []byte
	exitCode int
}

func runCommand(command []string, stdout io.Writer) (commandResult, error) {
	return runCommandWithEnv(command, stdout, nil)
}

func runCommandWithEnv(command []string, stdout io.Writer, env map[string]string) (commandResult, error) {
	var stdoutBuffer bytes.Buffer
	var stderrBuffer bytes.Buffer

	cmd := exec.Command(command[0], command[1:]...)
	if len(env) > 0 {
		cmd.Env = os.Environ()
		for key, value := range env {
			cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", key, value))
		}
	}
	if stdout == nil {
		cmd.Stdout = &stdoutBuffer
	} else {
		cmd.Stdout = stdout
	}
	cmd.Stderr = &stderrBuffer

	err := cmd.Run()
	exitCode := 0
	if err != nil {
		exitError, ok := err.(*exec.ExitError)
		if !ok {
			return commandResult{}, err
		}
		exitCode = exitError.ExitCode()
	}

	return commandResult{
		stdout:   stdoutBuffer.Bytes(),
		stderr:   stderrBuffer.Bytes(),
		exitCode: exitCode,
	}, nil
}

func processError(command []string, result commandResult) string {
	message := fmt.Sprintf("command failed with exit code %d: %s", result.exitCode, formatCommand(command))
	stderr := strings.TrimSpace(string(result.stderr))
	if stderr != "" {
		message = fmt.Sprintf("%s; %s", message, stderr)
	}
	return message
}

func formatCommand(command []string) string {
	return strings.Join(command, " ")
}

func expandUser(path string) string {
	if path == "" || path == "~" || !strings.HasPrefix(path, "~/") {
		if path == "~" {
			home, err := os.UserHomeDir()
			if err == nil {
				return home
			}
		}
		return path
	}

	home, err := os.UserHomeDir()
	if err != nil {
		return path
	}
	return filepath.Join(home, strings.TrimPrefix(path, "~/"))
}
