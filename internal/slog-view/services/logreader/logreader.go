package logreader

import (
	"bufio"
	"encoding/json"
	"github.com/lmika/awstools/internal/slog-view/models"
	"github.com/pkg/errors"
	"log"
	"os"
)

type Service struct {
}

func NewService() *Service {
	return &Service{}
}

func (s *Service) Open(filename string) (*models.LogFile, error) {
	f, err := os.Open(filename)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot open file: %v", filename)
	}
	defer f.Close()

	var lines []models.LogLine
	scanner := bufio.NewScanner(f)
	for scanner.Scan() {
		line := scanner.Text()

		var data interface{}
		if err := json.Unmarshal([]byte(line), &data); err != nil {
			log.Printf("invalid json line: %v", err)
			continue
		}

		lines = append(lines, models.LogLine{JSON: data})
	}
	if scanner.Err() != nil {
		return nil, errors.Wrapf(err, "unable to scan file: %v", filename)
	}

	return &models.LogFile{Lines: lines}, nil
}
