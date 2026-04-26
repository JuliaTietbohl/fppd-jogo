// jogo.go - Funções para manipular os elementos do jogo, como carregar o mapa e mover o personagem
package main

import (
	"bufio"
	"os"
	"math"
	"sync"
)

// Elemento representa qualquer objeto do mapa (parede, personagem, vegetação, etc)
type Elemento struct {
	simbolo   rune
	cor       Cor
	corFundo  Cor
	tangivel  bool // Indica se o elemento bloqueia passagem
}

// Jogo contém o estado atual do jogo
type Jogo struct {
	Mapa            [][]Elemento // grade 2D representando o mapa
	PosX, PosY      int          // posição atual do personagem
	UltimoVisitado  Elemento     // elemento que estava na posição do personagem antes de mover
	StatusMsg       string       // mensagem para a barra de status
	HP       int                 // pontos de vida do jogador
	MaxHP    int                 // vida máxima
	Inimigos []EstadoInimigo     // posições dos inimigos (cada um tem uma goroutine)
	Mu       sync.RWMutex        // protege só leituras concorrentes de renderização
	MoedasColetadas int			 // Objetivo: coletar moedas
	TotalMoedas     int			 // Contador moedas
	VidaExtraDisponivel bool     // true enquanto o item ainda está no mapa (só pode pegar 1 vez)
}

type EstadoInimigo struct {
	X, Y int
	Vivo bool
}
 
type MoveInimigo struct {
	ID int 
	DX int
	DY int
}
 
type EstadoRender struct {
	Mapa      [][]Elemento
	PosX      int
	PosY      int
	HP        int
	MaxHP     int
	StatusMsg string
	Inimigos  []EstadoInimigo
	MoedasColetadas int
	TotalMoedas     int
}

// Elementos visuais do jogo
var (
	Personagem = Elemento{'☺', CorCinzaEscuro, CorPadrao, true}
	Inimigo    = Elemento{'☠', CorVermelho, CorPadrao, true}
	Parede     = Elemento{'▤', CorParede, CorFundoParede, true}
	Vegetacao  = Elemento{'♣', CorVerde, CorPadrao, false}
	Vazio      = Elemento{' ', CorPadrao, CorPadrao, false}
	Moeda      = Elemento{'●', CorAmarelo, CorPadrao, false}
	VidaExtra = Elemento{'♥', CorVermelho, CorPadrao, false}
)

// Cria e retorna uma nova instância do jogo
func jogoNovo() Jogo {
	// O ultimo elemento visitado é inicializado como vazio
	// pois o jogo começa com o personagem em uma posição vazia
	return Jogo{UltimoVisitado: Vazio, HP: 5, MaxHP: 5, VidaExtraDisponivel: true}
}

// Lê um arquivo texto linha por linha e constrói o mapa do jogo
func jogoCarregarMapa(nome string, jogo *Jogo) error {
	arq, err := os.Open(nome)
	if err != nil {
		return err
	}
	defer arq.Close()
 
	scanner := bufio.NewScanner(arq)
	y := 0
	for scanner.Scan() {
		linha := scanner.Text()
		var linhaElems []Elemento
		for x, ch := range linha {
			e := Vazio
			switch ch {
			case Parede.simbolo:
				e = Parede
			case Inimigo.simbolo:
				jogo.Inimigos = append(jogo.Inimigos, EstadoInimigo{X: x, Y: y, Vivo: true})
				e = Vazio
			case Vegetacao.simbolo:
				e = Vegetacao
			case VidaExtra.simbolo:
    			e = VidaExtra
			case Moeda.simbolo:
    			jogo.TotalMoedas++
   	 			e = Moeda
			case Personagem.simbolo:
				jogo.PosX, jogo.PosY = x, y // registra a posição inicial do personagem
			}
			linhaElems = append(linhaElems, e)
		}
		jogo.Mapa = append(jogo.Mapa, linhaElems)
		y++
	}
	if err := scanner.Err(); err != nil {
		return err
	}
	return nil
}

// Verifica se o personagem pode se mover para a posição (x, y)
func jogoPodeMoverPara(jogo *Jogo, x, y int) bool {
	// Verifica se a coordenada Y está dentro dos limites verticais do mapa
	if y < 0 || y >= len(jogo.Mapa) {
		return false
	}

	// Verifica se a coordenada X está dentro dos limites horizontais do mapa
	if x < 0 || x >= len(jogo.Mapa[y]) {
		return false
	}

	// Verifica se o elemento de destino é tangível (bloqueia passagem)
	if jogo.Mapa[y][x].tangivel {
		return false
	}

	// Pode mover para a posição
	return true
}

// Move um elemento para a nova posição
func jogoMoverElemento(jogo *Jogo, x, y, dx, dy int) {
	nx, ny := x+dx, y+dy

	// Obtem elemento atual na posição
	elemento := jogo.Mapa[y][x] // guarda o conteúdo atual da posição

	jogo.Mapa[y][x] = jogo.UltimoVisitado     // restaura o conteúdo anterior
	jogo.UltimoVisitado = jogo.Mapa[ny][nx]   // guarda o conteúdo atual da nova posição
	jogo.Mapa[ny][nx] = elemento              // move o elemento
}

