package lsp

import (
	"bufio"
	"encoding/json"
	"io"
	"io/fs"
	"log"
	"mud/world"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"time"
)

const VALIDATE_DEBOUNCE = 300 * time.Millisecond
const DIAGNOSTIC_SOURCE = "mud"

// Owned by the server loop. Nothing else touches it
type server struct {
	// Absolute path to the data folder
	dataFolder string
	overlay *overlayFS
	output io.Writer

	shutdownRequested bool
	validateTimer *time.Timer
	// Set while a change is waiting out the debounce
	validatePending bool
	// From the last validation, see world.World.RoomFileNames
	roomNames map[string]string
	// URIs that had diagnostics in the last publish, so they can be cleared when their problems are fixed
	publishedUris map[string]bool
}

type initializeResult struct {
	Capabilities serverCapabilities `json:"capabilities"`
	ServerInfo serverInfo `json:"serverInfo"`
}

type serverCapabilities struct {
	TextDocumentSync textDocumentSyncOptions `json:"textDocumentSync"`
}

type textDocumentSyncOptions struct {
	OpenClose bool `json:"openClose"`
	// 1 is full document sync
	Change int `json:"change"`
	Save struct{} `json:"save"`
}

type serverInfo struct {
	Name string `json:"name"`
}

// Runs the language server over stdin and stdout until the client exits. Returns the process exit code
func Run(dataFolder string) int {
	return serve(dataFolder, os.Stdin, os.Stdout)
}

func serve(dataFolder string, input io.Reader, output io.Writer) int {
	absoluteDataFolder, err := filepath.Abs(dataFolder)
	if err != nil {
		log.Printf("Error resolving data folder %s: %s", dataFolder, err.Error())
		return 1
	}

	lspServer := &server {
		dataFolder: absoluteDataFolder,
		overlay: newOverlayFS(os.DirFS(absoluteDataFolder)),
		output: output,
		publishedUris: make(map[string]bool),
		roomNames: make(map[string]string),
	}

	// Stopped until the first change
	lspServer.validateTimer = time.NewTimer(VALIDATE_DEBOUNCE)
	lspServer.validateTimer.Stop()

	messages := make(chan *rpcMessage, 64)
	go readMessages(input, messages)

	log.Printf("Language server started for data folder %s.", absoluteDataFolder)
	return lspServer.run(messages)
}

// Sends every message read from the input to the channel, and closes it when the input ends or can't be read
func readMessages(input io.Reader, messages chan<- *rpcMessage) {
	reader := bufio.NewReader(input)
	for {
		message, err := readMessage(reader)
		if err != nil {
			if err != io.EOF {
				log.Printf("Error reading message: %s", err.Error())
			}
			close(messages)
			return
		}
		messages <- message
	}
}

func (lspServer *server) run(messages <-chan *rpcMessage) int {
	for {
		select {
		case message, ok := <-messages:
			if !ok {
				log.Printf("Input closed before exit.")
				return 1
			}
			if message.Method == "exit" {
				if lspServer.shutdownRequested {
					return 0
				}
				return 1
			}
			lspServer.handleMessage(message)
		case <-lspServer.validateTimer.C:
			lspServer.validate()
		}
	}
}

func (lspServer *server) handleMessage(message *rpcMessage) {
	isRequest := len(message.Id) != 0

	switch message.Method {
	case "initialize":
		result := initializeResult {
			Capabilities: serverCapabilities {
				TextDocumentSync: textDocumentSyncOptions {
					OpenClose: true,
					Change: 1,
				},
			},
			ServerInfo: serverInfo { Name: "mud" },
		}
		lspServer.respond(message.Id, result)
	case "initialized":
		lspServer.validate()
	case "shutdown":
		lspServer.shutdownRequested = true
		lspServer.respond(message.Id, nil)
	case "textDocument/didOpen":
		var params didOpenParams
		err := json.Unmarshal(message.Params, &params)
		if err != nil {
			log.Printf("warn: invalid didOpen params: %s", err.Error())
			return
		}
		lspServer.setDocument(params.TextDocument.Uri, params.TextDocument.Text)
	case "textDocument/didChange":
		var params didChangeParams
		err := json.Unmarshal(message.Params, &params)
		if err != nil {
			log.Printf("warn: invalid didChange params: %s", err.Error())
			return
		}
		// With full sync, the last change holds the whole document
		if len(params.ContentChanges) == 0 {
			return
		}
		lspServer.setDocument(params.TextDocument.Uri, params.ContentChanges[len(params.ContentChanges) - 1].Text)
	case "textDocument/didClose":
		var params didCloseParams
		err := json.Unmarshal(message.Params, &params)
		if err != nil {
			log.Printf("warn: invalid didClose params: %s", err.Error())
			return
		}
		path, ok := lspServer.uriToDataPath(params.TextDocument.Uri)
		if !ok {
			return
		}
		delete(lspServer.overlay.documents, path)
		lspServer.scheduleValidate()
	case "textDocument/didSave", "workspace/didChangeWatchedFiles":
		lspServer.scheduleValidate()
	case "mud/roomNames":
		// Answer with names that match the current buffers
		if lspServer.validatePending {
			lspServer.validate()
		}
		lspServer.respond(message.Id, lspServer.roomNames)
	default:
		if isRequest {
			lspServer.write(rpcErrorResponse {
				Jsonrpc: "2.0",
				Id: message.Id,
				Error: &rpcError {
					Code: RPC_ERROR_METHOD_NOT_FOUND,
					Message: "method not found: " + message.Method,
				},
			})
		}
	}
}

