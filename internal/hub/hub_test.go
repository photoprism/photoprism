package hub

import (
	"crypto/rand"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

func TestMain(m *testing.M) {
	log = logrus.StandardLogger()
	log.SetLevel(logrus.TraceLevel)

	ServiceURL = "https://hub-int.photoprism.app/v1/hello"

	code := m.Run()

	os.Exit(code)
}

// Token returns a random token with length of up to 10 characters.
func Token(size uint) string {
	if size > 10 || size < 1 {
		panic(fmt.Sprintf("size out of range: %d", size))
	}

	result := make([]byte, 0, 14)
	b := make([]byte, 8)

	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	randomInt := binary.BigEndian.Uint64(b)

	result = append(result, strconv.FormatUint(randomInt, 36)...)

	for i := len(result); i < cap(result); i++ {
		result = append(result, byte(123-(cap(result)-i)))
	}

	return string(result[:size])
}

func TestNewConfig(t *testing.T) {
	c := NewConfig("test", "testdata/new.yml", "zqkunt22r0bewti9", "test", "PhotoPrism/Test", "test")

	assert.IsType(t, &Config{}, c)
}

func TestNewRequest(t *testing.T) {
	r := NewRequest("test", "zqkunt22r0bewti9", "test", "test", "")

	assert.IsType(t, &Request{}, r)

	t.Logf("Request: %+v", r)

	if j, err := json.Marshal(r); err != nil {
		t.Fatal(err)
	} else {
		t.Logf("JSON: %s", j)
	}
}

func TestConfig_Refresh(t *testing.T) {
	t.Run("success", func(t *testing.T) {
		fileName := fmt.Sprintf("testdata/hub.%s.yml", Token(8))

		c := NewConfig("test", fileName, "zqkunt22r0bewti9", "test", "PhotoPrism/Test", "test")

		if err := c.Update(); err != nil {
			t.Fatal(err)
		}

		assert.Len(t, c.Key, 40)
		assert.Len(t, c.Secret, 32)
		assert.Equal(t, "test", c.Version)

		if sess, err := c.DecodeSession(false); err != nil {
			t.Fatal(err)
		} else if sess.Expired() {
			t.Fatalf("session expired: %+v", sess)
		} else {
			t.Logf("(1) session: %#v", sess)
		}

		if err := c.Save(); err != nil {
			t.Fatal(err)
		}

		defer os.Remove(fileName)

		assert.FileExists(t, fileName)

		if err := c.Update(); err != nil {
			t.Fatal(err)
		}

		assert.Len(t, c.Key, 40)
		assert.Len(t, c.Secret, 32)
		assert.Equal(t, "test", c.Version)

		if sess, err := c.DecodeSession(false); err != nil {
			t.Fatal(err)
		} else if sess.Expired() {
			t.Fatal("session expired")
		} else {
			t.Logf("(2) session: %#v", sess)
		}

		if err := c.Save(); err != nil {
			t.Fatal(err)
		}
		t.Logf("filename: %s", fileName)
		assert.FileExists(t, fileName)
	})
}

func TestConfig_DecodeSession(t *testing.T) {
	t.Run("hub3.yml", func(t *testing.T) {
		c := NewConfig("test", "testdata/hub3.yml", "zqkunt22r0bewti9", "test", "PhotoPrism/Test", "test")

		err := c.Load()

		assert.EqualError(t, err, "session expired")

		assert.Equal(t, "8dd8b115d052f91ac74b1c2475e305009366c487", c.Key)
		assert.Equal(t, "ddf4ce46afbf6c16a6bd8555ab1e4efb", c.Secret)
		assert.Equal(t, "7607796238c26b2d95007957b05c72d63f504346576bc2aa064a6dc54344de47d2ab38422bd1d061c067a16ef517e6054d8b7f5336c120431935518277fed45e49472aaf740cac1bc33ab2e362c767007a59e953e9973709", c.Session)
		assert.Equal(t, Status("unregistered"), c.Status)
		assert.Equal(t, "test", c.Version)
	})
}

func TestConfig_Load(t *testing.T) {
	t.Run("hub1.yml", func(t *testing.T) {
		c := NewConfig("test", "testdata/hub1.yml", "zqkunt22r0bewti9", "test", "PhotoPrism/Test", "test")

		if err := c.Load(); err != nil {
			t.Logf(err.Error())
		}

		assert.Equal(t, "b32e9ccdc90eb7c0f6f1b9fbc82b8a2b0e993304", c.Key)
		assert.Equal(t, "5991ea36a9611e9e00a8360c10b91567", c.Secret)
		assert.Equal(t, "3ef5685c6391a568731c8fc94ccad82d92dea60476c8b672990047c822248f45366fc0e8e812ad15e0b5ae1eb20e866235c56b", c.Session)
		assert.Equal(t, Status("unregistered"), c.Status)
		assert.Equal(t, "test", c.Version)
	})
	t.Run("hub2.yml", func(t *testing.T) {
		c := NewConfig("test", "testdata/hub2.yml", "zqkunt22r0bewti9", "test", "PhotoPrism/Test", "test")

		if err := c.Load(); err != nil {
			t.Logf(err.Error())
		}

		assert.Equal(t, "ab66cb5cfb3658dbea0a1433df048d900934ac68", c.Key)
		assert.Equal(t, "6b0f8440fe307d3120b3a4366350094b", c.Secret)
		assert.Equal(t, "c0ca88fc3094b70a1947b5b10f980a420cd6b1542a20f6f26ecc6a16f340473b9fb16b80be1078e86d886b3a8d46bf8184d147", c.Session)
		assert.Equal(t, Status("unregistered"), c.Status)
		assert.Equal(t, "test", c.Version)
	})
	t.Run("not existing filename", func(t *testing.T) {
		c := NewConfig("test", "testdata/hub_xxx.yml", "zqkunt22r0bewti9", "test", "PhotoPrism/Test", "test")

		if err := c.Load(); err == nil {
			t.Fatal("file should not exist")
		}

		assert.Equal(t, "", c.Key)
		assert.Equal(t, "", c.Secret)
		assert.Equal(t, "", c.Session)
	})
}

func TestConfig_Save(t *testing.T) {
	t.Run("existing filename", func(t *testing.T) {
		assert.FileExists(t, "testdata/hub1.yml")

		c := NewConfig("test", "testdata/hub1.yml", "zqkunt22r0bewti9", "test", "PhotoPrism/Test", "test")

		if err := c.Load(); err != nil {
			t.Logf(err.Error())
		}

		assert.Equal(t, "b32e9ccdc90eb7c0f6f1b9fbc82b8a2b0e993304", c.Key)
		assert.Equal(t, "5991ea36a9611e9e00a8360c10b91567", c.Secret)
		assert.Equal(t, "3ef5685c6391a568731c8fc94ccad82d92dea60476c8b672990047c822248f45366fc0e8e812ad15e0b5ae1eb20e866235c56b", c.Session)
		assert.Equal(t, Status("unregistered"), c.Status)
		assert.Equal(t, "test", c.Version)

		c.FileName = "testdata/hub-save.yml"

		if err := c.Save(); err != nil {
			t.Fatal(err)
		}

		defer os.Remove("testdata/hub-save.yml")

		assert.Equal(t, "b32e9ccdc90eb7c0f6f1b9fbc82b8a2b0e993304", c.Key)
		assert.Equal(t, "5991ea36a9611e9e00a8360c10b91567", c.Secret)
		assert.Equal(t, "3ef5685c6391a568731c8fc94ccad82d92dea60476c8b672990047c822248f45366fc0e8e812ad15e0b5ae1eb20e866235c56b", c.Session)
		assert.Equal(t, Status("unregistered"), c.Status)
		assert.Equal(t, "test", c.Version)

		assert.FileExists(t, "testdata/hub-save.yml")

		if err := c.Load(); err != nil {
			t.Logf(err.Error())
		}

		assert.Equal(t, "b32e9ccdc90eb7c0f6f1b9fbc82b8a2b0e993304", c.Key)
		assert.Equal(t, "5991ea36a9611e9e00a8360c10b91567", c.Secret)
		assert.Equal(t, "3ef5685c6391a568731c8fc94ccad82d92dea60476c8b672990047c822248f45366fc0e8e812ad15e0b5ae1eb20e866235c56b", c.Session)
		assert.Equal(t, Status("unregistered"), c.Status)
		assert.Equal(t, "test", c.Version)
	})
	t.Run("not existing filename", func(t *testing.T) {
		c := NewConfig("test", "testdata/hub_new.yml", "zqkunt22r0bewti9", "test", "PhotoPrism/Test", "test")
		c.Key = "F60F5B25D59C397989E3CD374F81CDD7710A4FCA"
		c.Secret = "foo"
		c.Session = "bar"

		assert.Equal(t, "F60F5B25D59C397989E3CD374F81CDD7710A4FCA", c.Key)
		assert.Equal(t, "foo", c.Secret)
		assert.Equal(t, "bar", c.Session)

		if err := c.Save(); err != nil {
			t.Fatal(err)
		}

		assert.Equal(t, "", c.Key)
		assert.Equal(t, "", c.Secret)
		assert.Equal(t, "", c.Session)

		assert.FileExists(t, "testdata/hub_new.yml")

		if err := os.Remove("testdata/hub_new.yml"); err != nil {
			t.Fatal(err)
		}
	})
}
