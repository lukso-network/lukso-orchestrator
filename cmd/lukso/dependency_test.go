package main

import (
	"github.com/lukso-network/lukso-orchestrator/shared/testutil/assert"
	"testing"
)

func TestClientDependency_ParseUrl(t *testing.T) {
	dependencyWithSprintf := &ClientDependency{
		baseUnixUrl:   "https://something.com/%s",
		baseDarwinUrl: "https://something.com/%s-darwin",
		name:          "dummy",
	}
	dependencyWithOutSprintf := &ClientDependency{
		baseUnixUrl:   "https://something.com/",
		baseDarwinUrl: "https://something.com/darwin",
		name:          "dummy",
	}
	tagName := "v6.6.6-dummy"
	t.Run("should parse macos flag", func(t *testing.T) {
		systemOs = macos
		assert.Equal(
			t,
			"https://something.com/v6.6.6-dummy-darwin",
			dependencyWithSprintf.ParseUrl(tagName),
		)
		assert.Equal(
			t,
			"https://something.com/darwin",
			dependencyWithOutSprintf.ParseUrl(tagName),
		)
	})
	t.Run("should work without flag flag", func(t *testing.T) {
		systemOs = ubuntu
		assert.Equal(
			t,
			"https://something.com/v6.6.6-dummy",
			dependencyWithSprintf.ParseUrl(tagName),
		)
		assert.Equal(
			t,
			"https://something.com/",
			dependencyWithOutSprintf.ParseUrl(tagName),
		)
	})
}
