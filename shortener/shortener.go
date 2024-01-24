package shortenerserver

import (
	"context"
	"go-url-shortener/cfg"
	"go-url-shortener/internal/storage"
	shortenerproto "go-url-shortener/pkg/proto"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"log"
	"net/url"
	"regexp"
	"strconv"
	"strings"
)

type LinkShortener struct {
	shortenerproto.LinkShortenerServer
	storage storage.UrlStorage
	env     *cfg.Env
	dictMap *map[rune]bool
	dict    []rune
}

func New(storage storage.UrlStorage, env *cfg.Env) *LinkShortener {
	dict := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	dict1 := append(dict, []rune(env.FillingChar)...)
	m := make(map[rune]bool)
	for _, d := range dict1 {
		m[d] = true
	}
	return &LinkShortener{
		storage: storage,
		env:     env,
		dictMap: &m,
		dict:    dict,
	}
}

func (s *LinkShortener) CreateShortLink(ctx context.Context, request *shortenerproto.CreateShortLinkRequest) (*shortenerproto.CreateShortLinkResponse, error) {
	found, id := s.storage.CheckExistence(request.Original)
	if found {
		return &shortenerproto.CreateShortLinkResponse{Short: GenerateShortLink(id, s.env.FillingChar, First(strconv.Atoi(s.env.ShortLen)), s.env.URLDomain, s.dict)}, nil
	}
	// check if it is url or not
	if s.env.CheckURLs == "1" {
		_, err := url.ParseRequestURI(request.Original)
		if err != nil {
			return &shortenerproto.CreateShortLinkResponse{Short: "NOT A URL"}, status.Errorf(codes.InvalidArgument, "NOT A URL")
		}
	}
	short := GenerateShortLink(First(s.storage.GetLastId()), s.env.FillingChar, First(strconv.Atoi(s.env.ShortLen)), s.env.URLDomain, s.dict)
	err := s.storage.Save(request.Original, First(s.storage.GetLastId()))
	if err != nil {
		log.Println(err)
		return nil, status.Errorf(codes.Internal, "INTERNAL ERROR")
	}
	return &shortenerproto.CreateShortLinkResponse{Short: short}, nil
}

func UrlChecker(short string, env *cfg.Env, dictMap *map[rune]bool) bool {
	// check short url length
	if len([]rune(short)) != len([]rune(env.URLDomain))+First(strconv.Atoi(env.ShortLen)) {
		return false
	}

	// check if it is url or not
	if env.CheckURLs == "1" {
		_, err := url.ParseRequestURI(short)
		if err != nil {
			return false
		}
	}

	// check if there are non-dictionary characters
	short = strings.Replace(short, env.URLDomain, "", 1)
	for _, char := range short {
		if !(*dictMap)[char] {
			return false
		}
	}

	// check format
	ok, _ := regexp.MatchString("^"+env.FillingChar+"*[a-zA-Z0-9]+$", short)
	return ok
}

func (s *LinkShortener) GetOriginalLink(ctx context.Context, request *shortenerproto.GetOriginalLinkRequest) (*shortenerproto.GetOriginalLinkResponse, error) {
	// Check short url length
	if !UrlChecker(request.Short, s.env, s.dictMap) {
		return &shortenerproto.GetOriginalLinkResponse{Original: "INVALID FORMAT"}, status.Errorf(codes.InvalidArgument, "INVALID FORMAT")
	}
	decoded := DecodeShortLink(request.Short, s.env.FillingChar, s.env.URLDomain, s.dict)
	original, err := s.storage.Load(decoded)
	if err != nil {
		// url not found
		return &shortenerproto.GetOriginalLinkResponse{Original: "NOT FOUND"}, status.Errorf(codes.NotFound, "NOT FOUND")
	}
	if original == "" {
		return &shortenerproto.GetOriginalLinkResponse{Original: "NOT FOUND"}, status.Errorf(codes.NotFound, "NOT FOUND")
	}
	return &shortenerproto.GetOriginalLinkResponse{Original: original}, nil
}

// Short link generation using base62
func GenerateShortLink(id int, fillerChar string, urlLen int, domain string, dict []rune) string {
	if id == 0 {
		return domain + strings.Repeat(fillerChar, urlLen-1) + string(dict[0])
	}
	base := len(dict)
	var digits []int
	for id > 0 {
		ch := id % base
		digits = append([]int{ch}, digits...)
		id = id / base
	}
	var result []rune
	for _, d := range digits {
		result = append(result, dict[d])
	}
	if len(result) < urlLen {
		return domain + strings.Repeat(fillerChar, urlLen-len(string(result))) + string(result)
	}
	return domain + string(result)
}

func DecodeShortLink(url string, fillerChar, domain string, dict []rune) int {
	base := len(dict)
	result := 0
	url = strings.Replace(url, domain, "", 1)
	url = strings.ReplaceAll(url, fillerChar, "")
	for _, ch := range url {
		for i, a := range dict {
			if a == ch {
				result = result*base + i
			}
		}
	}
	return result
}

func First[T, U any](val T, _ U) T {
	return val
}
