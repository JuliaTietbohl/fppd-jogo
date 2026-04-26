// interface.go - Interface gráfica do jogo usando termbox
// O código abaixo implementa a interface gráfica do jogo usando a biblioteca termbox-go.
// A biblioteca termbox-go é uma biblioteca de interface de terminal que permite desenhar
// elementos na tela, capturar eventos do teclado e gerenciar a aparência do terminal.

package main

import (
	"github.com/nsf/termbox-go"
	"fmt"
)

// Define um tipo Cor para encapsuladar as cores do termbox
type Cor = termbox.Attribute

// Definições de cores utilizadas no jogo
const (
	CorPadrao     Cor = termbox.ColorDefault
	CorCinzaEscuro    = termbox.ColorDarkGray
	CorVermelho       = termbox.ColorRed
	CorVerde          = termbox.ColorGreen
	CorParede         = termbox.ColorBlack | termbox.AttrBold | termbox.AttrDim
	CorFundoParede    = termbox.ColorDarkGray
	CorTexto          = termbox.ColorDarkGray
	CorHPCheio        = termbox.ColorGreen
	CorHPMedio        = termbox.ColorYellow
	CorHPBaixo        = termbox.ColorRed
	CorAmarelo		  = termbox.ColorYellow
)

// EventoTeclado representa uma ação detectada do teclado (como mover, sair ou interagir)
type EventoTeclado struct {
	Tipo  string // "sair", "interagir", "mover"
	Tecla rune   // Tecla pressionada, usada no caso de movimento
}

// Inicializa a interface gráfica usando termbox
func interfaceIniciar() {
	if err := termbox.Init(); err != nil {
		panic(err)
	}
}

// Encerra o uso da interface termbox
func interfaceFinalizar() {
	termbox.Close()
}

// Lê um evento do teclado e o traduz para um EventoTeclado
func interfaceLerEventoTeclado() EventoTeclado {
	ev := termbox.PollEvent()
	if ev.Type != termbox.EventKey {
		return EventoTeclado{}
	}
	if ev.Key == termbox.KeyEsc {
		return EventoTeclado{Tipo: "sair"}
	}
	if ev.Ch == 'e' {
		return EventoTeclado{Tipo: "interagir"}
	}
	return EventoTeclado{Tipo: "mover", Tecla: ev.Ch}
}

// Renderiza todo o estado atual do jogo na tela
func interfaceDesenharJogo(jogo *Jogo) {
	interfaceLimparTela()

	// Desenha todos os elementos do mapa
	for y, linha := range jogo.Mapa {
		for x, elem := range linha {
			interfaceDesenharElemento(x, y, elem)
		}
	}

	// Desenha o personagem sobre o mapa
	interfaceDesenharElemento(jogo.PosX, jogo.PosY, Personagem)

	// Desenha a barra de status
	interfaceDesenharBarraDeStatus(jogo)

	// Força a atualização do terminal
	interfaceAtualizarTela()
}

// Limpa a tela do terminal
func interfaceLimparTela() {
	termbox.Clear(CorPadrao, CorPadrao)
}

// Força a atualização da tela do terminal com os dados desenhados
func interfaceAtualizarTela() {
	termbox.Flush()
}

// Desenha um elemento na posição (x, y)
func interfaceDesenharElemento(x, y int, elem Elemento) {
	termbox.SetCell(x, y, elem.simbolo, elem.cor, elem.corFundo)
}

// Exibe uma barra de status com informações úteis ao jogador
func interfaceDesenharBarraDeStatus(jogo *Jogo) {
	for i, c := range jogo.StatusMsg {
		termbox.SetCell(i, len(jogo.Mapa)+1, c, CorTexto, CorPadrao)
	}

	msg := "Use WASD para mover e E para interagir. ESC para sair."
	for i, c := range msg {
		termbox.SetCell(i, len(jogo.Mapa)+3, c, CorTexto, CorPadrao)
	}
}

func interfaceDesenharEstado(estado EstadoRender) {
	interfaceLimparTela()
 
	// Desenha o mapa base
	for y, linha := range estado.Mapa {
		for x, elem := range linha {
			interfaceDesenharElemento(x, y, elem)
		}
	}
 
	// Desenha inimigos vivos
	for _, ini := range estado.Inimigos {
		if ini.Vivo {
			interfaceDesenharElemento(ini.X, ini.Y, Inimigo)
		}
	}
 
	// Desenha o personagem por cima
	interfaceDesenharElemento(estado.PosX, estado.PosY, Personagem)
 
	// Linha de status
	statusY := len(estado.Mapa) + 1
	for i, c := range estado.StatusMsg {
		termbox.SetCell(i, statusY, c, CorTexto, CorPadrao)
	}
 
	// Barra de HP colorida
	hpY := statusY + 1
	hpStr := fmt.Sprintf("HP: [%s] %d/%d", barraHPVisual(estado.HP, estado.MaxHP), estado.HP, estado.MaxHP)
	corHP := CorHPCheio
	if estado.HP <= estado.MaxHP/3 {
		corHP = CorHPBaixo
	} else if estado.HP <= (estado.MaxHP*2)/3 {
		corHP = CorHPMedio
	}
	for i, c := range hpStr {
		termbox.SetCell(i, hpY, c, corHP, CorPadrao)
	}

	// Contador de moedas
	moedaStr := fmt.Sprintf("Moedas: %d/%d  ●", estado.MoedasColetadas, estado.TotalMoedas)
	for i, ch := range moedaStr {
		termbox.SetCell(i, hpY+1, ch, termbox.ColorYellow, CorPadrao)
	}
 
	// Instruções fixas
	msg := "Use WASD para mover e E para interagir. ESC para sair."
	for i, c := range msg {
		termbox.SetCell(i, hpY+2, c, CorTexto, CorPadrao)
	}
 
	interfaceAtualizarTela()
}
 
// Barra de HP
func barraHPVisual(hp, maxHP int) string {
	const barLen = 10
	filled := 0
	if maxHP > 0 {
		filled = (hp * barLen) / maxHP
	}
	bar := ""
	for i := 0; i < barLen; i++ {
		if i < filled {
			bar += "█"
		} else {
			bar += "░"
		}
	}
	return bar
}

