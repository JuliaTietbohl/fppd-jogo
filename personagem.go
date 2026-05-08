// personagem.go - Funções para movimentação e ações do personagem
package main

import "fmt"

// Atualiza a posição do personagem com base na tecla pressionada (WASD)
func personagemMover(tecla rune, jogo *Jogo) {
	dx, dy := 0, 0
	switch tecla {
	case 'w': dy = -1 // Move para cima
	case 'a': dx = -1 // Move para a esquerda
	case 's': dy = 1  // Move para baixo
	case 'd': dx = 1  // Move para a direita
	}

	nx, ny := jogo.PosX+dx, jogo.PosY+dy
	// Verifica se o movimento é permitido e realiza a movimentação
	if jogoPodeMoverPara(jogo, nx, ny) {
		jogo.Mu.Lock()  
		jogoMoverElemento(jogo, jogo.PosX, jogo.PosY, dx, dy)
		jogo.PosX, jogo.PosY = nx, ny
		jogo.Mu.Unlock()    
	}
}

// Define o que ocorre quando o jogador pressiona a tecla de interação
// Neste exemplo, apenas exibe uma mensagem de status
// Você pode expandir essa função para incluir lógica de interação com objetos
// UTILIZADO PARA PEGAR VIDA EXTRA
func personagemInteragir(jogo *Jogo) {
	if jogo.VidaExtraDisponivel &&
		jogo.UltimoVisitado.simbolo == VidaExtra.simbolo {
		
		jogo.Mu.Lock() 
		jogo.UltimoVisitado = Vazio
		jogo.VidaExtraDisponivel = false
		jogo.Mu.Unlock()

		if jogo.HP < jogo.MaxHP {
			jogo.HP++
			jogo.StatusMsg = fmt.Sprintf("Vida extra coletada! HP: %d/%d", jogo.HP, jogo.MaxHP)
		} else {
			jogo.StatusMsg = "HP ja esta cheio!"
		}
	} else {
		jogo.StatusMsg = fmt.Sprintf("Interagindo em (%d, %d)", jogo.PosX, jogo.PosY)
	}
}

// Processa o evento do teclado e executa a ação correspondente
func personagemExecutarAcao(ev EventoTeclado, jogo *Jogo) bool {
	switch ev.Tipo {
	case "sair":
		// Retorna false para indicar que o jogo deve terminar
		return false
	case "interagir":
		// Executa a ação de interação
		personagemInteragir(jogo)
	case "mover":
		// Move o personagem com base na tecla
		personagemMover(ev.Tecla, jogo)
	}
	return true // Continua o jogo
}
