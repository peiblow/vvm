# Synx Lang

Synx governs decisions — it does not replace programming languages.

Synx is a declarative governance and execution language designed to validate, govern, and auditablely record external decisions in a deterministic and controlled manner before irreversible actions occur.

This project explores language design, compilers, and virtual machines with a focus on blockchain/smart contract primitives.

---

## ✨ Features

### Language Constructs

- **Contracts** - Top-level container for all declarations
- **Registries** - Persistent on-chain model declarations with version, owner, and purpose
- **Agents** - Validated entities tied to registries with hash verification
- **Policies** - Rule definitions with typed properties (e.g., credit limits, score ranges)
- **Custom Types** - User-defined structured types with typed fields
- **Functions** - Named functions with typed parameters and return types
- **Events** - Emit blockchain-style events with typed payloads

### Control Flow

- `if` / `else` conditionals
- `for` loops with initialization, condition, and increment
- `while` loops
- `require` statements for assertions (reverts on failure)
- `return` for function exits

### Compiler

- Transforms AST into compact **bytecode**
- Supports 2-byte jump addresses for programs up to 64KB
- Constant pool for strings, numbers, arrays, and objects
- Symbol table for variable/slot management
- Function metadata with argument tracking and type information
- **ContractArtifact** - Serializable compilation output with bytecode, functions, types, and initialized storage

### Virtual Machine (VM)

- **Stack-based** architecture
- **Call stack** for function calls and returns
- **Storage** system for runtime variables with pre-initialization support
- **Journal** for event logging with SHA-256 hashes
- **Function execution** - Execute specific functions by name with `RunFunction()`
- **InitStorage** - Pre-load immutable state (policies, registries) from deployed contracts
- Built-in operations: arithmetic, comparison, logical, array access

### Supported Types

- `UInt`, `String`, `Address`, `bool`
- Arrays (`[]interface{}`)
- Objects/Maps (`map[string]interface{}`)
- Custom user-defined types

---

## 📄 Example

```synx
contract Synx {
  registry Model CreditScoreFL {
    version: "1.0.0"
    owner: 0xABC123FF
    purpose: "credit_scoring"
  }

  agent CreditScoreFL {
    hash: 0xdfc2c348e0d71685ebfa1a6a999cbad256dccab83a4d66429c1fe504a4e81861
    version: "1.0.0"
    owner: 0xABC123FF
  }

  policy CreditPolicy {
    minScore: 700
    maxAmount: 100000
    ranges: [
      { min: 300, max: 599, limit: 1000 },
      { min: 600, max: 699, limit: 5000 },
      { min: 700, max: 799, limit: 20000 },
      { min: 800, max: 900, limit: 100000 }
    ]
  }

  type Decision {
    model_id: String
    client: Address
    score: UInt
    amount: UInt
  }

  fn getLimit(score: UInt): UInt {
    for (i = 0; i < len(CreditPolicy.ranges); i = i + 1) {
      range = CreditPolicy.ranges[i]
      if (score >= range.min && score <= range.max) {
        return range.limit
      }
    }
    return 0
  }

  fn approve(decision: Decision): bool {
    limit = getLimit(decision.score)
    require(decision.amount <= limit; "Amount exceeds limit")
    require(decision.score >= CreditPolicy.minScore; "Score too low")

    emit("CreditApproved", {
      client: decision.client,
      amount: decision.amount
    })
  }

  approve({
    model_id: "CreditScoreFL",
    client: 0xDEF456FF,
    score: 700,
    amount: 10000
  })
}
```

---

## 🚀 Running

```bash
# Deploy a contract (compiles and initializes)
go run . deploy

# Execute a specific function from a deployed contract
go run . exec <function_name> [args...]

# Example output for deploy:
# Registry '{CreditScoreFL}' created with hash: 0xdfc2c348...
# Agent 'CreditScoreFL' validated successfully
# Contract deployed successfully.
```

### Execution Model

The VM supports a **two-phase execution model**:

1. **Deploy Phase** - Compiles the contract, runs initialization code (registries, agents, policies, types), and captures the initialized storage state into the `ContractArtifact`.

2. **Exec Phase** - Executes a specific function by name with pre-loaded storage. Policies, registries, and types are immutable and available from the artifact.

```go
// Deploy: compile and initialize
artifact := compiler.Artifact()
vm := vm.NewFromArtifact(artifact)
vm.Run() // Runs initialization
artifact.InitStorage = vm.GetStorage() // Capture state

// Exec: run specific function with initialized state
vm.RunFunction("approve", decisionObject)
```

---

## 🏗️ Architecture

```
┌─────────────┐    ┌──────────┐    ┌──────────┐    ┌─────┐
│ Source Code │ -> │  Lexer   │ -> │  Parser  │ -> │ AST │
│   (.snx)    │    │ (tokens) │    │          │    │     │
└─────────────┘    └──────────┘    └──────────┘    └──┬──┘
                                                      │
                   ┌──────────┐    ┌──────────┐       │
                   │    VM    │ <- │ Compiler │ <─────┘
                   │ (execute)│    │(bytecode)│
                   └──────────┘    └──────────┘
```

### Project Structure

```
vvm/
├── main.go           # Entry point (deploy/exec modes)
├── commiter/         # Journal commit handlers
│   └── commiter.go
├── lexer/            # Tokenizer
│   ├── lexer.go
│   └── token.go
├── parser/           # AST generation
│   ├── parser.go
│   ├── expr.go
│   ├── stmt.go
│   └── types.go
├── ast/              # AST node definitions
│   ├── ast.go
│   ├── expressions.go
│   ├── statements.go
│   └── types.go
├── compiler/         # Bytecode generation
│   ├── compiler.go
│   ├── opcodes.go
│   ├── expr.go
│   ├── stmt.go
│   └── debug.go
└── vm/               # Virtual machine
    └── vm.go
```

---

## 📋 Opcodes

| Category   | Opcodes                                              |
| ---------- | ---------------------------------------------------- |
| Stack      | `CONST`, `PUSH`, `POP`, `DUP`, `SWAP`                |
| Arithmetic | `ADD`, `SUB`, `MUL`, `DIV`                           |
| Comparison | `GT`, `GT_EQ`, `LT`, `LT_EQ`, `EQ`, `DIFF`           |
| Control    | `JMP`, `JMP_IF`, `CALL`, `RET`, `HALT`               |
| Storage    | `STORE`, `SLOAD`, `DELETE`                           |
| Objects    | `PUSH_OBJECT`, `SET_PROPERTY`, `GET_PROPERTY`        |
| Arrays     | `ACCESS`, `LENGTH`                                   |
| Registry   | `REGISTRY_DECLARE`, `REGISTRY_GET`, `AGENT_VALIDATE` |
| Events     | `EMIT`, `ERR`, `REQUIRE`                             |
| I/O        | `PRINT`                                              |

---

## License

Apache License 2.0

Copyright (c) 2024–2026 Synx

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
