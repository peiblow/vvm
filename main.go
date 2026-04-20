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
			contract, err := os.ReadFile("contracts/governance_contract.snx")
			if err != nil {
				fmt.Println("Error reading contract file:", err)
				return
			}
			// Simulate a DEPLOY request
			deployReq := vm.DeployRequest{
				Hash:         "mockhash",
				ContractName: "SynxGovernance",
				Version:      "2.0.0",
				Owner:        "0xA1B2C3D4",
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
				Function:         "creditDecision",
				Args: map[string]interface{}{
					"request": map[string]interface{}{
						"client":      "0xCLIENT001",
						"model_id":    "credit_model_v1",
						"score":       750,
						"income":      80000,
						"debtPercent": 20,
						"amount":      50000,
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
