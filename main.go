package main

import (
	"encoding/json"
	"fmt"
	"net"
	"os"

	"github.com/peiblow/vvm/compiler"
	"github.com/peiblow/vvm/vm"
)

func main() {
	runtime := vm.NewRuntime()

	if len(os.Args) > 1 && os.Args[1] == "local" {
		fmt.Println("Running in local mode with mock runtime")
		func() {
			contract, err := os.ReadFile("contracts/00.snx")
			if err != nil {
				fmt.Println("Error reading contract file:", err)
				return
			}
			// Simulate a DEPLOY request
			deployReq := vm.DeployRequest{
				Hash:         "mockhash",
				ContractName: "MockContract",
				Version:      "0.1.0",
				Owner:        "0xMockOwner",
				Source:       []byte(contract),
			}

			msg := vm.WireMessage{
				Type: "DEPLOY",
				ID:   "mockdeploy1",
				Data: func() json.RawMessage {
					data, _ := json.Marshal(deployReq)
					return data
				}(),
			}
			deployRes := runtime.HandleDeploy(&msg)

			if !deployRes.Success {
				fmt.Printf("DEPLOY failed: %s\n", deployRes.Error)
				return
			}

			contractArtifactRes := deployRes.Data.(map[string]interface{})["contract_artifact"]

			artifactBytes, err := json.Marshal(contractArtifactRes.(*compiler.ContractArtifact))
			if err != nil {
				fmt.Println("Error marshaling artifact:", err)
				return
			}

			execReq := vm.ExecRequest{
				ArtifactHash:     "mockhash",
				ContractArtifact: json.RawMessage(artifactBytes),
				Function:         "pow",
				Args: map[string]interface{}{
					"input": map[string]interface{}{
						"commitment_id": "mock_commitment_id",
						"tx_hash":       "mock_tx_hash",
						"amount":        1000,
						"timestamp":     1234567890,
						"uniqueNonce":   "mock_unique_nonce",
					},
				},
			}
			execMsg := vm.WireMessage{
				Type: "EXEC",
				ID:   "mockexec1",
				Data: func() json.RawMessage {
					data, _ := json.Marshal(execReq)
					return data
				}(),
			}
			execRes := runtime.HandleExec(&execMsg)
			fmt.Printf("EXEC response: %+v\n", execRes)
		}()

		return
	}

	ln, err := net.Listen("tcp", ":8332")
	if err != nil {
		fmt.Println("Error starting server:", err)
		return
	}
	defer ln.Close()

	fmt.Println("VVM Runtime listening on :8332")

	for {
		conn, err := ln.Accept()
		if err != nil {
			fmt.Println("Error accepting connection:", err)
			continue
		}

		go runtime.HandleConnection(conn)
	}
}
