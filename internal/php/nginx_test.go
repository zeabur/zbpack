package php

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/zeabur/zbpack/pkg/types"
)

func TestEscape(t *testing.T) {
	assert.Equal(t, escape("a\nb"), "a\\nb")
	assert.Equal(t, escape("a$b"), "a\\$b")
}

func TestRetrieveNginxConf_Default(t *testing.T) {
	conf, err := RetrieveNginxConf(string(types.PHPApplicationDefault))

	assert.NoError(t, err)
	assert.Contains(t, conf, escape("try_files $uri $uri/ /index.php?$query_string;"))
}

func TestRetrieveNginxConf_AcgFaka(t *testing.T) {
	conf, err := RetrieveNginxConf(string(types.PHPApplicationAcgFaka))

	assert.NoError(t, err)
	assert.Contains(t, conf, escape("rewrite ^(.*)$ /index.php?s=$1 last; break;"))
}

func TestRetrieveNginxConf_Unknown(t *testing.T) {
	_, err := RetrieveNginxConf("unknown")

	assert.Error(t, err)
}
