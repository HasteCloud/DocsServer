package main

type GitConfig struct {
	Title string `json:"title" default:"Documentation"`
	Home string `json:"home" default:"README.md"`
	Style string `json:"style"`
	Icon string `json:"icon"`
	Root string `json:"root"`
	GenerateSubNavHeadings bool `json:"generateSubNavHeadings"`
	LowerSubNavHeadingBound int `json:"lowerSubNavHeadingBound"`
	UpperSubNavHeadingBound int `json:"upperSubNavHeadingBound"`
	NavItems []Item `json:"nav"`
	SubNavItems []Item `json:"subnav"`
}

type Item struct {
	Title string `json:"title"`
	Path string `json:"path"`
	Level string `json:"level"`
}

func (gitConfig GitConfig) New() GitConfig {
	if len(gitConfig.Title) == 0 {
		gitConfig.Title = "Documentation"
	}
	if len(gitConfig.Home) == 0 {
		gitConfig.Home = "README.md"
	}
	return gitConfig
}