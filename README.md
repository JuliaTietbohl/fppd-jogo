# Jogo de Terminal em Go

Este projeto é um pequeno jogo desenvolvido em Go que roda no terminal usando a biblioteca [termbox-go](https://github.com/nsf/termbox-go). O jogador controla um personagem que pode se mover por um mapa carregado de um arquivo de texto.

## Como funciona

- O mapa é carregado de um arquivo `.txt` contendo caracteres que representam diferentes elementos do jogo.
- O personagem se move com as teclas **W**, **A**, **S**, **D**.
- Pressione **E** para interagir com o ambiente.
- Pressione **ESC** para sair do jogo.

### Alterações feitas
Arquitetura concorrente adicionada ao projeto base do professor:

main.go — o loop principal foi substituído por goroutines comunicando via channels. O main agora cria os channels, lança as goroutines e aguarda o encerramento gracioso com sync.WaitGroup.
jogo.go — adicionadas structs (EstadoInimigo, MoveInimigo, EstadoRender) e funções para suportar concorrência: jogoSnapshot (cópia imutável do estado), jogoAplicarMoveInimigo (processa movimentos dos inimigos), jogoBFSNextStep (pathfinding por busca em largura), jogoColetarMoeda e jogoVerificarColisaoInimigos. Adicionados campos de HP, moedas e vida extra.
interface.go — adicionada interfaceDesenharEstado que renderiza a partir de um snapshot imutável em vez do estado direto do jogo, evitando race condition com o gameLoop.

-> goroutines.go — arquivo novo com 7 goroutines:

    inputLoop — lê teclado e envia eventos via channel
    gameLoop — lógica central com select multiplexando input, inimigos e tempo
    renderLoop — desenha snapshots recebidos via channel
    inimigoLoop × 3 — cada inimigo com uma com BFS independente perseguindo o jogador
    moedaLoop — monitora a condição de vitória (5 moedas coletadas)

personagem.go — personagemMover protegido com mutex para evitar data race com os inimigoLoops. personagemInteragir implementado para coletar vida extra (♥).

### Controles

| Tecla | Ação              |
|-------|-------------------|
| W     | Mover para cima   |
| A     | Mover para esquerda |
| S     | Mover para baixo  |
| D     | Mover para direita |
| E     | Interagir         |
| ESC   | Sair do jogo      |

## Como compilar

1. Instale o Go e clone este repositório.
2. Inicialize um novo módulo "jogo":

```bash
go mod init jogo
go get -u github.com/nsf/termbox-go
```

3. Compile o programa:

Linux:

```bash
go build -o jogo
```

Windows:

```bash
go build -o jogo.exe
```

Também é possivel compilar o projeto usando o comando `make` no Linux ou o script `build.bat` no Windows.

## Como executar

1. Certifique-se de ter o arquivo `mapa.txt` com um mapa válido.
2. Execute o programa no termimal:

```bash
./jogo
```

## Estrutura do projeto

- main.go — Ponto de entrada e loop principal
- interface.go — Entrada, saída e renderização com termbox
- jogo.go — Estruturas e lógica do estado do jogo
- personagem.go — Ações do jogador


