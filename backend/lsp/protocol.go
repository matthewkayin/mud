package lsp

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io"
	"strconv"
	"strings"
)

const RPC_ERROR_METHOD_NOT_FOUND = -32601

const DIAGNOSTIC_SEVERITY_ERROR = 1
const DIAGNOSTIC_SEVERITY_WARNING = 2

// A request, notification or response. A request has an id and a method, a notification only a method,
// and a response an id and either a result or an error
type rpcMessage struct {
	Jsonrpc string `json:"jsonrpc"`
	Id json.RawMessage `json:"id,omitempty"`
	Method string `json:"method,omitempty"`
	Params json.RawMessage `json:"params,omitempty"`
	Result json.RawMessage `json:"result,omitempty"`
	Error *rpcError `json:"error,omitempty"`
}

type rpcError struct {
	Code int `json:"code"`
	Message string `json:"message"`
}

type rpcResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Id json.RawMessage `json:"id"`
	// Always written, since a successful response must have a result even when it's null
	Result any `json:"result"`
}

type rpcErrorResponse struct {
	Jsonrpc string `json:"jsonrpc"`
	Id json.RawMessage `json:"id"`
	Error *rpcError `json:"error"`
}

type rpcNotification struct {
	Jsonrpc string `json:"jsonrpc"`
	Method string `json:"method"`
	Params any `json:"params"`
}

type textDocumentItem struct {
	Uri string `json:"uri"`
	Text string `json:"text"`
}

type textDocumentIdentifier struct {
	Uri string `json:"uri"`
}

type didOpenParams struct {
	TextDocument textDocumentItem `json:"textDocument"`
}

type contentChangeEvent struct {
	Text string `json:"text"`
}

type didChangeParams struct {
	TextDocument textDocumentIdentifier `json:"textDocument"`
	ContentChanges []contentChangeEvent `json:"contentChanges"`
}

type didCloseParams struct {
	TextDocument textDocumentIdentifier `json:"textDocument"`
}

// Line and Character are 0-based. Character counts UTF-16 code units
type Position struct {
	Line int `json:"line"`
	Character int `json:"character"`
}

type Range struct {
	Start Position `json:"start"`
	End Position `json:"end"`
}

type Diagnostic struct {
	Range Range `json:"range"`
	Severity int `json:"severity"`
	Source string `json:"source"`
	Message string `json:"message"`
}

type publishDiagnosticsParams struct {
	Uri string `json:"uri"`
	Diagnostics []Diagnostic `json:"diagnostics"`
}

// Reads one message framed with a Content-Length header
func readMessage(reader *bufio.Reader) (*rpcMessage, error) {
	contentLength := -1
	for {
		line, err := reader.ReadString('\n')
		if err != nil {
			return nil, err
		}
		line = strings.TrimRight(line, "\r\n")

		// A blank line ends the headers
		if line == "" {
			break
		}

		name, value, hasValue := strings.Cut(line, ":")
		if !hasValue {
			return nil, fmt.Errorf("invalid header line '%s'", line)
		}
		if strings.EqualFold(strings.TrimSpace(name), "Content-Length") {
			contentLength, err = strconv.Atoi(strings.TrimSpace(value))
			if err != nil {
				return nil, fmt.Errorf("invalid Content-Length '%s'", value)
			}
		}
	}
	if contentLength < 0 {
		return nil, fmt.Errorf("message has no Content-Length header")
	}

	body := make([]byte, contentLength)
	_, err := io.ReadFull(reader, body)
	if err != nil {
		return nil, err
	}

	var message rpcMessage
	err = json.Unmarshal(body, &message)
	if err != nil {
		return nil, fmt.Errorf("invalid message body: %s", err.Error())
	}

	return &message, nil
}

func writeMessage(writer io.Writer, message any) error {
	body, err := json.Marshal(message)
	if err != nil {
		return err
	}

	_, err = fmt.Fprintf(writer, "Content-Length: %d\r\n\r\n", len(body))
	if err != nil {
		return err
	}
	_, err = writer.Write(body)
	return err
}
