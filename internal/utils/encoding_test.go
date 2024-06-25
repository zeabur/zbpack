package utils

import (
	"encoding/base64"
	"strings"
	"testing"

	"github.com/samber/lo"
	"github.com/spf13/afero"
	"github.com/stretchr/testify/assert"
)

func TestReadToUTF8(t *testing.T) {
	t.Parallel()

	testmap := []struct {
		name          string
		contentBase64 string
		expected      string
		fail          bool
	}{
		{
			name:          "UTF-8",
			contentBase64: "SGVsbG8sIFdvcmxkIQo=",
			expected:      "Hello, World!\n",
		},
		{
			name:          "UTF-8 with BOM",
			contentBase64: "77u/SGVsbG8sIFdvcmxkIQo=",
			expected:      "Hello, World!\n",
		},
		{
			name:          "UTF-16 BE",
			contentBase64: "/v8ASABlAGwAbABvACwAIABXAG8AcgBsAGQAIQAK",
			expected:      "Hello, World!\n",
		},
		{
			name:          "UTF-16 LE (Windows Standard)",
			contentBase64: "//5hAGUAbgB1AG0APQA9ADMALgAxAC4AMQA1AA0ACgBhAGkAbwBoAHQAdABwAD0APQAzAC4AOQAuADUADQAKAGEAaQBvAHMAaQBnAG4AYQBsAD0APQAxAC4AMwAuADEADQAKAGEAbgBuAG8AdABhAHQAZQBkAC0AdAB5AHAAZQBzAD0APQAwAC4ANwAuADAADQAKAGEAcwB5AG4AYwAtAHQAaQBtAGUAbwB1AHQAPQA9ADQALgAwAC4AMwANAAoAYQB0AHQAcgBzAD0APQAyADMALgAyAC4AMAANAAoAYgBsAGkAbgBrAGUAcgA9AD0AMQAuADgALgAyAA0ACgBjAGUAcgB0AGkAZgBpAD0APQAyADAAMgA0AC4ANgAuADIADQAKAGMAaABhAHIAcwBlAHQALQBuAG8AcgBtAGEAbABpAHoAZQByAD0APQAzAC4AMwAuADIADQAKAGMAbABpAGMAawA9AD0AOAAuADEALgA3AA0ACgBjAG8AbABvAHIAYQBtAGEAPQA9ADAALgA0AC4ANgANAAoARABlAHAAcgBlAGMAYQB0AGUAZAA9AD0AMQAuADIALgAxADQADQAKAGYAbABhAHMAawA9AD0AMwAuADAALgAzAA0ACgBmAHIAbwB6AGUAbgBsAGkAcwB0AD0APQAxAC4ANAAuADEADQAKAGYAdQB0AHUAcgBlAD0APQAxAC4AMAAuADAADQAKAGkAZABuAGEAPQA9ADMALgA3AA0ACgBpAG0AcABvAHIAdABsAGkAYgAtAG0AZQB0AGEAZABhAHQAYQA9AD0ANwAuADIALgAxAA0ACgBpAHQAcwBkAGEAbgBnAGUAcgBvAHUAcwA9AD0AMgAuADIALgAwAA0ACgBqAGkAbgBqAGEAMgA9AD0AMwAuADEALgA0AA0ACgBsAGkAbgBlAC0AYgBvAHQALQBzAGQAawA9AD0AMwAuADEAMQAuADAADQAKAE0AYQByAGsAdQBwAFMAYQBmAGUAPQA9ADIALgAxAC4ANQANAAoAbQB1AGwAdABpAGQAaQBjAHQAPQA9ADYALgAwAC4ANQANAAoAcAB5AGQAYQBuAHQAaQBjAD0APQAyAC4ANwAuADQADQAKAHAAeQBkAGEAbgB0AGkAYwAtAGMAbwByAGUAPQA9ADIALgAxADgALgA0AA0ACgBwAHkAdABoAG8AbgAtAGQAYQB0AGUAdQB0AGkAbAA9AD0AMgAuADkALgAwAC4AcABvAHMAdAAwAA0ACgByAGUAcQB1AGUAcwB0AHMAPQA9ADIALgAzADEALgAwAA0ACgBzAGkAeAA9AD0AMQAuADEANgAuADAADQAKAHQAeQBwAGkAbgBnAC0AZQB4AHQAZQBuAHMAaQBvAG4AcwA9AD0ANAAuADEAMgAuADIADQAKAHUAcgBsAGwAaQBiADMAPQA9ADIALgAyAC4AMgANAAoAdwBlAHIAawB6AGUAdQBnAD0APQAzAC4AMAAuADMADQAKAHcAcgBhAHAAdAA9AD0AMQAuADEANgAuADAADQAKAHkAYQByAGwAPQA9ADEALgA5AC4ANAANAAoAegBpAHAAcAA9AD0AMwAuADEAOQAuADIADQAKAA==",
			expected: strings.ReplaceAll(`aenum==3.1.15
aiohttp==3.9.5
aiosignal==1.3.1
annotated-types==0.7.0
async-timeout==4.0.3
attrs==23.2.0
blinker==1.8.2
certifi==2024.6.2
charset-normalizer==3.3.2
click==8.1.7
colorama==0.4.6
Deprecated==1.2.14
flask==3.0.3
frozenlist==1.4.1
future==1.0.0
idna==3.7
importlib-metadata==7.2.1
itsdangerous==2.2.0
jinja2==3.1.4
line-bot-sdk==3.11.0
MarkupSafe==2.1.5
multidict==6.0.5
pydantic==2.7.4
pydantic-core==2.18.4
python-dateutil==2.9.0.post0
requests==2.31.0
six==1.16.0
typing-extensions==4.12.2
urllib3==2.2.2
werkzeug==3.0.3
wrapt==1.16.0
yarl==1.9.4
zipp==3.19.2
`, "\n", "\r\n"),
		},
		{
			name: "Big-5 (Unsupported)",
			// Big-5 is not supported by ReadToUTF8.
			contentBase64: "SGVsbG8sIFdvcmxkISCnQaZupUCsyaFJoUkK",
			fail:          true,
		},
	}

	for _, tt := range testmap {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			// decode base64
			decoded, err := base64.StdEncoding.DecodeString(tt.contentBase64)
			if !assert.NoError(t, err) {
				t.FailNow()
			}

			content, err := ReadToUTF8(decoded)
			if assert.NoError(t, err) {
				if tt.fail {
					assert.NotEqual(t, tt.expected, string(content))
				} else {
					assert.Equal(t, tt.expected, string(content))
				}
			}
		})
	}
}

