package main

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsExcludedNamespace(t *testing.T) {
	assert.False(t, isExcludedNamespace("a-namespace", []string{"kube-system", "kube-public"}))
	assert.False(t, isExcludedNamespace("default", []string{"kube-system", "kube-public"}))
	assert.True(t, isExcludedNamespace("kube-system", []string{"kube-system", "kube-public"}))
	assert.True(t, isExcludedNamespace("kube-public", []string{"kube-system", "kube-public"}))
	assert.False(t, isExcludedNamespace("ns", []string{}))
	assert.False(t, isExcludedNamespace("ns", nil))
}

func TestIsIncludedNamespace(t *testing.T) {
	assert.False(t, isIncludedNamespace("a-namespace", []string{"kube-system", "kube-public"}))
	assert.False(t, isIncludedNamespace("default", []string{"kube-system", "kube-public"}))
	assert.True(t, isIncludedNamespace("kube-system", []string{"kube-system", "kube-public"}))
	assert.True(t, isIncludedNamespace("kube-public", []string{"kube-system", "kube-public"}))
	assert.True(t, isIncludedNamespace("ns", []string{}))
	assert.True(t, isIncludedNamespace("ns", nil))
}
