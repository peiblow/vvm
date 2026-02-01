package vm

import (
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"net"
	"sync"

	"github.com/peiblow/vvm/compiler"
	"github.com/peiblow/vvm/lexer"
	"github.com/peiblow/vvm/parser"
)

type Runtime struct {
	contracts map[string]*compiler.ContractArtifact
	mu        sync.RWMutex
}

func NewRuntime() *Runtime {
	return &Runtime{
		contracts: make(map[string]*compiler.ContractArtifact),
	}
}

type WireMessage struct {
	Type string          `json:"type"` // DEPLOY, EXEC
	ID   string          `json:"id"`   // Request ID for correlation
	Data json.RawMessage `json:"data"` // Payload (depends on Type)
}

type WireResponse struct {
	Type    string      `json:"type"`
	ID      string      `json:"id"`
	Success bool        `json:"success"`
	Data    interface{} `json:"data,omitempty"`
	Error   string      `json:"error,omitempty"`
}

type DeployRequest struct {
	Hash         string
	ContractName string
	Version      string
	Owner        string
	Source       []byte
}

type ExecRequest struct {
	ContractID string                 `json:"contract_id"`
	Function   string                 `json:"function"`
	Args       map[string]interface{} `json:"args"`
}

func (r *Runtime) HandleConnection(conn net.Conn) {
	defer conn.Close()

	var frameLength uint32
	if err := binary.Read(conn, binary.BigEndian, &frameLength); err != nil {
		fmt.Println("Error reading frame length:", err)
		return
	}

	payload := make([]byte, frameLength)
	if _, err := io.ReadFull(conn, payload); err != nil {
		fmt.Println("Error reading payload:", err)
		return
	}

	var msg WireMessage
	if err := json.Unmarshal(payload, &msg); err != nil {
		fmt.Println("Error unmarshaling message:", err)
		return
	}

	response := r.processMessage(&msg)

	// Serializing response
	respBytes, err := json.Marshal(response)
	if err != nil {
		fmt.Println("Error marshaling response:", err)
		return
	}

	respLength := uint32(len(respBytes))
	if err := binary.Write(conn, binary.BigEndian, respLength); err != nil {
		fmt.Println("Error writing response length:", err)
		return
	}

	if _, err := conn.Write(respBytes); err != nil {
		fmt.Println("Error writing response payload:", err)
		return
	}
}

func (r *Runtime) processMessage(msg *WireMessage) WireResponse {
	switch msg.Type {
	case "DEPLOY":
		return r.handleDeploy(msg)
	case "EXEC":
		return r.handleExec(msg)
	case "PING":
		return WireResponse{
			Type:    "PONG",
			ID:      msg.ID,
			Success: true,
		}
	default:
		return WireResponse{
			Type:    "ERROR",
			ID:      msg.ID,
			Success: false,
			Error:   fmt.Sprintf("unknown message type: %s", msg.Type),
		}
	}
}

func (r *Runtime) handleDeploy(msg *WireMessage) WireResponse {
	var req DeployRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return WireResponse{
			Type:    "DEPLOY_RESPONSE",
			ID:      msg.ID,
			Success: false,
			Error:   fmt.Sprintf("invalid deploy request: %v", err),
		}
	}

	tokens := lexer.Tokenize(string(req.Source))
	ast := parser.Parse(tokens)

	cmpl := compiler.New()
	cmpl.CompileBlock(ast)

	artifact := cmpl.Artifact()

	initVM := NewFromArtifact(artifact)
	initResult := initVM.Run()
	if !initResult.Success {
		return WireResponse{
			Type:    "DEPLOY_RESPONSE",
			ID:      msg.ID,
			Success: false,
			Error:   fmt.Sprintf("initialization failed: %v", initResult.Error),
		}
	}

	artifact.InitStorage = initVM.GetStorage()

	r.mu.Lock()
	r.contracts[req.Hash] = artifact
	r.mu.Unlock()

	fmt.Println("Successful Deploy! ", req.Hash)

	return WireResponse{
		Type:    "DEPLOY_RESPONSE",
		ID:      msg.ID,
		Success: true,
		Data: map[string]interface{}{
			"contract_hash":    req.Hash,
			"contract_version": req.Version,
			"contract_name":    req.ContractName,
			"functions":        getFunctionNames(artifact),
			"agents":           getAgentsNames(artifact),
		},
	}
}

func (r *Runtime) handleExec(msg *WireMessage) WireResponse {
	var req ExecRequest
	if err := json.Unmarshal(msg.Data, &req); err != nil {
		return WireResponse{
			Type:    "EXEC_RESPONSE",
			ID:      msg.ID,
			Success: false,
			Error:   fmt.Sprintf("invalid exec request: %v", err),
		}
	}

	r.mu.RLock()
	artifact, exists := r.contracts[req.ContractID]
	r.mu.RUnlock()

	if !exists {
		return WireResponse{
			Type:    "EXEC_RESPONSE",
			ID:      msg.ID,
			Success: false,
			Error:   fmt.Sprintf("contract '%s' not found", req.ContractID),
		}
	}

	vm := NewFromArtifact(artifact)
	result := vm.RunFunction(req.Function, req.Args)

	if !result.Success {
		return WireResponse{
			Type:    "EXEC_RESPONSE",
			ID:      msg.ID,
			Success: false,
			Error:   result.Error.Error(),
		}
	}

	return WireResponse{
		Type:    "EXEC_RESPONSE",
		ID:      msg.ID,
		Success: true,
		Data: map[string]interface{}{
			"journal": result.Journal,
		},
	}
}

func (r *Runtime) GetContract(contractID string) (*compiler.ContractArtifact, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	artifact, exists := r.contracts[contractID]
	return artifact, exists
}

func getFunctionNames(artifact *compiler.ContractArtifact) []string {
	names := make([]string, 0, len(artifact.Functions))
	for name := range artifact.Functions {
		names = append(names, name)
	}
	return names
}

func getAgentsNames(artifact *compiler.ContractArtifact) []string {
	return []string{"agent1", "agent2"}
}
