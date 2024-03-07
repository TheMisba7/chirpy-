package storage

import (
	"bufio"
	"encoding/json"
	"io"
	"os"
	"sync"
	"syscall"
)

var chirpId int = 0

type Chirp struct {
	id   int
	Body string `json:"body"`
}
type dbStructure struct {
	Chirps map[int]Chirp `json:"chirps"`
}

type Storage struct {
	url string
	mu  sync.Mutex
}

func NewChirp() *Chirp {
	chirpId++
	return &Chirp{id: chirpId}
}

func NewStorage(url string) *Storage {
	s := &Storage{url: url}
	return s
}
func (s *Storage) GetById(id int) (*Chirp, bool) {
	data := s.readAll()

	chirp, ok := data.Chirps[id]
	return &chirp, ok

}
func (s *Storage) readAll() dbStructure {
	target := dbStructure{Chirps: make(map[int]Chirp)}
	data := s.Read()
	if len(data) > 0 {
		err := json.Unmarshal(s.Read(), &target)
		if err != nil {
			panic(err)
		}
	}
	return target
}
func (s *Storage) Add(chirp Chirp) map[int]Chirp {
	target := s.readAll()
	target.Chirps[chirp.id] = chirp
	bytes, err := json.Marshal(target)
	if err != nil {
		panic(err)
	}
	s.Write(bytes)
	return target.Chirps
}

func (s *Storage) Write(data []byte) {
	s.mu.Lock()
	defer s.mu.Unlock()
	file, err := os.OpenFile(s.url, os.O_CREATE, syscall.FILE_MAP_WRITE)
	if err != nil {
		panic(err)
	}
	file.Write(data)

}

func (s *Storage) Read() []byte {
	s.mu.Lock()
	defer s.mu.Unlock()
	file, err := os.Open(s.url)
	if err != nil {
		panic(err)
	}
	data, _ := io.ReadAll(bufio.NewReader(file))
	return data
}