// Overlays a document's text, if it's in the data folder, and schedules validation
func (lspServer *server) setDocument(uri string, text string) {
	path, ok := lspServer.uriToDataPath(uri)
	if !ok {
		return
	}
	lspServer.overlay.documents[path] = []byte(text)
	lspServer.scheduleValidate()
}

// Validates after the debounce time, restarting the wait if validation was already scheduled
func (lspServer *server) scheduleValidate() {
	lspServer.validatePending = true
	lspServer.validateTimer.Reset(VALIDATE_DEBOUNCE)
}

// Loads the world from the overlay and publishes diagnostics for every problem
func (lspServer *server) validate() {
	lspServer.validateTimer.Stop()
	lspServer.validatePending = false

	pathToDiagnostics, roomNames, err := diagnose(lspServer.overlay)
	if err != nil {
		log.Printf("Error loading world: %s", err.Error())
		return
	}
	lspServer.roomNames = roomNames

	// Publish
	newPublishedUris := make(map[string]bool)
	for path, diagnostics := range pathToDiagnostics {
		uri := lspServer.dataPathToUri(path)
		lspServer.publishDiagnostics(uri, diagnostics)
		newPublishedUris[uri] = true
	}

	// Clear files whose problems are fixed
	for uri := range lspServer.publishedUris {
		if !newPublishedUris[uri] {
			lspServer.publishDiagnostics(uri, []Diagnostic{})
		}
	}
	lspServer.publishedUris = newPublishedUris
}

// Loads the world from a data folder. Returns the diagnostics for its problems keyed by path relative to the data folder,
// and the name of every room file that loaded keyed by path relative to the rooms folder
func diagnose(dataFolder fs.FS) (map[string][]Diagnostic, map[string]string, error) {
	loadedWorld, problems, err := world.NewWorld(dataFolder, os.Stderr)
	if err != nil {
		return nil, nil, err
	}
	roomNames := loadedWorld.RoomFileNames
	loadedWorld.Quit()

	pathToDiagnostics := make(map[string][]Diagnostic)
	pathToSource := make(map[string][]byte)
	for _, problem := range problems {
		source, hasSource := pathToSource[problem.Path]
		if !hasSource {
			// Problems on a core folder have no source to read
			source, err = fs.ReadFile(dataFolder, problem.Path)
			if err != nil {
				source = nil
			}
			pathToSource[problem.Path] = source
		}

		severity := DIAGNOSTIC_SEVERITY_ERROR
		if problem.Severity == world.SCRIPT_PROBLEM_SEVERITY_WARNING {
			severity = DIAGNOSTIC_SEVERITY_WARNING
		}

		pathToDiagnostics[problem.Path] = append(pathToDiagnostics[problem.Path], Diagnostic {
			Range: problemRange(problem, source),
			Severity: severity,
			Source: DIAGNOSTIC_SOURCE,
			Message: problem.Message,
		})
	}

	return pathToDiagnostics, roomNames, nil
}

func (lspServer *server) publishDiagnostics(uri string, diagnostics []Diagnostic) {
	lspServer.write(rpcNotification {
		Jsonrpc: "2.0",
		Method: "textDocument/publishDiagnostics",
		Params: publishDiagnosticsParams {
			Uri: uri,
			Diagnostics: diagnostics,
		},
	})
}

func (lspServer *server) respond(id json.RawMessage, result any) {
	lspServer.write(rpcResponse {
		Jsonrpc: "2.0",
		Id: id,
		Result: result,
	})
}

func (lspServer *server) write(message any) {
	err := writeMessage(lspServer.output, message)
	if err != nil {
		log.Printf("Error writing message: %s", err.Error())
	}
}

// Returns the slash path of a file URI relative to the data folder. Not ok if it isn't a file in the data folder
func (lspServer *server) uriToDataPath(uri string) (string, bool) {
	parsedUri, err := url.Parse(uri)
	if err != nil || parsedUri.Scheme != "file" {
		return "", false
	}

	relativePath, err := filepath.Rel(lspServer.dataFolder, filepath.FromSlash(parsedUri.Path))
	if err != nil {
		return "", false
	}
	relativePath = filepath.ToSlash(relativePath)
	if relativePath == ".." || strings.HasPrefix(relativePath, "../") {
		return "", false
	}

	return relativePath, true
}

func (lspServer *server) dataPathToUri(path string) string {
	absolutePath := filepath.Join(lspServer.dataFolder, filepath.FromSlash(path))
	fileUrl := url.URL {
		Scheme: "file",
		Path: filepath.ToSlash(absolutePath),
	}
	return fileUrl.String()
}
