package main

import (
	"bufio"
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"edgetts/edgetts"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()

	voices, err := edgetts.ListVoices(ctx, "")
	if err != nil {
		log.Fatalf("list voices: %v", err)
	}

	for i, voice := range voices {
		fmt.Printf("%4d  %-36s %-6s %s\n", i+1, voice.ShortName, voice.Gender, voice.FriendlyName)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Print("\nВыберите голос: номер или ShortName: ")
	choice, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("read voice: %v", err)
	}
	choice = strings.TrimSpace(choice)

	voiceName, err := chooseVoice(voices, choice)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Print("Введите текст: ")
	text, err := reader.ReadString('\n')
	if err != nil {
		log.Fatalf("read text: %v", err)
	}
	text = strings.TrimSpace(text)
	if text == "" {
		log.Fatal("текст пустой")
	}

	file, err := os.Create("input.mp3")
	if err != nil {
		log.Fatalf("create input.mp3: %v", err)
	}
	defer file.Close()

	err = edgetts.WriteSpeech(ctx, file, text, edgetts.Options{Voice: voiceName})
	if err != nil {
		log.Fatalf("synthesize: %v", err)
	}

	fmt.Println("Готово: input.mp3")
}

func chooseVoice(voices []edgetts.Voice, choice string) (string, error) {
	if choice == "" {
		return "", fmt.Errorf("голос не выбран")
	}

	if n, err := strconv.Atoi(choice); err == nil {
		if n < 1 || n > len(voices) {
			return "", fmt.Errorf("номер голоса вне диапазона")
		}
		return voices[n-1].ShortName, nil
	}

	for _, voice := range voices {
		if strings.EqualFold(voice.ShortName, choice) {
			return voice.ShortName, nil
		}
	}
	return "", fmt.Errorf("голос %q не найден", choice)
}
