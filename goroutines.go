package main

import (
	"fmt"
	"sync"
	"time"
)

//frequência do loop de jogo (inimigos se movem a cada tick)
const tickInterval = 500 * time.Millisecond

// intervalo de movimento de cada inimigo
const inimigoMoveInterval = 600 * time.Millisecond

// --- Goroutine 1

func inputLoop(inputCh chan<- EventoTeclado, doneCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		ev := interfaceLerEventoTeclado()

		select {
		case <-doneCh:
			return
		case inputCh <- ev:
			if ev.Tipo == "sair" {
				return
			}
		}
	}
}

// --- Goroutine 2
func gameLoop(
	jogo *Jogo,
	inputCh <-chan EventoTeclado,
	renderCh chan<- EstadoRender,
	inimigoCh <-chan MoveInimigo,
	moedaCh <-chan struct{},
	fimCh chan struct{},
	doneCh <-chan struct{},
) {
	tick := time.NewTicker(tickInterval)
	defer tick.Stop()

	enviarSnapshot(jogo, renderCh)
	var fechouFim sync.Once
	for {
		select {

		// Case 1: encerramento solicitado externamente
		case <-doneCh:
			return

		// Case 2: tick periódico — atualiza mensagem de status com HP
		case <-tick.C:
    		enviarSnapshot(jogo, renderCh)

		// Case 3: movimento de um inimigo enviado por inimigoLoop
		case mv := <-inimigoCh:
			colidiu := jogoAplicarMoveInimigo(jogo, mv)
			if colidiu {
				jogo.Mu.Lock()
				if jogo.HP <= 0 {
					jogo.StatusMsg = "Você morreu! Pressione ESC para sair."
					fechouFim.Do(func() { close(fimCh) })
				} else {
					jogo.StatusMsg = fmt.Sprintf("Você foi atingido! HP: %s (%d/%d)",
						barraHP(jogo.HP, jogo.MaxHP), jogo.HP, jogo.MaxHP)
				}
				jogo.Mu.Unlock()
			}
			enviarSnapshot(jogo, renderCh)
		
		// Case 4: moedaLoop sinalizou que todas as moedas foram coletadas
		case <-moedaCh:
			jogo.Mu.Lock()
			jogo.StatusMsg = "Você venceu! Todas as moedas foram coletadas! ESC para sair."
			fechouFim.Do(func() { close(fimCh) })
			jogo.Mu.Unlock()
			enviarSnapshot(jogo, renderCh)

		// Case 5: input do jogador
		case ev := <-inputCh:
			switch ev.Tipo {
			case "sair":
				return
			case "mover":
    			personagemMover(ev.Tecla, jogo)
				jogoColetarMoeda(jogo)
				if jogoVerificarColisaoInimigos(jogo) {
					jogo.Mu.Lock()
					if jogo.HP <= 0 {
						jogo.StatusMsg = "Você morreu! Pressione ESC para sair."
						fechouFim.Do(func() { close(fimCh) })
					} else {
						jogo.StatusMsg = fmt.Sprintf("Você foi atingido! HP: %s (%d/%d)",
							barraHP(jogo.HP, jogo.MaxHP), jogo.HP, jogo.MaxHP)
					}
					jogo.Mu.Unlock()
				}
			case "interagir":
				personagemInteragir(jogo)
			}
			enviarSnapshot(jogo, renderCh)
		}
	}
}

// --- Goroutine 3
func renderLoop(renderCh <-chan EstadoRender, doneCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
	for {
		select {
		case <-doneCh:
			return
		case estado, ok := <-renderCh:
			if !ok {
				return
			}
			interfaceDesenharEstado(estado)
		}
	}
}

// --- Goroutines 4 5 e 6
func inimigoLoop(id int, jogo *Jogo, inimigoCh chan<- MoveInimigo, fimCh <-chan struct{}, doneCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()

	// Cada inimigo tem seu próprio ticker com variação para evitar sincronismo
	intervalo := inimigoMoveInterval + time.Duration(id*80)*time.Millisecond
	ticker := time.NewTicker(intervalo)
	defer ticker.Stop()

	for {
		select {
		case <-doneCh:
			return
		case <-fimCh:
            return
		case <-ticker.C:
			jogo.Mu.RLock()
			if id >= len(jogo.Inimigos) || !jogo.Inimigos[id].Vivo {
				jogo.Mu.RUnlock()
				return
			}
			ix, iy := jogo.Inimigos[id].X, jogo.Inimigos[id].Y
			px, py := jogo.PosX, jogo.PosY
			jogo.Mu.RUnlock()

			// BFS lê jogo.Mapa
			jogo.Mu.RLock()
			dx, dy := jogoBFSNextStep(jogo, ix, iy, px, py)
			jogo.Mu.RUnlock()

			if dx == 0 && dy == 0 {
				continue
			}

			// Envia movimento para o gameLoop processar
			select {
			case inimigoCh <- MoveInimigo{ID: id, DX: dx, DY: dy}:
			case <-doneCh:
				return
			case <-fimCh:
                return
			}
		}
	}
}

// --- Helpers ---
func enviarSnapshot(jogo *Jogo, renderCh chan<- EstadoRender) {
	snap := jogoSnapshot(jogo)
	select {
	case renderCh <- snap:
	default:
		//Descarta frame antigo para evitar que o gameLoop bloqueie esperando o renderLoop
	}
}

// barraHP gera uma representação visual da barra de vida
// Light Shade character e full block (procurei no google block elements)
func barraHP(hp, maxHP int) string {
	const barLen = 10
	filled := (hp * barLen) / maxHP
	bar := "["
	for i := 0; i < barLen; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	bar += "]"
	return bar
}

// --- Goroutine 7
func moedaLoop(jogo *Jogo, moedaCh chan<- struct{}, doneCh <-chan struct{}, wg *sync.WaitGroup) {
	defer wg.Done()
 
	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()
 
	for {
		select {
		case <-doneCh:
			return
		case <-ticker.C:
			jogo.Mu.RLock()
			coletadas := jogo.MoedasColetadas
			total := jogo.TotalMoedas
			jogo.Mu.RUnlock()
 
			// Para vitória: todas as moedas coletadas
			if total > 0 && coletadas >= total {
				select {
				case moedaCh <- struct{}{}:
				case <-doneCh:
				}
				return
			}
		}
	}
}