func TestClassicalReadAllFailed(t *testing.T) {
	t.Parallel()

	utf16Be := lo.Must(base64.StdEncoding.DecodeString("/v8ASABlAGwAbABvACwAIABXAG8AcgBsAGQAIQAK"))

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "test", utf16Be, 0644)

	content, err := afero.ReadFile(fs, "test")
	if assert.NoError(t, err) {
		// UTF-16 strings cannot be directly represented in UTF-8.
		assert.NotEqual(t, "Hello, World!\n", string(content))
	}
}

func TestReadFileToUTF8(t *testing.T) {
	t.Parallel()

	utf8 := lo.Must(base64.StdEncoding.DecodeString("SGVsbG8sIFdvcmxkIQo="))
	utf16Be := lo.Must(base64.StdEncoding.DecodeString("/v8ASABlAGwAbABvACwAIABXAG8AcgBsAGQAIQAK"))

	fs := afero.NewMemMapFs()
	_ = afero.WriteFile(fs, "test", utf16Be, 0644)

	content, err := ReadFileToUTF8(fs, "test")
	if assert.NoError(t, err) {
		assert.Equal(t, "Hello, World!\n", string(content))
	}

	_ = afero.WriteFile(fs, "test2", utf8, 0644)
	content, err = ReadFileToUTF8(fs, "test2")
	if assert.NoError(t, err) {
		assert.Equal(t, "Hello, World!\n", string(content))
	}
}