func jogoAplicarMoveInimigo(jogo *Jogo, mv MoveInimigo) bool {
	if mv.ID >= len(jogo.Inimigos) {
		return false
	}
	ini := &jogo.Inimigos[mv.ID]
	if !ini.Vivo {
		return false
	}
 
	nx, ny := ini.X+mv.DX, ini.Y+mv.DY
 
	// Colisão com o jogador causa dano mas não move
	if nx == jogo.PosX && ny == jogo.PosY {
		jogo.HP--
		if jogo.HP < 0 {
			jogo.HP = 0
		}
		return true
	}
 
	// Verifica colisão com outros inimigos vivos
	for i, outro := range jogo.Inimigos {
		if i != mv.ID && outro.Vivo && outro.X == nx && outro.Y == ny {
			return false
		}
	}
 
	// Move se a célula destino for livre
	if jogoPodeMoverPara(jogo, nx, ny) {
		jogo.Mu.Lock()
		ini.X, ini.Y = nx, ny
		jogo.Mu.Unlock() 
	}
	return false
}
 
func jogoSnapshot(jogo *Jogo) EstadoRender {
	mapaCopy := make([][]Elemento, len(jogo.Mapa))
	for i, linha := range jogo.Mapa {
		c := make([]Elemento, len(linha))
		copy(c, linha)
		mapaCopy[i] = c
	}
	inimCopy := make([]EstadoInimigo, len(jogo.Inimigos))
	copy(inimCopy, jogo.Inimigos)
 
	return EstadoRender{
		Mapa:      mapaCopy,
		PosX:      jogo.PosX,
		PosY:      jogo.PosY,
		HP:        jogo.HP,
		MaxHP:     jogo.MaxHP,
		StatusMsg: jogo.StatusMsg,
		Inimigos:  inimCopy,
		MoedasColetadas: jogo.MoedasColetadas,
		TotalMoedas:     jogo.TotalMoedas,
	}
}
 
// ============================================================
// >       BFS para pathfinding dos inimigos  (ALEST II)      <
// ============================================================
 
type ponto struct{ x, y int }
 
func jogoBFSNextStep(jogo *Jogo, sx, sy, tx, ty int) (int, int) {
	if sx == tx && sy == ty {
		return 0, 0
	}
 
	dirs := []ponto{{0, -1}, {0, 1}, {-1, 0}, {1, 0}}
	origem := ponto{sx, sy}
	visitado := map[ponto]ponto{origem: origem}
	fila := []ponto{origem}
 
	for len(fila) > 0 {
		atual := fila[0]
		fila = fila[1:]
 
		for _, d := range dirs {
			nx, ny := atual.x+d.x, atual.y+d.y
			viz := ponto{nx, ny}
			if _, visto := visitado[viz]; visto {
				continue
			}
			// Chegou ao destino reconstrói o primeiro passo do caminho
			if nx == tx && ny == ty {
				visitado[viz] = atual
				cur := viz
				for visitado[cur] != origem {
					cur = visitado[cur]
				}
				return cur.x - sx, cur.y - sy
			}
			if jogoPodeMoverPara(jogo, nx, ny) {
				visitado[viz] = atual
				fila = append(fila, viz)
			}
		}
	}
 
// tenta se aproximar do jogador andando na direção que diminui mais a distância em linha reta
// se não for acessível, retorna (0,0) e o inimigo fica parado.
	best, bestDist := ponto{0, 0}, math.MaxFloat64
	for _, d := range dirs {
		nx, ny := sx+d.x, sy+d.y
		if !jogoPodeMoverPara(jogo, nx, ny) {
			continue
		}
		dist := math.Sqrt(float64((nx-tx)*(nx-tx) + (ny-ty)*(ny-ty)))
		if dist < bestDist {
			bestDist = dist
			best = ponto{d.x, d.y}
		}
	}
	return best.x, best.y
}

func jogoMoverPersonagem(jogo *Jogo, dx, dy int) {
	nx, ny := jogo.PosX+dx, jogo.PosY+dy
	if jogoPodeMoverPara(jogo, nx, ny) {
		jogoMoverElemento(jogo, jogo.PosX, jogo.PosY, dx, dy)
		jogo.PosX, jogo.PosY = nx, ny
	}
}

func jogoColetarMoeda(jogo *Jogo) bool {
    // Verifica se o elemento sob o personagem é uma moeda
    if jogo.UltimoVisitado.simbolo == Moeda.simbolo {
		jogo.Mu.Lock()
        jogo.UltimoVisitado = Vazio
        jogo.MoedasColetadas++
		jogo.Mu.Unlock()  
        return true
    }
    return false
}

func jogoVerificarColisaoInimigos(jogo *Jogo) bool {
    colidiu := false
    for _, ini := range jogo.Inimigos {
        if !ini.Vivo {
            continue
        }
        // Mesmo lugar que o jogador
        if ini.X == jogo.PosX && ini.Y == jogo.PosY {
            jogo.HP--
            if jogo.HP < 0 {
                jogo.HP = 0
            }
            colidiu = true
        }
    }
    return colidiu
}
