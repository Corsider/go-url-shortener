package shortenerserver

import (
	"context"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"go-url-shortener/cfg"
	shortenerproto "go-url-shortener/pkg/proto"
	"testing"
)

type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) Save(original string, id int) error {
	args := m.Called(original, id)
	return args.Error(0)
}

func (m *MockStorage) Load(id int) (string, error) {
	args := m.Called(id)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) GetLastId() (int, error) {
	args := m.Called()
	return args.Int(0), args.Error(1)
}

func (m *MockStorage) CheckExistence(original string) (bool, int) {
	args := m.Called(original)
	return args.Bool(0), args.Int(1)
}

func (m *MockStorage) Close() error {
	args := m.Called()
	return args.Error(0)
}

func TestLinkShortener_CreateShortLink_default(t *testing.T) {
	mockStorage := new(MockStorage)
	env := &cfg.Env{
		FillingChar: "_",
		ShortLen:    "10",
		URLDomain:   "oz.on/",
	}
	linkShortener := New(mockStorage, env)

	mockStorage.On("CheckExistence", "original.com").Return(false, 0)
	mockStorage.On("GetLastId").Return(1, nil)
	mockStorage.On("Save", "original.com", 1).Return(nil)

	response, err := linkShortener.CreateShortLink(context.Background(), &shortenerproto.CreateShortLinkRequest{Original: "original.com"})
	assert.NoError(t, err)
	assert.Equal(t, "oz.on/_________b", response.Short)

	mockStorage.AssertCalled(t, "CheckExistence", "original.com")
	mockStorage.AssertCalled(t, "GetLastId")
	mockStorage.AssertCalled(t, "Save", "original.com", 1)
}

func TestLinkShortener_CreateShortLink_duplicate(t *testing.T) {
	mockStorage := new(MockStorage)
	env := &cfg.Env{
		FillingChar: "-",
		ShortLen:    "5",
		URLDomain:   "sho.rt/",
	}
	linkShortener := New(mockStorage, env)
	mockStorage.On("CheckExistence", "original.com").Return(true, 1)
	mockStorage.On("GetLastId").Return(1, nil)
	mockStorage.On("Save", "original.com", 1).Return(nil)

	response, err := linkShortener.CreateShortLink(context.Background(), &shortenerproto.CreateShortLinkRequest{Original: "original.com"})
	assert.NoError(t, err)
	assert.Equal(t, "sho.rt/----b", response.Short)
	mockStorage.AssertCalled(t, "CheckExistence", "original.com")
}

func TestLinkShortener_GetOriginalLink(t *testing.T) {
	mockStorage := new(MockStorage)
	env := &cfg.Env{
		FillingChar: "-",
		ShortLen:    "5",
		URLDomain:   "sho.rt/",
	}
	linkshortener := New(mockStorage, env)
	mockStorage.On("Load", 211788).Return("original.com", nil)

	response, err := linkshortener.GetOriginalLink(context.Background(), &shortenerproto.GetOriginalLinkRequest{Short: "sho.rt/--3f6"})
	assert.NoError(t, err)
	assert.Equal(t, "original.com", response.Original)
	mockStorage.AssertCalled(t, "Load", 211788)
}

func TestGenerateShortLink(t *testing.T) {
	dict := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	input := 0
	output := GenerateShortLink(input, "_", 5, "te.st/", dict)
	assert.Equal(t, "te.st/____a", output)

	input = 1
	output = GenerateShortLink(input, "_", 5, "te.st/", dict)
	assert.Equal(t, "te.st/____b", output)

	input = 124951343543
	output = GenerateShortLink(input, "_", 10, "te.st/", dict)
	assert.Equal(t, "te.st/___cmylgAR", output)
}

func TestDecodeShortLink(t *testing.T) {
	dict := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	input := "te.st/____a"
	output := DecodeShortLink(input, "_", "te.st/", dict)
	assert.Equal(t, 0, output)

	// Checking max possible link ID for 4-character encoding (62^4 - 1 = 14776335)
	input = "te.st/9999"
	output = DecodeShortLink(input, "_", "te.st/", dict)
	assert.Equal(t, 14776335, output)
}

func TestUrlChecker(t *testing.T) {
	dict := []rune("abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789")
	input := "https://te.st/--a"
	env := cfg.Env{
		FillingChar: "-",
		URLDomain:   "https://te.st/",
		ShortLen:    "3",
		CheckURLs:   "1",
	}
	dict1 := append(dict, []rune(env.FillingChar)...)
	m := make(map[rune]bool)
	for _, d := range dict1 {
		m[d] = true
	}
	output := UrlChecker(input, &env, &m)
	assert.Equal(t, true, output)

	env.ShortLen = "4"
	output = UrlChecker(input, &env, &m)
	assert.Equal(t, false, output)

	input = "https://te.st/a---"
	output = UrlChecker(input, &env, &m)
	assert.Equal(t, false, output)

	input = "https://te.st/----"
	output = UrlChecker(input, &env, &m)
	assert.Equal(t, false, false)

	input = "https://te.st/aoao"
	output = UrlChecker(input, &env, &m)
	assert.Equal(t, true, output)
}
