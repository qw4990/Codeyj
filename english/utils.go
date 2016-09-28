package main

import (
	"bufio"
	"strings"
)

// getWords gets all words in this sentence
func getWords(sentence string) []string {
	words := make([]string, 0, 20)
	reader := strings.NewReader(sentence)
	scanner := bufio.NewScanner(reader)
	scanner.Split(spliter)
	for scanner.Scan() {
		word := scanner.Text()
		words = append(words, word)
	}
	return words
}

func spliter(data []byte, atEOF bool) (advance int, token []byte, err error) {
	n := len(data)
	for advance < n && isSpace(data[advance]) { // remove leading spaces
		advance++
	}

	begin := advance
	data = data[begin:]

	if advance == n && atEOF == false {
		return 0, nil, nil // require more data
	} else if advance == n && atEOF {
		return advance, nil, nil
	}

	if isNumber(data[advance]) {
		advance, token, err = numberSpliter(data, atEOF)
	} else if advance+1 < n && data[advance] == '$' && isNumber(data[advance+1]) {
		advance, token, err = dollarSpliter(data, atEOF)
	} else if isLetter(data[advance]) {
		advance, token, err = wordSpliter(data, atEOF)
	} else {
		advance, token, err = 1, data[:1], nil
	}

	if err != nil {
		return 0, nil, err
	}
	return begin + advance, data[:advance], nil
}

func wordSpliter(data []byte, atEOF bool) (advance int, token []byte, err error) {
	n := len(data)
	for advance < n && (isLetter(data[advance]) || data[advance] == '\'') {
		advance++
	}

	if advance == n && atEOF == false {
		return 0, nil, nil // require more data
	}

	return advance, data[:advance], nil
}

func dollarSpliter(data []byte, atEOF bool) (advance int, token []byte, err error) {
	advance, _, err = numberSpliter(data[1:], atEOF)
	if err != nil {
		return 0, nil, err
	}

	return advance + 1, data[:advance+1], nil
}

func numberSpliter(data []byte, atEOF bool) (advance int, token []byte, err error) {
	n := len(data)
	for advance < n && isNumber(data[advance]) {
		advance++
	}

	if advance == n && atEOF == false {
		return 0, nil, nil // require more data
	}

	return advance, data[:advance], nil
}

func isLetter(b byte) bool {
	return (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z')
}

func isNumber(b byte) bool {
	return b >= '0' && b <= '9'
}

func isSpace(b byte) bool {
	return b == ' ' || b == '\t' || b == '\n' || b == '\r'
}
