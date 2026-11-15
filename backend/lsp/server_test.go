package lsp

import (
	"bufio"
	"bytes"
	"encoding/json"
	"io"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

const TEST_DATA_FOLDER = "testdata/data"

func TestDiagnose(test *testing.T) {
	type expectedDiagnostic struct {
		path string
		line int
		severity int
		// Empty for messages that come from Lua itself
		messageContains string
	}

	expectedDiagnostics := []expectedDiagnostic {
		{ path: "items/broken.lua", line: 5, severity: DIAGNOSTIC_SEVERITY_ERROR },
		{ path: "rooms/attic.lua", line: 2, severity: DIAGNOSTIC_SEVERITY_ERROR, messageContains: "invalid type for 'name'" },
		{ path: "rooms/cellar.lua", line: 6, severity: DIAGNOSTIC_SEVERITY_ERROR, messageContains: "item 'Broken Sword'" },
		{ path: "layout.json", line: 3, severity: DIAGNOSTIC_SEVERITY_ERROR, messageContains: "no west exit back" },
		{ path: "layout.json", line: 4, severity: DIAGNOSTIC_SEVERITY_ERROR, messageContains: "'cellar.lua', which is not a room file that loaded" },
		{ path: "layout.json", line: 5, severity: DIAGNOSTIC_SEVERITY_WARNING, messageContains: "used by more than one room" },
	}

	pathToDiagnostics, roomNames, err := diagnose(os.DirFS(TEST_DATA_FOLDER))
	if err != nil {
		test.Fatalf("unexpected error: %s", err.Error())
	}

	diagnosticCount := 0
	for _, diagnostics := range pathToDiagnostics {
		diagnosticCount += len(diagnostics)
	}
	if diagnosticCount != len(expectedDiagnostics) {
		test.Errorf("expected %d diagnostics, got %d: %+v", len(expectedDiagnostics), diagnosticCount, pathToDiagnostics)
	}

	for _, expected := range expectedDiagnostics {
		found := false
		for _, diagnostic := range pathToDiagnostics[expected.path] {
			if diagnostic.Range.Start.Line == expected.line &&
				diagnostic.Severity == expected.severity &&
				strings.Contains(diagnostic.Message, expected.messageContains) {
				found = true
				break
			}
		}
		if !found {
			test.Errorf("expected a diagnostic on %s line %d with severity %d containing '%s', got %+v",
				expected.path, expected.line, expected.severity, expected.messageContains, pathToDiagnostics[expected.path])
		}
	}

	expectedRoomNames := map[string]string { "hall.lua": "Hall" }
	if !maps.Equal(roomNames, expectedRoomNames) {
		test.Errorf("expected room names %v, got %v", expectedRoomNames, roomNames)
	}
}

func TestServe(test *testing.T) {
	absoluteDataFolder, err := filepath.Abs(TEST_DATA_FOLDER)
	if err != nil {
		test.Fatalf("unexpected error: %s", err.Error())
	}
	atticUri := "file://" + filepath.ToSlash(absoluteDataFolder) + "/rooms/attic.lua"
	layoutUri := "file://" + filepath.ToSlash(absoluteDataFolder) + "/layout.json"

	// All input up front. The server handles it in order, so the attic fix is still waiting out the debounce
	// when the room names request arrives
	var input bytes.Buffer
	inputMessages := []string {
		`{"jsonrpc":"2.0","id":1,"method":"initialize","params":{}}`,
		`{"jsonrpc":"2.0","method":"initialized","params":{}}`,
		`{"jsonrpc":"2.0","method":"textDocument/didOpen","params":{"textDocument":{"uri":"` + atticUri + `","text":"return { name = \"Attic\", description = \"Fixed\" }"}}}`,
		`{"jsonrpc":"2.0","id":2,"method":"mud/roomNames"}`,
		`{"jsonrpc":"2.0","id":3,"method":"unknown/method"}`,
		`{"jsonrpc":"2.0","id":4,"method":"shutdown"}`,
		`{"jsonrpc":"2.0","method":"exit"}`,
	}
	for _, message := range inputMessages {
		err = writeMessage(&input, json.RawMessage(message))
		if err != nil {
			test.Fatalf("unexpected error: %s", err.Error())
		}
	}

	var output bytes.Buffer
	exitCode := serve(TEST_DATA_FOLDER, &input, &output)
	if exitCode != 0 {
		test.Errorf("expected exit code 0, got %d", exitCode)
	}

	// Read every message the server wrote
	responses := make(map[string]*rpcMessage)
	uriToPublishes := make(map[string][][]Diagnostic)
	reader := bufio.NewReader(&output)
	for {
		message, err := readMessage(reader)
		if err == io.EOF {
			break
		}
		if err != nil {
			test.Fatalf("unexpected error reading output: %s", err.Error())
		}

		if message.Method == "textDocument/publishDiagnostics" {
			var params publishDiagnosticsParams
			err = json.Unmarshal(message.Params, &params)
			if err != nil {
				test.Fatalf("invalid publishDiagnostics params: %s", err.Error())
			}
			uriToPublishes[params.Uri] = append(uriToPublishes[params.Uri], params.Diagnostics)
			continue
		}
		responses[string(message.Id)] = message
	}

	// Initialize
	initializeResponse, exists := responses["1"]
	if !exists {
		test.Fatalf("no response to initialize")
	}
	var initialize initializeResult
	err = json.Unmarshal(initializeResponse.Result, &initialize)
	if err != nil {
		test.Fatalf("invalid initialize result: %s", err.Error())
	}
	if !initialize.Capabilities.TextDocumentSync.OpenClose || initialize.Capabilities.TextDocumentSync.Change != 1 {
		test.Errorf("expected open/close and full sync, got %+v", initialize.Capabilities.TextDocumentSync)
	}

	// Diagnostics. The attic is published with its error, then cleared once its fix is validated
	atticPublishes := uriToPublishes[atticUri]
	if len(atticPublishes) != 2 || len(atticPublishes[0]) != 1 || len(atticPublishes[1]) != 0 {
		test.Errorf("expected the attic to be published with one error and then cleared, got %+v", atticPublishes)
	}
	if len(uriToPublishes[layoutUri]) == 0 {
		test.Errorf("expected diagnostics on %s, got none", layoutUri)
	}

	// Room names include the fixed attic, since the pending validation runs before answering
	roomNamesResponse, exists := responses["2"]
	if !exists {
		test.Fatalf("no response to mud/roomNames")
	}
	var roomNames map[string]string
	err = json.Unmarshal(roomNamesResponse.Result, &roomNames)
	if err != nil {
		test.Fatalf("invalid mud/roomNames result: %s", err.Error())
	}
	expectedRoomNames := map[string]string { "hall.lua": "Hall", "attic.lua": "Attic" }
	if !maps.Equal(roomNames, expectedRoomNames) {
		test.Errorf("expected room names %v, got %v", expectedRoomNames, roomNames)
	}

	// Unknown request
	unknownResponse, exists := responses["3"]
	if !exists {
		test.Fatalf("no response to the unknown request")
	}
	if unknownResponse.Error == nil || unknownResponse.Error.Code != RPC_ERROR_METHOD_NOT_FOUND {
		test.Errorf("expected a method not found error, got %+v", unknownResponse.Error)
	}

	// Shutdown
	shutdownResponse, exists := responses["4"]
	if !exists {
		test.Fatalf("no response to shutdown")
	}
	if string(shutdownResponse.Result) != "null" {
		test.Errorf("expected a null shutdown result, got %s", string(shutdownResponse.Result))
	}
}

func TestServeExitWithoutShutdown(test *testing.T) {
	var input bytes.Buffer
	err := writeMessage(&input, json.RawMessage(`{"jsonrpc":"2.0","method":"exit"}`))
	if err != nil {
		test.Fatalf("unexpected error: %s", err.Error())
	}

	exitCode := serve(TEST_DATA_FOLDER, &input, io.Discard)
	if exitCode != 1 {
		test.Errorf("expected exit code 1, got %d", exitCode)
	}
}
