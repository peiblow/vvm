package main

const (
	OP_PUSH  = 0x01 // empilha valor
	OP_ADD   = 0x02 // soma valores do topo da stack
	OP_SUB   = 0x03 // subtração
	OP_MUL   = 0x04 // multiplicação
	OP_DIV   = 0x05 // divisão
	OP_POP   = 0x06 // remove topo da stack
	OP_DUP   = 0x07 // duplica topo da stack
	OP_SWAP  = 0x08 // troca dois valores do topo
	OP_PRINT = 0x09 // imprime valor do topo
	OP_NOP   = 0x0A // instrução nula

	OP_JMP    = 0x14 // salto incondicional
	OP_JMP_IF = 0x15 // salto condicional
	OP_CALL   = 0x16 // chamada de função
	OP_RET    = 0x17 // retorno de função

	OP_STORE  = 0x20 // armazena valor na storage
	OP_SLOAD  = 0x21 // carrega valor da storage
	OP_MSTORE = 0x22 // armazena valor na memoria
	OP_MLOAD  = 0x23 //carrega valor da memoria
	OP_DELETE = 0x24 // remove valor da storage

	OP_ML_RUN     = 0x40 // executa modelo de ML
	OP_ML_FEED    = 0x41 // envia dados para treinamento
	OP_ML_AGG     = 0x42 // agrega gradientes / atualização de modelo
	OP_ML_PREDICT = 0x43 // faz predição
	OP_ML_SCORE   = 0x44 // calcula métricas ou loss
	OP_ML_SAVE    = 0x45 // salva modelo
	OP_ML_LOAD    = 0x46 // carrega modelo

	OP_REWARD   = 0x50 // distribui recompensa
	OP_EMIT     = 0x51 // emite evento
	OP_TRANSFER = 0x52 // transfere valor entre contas

	OP_HALT = 0x00 // Fim do programa
)
