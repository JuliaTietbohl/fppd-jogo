// main.go - Loop principal e ponto de entrada
package main

import (
	"os"
	"sync"
)

func main() {
	// Inicializa a interface (termbox)
	interfaceIniciar()
	defer interfaceFinalizar()

	// Usa "mapa.txt" como arquivo padrão ou lê o primeiro argumento
	mapaFile := "mapa.txt"
	if len(os.Args) > 1 {
		mapaFile = os.Args[1]
	}

	// Inicializa o jogo
	jogo := jogoNovo()
	if err := jogoCarregarMapa(mapaFile, &jogo); err != nil {
		panic(err)
	}

	// Spawna a primeira moeda no mapa
	jogoSpawnarMoeda(&jogo)

	// channels de comunicação entre as goroutines
	inputCh      := make(chan EventoTeclado, 1)   // input -> gameLoop
	renderCh     := make(chan EstadoRender, 1)    // gameLoop -> renderLoop
	inimigoCh    := make(chan MoveInimigo, 10)    // inimigoLoop(s) -> gameLoop
	moedaCh      := make(chan struct{}, 1)		  // moedaLoop -> gameLoop (sinal de vitória)
	fimCh        := make(chan struct{})           // fechado pelo gameLoop morrer/vencer
	doneCh       := make(chan struct{})           // sinal de shutdown (fechado para broadcast)
	spawnMoedaCh := make(chan int, 1)             // moedaLoop -> gameLoop (spawna próxima moeda)
	piscarCh     := make(chan bool, 1)            // piscar coração
	var wg sync.WaitGroup

	//goroutine de input
	wg.Add(1)
	go inputLoop(inputCh, doneCh, &wg)

	// goroutine de render
	wg.Add(1)
	go renderLoop(renderCh, doneCh, &wg)

	//goroutine autônoma por inimigo encontrado no mapa
	for i := range jogo.Inimigos {
		wg.Add(1)
		go inimigoLoop(i, &jogo, inimigoCh, fimCh, doneCh, &wg)
	}

	// moedaLoop — monitora coleta e spawna próximas moedas
	wg.Add(1)
	go moedaLoop(&jogo, moedaCh, spawnMoedaCh, doneCh, &wg)

	wg.Add(1)
	go piscarLoop(&jogo, piscarCh, doneCh, fimCh, &wg)

	gameLoop(&jogo, inputCh, renderCh, inimigoCh, moedaCh, spawnMoedaCh, piscarCh, fimCh, doneCh)

	// Shutdown sinaliza todas as goroutines e aguarda conclusão
	close(doneCh)
	wg.Wait()
}