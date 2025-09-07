package main

const (
	OP_CONST   = 0x01
	OP_PUSH    = 0x02 // empilha valor
	OP_ADD     = 0x03 // soma valores do topo da stack
	OP_SUB     = 0x04 // subtração
	OP_MUL     = 0x05 // multiplicação
	OP_DIV     = 0x06 // divisão
	OP_GT      = 0x07 // Maior que
	OP_GT_EQ   = 0x08 // Maior igual que
	OP_LT      = 0x09 // Menor que
	OP_LT_EQ   = 0x0A // Menor igual que
	OP_EQ      = 0x0B // igual a
	OP_DIFF    = 0x0C // igual a
	OP_PLUS_EQ = 0x0D
	OP_POP     = 0x0E // remove topo da stack
	OP_DUP     = 0x0F // duplica topo da stack
	OP_SWAP    = 0x10 // troca dois valores do topo
	OP_PRINT   = 0x11 // imprime valor do topo
	OP_NOP     = 0x12 // instrução nula

	OP_JMP    = 0x13 // salto incondicional
	OP_JMP_IF = 0x14 // salto condicional
	OP_CALL   = 0x15 // chamada de função
	OP_RET    = 0x16 // retorno de função
	OP_ACCESS = 0x17 // access a especific array item
	OP_NULL   = 0x18 // null
	OP_LENGTH = 0x19 // retorna o tamanho de um array

	OP_STORE  = 0x1A // armazena valor na storage
	OP_SLOAD  = 0x1B // carrega valor da storage
	OP_MSTORE = 0x1C // armazena valor na memoria
	OP_MLOAD  = 0x1D //carrega valor da memoria
	OP_DELETE = 0x1E // remove valor da storage

	OP_ML_RUN     = 0x40 // executa modelo de ML
	OP_ML_FEED    = 0x41 // envia dados para treinamento
	OP_ML_AGG     = 0x42 // agrega gradientes / atualização de modelo
	OP_ML_PREDICT = 0x43 // faz predição
	OP_ML_SCORE   = 0x44 // calcula métricas ou loss
	OP_ML_SAVE    = 0x45 // salva modelo
	OP_ML_LOAD    = 0x46 // carrega modelo

	OP_REWARD       = 0x50 // distribui recompensa
	OP_EMIT         = 0x51 // emite evento
	OP_TRANSFER     = 0x52 // transfere valor entre contas
	OP_BALANCE_OF   = 0x53 // verifica saldo de endereço
	OP_REQUIRE      = 0x54 // verifica saldo de endereço
	OP_GET_PROPERTY = 0x55

	OP_HALT = 0x00 // Fim do programa
)
