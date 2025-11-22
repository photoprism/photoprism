package header

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsBot(t *testing.T) {
	t.Run("True", func(t *testing.T) {
		assert.True(t, IsBot("Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.6723.69 Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"))
		assert.True(t, IsBot("Mozilla/5.0 (Linux; Android 6.0.1; Nexus 5X Build/MMB29P) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/41.0.2272.96 Mobile Safari/537.36 (compatible; Googlebot/2.1; +http://www.google.com/bot.html)"))
		assert.True(t, IsBot("Mozilla/5.0 (compatible; Bingbot/2.0; +http://www.bing.com/bingbot.htm)"))
		assert.True(t, IsBot("Mozilla/5.0 (Linux; Android 7.0;) AppleWebKit/537.36 (KHTML, like Gecko) Mobile Safari/537.36 (compatible; PetalBot;+https://webmaster.petalsearch.com/site/petalbot)"))
		assert.True(t, IsBot("Mozilla/5.0 (Linux; Android 5.0) AppleWebKit/537.36 (KHTML, like Gecko) Mobile Safari/537.36 (compatible; Bytespider; spider-feedback@bytedance.com)"))
		assert.True(t, IsBot("Mozilla/5.0 (compatible; YandexBot/3.0; +http://yandex.com/bots)"))
		assert.True(t, IsBot("github-camo (b9ea4018)"))
	})
	t.Run("False", func(t *testing.T) {
		assert.False(t, IsBot("Mozilla/5.0 (Macintosh; Intel Mac OS X 10_9_2) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/52.0.2762.73 Safari/537.36"))
		assert.False(t, IsBot("Mozilla/5.0 (Windows NT 10.0) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/89.0.4389.114 Safari/537.36"))
	})
}
