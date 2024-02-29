package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"math"
	"math/rand"
	"os"
	"slices"
	"strings"
	"unicode"

	"github.com/fatih/color"
	"golang.org/x/text/runes"
	"golang.org/x/text/transform"
	"golang.org/x/text/unicode/norm"
)

type History struct {
	Word       string
	Correction []int
}

type TryResult struct {
	Correction []int // -1 = Nao contem, 0 = Contem, 1 = Correto
}

type Game struct {
	Word          string
	ValidWordList []string
	Hints         []string
	Histories     []History
	Chances       int
}

func (game *Game) readLines(r io.Reader, numberOfCharacters int) {
	normalize := transform.Chain(norm.NFD, runes.Remove(runes.In(unicode.Mn)), norm.NFC)
	words := make([]string, 0)
	sc := bufio.NewScanner(r)
	index := 0
	for sc.Scan() {
		word := sc.Text()
		if len(word) == numberOfCharacters {
			output, _, e := transform.String(normalize, sc.Text())
			if e != nil {
				panic(e)
			}
			words = append(words, strings.ToLower(output))
		}
		index++
	}
	game.ValidWordList = words
}

func (game *Game) getWord() {
	wordIndex := rand.Intn(len(game.ValidWordList))
	game.Word = game.ValidWordList[wordIndex]
}

func (game *Game) Init(numberOfCharacters int, numberOfChances int) {
	data, openErr := os.Open("./words.txt")
	if openErr != nil {
		fmt.Println("Erro ao ler arquivo de palavras")
	}
	game.readLines(data, numberOfCharacters)
	game.getWord()
	game.Chances = numberOfChances
}

func (game *Game) tryWord(word string) (bool, bool, bool) {
	index := slices.IndexFunc(game.ValidWordList, func(text string) bool { return strings.Compare(text, word) == 0 })
	if index == -1 {
		return false, true, false
	}

	correct := true
	result := TryResult{}
	for index, letter := range word {
		if game.Word[index] == byte(letter) {
			result.Correction = append(result.Correction, 1)
		} else if strings.Contains(game.Word, string(letter)) {
			result.Correction = append(result.Correction, 0)
			correct = false
		} else {
			result.Correction = append(result.Correction, -1)
			correct = false
		}
	}

	game.Histories = append(game.Histories, History{Word: word, Correction: result.Correction})

	actualChance := len(game.Histories)
	chancesOver := false
	if game.Chances != -1 && game.Chances-actualChance == 0 {
		chancesOver = true
	}

	return correct, false, chancesOver
}

func (game *Game) printScreen() {
	fmt.Print("\033[H\033[2J")
	redText := color.New(color.FgRed)
	greenText := color.New(color.FgGreen)
	yellowText := color.New(color.FgYellow)

	for historyIndex, history := range game.Histories {
		if game.Chances != -1 {
			matching := math.Round(float64(game.Chances) / 3)
			rating := math.Round(float64(game.Chances) / float64(historyIndex))
			if rating < matching {
				redText.Printf("%d", historyIndex+1)
				fmt.Printf(" - ")
			} else if rating < matching*2 {
				yellowText.Printf("%d", historyIndex+1)
				fmt.Printf(" - ")
			} else {
				greenText.Printf("%d", historyIndex+1)
				fmt.Printf(" - ")
			}
		} else {
			fmt.Printf("%d - ", historyIndex+1)
		}
		for index, correction := range history.Correction {
			if correction == 0 {
				yellowText.Printf("%c", history.Word[index])
			} else if correction == 1 {
				greenText.Printf("%c", history.Word[index])
			} else if correction == -1 {
				redText.Printf("%c", history.Word[index])
			}

			if len(history.Correction)-1 != index {
				fmt.Printf(" | ")
			} else {
				fmt.Print("\n")
			}
		}
	}
	fmt.Print("\n")
}

func (game *Game) gameLoop() {
	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\033[H\033[2J")

	for {
		if len(game.Hints) > 0 {
			fmt.Printf("A palavra a ser testada tem %d letras, digite 1 para dicas (%d restantes) ou digite a palavra a ser testada: ", len(game.Word), len(game.Hints))
		} else if game.Chances != -1 {
			fmt.Printf("A palavra a ser testada tem %d letras,você tem %d chances: ", len(game.Word), game.Chances-len(game.Histories))
		} else {
			fmt.Printf("A palavra a ser testada tem %d letras: ", len(game.Word))
		}
		word, err := reader.ReadString('\n')
		if err != nil {
			fmt.Println("Erro ao fazer leitura, tente novamente")
		}
		word = strings.Trim(word, "\n \r")

		if strings.Compare(word, "1") == 0 {
			game.printScreen()
			if len(game.Hints) > 0 {
				fmt.Printf("A dica é: %s\n", game.Hints[0])
				newArr := make([]string, 0)
				game.Hints = append(newArr, game.Hints[1:]...)
				continue
			}

			fmt.Println("Sem dicas disponiveis")
			continue
		}

		if strings.Compare(word, "0") == 0 {
			fmt.Printf("Desistiu!! QUE FEIOOOOOOOO, a palavra era: %s", game.Word)
			break
		}

		if len(word) != len(game.Word) {
			game.printScreen()
			fmt.Println("Tamanho incorreto, tente novamente!")
			continue
		}

		correct, invalidWord, choicesOver := game.tryWord(word)
		if correct {
			game.printScreen()
			fmt.Printf("PARABENS!!! Você acertou a palavra, ainda restaram %d tentativas!", game.Chances-len(game.Histories))
			break
		}

		if invalidWord {
			game.printScreen()
			fmt.Printf("A palavra '%s' não pertence ao dicionário brasileiro\n", word)
			continue
		}

		if choicesOver {
			game.printScreen()
			fmt.Printf("Infelizmente acabaram suas chances, a palavra era: %s.", game.Word)
			break
		}

		game.printScreen()
	}
}

func main() {
	numberOfCharacters := flag.Int("letters", 5, "Numero de caracteres das palavras do jogo!")
	numberOfChances := flag.Int("chances", -1, "Numero de changes de acertar a palavra")
	flag.Parse()

	game := Game{}
	game.Init(*numberOfCharacters, *numberOfChances)
	game.gameLoop()
}
