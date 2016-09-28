package main

import (
	"bufio"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"math/rand"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	filteredWordsRaw = "./data/filtered_words.raw"
	sentencesRaw     = "./data/sentences.raw"
	sentencesJSON    = "./data/sentences.json"
)

type sentence struct {
	Sentence string `json:"sentence"`
	Correct  int    `json:"correct"`
	Wrong    int    `json:"wrong"`
	Weight   int    `json:"weight"` // a number between 0 to 100
}

func newSentence(s string) sentence {
	return sentence{
		Sentence: s,
		Correct:  0,
		Wrong:    0,
		Weight:   50, // Weight is not used now
	}
}

type sentenceManager struct {
	filteredWords map[string]bool
	sentences     []sentence
	wordsIndex    map[string][]int // word -> [id0, id1...], id is the sentence's index in sentences slice

	sync.Mutex
}

var defaultSM *sentenceManager

func init() {
	var err error
	defaultSM, err = newSentenceManager()
	if err != nil {
		panic(err)
	}
}

// GetSentence x
func GetSentence(id int) []string {
	return defaultSM.getSentence(id)
}

// GenSentence x
func GenSentence(keyWord string) (int, []string) {
	return defaultSM.genSentence(keyWord)
}

// CheckSentence x
func CheckSentence(id int, words []string) (bool, []string) {
	return defaultSM.checkSentence(id, words)
}

func newSentenceManager() (*sentenceManager, error) {
	sm := &sentenceManager{
		filteredWords: make(map[string]bool),
		wordsIndex:    make(map[string][]int),
	}

	err := sm.loadFilteredWords()
	if err != nil {
		return nil, fmt.Errorf("loadFilteredWords err: %v", err)
	}

	err = sm.loadSentencesJSON()
	if err != nil {
		return nil, fmt.Errorf("loadSentencesJSON err: %v", err)
	}

	err = sm.loadSentencesRaw()
	if err != nil {
		return nil, fmt.Errorf("loadSentencesRaw err: %v", err)
	}

	go sm.async()

	for id := range sm.sentences {
		sm.createWordsIndex(id)
	}

	return sm, nil
}

func (sm *sentenceManager) getSentence(id int) []string {
	return getWords(sm.sentences[id].Sentence)
}

func (sm *sentenceManager) checkSentence(id int, words []string) (bool, []string) {
	correctWords := getWords(sm.sentences[id].Sentence)
	result := true
	for i := 0; i < len(words); i++ {
		if words[i] != correctWords[i] {
			words[i] = words[i] + "=>" + correctWords[i]
			result = false
		}
	}

	if result == true {
		sm.sentences[id].Correct++
	} else {
		sm.sentences[id].Wrong++
	}

	return result, words
}

// genSentence
func (sm *sentenceManager) genSentence(keyWord string) (int, []string) {
	id := -1
	if keyWord != "" {
		if _, ok := sm.wordsIndex[keyWord]; ok == true {
			ids := sm.wordsIndex[keyWord]
			id = ids[rand.Int()%len(ids)]
		}
	}
	if id == -1 {
		id = rand.Int() % len(sm.sentences)
	}

	sentence := sm.sentences[id]
	words := getWords(sentence.Sentence)
	for i, w := range words {
		if sm.isFilteredWord(w) == true {
			continue
		}
		tmp := rand.Int() % len(words)
		if tmp < 4 { // empty this word
			words[i] = ""
		}
	}

	return id, words
}

// isFilteredWord checks if this word should be filtered
func (sm *sentenceManager) isFilteredWord(word string) bool {
	if len(word) < 3 { // filter words which are too short
		return true
	}
	if strings.Contains(word, "'") == true || // filter words like Tom's
		(word[len(word)-1] == '.') == true || // filter words like U.S.
		strings.Contains(word, "-") == true { // filter words like
		return true
	}

	return sm.filteredWords[word] // filter words in filteredWordsSet
}

