package scanner

import (
	"bufio"
	"os"
	"strings"

	"github.com/JoaoDanielRufino/gcloc/pkg/analyzer"
	"github.com/JoaoDanielRufino/gcloc/pkg/gcloc/language"
)

type Scanner struct {
	supportedLanguages language.Languages
}

type scanResult struct {
	Metadata   analyzer.FileMetadata
	CodeLines  int
	BlankLines int
	Comments   int
}

func NewScanner(languages language.Languages) *Scanner {
	return &Scanner{
		supportedLanguages: languages,
	}
}

func (sc *Scanner) Scan(files []analyzer.FileMetadata) ([]scanResult, error) {
	var results []scanResult

	for _, file := range files {
		result, err := sc.scanFile(file)
		if err != nil {
			return results, err
		}

		results = append(results, result)
	}

	return results, nil
}

func (sc *Scanner) ChangeLanguages(languages language.Languages) {
	sc.supportedLanguages = languages
}

func (sc *Scanner) scanFile(file analyzer.FileMetadata) (scanResult, error) {
	result := scanResult{Metadata: file}
	isInBlockComment := false
	var closeBlockCommentToken string

	f, err := os.Open(file.FilePath)
	if err != nil {
		return result, err
	}
	defer f.Close()

	fileScanner := bufio.NewScanner(f)
	for fileScanner.Scan() {
		line := strings.TrimSpace(fileScanner.Text())

		result.CodeLines++

		if isInBlockComment {
			result.Comments++
			if sc.hasSecondMultiLineComment(line, closeBlockCommentToken) {
				isInBlockComment = false
			}
			continue
		}

		if sc.isBlankLine(line) {
			result.BlankLines++
			continue
		}

		if ok, secondCommentToken := sc.hasFirstMultiLineComment(file, line); ok {
			isInBlockComment = true
			closeBlockCommentToken = secondCommentToken
			result.Comments++
			continue
		}

		if sc.hasSingleLineComment(file, line) {
			result.Comments++
		}
	}

	return result, fileScanner.Err()
}

func (sc *Scanner) hasFirstMultiLineComment(file analyzer.FileMetadata, line string) (bool, string) {
	multiLineComments := sc.supportedLanguages[file.Language].MultiLineComments

	for _, multiLineComment := range multiLineComments {
		firstCommentToken := multiLineComment[0]
		if strings.HasPrefix(line, firstCommentToken) {
			return true, multiLineComment[1]
		}
	}

	return false, ""
}

func (sc *Scanner) hasSecondMultiLineComment(line, commentToken string) bool {
	return strings.Contains(line, commentToken)
}

func (sc *Scanner) hasSingleLineComment(file analyzer.FileMetadata, line string) bool {
	lineComments := sc.supportedLanguages[file.Language].LineComments

	for _, lineComment := range lineComments {
		if strings.HasPrefix(line, lineComment) {
			return true
		}
	}

	return false
}

func (sc *Scanner) isBlankLine(line string) bool {
	return len(line) == 0
}
