package service

import indexconfig "code-index-plugin/internal/index/config"

type Options = indexconfig.Options

func DefaultOptions() Options {
	return indexconfig.DefaultOptions()
}

func NormalizeProjectRoot(root string) string {
	return indexconfig.NormalizeProjectRoot(root)
}