// createWordsIndex creates index for unfiltered words
func (sm *sentenceManager) createWordsIndex(id int) {
	sentence := sm.sentences[id]
	words := getWords(sentence.Sentence)

	for _, w := range words {
		if sm.isFilteredWord(w) == false {
			sm.wordsIndex[w] = append(sm.wordsIndex[w], id)
		}
	}
}

// async syncs sentences to disk files
func (sm *sentenceManager) async() {
	ticker := time.Tick(time.Second * 30)
	for {
		err := sm.syncFilteredWords()
		if err != nil {
			panic(err)
		}

		err = sm.syncSentences()
		if err != nil {
			panic(err)
		}
		<-ticker
	}
}

// syncSentences syncs sentences to its disk file
func (sm *sentenceManager) syncSentences() error {
	sm.Lock()
	defer sm.Unlock()

	buf, err := json.Marshal(sm.sentences)
	if err != nil {
		return fmt.Errorf("marshal json err: %v", err)
	}

	tmpName := "/tmp/tmpsentences.json"
	tmpFile, err := os.OpenFile(tmpName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("open file err: %v", err)
	}

	_, err = tmpFile.Write(buf)
	if err != nil {
		return fmt.Errorf("write file err: %v", err)
	}
	tmpFile.Close()

	err = os.Rename(tmpName, sentencesJSON)
	if err != nil {
		return fmt.Errorf("rename err: %v", err)
	}

	return nil
}

// syncFilteredWords syncs filtered words to its disk file
func (sm *sentenceManager) syncFilteredWords() error {
	sm.Lock()
	defer sm.Unlock()

	tmpName := "/tmp/tmpwords.raw"
	tmpFile, err := os.OpenFile(tmpName, os.O_RDWR|os.O_CREATE, 0666)
	if err != nil {
		return fmt.Errorf("open file err: %v", err)
	}

	count := 0
	for word := range sm.filteredWords {
		word += "\t"
		count++
		if count == 10 {
			word += "\n"
			count = 0
		}

		_, err := tmpFile.Write([]byte(word))
		if err != nil {
			return fmt.Errorf("write file err: %v", err)
		}
	}
	tmpFile.Close()

	err = os.Rename(tmpName, filteredWordsRaw)
	if err != nil {
		return fmt.Errorf("rename err: %v", err)
	}

	return nil
}

// loadSentencesRaw loads sentences from raw file
func (sm *sentenceManager) loadSentencesRaw() error {
	f, err := os.Open(sentencesRaw)
	if err != nil {
		return fmt.Errorf("open file err: %v", err)
	}

	sentencesSet := make(map[string]bool)
	for _, s := range sm.sentences {
		sentencesSet[s.Sentence] = true
	}

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() == true {
		sentence := strings.TrimSpace(scanner.Text())
		if len(sentence) == 0 {
			continue
		}
		if sentencesSet[sentence] == true {
			continue
		}

		s := newSentence(sentence)
		sm.sentences = append(sm.sentences, s)
	}

	return nil
}

// loadSentencesJSON loads sentences from the file identified sentencesJSON
func (sm *sentenceManager) loadSentencesJSON() error {
	buf, err := ioutil.ReadFile(sentencesJSON)
	if err != nil {
		return fmt.Errorf("read file err: %v", err)
	}

	err = json.Unmarshal(buf, &sm.sentences)
	if err != nil {
		return fmt.Errorf("unmarshal json err: %v", err)
	}

	return nil
}

// loadFilteredWords loads filtered words to sm
func (sm *sentenceManager) loadFilteredWords() error {
	f, err := os.Open(filteredWordsRaw)
	if err != nil {
		return fmt.Errorf("open file err: %v", err)
	}

	scanner := bufio.NewScanner(f)
	scanner.Split(bufio.ScanWords)
	for scanner.Scan() {
		word := strings.TrimSpace(scanner.Text())
		sm.filteredWords[word] = true
	}

	return nil
}
