package config

import (
	"io/fs"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseEmptyPath(t *testing.T) {
	var testConfStruct struct {
		A int    `yaml:"a"`
		B string `yaml:"b"`
	}
	err := ParseConfig("", &testConfStruct)
	assert.Error(t, err)
}

func TestParseDirInsteadOfFile(t *testing.T) {
	var testConfStruct struct {
		A int    `yaml:"a"`
		B string `yaml:"b"`
	}
	err := ParseConfig("cmd", &testConfStruct)
	assert.Error(t, err)
}

func TestParseUnexistentFile(t *testing.T) {
	var testConfStruct struct {
		A int    `yaml:"a"`
		B string `yaml:"b"`
	}
	err := ParseConfig("/unexistent_file", &testConfStruct)
	assert.Error(t, err)
}

func TestValidPrimitives(t *testing.T) {
	var (
		testFilePath    = "/tmp/go_test_file"
		testFileContent = `
i1: -31
u1: 123
s1: Elise
s2: lemon stripe
`
		testConf struct {
			I1 int    `yaml:"i1"`
			U1 uint   `yaml:"u1"`
			S1 string `yaml:"s1"`
			S2 string `yaml:"s2"`
		}
	)
	err := os.WriteFile(testFilePath, []byte(testFileContent), fs.ModePerm)
	if err != nil {
		t.Skip("could not write temporary file")
	}
	require.NoError(t, ParseConfig(testFilePath, &testConf))
	assert.Equal(t, -31, testConf.I1)
	assert.Equal(t, uint(123), testConf.U1)
	assert.Equal(t, "Elise", testConf.S1)
	assert.Equal(t, "lemon stripe", testConf.S2)
}

func TestValidUrls(t *testing.T) {
	var (
		testFilePath    = "/tmp/go_test_file"
		testFileContent = `
u1: "http://test@neverssl.com"
u2: "https://yandex.ru/search/nothing"
u3: "https://www.youtube.com/watch?v=-gDinVAmtA0"
s1: skippppp
u4: "postgres://postUser:postPassword@1.1.1.1:5432/dbname?sssssss=no"
s2: skippppp
`
		testConf struct {
			U1 YamlURL `yaml:"u1"`
			U2 YamlURL `yaml:"u2"`
			U3 YamlURL `yaml:"u3"`
			U4 YamlURL `yaml:"u4"`
		}
	)
	err := os.WriteFile(testFilePath, []byte(testFileContent), fs.ModePerm)
	if err != nil {
		t.Skip("could not write temporary file")
	}
	require.NoError(t, ParseConfig(testFilePath, &testConf))
	assert.ElementsMatch(t,
		[]any{testConf.U1.Scheme, testConf.U2.Scheme, testConf.U3.Scheme, testConf.U4.Scheme},
		[]any{"http", "https", "https", "postgres"},
	)
	assert.ElementsMatch(t,
		[]any{testConf.U1.Host, testConf.U2.Host, testConf.U3.Host, testConf.U4.Host},
		[]any{"neverssl.com", "yandex.ru", "www.youtube.com", "1.1.1.1:5432"},
	)
	assert.ElementsMatch(t,
		[]any{testConf.U1.User.String(), testConf.U2.User.String(), testConf.U3.User.String(), testConf.U4.User.String()},
		[]any{"test", "", "", "postUser:postPassword"},
	)
	assert.ElementsMatch(t,
		[]any{testConf.U1.Path, testConf.U2.Path, testConf.U3.Path, testConf.U4.Path},
		[]any{"", "/search/nothing", "/watch", "/dbname"},
	)
}
